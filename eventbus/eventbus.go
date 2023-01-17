package eventbus

import "sync"

type subscribeCallback func(data any)

type publishData struct {
	message string
	data    any
}

type eventBus struct {
	messageToSubscribeCallback map[string][]subscribeCallback
	mu                         sync.RWMutex
	messageChannel             chan *publishData
}

func newEventBus() *eventBus {
	e := &eventBus{}
	e.messageToSubscribeCallback = make(map[string][]subscribeCallback, 0)
	e.messageChannel = make(chan *publishData)
	go e.listen()
	return e
}

var defaultEventBus = newEventBus()

func (e *eventBus) publish(message string, data any) {
	e.messageChannel <- &publishData{
		message: message,
		data:    data,
	}
}

func (e *eventBus) subscribe(message string, callback subscribeCallback) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.messageToSubscribeCallback[message] = append(e.messageToSubscribeCallback[message], callback)
}

func (e *eventBus) listen() {
	for {
		data := <-e.messageChannel

		go func() {
			e.mu.RLock()
			defer e.mu.RUnlock()

			for _, callback := range e.messageToSubscribeCallback[data.message] {
				callback(data.data)
			}
		}()
	}
}

func Publish(message string, data any) {
	defaultEventBus.publish(message, data)
}

func PublishAsync(message string, data any) {
	go defaultEventBus.publish(message, data)
}

func Subscribe(message string, callback subscribeCallback) {
	defaultEventBus.subscribe(message, callback)
}
