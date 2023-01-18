package eventbus

import "sync"

type SubscribeCallback func(data any)

type EventBus interface {
	PublishAsync(eventName string, data any)
	Subscribe(eventName string, callback SubscribeCallback)
}

func NewEventBus() EventBus {
	e := &eventBusImpl{}
	e.eventNameToSubscribeCallbacks = make(map[string][]SubscribeCallback, 0)
	return e
}

type eventBusImpl struct {
	eventNameToSubscribeCallbacks map[string][]SubscribeCallback
	mu                            sync.RWMutex
}

func (e *eventBusImpl) PublishAsync(eventName string, data any) {
	go func() {
		e.mu.RLock()
		defer e.mu.RUnlock()

		for _, callback := range e.eventNameToSubscribeCallbacks[eventName] {
			callback(data)
		}
	}()
}

func (e *eventBusImpl) Subscribe(eventName string, callback SubscribeCallback) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.eventNameToSubscribeCallbacks[eventName] = append(e.eventNameToSubscribeCallbacks[eventName], callback)
}
