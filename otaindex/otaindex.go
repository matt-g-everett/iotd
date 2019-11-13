package otaindex

import (
	"io/ioutil"
	"log"
	"regexp"
	"sort"
	"time"

	"github.com/fsnotify/fsnotify"
    "github.com/Masterminds/semver"
)



func check(e error) {
    if e != nil {
        panic(e)
    }
}

// OtaIndex contains OTA software by type and by version.
type OtaIndex struct {
	Entries map[string][]*semver.Version
	C chan bool
	watcher *fsnotify.Watcher
	directory string
}

// NewOtaIndex creates a new OtaIndex object.
func NewOtaIndex(directory string) *OtaIndex {
	o := new(OtaIndex)
	o.directory = directory
	o.Entries = make(map[string][]*semver.Version)
	o.C = make(chan bool)

	var err error
	o.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	o.watcher.Add(o.directory)
	return o
}

func (o *OtaIndex) reload() {
	binPattern := regexp.MustCompile(`^([^_]+)_(.*)\.bin$`)

	// The binary files are in the ota directory
    fileInfo, err := ioutil.ReadDir(o.directory)
	check(err)

	// Reset the Entries
	o.Entries = make(map[string][]*semver.Version)

    // Sort the binary files by software type
    for _, file := range fileInfo {
        m := binPattern.FindStringSubmatch(file.Name())
        version, err := semver.NewVersion(m[2])
        check(err)
        o.Entries[m[1]] = append(o.Entries[m[1]], version)
    }

    // Sort the binary files by version so latest is the first one
    for softwareType, versions := range o.Entries {
        sort.Slice(versions, func(a, b int) bool {
            return versions[a].GreaterThan(versions[b])
        })
        o.Entries[softwareType] = versions
	}

	log.Printf("%v\n", o.Entries)
	for softwareType := range o.Entries {
        log.Println(softwareType, ":", o.Entries[softwareType][0].String())
    }
}

// WatchDirectory watches the directory associated with the OtaIndex.
func (o *OtaIndex) WatchDirectory() {
	debounced := debounce(o.watcher.Events)

	for {
		select {
		case <-debounced: // event, ok := <-o.watcher.Events:
			o.reload()
			o.C <- true
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

	go (func() {
		for {
			select {
			case <-toDebounce:
				// Stop the timer and drain the queue if necessary
				t.Stop()
				select {
				case <-t.C:
				default:
				}

				// Reset the timer
				t.Reset(500 * time.Millisecond)

			case <-t.C:
				// The timer reached completion, relay the message
				c <- true
			}
		}
	})()

	return c
}

// GetLatest gets the version for the specified software type
func (o *OtaIndex) GetLatest(softwareType string) (version *semver.Version, found bool) {
	var versions []*semver.Version
	versions, found = o.Entries[softwareType]
	if found {
		if len(versions) > 0 {
			version = versions[0]
		} else {
			found = false
		}
	}

	return
}
