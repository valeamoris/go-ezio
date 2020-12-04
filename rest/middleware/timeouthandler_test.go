package middleware

import (
	"github.com/labstack/echo/v4"
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

func TestTimeout(t *testing.T) {
	timeoutHandler := TimeoutMiddleware(time.Millisecond)
	handler := timeoutHandler(func(ctx echo.Context) error {
		time.Sleep(time.Minute)
		return nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusGatewayTimeout, resp.Code)
}

func TestWithinTimeout(t *testing.T) {
	timeoutHandler := TimeoutMiddleware(time.Second)
	handler := timeoutHandler(func(ctx echo.Context) error {
		time.Sleep(time.Millisecond)
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

func TestWithoutTimeout(t *testing.T) {
	timeoutHandler := TimeoutMiddleware(0)
	handler := timeoutHandler(func(ctx echo.Context) error {
		time.Sleep(100 * time.Millisecond)
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
