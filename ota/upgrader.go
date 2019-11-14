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
)

type versionReport struct {
    IP string `json:"ip"`
    Version string `json:"version"`
}

// Upgrader manages message
type Upgrader struct {
	client mqtt.Client
	index *Index
	options *mqtt.ClientOptions
}

// NewUpgrader instantiates an Upgrader object.
func NewUpgrader(directory string) *Upgrader {
	u := new(Upgrader)
	u.options = mqtt.NewClientOptions().
		AddBroker("tcp://***REMOVED***:31883").
		SetClientID("iotd").
		SetUsername("***REMOVED***").
		SetPassword("***REMOVED***").
		SetKeepAlive(30 * time.Second).
		SetPingTimeout(5 * time.Second)
	u.index = NewIndex(directory)

	return u
}

func (u *Upgrader) handleVersionMessage(client mqtt.Client, msg mqtt.Message) {
	log.Printf("SUBSCRIBE TOPIC: %s\n", msg.Topic())
    log.Printf("SUBSCRIBE MSGID: %d\n", msg.MessageID())

    var versionMsg versionReport
    json.Unmarshal(msg.Payload(), &versionMsg)
    log.Printf("%#v\n", versionMsg)
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
    mqtt.DEBUG = log.New(os.Stdout, "", 0)
    mqtt.ERROR = log.New(os.Stdout, "", 0)

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
