package main

import (
	"os"
	"path/filepath"
)

var isVerbose bool
var isQuiet bool

func main() {
	logger.msg("Start.")

	checkVersion()

	// Parse arguments

	args, err := parseArguments()
	if err != nil {
		outError("Error parsing aruments: %s", err)
		exit(1)
	}

	logger.msg("Arguments passed: %+v", args)

	isVerbose = args["--verbose"].(bool)
	isQuiet = args["--quiet"].(bool)

	var configPath string
	if args["--config"] == nil {
		configPath = ""
	} else {
		configPath = args["--config"].(string)
	}

	// Process arguments
	var rc RC
	if configPath == "" {
		rc, err = readRC()
		if err != nil {
			outError("Error reading rc file: %s", err)
			exit(1)
		}
		if rc.Config.Path == "" {
			outError("Config file not specified.")
			exit(1)
		}
	} else {
		// Save to RC file
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			outError("Bad config path: %s", err)
			exit(1)
		}

		rc, err = saveRC(configPath)
		if err != nil {
			outError("Cannot save rc file: %s", err)
			exit(1)
		}
	}

	// Process config

	config, err := configurationFromFile(rc.Config.Path)
	if err != nil {
		outError("Error reading configuration from file %s: %s", rc.Config.Path, err)
		exit(1)
	}

	// Preparations

	err = os.MkdirAll(config.Directories.Backup, 0755)
	if err != nil && !os.IsExist(err) {
		outError("Error creating backup directory: %s", err)
		exit(1)
	}

	if isVerbose {
		outVerbose("Dotfiles root: %s", config.Directories.Dotfiles)
		outVerbose("Dotfiles src: %s", config.Directories.Sources)
		outVerbose("Destination dir: %s", config.Directories.Destination)
	}

	err = cleanDeadSymlinks(config.Directories.Destination)
	if err != nil {
		outError("Error cleaning dead symlinks: %s", err)
		exit(1)
	}

	srcDirAbs := config.Directories.Dotfiles
	if config.Directories.Sources != "" {
		if _, err = os.Stat(config.Directories.Sources); os.IsNotExist(err) {
			outError("Sources directory `%s' does not exist.", config.Directories.Sources)
			exit(1)
		}
		if err != nil {
			outError("Error reading sources directory `%s': %s", config.Directories.Sources, err)
			exit(1)
		}
		srcDirAbs += "/" + config.Directories.Sources
	}

	mapping := make(map[string]string)

	if len(config.Mapping) == 0 {
		// install all the things
		outVerbose("Mapping is not specified - install all the things")
		dir, err := os.Open(srcDirAbs)
		if err != nil {
			outError("Error reading dotfiles source dir: %s", err)
			exit(1)
		}

		defer dir.Close()

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

	outInfo("Installing dotfiles...")
	for src, dest := range mapping {
		srcAbs := srcDirAbs + "/" + src
		destAbs := config.Directories.Destination + "/" + dest

		exists, err := isExists(srcAbs)
		if !exists {
			outWarn("Source file %s does not exist", srcAbs)
			continue
		}

		if err != nil {
			outError("Error processing source file %s: %s", src, err)
			exit(1)
		}

		needSymlink, needBackup, err := processDest(srcAbs, destAbs)
		if err != nil {
			outError("Error processing destination file %s: %s", destAbs, err)
			exit(1)
		}

		if !needSymlink {
			continue
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

	outInfo("All done (─‿‿─)")
	exit(0)
}
