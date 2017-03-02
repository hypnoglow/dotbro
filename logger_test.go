package main

import (
	"bytes"
	"log"
	"testing"
)

func TestNewDebugLogger(t *testing.T) {
	NewDebugLogger(nil)
}

func TestDebugLogger_Write(t *testing.T) {
	cases := []struct {
		format   string
		argument string
		expected string
	}{
		{
			format:   "This is the %s simple log entry.",
			argument: "first",
			expected: "logger_test.go:38 This is the first simple log entry.\n",
		},
	}

	// at first, write to empty logger
	emptyLogger := NewDebugLogger(nil)
	emptyLogger.Write("Some text that will not be printed anywhere.")

	// next, check actual logger
	buf := bytes.NewBufferString("")
	log := log.New(buf, "", 0)
	logger := NewDebugLogger(log)

	for _, c := range cases {
		buf.Reset()

		logger.Write(c.format, c.argument) // this line will be included in log
		if buf.String() != c.expected {
			t.Errorf("Expected %q but got %q", c.expected, buf.String())
		}
	}
}
