package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/logrusorgru/aurora"
)

// FakeWriteLog is a fake WriteLog.
type FakeWriteLog struct{}

func (f *FakeWriteLog) Write(format string, v ...interface{}) {
	return
}

// FakeLevelLog is a fake LevelLog
type FakeLevelLog struct{}

func (o *FakeLevelLog) Debug(format string, v ...interface{}) {
	return
}

func (o *FakeLevelLog) Info(format string, v ...interface{}) {
	return
}

func (o *FakeLevelLog) Warning(format string, v ...interface{}) {
	return
}

func (o *FakeLevelLog) Error(format string, v ...interface{}) {
	return
}

// Tests

func TestNewWriteLog(t *testing.T) {
	NewWriteLog(nil)
}

func TestWriteLog_Write(t *testing.T) {
	cases := []struct {
		format   string
		argument string
		expected string
	}{
		{
			format:   "This is the %s simple log entry.",
			argument: "first",
			expected: "This is the first simple log entry.\n",
		},
	}

	// at first, write to empty logger
	emptyLogger := NewWriteLog(nil)
	emptyLogger.Write("Some text that will not be printed anywhere. Just check the method works.")

	// next, check actual logger
	buf := bytes.NewBufferString("")
	logger := NewWriteLog(buf)

	for _, c := range cases {
		buf.Reset()

		logger.Write(c.format, c.argument)
		if !strings.HasSuffix(buf.String(), c.expected) {
			t.Errorf("Expected that %q has suffix %q, but is hasn't", buf.String(), c.expected)
		}
	}
}

func TestNewLevelLog(t *testing.T) {
	NewLevelLog(LoggerModeNormal, os.Stdout, &FakeWriteLog{})
}

func TestLevelLog_Debug(t *testing.T) {
	cases := []struct {
		mode     LoggerMode
		format   string
		argument string
		expected string
	}{
		{
			mode:     LoggerModeVerbose,
			format:   "This is a sample verbose output that will be shown in %s mode.",
			argument: "verbose",
			expected: "This is a sample verbose output that will be shown in verbose mode.\n",
		},
		{
			mode:     LoggerModeNormal,
			format:   "This is a sample verbose output that will not be shown in %s mode.",
			argument: "normal",
			expected: "",
		},
		{
			mode:     LoggerModeQuiet,
			format:   "This is a sample verbose output that will not be shown in %s mode.",
			argument: "quiet",
			expected: "",
		},
	}

	output := bytes.NewBufferString("")
	l := NewLevelLog(LoggerModeVerbose, output, &FakeWriteLog{})

	for _, c := range cases {
		output.Reset()

		l.Mode = c.mode
		l.Debug(c.format, c.argument)

		if output.String() != c.expected {
			t.Errorf("Expected %q but got %q", c.expected, output.String())
		}
	}
}

func TestOutputer_OutInfo(t *testing.T) {
	cases := []struct {
		mode     LoggerMode
		format   string
		argument string
		expected string
	}{
		{
			mode:     LoggerModeVerbose,
			format:   "This is a sample info output that will be shown in %s mode.",
			argument: "verbose",
			expected: "This is a sample info output that will be shown in verbose mode.\n",
		},
		{
			mode:     LoggerModeNormal,
			format:   "This is a sample info output that will be shown in %s mode.",
			argument: "normal",
			expected: "This is a sample info output that will be shown in normal mode.\n",
		},
		{
			mode:     LoggerModeQuiet,
			format:   "This is a sample verbose output that will not be shown in %s mode.",
			argument: "quiet",
			expected: "",
		},
	}

	output := bytes.NewBufferString("")
	l := NewLevelLog(LoggerModeVerbose, output, &FakeWriteLog{})

	for _, c := range cases {
		output.Reset()

		l.Mode = c.mode
		l.Info(c.format, c.argument)

		if output.String() != c.expected {
			t.Errorf("Expected %q but got %q", c.expected, output.String())
		}
	}
}

func TestOutputer_OutWarn(t *testing.T) {
	cases := []struct {
		mode     LoggerMode
		format   string
		argument string
		expected string
	}{
		{
			mode:     LoggerModeVerbose,
			format:   "This is a sample warn output that will be shown in %s mode.",
			argument: "verbose",
			expected: fmt.Sprintf("%s: This is a sample warn output that will be shown in verbose mode.\n", Brown("WARN")),
		},
		{
			mode:     LoggerModeNormal,
			format:   "This is a sample warn output that will be shown in %s mode.",
			argument: "normal",
			expected: fmt.Sprintf("%s: This is a sample warn output that will be shown in normal mode.\n", Brown("WARN")),
		},
		{
			mode:     LoggerModeQuiet,
			format:   "This is a sample warn output that will be shown in %s mode.",
			argument: "quiet",
			expected: fmt.Sprintf("%s: This is a sample warn output that will be shown in quiet mode.\n", Brown("WARN")),
		},
	}

	output := bytes.NewBufferString("")
	l := NewLevelLog(LoggerModeVerbose, output, &FakeWriteLog{})

	for _, c := range cases {
		output.Reset()

		l.Mode = c.mode
		l.Warning(c.format, c.argument)

		if output.String() != c.expected {
			t.Errorf("Expected %q but got %q", c.expected, output.String())
		}
	}
}

func TestOutputer_OutError(t *testing.T) {
	cases := []struct {
		mode     LoggerMode
		format   string
		argument string
		expected string
	}{
		{
			mode:     LoggerModeVerbose,
			format:   "This is a sample error output that will be shown in %s mode.",
			argument: "verbose",
			expected: fmt.Sprintf("%s: This is a sample error output that will be shown in verbose mode.\n", Red("ERROR")),
		},
		{
			mode:     LoggerModeNormal,
			format:   "This is a sample error output that will be shown in %s mode.",
			argument: "normal",
			expected: fmt.Sprintf("%s: This is a sample error output that will be shown in normal mode.\n", Red("ERROR")),
		},
		{
			mode:     LoggerModeQuiet,
			format:   "This is a sample error output that will be shown in %s mode.",
			argument: "quiet",
			expected: fmt.Sprintf("%s: This is a sample error output that will be shown in quiet mode.\n", Red("ERROR")),
		},
	}

	output := bytes.NewBufferString("")
	l := NewLevelLog(LoggerModeVerbose, output, &FakeWriteLog{})

	for _, c := range cases {
		output.Reset()

		l.Mode = c.mode
		l.Error(c.format, c.argument)

		if output.String() != c.expected {
			t.Errorf("Expected %q but got %q", c.expected, output.String())
		}
	}
}
