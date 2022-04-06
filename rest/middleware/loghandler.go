package middleware

import (
	"bytes"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/valeamoris/go-ezio/rest/internal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/timex"
	"github.com/zeromicro/go-zero/core/utils"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

const slowThreshold = time.Millisecond * 500

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
	duration := timer.Duration()
	buf.WriteString(fmt.Sprintf("%d - %s - %s - %s - %s - %s",
		ctx.Response().Status, ctx.Request().Method, ctx.Request().RequestURI, ctx.RealIP(), ctx.Request().UserAgent(),
		timex.ReprOfDuration(duration)))
	if duration > slowThreshold {
		logx.Slowf("[HTTP] %d - %s - %s - %s - %s - slowcall(%s)",
			ctx.Response().Status, ctx.Request().Method, ctx.Request().RequestURI, ctx.RealIP(), ctx.Request().UserAgent(),
			timex.ReprOfDuration(duration))
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
		logx.Info(buf.String())
	} else {
		logx.Error(buf.String())
	}
}

func logDetails(ctx echo.Context, pool *sync.Pool, timer *utils.ElapsedTimer, logs *internal.LogCollector) {
	buf := pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer pool.Put(buf)

	duration := timer.Duration()
	buf.WriteString(fmt.Sprintf("%d - %s - %s - %s\n=> %s\n",
		ctx.Response().Status, ctx.Request().Method, ctx.RealIP(), timex.ReprOfDuration(duration), dumpRequest(ctx.Request())))
	if duration > slowThreshold {
		logx.Slowf("[HTTP] %d - %s - %s - slowcall(%s)\n=> %s\n",
			ctx.Response().Status, ctx.Request().Method, ctx.RealIP(), timex.ReprOfDuration(duration), dumpRequest(ctx.Request()))
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
		logx.Info(buf.String())
	} else {
		logx.Error(buf.String())
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
