package log4go

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

func (w *TimeFileLogWriter) LogWrite(rec *LogRecord) {
	if !LogWithBlocking {
		if len(w.rec) >= LogBufferLength {
			return
		}
	}

	w.rec <- rec
}

func (w *TimeFileLogWriter) Close() {
	w.WaitForEnd(w.rec)
	close(w.rec)
}

func NewTimeFileLogWriter(fileName, when string, backupCount int) *TimeFileLogWriter {
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

	// open the file for the first time
	if err := w.initRotate(); err != nil {
		fmt.Fprintf(os.Stderr, "NewTimeFileLogWriter(%q): %s\n", w.filename, err)
		return nil
	}
	go func() {
		defer func() {
			if w.file != nil {
				w.file.Close()
			}
		}()

		for rec := range w.rec {
			if w.EndNotify(rec) {
				return
			}

			if w.shouldRollover() {
				if err := w.initRotate(); err != nil {
					return
				}
			}

			var err error
			if rec.Binary != nil {
				_, err = w.file.Write(rec.Binary)
			} else {
				_, err = fmt.Fprint(w.file, FormatLogRecord(w.format, rec))
			}

			if err != nil {
				return
			}
		}
	}()

	return w

}

func (w *TimeFileLogWriter) shouldRollover() bool {
	t := time.Now().Unix()
	if t >= w.rolloverAt {
		return true
	} else {
		return false
	}
}

func (w *TimeFileLogWriter) initRotate() error {

}

func (w *TimeFileLogWriter) prepare() {
	var regRule string
	switch w.when {
	case "M":
		w.interval = 60
		w.suffix = "%Y-%m-%d_%H-%M"
		regRule = `^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}$`
	case "H", "NEXTHOUR":
		w.interval = 60 * 60
		w.suffix = "%Y-%m-%d_%H"
		regRule = `^\d{4}-\d{2}-\d{2}_\d{2}$`
	case "D", "MIDNIGHT":
		w.interval = 60 * 60 * 24
		w.suffix = "%Y-%m-%d"
		regRule = `^\d{4}-\d{2}-\d{2}$`
	default:
		// default is "D"
		w.interval = 60 * 60 * 24
		w.suffix = "%Y-%m-%d"
		regRule = `^\d{4}-\d{2}-\d{2}$`
	}
	w.fileFilter = regexp.MustCompile(regRule)
	fInfo, err := os.Stat(w.filename)
	var t time.Time
	if err == nil {
		t = fInfo.ModTime()
	} else {
		t = time.Now()
	}
	w.rolloverAt = w.computeRollover(t) // 计算切割
}

func (w *TimeFileLogWriter) computeRollover(curTime time.Time) int64 {
	var result int64
	if w.when == "MIDNIGHT" {

	} else if w.when == "NEXTHOUR" {

	} else {
		result = curTime.Unix() + w.interval
	}
	return result
}

func (w *TimeFileLogWriter) SetFormat(format string) *TimeFileLogWriter {
	w.format = format
	return w
}

func (w *TimeFileLogWriter) Name() string {
	return w.filename
}

func (w *TimeFileLogWriter) QueueLen() int {
	return len(w.rec)
}
