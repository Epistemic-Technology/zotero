.PHONY: build clean test test-unit test-integration test-all help zotero-cli

# Go parameters
GOEXPERIMENT := jsonv2
BINARY_DIR := bin

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

build: ## Build all binaries
	GOEXPERIMENT=$(GOEXPERIMENT) go build -o $(BINARY_DIR)/ ./cmd/...

zotero-cli: ## Build zotero-cli binary
	GOEXPERIMENT=$(GOEXPERIMENT) go build -o $(BINARY_DIR)/zotero-cli ./cmd/zotero-cli

clean: ## Remove build artifacts
	rm -rf $(BINARY_DIR)

test: test-unit ## Run unit tests (default, fast)

test-unit: ## Run unit tests only (mock tests)
	go test ./zotero -v

test-integration: ## Run integration tests (requires credentials)
	@if [ -f .env ]; then \
		set -a; . ./.env; set +a; go test ./tests -v; \
	else \
		go test ./tests -v; \
	fi

test-all: ## Run all tests (unit + integration)
	@$(MAKE) test-unit
	@$(MAKE) test-integration
