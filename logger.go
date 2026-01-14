package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func newDebugLogger(ctx context.Context) *slog.Logger {
	var output io.Writer = io.Discard

	filename := os.ExpandEnv(logFilepath)
	file, err := openLogFile(filename)
	if err == nil {
		output = file
	} else {
		slog.WarnContext(ctx, fmt.Sprintf("Cannot use log file %s. Reason: %s", filename, err))
	}

	handler := slog.NewTextHandler(output, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})

	logger := slog.New(handler)

	return logger
}

func openLogFile(filename string) (*os.File, error) {
	if err := osfs.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return f, nil
}
