package main

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/logrusorgru/aurora"
)

// FormatWriter describes a writer that can write in printf-like format.
type FormatWriter interface {
	// Write writes message in printf-like format.
	Write(format string, v ...interface{})
}

// WriteLog is a wrapper around log.Logger that can only Write in printf-like format.
type WriteLog struct {
	logger *log.Logger
}

// NewWriteLog creates a new WriteLog.
func NewWriteLog(w io.Writer) WriteLog {
	if w == nil {
		return WriteLog{}
	}

	return WriteLog{
		logger: log.New(w, "", log.Ldate|log.Ltime),
	}
}

func (fl WriteLog) Write(format string, v ...interface{}) {
	if fl.logger == nil {
		return
	}

	msg := fmt.Sprintf(format, v...)

	// Prepend file and line.
	_, filename, line, _ := runtime.Caller(1)
	msg = fmt.Sprintf("%s:%d %s", filepath.Base(filename), line, msg)

	fl.logger.Println(strings.TrimSpace(msg))
}

// LoggerMode is a mode that logger currently on.
type LoggerMode byte

const (
	LoggerModeQuiet   LoggerMode = 3
	LoggerModeNormal             = 6
	LoggerModeVerbose            = 9
)

// LeveLogger describes a logger that can log messages with various level of importance.
type LevelLogger interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warning(format string, v ...interface{})
	Error(format string, v ...interface{})
}

// LevelLog is a logger that can log messages with various level of importance.
// Also, it writes every message regardless of level to a secondary writer named formatWriter.
type LevelLog struct {
	Mode LoggerMode

	// writer in a main writer for the log.
	writer io.Writer

	// formatWriter is a secondary writer, where LevelLog writes each message regardless of level and log mode.
	formatWriter FormatWriter
}

// NewLevelLog creates a new LevelLog.
func NewLevelLog(mode LoggerMode, output io.Writer, fw FormatWriter) LevelLog {
	return LevelLog{
		Mode:         mode,
		writer:       output,
		formatWriter: fw,
	}
}

// Debug prints the message to the writer if the Mode is at least LoggerModeVerbose.
func (o *LevelLog) Debug(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	o.formatWriter.Write(msg)

	if o.Mode < LoggerModeVerbose {
		return
	}

	fmt.Fprintln(o.writer, msg)
}

// Info prints the message to the writer if the Mode is at least LoggerModeNormal.
func (o *LevelLog) Info(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	o.formatWriter.Write(msg)

	if o.Mode < LoggerModeNormal {
		return
	}

	fmt.Fprintln(o.writer, msg)
}

// Warning prints warning message to the writer.
func (o *LevelLog) Warning(format string, v ...interface{}) {
	o.formatWriter.Write(fmt.Sprintf("WARN: %s", fmt.Sprintf(format, v...)))

	fmt.Fprintln(o.writer, fmt.Sprintf("%s: %s", Brown("WARN"), fmt.Sprintf(format, v...)))
}

// Error prints error message to the writer.
func (o *LevelLog) Error(format string, v ...interface{}) {
	o.formatWriter.Write(fmt.Sprintf("ERROR: %s", fmt.Sprintf(format, v...)))

	// TODO: write to stderr?
	fmt.Fprintln(o.writer, fmt.Sprintf("%s: %s", Red("ERROR"), fmt.Sprintf(format, v...)))
}
