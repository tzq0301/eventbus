package eventbus

import (
	"context"
)

var _ Handler = &InMemoryHandler{}

type InMemoryHandler struct {
	subCh  chan SubCmd
	pubCh  chan PubCmd
	worker chan func() // use another goroutine to execute callback, to avoid deadlock while Publish in Subscribe
}

func NewInMemoryHandler() *InMemoryHandler {
	handler := InMemoryHandler{
		subCh:  make(chan SubCmd),
		pubCh:  make(chan PubCmd),
		worker: make(chan func()),
	}

	go handler.handle()

	return &handler
}

func (h *InMemoryHandler) Subscribe(_ context.Context, cmd SubCmd) error {
	h.subCh <- cmd
	return nil
}

func (h *InMemoryHandler) Publish(_ context.Context, cmd PubCmd) error {
	h.pubCh <- cmd
	return nil
}

func (h *InMemoryHandler) handle() {
	go h.work()

	m := make(map[Event][]Callback)

	for {
		select {
		case cmd := <-h.subCh:
			m[cmd.Event] = append(m[cmd.Event], cmd.Callback)
		case cmd := <-h.pubCh:
			for _, callback := range m[cmd.Event] {
				h.worker <- func() {
					callback(cmd.Event, cmd.Payload)
				}
			}
		}
	}
}

func (h *InMemoryHandler) work() {
	for {
		select {
		case f := <-h.worker:
			f()
		}
	}
}
