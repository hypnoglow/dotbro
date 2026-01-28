# dotbro - simple yet effective dotfiles manager.

[![CI](https://github.com/hypnoglow/dotbro/actions/workflows/ci.yaml/badge.svg)](https://github.com/hypnoglow/dotbro/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hypnoglow/dotbro)](https://goreportcard.com/report/github.com/hypnoglow/dotbro)

*Dotbro* is a tool which helps you install and keep your dotfiles up to date.

## tl;dr

1. Create simple [profile](#profiles).

2. Run dotbro first time:
    
    dotbro --config path/to/your/dotbro.toml
    
3. Next time just execute:

    dotbro

## Dotfiles? What?

Read about dotfiles on [GitHub page](https://dotfiles.github.io/).
I think [this article](https://webpro.nl/articles/getting-started-with-dotfiles) by *Lars Kappert* will give you enough understanding.
So, if you don't have your dotfiles repository yet, it's time to create it. Next, read further to explore an easy way to manage your dotfiles. 

## Motivation

Dotfiles are generally stored in VCS and symlinked from repo directory to your `$HOME` - this is one of the best patterns because you can track changes easily and commit them to your dotfiles repo. However, this pattern does not offer you any way to install your dotfiles, so often people end up writing their own script e.g. in bash, which is not good for long-term purposes (I know that because I had one).

This tool was made to deal with dotfiles installation, so you don't waste your time writing your install scripts and focus only on your dotfiles themselves.

*Dotbro* takes on the routine. The main task - installing your dotfiles in one command on any of your machines.

## Features

### Simple Configuration

All you need is simple [profile](#profiles) in JSON or TOML format.

The extra benefit is that you do not need any special tooling if you use multiple different operation systems, e.g Linux and macOS.
You can use one single dotfiles repository with multiple dotbro profiles inside - one for each OS.
What can be easier?

### Explicit Mapping

You may want to (or you do already) store your dotfiles in a neat way using named directories like `bash/bashrc`.
Obviously, you want to symlink it to proper place `$HOME/.bashrc`.
This is easily done by writing such string in `[mapping]` section:

```toml
"bash/bashrc" = ".bashrc"
```

### Smart Profiles

First time you run dotbro, specify the profile (`dotbro.toml` file you created).
Dotbro remembers path to this file and use it in further runs.

### Automatic Cleanup

Dotbro cleans broken symlinks in your destination path (`$HOME` by default).

### `add` command

Dotbro can automate routine of adding files to your dotfiles repo with one single
command. It does a backup copy, moves the file and creates a symlink to your file.
After that you only need to add this file to your dotbro profile (*I'm working on automation of this*) and commit that file to your repo.

## Installation

### [Go](https://go.dev/doc/install) toolchain

    go install github.com/hypnoglow/dotbro@latest

This downloads the source code, builds and installs the latest version of dotbro.
Then you can use `dotbro` command right away.

### Arch Linux

`dotbro` package is available in AUR:

https://aur.archlinux.org/packages/dotbro/

### Precompiled binary

Coming soon ...

## Configuration

### Profiles

*Profile* is a typical [dotbro.toml](dotbro.toml) (or `dotbro.json`) file. It defines directories and mappings.

You can have multiple profiles for different purposes. For example, you can have a profile for your work environment and another for your personal environment.

Profile can be either TOML or JSON file.
TOML is peferred, because it's a bit clearer and allows comments.
However, JSON is good option for profiles without explicit mapping, it's short and simple.

Example of a simple profile in TOML format:

https://github.com/hypnoglow/dotbro/blob/2493e151a4ed08b60ab7d1266a36dde23349b258/profile.toml#L1-L75

See more examples in [config_examples](https://github.com/hypnoglow/dotbro/tree/master/config_examples) directory of this repo.

#### Options

Profile has 3 sections:
- directories
- mapping
- files

#### Directories

Option | Description | Example | Default
--- | --- | --- | ---
dotfiles | Directory of your dotfiles repository. | `$HOME/dotfiles` | Directory of your profile file.
sources | Directory relative to `dotfiles` where dotfiles are stored. You want to set this option if you keep your dotfiles in a subdirectory of your repo. By default this is empty, assuming your dotfiles are on the first level of `dotfiles` directory. | `src` | none
destination | Your dotfiles will be linked there. | `$HOME` | `$HOME`
backup | Your original files will be backuped there. | `$HOME/backups/dotfiles` | `$HOME/.dotfiles~`

#### Mapping

Each option here represents source file and destination file.  
Example: your dotfiles directory is `$HOME/dotfiles`. In that directory, you have folder `vim` and file `vimrc` in that folder, so path is `$HOME/dotfiles/vim/vimrc`.
In `directories` section you have already specified `dotfiles = "$HOME/dotfiles"`. So to install your `vimrc` properly you need to specify such line in mapping section:
```
"vim/vimrc" = ".vimrc"
```

Also, mapping is optional. If you do not specify any mapping, `dotbro` will symlink all files from your dotfiles directory to your destination directory respectively. If you do want this approach, but want some files to be excluded, see [Files](#files) section.

#### Files

As said above, this section is for symlinking all dotfiles without mapping specification.

Option | Description | Example | Default
--- | --- | --- | ---
excludes | Files to exclude from being installed | `excludes = ["README.md", "dotbro.toml"]` | none

Summing up, your profile without mapping will look like this:

```toml
[directories]

dotfiles = "$HOME/dotfiles"

[mapping]

excludes = [
    "README.md",
    "dotbro.toml"
]
```

## Usage

Take a look at usage info running:

    dotbro --help

If you haven't prepared your `dotbro.toml` yet, it's time to do it.
When your profile is ready, run:

    dotbro -c <path/to/dotbro.toml>

This installs your dotfiles.

Further runs you can omit profile - dotbro have remembered it for you.
So just run:

    dotbro

To move a file to your dotfiles, perform an `add` command:

    dotbro add ./path-to-file

## Issues

If you experience any problems, please submit an issue and attach dotbro log file,
which can be found at `$HOME/.dotbro/dotbro.log`.

## License

[MIT](https://github.com/hypnoglow/dotbro/blob/master/LICENSE.md)
