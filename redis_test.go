package eventbus

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	rediscontainers "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestRedisGroupCreate(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()

	redisContainer := must(rediscontainers.RunContainer(ctx, testcontainers.WithImage("redis:latest")))

	rdb := redis.NewClient(must(redis.ParseURL(must(redisContainer.ConnectionString(ctx)))))

	eventbus := New(NewRedisHandler(
		WithRedisRDB(rdb),
		WithRedisStreamPrefix("event:"),
		WithRedisPullInterval(100*time.Millisecond),
		WithRedisPullTimeout(2*time.Second),
		WithRedisCountPerPull(3),
		WithRedisErrHandler(func(err error) {
			require.NoError(t, err, err.Error())
		}),
	))

	event := "event"
	require.NoError(t, eventbus.Subscribe(ctx, event, func(event Event, payload Payload) {
	}))
	require.NoError(t, eventbus.Subscribe(ctx, event, func(event Event, payload Payload) {
	}))
}

func TestRedisPublish(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.TODO()

	redisContainer := must(rediscontainers.RunContainer(ctx, testcontainers.WithImage("redis:latest")))

	rdb := redis.NewClient(must(redis.ParseURL(must(redisContainer.ConnectionString(ctx)))))

	eventbus := New(NewRedisHandler(
		WithRedisRDB(rdb),
		WithRedisStreamPrefix("event:"),
		WithRedisPullInterval(100*time.Millisecond),
		WithRedisPullTimeout(2*time.Second),
		WithRedisCountPerPull(3),
		WithRedisErrHandler(func(err error) {
			require.NoErrorf(t, err, err.Error())
		}),
	))

	event := "event"
	require.NoError(t, eventbus.Publish(ctx, event, nil))
}

func TestRedisHandler(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.TODO()

	redisContainer := must(rediscontainers.RunContainer(ctx, testcontainers.WithImage("redis:latest")))

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

	event := "event"

	type SubData struct {
		Name string
	}

	type Data struct {
		ID      int
		SubData *SubData
	}

	total := 1000

	for i := 0; i < total; i++ {
		require.NoError(t, eventbus.Publish(ctx, event, Data{
			ID: i,
			SubData: &SubData{
				Name: fmt.Sprintf("name-%d", i),
			},
		}))
	}

	var wg sync.WaitGroup
	wg.Add(total * 2)

	hit := make(chan int, total*2)

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

		require.NoError(t, eventbus.Subscribe(ctx, event, func(_ Event, payload Payload) {
			defer wg.Done()
			var data Data
			require.NoError(t, mapstructure.Decode(payload, &data))
			hit <- data.ID
		}))
	}

	for i := total; i < total*2; i++ {
		require.NoError(t, eventbus.Publish(ctx, event, Data{
			ID: i,
			SubData: &SubData{
				Name: fmt.Sprintf("name-%d", i),
			},
		}))
	}

	wg.Wait()

	require.EqualValues(t, total*2, len(hit))
}

func TestRedisPubInSub(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	redisContainer := must(rediscontainers.RunContainer(ctx, testcontainers.WithImage("redis:latest")))
	rdb := redis.NewClient(must(redis.ParseURL(must(redisContainer.ConnectionString(ctx)))))
	eventbus := New(NewRedisHandler(WithRedisRDB(rdb), WithRedisErrHandler(func(err error) {
		require.NoError(t, err)
	})))

	outerEvent := "outer"
	innerEvent := "inner"

	_ = eventbus.Subscribe(ctx, innerEvent, func(_ Event, _ Payload) {
		cancel()
	})

	_ = eventbus.Subscribe(ctx, outerEvent, func(_ Event, _ Payload) {
		_ = eventbus.Publish(ctx, innerEvent, nil)
	})

	_ = eventbus.Publish(ctx, outerEvent, nil)

	select {
	case <-ctx.Done():
	}
}

func TestCase(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()
	redisContainer := must(rediscontainers.RunContainer(ctx, testcontainers.WithImage("redis:latest")))
	rdb := redis.NewClient(must(redis.ParseURL(must(redisContainer.ConnectionString(ctx)))))

	eventbus := New(NewRedisHandler(WithRedisRDB(rdb), WithRedisErrHandler(func(err error) {
		require.NoError(t, err)
	})))

	const (
		UserEvent  Event = "user"
		EmailEvent Event = "email"
		LarkEvent  Event = "lark"
	)

	type (
		UserEventData struct {
			UserID string
		}

		EmailEventData struct {
			Dest string
		}

		LarkEventData struct {
			AccountID string
		}
	)

	var wg sync.WaitGroup
	wg.Add(2)

	mockUserID := uuid.New().String()

	require.NoError(t, eventbus.Subscribe(ctx, UserEvent, func(_ Event, payload Payload) {
		var data UserEventData
		require.NoError(t, mapstructure.Decode(payload, &data))

		require.NoError(t, eventbus.Publish(ctx, EmailEvent, EmailEventData{
			Dest: fmt.Sprintf("%s@gmail.com", data.UserID),
		}))

		require.NoError(t, eventbus.Publish(ctx, LarkEvent, LarkEventData{
			AccountID: fmt.Sprintf("lark@%s", data.UserID),
		}))
	}))

	require.NoError(t, eventbus.Subscribe(ctx, EmailEvent, func(_ Event, payload Payload) {
		var data EmailEventData
		require.NoError(t, mapstructure.Decode(payload, &data))
		require.True(t, strings.HasPrefix(data.Dest, mockUserID))
		wg.Done()
	}))

	require.NoError(t, eventbus.Subscribe(ctx, LarkEvent, func(_ Event, payload Payload) {
		var data LarkEventData
		require.NoError(t, mapstructure.Decode(payload, &data))
		require.True(t, strings.HasSuffix(data.AccountID, mockUserID))
		wg.Done()
	}))

	require.NoError(t, eventbus.Publish(ctx, UserEvent, UserEventData{
		UserID: mockUserID,
	}))

	wg.Wait()
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
