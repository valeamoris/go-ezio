package middleware

import (
	"bytes"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/timex"
	"github.com/tal-tech/go-zero/core/trace/tracespec"
	"github.com/tal-tech/go-zero/core/utils"
	"github.com/uber/jaeger-client-go"
	"github.com/valeamoris/go-ezio/rest/internal"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

const slowThreshold = time.Millisecond * 500

type jaegerTracer struct {
	ctx jaeger.SpanContext
}

func newJaegerTracer(ctx jaeger.SpanContext) *jaegerTracer {
	return &jaegerTracer{ctx: ctx}
}

func (j *jaegerTracer) TraceId() string {
	return j.ctx.SpanID().String()
}

func (j *jaegerTracer) SpanId() string {
	return j.ctx.TraceID().String()
}

func (j *jaegerTracer) Visit(fn func(key string, val string) bool) {
	return
}

func (j *jaegerTracer) Finish() {
	return
}

func (j *jaegerTracer) Fork(ctx context.Context, serviceName, operationName string) (context.Context, tracespec.Trace) {
	return ctx, j
}

func (j *jaegerTracer) Follow(ctx context.Context, serviceName, operationName string) (context.Context, tracespec.Trace) {
	return ctx, j
}

type emptyTracer struct{}

func (e emptyTracer) TraceId() string {
	return ""
}

func (e emptyTracer) SpanId() string {
	return ""
}

func (e emptyTracer) Visit(fn func(key string, val string) bool) {
	return
}

func (e emptyTracer) Finish() {
	return
}

func (e emptyTracer) Fork(ctx context.Context, serviceName, operationName string) (context.Context, tracespec.Trace) {
	return ctx, e
}

func (e emptyTracer) Follow(ctx context.Context, serviceName, operationName string) (context.Context, tracespec.Trace) {
	return ctx, e
}

func newEmptyTracer() *emptyTracer {
	return &emptyTracer{}
}

func LogMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	pool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}

	return func(ctx echo.Context) error {
		timer := utils.NewElapsedTimer()
		logs := new(internal.LogCollector)
		ctx.SetRequest(ctx.Request().WithContext(context.WithValue(ctx.Request().Context(), internal.LogContext, logs)))
		err := next(ctx)
		logBrief(ctx, pool, timer, logs)
		return err
	}
}

func DetailedLogMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	pool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}
	return func(ctx echo.Context) error {
		timer := utils.NewElapsedTimer()
		logs := new(internal.LogCollector)
		ctx.SetRequest(ctx.Request().WithContext(context.WithValue(ctx.Request().Context(), internal.LogContext, logs)))
		err := next(ctx)
		logDetails(ctx, pool, timer, logs)
		return err
	}
}

func logBrief(ctx echo.Context, pool *sync.Pool, timer *utils.ElapsedTimer, logs *internal.LogCollector) {
	buf := pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer pool.Put(buf)

	// 需要注意的是trace中间件需要放到log前
	var nCtx context.Context
	oCtx := ctx.Request().Context()
	sp := opentracing.SpanFromContext(oCtx)
	if sp != nil {
		jaegerCtx := sp.Context().(jaeger.SpanContext)
		nCtx = context.WithValue(oCtx, tracespec.TracingKey, newJaegerTracer(jaegerCtx))
	} else {
		nCtx = context.WithValue(oCtx, tracespec.TracingKey, newEmptyTracer())
	}
	duration := timer.Duration()
	buf.WriteString(fmt.Sprintf("%d - %s - %s - %s - %s",
		ctx.Response().Status, ctx.Request().RequestURI, ctx.RealIP(), ctx.Request().UserAgent(), timex.ReprOfDuration(duration)))
	if duration > slowThreshold {
		logx.WithContext(nCtx).Slowf("[HTTP] %d - %s - %s - %s - slowcall(%s)",
			ctx.Response().Status, ctx.Request().RequestURI, ctx.RealIP(), ctx.Request().UserAgent(), timex.ReprOfDuration(duration))
	}

	ok := isOkResponse(ctx.Response().Status)
	if !ok {
		buf.WriteString(fmt.Sprintf("\n%s", dumpRequest(ctx.Request())))
	}
	body := logs.Flush()
	if len(body) > 0 {
		buf.WriteString(fmt.Sprintf("\n%s", body))
	}

	if ok {
		logx.WithContext(nCtx).Info(buf.String())
	} else {
		logx.WithContext(nCtx).Error(buf.String())
	}
}

func logDetails(ctx echo.Context, pool *sync.Pool, timer *utils.ElapsedTimer, logs *internal.LogCollector) {
	buf := pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer pool.Put(buf)

	var nCtx context.Context
	oCtx := ctx.Request().Context()
	sp := opentracing.SpanFromContext(oCtx)
	if sp != nil {
		jaegerCtx := sp.Context().(jaeger.SpanContext)
		nCtx = context.WithValue(oCtx, tracespec.TracingKey, newJaegerTracer(jaegerCtx))
	} else {
		nCtx = context.WithValue(oCtx, tracespec.TracingKey, newEmptyTracer())
	}
	duration := timer.Duration()
	buf.WriteString(fmt.Sprintf("%d - %s - %s\n=> %s\n",
		ctx.Response().Status, ctx.RealIP(), timex.ReprOfDuration(duration), dumpRequest(ctx.Request())))
	if duration > slowThreshold {
		logx.WithContext(nCtx).Slowf("[HTTP] %d - %s - slowcall(%s)\n=> %s\n",
			ctx.Response().Status, ctx.RealIP(), timex.ReprOfDuration(duration), dumpRequest(ctx.Request()))
	}

	ok := isOkResponse(ctx.Response().Status)
	if !ok {
		buf.WriteString(fmt.Sprintf("\n%s", dumpRequest(ctx.Request())))
	}
	body := logs.Flush()
	if len(body) > 0 {
		buf.WriteString(fmt.Sprintf("\n%s", body))
	}
	if ok {
		logx.WithContext(nCtx).Info(buf.String())
	} else {
		logx.WithContext(nCtx).Error(buf.String())
	}
}

func dumpRequest(r *http.Request) string {
	reqContent, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err.Error()
	} else {
		return string(reqContent)
	}
}

func isOkResponse(code int) bool {
	// not server error
	return code < http.StatusInternalServerError
}
