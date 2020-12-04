package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/valeamoris/go-ezio/rest/internal"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func TestLogHandler(t *testing.T) {
	handlers := []func(next echo.HandlerFunc) echo.HandlerFunc{
		LogMiddleware,
		DetailedLogMiddleware,
	}

	e := echo.New()
	for _, logHandler := range handlers {
		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		handler := logHandler(func(ctx echo.Context) error {
			ctx.Request().Context().Value(internal.LogContext).(*internal.LogCollector).Append("anything")
			ctx.Response().Header().Set("X-Test", "test")
			return ctx.String(http.StatusServiceUnavailable, "content")
		})

		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)
		err := handler(ctx)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
		assert.Equal(t, "test", resp.Header().Get("X-Test"))
		assert.Equal(t, "content", resp.Body.String())
	}
}

func TestLogHandlerSlow(t *testing.T) {
	handlers := []func(next echo.HandlerFunc) echo.HandlerFunc{
		LogMiddleware,
		DetailedLogMiddleware,
	}

	e := echo.New()
	for _, logHandler := range handlers {
		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		handler := logHandler(func(ctx echo.Context) error {
			time.Sleep(slowThreshold + time.Millisecond*50)
			return nil
		})

		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)
		err := handler(ctx)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
	}
}

func BenchmarkLogHandler(b *testing.B) {
	b.ReportAllocs()

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	handler := LogMiddleware(func(ctx echo.Context) error {
		ctx.Response().WriteHeader(http.StatusOK)
		return nil
	})

	e := echo.New()
	for i := 0; i < b.N; i++ {
		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)
		_ = handler(ctx)
	}
}
