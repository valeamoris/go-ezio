package rest

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/tal-tech/go-zero/core/breaker"
	"github.com/tal-tech/go-zero/core/logx"
	"log"
	"net/http"
)

type (
	runOptions struct {
		start func(*engine) error
		close func(*engine) error
	}

	RunOption func(*Server)

	Server struct {
		engine *engine
		opts   runOptions
	}
)

func MustNewServer(c Conf, opts ...RunOption) *Server {
	engine, err := NewServer(c, opts...)
	if err != nil {
		log.Fatal(err)
	}

	return engine
}

func NewServer(c Conf, opts ...RunOption) (*Server, error) {
	if err := c.SetUp(); err != nil {
		return nil, err
	}

	server := &Server{
		engine: newEngine(c),
		opts: runOptions{
			start: func(srv *engine) error {
				return srv.Start()
			},
			close: func(srv *engine) error {
				return srv.Close()
			},
		},
	}

	for _, opt := range opts {
		opt(server)
	}

	return server, nil
}

// 设置验证器
func WithValidator(validator Validator) RunOption {
	return func(srv *Server) {
		srv.engine.Validator = validator
	}
}

func WithRenderer(renderer Renderer) RunOption {
	return func(srv *Server) {
		srv.engine.Renderer = renderer
	}
}

func WithErrorHandler(errHandler func(err error, ctx Context)) RunOption {
	return func(srv *Server) {
		srv.engine.HTTPErrorHandler = errHandler
	}
}

func WithBreakerRejectHandler(rejectHandler func(promise breaker.Promise, err error)) RunOption {
	return func(srv *Server) {
		srv.engine.rejectHandler = rejectHandler
	}
}

func (e *Server) NewContext(r *http.Request, w http.ResponseWriter) Context {
	return e.engine.NewContext(r, w)
}

func (e *Server) Start() {
	handlerError(e.opts.start(e.engine))
}

// graceful shutdown
func (e *Server) Shutdown(ctx context.Context) error {
	return e.engine.Shutdown(ctx)
}

func (e *Server) Stop() {
	_ = e.opts.close(e.engine)
	_ = logx.Close()
}

func (e *Server) Group(g Group, opts ...RouteOption) {
	for _, opt := range opts {
		opt(&g)
	}
	e.engine.AddGroup(g)
}

func (e *Server) Use(middlewares ...Middleware) {
	for _, m := range middlewares {
		e.engine.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return echo.HandlerFunc(m(HandlerFunc(next)))
		})
	}
}

func handlerError(err error) {
	if err == nil || err == http.ErrServerClosed {
		return
	}

	logx.Error(err)
	panic(err)
}

// 校验Jwt
func WithJwt(secret string, claims jwt.Claims) RouteOption {
	return func(r *Group) {
		validateSecret(secret)
		r.jwt.enabled = true
		r.jwt.secret = secret
		r.jwt.claims = claims
	}
}

func validateSecret(secret string) {
	if len(secret) < 8 {
		panic("secret's length can't be less than 8")
	}
}

func WithMiddlewares(ms ...Middleware) RouteOption {
	return func(r *Group) {
		r.middlewares = ms
	}
}

func WithPriority() RouteOption {
	return func(r *Group) {
		r.priority = true
	}
}

func WithShedding() RouteOption {
	return func(r *Group) {
		r.shedding = true
	}
}

func WithTimeoutDisabled() RouteOption {
	return func(r *Group) {
		r.timeoutDisabled = true
	}
}

func WithStatic(prefix, root string) RouteOption {
	return func(r *Group) {
		r.static.enabled = true
		r.static.prefix = prefix
		r.static.root = root
	}
}
