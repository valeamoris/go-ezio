package redis

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	red "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestRedis_Exists(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		cnt, err := client.Exists(ctx, "a").Result()
		ok := cnt != 0
		assert.Nil(t, err)
		assert.False(t, ok)
		_, err = client.Set(ctx, "a", "b", 0).Result()
		assert.Nil(t, err)
		cnt, err = client.Exists(ctx, "a").Result()
		ok = cnt != 0
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestRedis_Eval(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		_, err := client.Eval(ctx, `redis.call("EXISTS", KEYS[1])`, []string{"notexist"}).Result()
		assert.Equal(t, Nil, err)
		err = client.Set(ctx, "key1", "value1", 0).Err()
		assert.Nil(t, err)
		_, err = client.Eval(ctx, `redis.call("EXISTS", KEYS[1])`, []string{"key1"}).Result()
		assert.Equal(t, Nil, err)
		val, err := client.Eval(ctx, `return redis.call("EXISTS", KEYS[1])`, []string{"key1"}).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
	})
}

func TestRedis_Hgetall(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		assert.Nil(t, client.HSet(ctx, "a", "aa", "aaa").Err())
		assert.Nil(t, client.HSet(ctx, "a", "bb", "bbb").Err())
		vals, err := client.HGetAll(ctx, "a").Result()
		assert.Nil(t, err)
		assert.EqualValues(t, map[string]string{
			"aa": "aaa",
			"bb": "bbb",
		}, vals)
	})
}

func TestRedis_Hvals(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		assert.Nil(t, client.HSet(ctx, "a", "aa", "aaa").Err())
		assert.Nil(t, client.HSet(ctx, "a", "bb", "bbb").Err())
		vals, err := client.HVals(ctx, "a").Result()
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aaa", "bbb"}, vals)
	})
}

func TestRedis_Hsetnx(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		assert.Nil(t, client.HSet(ctx, "a", "aa", "aaa").Err())
		assert.Nil(t, client.HSet(ctx, "a", "bb", "bbb").Err())
		ok, err := client.HSetNX(ctx, "a", "bb", "ccc").Result()
		assert.Nil(t, err)
		assert.False(t, ok)
		ok, err = client.HSetNX(ctx, "a", "dd", "ddd").Result()
		assert.Nil(t, err)
		assert.True(t, ok)
		vals, err := client.HVals(ctx, "a").Result()
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aaa", "bbb", "ddd"}, vals)
	})
}

func TestRedis_HdelHlen(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		assert.Nil(t, client.HSet(ctx, "a", "aa", "aaa").Err())
		assert.Nil(t, client.HSet(ctx, "a", "bb", "bbb").Err())
		num, err := client.HLen(ctx, "a").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), num)
		val, err := client.HDel(ctx, "a", "aa").Result()
		assert.Nil(t, err)
		assert.True(t, val == 1)
		vals, err := client.HVals(ctx, "a").Result()
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"bbb"}, vals)
	})
}

func TestRedis_HIncrBy(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		val, err := client.HIncrBy(ctx, "key", "field", 2).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
		val, err = client.HIncrBy(ctx, "key", "field", 3).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(5), val)
	})
}

func TestRedis_Hkeys(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		assert.Nil(t, client.HSet(ctx, "a", "aa", "aaa").Err())
		assert.Nil(t, client.HSet(ctx, "a", "bb", "bbb").Err())
		vals, err := client.HKeys(ctx, "a").Result()
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aa", "bb"}, vals)
	})
}

func TestRedis_Hmget(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		assert.Nil(t, client.HSet(ctx, "a", "aa", "aaa").Err())
		assert.Nil(t, client.HSet(ctx, "a", "bb", "bbb").Err())
		vals, err := client.HMGet(ctx, "a", "aa", "bb").Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []interface{}{"aaa", "bbb"}, vals)
		vals, err = client.HMGet(ctx, "a", "aa", "no", "bb").Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []interface{}{"aaa", nil, "bbb"}, vals)
	})
}

func TestRedis_Hmset(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		values := make([]interface{}, 0)
		m := map[string]string{
			"aa": "aaa",
			"bb": "bbb",
		}
		for k, v := range m {
			values = append(values, []interface{}{k, v}...)
		}
		assert.Nil(t, client.HMSet(ctx, "a", values...).Err())
		vals, err := client.HMGet(ctx, "a", "aa", "bb").Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []interface{}{"aaa", "bbb"}, vals)
	})
}

