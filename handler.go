package eventbus

import (
	"context"
)

type SubCmd struct {
	Event    Event
	Callback Callback
}

type PubCmd struct {
	Event   Event
	Payload Payload
}

type Handler interface {
	Subscribe(ctx context.Context, cmd SubCmd) error
	Publish(ctx context.Context, cmd PubCmd) error
}
