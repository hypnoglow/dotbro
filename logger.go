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

// ILoger describes debug logger.
// TODO: rename. This is a bad name (because Logger was not available).
type ILogger interface {
	Msg(format string, v ...interface{})
}

// Logger is just an internal logger for debugging purposes.
type Logger struct {
	*log.Logger
}

var logger *Logger

// Msg logs message to log file
func (lg *Logger) Msg(format string, v ...interface{}) {
	if lg == nil {
		return
	}

	msg := fmt.Sprintf(format, v...)

	// Prepend file and line
	_, filename, line, _ := runtime.Caller(1)
	msg = fmt.Sprintf("%s:%d %s", filepath.Base(filename), line, msg)

	lg.Println(strings.TrimSpace(msg))
}

func init() {
	var filename = os.ExpandEnv(logFilepath)

	err := CreatePath(osDirCheckMaker, filename)
	if err != nil {
		outputer.OutWarn("Cannot use log file %s. Reason: %s", filename, err)
		return
	}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		outputer.OutWarn("Cannot use log file %s. Reason: %s", filename, err)
		return
	}

	logger = &Logger{
		log.New(f, "", log.Ldate|log.Ltime),
	}

	logger.Msg("Init logger")
}

// exit actually calls os.Exit after logger logs exit message.
func exit(exitCode int) {
	logger.Msg("Exit with code %d.", exitCode)
	os.Exit(exitCode)
}
