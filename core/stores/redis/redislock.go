package redis

import (
	"context"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"

	red "github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
)

// redis分布式锁

const (
	letters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lockCommand = `if redis.call("GET", KEYS[1]) == ARGV[1] then
    redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
    return "OK"
else
    return redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
end`
	delCommand = `if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end`
	randomLen       = 16
	tolerance       = 500 // milliseconds
	millisPerSecond = 1000
)

type Lock struct {
	store   Node
	seconds uint32
	key     string
	id      string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewRedisLock(store Node, key string) *Lock {
	return &Lock{
		store: store,
		key:   key,
		id:    randomStr(randomLen),
	}
}

func (rl *Lock) Acquire(ctx context.Context) (bool, error) {
	seconds := atomic.LoadUint32(&rl.seconds)
	resp, err := rl.store.Eval(ctx, lockCommand, []string{rl.key}, []string{
		rl.id, strconv.Itoa(int(seconds)*millisPerSecond + tolerance)}).Result()
	if err == red.Nil {
		return false, nil
	} else if err != nil {
		logx.Errorf("Error on acquiring lock for %s, %s", rl.key, err.Error())
		return false, err
	} else if resp == nil {
		return false, nil
	}

	reply, ok := resp.(string)
	if ok && reply == "OK" {
		return true, nil
	} else {
		logx.Errorf("Unknown reply when acquiring lock for %s: %v", rl.key, resp)
		return false, nil
	}
}

func (rl *Lock) Release(ctx context.Context) (bool, error) {
	resp, err := rl.store.Eval(ctx, delCommand, []string{rl.key}, []string{rl.id}).Result()
	if err != nil {
		return false, err
	}

	if reply, ok := resp.(int64); !ok {
		return false, nil
	} else {
		return reply == 1, nil
	}
}

func (rl *Lock) SetExpire(seconds int) {
	atomic.StoreUint32(&rl.seconds, uint32(seconds))
}

func randomStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
