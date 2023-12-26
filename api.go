package eventbus

import (
	"context"
)

var defaultEventBus = New(NewInMemoryHandler())

func SetDefault(eventbus *EventBus) {
	defaultEventBus = eventbus
}

func Subscribe(ctx context.Context, source Source, callback Callback) error {
	return defaultEventBus.Subscribe(ctx, source, callback)
}

func Publish(ctx context.Context, source Source, payload Payload) error {
	return defaultEventBus.Publish(ctx, source, payload)
}
