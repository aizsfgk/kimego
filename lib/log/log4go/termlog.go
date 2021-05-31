package log4go

import (
	"fmt"
	"io"
	"os"
)

var stdout io.Writer = os.Stdout

type ConsoleLogWriter chan *LogRecord

func NewConsoleLogWriter() ConsoleLogWriter {
	records := make(ConsoleLogWriter, LogBufferLength)
	go records.run(stdout)
	return records
}

func (w ConsoleLogWriter) run(out io.Writer) {
	var (
		timeStr   string
		timeStrAt int64
	)

	for rec := range w {
		if at := rec.Created.UnixNano() / 1e9; at != timeStrAt {
			timeStr, timeStrAt = rec.Created.Format("01/02/06 15:04:05"), at
		}
		fmt.Fprint(out, "[", timeStr, "] [", levelStrings[rec.Level], "] ", rec.Message, "\n")
	}
}

func (w ConsoleLogWriter) LogWrite(rec *LogRecord) {
	if !LogWithBlocking {
		if len(w) >= LogBufferLength {
			return
		}
	}
	w <- rec
}

func (w ConsoleLogWriter) Close() {
	close(w)
}
