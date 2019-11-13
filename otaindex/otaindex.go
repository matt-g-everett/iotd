package otaindex

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
    "github.com/Masterminds/semver"
)

// OtaIndex contains OTA software by type and by version.
type OtaIndex struct {
	Entries map[string][]*semver.Version
	watcher *fsnotify.Watcher
}

// NewOtaIndex creates a new OtaIndex object.
func NewOtaIndex(directory string) *OtaIndex {
	o := new(OtaIndex)
	o.Entries = make(map[string][]*semver.Version)
	var err error
	o.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	o.watcher.Add(directory)
	return o
}

// WatchDirectory watches the directory associated with the OtaIndex.
func (o *OtaIndex) WatchDirectory() {

	debounced := debounce(o.watcher.Events)

	for {
		select {
		case <-debounced: // event, ok := <-o.watcher.Events:
			// if !ok {
			// 	return
			// }
			// log.Println("event:", event)
			// if event.Op & fsnotify.Write == fsnotify.Write {
			// 	log.Println("modified file:", event.Name)
			// }
			log.Println("File changed")
		case err, ok := <-o.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func debounce(toDebounce chan fsnotify.Event) chan bool {
	c := make(chan bool)
	t := time.NewTimer(0)
	send := false

	go (func() {
		for {
			select {
			case <-toDebounce:
				log.Println("Received message")
				t.Stop()
				select {
				case <-t.C:
				default:
				}

				send = true
				t.Reset(500 * time.Millisecond)

			case <-t.C:
				log.Println("Timer received")
				if (send) {
					log.Println("Relaying message")
					send = false
					c <- true
				}
			}
		}
	})()

	return c
}
