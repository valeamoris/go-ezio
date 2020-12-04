package rest

import "github.com/labstack/echo/v4"

type (
	jwtSetting struct {
		enabled    bool
		secret     string
		prevSecret string
	}

	RouteOption func(r *Group)

	Route struct {
		Method  string
		Path    string
		Handler echo.HandlerFunc
	}

	Group struct {
		Prefix   string
		priority bool
		jwt      jwtSetting
		Routes   []Route
		echo.Group
		middlewares []Middleware
	}

	Middleware echo.MiddlewareFunc
)
