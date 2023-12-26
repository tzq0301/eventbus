package eventbus

import (
	"context"
)

var _ Handler = &InMemoryHandler{}

type InMemoryHandler struct {
}

func NewInMemoryHandler() *InMemoryHandler {
	return &InMemoryHandler{}
}

func (h *InMemoryHandler) Subscribe(ctx context.Context, source Source, callback Callback) error {
	// TODO implement me
	panic("implement me")
}

func (h *InMemoryHandler) Unsubscribe(ctx context.Context, source Source) error {
	// TODO implement me
	panic("implement me")
}

func (h *InMemoryHandler) Publish(ctx context.Context, source Source, payload Payload) error {
	// TODO implement me
	panic("implement me")
}
