package main

import (
	"bytes"
	"os"
	"testing"
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
			expected: "WARN: This is a sample warn output that will be shown in verbose mode.\n",
		},
		{
			mode:     OutputerModeNormal,
			format:   "This is a sample warn output that will be shown in %s mode.",
			argument: "normal",
			expected: "WARN: This is a sample warn output that will be shown in normal mode.\n",
		},
		{
			mode:     OutputerModeQuiet,
			format:   "This is a sample warn output that will be shown in %s mode.",
			argument: "quiet",
			expected: "WARN: This is a sample warn output that will be shown in quiet mode.\n",
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
			expected: "ERRO: This is a sample error output that will be shown in verbose mode.\n",
		},
		{
			mode:     OutputerModeNormal,
			format:   "This is a sample error output that will be shown in %s mode.",
			argument: "normal",
			expected: "ERRO: This is a sample error output that will be shown in normal mode.\n",
		},
		{
			mode:     OutputerModeQuiet,
			format:   "This is a sample error output that will be shown in %s mode.",
			argument: "quiet",
			expected: "ERRO: This is a sample error output that will be shown in quiet mode.\n",
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
