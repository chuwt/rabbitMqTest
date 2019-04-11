package consumer

import (
    "sync"
    "fmt"
)

var TtestConsumer *TestConsumer

func init() {
    TtestConsumer = &TestConsumer{
        Mutex: sync.Mutex{},
    }
}

type TestConsumer struct {
    Mutex sync.Mutex // 声明了一个全局锁
}

func (c *TestConsumer) Name() string {
    return "测试消费者"
}

func (c *TestConsumer) QueueName() string {
    return "testQueue"
}

func (c *TestConsumer) RouterKey() string {
    return "testRouter"
}

func (c *TestConsumer) Exchange() string {
    return "testExchange"
}

func (c *TestConsumer) OnReceive(msg []byte, number int) bool {
    fmt.Println(fmt.Sprintf("my number is %d, i receiver a message: %s", number, string(msg)))
    return true
}

func (c *TestConsumer) ReceiverNumber() int {
    return 2
}