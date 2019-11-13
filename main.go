package main

import (
	"encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "time"
    "reflect"

    "github.com/eclipse/paho.mqtt.golang"
    "github.com/matt-g-everett/iotd/otaindex"
)

type versionReport struct {
    IP string `json:"ip"`
    Version string `json:"version"`
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    fmt.Printf("SUBSCRIBE TOPIC: %s\n", msg.Topic())
    fmt.Printf("SUBSCRIBE MSGID: %d\n", msg.MessageID())

    var versionMsg versionReport
    json.Unmarshal(msg.Payload(), &versionMsg)
    fmt.Printf("%#v\n", versionMsg)
}

func publish(c mqtt.Client, otaIndex *otaindex.OtaIndex) {
    v, found := otaIndex.GetLatest("logger")
    if found {
        dat, err := ioutil.ReadFile(fmt.Sprintf("data/ota/%s_%s.bin", "logger", v.String()))
        check(err)

        token := c.Publish("home/upgrade", 1, false, string(dat))
        fmt.Println("PUBLISH TOKEN: ", reflect.TypeOf(token))
        pubToken := token.(*mqtt.PublishToken)
        fmt.Println("PUBLISH MSGID: ", pubToken.MessageID())
        token.Wait()
    }
}

func main() {
    mqtt.DEBUG = log.New(os.Stdout, "", 0)
    mqtt.ERROR = log.New(os.Stdout, "", 0)

    opts := mqtt.NewClientOptions().
        AddBroker("tcp://***REMOVED***:31883").
        SetClientID("iotd").
        SetUsername("***REMOVED***").
        SetPassword("***REMOVED***").
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

    otaIndex := otaindex.NewOtaIndex("data/ota")
    go otaIndex.WatchDirectory()

    publishTimer := time.NewTicker(5 * time.Second)
    for {
        select {
        case <-otaIndex.C:
            log.Println("###### RELOADED")
        case <-publishTimer.C:
            publish(c, otaIndex)
        }
    }
}
