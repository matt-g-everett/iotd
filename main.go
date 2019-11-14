package main

import (
    "github.com/matt-g-everett/iotd/ota"
)

func main() {
    upgrader := ota.NewUpgrader("data/ota")
    upgrader.Run()
}
