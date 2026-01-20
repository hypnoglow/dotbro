.PHONY: all
all: deps build

.PHONY: deps
deps:
	@go mod tidy

.PHONY: build
build:
	@go build -o ./bin/dotbro .

.PHONY: install
install:
	@go install .

.PHONY: test
test:
	@go test ./...

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: clean
clean:
	@rm -fv ./bin/*
