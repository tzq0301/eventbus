package eventbus

import (
	"context"
)

type (
	Event    = string
	Payload  = any
	Callback func(event Event, payload Payload)
)

type EventBus struct {
	handler Handler
}

func New(handler Handler) *EventBus {
	return &EventBus{
		handler: handler,
	}
}

func (b *EventBus) Subscribe(ctx context.Context, event Event, callback Callback) error {
	return b.handler.Subscribe(ctx, SubCmd{
		Event:    event,
		Callback: callback,
	})
}

func (b *EventBus) Publish(ctx context.Context, event Event, payload Payload) error {
	return b.handler.Publish(ctx, PubCmd{
		Event:   event,
		Payload: payload,
	})
}
