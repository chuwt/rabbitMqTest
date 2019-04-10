package consumer

type Receiver interface {
    Name() string // 名称
    QueueName() string // 获取接收者需要监听的队列
    RouterKey() string // 这个队列绑定的路由
    Exchange() string // 绑定的exchange
    //OnError(error)         // 处理遇到的错误，当RabbitMQ对象发生了错误，他需要告诉接收者处理错误
    OnReceive([]byte, int) bool // 处理收到的消息, 这里需要告知RabbitMQ对象消息是否处理成功
    ReceiverNumber() int         // 一次启动的 receiver 数量
}