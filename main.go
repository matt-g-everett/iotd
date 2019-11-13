package main

import (
	"encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "sort"
    "time"
    "reflect"
    "regexp"

    "github.com/eclipse/paho.mqtt.golang"
    "github.com/Masterminds/semver"
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

func main() {
    mqtt.DEBUG = log.New(os.Stdout, "", 0)
    mqtt.ERROR = log.New(os.Stdout, "", 0)

    binPattern := regexp.MustCompile(`^([^_]+)_(.*)\.bin$`)

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

    // The binary files are the ota directory
    fileInfo, err := ioutil.ReadDir("data/ota")
    check(err)

    // Sort the binary files by software type
    otaIndex := otaindex.NewOtaIndex("data/ota")
    for _, file := range fileInfo {
        m := binPattern.FindStringSubmatch(file.Name())
        version, err := semver.NewVersion(m[2])
        check(err)
        otaIndex.Entries[m[1]] = append(otaIndex.Entries[m[1]], version)
    }

    // Sort the binary files by version so latest is the first one
    for softwareType, versions := range otaIndex.Entries {
        sort.Slice(versions, func(a, b int) bool {
            return versions[a].GreaterThan(versions[b])
        })
        otaIndex.Entries[softwareType] = versions
    }

    fmt.Printf("%v\n", otaIndex.Entries)

    for softwareType := range otaIndex.Entries {
        println(softwareType, ":", otaIndex.Entries[softwareType][0].String())
    }

    dat, err := ioutil.ReadFile(fmt.Sprintf("data/ota/%s_%s.bin", "logger", otaIndex.Entries["logger"][0].String()))
    check(err)

    go otaIndex.WatchDirectory()

    for {
        token := c.Publish("home/upgrade", 1, false, string(dat))
        fmt.Println("PUBLISH TOKEN: ", reflect.TypeOf(token))
        pubToken := token.(*mqtt.PublishToken)
        fmt.Println("PUBLISH MSGID: ", pubToken.MessageID())
        token.Wait()
        time.Sleep(5 * time.Second)
    }
}
