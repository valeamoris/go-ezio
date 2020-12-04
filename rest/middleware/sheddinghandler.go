package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/tal-tech/go-zero/core/load"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/stat"
	"net/http"
	"sync"
)

const serviceType = "api"

var (
	sheddingStat *load.SheddingStat
	lock         sync.Mutex
)

func SheddingMiddleware(shedder load.Shedder, metrics *stat.Metrics) echo.MiddlewareFunc {
	if shedder == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx echo.Context) error {
				return next(ctx)
			}
		}
	}

	ensureSheddingStat()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			sheddingStat.IncrementTotal()
			promise, err := shedder.Allow()
			if err != nil {
				metrics.AddDrop()
				sheddingStat.IncrementDrop()
				logx.Errorf("[http] dropped, %s - %s - %s",
					ctx.Request().RequestURI, ctx.RealIP(), ctx.Request().UserAgent())
				ctx.Response().WriteHeader(http.StatusServiceUnavailable)
				return nil
			}

			err = next(ctx)
			if err != nil || ctx.Response().Status == http.StatusServiceUnavailable {
				promise.Fail()
			} else {
				sheddingStat.IncrementPass()
				promise.Pass()
			}
			return err
		}
	}
}

func ensureSheddingStat() {
	lock.Lock()
	if sheddingStat == nil {
		sheddingStat = load.NewSheddingStat(serviceType)
	}
	lock.Unlock()
}
