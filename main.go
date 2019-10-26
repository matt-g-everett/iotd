package main

import (
    "fmt"
    "log"
    "os"
    "time"

    "github.com/eclipse/paho.mqtt.golang"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    fmt.Printf("TOPIC: %s\n", msg.Topic())
    fmt.Printf("MSG: %s\n", msg.Payload())
}

func main() {
    mqtt.DEBUG = log.New(os.Stdout, "", 0)
    mqtt.ERROR = log.New(os.Stdout, "", 0)

    opts := mqtt.NewClientOptions().AddBroker("tcp://***REMOVED***:31883").SetClientID("iotd").SetUsername("***REMOVED***").SetPassword("***REMOVED***")
    opts.SetKeepAlive(2 * time.Second)
    opts.SetDefaultPublishHandler(f)
    opts.SetPingTimeout(1 * time.Second)

    c := mqtt.NewClient(opts)
    if token := c.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    if token := c.Subscribe("home/versions", 1, nil); token.Wait() && token.Error() != nil {
        fmt.Println(token.Error())
        os.Exit(1)
    }

    for {
        fmt.Println("hello world")
        token := c.Publish("home/upgrade", 1, false, "Let's upgrade!")
        token.Wait()
        time.Sleep(5 * time.Second)
    }
}
