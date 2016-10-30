package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

var isVerbose bool
var isQuiet bool

func main() {
	logger.msg("Start.")

	// Parse arguments

	args, err := parseArguments()
	if err != nil {
		outError("Error parsing aruments: %s", err)
		exit(1)
	}

	logger.msg("Arguments passed: %+v", args)

	isVerbose = args["--verbose"].(bool)
	isQuiet = args["--quiet"].(bool)
	configPath := getConfigPath(args["--config"])

	// Process config

	config, err := NewConfiguration(configPath)
	if err != nil {
		outError("Error reading configuration from file %s: %s", configPath, err)
		exit(1)
	}

	// Preparations

	err = os.MkdirAll(config.Directories.Backup, 0700)
	if err != nil && !os.IsExist(err) {
		outError("Error creating backup directory: %s", err)
		exit(1)
	}

	if isVerbose {
		outVerbose("Dotfiles root: %s", config.Directories.Dotfiles)
		outVerbose("Dotfiles src: %s", config.Directories.Sources)
		outVerbose("Destination dir: %s", config.Directories.Destination)
	}

	// Select action

	switch {
	case args["add"]:
		filename := args["<filename>"].(string)
		if err = addAction(filename, config); err != nil {
			outError("%s", err)
			exit(1)
		}

		outInfo("`%s` was successfully added to your dotfiles!", filename)
		exit(0)
	default:
		// Default action: install
		if err = installAction(config); err != nil {
			outError("%s", err)
			exit(1)
		}

		outInfo("All done (─‿‿─)")
		exit(0)
	}
}

func addAction(filename string, config *Configuration) error {
	fileInfo, err := os.Lstat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s: no such file or directory", filename)
		}
		return err
	}

	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		return fmt.Errorf("Cannot add file %s - it is a symlink", filename)
	}

	if fileInfo.Mode().IsDir() {
		return fmt.Errorf("Cannot add dir %s - directories are not supported yet.", filename)
	}

	outVerbose("Adding file `%s` to dotfiles root `%s`", filename, config.Directories.Dotfiles)

	// backup file
	err = backupCopy(filename, config.Directories.Backup)
	if err != nil {
		return fmt.Errorf("Cannot backup file %s: %s", filename, err)
	}

	// move file to dotfiles root
	newPath := config.Directories.Dotfiles + "/" + path.Base(filename)
	if err = os.Rename(filename, newPath); err != nil {
		return err
	}

	// Add a symlink to the moved file
	if err = setSymlink(newPath, filename); err != nil {
		return err
	}

	// TODO: write to config file

	return nil
}

func installAction(config *Configuration) error {
	// Default action: install

	err := cleanDeadSymlinks(config.Directories.Destination)
	if err != nil {
		return fmt.Errorf("Error cleaning dead symlinks: %s", err)
	}

	srcDirAbs := config.Directories.Dotfiles
	if config.Directories.Sources != "" {
		if _, err = os.Stat(config.Directories.Sources); os.IsNotExist(err) {
			return fmt.Errorf("Sources directory `%s' does not exist.", config.Directories.Sources)
		}
		if err != nil {
			return fmt.Errorf("Error reading sources directory `%s': %s", config.Directories.Sources, err)
		}
		srcDirAbs += "/" + config.Directories.Sources
	}

	mapping := getMapping(config, srcDirAbs)

	outInfo("Installing dotfiles...")
	for src, dst := range mapping {
		installDotfile(src, dst, config, srcDirAbs)
	}

	return nil
}

func getConfigPath(configArg interface{}) string {
	var configPath string
	if configArg != nil {
		configPath = configArg.(string)
	}

	rc := NewRC()
	var err error

	// If config param is not passed to dotbro, read it from RC file.
	if configPath == "" {
		if err = rc.Load(); err != nil {
			outError("Error reading rc file: %s", err)
			exit(1)
		}

		if rc.Config.Path == "" {
			outError("Config file not specified.")
			exit(1)
		}

		outVerbose("Got config path from file `%s`", RCFilepath)
		return rc.Config.Path
	}

	// Save to RC file
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		outError("Bad config path: %s", err)
		exit(1)
	}

	if err = rc.Save(configPath); err != nil {
		outError("Cannot save rc file: %s", err)
		exit(1)
	}

	outVerbose("Saved config path to file `%s`", RCFilepath)
	return rc.Config.Path
}

func getMapping(config *Configuration, srcDirAbs string) map[string]string {
	mapping := make(map[string]string)

	if len(config.Mapping) == 0 {
		// install all the things
		outVerbose("Mapping is not specified - install all the things")
		dir, err := os.Open(srcDirAbs)
		if err != nil {
			outError("Error reading dotfiles source dir: %s", err)
			exit(1)
		}

		defer func() {
			if err = dir.Close(); err != nil {
				outWarn("Error closing dir %s: $s", srcDirAbs, err.Error())
			}
		}()

		files, err := dir.Readdir(0)
		if err != nil {
			outError("Error reading dotfiles source dir: %s", err)
			exit(1)
		}

		for _, fileInfo := range files {
			mapping[fileInfo.Name()] = fileInfo.Name()
		}

		// filter excludes
		for _, exclude := range config.Files.Excludes {
			if _, ok := mapping[exclude]; ok {
				delete(mapping, exclude)
			}
		}
	} else {
		// install by mapping
		if len(config.Files.Excludes) > 0 {
			outWarn("Excludes in config make no sense when mapping is specified, omitting them.")
		}

		mapping = config.Mapping
	}

	return mapping
}

func installDotfile(src, dest string, config *Configuration, srcDirAbs string) {
	srcAbs := srcDirAbs + "/" + src
	destAbs := config.Directories.Destination + "/" + dest

	exists, err := isExists(srcAbs)
	if err != nil {
		outError("Error processing source file %s: %s", src, err)
		exit(1)
	}

	if !exists {
		outWarn("Source file %s does not exist", srcAbs)
		return
	}

	needSymlink, err := NeedSymlink(srcAbs, destAbs)
	if err != nil {
		outError("Error processing destination file %s: %s", destAbs, err)
		exit(1)
	}

	if !needSymlink {
		return
	}

	needBackup, err := needBackup(destAbs)
	if err != nil {
		outError("Error processing destination file %s: %s", destAbs, err)
		exit(1)
	}

	if needBackup {
		err = backup(dest, destAbs, config.Directories.Backup)
		if err != nil {
			outError("Error backuping file %s: %s", destAbs, err)
			exit(1)
		}
	}

	err = setSymlink(srcAbs, destAbs)
	if err != nil {
		outError("Error creating symlink from %s to %s: %s", srcAbs, destAbs, err)
		exit(1)
	}
}
