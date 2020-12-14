package rest

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/rest"
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

func MustNewServer(c rest.RestConf, opts ...RunOption) *Server {
	engine, err := NewServer(c, opts...)
	if err != nil {
		log.Fatal(err)
	}

	return engine
}

func NewServer(c rest.RestConf, opts ...RunOption) (*Server, error) {
	if len(opts) > 1 {
		return nil, errors.New("only one RunOption is allowed")
	}

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

func (e *Server) Start() {
	handlerError(e.opts.start(e.engine))
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
		e.engine.Use(echo.MiddlewareFunc(m))
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
func WithJwt(secret string) RouteOption {
	return func(r *Group) {
		validateSecret(secret)
		r.jwt.enabled = true
		r.jwt.secret = secret
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
