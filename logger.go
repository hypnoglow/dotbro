package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const logFilepath = "${HOME}/.dotbro/dotbro.log"

// Logger is just an internal logger for debugging purposes.
type Logger struct {
	*log.Logger
}

var logger *Logger

func init() {
	var filename = os.ExpandEnv(logFilepath)

	err := createPath(filename)
	if err != nil {
		outWarn("Cannot use log file %s. Reason: %s", filename, err)
		return
	}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		outWarn("Cannot use log file %s. Reason: %s", filename, err)
		return
	}

	logger = &Logger{
		log.New(f, "", log.Ldate|log.Ltime),
	}

	logger.msg("Init logger")
}

// msg logs message to log file
func (lg *Logger) msg(format string, v ...interface{}) {
	if lg == nil {
		return
	}

	msg := fmt.Sprintf(format, v...)

	// Prepend file and line
	_, filename, line, _ := runtime.Caller(1)
	msg = fmt.Sprintf("%s:%d %s", filepath.Base(filename), line, msg)

	lg.Println(strings.TrimSpace(msg))
}

// exit actually calls os.Exit after logger logs exit message.
func exit(exitCode int) {
	logger.msg("Exit with code %d.", exitCode)
	os.Exit(exitCode)
}
