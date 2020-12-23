package middleware

import (
	"compress/gzip"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

const gzipEncoding = "gzip"

func GunzipMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if strings.Contains(ctx.Request().Header.Get(echo.HeaderContentEncoding), gzipEncoding) {
			reader, err := gzip.NewReader(ctx.Request().Body)
			if err != nil {
				ctx.Response().WriteHeader(http.StatusBadRequest)
				return nil
			}
			r := ctx.Request()
			r.Body = reader
			ctx.SetRequest(r)
		}
		return next(ctx)
	}
}
