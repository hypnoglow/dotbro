package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	. "github.com/logrusorgru/aurora"
)

const logFilepath = "${HOME}/.dotbro/dotbro.log"

var (
	osfs = new(OSFS)
)

func main() {
	ctx := context.Background()

	// TODO:
	// 1. make slog output colored
	// 2. log both to file and stderr (without color)

	debugLogger := newDebugLogger(ctx)

	var outputer = NewOutputer(OutputerModeNormal, os.Stdout, nil)

	debugLogger.Debug("Start.")

	// Parse arguments

	args, err := ParseArguments(nil)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error parsing aruments: %s", err))
		exit(1, debugLogger)
	}

	debugLogger.Debug(fmt.Sprintf("Arguments passed: %+v", args))

	switch {
	case args["--verbose"].(bool):
		outputer.Mode = OutputerModeVerbose
	case args["--quiet"].(bool):
		outputer.Mode = OutputerModeQuiet
	default:
		outputer.Mode = OutputerModeNormal
	}

	// Process config

	configPaths := getConfigPath(ctx, args["--config"], debugLogger)

	for _, configPath := range configPaths {
		debugLogger.Debug(fmt.Sprintf("Parsing config file %s", configPath))
		config, err := NewConfiguration(configPath)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Cannot read configuration from file %s : %s.\n", configPath, err))
			slog.InfoContext(ctx, fmt.Sprintf("%s: Maybe you have renamed your config file?\nIf so, run dotbro with '--config' argument (see 'dotbro --help' for details).", Magenta("TIP")))
			exit(1, debugLogger)
		}

		// Preparations

		err = os.MkdirAll(config.Directories.Backup, 0700)
		if err != nil && !os.IsExist(err) {
			slog.ErrorContext(ctx, fmt.Sprintf("Error creating backup directory: %s", err))
			exit(1, debugLogger)
		}

		slog.DebugContext(ctx, fmt.Sprintf("Dotfiles root: %s", Brown(config.Directories.Dotfiles)))
		slog.DebugContext(ctx, fmt.Sprintf("Dotfiles src: %s", Brown(config.Directories.Sources)))
		slog.DebugContext(ctx, fmt.Sprintf("Destination dir: %s", Brown(config.Directories.Destination)))

		// Select action

		switch {
		case args["add"]:
			filename := args["<filename>"].(string)
			if err = addAction(ctx, filename, config); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("%s", err))
				exit(1, debugLogger)
			}

			slog.InfoContext(ctx, fmt.Sprintf("\n%s was successfully added to your dotfiles!", Brown(filename)))
			exit(0, debugLogger)
		case args["clean"]:
			if err = cleanAction(ctx, config); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("%s", err))
				exit(1, debugLogger)
			}

			slog.InfoContext(ctx, "\nCleaned!")
			exit(0, debugLogger)
		default:
			// Default action: install
			if err = installAction(ctx, config, debugLogger); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("%s", err))
				exit(1, debugLogger)
			}

			slog.InfoContext(ctx, "\nAll done (─‿‿─)")
			exit(0, debugLogger)
		}
	}
}

func addAction(ctx context.Context, filename string, config *Configuration) error {
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

	slog.DebugContext(ctx, fmt.Sprintf("Adding file %s to dotfiles root %s", Brown(filename), Brown(config.Directories.Dotfiles)))

	// backup file
	backupPath := config.Directories.Backup + "/" + path.Base(filename)
	if err = Copy(osfs, filename, backupPath); err != nil {
		return fmt.Errorf("Cannot backup file %s: %s", filename, err)
	}
	slog.InfoContext(ctx, fmt.Sprintf("  %s backup %s to %s", Green("→"), Brown(filename), Brown(backupPath)))

	// Move file to dotfiles root
	newPath := config.Directories.Dotfiles + "/" + path.Base(filename)
	if err = os.Rename(filename, newPath); err != nil {
		return err
	}

	linker := NewLinker(osfs)

	// Add a symlink to the moved file
	if err = linker.SetSymlink(newPath, filename); err != nil {
		return err
	}

	// TODO: write to config file

	return nil
}

func cleanAction(ctx context.Context, config *Configuration) error {
	cleaner := NewCleaner(osfs)
	if err := cleaner.CleanDeadSymlinks(ctx, config.Directories.Destination); err != nil {
		return fmt.Errorf("Error cleaning dead symlinks: %s", err)
	}

	return nil
}

