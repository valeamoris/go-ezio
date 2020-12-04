package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/tal-tech/go-zero/core/stat"
	"github.com/tal-tech/go-zero/core/timex"
)

func MetricMiddleware(metrics *stat.Metrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			startTime := timex.Now()
			defer func() {
				metrics.Add(stat.Task{
					Duration: timex.Since(startTime),
				})
			}()
			return next(ctx)
		}
	}
}
