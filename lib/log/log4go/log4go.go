package log4go

import (
	"strings"
	"time"
)

type LevelType int

const (
	FINE LevelType = iota
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	FATAL
)

type LogRecord struct {
	Level   LevelType     // The log LevelType
	Created time.Time // The time at which the log message was created (nanoseconds)
	Source  string    // The message source
	Message string    // The log message
	Binary  []byte    // binary log message
}

type LogWriter interface {
	LogWrite(rec *LogRecord)

	Close()
}

type Filter struct {
	Level LevelType
	LogWriter
}

type Logger map[string]*Filter

func WhenIsValid(when string) bool {
	switch strings.ToUpper(when) {
	case "MIDNIGHT", "NEXTHOUR", "M", "H", "D":
		return true
	default:
		return false
	}
}
