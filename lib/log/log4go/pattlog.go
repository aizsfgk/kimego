package log4go

import (
	"bytes"
	"fmt"
	"sync"
)

const (
	FORMAT_DEFAULT = "[%D %T] [%L] (%S) %M"
)

type formatCacheType struct {
	LastUpdateSeconds    int64
	shortTime, shortDate string
	longTime, longDate   string
}

var (
	formatCache = &formatCacheType{}
	formatMutex sync.Mutex
)

var (
	bufPool sync.Pool
)

func newBuf() *bytes.Buffer {
	if v := bufPool.Get(); v != nil {
		return v.(*bytes.Buffer)
	}

	return bytes.NewBuffer(make([]byte, 0, 4096))
}

func putBuf(bb *bytes.Buffer) {
	bb.Reset()
	bufPool.Put(bb)
}

func FormatLogRecord(format string, lr *LogRecord) string {
	if lr == nil {
		return "<nil>"
	}

	if len(format) == 0 {
		return ""
	}

	out := newBuf()
	defer putBuf(out)

	secs := lr.Created.UnixNano() / 1e9

	formatMutex.Lock()
	cache := *formatCache
	formatMutex.Unlock()

	if cache.LastUpdateSeconds != secs {
		month, day, year := lr.Created.Month(), lr.Created.Day(), lr.Created.Year()
		hour, minute, second := lr.Created.Hour(), lr.Created.Minute(), lr.Created.Second()
		zone, _ := lr.Created.Zone()
		updated := &formatCacheType{
			LastUpdateSeconds: secs,
			shortTime:         fmt.Sprintf("%02d:%02d", hour, minute),
			shortDate:         fmt.Sprintf("%02d/%02d/%02d", month, day, year%100),
			longTime:          fmt.Sprintf("%02d:%02d:%02d %s", hour, minute, second, zone),
			longDate:          fmt.Sprintf("%04d/%02d/%02d", year, month, day),
		}
		formatMutex.Lock()
		cache = *updated
		formatCache = updated
		formatMutex.Unlock()
	}

	pieces := bytes.Split([]byte(format), []byte{'%'})

	for i, piece := range pieces {
		if i > 0 && len(piece) > 0 {
			switch piece[0] {
			case 'T':
				out.WriteString(cache.longTime)
			case 't':
				out.WriteString(cache.shortTime)
			case 'D':
				out.WriteString(cache.longDate)
			case 'd':
				out.WriteString(cache.shortDate)
			case 'L':
				out.WriteString(levelStrings[lr.Level])
			case 'P':
				out.WriteString(LogProcessId)
			case 'S':
				out.WriteString(lr.Source)
			case 'M':
				out.WriteString(lr.Message)
			}
			if len(piece) > 1 {
				out.Write(piece[1:])
			}
		} else if len(piece) > 0 {
			out.Write(piece)
		}
	}
	out.WriteByte('\n')

	return out.String()

}
