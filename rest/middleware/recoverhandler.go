package middleware

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/valeamoris/go-ezio/rest/internal"
	"net/http"
	"runtime/debug"
)

func RecoverMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		defer func() {
			if r := recover(); r != nil {
				internal.Error(ctx, fmt.Sprintf("%v\n%s", r, debug.Stack()))
				ctx.Response().WriteHeader(http.StatusInternalServerError)
			}
		}()

		return next(ctx)
	}
}
