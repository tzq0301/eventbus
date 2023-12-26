package eventbus

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	rediscontainers "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestRedisGroupCreate(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.TODO()

	redisContainer := must(rediscontainers.RunContainer(ctx, testcontainers.WithImage("redis:7")))

	rdb := redis.NewClient(must(redis.ParseURL(must(redisContainer.ConnectionString(ctx)))))

	eventbus := New(NewRedisHandler(
		WithRedisRDB(rdb),
		WithRedisStreamPrefix("event:"),
		WithRedisPullInterval(100*time.Millisecond),
		WithRedisPullTimeout(2*time.Second),
		WithRedisCountPerPull(3),
		WithRedisErrHandler(func(err error) {
			panic(err)
		}),
	))

	source := "source"
	require.NoError(t, eventbus.Subscribe(ctx, source, nil))
	require.NoError(t, eventbus.Subscribe(ctx, source, nil))
}

func TestRedisPublish(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.TODO()

	redisContainer := must(rediscontainers.RunContainer(ctx, testcontainers.WithImage("redis:7")))

	rdb := redis.NewClient(must(redis.ParseURL(must(redisContainer.ConnectionString(ctx)))))

	eventbus := New(NewRedisHandler(
		WithRedisRDB(rdb),
		WithRedisPullInterval(100*time.Millisecond),
		WithRedisPullTimeout(2*time.Second),
		WithRedisCountPerPull(3),
		WithRedisErrHandler(func(err error) {
			panic(err)
		}),
	))

	source := "source"
	require.NoError(t, eventbus.Publish(ctx, source, nil))
}

func TestRedisHandler(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.TODO()

	redisContainer := must(rediscontainers.RunContainer(ctx, testcontainers.WithImage("redis:7")))

	rdb := redis.NewClient(must(redis.ParseURL(must(redisContainer.ConnectionString(ctx)))))

	eventbus := New(NewRedisHandler(
		WithRedisRDB(rdb),
		WithRedisPullInterval(10*time.Millisecond),
		WithRedisPullTimeout(2*time.Second),
		WithRedisCountPerPull(3),
		WithRedisErrHandler(func(err error) {
			panic(err)
		}),
	))

	source := "source"

	type SubData struct {
		Name string
	}

	type Data struct {
		ID      int
		SubData *SubData
	}

	total := 1000

	for i := 0; i < total; i++ {
		require.NoError(t, eventbus.Publish(ctx, source, Data{
			ID: i,
			SubData: &SubData{
				Name: fmt.Sprintf("name-%d", i),
			},
		}))
	}

	var wg sync.WaitGroup
	wg.Add(total * 2)

	marked := make([]bool, total*2)
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		eventbus := New(NewRedisHandler(
			WithRedisRDB(rdb),
			WithRedisPullInterval(10*time.Millisecond),
			WithRedisPullTimeout(2*time.Second),
			WithRedisCountPerPull(30),
			WithRedisErrHandler(func(err error) {
				panic(err)
			}),
		))

		require.NoError(t, eventbus.Subscribe(ctx, source, func(_ Source, payload Payload) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			var data Data
			require.NoError(t, mapstructure.Decode(payload, &data))
			marked[data.ID] = true
		}))
	}

	for i := total; i < total*2; i++ {
		require.NoError(t, eventbus.Publish(ctx, source, Data{
			ID: i,
			SubData: &SubData{
				Name: fmt.Sprintf("name-%d", i),
			},
		}))
	}

	wg.Wait()

	var counter int64
	for _, b := range marked {
		if b {
			counter++
		}
	}

	require.EqualValues(t, total*2, counter)
}

func TestRedisPubInSub(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	redisContainer := must(rediscontainers.RunContainer(ctx, testcontainers.WithImage("redis:7")))
	rdb := redis.NewClient(must(redis.ParseURL(must(redisContainer.ConnectionString(ctx)))))
	eventbus := New(NewRedisHandler(WithRedisRDB(rdb)))

	outerSource := "outer"
	innerSource := "inner"

	_ = eventbus.Subscribe(ctx, innerSource, func(_ Source, _ Payload) {
		cancel()
	})

	_ = eventbus.Subscribe(ctx, outerSource, func(_ Source, _ Payload) {
		_ = eventbus.Publish(ctx, innerSource, nil)
	})

	_ = eventbus.Publish(ctx, outerSource, nil)

	select {
	case <-ctx.Done():
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
