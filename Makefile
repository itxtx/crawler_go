# Makefile for crawler_go project
# Minimum Chrome version required for JavaScript testing
CHROME_MIN_VERSION = 118

# Default Go test flags
GO_TEST_FLAGS = -v
GO_SHORT_FLAGS = -short

# Detect Chrome installation and set CHROME_PATH
ifeq ($(shell uname -s),Darwin)
	# macOS
	CHROME_CANDIDATES = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
						"/Applications/Chromium.app/Contents/MacOS/Chromium"
else ifeq ($(shell uname -s),Linux)
	# Linux
	CHROME_CANDIDATES = "/usr/bin/google-chrome" \
						"/usr/bin/google-chrome-stable" \
						"/usr/bin/chromium" \
						"/usr/bin/chromium-browser" \
						"/snap/bin/chromium"
else
	# Windows (through WSL or Git Bash)
	CHROME_CANDIDATES = "/mnt/c/Program Files/Google/Chrome/Application/chrome.exe" \
						"/mnt/c/Program Files (x86)/Google/Chrome/Application/chrome.exe"
endif

# Function to check if Chrome is available
define check_chrome
	$(eval CHROME_PATH := $(shell for chrome in $(CHROME_CANDIDATES); do \
		if [ -f "$$chrome" ] || [ -L "$$chrome" ]; then \
			echo "$$chrome"; \
			break; \
		fi; \
	done))
	$(if $(CHROME_PATH),,$(error Chrome/Chromium not found. Please install Chrome/Chromium $(CHROME_MIN_VERSION)+ for JavaScript testing))
endef

# Function to check Chrome version
define check_chrome_version
	$(eval CHROME_VERSION := $(shell "$(CHROME_PATH)" --version 2>/dev/null | grep -oE '[0-9]+' | head -1))
	$(if $(shell [ $(CHROME_VERSION) -ge $(CHROME_MIN_VERSION) ] && echo "ok"),,$(error Chrome version $(CHROME_VERSION) is too old. Minimum version $(CHROME_MIN_VERSION) required))
endef

.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.PHONY: check-chrome
check-chrome: ## Check if Chrome is installed and set CHROME_PATH
	$(call check_chrome)
	$(call check_chrome_version)
	@echo "✓ Chrome found at: $(CHROME_PATH)"
	@echo "✓ Chrome version: $(CHROME_VERSION) (minimum: $(CHROME_MIN_VERSION))"
	@echo "✓ CHROME_PATH=$(CHROME_PATH)"

.PHONY: export-chrome-path
export-chrome-path: check-chrome ## Export CHROME_PATH environment variable
	$(call check_chrome)
	@echo "export CHROME_PATH='$(CHROME_PATH)'"

.PHONY: test
test: ## Run basic tests (short mode, skips JS tests)
	@echo "Running tests in short mode (skipping JavaScript tests)..."
	go test $(GO_TEST_FLAGS) $(GO_SHORT_FLAGS) ./...

.PHONY: test-js
test-js: check-chrome ## Run JavaScript tests (requires Chrome)
	$(call check_chrome)
	@echo "Running JavaScript tests with Chrome at: $(CHROME_PATH)"
	@export CHROME_PATH="$(CHROME_PATH)" && go test $(GO_TEST_FLAGS) ./tests -run ".*JS.*|.*JavaScript.*" -timeout 30s

.PHONY: test-jsengine
test-jsengine: check-chrome ## Run JavaScript engine tests specifically
	$(call check_chrome)
	@echo "Running JavaScript engine tests with Chrome at: $(CHROME_PATH)"
	@export CHROME_PATH="$(CHROME_PATH)" && go test $(GO_TEST_FLAGS) ./... -run TestJSEngine -timeout 30s

.PHONY: baseline-jsengine
baseline-jsengine: check-chrome ## Run JavaScript engine baseline tests with comprehensive analysis
	$(call check_chrome)
	@echo "Running JavaScript engine baseline testing..."
	@export CHROME_PATH="$(CHROME_PATH)" && ./scripts/baseline-jsengine-tests.sh

.PHONY: test-all
test-all: check-chrome ## Run all tests including JavaScript tests
	$(call check_chrome)
	@echo "Running all tests with Chrome at: $(CHROME_PATH)"
	@export CHROME_PATH="$(CHROME_PATH)" && go test $(GO_TEST_FLAGS) ./... -timeout 60s

.PHONY: test-short
test-short: ## Explicitly run tests in short mode
	@echo "Running tests in short mode (skipping JavaScript tests)..."
	go test $(GO_TEST_FLAGS) $(GO_SHORT_FLAGS) ./...

.PHONY: lint
lint: ## Run linter (golangci-lint)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, running go vet and go fmt instead..."; \
		go vet ./...; \
		go fmt ./...; \
		if [ -n "$$(gofmt -l .)" ]; then \
			echo "Code is not formatted. Run 'go fmt ./...' to fix."; \
			exit 1; \
		fi; \
	fi

.PHONY: install-lint
install-lint: ## Install golangci-lint
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	else \
		echo "golangci-lint is already installed"; \
	fi

.PHONY: build
build: ## Build the project
	go build -o bin/crawler_go .

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	go clean

.PHONY: deps
deps: ## Download dependencies
	go mod tidy
	go mod download

.PHONY: ci-setup
ci-setup: ## Setup for CI environment (exports Chrome path)
	$(call check_chrome)
	@echo "CHROME_PATH=$(CHROME_PATH)" >> $$GITHUB_ENV

.PHONY: verify-chrome
verify-chrome: check-chrome ## Verify Chrome installation and version
	$(call check_chrome)
	$(call check_chrome_version)
	@echo "Chrome verification completed successfully"
	@"$(CHROME_PATH)" --version

# Development targets
.PHONY: dev-setup
dev-setup: deps install-lint verify-chrome ## Setup development environment
	@echo "Development environment setup complete"

.PHONY: dev-test
dev-test: test-short ## Run development tests (short mode)

.PHONY: dev-test-full
dev-test-full: test-all ## Run full development tests

# CI/CD targets
.PHONY: ci-test
ci-test: ## Run tests in CI environment
	@if [ -n "$$CHROME_PATH" ] && [ -f "$$CHROME_PATH" ]; then \
		echo "Chrome detected in CI, running all tests..."; \
		go test $(GO_TEST_FLAGS) ./... -timeout 60s; \
	else \
		echo "Chrome not available in CI, running short tests only..."; \
		go test $(GO_TEST_FLAGS) $(GO_SHORT_FLAGS) ./... -timeout 30s; \
	fi