func TestRedis_Incr(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		val, err := client.Incr(ctx, "a").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
		val, err = client.Incr(ctx, "a").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
	})
}

func TestRedis_IncrBy(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		val, err := client.IncrBy(ctx, "a", 2).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
		val, err = client.IncrBy(ctx, "a", 3).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(5), val)
	})
}

func TestRedis_Keys(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		err := client.Set(ctx, "key1", "value1", 0).Err()
		assert.Nil(t, err)
		err = client.Set(ctx, "key2", "value2", 0).Err()
		assert.Nil(t, err)
		keys, err := client.Keys(ctx, "*").Result()
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys)
	})
}

func TestRedis_HyperLogLog(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		client.Ping(ctx)

		_, err := client.PFAdd(ctx, "key1").Result()
		assert.NotNil(t, err)
		_, err = client.PFCount(ctx, "*").Result()
		assert.NotNil(t, err)
		err = client.PFMerge(ctx, "*").Err()
		assert.NotNil(t, err)
	})
}

func TestRedis_List(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		val, err := client.LPush(ctx, "key", "value1", "value2").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
		val, err = client.RPush(ctx, "key", "value3", "value4").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(4), val)
		val, err = client.LLen(ctx, "key").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(4), val)
		vals, err := client.LRange(ctx, "key", 0, 10).Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value2", "value1", "value3", "value4"}, vals)
		v, err := client.LPop(ctx, "key").Result()
		assert.Nil(t, err)
		assert.Equal(t, "value2", v)
		val, err = client.LPush(ctx, "key", "value1", "value2").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(5), val)
		val, err = client.RPush(ctx, "key", "value3", "value3").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(7), val)
		n, err := client.LRem(ctx, "key", 2, "value1").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), n)
		vals, err = client.LRange(ctx, "key", 0, 10).Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value2", "value3", "value4", "value3", "value3"}, vals)
		n, err = client.LRem(ctx, "key", -2, "value3").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), n)
		vals, err = client.LRange(ctx, "key", 0, 10).Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value2", "value3", "value4"}, vals)
	})
}

func TestRedis_Mget(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		err := client.Set(ctx, "key1", "value1", 0).Err()
		assert.Nil(t, err)
		err = client.Set(ctx, "key2", "value2", 0).Err()
		assert.Nil(t, err)
		vals, err := client.MGet(ctx, "key1", "key0", "key2", "key3").Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []interface{}{"value1", nil, "value2", nil}, vals)
	})
}

func TestRedis_SetBit(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		err := client.SetBit(ctx, "key", 1, 1).Err()
		assert.Nil(t, err)
	})
}

func TestRedis_GetBit(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		err := client.SetBit(ctx, "key", 2, 1).Err()
		assert.Nil(t, err)
		val, err := client.GetBit(ctx, "key", 2).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
	})
}

func TestRedis_Persist(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		ok, err := client.Persist(ctx, "key").Result()
		assert.Nil(t, err)
		assert.False(t, ok)
		err = client.Set(ctx, "key", "value", 0).Err()
		assert.Nil(t, err)
		ok, err = client.Persist(ctx, "key").Result()
		assert.Nil(t, err)
		assert.False(t, ok)
		err = client.Expire(ctx, "key", 5*time.Second).Err()
		assert.Nil(t, err)
		ok, err = client.Persist(ctx, "key").Result()
		assert.Nil(t, err)
		assert.True(t, ok)
		err = client.ExpireAt(ctx, "key", time.Now().Add(5*time.Second)).Err()
		assert.Nil(t, err)
		ok, err = client.Persist(ctx, "key").Result()
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestRedis_Ping(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		reply, err := client.Ping(ctx).Result()
		assert.NoError(t, err)
		assert.True(t, reply == "PONG")
	})
}

func TestRedis_Scan(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		err := client.Set(ctx, "key1", "value1", 0).Err()
		assert.Nil(t, err)
		err = client.Set(ctx, "key2", "value2", 0).Err()
		assert.Nil(t, err)
		keys, _, err := client.Scan(ctx, 0, "*", 100).Result()
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys)
	})
}

func TestRedis_Sscan(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		key := "list"
		var list []string
		for i := 0; i < 1550; i++ {
			list = append(list, randomStr(i))
		}
		lens, err := client.SAdd(ctx, key, list).Result()
		assert.Nil(t, err)
		assert.Equal(t, lens, int64(1550))

		var cursor uint64 = 0
		sum := 0
		for {
			keys, next, err := client.SScan(ctx, key, cursor, "", 100).Result()
			assert.Nil(t, err)
			sum += len(keys)
			if next == 0 {
				break
			}
			cursor = next
		}

		assert.Equal(t, sum, 1550)
		_, err = client.Del(ctx, key).Result()
		assert.Nil(t, err)
	})
}

