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

// App is the main application structure.
type App struct {
	logger *slog.Logger
	config *Configuration
}

func main() {
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

	app := &App{
		logger: newConsoleLogger(logLevel),
	}
	app.Run(args)
}

// Run runs the main application logic.
func (a *App) Run(args map[string]any) {
	ctx := context.Background()

	a.logger.DebugContext(ctx, "Start")
	a.logger.DebugContext(ctx, "Arguments passed", slog.Any("args", args))

	// Process config
	configPaths := a.getConfigPath(ctx, args["--config"])

	for _, configPath := range configPaths {
		a.logger.DebugContext(ctx, "Loading config file", slog.String("path", configPath))
		var err error
		a.config, err = NewConfiguration(configPath)
		if err != nil {
			a.logger.ErrorContext(ctx, "Cannot read configuration from file", slog.String("path", configPath), slog.Any("error", err))
			a.logger.InfoContext(ctx, "Maybe you have renamed your config file?\nIf so, run dotbro with '--config' argument (see 'dotbro --help' for details).", slog.String("tip", "TIP"))
			a.exit(1)
		}

		// Preparations
		err = os.MkdirAll(a.config.Directories.Backup, 0700)
		if err != nil && !os.IsExist(err) {
			a.logger.ErrorContext(ctx, "Error creating backup directory", slog.Any("error", err))
			a.exit(1)
		}

		a.logger.DebugContext(ctx, "Config directories",
			slog.String("dotfiles", a.config.Directories.Dotfiles),
			slog.String("sources", a.config.Directories.Sources),
			slog.String("destination", a.config.Directories.Destination),
			slog.String("backup", a.config.Directories.Backup))

		// Select action
		switch {
		case args["add"]:
			// TODO: add support for multiple configs
			filename := args["<filename>"].(string)
			if err = a.addAction(ctx, filename); err != nil {
				a.logger.ErrorContext(ctx, "Add action failed", slog.Any("error", err))
				a.exit(1)
			}

			a.logger.InfoContext(ctx, "File was successfully added to your dotfiles!", slog.String("path", filename))
			a.logger.InfoContext(ctx, "All done (─‿‿─)")
			a.exit(0)
		case args["clean"]:
			// TODO: add support for multiple configs
			if err = a.cleanAction(ctx); err != nil {
				a.logger.ErrorContext(ctx, "Clean action failed", slog.Any("error", err))
				a.exit(1)
			}

			a.logger.InfoContext(ctx, "Cleaned!")
			a.logger.InfoContext(ctx, "All done (─‿‿─)")
			a.exit(0)
		default:
			// Default action: install
			if err = a.installAction(ctx); err != nil {
				a.logger.ErrorContext(ctx, "Install action failed", slog.Any("error", err))
				a.exit(1)
			}
		}
	}

	a.logger.InfoContext(ctx, "All done (─‿‿─)")
	a.exit(0)
}

func (a *App) addAction(ctx context.Context, filename string) error {
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

	a.logger.DebugContext(ctx, "Adding file to dotfiles root",
		slog.String("src", filename),
		slog.String("dst", a.config.Directories.Dotfiles))

	// backup file
	backupPath := a.config.Directories.Backup + "/" + path.Base(filename)
	if err = Copy(osfs, filename, backupPath); err != nil {
		return fmt.Errorf("Cannot backup file %s: %s", filename, err)
	}
	a.logger.InfoContext(ctx, "backup",
		slog.String("status", "→"),
		slog.String("src", filename),
		slog.String("dst", backupPath))

	// Move file to dotfiles root
	newPath := a.config.Directories.Dotfiles + "/" + path.Base(filename)
	if err = os.Rename(filename, newPath); err != nil {
		return err
	}

	linker := NewLinker(osfs, a.logger)

	// Add a symlink to the moved file
	if err = linker.SetSymlink(newPath, filename); err != nil {
		return err
	}

	// TODO: write to config file

	return nil
}

func (a *App) cleanAction(ctx context.Context) error {
	cleaner := NewCleaner(osfs, a.logger)
	if err := cleaner.CleanDeadSymlinks(ctx, a.config.Directories.Destination); err != nil {
		return fmt.Errorf("Error cleaning dead symlinks: %s", err)
	}

	return nil
}

func (a *App) installAction(ctx context.Context) error {
	// Default action: install
	cleaner := NewCleaner(osfs, a.logger)
	err := cleaner.CleanDeadSymlinks(ctx, a.config.Directories.Destination)
	if err != nil {
		return fmt.Errorf("Error cleaning dead symlinks: %s", err)
	}

	srcDirAbs := a.config.Directories.Dotfiles
	if a.config.Directories.Sources != "" {
		if _, err = os.Stat(a.config.Directories.Sources); os.IsNotExist(err) {
			return fmt.Errorf("Sources directory `%s' does not exist.", a.config.Directories.Sources)
		}
		if err != nil {
			return fmt.Errorf("Error reading sources directory `%s': %s", a.config.Directories.Sources, err)
		}
		srcDirAbs += "/" + a.config.Directories.Sources
	}

	mapping := a.getMapping(ctx, srcDirAbs)
	linker := NewLinker(osfs, a.logger)

	a.logger.InfoContext(ctx, "--> Installing dotfiles...", slog.String("config", a.config.Filepath))

	// filter mapping:
	// - non-existent files
	// - already installed files
	filterMapping(mapping, func(src, dst string) bool {
		srcAbs := path.Join(srcDirAbs, src)
		destAbs := path.Join(a.config.Directories.Destination, dst)
		if _, err := osfs.Stat(srcAbs); err != nil {
			if osfs.IsNotExist(err) {
				a.logger.WarnContext(ctx, "Source file does not exist", slog.String("path", srcAbs))
				return false
			}

			a.logger.ErrorContext(ctx, "Error processing source file", slog.String("path", src), slog.Any("error", err))
			a.exit(1)
		}
		needSymlink, err := linker.NeedSymlink(ctx, srcAbs, destAbs)
		if err != nil {
			a.logger.ErrorContext(ctx, "Error processing destination file", slog.String("path", destAbs), slog.Any("error", err))
			a.exit(1)
		}
		return needSymlink
	})

	if len(mapping) == 0 {
		return nil
	}

	a.logger.InfoContext(ctx, "Installing dotfiles",
		slog.String("src", srcDirAbs),
		slog.String("dst", a.config.Directories.Destination))
	for src, dst := range mapping {
		a.installDotfile(ctx, src, dst, linker, srcDirAbs)
	}

	return nil
}

