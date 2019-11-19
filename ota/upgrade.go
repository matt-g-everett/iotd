package ota

import (
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/Masterminds/semver"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// An Upgrade is a representation of a upgrade job to be completed.
type Upgrade struct {
	upgrader *Upgrader
	client mqtt.Client
	softwareType string
	version *semver.Version
	channelName string
}

// NewUpgrade creates an upgrade job.
func NewUpgrade(upgrader *Upgrader, softwareType string, version *semver.Version) *Upgrade {
	u := new(Upgrade)
	u.upgrader = upgrader
	u.softwareType = softwareType
	u.version = version
	u.channelName = "home/ota/upgrade-" + randomString(8)

	return u
}

func (u *Upgrade) advertise() {
    v, found := u.upgrader.index.GetLatest(u.softwareType)
    if found {
		filePath := fmt.Sprintf("%s/%s_%s.bin", u.upgrader.directory, u.softwareType, v.String())
		log.Printf("filepath: %s\n", filePath)
		dat, err := ioutil.ReadFile(filePath)
        check(err)
		crc := crc32.ChecksumIEEE(dat)
		log.Printf("crc32: %08x\n", crc)

		message := fmt.Sprintf("%s\n%s\n%s\n%08x", "logger", v.String(), u.channelName, crc)
		log.Println(message)
		token := u.upgrader.client.Publish("home/ota/advertise", 0, false, message)
        pubToken := token.(*mqtt.PublishToken)
        log.Printf("Sent message %d on home/ota/advertise\n", pubToken.MessageID())
		if token.WaitTimeout(10 * time.Second) {
			// Give clients an opportunity to subscribe
			time.Sleep(5 * time.Second)

			token = u.upgrader.client.Publish(u.channelName, 0, true, dat)
			pubToken = token.(*mqtt.PublishToken)
			log.Printf("Sent message %d on %s\n", pubToken.MessageID(), u.channelName)
			if !token.WaitTimeout(10 * time.Second) {
				log.Println("Software failed to send.")
			}
		} else {
			log.Println("Failed to advertise software.")
		}

		// Unhook from the parent so it can do other upgrades
		u.upgrader.upgrade = nil
		log.Printf("Software %s sent, exiting upgrade sequence on %s.\n", filePath, u.channelName)
    }
}