func TestRedis_Set(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		num, err := client.SAdd(ctx, "key", 1, 2, 3, 4).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(4), num)
		val, err := client.SCard(ctx, "key").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(4), val)
		ok, err := client.SIsMember(ctx, "key", 2).Result()
		assert.Nil(t, err)
		assert.True(t, ok)
		num, err = client.SRem(ctx, "key", 3, 4).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), num)
		vals, err := client.SMembers(ctx, "key").Result()
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"1", "2"}, vals)
		members, err := client.SRandMemberN(ctx, "key", 1).Result()
		assert.Nil(t, err)
		assert.Len(t, members, 1)
		assert.Contains(t, []string{"1", "2"}, members[0])
		member, err := client.SPop(ctx, "key").Result()
		assert.Nil(t, err)
		assert.Contains(t, []string{"1", "2"}, member)
		vals, err = client.SMembers(ctx, "key").Result()
		assert.Nil(t, err)
		assert.NotContains(t, vals, member)
		num, err = client.SAdd(ctx, "key1", 1, 2, 3, 4).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(4), num)
		num, err = client.SAdd(ctx, "key2", 2, 3, 4, 5).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(4), num)
		vals, err = client.SUnion(ctx, "key1", "key2").Result()
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"1", "2", "3", "4", "5"}, vals)
		num, err = client.SUnionStore(ctx, "key3", "key1", "key2").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(5), num)
		vals, err = client.SDiff(ctx, "key1", "key2").Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"1"}, vals)
		num, err = client.SDiffStore(ctx, "key4", "key1", "key2").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), num)
	})
}

func TestRedis_SetGetDel(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		err := client.Set(ctx, "hello", "world", 0).Err()
		assert.Nil(t, err)
		val, err := client.Get(ctx, "hello").Result()
		assert.Nil(t, err)
		assert.Equal(t, "world", val)
		ret, err := client.Del(ctx, "hello").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), ret)
	})
}

func TestRedis_SetExNx(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		err := client.Set(ctx, "hello", "world", time.Second*5).Err()
		assert.Nil(t, err)
		ok, err := client.SetNX(ctx, "hello", "newworld", 0).Result()
		assert.Nil(t, err)
		assert.False(t, ok)
		ok, err = client.SetNX(ctx, "newhello", "newworld", 0).Result()
		assert.Nil(t, err)
		assert.True(t, ok)
		val, err := client.Get(ctx, "hello").Result()
		assert.Nil(t, err)
		assert.Equal(t, "world", val)
		val, err = client.Get(ctx, "newhello").Result()
		assert.Nil(t, err)
		assert.Equal(t, "newworld", val)
		ttl, err := client.TTL(ctx, "hello").Result()
		assert.Nil(t, err)
		assert.True(t, ttl > 0)
		ok, err = client.SetNX(ctx, "newhello", "newworld", 5*time.Second).Result()
		assert.Nil(t, err)
		assert.False(t, ok)
		num, err := client.Del(ctx, "newhello").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), num)
		ok, err = client.SetNX(ctx, "newhello", "newworld", 5*time.Second).Result()
		assert.Nil(t, err)
		assert.True(t, ok)
		val, err = client.Get(ctx, "newhello").Result()
		assert.Nil(t, err)
		assert.Equal(t, "newworld", val)
	})
}

func TestRedis_SetGetDelHashField(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		err := client.HSet(ctx, "key", "field", "value").Err()
		assert.Nil(t, err)
		val, err := client.HGet(ctx, "key", "field").Result()
		assert.Nil(t, err)
		assert.Equal(t, "value", val)
		ok, err := client.HExists(ctx, "key", "field").Result()
		assert.Nil(t, err)
		assert.True(t, ok)
		ret, err := client.HDel(ctx, "key", "field").Result()
		assert.Nil(t, err)
		assert.True(t, ret == 1)
		ok, err = client.HExists(ctx, "key", "field").Result()
		assert.Nil(t, err)
		assert.False(t, ok)
	})
}

