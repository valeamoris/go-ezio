package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-client-go"
	"net/http"
	"net/http/httptest"
	"testing"
)

const tradeId = "4f0b2ce95792ae3"

func TestTracingHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Set("x-b3-traceid", tradeId)
	req.Header.Set("x-b3-spanid", tradeId)
	md, closer := TracingMiddleware("test")
	defer closer.Close()

	handler := md(func(ctx echo.Context) error {
		span := opentracing.SpanFromContext(ctx.Request().Context())
		assert.NotNil(t, span)
		spanCtx := span.Context().(jaeger.SpanContext)
		assert.Equal(t, tradeId, spanCtx.TraceID().String())
		return nil
	})

	resp := httptest.NewRecorder()
	ctx := e.NewContext(req, resp)
	err := handler(ctx)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
}
