package middleware

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/tal-tech/go-zero/core/breaker"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/stat"
	"net/http"
	"strings"
)

const breakerSeparator = "://"

func BreakerMiddleware(method, path string, metrics *stat.Metrics, rejectHandler func(promise breaker.Promise, err error)) echo.MiddlewareFunc {
	brk := breaker.NewBreaker(
		breaker.WithName(strings.Join([]string{method, path}, breakerSeparator)),
	)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			promise, err := brk.Allow()
			if err != nil {
				metrics.AddDrop()
				logx.Errorf("[http] dropped, %s - %s - %s",
					ctx.Request().RequestURI, ctx.RealIP(), ctx.Request().UserAgent())
				ctx.Response().WriteHeader(http.StatusServiceUnavailable)
				return nil
			}
			defer func() {
				if ctx.Response().Status < http.StatusInternalServerError {
					promise.Accept()
				} else {
					promise.Reject(fmt.Sprintf("%d %s", ctx.Response().Status, http.StatusText(ctx.Response().Status)))
				}
			}()
			err = next(ctx)
			if err != nil {
				rejectHandler(promise, err)
			}
			return err
		}
	}
}
