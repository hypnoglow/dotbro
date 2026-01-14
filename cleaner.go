package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type Cleaner struct {
	os     OS
	logger *slog.Logger
}

func NewCleaner(os OS, logger *slog.Logger) Cleaner {
	return Cleaner{
		os:     os,
		logger: logger,
	}
}

func (c *Cleaner) CleanDeadSymlinks(ctx context.Context, dirPath string) error {
	dir, err := c.os.Open(dirPath)
	if err != nil {
		return err
	}

	defer func() {
		dirCloseErr := dir.Close()
		if err == nil {
			err = dirCloseErr
		}
	}()

	dirInfo, err := dir.Stat()
	if err != nil {
		return err
	}

	if !dirInfo.IsDir() {
		return fmt.Errorf("Specified dirPath %s is not a directory", dirPath)
	}

	files, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	return c.cleanFiles(ctx, dirPath, files)
}

// Checks each file, if it is a bad symlink - removes it.
func (c *Cleaner) cleanFiles(ctx context.Context, dirPath string, files []os.FileInfo) error {
	removedAny := false
	for _, fileInfo := range files {
		if fileInfo.Mode()&os.ModeSymlink != os.ModeSymlink {
			continue
		}

		filepath := dirPath + "/" + fileInfo.Name()
		_, err := c.os.Stat(filepath)
		if err == nil {
			// symlink is correct
			continue
		}

		if !os.IsNotExist(err) {
			return err
		}

		// file not exists => bad symlink, remove it

		if err := c.os.Remove(filepath); err != nil {
			return err
		}

		if !removedAny {
			removedAny = true
			c.logger.InfoContext(ctx, "Cleaning dead symlinks...")
		}

		c.logger.InfoContext(ctx, "removed broken symlink",
			slog.String("status", "✓"),
			slog.String("path", filepath))
	}

	return nil
}
