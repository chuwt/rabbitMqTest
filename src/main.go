package main

import (
    "rabbitMqTest/src/mqService"
    "time"
    "fmt"
)

func main(){
    // publisher
    for {
    mqService.RabbitMq.SafePublish(time.Now().String(), "testExchange", "testRouterKey")
    fmt.Println("send msg "+time.Now().String())
    time.Sleep(time.Second*1)
    }

    // consumer
    //mqService.RabbitMq.RunConsumer()
}