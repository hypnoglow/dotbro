package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/lmittmann/tint"
)

// multiHandler duplicates log records to multiple handlers.
type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

func newConsoleLogger(level slog.Level) *slog.Logger {
	var fileOutput io.Writer = io.Discard
	filename := os.ExpandEnv(logFilepath)
	file, err := openLogFile(filename)
	if err == nil {
		fileOutput = file
	} else {
		fmt.Fprintf(os.Stderr, "Warning: Cannot use log file %s. Reason: %s\n", filename, err)
	}

	// File handler - plain text without colors, debug level
	fileHandler := slog.NewTextHandler(fileOutput, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})

	// Console handler - tint with colors
	consoleHandler := tint.NewHandler(os.Stderr, &tint.Options{
		Level:      level,
		TimeFormat: "15:04:05",
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Apply semantic colors to specific attributes
			switch a.Key {
			case "path", "src", "dst":
				// Brown/yellow color for paths (ANSI 3 = yellow/brown)
				return tint.Attr(3, a)
			case "status":
				// Green color for status indicators (ANSI 2 = green)
				return tint.Attr(2, a)
			case "tip":
				// Magenta color for tips (ANSI 5 = magenta)
				return tint.Attr(5, a)
			}
			return a
		},
	})

	// Multi-handler to write to both file and console
	multi := &multiHandler{
		handlers: []slog.Handler{fileHandler, consoleHandler},
	}

	return slog.New(multi)
}

func newDiscardLogger() *slog.Logger {
	handler := slog.NewTextHandler(io.Discard, nil)
	return slog.New(handler)
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
