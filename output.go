package main

import (
	"fmt"
	"os"
)

// outVerbose prints message to stdout if program is running in verbose mode.
func outVerbose(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	logger.msg(msg)

	if isQuiet {
		return
	}

	if !isVerbose {
		return
	}

	fmt.Fprintln(os.Stdout, msg)
}

// outInfo prints message to stdout.
func outInfo(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	logger.msg(msg)

	if isQuiet {
		return
	}

	fmt.Fprintln(os.Stdout, msg)
}

// outWarn prints warning message to stdout.
func outWarn(format string, v ...interface{}) {
	msg := fmt.Sprintf("WARN: %s", fmt.Sprintf(format, v...))
	logger.msg(msg)

	fmt.Fprintln(os.Stderr, msg)
}

// outError prints error message to stdout.
func outError(format string, v ...interface{}) {
	msg := fmt.Sprintf("ERRO: %s", fmt.Sprintf(format, v...))
	logger.msg(msg)

	fmt.Fprintln(os.Stderr, msg)
}
