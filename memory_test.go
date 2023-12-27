package eventbus

import (
	"context"
	"sync"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

func TestInMemoryHandler(t *testing.T) {
	eventbus := New(NewInMemoryHandler())

	ctx := context.TODO()
	event := "test"
	var wg sync.WaitGroup

	_ = eventbus.Subscribe(ctx, event, func(_ Event, _ Payload) {
		wg.Done()
	})

	total := 100
	wg.Add(total)
	for i := 0; i < total; i++ {
		_ = eventbus.Publish(ctx, event, nil)
	}

	wg.Wait()
}

func TestPayload(t *testing.T) {
	type Data struct {
		ID   int
		Name string
	}

	eventbus := New(NewInMemoryHandler())

	ctx := context.TODO()
	event := "test"

	var wg sync.WaitGroup
	wg.Add(1)

	mock := Data{1, "foo"}

	_ = eventbus.Subscribe(ctx, event, func(_ Event, payload Payload) {
		defer wg.Done()

		var data Data
		require.NoError(t, mapstructure.Decode(payload, &data))

		require.EqualValues(t, mock.ID, data.ID)
		require.Equal(t, mock.Name, data.Name)
	})

	_ = eventbus.Publish(ctx, event, mock)

	wg.Wait()
}

func TestPubInSub(t *testing.T) {
	eventbus := New(NewInMemoryHandler())

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	outerEvent := "outer"
	innerEVent := "inner"

	_ = eventbus.Subscribe(ctx, innerEVent, func(_ Event, _ Payload) {
		cancel()
	})

	_ = eventbus.Subscribe(ctx, outerEvent, func(_ Event, _ Payload) {
		_ = eventbus.Publish(ctx, innerEVent, nil)
	})

	_ = eventbus.Publish(ctx, outerEvent, nil)

	select {
	case <-ctx.Done():
	}
}

func TestNested(t *testing.T) {
	eventbus := New(NewInMemoryHandler())

	ctx := context.TODO()

	events := []string{"A", "B", "C"}

	done := make(chan struct{})

	for i := range events {
		i := i
		_ = eventbus.Subscribe(ctx, events[i], func(_ Event, _ Payload) {
			_ = eventbus.Publish(ctx, events[(i+1)%len(events)], nil)
			done <- struct{}{}
		})
	}

	_ = eventbus.Publish(ctx, events[0], nil)

	counter := 0
	for range done {
		counter++
		if counter == len(events) {
			break
		}
	}
}
