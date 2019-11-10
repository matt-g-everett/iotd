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
)

type VersionReport struct {
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

    var versionReport VersionReport
    json.Unmarshal(msg.Payload(), &versionReport)
    fmt.Printf("%#v\n", versionReport)
}

func main() {
    mqtt.DEBUG = log.New(os.Stdout, "", 0)
    mqtt.ERROR = log.New(os.Stdout, "", 0)

    binPattern := regexp.MustCompile(`^([^_]+)_(.*)\.bin$`)

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

    // The binary files are the ota directory
    fileInfo, err := ioutil.ReadDir("data/ota")
    check(err)

    // Sort the binary files by software type
    binaryIndex := make(map[string][]*semver.Version)
    for _, file := range fileInfo {
        m := binPattern.FindStringSubmatch(file.Name())
        version, err := semver.NewVersion(m[2])
        check(err)
        binaryIndex[m[1]] = append(binaryIndex[m[1]], version)
    }

    // Sort the binary files by version so latest is the first one
    for softwareType, versions := range binaryIndex {
        sort.Slice(versions, func(a, b int) bool {
            return versions[a].GreaterThan(versions[b])
        })
        binaryIndex[softwareType] = versions
    }

    fmt.Printf("%v\n", binaryIndex)

    for softwareType := range binaryIndex {
        println(softwareType, ":", binaryIndex[softwareType][0].String())
    }

    dat, err := ioutil.ReadFile(fmt.Sprintf("data/ota/%s_%s.bin", "logger", binaryIndex["logger"][0].String()))
    check(err)

    for {
        token := c.Publish("home/upgrade", 1, false, string(dat))
        fmt.Println("PUBLISH TOKEN: ", reflect.TypeOf(token))
        pubToken := token.(*mqtt.PublishToken)
        fmt.Println("PUBLISH MSGID: ", pubToken.MessageID())
        token.Wait()
        time.Sleep(5 * time.Second)
    }
}
