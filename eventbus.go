package eventbus

import (
	"context"
)

type (
	Source   string
	Payload  any
	Callback func(source Source, payload Payload) error
)

type EventBus struct {
	handler Handler
}

func NewEventBus(handler Handler) *EventBus {
	return &EventBus{
		handler: handler,
	}
}

func (b *EventBus) Subscribe(ctx context.Context, source Source, callback Callback) error {
	return b.handler.Subscribe(ctx, source, callback)
}

func (b *EventBus) Unsubscribe(ctx context.Context, source Source) error {
	return b.handler.Unsubscribe(ctx, source)
}

func (b *EventBus) Publish(ctx context.Context, source Source, payload Payload) error {
	return b.handler.Publish(ctx, source, payload)
}
