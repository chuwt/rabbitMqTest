package mqService

import "rabbitMqTest/src/mqService/consumer"

func init() {
    // 注册消费者
    RabbitMq.RegisterConsumer(consumer.TtestConsumer)
}