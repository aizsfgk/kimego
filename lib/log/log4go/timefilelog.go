package log4go

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type TimeFileLogWriter struct {
	LogCloser // for Elegant exit

	rec chan *LogRecord

	filename     string
	baseFilename string
	file         *os.File

	format string

	when        string // "D", "H", "M", "MIDNIGHT", "NEXTHOUR"
	backupCount int

	interval   int64
	suffix     string
	fileFilter *regexp.Regexp // for removing old log files

	rolloverAt int64 // time.Unix()
}

func WhenIsValid(when string) bool {
	switch strings.ToUpper(when) {
	case "MIDNIGHT", "NEXTHOUR", "M", "H", "D":
		return true
	default:
		return false
	}
}

func NewTimeFileLogWriter(fileName, when string, backupCount int) TimeFileLogWriter {
	if !WhenIsValid(when) {
		fmt.Fprintf(os.Stderr, "NewTimeFileLogWriter(%q): invalid valude of when:%s\n", fileName, when)
		return nil
	}

	when = strings.ToUpper(when)

	w := &TimeFileLogWriter{
		rec:         make(chan *LogRecord, LogBufferLength),
		filename:    fileName,
		format:      "[%D %T] [%L] (%S) %M",
		when:        when,
		backupCount: backupCount,
	}

	writersInfo = append(writersInfo, w)

	// init LogCloser
	w.LogCloserInit()

	// get abs path
	if path, err := filepath.Abs(fileName); err != nil {
		fmt.Fprintf(os.Stderr, "NewTimeFileLogWriter(%q): %s\n", w.filename, err)
		return nil
	} else {
		w.baseFilename = path
	}
	//
	w.prepare()

	return w

}

func (w *TimeFileLogWriter) Name() string {
	return w.filename
}

func (w *TimeFileLogWriter) QueueLen() int {
	return len(w.rec)
}
