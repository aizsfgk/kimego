package log

import (
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	if err := Init("test", "INFO", "./log/log", true, "M", 2); err != nil {
		t.Error("log.Init() fail")
	}

	if err := Init("test", "Info", "./log/log", true, "M", 5); err == nil {
		t.Error("fail in process reentering log.Init()")
	}

	for i:=0; i<100; i++ {
		Logger.Warn("waring msg:%d", i)
		Logger.Info("info msg: %d", i)
	}

	time.Sleep(50 * time.Millisecond)
}
