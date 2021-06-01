package main

import (
	"github.com/aizsfgk/kimego/lib/log"
	"time"
)

func main() {
	err := log.Init("example", "DEBUG", "./log/",
		true, "M", 3)
	if err != nil {
		panic(err)
	}

	log.Logger.Info("i am test log content...")
	log.Logger.Warn("i am test log content...")
	log.Logger.Info("i am test log content...")

	time.Sleep(10 * time.Second)

}
