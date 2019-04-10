package mqService

const (
    ExchangeDirectType = "direct"
)

var ExchangeMap = map[string]string{
    // exchangeName, exchangeType
    "testExchange": ExchangeDirectType,
}

var QueueMap = map[string]string{
    // queueName, exchangeName
    "testQueue": "testExchange",
}

var RouterKeyMap = map[string]string{
    // queueName, routerKey
    "testQueue": "testRouterKey",
}