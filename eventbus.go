package eventbus

import (
	"context"
)

type (
	Source   = string
	Payload  = any
	Callback func(source Source, payload Payload)
)

type EventBus struct {
	handler Handler
}

func New(handler Handler) *EventBus {
	return &EventBus{
		handler: handler,
	}
}

func (b *EventBus) Subscribe(ctx context.Context, source Source, callback Callback) error {
	return b.handler.Subscribe(ctx, SubCmd{
		Source:   source,
		Callback: callback,
	})
}

func (b *EventBus) Publish(ctx context.Context, source Source, payload Payload) error {
	return b.handler.Publish(ctx, PubCmd{
		Source:  source,
		Payload: payload,
	})
}
