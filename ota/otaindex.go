package ota

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

// Index contains OTA software by type and by version.
type Index struct {
    Entries map[string][]*semver.Version
    C chan bool
    watcher *fsnotify.Watcher
    directory string
}

// NewIndex creates a new Index object.
func NewIndex(directory string) *Index {
    i := new(Index)
    i.directory = directory
    i.Entries = make(map[string][]*semver.Version)
    i.C = make(chan bool)

    var err error
    i.watcher, err = fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    i.watcher.Add(i.directory)
    return i
}

func (i *Index) reload() {
    binPattern := regexp.MustCompile(`^([^_]+)_(.*)\.bin$`)

    // The binary files are in the ota directory
    fileInfo, err := ioutil.ReadDir(i.directory)
    check(err)

    // Reset the Entries
    i.Entries = make(map[string][]*semver.Version)

    // Sort the binary files by software type
    for _, file := range fileInfo {
        m := binPattern.FindStringSubmatch(file.Name())
        version, err := semver.NewVersion(m[2])
        check(err)
        i.Entries[m[1]] = append(i.Entries[m[1]], version)
    }

    // Sort the binary files by version so latest is the first one
    for softwareType, versions := range i.Entries {
        sort.Slice(versions, func(a, b int) bool {
            return versions[a].GreaterThan(versions[b])
        })
        i.Entries[softwareType] = versions
    }

    log.Printf("Reloaded index: %v\n", i.Entries)
    for softwareType := range i.Entries {
        log.Println(softwareType, ":", i.Entries[softwareType][0].String())
    }
}

// WatchDirectory watches the directory associated with the Index.
func (i *Index) WatchDirectory() {
    debounced := debounce(i.watcher.Events)

    for {
        select {
        case <-debounced:
            i.reload()

            // Non-blocking message send
            select {
            case i.C <- true:
            default:
            }
        case err, ok := <-i.watcher.Errors:
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
func (i *Index) GetLatest(softwareType string) (version *semver.Version, found bool) {
    var versions []*semver.Version
    versions, found = i.Entries[softwareType]
    if found {
        if len(versions) > 0 {
            version = versions[0]
        } else {
            found = false
        }
    }

    return
}
