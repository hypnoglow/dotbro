package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
)

const logFilepath = "${HOME}/.dotbro/dotbro.log"

var (
	osfs = new(OSFS)
)

func main() {
	ctx := context.Background()

	// Parse arguments
	args, err := ParseArguments(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %s\n", err)
		os.Exit(1)
	}

	// Determine log level based on flags
	var logLevel slog.Level
	switch {
	case args["--verbose"].(bool):
		logLevel = slog.LevelDebug
	case args["--quiet"].(bool):
		logLevel = slog.LevelWarn
	default:
		logLevel = slog.LevelInfo
	}

	logger := newConsoleLogger(logLevel)
	logger.Debug("Start.")
	logger.Debug(fmt.Sprintf("Arguments passed: %+v", args))

	// Process config
	configPaths := getConfigPath(ctx, args["--config"], logger)

	for _, configPath := range configPaths {
		logger.Debug(fmt.Sprintf("Parsing config file %s", configPath))
		config, err := NewConfiguration(configPath)
		if err != nil {
			logger.ErrorContext(ctx, fmt.Sprintf("Cannot read configuration from file %s : %s.\n", configPath, err))
			logger.InfoContext(ctx, "Maybe you have renamed your config file?\nIf so, run dotbro with '--config' argument (see 'dotbro --help' for details).", slog.String("tip", "TIP"))
			exit(1, logger)
		}

		// Preparations
		err = os.MkdirAll(config.Directories.Backup, 0700)
		if err != nil && !os.IsExist(err) {
			logger.ErrorContext(ctx, fmt.Sprintf("Error creating backup directory: %s", err))
			exit(1, logger)
		}

		logger.DebugContext(ctx, "Dotfiles root", slog.String("path", config.Directories.Dotfiles))
		logger.DebugContext(ctx, "Dotfiles src", slog.String("path", config.Directories.Sources))
		logger.DebugContext(ctx, "Destination dir", slog.String("path", config.Directories.Destination))

		// Select action
		switch {
		case args["add"]:
			filename := args["<filename>"].(string)
			if err = addAction(ctx, filename, config, logger); err != nil {
				logger.ErrorContext(ctx, fmt.Sprintf("%s", err))
				exit(1, logger)
			}

			logger.InfoContext(ctx, "File was successfully added to your dotfiles!", slog.String("path", filename))
			exit(0, logger)
		case args["clean"]:
			if err = cleanAction(ctx, config, logger); err != nil {
				logger.ErrorContext(ctx, fmt.Sprintf("%s", err))
				exit(1, logger)
			}

			logger.InfoContext(ctx, "\nCleaned!")
			exit(0, logger)
		default:
			// Default action: install
			if err = installAction(ctx, config, logger); err != nil {
				logger.ErrorContext(ctx, fmt.Sprintf("%s", err))
				exit(1, logger)
			}

			logger.InfoContext(ctx, "\nAll done (─‿‿─)")
			exit(0, logger)
		}
	}
}

func addAction(ctx context.Context, filename string, config *Configuration, logger *slog.Logger) error {
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

	logger.DebugContext(ctx, "Adding file to dotfiles root",
		slog.String("src", filename),
		slog.String("dst", config.Directories.Dotfiles))

	// backup file
	backupPath := config.Directories.Backup + "/" + path.Base(filename)
	if err = Copy(osfs, filename, backupPath); err != nil {
		return fmt.Errorf("Cannot backup file %s: %s", filename, err)
	}
	logger.InfoContext(ctx, "backup",
		slog.String("status", "→"),
		slog.String("src", filename),
		slog.String("dst", backupPath))

	// Move file to dotfiles root
	newPath := config.Directories.Dotfiles + "/" + path.Base(filename)
	if err = os.Rename(filename, newPath); err != nil {
		return err
	}

	linker := NewLinker(osfs, logger)

	// Add a symlink to the moved file
	if err = linker.SetSymlink(newPath, filename); err != nil {
		return err
	}

	// TODO: write to config file

	return nil
}

func cleanAction(ctx context.Context, config *Configuration, logger *slog.Logger) error {
	cleaner := NewCleaner(osfs, logger)
	if err := cleaner.CleanDeadSymlinks(ctx, config.Directories.Destination); err != nil {
		return fmt.Errorf("Error cleaning dead symlinks: %s", err)
	}

	return nil
}

