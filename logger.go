package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"
)

// LogWriter describes a writer for the log.
type LogWriter interface {
	Write(format string, v ...interface{})
}

// DebugLogger is a logger that logs output to a log.Logger for debugging purposes.
type DebugLogger struct {
	Log *log.Logger
}

// NewDebugLogger creates a new DebugLogger.
func NewDebugLogger(log *log.Logger) DebugLogger {
	return DebugLogger{
		Log: log,
	}
}

func (dl DebugLogger) Write(format string, v ...interface{}) {
	if dl.Log == nil {
		return
	}

	msg := fmt.Sprintf(format, v...)

	// Prepend file and line
	_, filename, line, _ := runtime.Caller(1)
	msg = fmt.Sprintf("%s:%d %s", filepath.Base(filename), line, msg)

	dl.Log.Println(strings.TrimSpace(msg))
}
