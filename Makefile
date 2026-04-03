SHELL := /bin/bash
.DEFAULT_GOAL := help

.PHONY: help build fmt test

help:
	@printf "Available targets:\n"
	@printf "  make build                      Build the harbour binary\n"
	@printf "  make fmt                        Format the Go source\n"
	@printf "  make test                       Run the Go tests\n"

build:
	mkdir -p bin
	GOOS=darwin GOARCH=arm64 go build -o bin/harbour ./cmd/harbour

fmt:
	gofmt -w ./cmd/harbour/*.go

test:
	go test ./...
