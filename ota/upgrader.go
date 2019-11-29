package ota

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"

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
    directory string
    upgrade *Upgrade
}

// NewUpgrader instantiates an Upgrader object.
func NewUpgrader(client mqtt.Client, directory string) *Upgrader {
    u := new(Upgrader)
    u.client = client
    u.directory = directory
    u.index = NewIndex(u.directory)
    u.upgrade = nil

    return u
}

// HandleVersionMessage handles an iotp version message over MQTT.
func (u *Upgrader) HandleVersionMessage(client mqtt.Client, msg mqtt.Message) {
    log.Printf("Received msg %d on %s: %s\n", msg.MessageID(), msg.Topic(), msg.Payload())

    var versionMsg versionReport
    json.Unmarshal(msg.Payload(), &versionMsg)
    log.Printf("%#v\n", versionMsg)

    latest, found := u.index.GetLatest(versionMsg.Type)
    if found {
        semVersion := versionMsg.GetSemVer()
        if latest.GreaterThan(semVersion) {
            log.Printf("###### We should upgrade %s %s @ %s to %s.\n", versionMsg.Type, versionMsg.IP, versionMsg.Version, latest)
            if u.upgrade == nil {
                u.upgrade = NewUpgrade(u, versionMsg.Type, semVersion)
                u.upgrade.advertise()
            } else {
                log.Println("Upgrade already in progress")
            }
        } else {
            log.Printf("###### %s %s @ %s is up to date.\n", versionMsg.Type, versionMsg.IP, versionMsg.Version)
        }
    }
}

func (u *Upgrader) publish() {
    v, found := u.index.GetLatest("logger")
    if found {
        dat, err := ioutil.ReadFile(fmt.Sprintf("data/ota/%s_%s.bin", "logger", v.String()))
        check(err)

        token := u.client.Publish("home/ota/upgradechannel", 0, false, string(dat))
        pubToken := token.(*mqtt.PublishToken)
        log.Println("PUBLISH MSGID: ", pubToken.MessageID())
        token.Wait()
    }
}

// Run the Upgrader.
func (u *Upgrader) Run() {
    u.index.WatchDirectory()
}
