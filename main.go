package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	. "github.com/logrusorgru/aurora"
)

const logFilepath = "${HOME}/.dotbro/dotbro.log"

var debugLogger DebugLogger

var (
	osfs = new(OSFS)
)

func main() {
	var outputer = NewOutputer(OutputerModeNormal, os.Stdout, debugLogger)
	initLogger(&outputer)
	outputer.Logger = debugLogger

	debugLogger.Write("Start.")

	// Parse arguments

	args, err := ParseArguments(nil)
	if err != nil {
		outputer.OutError("Error parsing aruments: %s", err)
		exit(1)
	}

	debugLogger.Write("Arguments passed: %+v", args)

	switch {
	case args["--verbose"].(bool):
		outputer.Mode = OutputerModeVerbose
	case args["--quiet"].(bool):
		outputer.Mode = OutputerModeQuiet
	default:
		outputer.Mode = OutputerModeNormal
	}

	// Process config

	configPath := getConfigPath(args["--config"], &outputer)
	debugLogger.Write("Parsing config file %s", configPath)
	config, err := NewConfiguration(configPath)
	if err != nil {
		outputer.OutError("Cannot read configuration from file %s : %s.\n", configPath, err)
		outputer.OutInfo("%s: Maybe you have renamed your config file?\nIf so, run dotbro with '--config' argument (see 'dotbro --help' for details).", Magenta("TIP"))
		exit(1)
	}

	// Preparations

	err = os.MkdirAll(config.Directories.Backup, 0700)
	if err != nil && !os.IsExist(err) {
		outputer.OutError("Error creating backup directory: %s", err)
		exit(1)
	}

	outputer.OutVerbose("Dotfiles root: %s", Brown(config.Directories.Dotfiles))
	outputer.OutVerbose("Dotfiles src: %s", Brown(config.Directories.Sources))
	outputer.OutVerbose("Destination dir: %s", Brown(config.Directories.Destination))

	// Select action

	switch {
	case args["add"]:
		filename := args["<filename>"].(string)
		if err = addAction(filename, config, &outputer); err != nil {
			outputer.OutError("%s", err)
			exit(1)
		}

		outputer.OutInfo("\n%s was successfully added to your dotfiles!", Brown(filename))
		exit(0)
	case args["clean"]:
		if err = cleanAction(config, &outputer); err != nil {
			outputer.OutError("%s", err)
			exit(1)
		}

		outputer.OutInfo("\nCleaned!")
		exit(0)
	default:
		// Default action: install
		if err = installAction(config, &outputer); err != nil {
			outputer.OutError("%s", err)
			exit(1)
		}

		outputer.OutInfo("\nAll done (─‿‿─)")
		exit(0)
	}
}

func initLogger(outputer IOutputer) {
	var filename = os.ExpandEnv(logFilepath)

	if err := osfs.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		outputer.OutWarn("Cannot use log file %s. Reason: %s", filename, err)
		return
	}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		outputer.OutWarn("Cannot use log file %s. Reason: %s", filename, err)
		return
	}

	debugLogger = NewDebugLogger(log.New(f, "", log.Ldate|log.Ltime))
}

func addAction(filename string, config *Configuration, outputer IOutputer) error {
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

	outputer.OutVerbose("Adding file %s to dotfiles root %s", Brown(filename), Brown(config.Directories.Dotfiles))

	// backup file
	backupPath := config.Directories.Backup + "/" + path.Base(filename)
	if err = Copy(osfs, filename, backupPath); err != nil {
		return fmt.Errorf("Cannot backup file %s: %s", filename, err)
	}
	outputer.OutInfo("  %s backup %s to %s", Green("→"), Brown(filename), Brown(backupPath))

	// Move file to dotfiles root
	newPath := config.Directories.Dotfiles + "/" + path.Base(filename)
	if err = os.Rename(filename, newPath); err != nil {
		return err
	}

	linker := NewLinker(outputer, osfs)

	// Add a symlink to the moved file
	if err = linker.SetSymlink(newPath, filename); err != nil {
		return err
	}

	// TODO: write to config file

	return nil
}

func cleanAction(config *Configuration, outputer IOutputer) error {
	cleaner := NewCleaner(outputer, osfs)
	if err := cleaner.CleanDeadSymlinks(config.Directories.Destination); err != nil {
		return fmt.Errorf("Error cleaning dead symlinks: %s", err)
	}

	return nil
}

func installAction(config *Configuration, outputer IOutputer) error {
	// Default action: install
	cleaner := NewCleaner(outputer, osfs)
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

	mapping := getMapping(config, srcDirAbs, outputer)
	linker := NewLinker(outputer, osfs)

	outputer.OutInfo("--> Installing dotfiles...")

	// filter mapping:
	// - non-existent files
	// - already installed files
	filterMapping(mapping, func(src, dst string) bool {
		srcAbs := path.Join(srcDirAbs, src)
		destAbs := path.Join(config.Directories.Destination, dst)
		if _, err := osfs.Stat(srcAbs); err != nil {
			if osfs.IsNotExist(err) {
				outputer.OutWarn("Source file %s does not exist", srcAbs)
				return false
			}

			outputer.OutError("Error processing source file %s: %s", src, err)
			exit(1)
		}
		needSymlink, err := linker.NeedSymlink(srcAbs, destAbs)
		if err != nil {
			outputer.OutError("Error processing destination file %s: %s", destAbs, err)
			exit(1)
		}
		return needSymlink
	})

	if len(mapping) == 0 {
		return nil
	}

	outputer.OutInfo("From %s to %s :", Brown(srcDirAbs), Brown(config.Directories.Destination))
	for src, dst := range mapping {
		installDotfile(src, dst, linker, config, srcDirAbs, outputer)
	}

	return nil
}

func getConfigPath(configArg interface{}, outputer IOutputer) string {
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

		outputer.OutVerbose("Got config path from file %s", Brown(RCFilepath))
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

	outputer.OutVerbose("Saved config path to file %s", Brown(RCFilepath))
	return rc.Config.Path
}

func getMapping(config *Configuration, srcDirAbs string, outputer IOutputer) map[string]string {
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

func filterMapping(mapping map[string]string, callback func(src, dst string) bool) {
	for src, dst := range mapping {
		if !callback(src, dst) {
			delete(mapping, src)
		}
	}
}

func installDotfile(src, dest string, linker Linker, config *Configuration, srcDirAbs string, outputer IOutputer) {
	srcAbs := path.Join(srcDirAbs, src)
	destAbs := path.Join(config.Directories.Destination, dest)

	needBackup, err := linker.NeedBackup(destAbs)
	if err != nil {
		outputer.OutError("Error processing destination file %s: %s", destAbs, err)
		exit(1)
	}

	if needBackup {
		oldpath := destAbs
		newpath := config.Directories.Backup + "/" + dest
		err = linker.Move(oldpath, newpath)
		if err != nil {
			outputer.OutError("Error on file backup %s: %s", oldpath, err)
			exit(1)
		}
	}

	err = linker.SetSymlink(srcAbs, destAbs)
	if err != nil {
		outputer.OutError("Error creating symlink from %s to %s: %s", srcAbs, destAbs, err)
		exit(1)
	}

	outputer.OutInfo("  %s set symlink %s -> %s", Green("+"), Brown(src), Brown(dest))
}

// exit actually calls os.Exit after logger logs exit message.
func exit(exitCode int) {
	debugLogger.Write("Exit with code %d.", exitCode)
	os.Exit(exitCode)
}
