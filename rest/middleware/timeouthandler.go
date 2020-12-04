package middleware

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

const reason = "Request Timeout"

func TimeoutMiddleware(duration time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		if duration > 0 {
			return func(ctx echo.Context) error {
				var cancelCtx context.CancelFunc
				newCtx, cancelCtx := context.WithTimeout(ctx.Request().Context(), duration)
				defer cancelCtx()
				done := make(chan error, 1)

				go func() {
					ctx.SetRequest(ctx.Request().WithContext(newCtx))
					done <- next(ctx)
				}()

				select {
				case <-newCtx.Done():
					return ctx.String(http.StatusGatewayTimeout, newCtx.Err().Error())
				case err := <-done:
					return err
				}
			}
		} else {
			return next
		}
	}
}
