# Makefile for go-services project
# This file provides various targets for building, testing, and managing the project

# Project configuration
PROJECT_NAME := "go-services"
PKG := "github.com/a-castellano/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

# Phony targets declaration
.PHONY: all build clean test test_integration test_messagebroker coverage coverhtml lint race msan

# Default target - builds the project
all: build

# Lint the Go source files using go vet
lint: ## Lint the files
	@go vet ./...

# Run unit tests only (excludes integration tests)
test: ## Run unit tests
	@go test --tags=unit_tests -short ./...

# Run integration tests only (requires external services like Redis and RabbitMQ)
test_integration: ## Run integration tests
	@go test --tags=integration_tests -short ./...

# Run all messagebroker-related tests (both unit and integration)
test_messagebroker: ## Run messagebroker related tests
	@go test --tags=messagebroker_tests -short ./...

# Run only messagebroker unit tests
test_messagebroker_unit: ## Run messagebroker unit tests
	@go test --tags=messagebroker_unit_tests -short ./...

# Run all memorydatabase-related tests (both unit and integration)
test_memorydatabase: ## Run memorydatabase related tests
	@go test --tags=memorydatabase_tests -short ./...

# Run only memorydatabase unit tests
test_memorydatabase_unit: ## Run memorydatabase unit tests
	@go test --tags=memorydatabase_unit_tests -short ./...

# Run tests with data race detector enabled
race: ## Run data race detector
	@go test -race -short ./...

# Run tests with memory sanitizer enabled
msan: ## Run memory sanitizer
	@go test -msan -short ./...

# Generate global code coverage report in text format
coverage: ## Generate global code coverage report
	./scripts/coverage.sh;

# Generate global code coverage report in HTML format
coverhtml: ## Generate global code coverage report in HTML
	./scripts/coverage.sh html;

# Build target (commented out as this is a library project)
#build: ## Build the binary file
#	@go build -v $(PKG)

# Clean up build artifacts
clean: ## Remove previous build
	@rm -f $(PROJECT_NAME)

# Display help information for all available targets
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
