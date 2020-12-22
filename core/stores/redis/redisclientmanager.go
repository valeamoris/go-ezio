package redis

import (
	"github.com/tal-tech/go-zero/core/breaker"
	"io"

	red "github.com/go-redis/redis/v8"
	"github.com/tal-tech/go-zero/core/syncx"
)

const (
	defaultDatabase = 0
	maxRetries      = 3
	idleConns       = 8
)

var clientManager = syncx.NewResourceManager()

func getClient(server, pass string, brk breaker.Breaker) (*red.Client, error) {
	val, err := clientManager.GetResource(server, func() (io.Closer, error) {
		store := red.NewClient(&red.Options{
			Addr:         server,
			Password:     pass,
			DB:           defaultDatabase,
			MaxRetries:   maxRetries,
			MinIdleConns: idleConns,
		})
		store.AddHook(&hook{
			brk: brk,
		})
		return store, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(*red.Client), nil
}
