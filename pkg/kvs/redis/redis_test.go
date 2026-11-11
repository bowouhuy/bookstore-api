package redis

import (
	"context"
	"fmt"
	"testing"

	goredis "github.com/redis/go-redis/v9"
)

// requireRedis skips the test when no Redis server is reachable on the default address.
func requireRedis(t *testing.T) {
	t.Helper()

	c := goredis.NewClient(&goredis.Options{Addr: "127.0.0.1:6379"})
	defer c.Close()

	if err := c.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis not available: %v", err)
	}
}

func TestRedisClient_Set(t *testing.T) {
	requireRedis(t)

	ctx := context.Background()
	redis, err := New()
	if err != nil {
		t.Fatal(err)
	}

	defer redis.Close()

	_, err = redis.Set(ctx, "key", "value", 0)
	if err != nil {
		t.Fatal(err)
	}

	val, err := redis.Get(ctx, "key")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(fmt.Sprintf("success get value: %s", val))

	redis.FlushAll(ctx)
}

func TestRedisClient_HSet(t *testing.T) {
	requireRedis(t)

	ctx := context.Background()
	redis, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer redis.Close()

	_, err = redis.HSet(ctx, "key", "field", "value")
	if err != nil {
		t.Fatal(err)
	}

	_, err = redis.HSet(ctx, "key", "field2", "value2")
	if err != nil {
		t.Fatal(err)
	}

	val, err := redis.HGetAll(ctx, "key")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(fmt.Sprintf("success get value: %s", val))

	redis.FlushAll(ctx)
}
