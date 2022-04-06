package internal

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"sync"
)

var LogContext = contextKey("request_logs")

type LogCollector struct {
	Messages []string
	lock     sync.Mutex
}

func (lc *LogCollector) Append(msg string) {
	lc.lock.Lock()
	lc.Messages = append(lc.Messages, msg)
	lc.lock.Unlock()
}

func (lc *LogCollector) takeAll() []string {
	lc.lock.Lock()
	messages := lc.Messages
	lc.Messages = nil
	lc.lock.Unlock()

	return messages
}

func (lc *LogCollector) Flush() string {
	var buffer bytes.Buffer

	start := true
	for _, message := range lc.takeAll() {
		if start {
			start = false
		} else {
			buffer.WriteByte('\n')
		}
		buffer.WriteString(message)
	}

	return buffer.String()
}

func Error(ctx echo.Context, v ...interface{}) {
	logx.ErrorCaller(1, format(ctx, v...))
}

func Errorf(ctx echo.Context, format string, v ...interface{}) {
	logx.ErrorCaller(1, formatf(ctx, format, v...))
}

func Info(ctx echo.Context, v ...interface{}) {
	appendLog(ctx, format(ctx, v...))
}

func Infof(ctx echo.Context, format string, v ...interface{}) {
	appendLog(ctx, formatf(ctx, format, v...))
}

func appendLog(ctx echo.Context, message string) {
	logs := ctx.Request().Context().Value(LogContext)
	if logs != nil {
		logs.(*LogCollector).Append(message)
	}
}

func format(ctx echo.Context, v ...interface{}) string {
	return formatWithCtx(ctx, fmt.Sprint(v...))
}

func formatf(ctx echo.Context, format string, v ...interface{}) string {
	return formatWithCtx(ctx, fmt.Sprintf(format, v...))
}

func formatWithCtx(ctx echo.Context, v string) string {
	return fmt.Sprintf("(%s - %s) %s", ctx.Request().RequestURI, ctx.RealIP(), v)
}

type contextKey string

func (c contextKey) String() string {
	return "rest/internal context key " + string(c)
}
