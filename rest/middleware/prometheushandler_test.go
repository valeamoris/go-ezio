package middleware

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromMetricHandler(t *testing.T) {
	prometheusHandler := PrometheusMiddleware()
	handler := prometheusHandler(func(ctx echo.Context) error {
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
