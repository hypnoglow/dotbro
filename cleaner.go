package main

import (
	"fmt"
	"os"
)

func cleanDeadSymlinks(dirPath string) error {
	dir, err := os.Open(dirPath)
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

	return cleanFiles(dirPath, files)
}

func cleanFiles(dirPath string, files []os.FileInfo) error {
	removedAny := false
	for _, fileInfo := range files {
		if fileInfo.Mode()&os.ModeSymlink != os.ModeSymlink {
			continue
		}

		filepath := dirPath + "/" + fileInfo.Name()
		removed, err := checkLinkIfBadRemove(filepath)
		if err != nil {
			return err
		}

		if !removed {
			continue
		}

		if !removedAny {
			removedAny = true
			outInfo("Cleaning dead symlinks...")
		}

		outInfo("  ✓ %s has been removed (broken symlink)", filepath)
	}

	return nil
}

func checkLinkIfBadRemove(filepath string) (bool, error) {
	target, err := os.Readlink(filepath)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(target)
	if err == nil {
		// File is ok, no need to remove.
		return false, nil
	}

	if !os.IsNotExist(err) {
		return false, err
	}

	if err := os.Remove(filepath); err != nil {
		return false, err
	}

	return true, nil
}
