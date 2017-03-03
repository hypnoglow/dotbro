package main

import (
	"fmt"
	"os"

	. "github.com/logrusorgru/aurora"
)

type Cleaner struct {
	outputer IOutputer
	os       OS
}

func NewCleaner(outputer IOutputer, os OS) Cleaner {
	return Cleaner{
		outputer: outputer,
		os:       os,
	}
}

func (c *Cleaner) CleanDeadSymlinks(dirPath string) error {
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

	return c.cleanFiles(dirPath, files)
}

// Checks each file, if it is a bad symlink - removes it.
func (c *Cleaner) cleanFiles(dirPath string, files []os.FileInfo) error {
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
			c.outputer.OutInfo("Cleaning dead symlinks...")
		}

		c.outputer.OutInfo("  %s %s has been removed (broken symlink)", Green("✓"), Brown(filepath))
	}

	return nil
}
