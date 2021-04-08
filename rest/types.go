package rest

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

type (
	jwtSetting struct {
		enabled    bool
		secret     string
		prevSecret string
		claims     jwt.Claims
	}

	RouteOption func(r *Group)

	HandlerFunc func(ctx Context) error

	Route struct {
		Method  string
		Path    string
		Handler HandlerFunc
	}

	Group struct {
		Prefix   string
		priority bool
		jwt      jwtSetting
		static   staticSetting
		// should open shedding
		shedding      bool
		enableBreaker bool
		// should enable timeout middleware
		timeoutDisabled bool
		Routes          []Route
		echo.Group
		middlewares []Middleware
	}

	Validator = echo.Validator

	Context = echo.Context

	Renderer = echo.Renderer

	MiddlewareFunc = func(next HandlerFunc) HandlerFunc

	Middleware = MiddlewareFunc

	staticSetting struct {
		enabled bool
		prefix  string
		root    string
	}
)
