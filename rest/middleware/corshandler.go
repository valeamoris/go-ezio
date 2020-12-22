package middleware

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/valeamoris/go-ezio/rest"
)

func CORSMiddleware() rest.MiddlewareFunc {
	return func(next rest.HandlerFunc) rest.HandlerFunc {
		return rest.HandlerFunc(echoMiddleware.CORS()(echo.HandlerFunc(next)))
	}
}
