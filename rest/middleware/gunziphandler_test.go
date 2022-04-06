package middleware

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/codec"
)

func TestGunzipHandler(t *testing.T) {
	const message = "hello world"
	var wg sync.WaitGroup
	wg.Add(1)
	handler := GunzipMiddleware(func(ctx echo.Context) error {
		body, err := ioutil.ReadAll(ctx.Request().Body)
		assert.Nil(t, err)
		assert.Equal(t, string(body), message)
		wg.Done()
		return nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "http://localhost",
		bytes.NewReader(codec.Gzip([]byte(message))))
	req.Header.Set(echo.HeaderContentEncoding, gzipEncoding)
	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	wg.Wait()
}

func TestGunzipHandler_NoGzip(t *testing.T) {
	const message = "hello world"
	var wg sync.WaitGroup
	wg.Add(1)
	handler := GunzipMiddleware(func(ctx echo.Context) error {
		body, err := ioutil.ReadAll(ctx.Request().Body)
		assert.Nil(t, err)
		assert.Equal(t, string(body), message)
		wg.Done()
		return nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "http://localhost",
		strings.NewReader(message))
	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	wg.Wait()
}

func TestGunzipHandler_NoGzipButTelling(t *testing.T) {
	const message = "hello world"
	handler := GunzipMiddleware(func(context echo.Context) error {
		return nil
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "http://localhost",
		strings.NewReader(message))
	req.Header.Set(echo.HeaderContentEncoding, gzipEncoding)
	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}
