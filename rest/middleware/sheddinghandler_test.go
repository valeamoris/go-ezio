package middleware

import (
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/load"
	"github.com/zeromicro/go-zero/core/stat"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func TestSheddingHandlerAccept(t *testing.T) {
	metrics := stat.NewMetrics("unit-test")
	shedder := mockShedder{
		allow: true,
	}
	sheddingHandler := SheddingMiddleware(shedder, metrics)
	handler := sheddingHandler(func(ctx echo.Context) error {
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

func TestSheddingHandlerFail(t *testing.T) {
	metrics := stat.NewMetrics("unit-test")
	shedder := mockShedder{
		allow: true,
	}
	sheddingHandler := SheddingMiddleware(shedder, metrics)
	handler := sheddingHandler(func(ctx echo.Context) error {
		ctx.Response().WriteHeader(http.StatusServiceUnavailable)
		return nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
}

func TestSheddingHandlerReject(t *testing.T) {
	metrics := stat.NewMetrics("unit-test")
	shedder := mockShedder{
		allow: false,
	}
	sheddingHandler := SheddingMiddleware(shedder, metrics)
	handler := sheddingHandler(func(ctx echo.Context) error {
		ctx.Response().WriteHeader(http.StatusOK)
		return nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
}

func TestSheddingHandlerNoShedding(t *testing.T) {
	metrics := stat.NewMetrics("unit-test")
	sheddingHandler := SheddingMiddleware(nil, metrics)
	handler := sheddingHandler(func(ctx echo.Context) error {
		ctx.Response().WriteHeader(http.StatusOK)
		return nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
}

type mockShedder struct {
	allow bool
}

func (s mockShedder) Allow() (load.Promise, error) {
	if s.allow {
		return mockPromise{}, nil
	} else {
		return nil, load.ErrServiceOverloaded
	}
}

type mockPromise struct {
}

func (p mockPromise) Pass() {
}

func (p mockPromise) Fail() {
}
