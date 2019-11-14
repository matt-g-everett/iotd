package main

import (
    "log"
    "os"
    "time"

    "github.com/eclipse/paho.mqtt.golang"
    "github.com/matt-g-everett/iotd/ota"
)

func main() {
    mqtt.DEBUG = log.New(os.Stdout, "", 0)
    mqtt.ERROR = log.New(os.Stdout, "", 0)

    options := mqtt.NewClientOptions().
        AddBroker("tcp://192.168.1.210:31883").
        SetClientID("iotd").
        SetUsername("homeauto").
        SetPassword("fleabags").
        SetKeepAlive(30 * time.Second).
        SetPingTimeout(5 * time.Second)

    upgrader := ota.NewUpgrader("data/ota", options)
    upgrader.Run()
}
