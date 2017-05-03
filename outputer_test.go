package main

import (
	"bytes"
	"os"
	"testing"

	. "github.com/logrusorgru/aurora"
	"fmt"
)

type FakeLogWriterForOutputer struct{}

func (f *FakeLogWriterForOutputer) Write(format string, v ...interface{}) {
	return
}

func TestNewOutputer(t *testing.T) {
	NewOutputer(OutputerModeNormal, os.Stdout, &FakeLogWriterForOutputer{})
}

func TestOutputer_OutVerbose(t *testing.T) {
	cases := []struct {
		mode     OutputerMode
		format   string
		argument string
		expected string
	}{
		{
			mode:     OutputerModeVerbose,
			format:   "This is a sample verbose output that will be shown in %s mode.",
			argument: "verbose",
			expected: "This is a sample verbose output that will be shown in verbose mode.\n",
		},
		{
			mode:     OutputerModeNormal,
			format:   "This is a sample verbose output that will not be shown in %s mode.",
			argument: "normal",
			expected: "",
		},
		{
			mode:     OutputerModeQuiet,
			format:   "This is a sample verbose output that will not be shown in %s mode.",
			argument: "quiet",
			expected: "",
		},
	}

	output := bytes.NewBufferString("")
	outputer := NewOutputer(OutputerModeVerbose, output, &FakeLogWriterForOutputer{})

	for _, c := range cases {
		output.Reset()

		outputer.Mode = c.mode
		outputer.OutVerbose(c.format, c.argument)

		if output.String() != c.expected {
			t.Errorf("Expected %q but got %q", c.expected, output.String())
		}
	}
}

func TestOutputer_OutInfo(t *testing.T) {
	cases := []struct {
		mode     OutputerMode
		format   string
		argument string
		expected string
	}{
		{
			mode:     OutputerModeVerbose,
			format:   "This is a sample info output that will be shown in %s mode.",
			argument: "verbose",
			expected: "This is a sample info output that will be shown in verbose mode.\n",
		},
		{
			mode:     OutputerModeNormal,
			format:   "This is a sample info output that will be shown in %s mode.",
			argument: "normal",
			expected: "This is a sample info output that will be shown in normal mode.\n",
		},
		{
			mode:     OutputerModeQuiet,
			format:   "This is a sample verbose output that will not be shown in %s mode.",
			argument: "quiet",
			expected: "",
		},
	}

	output := bytes.NewBufferString("")
	outputer := NewOutputer(OutputerModeVerbose, output, &FakeLogWriterForOutputer{})

	for _, c := range cases {
		output.Reset()

		outputer.Mode = c.mode
		outputer.OutInfo(c.format, c.argument)

		if output.String() != c.expected {
			t.Errorf("Expected %q but got %q", c.expected, output.String())
		}
	}
}

func TestOutputer_OutWarn(t *testing.T) {
	cases := []struct {
		mode     OutputerMode
		format   string
		argument string
		expected string
	}{
		{
			mode:     OutputerModeVerbose,
			format:   "This is a sample warn output that will be shown in %s mode.",
			argument: "verbose",
			expected: fmt.Sprintf("%s: This is a sample warn output that will be shown in verbose mode.\n", Brown("WARN")),
		},
		{
			mode:     OutputerModeNormal,
			format:   "This is a sample warn output that will be shown in %s mode.",
			argument: "normal",
			expected: fmt.Sprintf("%s: This is a sample warn output that will be shown in normal mode.\n", Brown("WARN")),
		},
		{
			mode:     OutputerModeQuiet,
			format:   "This is a sample warn output that will be shown in %s mode.",
			argument: "quiet",
			expected: fmt.Sprintf("%s: This is a sample warn output that will be shown in quiet mode.\n", Brown("WARN")),
		},
	}

	output := bytes.NewBufferString("")
	outputer := NewOutputer(OutputerModeVerbose, output, &FakeLogWriterForOutputer{})

	for _, c := range cases {
		output.Reset()

		outputer.Mode = c.mode
		outputer.OutWarn(c.format, c.argument)

		if output.String() != c.expected {
			t.Errorf("Expected %q but got %q", c.expected, output.String())
		}
	}
}

func TestOutputer_OutError(t *testing.T) {
	cases := []struct {
		mode     OutputerMode
		format   string
		argument string
		expected string
	}{
		{
			mode:     OutputerModeVerbose,
			format:   "This is a sample error output that will be shown in %s mode.",
			argument: "verbose",
			expected: fmt.Sprintf("%s: This is a sample error output that will be shown in verbose mode.\n", Red("ERROR")),
		},
		{
			mode:     OutputerModeNormal,
			format:   "This is a sample error output that will be shown in %s mode.",
			argument: "normal",
			expected: fmt.Sprintf("%s: This is a sample error output that will be shown in normal mode.\n", Red("ERROR")),
		},
		{
			mode:     OutputerModeQuiet,
			format:   "This is a sample error output that will be shown in %s mode.",
			argument: "quiet",
			expected: fmt.Sprintf("%s: This is a sample error output that will be shown in quiet mode.\n", Red("ERROR")),
		},
	}

	output := bytes.NewBufferString("")
	outputer := NewOutputer(OutputerModeVerbose, output, &FakeLogWriterForOutputer{})

	for _, c := range cases {
		output.Reset()

		outputer.Mode = c.mode
		outputer.OutError(c.format, c.argument)

		if output.String() != c.expected {
			t.Errorf("Expected %q but got %q", c.expected, output.String())
		}
	}
}

type FakeOutputer struct{}

func (o *FakeOutputer) OutVerbose(format string, v ...interface{}) {
	return
}

func (o *FakeOutputer) OutInfo(format string, v ...interface{}) {
	return
}

func (o *FakeOutputer) OutWarn(format string, v ...interface{}) {
	return
}

func (o *FakeOutputer) OutError(format string, v ...interface{}) {
	return
}
