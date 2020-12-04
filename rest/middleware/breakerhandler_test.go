package middleware

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/stat"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	logx.Disable()
	stat.SetReporter(nil)
}

func TestBreakerMiddlewareAccept(t *testing.T) {
	metrics := stat.NewMetrics("unit-test")
	breakerHandler := BreakerMiddleware(http.MethodGet, "/", metrics)
	handler := breakerHandler(func(ctx echo.Context) error {
		ctx.Response().Header().Set("X-Test", "test")
		return ctx.String(http.StatusOK, "content")
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Set("X-Test", "test")
	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "test", resp.Header().Get("X-Test"))
	assert.Equal(t, "content", resp.Body.String())
}

func TestBreakerMiddlewareFail(t *testing.T) {
	metrics := stat.NewMetrics("unit-test")
	breakerHandler := BreakerMiddleware(http.MethodGet, "/", metrics)
	handler := breakerHandler(func(ctx echo.Context) error {
		ctx.Response().WriteHeader(http.StatusBadGateway)
		return nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadGateway, resp.Code)
}

func TestBreakerMiddleware_4XX(t *testing.T) {
	metrics := stat.NewMetrics("unit-test")
	breakerHandler := BreakerMiddleware(http.MethodGet, "/", metrics)
	handler := breakerHandler(func(ctx echo.Context) error {
		ctx.Response().WriteHeader(http.StatusBadRequest)
		return nil
	})

	e := echo.New()
	for i := 0; i < 1000; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)
		err := handler(ctx)
		assert.Nil(t, err)
	}

	const tries = 100
	var pass int
	for i := 0; i < tries; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)
		err := handler(ctx)
		assert.Nil(t, err)
		if resp.Code == http.StatusBadRequest {
			pass++
		}
	}

	assert.Equal(t, tries, pass)
}

func TestBreakerMiddlewareReject(t *testing.T) {
	metrics := stat.NewMetrics("unit-test")
	breakerHandler := BreakerMiddleware(http.MethodGet, "/", metrics)
	handler := breakerHandler(func(ctx echo.Context) error {
		ctx.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	})

	e := echo.New()
	for i := 0; i < 1000; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)
		err := handler(ctx)
		assert.Nil(t, err)
	}

	var drops int
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)
		err := handler(ctx)
		assert.Nil(t, err)
		if resp.Code == http.StatusServiceUnavailable {
			drops++
		}
	}

	assert.True(t, drops >= 80, fmt.Sprintf("expected to be greater than 80, but got %d", drops))
}
