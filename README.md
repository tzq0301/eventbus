# Golang EventBus

EventBus 是进程内的消息队列，用于异步处理消息数据

```shell
go get github.com/tzq0301/eventbus
```

## How to use

参考 eventbus/api_test.go 中的 `TestEventBus` 测试方法

eventbus 提供了默认的实现，可以直接调用 `eventbus.Subscribe` 与 `eventbus.PublishAsync`（eventbus/api.go）

```go
struct Example {}

eventbus.Subscribe("some.event.name", func(data any) {
    // use data
    example := data.(Example)
    // use example
})

eventbus.PublishAsync("some.event.name", Example{})
```

也可以使用 `eventbus.NewEventBus` 方法创建 `EventBus` 的一个默认实现实例，或者自行实现 `eventbus.EventBus` 接口
