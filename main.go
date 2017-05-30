package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	. "github.com/logrusorgru/aurora"
)

const logFilepath = "${HOME}/.dotbro/dotbro.log"

var (
	debugLog WriteLog
	outLog   LevelLog
	osfs     = new(OSFS)
)

func main() {
	logFile, err := getLogFile(logFilepath)
	debugLog = NewWriteLog(logFile)
	outLog = NewLevelLog(LoggerModeNormal, os.Stdout, debugLog)
	if err != nil {
		// debugLog will not write to the file. Notify the user.
		outLog.Warning(err.Error())
	}

	debugLog.Write("Start.")

	// Parse arguments

	args, err := ParseArguments(nil)
	if err != nil {
		outLog.Error("Error parsing aruments: %s", err)
		exit(1)
	}

	debugLog.Write("Arguments passed: %+v", args)

	switch {
	case args["--verbose"].(bool):
		outLog.Mode = LoggerModeVerbose
	case args["--quiet"].(bool):
		outLog.Mode = LoggerModeQuiet
	default:
		outLog.Mode = LoggerModeNormal
	}

	// Process config

	configPath := getConfigPath(args["--config"])
	debugLog.Write("Parsing config file %s", configPath)
	config, err := NewConfiguration(configPath)
	if err != nil {
		outLog.Error("Cannot read configuration from file %s : %s.\n", configPath, err)
		outLog.Info("%s: Maybe you have renamed your config file?\nIf so, run dotbro with '--config' argument (see 'dotbro --help' for details).", Magenta("TIP"))
		exit(1)
	}

	// Preparations

	err = os.MkdirAll(config.Directories.Backup, 0700)
	if err != nil && !os.IsExist(err) {
		outLog.Error("Error creating backup directory: %s", err)
		exit(1)
	}

	outLog.Debug("Dotfiles root: %s", Brown(config.Directories.Dotfiles))
	outLog.Debug("Dotfiles src: %s", Brown(config.Directories.Sources))
	outLog.Debug("Destination dir: %s", Brown(config.Directories.Destination))

	// Select action

	switch {
	case args["add"]:
		filename := args["<filename>"].(string)
		if err = addAction(filename, config); err != nil {
			outLog.Error("%s", err)
			exit(1)
		}

		outLog.Info("\n%s was successfully added to your dotfiles!", Brown(filename))
		exit(0)
	case args["clean"]:
		if err = cleanAction(config); err != nil {
			outLog.Error("%s", err)
			exit(1)
		}

		outLog.Info("\nCleaned!")
		exit(0)
	default:
		// Default action: install
		if err = installAction(config); err != nil {
			outLog.Error("%s", err)
			exit(1)
		}

		outLog.Info("\nAll done (─‿‿─)")
		exit(0)
	}
}

func getLogFile(filename string) (*os.File, error) {
	return nil, nil
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return nil, fmt.Errorf("Cannot use log file %s. Reason: %s\n", filename, err)
	}

	logFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("Cannot use log file %s. Reason: %s\n", filename, err)
	}

	// TODO: If file exists, truncate the file to some reasonable amount of lines.

	return logFile, nil
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

	outLog.Debug("Adding file %s to dotfiles root %s", Brown(filename), Brown(config.Directories.Dotfiles))

	// backup file
	backupPath := config.Directories.Backup + "/" + path.Base(filename)
	if err = Copy(osfs, filename, backupPath); err != nil {
		return fmt.Errorf("Cannot backup file %s: %s", filename, err)
	}
	outLog.Info("  %s backup %s to %s", Green("→"), Brown(filename), Brown(backupPath))

	// Move file to dotfiles root
	newPath := config.Directories.Dotfiles + "/" + path.Base(filename)
	if err = os.Rename(filename, newPath); err != nil {
		return err
	}

	linker := NewLinker(&outLog, osfs)

	// Add a symlink to the moved file
	if err = linker.SetSymlink(newPath, filename); err != nil {
		return err
	}

	// TODO: write to config file

	return nil
}

