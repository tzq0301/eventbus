package eventbus

import (
	"context"
	"sync"
	"testing"
)

func TestAPI(t *testing.T) {
	ctx := context.TODO()

	SetDefault(New(NewInMemoryHandler()))

	event := "test"

	type Data struct {
		ID   int
		Name string
	}

	mockData := Data{
		ID:   10,
		Name: "hello",
	}

	var wg sync.WaitGroup
	wg.Add(1)

	_ = Subscribe(ctx, event, func(_ Event, payload Payload) {
		defer wg.Done()
		data := payload.(Data)
		if data.ID != mockData.ID || data.Name != mockData.Name {
			t.Fail()
		}
	})

	_ = Publish(ctx, event, mockData)

	wg.Wait()
}
