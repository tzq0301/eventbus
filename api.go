package eventbus

import (
	"context"
	"sync/atomic"
)

var defaultEventBus atomic.Pointer[EventBus]

func init() {
	defaultEventBus.Store(New(NewInMemoryHandler()))
}

func SetDefault(eventbus *EventBus) {
	defaultEventBus.Store(eventbus)
}

func Subscribe(ctx context.Context, source Source, callback Callback) error {
	return defaultEventBus.Load().Subscribe(ctx, source, callback)
}

func Publish(ctx context.Context, source Source, payload Payload) error {
	return defaultEventBus.Load().Publish(ctx, source, payload)
}
