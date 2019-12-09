package main

import (
    "flag"
    "log"
    "os"
    "time"

    "github.com/eclipse/paho.mqtt.golang"
    "github.com/matt-g-everett/iotd/ota"
    "gopkg.in/yaml.v2"
)

type app struct {
    Config ota.Config
    Client mqtt.Client
    Upgrader *ota.Upgrader
}

func newApp() *app {
    a := new(app)
    return a
}

func (a *app) handleOnConnect(client mqtt.Client) {
    if token := client.Subscribe(a.Config.Mqtt.Topics.Report, 0, a.Upgrader.HandleVersionMessage); token.Wait() && token.Error() != nil {
        log.Println(token.Error())
        os.Exit(1)
    }
}

func (a *app) run() {
    if token := a.Client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }
    a.Upgrader.Run()
}

func (a *app) readConfig(configPath string) {
    f, err := os.Open(configPath)
    if err != nil {
        panic(err)
    }

    decoder := yaml.NewDecoder(f)
    err = decoder.Decode(&a.Config)
    if err != nil {
        panic(err)
    }
}

func main() {
    // mqtt.DEBUG = log.New(os.Stdout, "", 0)
    mqtt.ERROR = log.New(os.Stdout, "", 0)

    // Parse command line parameters
    configPath := flag.String("config", "config.yaml", "YAML config file.")
    flag.Parse()

    // Read the config
    a := newApp()
    a.readConfig(*configPath)
    log.Printf("Config: %+v", a.Config)

    options := mqtt.NewClientOptions().
        AddBroker(a.Config.Mqtt.URL).
        SetClientID("iotd").
        SetUsername(a.Config.Mqtt.Username).
        SetPassword(a.Config.Mqtt.Password).
        SetKeepAlive(30 * time.Second).
        SetPingTimeout(5 * time.Second).
        SetOnConnectHandler(a.handleOnConnect)
    client := mqtt.NewClient(options)

    a.Client = client
    a.Upgrader = ota.NewUpgrader(client, a.Config)

    a.run()
}
