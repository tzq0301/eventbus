package eventbus

import (
	"context"
)

var defaultEventBus = New(NewInMemoryHandler())

func SetDefault(eventbus *EventBus) {
	defaultEventBus = eventbus
}

func Subscribe(ctx context.Context, event Event, callback Callback) error {
	return defaultEventBus.Subscribe(ctx, event, callback)
}

func Publish(ctx context.Context, event Event, payload Payload) error {
	return defaultEventBus.Publish(ctx, event, payload)
}
