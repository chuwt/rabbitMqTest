package main

import (
    "rabbitMqTest/src/mqService"
    )

func main(){
    // publisher
    //for {
    //    mqService.RabbitMq.SafePublish(time.Now().String(), "testExchange", "testRouterKey")
    //    time.Sleep(time.Second*60)
    //}

    // consumer
    mqService.RabbitMq.RunConsumer()
}