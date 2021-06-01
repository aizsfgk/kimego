package log4go

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aizsfgk/kimego/lib/strftime"
)

const (
	MIDNIGHT = 24 * 60 * 60
	NEXTHOUT = 60 * 60
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
				fmt.Fprintf(os.Stderr, "NewTimeFileLogWriter goroutine err: %s\n", err.Error())
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
	if w.file != nil {
		w.file.Close()
	}

	if w.shouldRollover() {
		if err := w.moveToBackup(); err != nil {
			return err
		}
	}

	// remove files, according to backupCount
	if w.backupCount > 0 {
		for _, fileName := range w.getFilesToDelete() {
			os.Remove(fileName)
		}
	}

	// open the log file
	fd, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	w.file = fd

	w.adjustRolloverAt()

	return nil
}

// moveToBackup renames file to backup name
func (w *TimeFileLogWriter) moveToBackup() error {
	_, err := os.Lstat(w.filename)
	if err != nil { // file exists

		// get the time that this sequence started at and make it a TimeTuple
		t := time.Unix(w.rolloverAt-w.interval, 0).Local()
		fname := w.baseFilename + "." + strftime.Format(w.suffix, t)

		// remove the file with fname if exist
		if _, err = os.Stat(fname); err == nil {
			err = os.Remove(fname)
			if err != nil {
				return fmt.Errorf("moveToBackup err : %s\n", err.Error())
			}
		}

		// rename the file to it's new found home
		err = os.Rename(w.baseFilename, fname)
		if err != nil {
			return fmt.Errorf("moveToBackup err: %s\n", err.Error())
		}

	}
	return nil
}

// getFilesToDelete determines the files to delte when rolling over
func (w *TimeFileLogWriter) getFilesToDelete() []string {
	dirName := filepath.Dir(w.baseFilename)
	baseName := filepath.Base(w.baseFilename)

	result := []string{}
	fileInfos, err := ioutil.ReadDir(dirName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
		return result
	}

	prefix := baseName + "."
	plen := len(prefix)

	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()
		if len(fileName) >= plen {
			if fileName[:plen] == prefix {
				suffix := fileName[plen:]
				if w.fileFilter.MatchString(suffix) {
					result = append(result, filepath.Join(dirName, fileName))
				}
			}
		}
	}

	sort.Sort(sort.StringSlice(result)) // 对文件进行排序

	if len(result) < w.backupCount {
		result = result[0:0]
	} else {
		result = result[:len(result)-w.backupCount]
	}
	return result
}

// 调整下次切割时间
func (w *TimeFileLogWriter) adjustRolloverAt() {
	curTime := time.Now()
	newRolloverAt := w.computeRollover(curTime)

	for newRolloverAt <= curTime.Unix() {
		newRolloverAt = newRolloverAt + w.interval
	}
	w.rolloverAt = newRolloverAt
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
		t := curTime.Local()
		r := MIDNIGHT - (t.Hour()*60 + t.Minute()*60 + t.Second())
		result = curTime.Unix() + int64(r)
	} else if w.when == "NEXTHOUR" {
		t := curTime.Local()
		r := NEXTHOUT - (t.Minute()*60 + t.Second())
		result = curTime.Unix() + int64(r)
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
