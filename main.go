package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

var outputer Outputer

type OsStater struct{}

func (s *OsStater) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (s *OsStater) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

type OsDirCheckMaker struct {
	OsStater
}

func (dcm *OsDirCheckMaker) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

var (
	osStater        = new(OsStater)
	osDirCheckMaker = new(OsDirCheckMaker)
)

func main() {
	outputer = NewOutputer(OutputerModeNormal, os.Stdout, logger)
	logger.Msg("Start.")

	// Parse arguments

	args, err := parseArguments()
	if err != nil {
		outputer.OutError("Error parsing aruments: %s", err)
		exit(1)
	}

	logger.Msg("Arguments passed: %+v", args)

	switch {
	case args["--verbose"].(bool):
		outputer.Mode = OutputerModeVerbose
	case args["--quiet"].(bool):
		outputer.Mode = OutputerModeQuiet
	default:
		outputer.Mode = OutputerModeNormal
	}

	// Process config

	configPath := getConfigPath(args["--config"])
	config, err := NewConfiguration(configPath)
	if err != nil {
		outputer.OutError("Error reading configuration from file %s: %s", configPath, err)
		exit(1)
	}

	// Preparations

	err = os.MkdirAll(config.Directories.Backup, 0700)
	if err != nil && !os.IsExist(err) {
		outputer.OutError("Error creating backup directory: %s", err)
		exit(1)
	}

	outputer.OutVerbose("Dotfiles root: %s", config.Directories.Dotfiles)
	outputer.OutVerbose("Dotfiles src: %s", config.Directories.Sources)
	outputer.OutVerbose("Destination dir: %s", config.Directories.Destination)

	// Select action

	switch {
	case args["add"]:
		filename := args["<filename>"].(string)
		if err = addAction(filename, config); err != nil {
			outputer.OutError("%s", err)
			exit(1)
		}

		outputer.OutInfo("`%s` was successfully added to your dotfiles!", filename)
		exit(0)
	default:
		// Default action: install
		if err = installAction(config); err != nil {
			outputer.OutError("%s", err)
			exit(1)
		}

		outputer.OutInfo("All done (─‿‿─)")
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

	outputer.OutVerbose("Adding file `%s` to dotfiles root `%s`", filename, config.Directories.Dotfiles)

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

	outputer.OutInfo("Installing dotfiles...")
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
			outputer.OutError("Error reading rc file: %s", err)
			exit(1)
		}

		if rc.Config.Path == "" {
			outputer.OutError("Config file not specified.")
			exit(1)
		}

		outputer.OutVerbose("Got config path from file `%s`", RCFilepath)
		return rc.Config.Path
	}

	// Save to RC file
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		outputer.OutError("Bad config path: %s", err)
		exit(1)
	}

	rc.SetPath(configPath)

	if err = rc.Save(); err != nil {
		outputer.OutError("Cannot save rc file: %s", err)
		exit(1)
	}

	outputer.OutVerbose("Saved config path to file `%s`", RCFilepath)
	return rc.Config.Path
}

func getMapping(config *Configuration, srcDirAbs string) map[string]string {
	mapping := make(map[string]string)

	if len(config.Mapping) == 0 {
		// install all the things
		outputer.OutVerbose("Mapping is not specified - install all the things")
		dir, err := os.Open(srcDirAbs)
		if err != nil {
			outputer.OutError("Error reading dotfiles source dir: %s", err)
			exit(1)
		}

		defer func() {
			if err = dir.Close(); err != nil {
				outputer.OutWarn("Error closing dir %s: $s", srcDirAbs, err.Error())
			}
		}()

		files, err := dir.Readdir(0)
		if err != nil {
			outputer.OutError("Error reading dotfiles source dir: %s", err)
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
			outputer.OutWarn("Excludes in config make no sense when mapping is specified, omitting them.")
		}

		mapping = config.Mapping
	}

	return mapping
}

func installDotfile(src, dest string, config *Configuration, srcDirAbs string) {
	srcAbs := srcDirAbs + "/" + src
	destAbs := config.Directories.Destination + "/" + dest

	exists, err := IsExists(osStater, srcAbs)
	if err != nil {
		outputer.OutError("Error processing source file %s: %s", src, err)
		exit(1)
	}

	if !exists {
		outputer.OutWarn("Source file %s does not exist", srcAbs)
		return
	}

	needSymlink, err := needSymlink(srcAbs, destAbs)
	if err != nil {
		outputer.OutError("Error processing destination file %s: %s", destAbs, err)
		exit(1)
	}

	if !needSymlink {
		return
	}

	needBackup, err := needBackup(destAbs)
	if err != nil {
		outputer.OutError("Error processing destination file %s: %s", destAbs, err)
		exit(1)
	}

	if needBackup {
		err = backup(dest, destAbs, config.Directories.Backup)
		if err != nil {
			outputer.OutError("Error backuping file %s: %s", destAbs, err)
			exit(1)
		}
	}

	err = setSymlink(srcAbs, destAbs)
	if err != nil {
		outputer.OutError("Error creating symlink from %s to %s: %s", srcAbs, destAbs, err)
		exit(1)
	}
}
