package mqService

import (
    "sync"
    "github.com/streadway/amqp"
    "github.com/rs/zerolog/log"
    "rabbitMqTest/src/config"
    "rabbitMqTest/src/mqService/consumer"
    "fmt"
)

var RabbitMq *RMQ

type RMQ struct {
    wg        sync.WaitGroup
    Channel   *amqp.Channel
    receivers []consumer.Receiver
}

func init() {
    RabbitMq = &RMQ{}
    //RabbitMq.InitMq()
}

func (r *RMQ) InitMq() {
    r.NewChannel(config.Config.RabbitMQ.URL)
    // 注册exchange
    for exchangeName, exchangeType := range ExchangeMap {
        if err := r.NewExchange(exchangeName, exchangeType); err != nil {
            r.Close()
            panic("RabbitMQ 初始化exchange失败: " + err.Error())
        }
    }
    // 注册queue
    for queueName, exchangeName := range QueueMap {
        if err := r.NewQueue(queueName); err != nil {
            r.Close()
            panic("RabbitMQ 初始化queue失败: " + err.Error())
        } else {
            // 绑定这个queue和exchange的routerKey
            if routerKey, ok := RouterKeyMap[queueName]; ok {
                // 绑定exchange和queue
                if err := r.ExchangeBindQueue(exchangeName, queueName, routerKey); err != nil {
                    r.Close()
                    panic("RabbitMQ bind queue失败: " + err.Error())
                }
            }
        }
    }
}

func (r *RMQ) NewChannel(url string) {
    rabbitMqConn, err := amqp.Dial(url)
    if err != nil {
        panic("RabbitMQ 初始化失败: " + err.Error())
    }
    r.Channel, err = rabbitMqConn.Channel()
    if err != nil {
        panic("初始化Channel失败: " + err.Error())
    }
}

func (r *RMQ) NewExchange(exchangeName, exchangeType string) error {
    // 申明Exchange
    err := r.Channel.ExchangeDeclare(
        exchangeName, // exchange
        exchangeType, // type
        true,         // durable
        false,        // autoDelete
        false,        // internal
        false,        // noWait
        nil,          // args
    )

    if nil != err {
        log.Error().Str("err", err.Error()).Msg("创建exchange失败")
        return err
    }
    return nil
}

func (r *RMQ) NewQueue(queueName string) error {
    // declare queue 创建queue
    _, err := r.Channel.QueueDeclare(
        queueName, // name
        true,      // durable 持久化
        false,     // delete when unused
        false,     // exclusive
        false,     // no-wait
        nil,       // arguments
    )
    if err != nil {
        log.Error().Str("error", err.Error()).Msg("创建queue错误")
        return err
        //panic("创建queue失败")
    }
    return nil
}

func (r *RMQ) ExchangeBindQueue(exchangeName, queueName, routerKey string) error {
    err := r.Channel.QueueBind(
        queueName,    // queue name
        routerKey,    // routing key
        exchangeName, // exchange
        false,        // no-wait
        nil,
    )
    if err != nil {
        log.Error().Str("error", err.Error()).Msg("绑定queue错误")
        return err
    }
    return nil
}

func (r *RMQ) Close() {
    r.Channel.Close()
}

func (r *RMQ) RegisterConsumer(receiver consumer.Receiver) {
    r.receivers = append(r.receivers, receiver)
    log.Info().Str("name", receiver.Name()).Msg("register receiver ok..")
}

func (r *RMQ) SafePublish(msg, exchangeName, routerKey string) bool {
    maxRepeat := 0
    for {
        if err := r.Publish(msg, exchangeName, routerKey); err != nil {
            r.Channel.Close()
            if maxRepeat == config.Config.RabbitMQ.ErrorRepeat {
                log.Error().Msg("断线重连失败")
                return false
            }
            log.Info().Int("重连次数", maxRepeat).Msg("断线重连开始")
            r.NewChannel(config.Config.RabbitMQ.URL)
            maxRepeat += 1
            continue
        }
        return true
    }

}

func (r *RMQ) Publish(msg, exchangeName, routerKey string) error {
    // 推送消息
    err := RabbitMq.Channel.Publish(
        exchangeName, // exchange
        routerKey,    // routing key
        false,        // mandatory
        false,        // immediate
        amqp.Publishing{
            ContentType:  "text/plain",
            Body:         []byte(msg),
            DeliveryMode: 2, // 持久化
        },
    )
    if err != nil {
        log.Error().Str("error", err.Error()).Msg("channel连接断开，推送消息失败")
        return err
    }
    return nil
}

func (r *RMQ) RunConsumer() {
    // 创建channel
    r.NewChannel(config.Config.RabbitMQ.URL)
    log.Info().Msg("start run consumers ... ")
    defer r.Channel.Close()
    for _, receiver := range r.receivers {
        //  创建queue，绑定ex
        err := r.BindConsumer(receiver)
        if err != nil {
            continue
        }
        for th := 0; th < receiver.ReceiverNumber(); th ++ {
            log.Info().Str("name", fmt.Sprintf("%s%d", receiver.Name(), th)).Msg("run consumer")
            r.wg.Add(1)
            go r.listen(receiver, th) // 每个接收者单独启动一个或多个goroutine用来初始化queue并接收消息
        }
    }
    r.wg.Wait()
}

func (r *RMQ) BindConsumer(receiver consumer.Receiver) error {
    // 这里获取每个接收者需要监听的队列和路由
    queueName := receiver.QueueName()
    routerKey := receiver.RouterKey()
    exchange := receiver.Exchange()
    // 创建queue
    if err := r.NewQueue(queueName); err != nil {
        log.Error().Str("receiver", receiver.Name()).Str("queueName", receiver.Name()).Msg("create receiver queue error")
        return err
    }
    // 绑定ex
    if err := r.ExchangeBindQueue(exchange, queueName, routerKey); err != nil {
        log.Error().Str("receiver", receiver.Name()).Str("exchange", exchange).Str("queueName", queueName).Str("routerKey", routerKey).Msg("bind receiver exchange error")
        return err
    }
    return nil
}

func (r *RMQ) listen(receiver consumer.Receiver, th int) {
    defer r.wg.Done()
    // 这里获取每个接收者需要监听的队列和路由
    queueName := receiver.QueueName()

    // 消息缓存，
    r.Channel.Qos(10, 0, true)
    // 获取消费通道
    msgs, err := r.Channel.Consume(
        queueName, // queue
        fmt.Sprintf("%s%d", receiver.Name(), th),        // consumer
        false,     // auto-ack
        false,     // exclusive
        false,     // no-local
        false,     // no-wait
        nil,       // args
    )
    if nil != err {
        // todo 处理错误
        log.Error().Str("queueName", queueName).Str("error", err.Error()).Msg("获取消费队列失败")
        return
    }

    // 使用callback消费数据
    for msg := range msgs {
        // 当接收者消息处理失败的时候，
        // 比如网络问题导致的数据库连接失败，redis连接失败等等这种
        // 通过重试可以成功的操作，那么这个时候是需要重试的
        // 直到数据处理成功后再返回，然后才会回复rabbitmq ack
        //for !receiver.OnReceive(msg.Body, th) {
        //	//log.Warn().Msg("receiver 数据处理失败，将要重试")
        //	time.Sleep(1 * time.Second)
        //}
        receiver.OnReceive(msg.Body, th)
        // 确认收到本条消息, multiple必须为false
        msg.Ack(false)
    }
}