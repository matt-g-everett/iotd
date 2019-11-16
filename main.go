package main

import (
    "log"
    "os"
    "time"

    "github.com/eclipse/paho.mqtt.golang"
    "github.com/matt-g-everett/iotd/ota"
)

func main() {
    // mqtt.DEBUG = log.New(os.Stdout, "", 0)
    mqtt.ERROR = log.New(os.Stdout, "", 0)

    options := mqtt.NewClientOptions().
        AddBroker("tcp://***REMOVED***:31883").
        SetClientID("iotd").
        SetUsername("***REMOVED***").
        SetPassword("***REMOVED***").
        SetKeepAlive(30 * time.Second).
        SetPingTimeout(5 * time.Second)

    upgrader := ota.NewUpgrader("data/ota", options)
    upgrader.Run()
}
