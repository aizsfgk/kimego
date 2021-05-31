// package log
// log is core for app, so we create private log library first.
package log

import (
	"errors"
	"fmt"
	"sync"

	"kimego/lib/log/log4go"
)

var (
	initialized bool = false
	mutex sync.Mutex
	Logger log4go.Logger
)

func Init(proname string, level string, logDir string,
	hasStdOut bool, when string, backupCount int) error {
	mutex.Lock()
	defer mutex.Unlock()

	if initialized {
		return errors.New("has initialized")
	}

	var err error
	err, Logger = Create(proname, level, logDir, hasStdOut, when, backupCount)
	if err != nil {
		return err
	}

	initialized = true
	return nil
}

func Create(proname string, level string, logDir string,
	hasStdOut bool, when string, backupCount int) (log4go.Logger, error) {

	if !log4go.WhenIsValid(when) {
		return nil, fmt.Errorf("invalid valud of when: %s", when)
	}

}

func strLevel2IntLeve(level string) int {
	return 0
}

