package eventbus

import (
	"context"
	"sync"
	"testing"
)

func TestInMemoryHandler(t *testing.T) {
	eventbus := New(NewInMemoryHandler())

	ctx := context.TODO()
	source := "test"
	var wg sync.WaitGroup

	_ = eventbus.Subscribe(ctx, source, func(_ Source, _ Payload) {
		wg.Done()
	})

	total := 100
	wg.Add(total)
	for i := 0; i < total; i++ {
		_ = eventbus.Publish(ctx, source, nil)
	}

	wg.Wait()
}

func TestPubInSub(t *testing.T) {
	eventbus := New(NewInMemoryHandler())

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	outerSource := "outer"
	innerSource := "inner"

	_ = eventbus.Subscribe(ctx, innerSource, func(_ Source, _ Payload) {
		cancel()
	})

	_ = eventbus.Subscribe(ctx, outerSource, func(_ Source, _ Payload) {
		_ = eventbus.Publish(ctx, innerSource, nil)
	})

	_ = eventbus.Publish(ctx, outerSource, nil)

	select {
	case <-ctx.Done():
	}
}

func TestNested(t *testing.T) {
	eventbus := New(NewInMemoryHandler())

	ctx := context.TODO()

	sources := []string{"A", "B", "C"}

	done := make(chan struct{})

	for i := range sources {
		i := i
		_ = eventbus.Subscribe(ctx, sources[i], func(_ Source, _ Payload) {
			_ = eventbus.Publish(ctx, sources[(i+1)%len(sources)], nil)
			done <- struct{}{}
		})
	}

	_ = eventbus.Publish(ctx, sources[0], nil)

	counter := 0
	for range done {
		counter++
		if counter == len(sources) {
			break
		}
	}
}
