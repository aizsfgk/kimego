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

	log.Logger.Info("i am test log content 1...")
	log.Logger.Warn("i am test log content2...")
	log.Logger.Info("i am test log content.3..")

	time.Sleep(10 * time.Second)

	log.Logger.Debug("i am test log content4...")
	log.Logger.Warn("i am test log content5...")
	log.Logger.Error("i am test log content6...")

	time.Sleep(30 * time.Second)

	log.Logger.Debug("Debug Info7: %s, %d", "i am test log content...", 10)
	log.Logger.Warn("i am test log8 content...")
	log.Logger.Error("i am test log9 content...")

	time.Sleep(30 * time.Second)


	log.Logger.Debug("Debug Info10: %s, %d", "i am test log content...", 11)
	log.Logger.Warn("i am test log11 content...")
	log.Logger.Error("i am test log12 content...")

	time.Sleep(10 * time.Second)


}
