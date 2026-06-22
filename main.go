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
	logger  *slog.Logger
	profile *Profile
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
func (app *App) Run(args map[string]any) {
	ctx := context.Background()

	app.logger.DebugContext(ctx, "Start")
	app.logger.DebugContext(ctx, "Arguments passed", slog.Any("args", args))

	if args["add"].(bool) {
		profilePath := app.getCurrentProfilePath(ctx, args["--config"])
		if err := app.loadProfile(ctx, profilePath); err != nil {
			app.logger.ErrorContext(ctx, "Cannot read profile", slog.String("path", profilePath), slog.Any("error", err))
			app.logger.InfoContext(ctx, "Maybe you have renamed your profile file?\nIf so, run dotbro with '--config' argument (see 'dotbro --help' for details).", slog.String("tip", "TIP"))
			app.exit(1)
		}

		filename := args["<filename>"].(string)
		if err := app.addAction(ctx, filename); err != nil {
			app.logger.ErrorContext(ctx, "Add action failed", slog.Any("error", err))
			app.exit(1)
		}

		app.logger.InfoContext(ctx, "File was successfully added to your dotfiles!", slog.String("path", filename))
		app.logger.InfoContext(ctx, "All done (─‿‿─)")
		app.exit(0)
	}

	// Process profiles
	profilePaths := app.getProfilePaths(ctx, args["--config"])

	for _, profilePath := range profilePaths {
		if err := app.loadProfile(ctx, profilePath); err != nil {
			app.logger.ErrorContext(ctx, "Cannot read profile", slog.String("path", profilePath), slog.Any("error", err))
			app.logger.InfoContext(ctx, "Maybe you have renamed your profile file?\nIf so, run dotbro with '--config' argument (see 'dotbro --help' for details).", slog.String("tip", "TIP"))
			app.exit(1)
		}

		// Select action
		switch {
		case args["clean"]:
			// TODO: add support for multiple configs
			if err := app.cleanAction(ctx); err != nil {
				app.logger.ErrorContext(ctx, "Clean action failed", slog.Any("error", err))
				app.exit(1)
			}

			app.logger.InfoContext(ctx, "Cleaned!")
			app.logger.InfoContext(ctx, "All done (─‿‿─)")
			app.exit(0)
		default:
			// Default action: install
			if err := app.installAction(ctx); err != nil {
				app.logger.ErrorContext(ctx, "Install action failed", slog.Any("error", err))
				app.exit(1)
			}
		}
	}

	app.logger.InfoContext(ctx, "All done (─‿‿─)")
	app.exit(0)
}

func (app *App) cleanAction(ctx context.Context) error {
	cleaner := NewCleaner(osfs, app.logger)
	if err := cleaner.CleanDeadSymlinks(ctx, app.profile.DestinationDir()); err != nil {
		return fmt.Errorf("Error cleaning dead symlinks: %s", err)
	}

	return nil
}

func (app *App) installAction(ctx context.Context) error {
	// Default action: install
	cleaner := NewCleaner(osfs, app.logger)
	err := cleaner.CleanDeadSymlinks(ctx, app.profile.DestinationDir())
	if err != nil {
		return fmt.Errorf("Error cleaning dead symlinks: %s", err)
	}

	srcDirAbs := app.profile.DotfilesDir()
	if app.profile.SourcesDir() != "" {
		if _, err = os.Stat(app.profile.SourcesDir()); os.IsNotExist(err) {
			return fmt.Errorf("Sources directory `%s' does not exist.", app.profile.SourcesDir())
		}
		if err != nil {
			return fmt.Errorf("Error reading sources directory `%s': %s", app.profile.SourcesDir(), err)
		}
		srcDirAbs += "/" + app.profile.SourcesDir()
	}

	mapping := app.getMapping(ctx, srcDirAbs)
	linker := NewLinker(osfs, app.logger)

	app.logger.InfoContext(ctx, "--> Installing dotfiles...", slog.String("profile", app.profile.Filepath()))

	// filter mapping:
	// - non-existent files
	// - already installed files
	filterMapping(mapping, func(src, dst string) bool {
		srcAbs := path.Join(srcDirAbs, src)
		destAbs := path.Join(app.profile.DestinationDir(), dst)
		if _, err := osfs.Stat(srcAbs); err != nil {
			if osfs.IsNotExist(err) {
				app.logger.WarnContext(ctx, "Source file does not exist", slog.String("path", srcAbs))
				return false
			}

			app.logger.ErrorContext(ctx, "Error processing source file", slog.String("path", src), slog.Any("error", err))
			app.exit(1)
		}
		needSymlink, err := linker.NeedSymlink(ctx, srcAbs, destAbs)
		if err != nil {
			app.logger.ErrorContext(ctx, "Error processing destination file", slog.String("path", destAbs), slog.Any("error", err))
			app.exit(1)
		}
		return needSymlink
	})

	if len(mapping) == 0 {
		return nil
	}

	app.logger.InfoContext(ctx, "Installing dotfiles",
		slog.String("src", srcDirAbs),
		slog.String("dst", app.profile.DestinationDir()))
	for src, dsts := range mapping {
		for _, dst := range dsts {
			app.installDotfile(ctx, src, dst, linker, srcDirAbs)
		}
	}

	return nil
}