func TestRedis_SortedSet(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		zval, err := client.ZAdd(ctx, "key", &Z{Score: 1, Member: "value1"}).Result()
		assert.Nil(t, err)
		assert.True(t, zval == 1)
		zval, err = client.ZAdd(ctx, "key", &Z{Score: 2, Member: "value1"}).Result()
		assert.Nil(t, err)
		assert.False(t, zval == 1)
		val, err := client.ZScore(ctx, "key", "value1").Result()
		assert.Nil(t, err)
		assert.Equal(t, float64(2), val)
		val, err = client.ZIncrBy(ctx, "key", 3, "value1").Result()
		assert.Nil(t, err)
		assert.Equal(t, float64(5), val)
		val, err = client.ZScore(ctx, "key", "value1").Result()
		assert.Nil(t, err)
		assert.Equal(t, float64(5), val)
		zval, err = client.ZAdd(ctx, "key", &Z{
			Member: "value2",
			Score:  6,
		}, &Z{
			Member: "value3",
			Score:  7,
		}).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), zval)
		pairs, err := client.ZRevRangeWithScores(ctx, "key", 1, 3).Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []Z{
			{
				Member: "value2",
				Score:  6,
			},
			{
				Member: "value1",
				Score:  5,
			},
		}, pairs)
		rank, err := client.ZRank(ctx, "key", "value2").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), rank)
		rank, err = client.ZRevRank(ctx, "key", "value1").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), rank)
		_, err = client.ZRank(ctx, "key", "value4").Result()
		assert.Equal(t, Nil, err)
		num, err := client.ZRem(ctx, "key", "value2", "value3").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), num)
		zval, err = client.ZAdd(ctx, "key", &Z{Score: 6, Member: "value2"}).Result()
		assert.Nil(t, err)
		assert.True(t, zval == 1)
		zval, err = client.ZAdd(ctx, "key", &Z{Score: 7, Member: "value3"}).Result()
		assert.Nil(t, err)
		assert.True(t, zval == 1)
		zval, err = client.ZAdd(ctx, "key", &Z{Score: 8, Member: "value4"}).Result()
		assert.Nil(t, err)
		assert.True(t, zval == 1)
		num, err = client.ZRemRangeByScore(ctx, "key", "6", "7").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), num)
		zval, err = client.ZAdd(ctx, "key", &Z{Score: 6, Member: "value2"}).Result()
		assert.Nil(t, err)
		assert.True(t, zval == 1)
		zval, err = client.ZAdd(ctx, "key", &Z{Score: 7, Member: "value3"}).Result()
		assert.Nil(t, err)
		assert.True(t, zval == 1)
		num, err = client.ZCount(ctx, "key", "6", "7").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), num)
		num, err = client.ZRemRangeByRank(ctx, "key", 1, 2).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), num)
		card, err := client.ZCard(ctx, "key").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), card)
		vals, err := client.ZRange(ctx, "key", 0, -1).Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value1", "value4"}, vals)
		vals, err = client.ZRevRange(ctx, "key", 0, -1).Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value4", "value1"}, vals)
		pairs, err = client.ZRangeWithScores(ctx, "key", 0, -1).Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []Z{
			{
				Member: "value1",
				Score:  5,
			},
			{
				Member: "value4",
				Score:  8,
			},
		}, pairs)
		pairs, err = client.ZRangeByScoreWithScores(ctx, "key", &red.ZRangeBy{
			Min: "5",
			Max: "8",
		}).Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []Z{
			{
				Member: "value1",
				Score:  5,
			},
			{
				Member: "value4",
				Score:  8,
			},
		}, pairs)
		pairs, err = client.ZRangeByScoreWithScores(ctx, "key", &red.ZRangeBy{
			Min:    "5",
			Max:    "8",
			Offset: 1,
			Count:  1,
		}).Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []Z{
			{
				Member: "value4",
				Score:  8,
			},
		}, pairs)
		pairs, err = client.ZRangeByScoreWithScores(ctx, "key", &red.ZRangeBy{
			Min:    "5",
			Max:    "8",
			Offset: 1,
			Count:  0,
		}).Result()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(pairs))
		pairs, err = client.ZRevRangeByScoreWithScores(ctx, "key", &red.ZRangeBy{
			Min: "5",
			Max: "8",
		}).Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []Z{
			{
				Member: "value4",
				Score:  8,
			},
			{
				Member: "value1",
				Score:  5,
			},
		}, pairs)
		pairs, err = client.ZRevRangeByScoreWithScores(ctx, "key", &red.ZRangeBy{
			Min:    "5",
			Max:    "8",
			Offset: 1,
			Count:  1,
		}).Result()
		assert.Nil(t, err)
		assert.EqualValues(t, []Z{
			{
				Member: "value1",
				Score:  5,
			},
		}, pairs)
		pairs, err = client.ZRevRangeByScoreWithScores(ctx, "key", &red.ZRangeBy{
			Min:    "5",
			Max:    "8",
			Offset: 1,
			Count:  0,
		}).Result()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(pairs))
	})
}

