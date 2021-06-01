package log4go

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type LevelType int

const (
	FINE  LevelType = iota // 0
	DEBUG                  // 1
	TRACE
	INFO
	WARN
	ERROR
	FATAL
)

var (
	levelStrings = [...]string{
		"FINE",
		"DEBUG",
		"TRACE",
		"INFO",
		"WARN",
		"ERROR",
		"FATAL",
	}
)

var (
	LogBufferLength = 1024
	LogWithBlocking = true
	LogFormat       = FORMAT_DEFAULT
	LogProcessId    = "0"
	EnableSrcForBinLog = true
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

func setLogProcessId() {
	LogProcessId = strconv.Itoa(os.Getpid())
}

func SetLogFormat(format string) {
	LogFormat = format
	if strings.Contains(LogFormat, "%P") {
		setLogProcessId()
	}
}

func SetLogBufferLength(bufferLen int) {
	LogBufferLength = bufferLen
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

func (log Logger) intLogc(lvl LevelType, closure func() string) {
	skip := true

	// Determine if any logging will be done
	for _, filt := range log {
		if lvl >= filt.Level {
			skip = false
			break
		}
	}
	if skip {
		return
	}

	// Determine caller func
	pc, _, lineno, ok := runtime.Caller(2)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
	}

	// Make the log record
	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Source:  src,
		Message: closure(),
		Binary: nil,
	}

	// Dispatch the logs
	for _, filt := range log {
		if lvl < filt.Level {
			continue
		}
		filt.LogWrite(rec)
	}
}

func (log Logger) intLogb(lvl LevelType, data []byte) {
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

	if len(data) == 0 {
		// no data
		return
	}

	src := ""
	if EnableSrcForBinLog {
		pc, _, lineno, ok := runtime.Caller(2)
		if ok {
			src = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
		}
	}

	// make the log record
	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Source:  src,
		Message: "",
		Binary:  data,
	}

	for _, filt := range log {
		if lvl < filt.Level {
			continue
		}
		filt.LogWrite(rec)
	}
}

func (log Logger) Fine(arg0 interface{}, args ...interface{}) {
	const (
		lvl = FINE
	)

	switch first := arg0.(type) {
	case string:
		log.intLogf(lvl, first, args...)
	case func() string:
		log.intLogc(lvl, first)
	default:
		log.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func (log Logger) Debug(arg0 interface{}, args ...interface{}) {
	const (
		lvl = DEBUG
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		log.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		log.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		log.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func (log Logger) Trace(arg0 interface{}, args ...interface{}) {
	const (
		lvl = TRACE
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		log.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		log.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		log.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func (log Logger) Info(arg0 interface{}, args ...interface{}) {
	const (
		lvl = INFO
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		log.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		log.intLogc(lvl, first)
	case []byte:
		// Log the binary log message
		log.intLogb(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		log.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func (log Logger) Warn(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = WARN
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

func (log Logger) Error(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = ERROR
	)
	var msg string
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		msg = fmt.Sprintf(first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		msg = first()
	default:
		// Build a format string so that it will be similar to Sprint
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
	log.intLogf(lvl, msg)
	return errors.New(msg)
}

func (log Logger) Fatal(arg0 interface{}, args ...interface{}) error  {
	const (
		lvl = FATAL
	)
	var msg string
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		msg = fmt.Sprintf(first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		msg = first()
	default:
		// Build a format string so that it will be similar to Sprint
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
	log.intLogf(lvl, msg)
	return errors.New(msg)
}
