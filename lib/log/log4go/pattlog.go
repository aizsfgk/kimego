package log4go

func FormatLogRecord(format string, lr *LogRecord) string {
	return lr.Message
}
