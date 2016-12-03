package main

import (
	"fmt"
	"io"
)

type OutputerMode string

const (
	OutputerModeQuiet   OutputerMode = "3"
	OutputerModeNormal  OutputerMode = "6"
	OutputerModeVerbose OutputerMode = "9"
)

// Outputer is a logger that shows the output to the user.
type Outputer struct {
	Mode   OutputerMode
	Output io.Writer
	Logger ILogger // TODO: refactor to interface
}

// NewOutputer creates a new Outputer.
func NewOutputer(mode OutputerMode, output io.Writer, logger ILogger) Outputer {
	return Outputer{
		Mode:   mode,
		Output: output,
		Logger: logger,
	}
}

// OutVerbose prints the message to the Output if the Mode is at least OutputerModeVerbose.
func (o *Outputer) OutVerbose(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	o.Logger.Msg(msg)

	if o.Mode < OutputerModeVerbose {
		return
	}

	fmt.Fprintln(o.Output, msg)
}

// OutInfo prints the message to the Output if the Mode is at least OutputerModeNormal.
func (o *Outputer) OutInfo(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	o.Logger.Msg(msg)

	if o.Mode < OutputerModeNormal {
		return
	}

	fmt.Fprintln(o.Output, msg)
}

// OutWarn prints warning message to stdout. TODO
func (o *Outputer) OutWarn(format string, v ...interface{}) {
	msg := fmt.Sprintf("WARN: %s", fmt.Sprintf(format, v...))

	o.Logger.Msg(msg)

	fmt.Fprintln(o.Output, msg)
}

// OutError prints error message to stdout.
func (o *Outputer) OutError(format string, v ...interface{}) {
	msg := fmt.Sprintf("ERRO: %s", fmt.Sprintf(format, v...))

	o.Logger.Msg(msg)

	// TODO: write to stderr
	fmt.Fprintln(o.Output, msg)
}
