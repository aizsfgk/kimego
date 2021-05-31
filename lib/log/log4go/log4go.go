package log4go

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
)

type LevelType int

const (
	FINE  LevelType = iota // 0
	DEBUG                  // 1
	TRACE
	INFO
	WARNING
	ERROR
	FATAL
)

var (
	levelStrings = [...]string{
		"FINE ",
		"DEBUG",
		"TRACE",
		"INFO ",
		"WARN ",
		"ERROR",
		"FATAL",
	}
)

var (
	LogBufferLength = 1024
	LogWithBlocking = true
)

type LogRecord struct {
	Level   LevelType // The log LevelType
	Created time.Time // The time at which the log message was created (nanoseconds)
	Source  string    // The message source
	Message string    // The log message
	Binary  []byte    // binary log message
}

type LogWriter interface {
	LogWrite(rec *LogRecord)

	Close()
}

type WriterInfo interface {
	Name() string
	QueueLen() int
}

type WriterInfoArray []WriterInfo

var writersInfo WriterInfoArray = make(WriterInfoArray, 0)

type Filter struct {
	Level LevelType
	LogWriter
}

type Logger map[string]*Filter

// LogCloser
type LogCloser struct {
	IsEnd chan bool
}

func (lc *LogCloser) LogCloserInit() {
	lc.IsEnd = make(chan bool)
}

func (lc *LogCloser) EndNotify(lr *LogRecord) bool {
	if lr == nil && lc.IsEnd != nil {
		lc.IsEnd <- true
		return true
	}
	return false
}

func (lc *LogCloser) WaitForEnd(rec chan *LogRecord) {
	rec <- nil
	if lc.IsEnd != nil {
		<-lc.IsEnd
	}
}

func (log Logger) AddFilter(name string, lvl LevelType, writer LogWriter) Logger {
	log[name] = &Filter{lvl, writer}
	return log
}

func (log Logger) Warn(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = WARNING
	)

	var msg string
	switch first := arg0.(type) {
	case string:
		msg = fmt.Sprintf(first, args...)
	case func() string:
		msg = first()
	default:
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}

	log.intLogf(lvl, msg)
	return errors.New(msg)
}

func (log Logger) Info(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = INFO
	)

	var msg string
	switch first := arg0.(type) {
	case string:
		msg = fmt.Sprintf(first, args...)
	case func() string:
		msg = first()
	default:
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}

	log.intLogf(lvl, msg)
	return errors.New(msg)
}

func (log Logger) intLogf(lvl LevelType, format string, args ...interface{}) {
	skip := true

	for _, filt := range log {
		if lvl >= filt.Level {
			skip = false
			break
		}
	}

	if skip {
		return
	}

	pc, _, lineno, ok := runtime.Caller(2)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
	}

	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	// make the log record
	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Source:  src,
		Message: msg,
		Binary:  nil,
	}

	for _, filt := range log {
		if lvl < filt.Level {
			continue
		}
		filt.LogWrite(rec)
	}
}
