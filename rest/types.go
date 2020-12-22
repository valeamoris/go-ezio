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

	HandlerFunc = echo.HandlerFunc

	Route struct {
		Method  string
		Path    string
		Handler HandlerFunc
	}

	Group struct {
		Prefix   string
		priority bool
		jwt      jwtSetting
		Routes   []Route
		echo.Group
		middlewares []Middleware
	}

	Validator = echo.Validator

	Context = echo.Context

	Middleware echo.MiddlewareFunc
)
