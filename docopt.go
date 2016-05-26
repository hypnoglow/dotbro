package main

import "github.com/docopt/docopt-go"

const version = "0.0.1"

func parseArguments() (map[string]interface{}, error) {
	usage := `dotbro - simple yet effective dotfiles manager.

Usage:
  dotbro [options] [--config=<filepath>]
  dotbro -h | --help
  dotbro --version

Options:
  -c --config=<filepath>  Dotbro's configuration file in JSON or TOML format.
  -h --help               Show this helpful info.
  -q --quiet              Quiet mode. Do not print any output, except warnings
                          and errors.
  -v --verbose            Verbose mode. Detailed output.
  -V --version            Show version.
`

	return docopt.Parse(usage, nil, true, "dotbro "+version, false)
}
