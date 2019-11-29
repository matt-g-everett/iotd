package main

import (
    "log"
    "os"
    "time"

    "github.com/eclipse/paho.mqtt.golang"
    "github.com/matt-g-everett/iotd/ledtx"
    "github.com/matt-g-everett/iotd/ota"
)

type app struct {
    Client mqtt.Client
    Upgrader *ota.Upgrader
    Streamer *ledtx.Streamer
}

func newApp() *app {
    a := new(app)
    return a
}

func (a *app) handleOnConnect(client mqtt.Client) {
    if token := client.Subscribe("home/ota/report", 1, a.Upgrader.HandleVersionMessage); token.Wait() && token.Error() != nil {
        log.Println(token.Error())
        os.Exit(1)
    }
}

func (a *app) run() {
    if token := a.Client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }
    go a.Streamer.Run()
    a.Upgrader.Run()
}

func main() {
    // mqtt.DEBUG = log.New(os.Stdout, "", 0)
    mqtt.ERROR = log.New(os.Stdout, "", 0)

    a := newApp()

    options := mqtt.NewClientOptions().
        AddBroker("tcp://192.168.1.210:31883").
        SetClientID("iotd").
        SetUsername("homeauto").
        SetPassword("fleabags").
        SetKeepAlive(30 * time.Second).
        SetPingTimeout(5 * time.Second).
        SetOnConnectHandler(a.handleOnConnect)
    client := mqtt.NewClient(options)

    a.Client = client
    a.Upgrader = ota.NewUpgrader(client, "data/ota")
    a.Streamer = ledtx.NewStreamer(client)

    a.run()
}
