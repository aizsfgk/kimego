// package log
// log is core for app, so we create private log library first.
package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aizsfgk/kimego/lib/log/log4go"
)

var (
	initialized bool = false
	mutex       sync.Mutex
	Logger      log4go.Logger
)

func Init(proname string, level string, logDir string,
	hasStdOut bool, when string, backupCount int) error {
	mutex.Lock()
	defer mutex.Unlock()

	if initialized {
		return errors.New("has initialized")
	}

	var err error
	Logger, err = Create(proname, level, logDir, hasStdOut, when, backupCount)
	if err != nil {
		return err
	}

	initialized = true
	return nil
}

func Create(proname string, levelStr string, logDir string,
	hasStdOut bool, when string, backupCount int) (log4go.Logger, error) {

	if !log4go.WhenIsValid(when) {
		return nil, fmt.Errorf("invalid valud of when: %s", when)
	}

	if err := logDirCreate(logDir); err != nil {
		//log4go.Error("Init(), in logDirCreate(%s)", logDir)
		return nil, err
	}

	level := stringToLevel(levelStr)

	logger := make(log4go.Logger)
	if hasStdOut {
		logger.AddFilter("stdout", level, log4go.NewConsoleLogWriter())
	}

	// default file
	fileName := filenameGen(proname, logDir, false)
	logWriter := log4go.NewTimeFileLogWriter(fileName, when, backupCount)
	if logWriter == nil {
		return nil, fmt.Errorf("error in log4go.NewTimeFileLogWriter(%s)", fileName)
	}
	logWriter.SetFormat(log4go.LogFormat)
	logger.AddFilter("log", level, logWriter)

	// warning level up file
	fileNameW := filenameGen(proname, logDir, true)
	logWriterW := log4go.NewTimeFileLogWriter(fileNameW, when, backupCount)
	if logWriterW == nil {
		return nil, fmt.Errorf("error in log4go.NewTimeFileLogWriter(%s)", fileNameW)
	}
	logWriterW.SetFormat(log4go.LogFormat)
	logger.AddFilter("log_wf", log4go.WARN, logWriterW)

	return logger, nil

}

func logDirCreate(logDir string) error {
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err = os.MkdirAll(logDir, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func filenameGen(progName, logDir string, isErrLog bool) string {
	logDir = strings.TrimSuffix(logDir, "/")

	var fileName string
	if isErrLog {
		fileName = filepath.Join(logDir, progName+".wf.log")
	} else {
		fileName = filepath.Join(logDir, progName+".log")
	}
	return fileName
}

func stringToLevel(levelStr string) log4go.LevelType {
	var level log4go.LevelType

	levelStr = strings.ToUpper(levelStr)

	switch levelStr {
	case "DEBUG":
		level = log4go.DEBUG
	case "TRACE":
		level = log4go.TRACE
	case "INFO":
		level = log4go.INFO
	case "WARN":
		level = log4go.WARN
	case "ERROR":
		level = log4go.ERROR
	case "FATAL":
		level = log4go.FATAL
	default:
		level = log4go.INFO
	}
	return level
}
