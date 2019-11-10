package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "time"
    "reflect"

    "github.com/eclipse/paho.mqtt.golang"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    fmt.Printf("SUBSCRIBE TOPIC: %s\n", msg.Topic())
    fmt.Printf("SUBSCRIBE MSG: %s\n", msg.Payload())
    fmt.Printf("SUBSCRIBE MSGID: %d\n", msg.MessageID())
}

func main() {
    mqtt.DEBUG = log.New(os.Stdout, "", 0)
    mqtt.ERROR = log.New(os.Stdout, "", 0)

    opts := mqtt.NewClientOptions().
        AddBroker("tcp://192.168.1.210:31883").
        SetClientID("iotd").
        SetUsername("homeauto").
        SetPassword("fleabags").
        SetKeepAlive(30 * time.Second).
        SetDefaultPublishHandler(f).
        SetPingTimeout(5 * time.Second)

    c := mqtt.NewClient(opts)
    if token := c.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    if token := c.Subscribe("home/versions", 1, nil); token.Wait() && token.Error() != nil {
        fmt.Println(token.Error())
        os.Exit(1)
    }

    dat, err := ioutil.ReadFile("data/lorem.txt")
    check(err)
    fmt.Print(string(dat))

    for {
        token := c.Publish("home/upgrade", 1, false, string(dat))
        fmt.Println("PUBLISH TOKEN: ", reflect.TypeOf(token))
        pubToken := token.(*mqtt.PublishToken)
        fmt.Println("PUBLISH MSGID: ", pubToken.MessageID())
        token.Wait()
        time.Sleep(5 * time.Second)
    }
}
