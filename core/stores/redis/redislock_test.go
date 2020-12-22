package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tal-tech/go-zero/core/stringx"
)

func TestRedisLock(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		key := stringx.Rand()
		firstLock := NewRedisLock(client, key)
		firstLock.SetExpire(5)
		firstAcquire, err := firstLock.Acquire(ctx)
		assert.Nil(t, err)
		assert.True(t, firstAcquire)

		secondLock := NewRedisLock(client, key)
		secondLock.SetExpire(5)
		againAcquire, err := secondLock.Acquire(ctx)
		assert.Nil(t, err)
		assert.False(t, againAcquire)

		release, err := firstLock.Release(ctx)
		assert.Nil(t, err)
		assert.True(t, release)

		endAcquire, err := secondLock.Acquire(ctx)
		assert.Nil(t, err)
		assert.True(t, endAcquire)
	})
}