func (a *App) getConfigPath(ctx context.Context, configArg any) []string {
	var configPath string
	if configArg != nil {
		configPath = configArg.(string)
	}

	cfg := NewConfig(a.logger)

	if err := cfg.Load(ctx); err != nil {
		a.logger.ErrorContext(ctx, "Error reading config file", slog.Any("error", err))
		a.exit(1)
	}

	// If config param is not passed to dotbro, use paths from config file.
	if configPath == "" {
		paths := cfg.GetProfilePaths()
		if len(paths) == 0 {
			a.logger.ErrorContext(ctx, "Config file not specified.")
			a.exit(1)
		}

		a.logger.DebugContext(ctx, "Using config paths", slog.Int("count", len(paths)))
		for i, p := range paths {
			a.logger.DebugContext(ctx, "Config path", slog.Int("index", i+1), slog.String("path", p))
		}
		return paths
	}

	// Add new config path to state config file
	var err error
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		a.logger.ErrorContext(ctx, "Bad config path", slog.Any("error", err))
		a.exit(1)
	}

	cfg.AddProfile(configPath)

	if err = cfg.Save(ctx); err != nil {
		a.logger.ErrorContext(ctx, "Cannot save config file", slog.Any("error", err))
		a.exit(1)
	}
	return []string{configPath}
}

func (a *App) getMapping(ctx context.Context, srcDirAbs string) map[string]string {
	mapping := make(map[string]string)

	if len(a.config.Mapping) == 0 {
		// install all the things
		a.logger.DebugContext(ctx, "Mapping is not specified - install all the things")
		dir, err := os.Open(srcDirAbs)
		if err != nil {
			a.logger.ErrorContext(ctx, "Error reading dotfiles source dir", slog.Any("error", err))
			a.exit(1)
		}

		defer func() {
			if err = dir.Close(); err != nil {
				a.logger.WarnContext(ctx, "Error closing dir", slog.String("path", srcDirAbs), slog.Any("error", err))
			}
		}()

		files, err := dir.Readdir(0)
		if err != nil {
			a.logger.ErrorContext(ctx, "Error reading dotfiles source dir", slog.Any("error", err))
			a.exit(1)
		}

		for _, fileInfo := range files {
			mapping[fileInfo.Name()] = fileInfo.Name()
		}

		// filter excludes
		for _, exclude := range a.config.Files.Excludes {
			delete(mapping, exclude)
		}
	} else {
		// install by mapping
		if len(a.config.Files.Excludes) > 0 {
			a.logger.WarnContext(ctx, "Excludes in config make no sense when mapping is specified, omitting them.")
		}

		mapping = a.config.Mapping
	}

	return mapping
}

func (a *App) installDotfile(ctx context.Context, src, dest string, linker Linker, srcDirAbs string) {
	srcAbs := path.Join(srcDirAbs, src)
	destAbs := path.Join(a.config.Directories.Destination, dest)

	needBackup, err := linker.NeedBackup(destAbs)
	if err != nil {
		a.logger.ErrorContext(ctx, "Error processing destination file", slog.String("path", destAbs), slog.Any("error", err))
		a.exit(1)
	}

	if needBackup {
		oldpath := destAbs
		newpath := a.config.Directories.Backup + "/" + dest
		err = linker.Move(ctx, oldpath, newpath)
		if err != nil {
			a.logger.ErrorContext(ctx, "Error on file backup", slog.String("path", oldpath), slog.Any("error", err))
			a.exit(1)
		}
	}

	err = linker.SetSymlink(srcAbs, destAbs)
	if err != nil {
		a.logger.ErrorContext(ctx, "Error creating symlink", slog.String("src", srcAbs), slog.String("dst", destAbs), slog.Any("error", err))
		a.exit(1)
	}

	a.logger.InfoContext(ctx, "set symlink",
		slog.String("status", "+"),
		slog.String("src", src),
		slog.String("dst", dest))
}

// exit actually calls os.Exit after logger logs exit message.
func (a *App) exit(exitCode int) {
	a.logger.Debug("Exit", slog.Int("code", exitCode))
	os.Exit(exitCode)
}

func filterMapping(mapping map[string]string, callback func(src, dst string) bool) {
	for src, dst := range mapping {
		if !callback(src, dst) {
			delete(mapping, src)
		}
	}
}
