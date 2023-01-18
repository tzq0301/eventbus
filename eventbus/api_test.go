package eventbus

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

const (
	EventCreate = "event.create"
	EventUpdate = "event.update"
	EventDelete = "event.delete"
)

type EventCreateData struct {
	EventId int64
}

type EventUpdateData struct {
	EventId      int64
	OperatorName string
}

type EventDeleteData struct {
	EventId      int64
	OperatorName string
}

func InitLark() {
	SubScribe(EventCreate, func(data any) {
		eventCreateData := data.(EventCreateData)
		larkNotify(fmt.Sprintf("Create Event, Id = %v", eventCreateData.EventId))
	})

	SubScribe(EventUpdate, func(data any) {
		eventUpdateData := data.(EventUpdateData)
		larkNotify(fmt.Sprintf("Update Event, Id = %v, Updater = %v", eventUpdateData.EventId, eventUpdateData.OperatorName))
	})

	SubScribe(EventDelete, func(data any) {
		eventDeleteData := data.(EventDeleteData)
		larkNotify(fmt.Sprintf("Delete Event, Id = %v, Deleter = %v", eventDeleteData.EventId, eventDeleteData.OperatorName))
	})
}

func InitResource() {
	SubScribe(EventCreate, func(data any) {
		eventCreateData := data.(EventCreateData)
		createResource(eventCreateData.EventId)
	})

	SubScribe(EventDelete, func(data any) {
		eventDeleteData := data.(EventDeleteData)
		deleteResource(eventDeleteData.EventId)
	})
}

func TestEventBus(t *testing.T) {
	InitLark()
	InitResource()

	var wg sync.WaitGroup

	wg.Add(4)
	for i := 0; i < 4; i++ {
		i64 := int64(i)
		go func() {
			PublishAsync(EventCreate, EventCreateData{
				EventId: i64,
			})
			wg.Done()
		}()
	}

	wg.Wait()
	time.Sleep(time.Second)

	fmt.Println("========================================")

	wg.Add(3)
	for i := 0; i < 3; i++ {
		i64 := int64(i)
		go func() {
			PublishAsync(EventUpdate, EventUpdateData{
				EventId:      i64,
				OperatorName: "Tom",
			})
			wg.Done()
		}()
	}

	wg.Wait()
	time.Sleep(time.Second)

	fmt.Println("========================================")

	wg.Add(2)
	for i := 0; i < 2; i++ {
		i64 := int64(i)
		go func() {
			PublishAsync(EventDelete, EventDeleteData{
				EventId:      i64,
				OperatorName: "Sam",
			})
			wg.Done()
		}()
	}

	wg.Wait()
	time.Sleep(time.Second)
}

func larkNotify(message string) {
	fmt.Printf("Lark: %v\n", message)
}

func createResource(id int64) {
	fmt.Printf("Create Resource, id = %v\n", id)
}

func deleteResource(id int64) {
	fmt.Printf("Delete Resource, id = %v\n", id)
}
