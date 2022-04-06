package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/valeamoris/go-ezio/rest/internal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/syncx"
	"net/http"
)

func MaxConnMiddleware(n int) echo.MiddlewareFunc {
	if n <= 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx echo.Context) error {
				return next(ctx)
			}
		}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		latch := syncx.NewLimit(n)
		return func(ctx echo.Context) error {
			if latch.TryBorrow() {
				defer func() {
					if err := latch.Return(); err != nil {
						logx.Error(err)
					}
				}()

				return next(ctx)
			} else {
				internal.Errorf(ctx, "concurrent connections over %d, reject with code %d",
					n, http.StatusServiceUnavailable)
				ctx.Response().WriteHeader(http.StatusServiceUnavailable)
				return nil
			}
		}
	}
}
