package eventbus

import (
	"context"
)

type SubCmd struct {
	Source   Source
	Callback Callback
}

type PubCmd struct {
	Source  Source
	Payload Payload
}

type Handler interface {
	Subscribe(ctx context.Context, cmd SubCmd) error
	Publish(ctx context.Context, cmd PubCmd) error
}