func installAction(ctx context.Context, config *Configuration, logger *slog.Logger) error {
	// Default action: install
	cleaner := NewCleaner(osfs, logger)
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
	linker := NewLinker(osfs, logger)

	logger.InfoContext(ctx, "--> Installing dotfiles...")

	// filter mapping:
	// - non-existent files
	// - already installed files
	filterMapping(mapping, func(src, dst string) bool {
		srcAbs := path.Join(srcDirAbs, src)
		destAbs := path.Join(config.Directories.Destination, dst)
		if _, err := osfs.Stat(srcAbs); err != nil {
			if osfs.IsNotExist(err) {
				logger.WarnContext(ctx, fmt.Sprintf("Source file %s does not exist", srcAbs))
				return false
			}

			logger.ErrorContext(ctx, fmt.Sprintf("Error processing source file %s: %s", src, err))
			exit(1, logger)
		}
		needSymlink, err := linker.NeedSymlink(ctx, srcAbs, destAbs)
		if err != nil {
			logger.ErrorContext(ctx, fmt.Sprintf("Error processing destination file %s: %s", destAbs, err))
			exit(1, logger)
		}
		return needSymlink
	})

	if len(mapping) == 0 {
		return nil
	}

	logger.InfoContext(ctx, "Installing dotfiles",
		slog.String("src", srcDirAbs),
		slog.String("dst", config.Directories.Destination))
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
		logger.ErrorContext(ctx, fmt.Sprintf("Error reading rc file: %s", err))
		exit(1, logger)
	}

	// If config param is not passed to dotbro, use paths from RC file.
	if configPath == "" {
		paths := rc.GetPaths()
		if len(paths) == 0 {
			logger.ErrorContext(ctx, "Config file not specified.")
			exit(1, logger)
		}

		logger.DebugContext(ctx, "Got config paths from file", slog.String("path", RCFilepath))
		return paths
	}

	// Add new config path to RC file
	var err error
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		logger.ErrorContext(ctx, fmt.Sprintf("Bad config path: %s", err))
		exit(1, logger)
	}

	rc.SetPath(configPath)

	if err = rc.Save(); err != nil {
		logger.ErrorContext(ctx, fmt.Sprintf("Cannot save rc file: %s", err))
		exit(1, logger)
	}

	logger.DebugContext(ctx, "Saved config path to file", slog.String("path", RCFilepath))
	return []string{configPath}
}

func getMapping(ctx context.Context, config *Configuration, srcDirAbs string, logger *slog.Logger) map[string]string {
	mapping := make(map[string]string)

	if len(config.Mapping) == 0 {
		// install all the things
		logger.DebugContext(ctx, "Mapping is not specified - install all the things")
		dir, err := os.Open(srcDirAbs)
		if err != nil {
			logger.ErrorContext(ctx, fmt.Sprintf("Error reading dotfiles source dir: %s", err))
			exit(1, logger)
		}

		defer func() {
			if err = dir.Close(); err != nil {
				logger.WarnContext(ctx, fmt.Sprintf("Error closing dir %s: %s", srcDirAbs, err.Error()))
			}
		}()

		files, err := dir.Readdir(0)
		if err != nil {
			logger.ErrorContext(ctx, fmt.Sprintf("Error reading dotfiles source dir: %s", err))
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
			logger.WarnContext(ctx, "Excludes in config make no sense when mapping is specified, omitting them.")
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
		logger.ErrorContext(ctx, fmt.Sprintf("Error processing destination file %s: %s", destAbs, err))
		exit(1, logger)
	}

	if needBackup {
		oldpath := destAbs
		newpath := config.Directories.Backup + "/" + dest
		err = linker.Move(ctx, oldpath, newpath)
		if err != nil {
			logger.ErrorContext(ctx, fmt.Sprintf("Error on file backup %s: %s", oldpath, err))
			exit(1, logger)
		}
	}

	err = linker.SetSymlink(srcAbs, destAbs)
	if err != nil {
		logger.ErrorContext(ctx, fmt.Sprintf("Error creating symlink from %s to %s: %s", srcAbs, destAbs, err))
		exit(1, logger)
	}

	logger.InfoContext(ctx, "set symlink",
		slog.String("status", "+"),
		slog.String("src", src),
		slog.String("dst", dest))
}

// exit actually calls os.Exit after logger logs exit message.
func exit(exitCode int, logger *slog.Logger) {
	logger.Debug(fmt.Sprintf("Exit with code %d.", exitCode))
	os.Exit(exitCode)
}
