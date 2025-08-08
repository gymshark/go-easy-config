# Makefile for go-easy-config project

.PHONY: all build clean test fmt setup

all: setup test build

setup:
	@echo "Setting up project..."
	@go mod tidy

test: setup
	@echo "Running tests..."
	@go test ./... -v -race

test-bench: setup
	@echo "Running benchmarks..."
	@go test -bench . -benchmem

fmt:
	@echo "Formatting Go code..."
	@gofmt -s -w -l .
	@echo "Formatting complete."

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf ./bin
	@rm -rf ./terraform/bin
	@echo "Clean complete."