func cleanAction(config *Configuration) error {
	cleaner := NewCleaner(&outLog, osfs)
	if err := cleaner.CleanDeadSymlinks(config.Directories.Destination); err != nil {
		return fmt.Errorf("Error cleaning dead symlinks: %s", err)
	}

	return nil
}

func installAction(config *Configuration) error {
	// Default action: install
	cleaner := NewCleaner(&outLog, osfs)
	err := cleaner.CleanDeadSymlinks(config.Directories.Destination)
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
	linker := NewLinker(&outLog, osfs)

	outLog.Info("Installing dotfiles...")
	for src, dst := range mapping {
		installDotfile(src, dst, linker, config, srcDirAbs)
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
			outLog.Error("Error reading rc file: %s", err)
			exit(1)
		}

		if rc.Config.Path == "" {
			outLog.Error("Config file not specified.")
			exit(1)
		}

		outLog.Debug("Got config path from file %s", Brown(RCFilepath))
		return rc.Config.Path
	}

	// Save to RC file
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		outLog.Error("Bad config path: %s", err)
		exit(1)
	}

	rc.SetPath(configPath)

	if err = rc.Save(); err != nil {
		outLog.Error("Cannot save rc file: %s", err)
		exit(1)
	}

	outLog.Debug("Saved config path to file %s", Brown(RCFilepath))
	return rc.Config.Path
}

func getMapping(config *Configuration, srcDirAbs string) map[string]string {
	mapping := make(map[string]string)

	if len(config.Mapping) == 0 {
		// install all the things
		outLog.Debug("Mapping is not specified - install all the things")
		dir, err := os.Open(srcDirAbs)
		if err != nil {
			outLog.Error("Error reading dotfiles source dir: %s", err)
			exit(1)
		}

		defer func() {
			if err = dir.Close(); err != nil {
				outLog.Warning("Error closing dir %s: $s", srcDirAbs, err.Error())
			}
		}()

		files, err := dir.Readdir(0)
		if err != nil {
			outLog.Error("Error reading dotfiles source dir: %s", err)
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
			outLog.Warning("Excludes in config make no sense when mapping is specified, omitting them.")
		}

		mapping = config.Mapping
	}

	return mapping
}

func installDotfile(src, dest string, linker Linker, config *Configuration, srcDirAbs string) {
	srcAbs := srcDirAbs + "/" + src
	destAbs := config.Directories.Destination + "/" + dest

	_, err := osfs.Stat(srcAbs)
	if osfs.IsNotExist(err) {
		outLog.Warning("Source file %s does not exist", srcAbs)
		return
	}
	if err != nil {
		outLog.Error("Error processing source file %s: %s", src, err)
		exit(1)
	}

	needSymlink, err := linker.NeedSymlink(srcAbs, destAbs)
	if err != nil {
		outLog.Error("Error processing destination file %s: %s", destAbs, err)
		exit(1)
	}

	if !needSymlink {
		return
	}

	needBackup, err := linker.NeedBackup(destAbs)
	if err != nil {
		outLog.Error("Error processing destination file %s: %s", destAbs, err)
		exit(1)
	}

	if needBackup {
		oldpath := destAbs
		newpath := config.Directories.Backup + "/" + dest
		err = linker.Move(oldpath, newpath)
		if err != nil {
			outLog.Error("Error on file backup %s: %s", oldpath, err)
			exit(1)
		}
	}

	err = linker.SetSymlink(srcAbs, destAbs)
	if err != nil {
		outLog.Error("Error creating symlink from %s to %s: %s", srcAbs, destAbs, err)
		exit(1)
	}
}

// exit actually calls os.Exit after logger logs exit message.
func exit(exitCode int) {
	debugLog.Write("Exit with code %d.", exitCode)
	os.Exit(exitCode)
}