func TestRedis_Pipelined(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		_, err := client.Pipelined(ctx,
			func(pipe Pipeliner) error {
				pipe.Incr(ctx, "pipelined_counter")
				pipe.Expire(ctx, "pipelined_counter", time.Hour)
				pipe.ZAdd(ctx, "zadd", &Z{Score: 12, Member: "zadd"})
				return nil
			},
		)
		assert.Nil(t, err)
		ttl, err := client.TTL(ctx, "pipelined_counter").Result()
		assert.Nil(t, err)
		assert.Equal(t, time.Second*3600, ttl)
		value, err := client.Get(ctx, "pipelined_counter").Result()
		assert.Nil(t, err)
		assert.Equal(t, "1", value)
		score, err := client.ZScore(ctx, "zadd", "zadd").Result()
		assert.Nil(t, err)
		assert.Equal(t, float64(12), score)
	})
}

func TestRedisScriptLoad(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		client.Ping(ctx)
		_, err := client.ScriptLoad(ctx, "foo").Result()
		assert.NotNil(t, err)
	})
}

func TestRedisToStrings(t *testing.T) {
	vals := toStrings([]interface{}{1, 2})
	assert.EqualValues(t, []string{"1", "2"}, vals)
}

func TestRedisBlpop(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		client.Ping(ctx)
		_, err := client.BLPop(ctx, blockingQueryTimeout, "foo").Result()
		assert.NotNil(t, err)
	})
}

func TestRedisGeo(t *testing.T) {
	runOnRedis(t, func(ctx context.Context, client Node) {
		client.Ping(ctx)
		var geoLocation = []*GeoLocation{{Longitude: 13.361389, Latitude: 38.115556, Name: "Palermo"}, {Longitude: 15.087269, Latitude: 37.502669, Name: "Catania"}}
		v, err := client.GeoAdd(ctx, "sicily", geoLocation...).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), v)
		v2, err := client.GeoDist(ctx, "sicily", "Palermo", "Catania", "m").Result()
		assert.Nil(t, err)
		assert.Equal(t, 166274, int(v2))
		// GeoHash not support
		v3, err := client.GeoPos(ctx, "sicily", "Palermo", "Catania").Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(v3[0].Longitude), int64(13))
		assert.Equal(t, int64(v3[0].Latitude), int64(38))
		assert.Equal(t, int64(v3[1].Longitude), int64(15))
		assert.Equal(t, int64(v3[1].Latitude), int64(37))
		v4, err := client.GeoRadius(ctx, "sicily", 15, 37, &red.GeoRadiusQuery{WithDist: true, Unit: "km", Radius: 200}).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(v4[0].Dist), int64(190))
		assert.Equal(t, int64(v4[1].Dist), int64(56))
		var geoLocation2 = []*GeoLocation{{Longitude: 13.583333, Latitude: 37.316667, Name: "Agrigento"}}
		v5, err := client.GeoAdd(ctx, "sicily", geoLocation2...).Result()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), v5)
		v6, err := client.GeoRadiusByMember(ctx, "sicily", "Agrigento", &red.GeoRadiusQuery{Unit: "km", Radius: 100}).Result()
		assert.Nil(t, err)
		assert.Equal(t, v6[0].Name, "Agrigento")
		assert.Equal(t, v6[1].Name, "Palermo")
	})
}

func runOnRedis(t *testing.T, fn func(ctx context.Context, client Node)) {
	s, err := miniredis.Run()
	assert.Nil(t, err)
	defer func() {
		client, err := clientManager.GetResource(s.Addr(), func() (io.Closer, error) {
			return nil, errors.New("should already exist")
		})
		if err != nil {
			t.Error(err)
		}

		if client != nil {
			_ = client.Close()
		}
	}()

	ctx := context.TODO()
	node, err := NewRedis(s.Addr(), NodeType)
	if err != nil {
		t.Error(err)
		return
	}
	fn(ctx, node)
}

type mockedNode struct {
	Node
}

func (n mockedNode) BLPop(ctx context.Context, timeout time.Duration, keys ...string) *red.StringSliceCmd {
	return red.NewStringSliceCmd(context.TODO(), "foo", "bar")
}
