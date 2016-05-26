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

	defer dir.Close()

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

	removedAny := false

	for _, fileInfo := range files {
		if fileInfo.Mode()&os.ModeSymlink != os.ModeSymlink {
			continue
		}

		filepath := dirPath + "/" + fileInfo.Name()

		target, err := os.Readlink(filepath)
		if err != nil {
			return err
		}

		_, err = os.Stat(target)
		if os.IsNotExist(err) {
			os.Remove(filepath)

			if removedAny == false {
				removedAny = true
				outInfo("Cleaning dead symlinks...")
			}
			outInfo("  ✓ %s has been removed (broken symlink)", filepath)
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}
