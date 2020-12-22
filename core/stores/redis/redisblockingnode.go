package redis

import (
	"fmt"

	red "github.com/go-redis/redis/v8"
)

type ClosableNode interface {
	Node
	Close()
}

func CreateBlockingNode(r *Redis) (red.UniversalClient, error) {
	timeout := readWriteTimeout + blockingQueryTimeout

	switch r.Type {
	case NodeType:
		client := red.NewClient(&red.Options{
			Addr:         r.Addr,
			Password:     r.Pass,
			DB:           defaultDatabase,
			MaxRetries:   maxRetries,
			PoolSize:     1,
			MinIdleConns: 1,
			ReadTimeout:  timeout,
		})
		return client, nil
	case ClusterType:
		client := red.NewClusterClient(&red.ClusterOptions{
			Addrs:        []string{r.Addr},
			Password:     r.Pass,
			MaxRetries:   maxRetries,
			PoolSize:     1,
			MinIdleConns: 1,
			ReadTimeout:  timeout,
		})
		return client, nil
	default:
		return nil, fmt.Errorf("unknown redis type: %s", r.Type)
	}
}