func installAction(ctx context.Context, config *Configuration, logger *slog.Logger) error {
	// Default action: install
	cleaner := NewCleaner(osfs)
	err := cleaner.CleanDeadSymlinks(ctx, config.Directories.Destination)
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

	mapping := getMapping(ctx, config, srcDirAbs, logger)
	linker := NewLinker(osfs)

	slog.InfoContext(ctx, "--> Installing dotfiles...")

	// filter mapping:
	// - non-existent files
	// - already installed files
	filterMapping(mapping, func(src, dst string) bool {
		srcAbs := path.Join(srcDirAbs, src)
		destAbs := path.Join(config.Directories.Destination, dst)
		if _, err := osfs.Stat(srcAbs); err != nil {
			if osfs.IsNotExist(err) {
				slog.WarnContext(ctx, fmt.Sprintf("Source file %s does not exist", srcAbs))
				return false
			}

			slog.ErrorContext(ctx, fmt.Sprintf("Error processing source file %s: %s", src, err))
			exit(1, logger)
		}
		needSymlink, err := linker.NeedSymlink(ctx, srcAbs, destAbs)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error processing destination file %s: %s", destAbs, err))
			exit(1, logger)
		}
		return needSymlink
	})

	if len(mapping) == 0 {
		return nil
	}

	slog.InfoContext(ctx, fmt.Sprintf("From %s to %s :", Brown(srcDirAbs), Brown(config.Directories.Destination)))
	for src, dst := range mapping {
		installDotfile(ctx, src, dst, linker, config, srcDirAbs, logger)
	}

	return nil
}

func getConfigPath(ctx context.Context, configArg any, logger *slog.Logger) []string {
	var configPath string
	if configArg != nil {
		configPath = configArg.(string)
	}

	rc := NewRC()

	// Always load RC file (ignore error if file doesn't exist)
	if err := rc.Load(); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error reading rc file: %s", err))
		exit(1, logger)
	}

	// If config param is not passed to dotbro, use paths from RC file.
	if configPath == "" {
		paths := rc.GetPaths()
		if len(paths) == 0 {
			slog.ErrorContext(ctx, "Config file not specified.")
			exit(1, logger)
		}

		slog.DebugContext(ctx, fmt.Sprintf("Got config paths from file %s", Brown(RCFilepath)))
		return paths
	}

	// Add new config path to RC file
	var err error
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Bad config path: %s", err))
		exit(1, logger)
	}

	rc.SetPath(configPath)

	if err = rc.Save(); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Cannot save rc file: %s", err))
		exit(1, logger)
	}

	slog.DebugContext(ctx, fmt.Sprintf("Saved config path to file %s", Brown(RCFilepath)))
	return []string{configPath}
}

func getMapping(ctx context.Context, config *Configuration, srcDirAbs string, logger *slog.Logger) map[string]string {
	mapping := make(map[string]string)

	if len(config.Mapping) == 0 {
		// install all the things
		slog.DebugContext(ctx, "Mapping is not specified - install all the things")
		dir, err := os.Open(srcDirAbs)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error reading dotfiles source dir: %s", err))
			exit(1, logger)
		}

		defer func() {
			if err = dir.Close(); err != nil {
				slog.WarnContext(ctx, fmt.Sprintf("Error closing dir %s: %s", srcDirAbs, err.Error()))
			}
		}()

		files, err := dir.Readdir(0)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error reading dotfiles source dir: %s", err))
			exit(1, logger)
		}

		for _, fileInfo := range files {
			mapping[fileInfo.Name()] = fileInfo.Name()
		}

		// filter excludes
		for _, exclude := range config.Files.Excludes {
			delete(mapping, exclude)
		}
	} else {
		// install by mapping
		if len(config.Files.Excludes) > 0 {
			slog.WarnContext(ctx, "Excludes in config make no sense when mapping is specified, omitting them.")
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

func installDotfile(ctx context.Context, src, dest string, linker Linker, config *Configuration, srcDirAbs string, logger *slog.Logger) {
	srcAbs := path.Join(srcDirAbs, src)
	destAbs := path.Join(config.Directories.Destination, dest)

	needBackup, err := linker.NeedBackup(destAbs)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error processing destination file %s: %s", destAbs, err))
		exit(1, logger)
	}

	if needBackup {
		oldpath := destAbs
		newpath := config.Directories.Backup + "/" + dest
		err = linker.Move(ctx, oldpath, newpath)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error on file backup %s: %s", oldpath, err))
			exit(1, logger)
		}
	}

	err = linker.SetSymlink(srcAbs, destAbs)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error creating symlink from %s to %s: %s", srcAbs, destAbs, err))
		exit(1, logger)
	}

	slog.InfoContext(ctx, fmt.Sprintf("  %s set symlink %s -> %s", Green("+"), Brown(src), Brown(dest)))
}

// exit actually calls os.Exit after logger logs exit message.
func exit(exitCode int, logger *slog.Logger) {
	logger.Debug(fmt.Sprintf("Exit with code %d.", exitCode))
	os.Exit(exitCode)
}
