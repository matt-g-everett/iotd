package ota

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "time"
    "reflect"

    "github.com/eclipse/paho.mqtt.golang"
    "github.com/Masterminds/semver"
)

type versionReport struct {
    IP string `json:"ip"`
    Type string `json:"type"`
    Version string `json:"version"`
}

func (vr *versionReport) GetSemVer() *semver.Version {
    sv, _ := semver.NewVersion(vr.Version)
    return sv
}

// Upgrader manages message
type Upgrader struct {
    client mqtt.Client
    index *Index
    options *mqtt.ClientOptions
}

// NewUpgrader instantiates an Upgrader object.
func NewUpgrader(directory string, opts *mqtt.ClientOptions) *Upgrader {
    u := new(Upgrader)
    u.options = opts
    u.index = NewIndex(directory)

    return u
}

func (u *Upgrader) handleVersionMessage(client mqtt.Client, msg mqtt.Message) {
    log.Printf("SUBSCRIBE TOPIC: %s\n", msg.Topic())
    log.Printf("SUBSCRIBE MSGID: %d\n", msg.MessageID())

    var versionMsg versionReport
    json.Unmarshal(msg.Payload(), &versionMsg)
    log.Printf("%#v\n", versionMsg)

    latest, found := u.index.GetLatest(versionMsg.Type)
    if found {
        if latest.GreaterThan(versionMsg.GetSemVer()) {
            log.Printf("###### We should upgrade %s %s @ %s to %s.\n", versionMsg.Type, versionMsg.IP, versionMsg.Version, latest)
        } else {
            log.Printf("###### %s %s @ %s is up to date.\n", versionMsg.Type, versionMsg.IP, versionMsg.Version)
        }
    }
    //if versionMsg.Version
}

func (u *Upgrader) publish() {
    v, found := u.index.GetLatest("logger")
    if found {
        dat, err := ioutil.ReadFile(fmt.Sprintf("data/ota/%s_%s.bin", "logger", v.String()))
        check(err)

        token := u.client.Publish("home/upgrade", 1, false, string(dat))
        fmt.Println("PUBLISH TOKEN: ", reflect.TypeOf(token))
        pubToken := token.(*mqtt.PublishToken)
        fmt.Println("PUBLISH MSGID: ", pubToken.MessageID())
        token.Wait()
    }
}

// Run the Upgrader.
func (u *Upgrader) Run() {
    u.client = mqtt.NewClient(u.options)
    if token := u.client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    if token := u.client.Subscribe("home/versions", 1, u.handleVersionMessage); token.Wait() && token.Error() != nil {
        fmt.Println(token.Error())
        os.Exit(1)
    }

    go u.index.WatchDirectory()

    publishTimer := time.NewTicker(5 * time.Second)
    for {
        select {
        case <-u.index.C:
            log.Println("###### RELOADED")
        case <-publishTimer.C:
            u.publish()
        }
    }
}
