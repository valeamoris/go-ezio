package middleware

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/valeamoris/go-ezio/rest"
)

func SecurityMiddleware() rest.MiddlewareFunc {
	return func(next rest.HandlerFunc) rest.HandlerFunc {
		return rest.HandlerFunc(echoMiddleware.Secure()(echo.HandlerFunc(next)))
	}
}
