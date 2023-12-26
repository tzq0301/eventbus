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

	m := make(map[Source][]Callback)

	for {
		select {
		case cmd := <-h.subCh:
			m[cmd.Source] = append(m[cmd.Source], cmd.Callback)
		case cmd := <-h.pubCh:
			for _, callback := range m[cmd.Source] {
				h.worker <- func() {
					callback(cmd.Source, cmd.Payload)
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
