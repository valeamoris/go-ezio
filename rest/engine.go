package rest

import (
	"fmt"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/bytes"
	"github.com/valeamoris/go-ezio/rest/middleware"
	"github.com/zeromicro/go-zero/core/breaker"
	"github.com/zeromicro/go-zero/core/load"
	"github.com/zeromicro/go-zero/core/stat"
	"github.com/zeromicro/go-zero/core/sysx"
	"io"
	"time"
)

// 1000m代表100%
const topCpuUsage = 1000

type (
	engine struct {
		conf Conf
		*echo.Echo
		// 系统负载
		shedder load.Shedder
		// 预警负载 取最高和配置负载平均
		priorityShedder load.Shedder
		closers         []io.Closer
		groups          []Group
		rejectHandler   func(promise breaker.Promise, err error)
	}
)

func newEngine(conf Conf) *engine {
	srv := &engine{
		conf: conf,
		Echo: echo.New(),
	}
	if conf.CpuThreshold > 0 {
		srv.shedder = load.NewAdaptiveShedder(load.WithCpuThreshold(conf.CpuThreshold))
		srv.priorityShedder = load.NewAdaptiveShedder(load.WithCpuThreshold(
			(conf.CpuThreshold + topCpuUsage) >> 1))
	}
	return srv
}

func (s *engine) AddGroup(g Group) {
	s.groups = append(s.groups, g)
}

func (s *engine) bindGroup(g Group, metrics *stat.Metrics) error {
	// todo 签名
	s.signatureVerifier()

	group := s.Group(g.Prefix)

	if g.shedding {
		// 自定义负载保护
		group.Use(middleware.SheddingMiddleware(s.getShedder(g.priority), metrics))
	}

	if !g.timeoutDisabled {
		// 超时
		group.Use(middleware.TimeoutMiddleware(time.Duration(s.conf.Timeout) * time.Millisecond))
	}

	// JWT的认证中间件
	if g.jwt.enabled {
		conf := echoMiddleware.DefaultJWTConfig
		conf.Claims = g.jwt.claims
		conf.SigningKey = []byte(g.jwt.secret)
		group.Use(echoMiddleware.JWTWithConfig(conf))
	}

	if g.static.enabled {
		group.Static(g.static.prefix, g.static.root)
	}

	for _, m := range g.middlewares {
		middlewareWrapper := func(m Middleware) echo.MiddlewareFunc {
			return func(next echo.HandlerFunc) echo.HandlerFunc {
				return echo.HandlerFunc(m(HandlerFunc(next)))
			}
		}

		group.Use(middlewareWrapper(m))
	}

	for _, route := range g.Routes {
		s.bindRoute(group, metrics, route)
	}
	return nil
}

func (s *engine) getShedder(priority bool) load.Shedder {
	if priority && s.priorityShedder != nil {
		return s.priorityShedder
	}
	return s.shedder
}

func (s *engine) bindRoute(g *echo.Group, metrics *stat.Metrics, route Route) {
	g.Add(route.Method, route.Path, echo.HandlerFunc(route.Handler)) // todo 断路器逻辑还有点问题，下次再优化
	// middleware.BreakerMiddleware(route.Method, route.Path, metrics, s.rejectHandler),

}

func (s *engine) bindRoutes() error {
	metrics := s.createMetrics()

	traceMiddleware, closer := middleware.TracingMiddleware(sysx.Hostname())
	s.closers = append(s.closers, closer)

	// 追踪
	s.Echo.Use(traceMiddleware)
	// 日志记录
	s.Echo.Use(s.getLogMiddleware())
	// 单连接最大连接数
	s.Echo.Use(middleware.MaxConnMiddleware(s.conf.MaxConns))
	// recover恢复
	s.Echo.Use(middleware.RecoverMiddleware)
	// 数据统计
	s.Echo.Use(middleware.MetricMiddleware(metrics))
	// 全局初始化prometheus中间件
	// prometheus监控
	s.Echo.Use(middleware.PrometheusMiddleware())

	// 最大body limit
	s.Echo.Use(echoMiddleware.BodyLimit(bytes.Format(s.conf.MaxBytes)))
	// gzip request的支持
	s.Echo.Use(middleware.GunzipMiddleware)
	for _, fr := range s.groups {
		if err := s.bindGroup(fr, metrics); err != nil {
			return err
		}
	}
	return nil
}

func (s *engine) createMetrics() *stat.Metrics {
	var metrics *stat.Metrics

	if len(s.conf.Name) > 0 {
		metrics = stat.NewMetrics(s.conf.Name)
	} else {
		metrics = stat.NewMetrics(fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port))
	}

	return metrics
}

func (s *engine) getLogMiddleware() echo.MiddlewareFunc {
	if s.conf.Verbose {
		return middleware.DetailedLogMiddleware
	} else {
		return middleware.LogMiddleware
	}
}

// todo 签名
func (s *engine) signatureVerifier() {}

func (s *engine) startGroup() error {
	if err := s.bindRoutes(); err != nil {
		return err
	}
	return s.Echo.Start(fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port))
}

func (s *engine) Start() error {
	return s.startGroup()
}

func (s *engine) Close() error {
	for _, closer := range s.closers {
		closer.Close()
	}
	return s.Echo.Close()
}
