package redis

import (
	"context"
	"github.com/tal-tech/go-zero/core/breaker"
	"strings"
	"time"

	red "github.com/go-redis/redis/v8"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/mapping"
	"github.com/tal-tech/go-zero/core/timex"
)

const (
	redisContextKey = "redisContextKey"
)

type hook struct {
	brk breaker.Breaker
}

type hookContainer struct {
	start   time.Duration
	promise breaker.Promise
}

func (h hook) BeforeProcess(ctx context.Context, cmd red.Cmder) (context.Context, error) {
	p, err := h.brk.Allow()
	if err != nil {
		return ctx, err
	}
	c := &hookContainer{
		start:   timex.Now(),
		promise: p,
	}
	return context.WithValue(ctx, redisContextKey, c), nil
}

func (h hook) AfterProcess(ctx context.Context, cmd red.Cmder) error {
	c := ctx.Value(redisContextKey)
	if c == nil {
		return nil
	}
	container := c.(*hookContainer)
	go func() {
		duration := timex.Since(container.start)
		if duration > slowThreshold {
			var buf strings.Builder
			for i, arg := range cmd.Args() {
				if i > 0 {
					buf.WriteByte(' ')
				}
				buf.WriteString(mapping.Repr(arg))
			}
			logx.WithDuration(duration).Slowf("[REDIS] slowcall on executing: %s", buf.String())
		}
	}()
	err := cmd.Err()
	if acceptable(err) {
		container.promise.Accept()
	} else {
		container.promise.Reject(err.Error())
	}
	return nil
}

func (h hook) BeforeProcessPipeline(ctx context.Context, cmds []red.Cmder) (context.Context, error) {
	return ctx, nil
}

func (h hook) AfterProcessPipeline(ctx context.Context, cmds []red.Cmder) error {
	return nil
}
