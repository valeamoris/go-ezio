package redis

import (
	"github.com/tal-tech/go-zero/core/breaker"
	"io"

	red "github.com/go-redis/redis/v8"
	"github.com/tal-tech/go-zero/core/syncx"
)

var clusterManager = syncx.NewResourceManager()

func getCluster(server, pass string, brk breaker.Breaker) (*red.ClusterClient, error) {
	val, err := clusterManager.GetResource(server, func() (io.Closer, error) {
		store := red.NewClusterClient(&red.ClusterOptions{
			Addrs:        []string{server},
			Password:     pass,
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

	return val.(*red.ClusterClient), nil
}
