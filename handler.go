package eventbus

import (
	"context"
)

type Handler interface {
	Subscribe(ctx context.Context, source Source, callback Callback) error

	Publish(ctx context.Context, source Source, payload Payload) error
}
