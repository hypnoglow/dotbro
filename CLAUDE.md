# CLAUDE.md

Development instructions for Claude.

## Build, Test, and Lint

This project uses a Makefile for common development tasks. Always use these commands:

- **Build**: `make build` - compiles the project to `./bin/dotbro`
- **Test**: `make test` - runs all tests
- **Lint**: `make lint` - runs golangci-lint
- **Dependencies**: `make deps` - tidies Go modules
- **Clean**: `make clean` - removes built binaries

### Important
Always run `make test` before committing changes to ensure all tests pass.
