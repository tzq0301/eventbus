package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ Handler = &RedisHandler{}

const (
	redisDummyGroup    = "dummy"
	redisDummyConsumer = "consumer"
	redisDummyKey      = "key"

	redisDefaultPullInterval = 500 * time.Millisecond
	redisDefaultTimeout      = 5 * time.Second
	redisDefaultCountPerPull = 10
)

var (
	redisDefaultPullErrHandler = func(err error) {}
)

type RedisHandler struct {
	rdb *redis.Client

	mu        sync.RWMutex
	callbacks map[string][]Callback // stream -> []Callback

	streamPrefix string // should like "bu:event:" or empty

	pullInterval time.Duration // interval of pulling data from Redis
	pullTimeout  time.Duration // context.WithTimeout
	countPerPull int64         // COUNT argument in XREADGROUP

	errHandler func(err error)
}

type NewRedisHandlerOptions struct {
	rdb          *redis.Client
	streamPrefix string        // should like "bu:event:" or empty
	pullInterval time.Duration // interval of pulling data from Redis
	pullTimeout  time.Duration // context.WithTimeout
	countPerPull int64         // COUNT argument in XREADGROUP
	errHandler   func(err error)
}

type NewRedisHandlerOption func(options *NewRedisHandlerOptions)

func WithRedisRDB(rdb *redis.Client) NewRedisHandlerOption {
	return func(options *NewRedisHandlerOptions) {
		options.rdb = rdb
	}
}

func WithRedisStreamPrefix(prefix string) NewRedisHandlerOption {
	return func(options *NewRedisHandlerOptions) {
		options.streamPrefix = prefix
	}
}

func WithRedisPullInterval(interval time.Duration) NewRedisHandlerOption {
	return func(options *NewRedisHandlerOptions) {
		options.pullInterval = interval
	}
}

func WithRedisPullTimeout(timeout time.Duration) NewRedisHandlerOption {
	return func(options *NewRedisHandlerOptions) {
		options.pullTimeout = timeout
	}
}

func WithRedisCountPerPull(count int64) NewRedisHandlerOption {
	return func(options *NewRedisHandlerOptions) {
		options.countPerPull = count
	}
}

func WithRedisErrHandler(errHandler func(err error)) NewRedisHandlerOption {
	return func(options *NewRedisHandlerOptions) {
		options.errHandler = errHandler
	}
}

func NewRedisHandler(options ...NewRedisHandlerOption) *RedisHandler {
	opts := NewRedisHandlerOptions{
		pullInterval: redisDefaultPullInterval,
		pullTimeout:  redisDefaultTimeout,
		countPerPull: redisDefaultCountPerPull,
		errHandler:   redisDefaultPullErrHandler,
	}

	for _, option := range options {
		option(&opts)
	}

	handler := RedisHandler{
		rdb:          opts.rdb,
		callbacks:    make(map[string][]Callback),
		streamPrefix: opts.streamPrefix,
		pullInterval: opts.pullInterval,
		pullTimeout:  opts.pullTimeout,
		countPerPull: opts.countPerPull,
		errHandler:   opts.errHandler,
	}

	go handler.handle()

	return &handler
}

func (h *RedisHandler) Subscribe(ctx context.Context, cmd SubCmd) error {
	stream := h.streamPrefix + cmd.Source

	err := h.createStreamGroupIfNotExist(ctx, stream)
	if err != nil {
		return fmt.Errorf("subscribe event with source=%q: %w", cmd.Source, err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.callbacks[stream] = append(h.callbacks[stream], cmd.Callback)

	return nil
}

func (h *RedisHandler) Publish(ctx context.Context, cmd PubCmd) error {
	stream := h.streamPrefix + cmd.Source

	err := h.createStreamGroupIfNotExist(ctx, stream)
	if err != nil {
		return fmt.Errorf("publich event with source=%q: %w", cmd.Source, err)
	}

	payload, err := json.Marshal(cmd.Payload)
	if err != nil {
		return fmt.Errorf("convert to json string: %w", err)
	}

	xAddArgs := &redis.XAddArgs{
		Stream: stream,
		ID:     "*",
		Values: []interface{}{redisDummyKey, payload},
	}

	err = h.rdb.XAdd(ctx, xAddArgs).Err()
	if err != nil {
		return fmt.Errorf("redis xadd stream=%q: %w", stream, err)
	}

	return nil
}

func (h *RedisHandler) handle() {
	ticker := time.NewTicker(h.pullInterval)

	for range ticker.C {
		h.handleEachLoop()
	}
}

func (h *RedisHandler) handleEachLoop() {
	ctx, cancel := context.WithTimeout(context.TODO(), h.pullTimeout)
	defer cancel()

	h.mu.Lock()
	defer h.mu.Unlock()

	var xStreams []redis.XStream

	for stream := range h.callbacks {
		xReadGroupArgs := &redis.XReadGroupArgs{
			Group:    redisDummyGroup,
			Consumer: redisDummyConsumer,
			Streams:  []string{stream, ">"},
			Count:    h.countPerPull,
			NoAck:    true,
		}

		result, err := h.rdb.XReadGroup(ctx, xReadGroupArgs).Result()
		if err != nil {
			h.errHandler(fmt.Errorf("redis XReadGroup stream=%q: %w", stream, err))
			continue
		}

		xStreams = append(xStreams, result...)
	}

	for _, xStream := range xStreams {
		stream := xStream.Stream
		for _, callback := range h.callbacks[stream] {
			for _, message := range xStream.Messages {
				bytes := []byte(message.Values[redisDummyKey].(string))
				var payload Payload
				if err := json.Unmarshal(bytes, &payload); err != nil {
					h.errHandler(err)
					continue
				}
				callback(stream, payload)
			}
		}
	}
}

func (h *RedisHandler) createStreamGroupIfNotExist(ctx context.Context, stream string) error {
	err := h.rdb.XGroupCreateMkStream(ctx, stream, redisDummyGroup, "$").Err()
	if err != nil && !redis.HasErrorPrefix(err, "BUSYGROUP") {
		return fmt.Errorf("create redis stream group, stream=%q, group=%q: %w", stream, redisDummyGroup, err)
	}
	return nil
}
