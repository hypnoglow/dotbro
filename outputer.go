package main

import (
	"fmt"
	"io"

	. "github.com/logrusorgru/aurora"
)

type OutputerMode string

const (
	OutputerModeQuiet   OutputerMode = "3"
	OutputerModeNormal  OutputerMode = "6"
	OutputerModeVerbose OutputerMode = "9"
)

type IOutputer interface {
	OutVerbose(format string, v ...interface{})
	OutInfo(format string, v ...interface{})
	OutWarn(format string, v ...interface{})
	OutError(format string, v ...interface{})
}

// Outputer is a logger that shows the output to the user.
type Outputer struct {
	Mode   OutputerMode
	Output io.Writer
	Logger LogWriter
}

// NewOutputer creates a new Outputer.
func NewOutputer(mode OutputerMode, output io.Writer, logger LogWriter) Outputer {
	return Outputer{
		Mode:   mode,
		Output: output,
		Logger: logger,
	}
}

// OutVerbose prints the message to the Output if the Mode is at least OutputerModeVerbose.
func (o *Outputer) OutVerbose(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	o.Logger.Write(msg)

	if o.Mode < OutputerModeVerbose {
		return
	}

	fmt.Fprintln(o.Output, msg)
}

// OutInfo prints the message to the Output if the Mode is at least OutputerModeNormal.
func (o *Outputer) OutInfo(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	o.Logger.Write(msg)

	if o.Mode < OutputerModeNormal {
		return
	}

	fmt.Fprintln(o.Output, msg)
}

// OutWarn prints warning message to stdout. TODO
func (o *Outputer) OutWarn(format string, v ...interface{}) {
	o.Logger.Write(fmt.Sprintf("WARN: %s", fmt.Sprintf(format, v...)))

	fmt.Fprintln(o.Output, fmt.Sprintf("%s: %s", Brown("WARN"), fmt.Sprintf(format, v...)))
}

// OutError prints error message to stdout.
func (o *Outputer) OutError(format string, v ...interface{}) {
	o.Logger.Write(fmt.Sprintf("ERROR: %s", fmt.Sprintf(format, v...)))

	// TODO: write to stderr
	fmt.Fprintln(o.Output, fmt.Sprintf("%s: %s", Red("ERROR"), fmt.Sprintf(format, v...)))
}
