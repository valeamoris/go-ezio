package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tal-tech/go-zero/core/stat"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricMiddleware(t *testing.T) {
	metrics := stat.NewMetrics("unit-test")
	metricHandler := MetricMiddleware(metrics)
	handler := metricHandler(func(context echo.Context) error {
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
