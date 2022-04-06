package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	red "github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/breaker"
	"github.com/zeromicro/go-zero/core/mapping"
)

const (
	ClusterType = "cluster"
	NodeType    = "node"
	Nil         = red.Nil

	blockingQueryTimeout = 5 * time.Second
	readWriteTimeout     = 2 * time.Second

	slowThreshold = time.Millisecond * 100
)

var ErrNilNode = errors.New("nil redis node")

type (
	Pair struct {
		Key   string
		Score int64
	}

	// thread-safe
	Redis struct {
		Addr string
		Type string
		Pass string
		brk  breaker.Breaker
	}

	Node = red.UniversalClient

	// GeoLocation is used with GeoAdd to add geospatial location.
	GeoLocation = red.GeoLocation
	// GeoRadiusQuery is used with GeoRadius to query geospatial index.
	GeoRadiusQuery = red.GeoRadiusQuery
	GeoPos         = red.GeoPos

	Pipeliner = red.Pipeliner

	// Z represents sorted set member.
	Z        = red.Z
	FloatCmd = red.FloatCmd
)

func NewRedis(redisAddr, redisType string, redisPass ...string) (red.UniversalClient, error) {
	var pass string
	for _, v := range redisPass {
		pass = v
	}

	return getRedis(&Redis{
		Addr: redisAddr,
		Type: redisType,
		Pass: pass,
		brk:  breaker.NewBreaker(),
	})
}

func (s *Redis) String() string {
	return s.Addr
}

func (s *Redis) scriptLoad(ctx context.Context, script string) (string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return "", err
	}

	return conn.ScriptLoad(ctx, script).Result()
}

func acceptable(err error) bool {
	return err == nil || err == red.Nil
}

func getRedis(r *Redis) (red.UniversalClient, error) {
	switch r.Type {
	case ClusterType:
		return getCluster(r.Addr, r.Pass, r.brk)
	case NodeType:
		return getClient(r.Addr, r.Pass, r.brk)
	default:
		return nil, fmt.Errorf("redis type '%s' is not supported", r.Type)
	}
}

func toStrings(vals []interface{}) []string {
	ret := make([]string, len(vals))
	for i, val := range vals {
		if val == nil {
			ret[i] = ""
		} else {
			switch val := val.(type) {
			case string:
				ret[i] = val
			default:
				ret[i] = mapping.Repr(val)
			}
		}
	}
	return ret
}