func (app *App) getProfilePaths(ctx context.Context, profileArg any) []string {
	var profilePath string
	if profileArg != nil {
		profilePath = profileArg.(string)
	}

	cfg := NewConfig(
		app.logger,
		defaultConfigFilepath,
		defaultLegacyConfigFilepath,
	)

	if err := cfg.Load(ctx); err != nil {
		app.logger.ErrorContext(ctx, "Error reading config", slog.Any("error", err))
		app.exit(1)
	}

	// If profile path is not passed to dotbro, use paths from config.
	if profilePath == "" {
		paths := cfg.GetProfilePaths()
		if len(paths) == 0 {
			app.logger.ErrorContext(ctx, "Profile not specified.")
			app.exit(1)
		}

		app.logger.DebugContext(ctx, "Using profile paths from config", slog.Int("count", len(paths)))
		for i, p := range paths {
			app.logger.DebugContext(ctx, "Profile path", slog.Int("index", i+1), slog.String("path", p))
		}
		return paths
	}

	// Add new profile path to config
	var err error
	profilePath, err = filepath.Abs(profilePath)
	if err != nil {
		app.logger.ErrorContext(ctx, "Bad profile path", slog.Any("error", err))
		app.exit(1)
	}

	cfg.AddProfile(profilePath)

	if err = cfg.Save(ctx); err != nil {
		app.logger.ErrorContext(ctx, "Cannot save config", slog.Any("error", err))
		app.exit(1)
	}
	return []string{profilePath}
}

func (app *App) getMapping(ctx context.Context, srcDirAbs string) map[string]Destinations {
	if len(app.profile.Data().Mapping) == 0 {
		// install all the things
		app.logger.DebugContext(ctx, "Mapping is not specified - install all the things")
		dir, err := os.Open(srcDirAbs)
		if err != nil {
			app.logger.ErrorContext(ctx, "Error reading dotfiles source dir", slog.Any("error", err))
			app.exit(1)
		}

		defer func() {
			if err = dir.Close(); err != nil {
				app.logger.WarnContext(ctx, "Error closing dir", slog.String("path", srcDirAbs), slog.Any("error", err))
			}
		}()

		files, err := dir.Readdir(0)
		if err != nil {
			app.logger.ErrorContext(ctx, "Error reading dotfiles source dir", slog.Any("error", err))
			app.exit(1)
		}

		mapping := make(map[string]Destinations, len(files))
		for _, fileInfo := range files {
			mapping[fileInfo.Name()] = Destinations{fileInfo.Name()}
		}

		// filter excludes
		for _, exclude := range app.profile.Data().Files.Excludes {
			delete(mapping, exclude)
		}

		return mapping
	}

	// install by mapping
	if len(app.profile.Data().Files.Excludes) > 0 {
		app.logger.WarnContext(ctx, "Excludes in config make no sense when mapping is specified, omitting them.")
	}

	return app.profile.Data().Mapping
}

func (app *App) installDotfile(ctx context.Context, src, dest string, linker Linker, srcDirAbs string) {
	srcAbs := path.Join(srcDirAbs, src)
	destAbs := path.Join(app.profile.DestinationDir(), dest)

	needBackup, err := linker.NeedBackup(destAbs)
	if err != nil {
		app.logger.ErrorContext(ctx, "Error processing destination file", slog.String("path", destAbs), slog.Any("error", err))
		app.exit(1)
	}

	if needBackup {
		oldpath := destAbs
		newpath := app.profile.BackupDir() + "/" + dest
		err = linker.Move(ctx, oldpath, newpath)
		if err != nil {
			app.logger.ErrorContext(ctx, "Error on file backup", slog.String("path", oldpath), slog.Any("error", err))
			app.exit(1)
		}
	}

	err = linker.SetSymlink(srcAbs, destAbs)
	if err != nil {
		app.logger.ErrorContext(ctx, "Error creating symlink", slog.String("src", srcAbs), slog.String("dst", destAbs), slog.Any("error", err))
		app.exit(1)
	}

	app.logger.InfoContext(ctx, "set symlink",
		slog.String("status", "+"),
		slog.String("src", src),
		slog.String("dst", dest))
}

// exit actually calls os.Exit after logger logs exit message.
func (app *App) exit(exitCode int) {
	app.logger.Debug("Exit", slog.Int("code", exitCode))
	os.Exit(exitCode)
}

func filterMapping(mapping map[string]Destinations, callback func(src, dst string) bool) {
	for src, dsts := range mapping {
		kept := dsts[:0]
		for _, dst := range dsts {
			if callback(src, dst) {
				kept = append(kept, dst)
			}
		}
		if len(kept) == 0 {
			delete(mapping, src)
		} else {
			mapping[src] = kept
		}
	}
}
