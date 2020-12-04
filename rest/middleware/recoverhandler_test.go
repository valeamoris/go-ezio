package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func TestWithPanic(t *testing.T) {
	handler := RecoverMiddleware(func(ctx echo.Context) error {
		panic("whatever")
		return nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestWithoutPanic(t *testing.T) {
	handler := RecoverMiddleware(func(ctx echo.Context) error {
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
