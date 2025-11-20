# Makefile for podcast-transcribe tool

# Build variables
BINARY_NAME=podcast-transcribe
BUILD_DIR=build
CMD_DIR=cmd/podcast-transcribe
INSTALL_PATH=/usr/local/bin

# Whisper.cpp variables
WHISPER_CPP_VERSION=v1.8.2
WHISPER_CPP_DIR=whisper.cpp
WHISPER_LIB=$(WHISPER_CPP_DIR)/libwhisper.a

# Go build flags
GOTOOLCHAIN=local
GO=GOTOOLCHAIN=$(GOTOOLCHAIN) go
GOFLAGS=-v
LDFLAGS=-w -s

# Platform detection
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Darwin)
	PLATFORM=darwin
	ifeq ($(UNAME_M),arm64)
		ARCH=arm64
		# Enable Metal acceleration for M1/M2/M3 Macs
		WHISPER_METAL=1
	else
		ARCH=amd64
	endif
else ifeq ($(UNAME_S),Linux)
	PLATFORM=linux
	ARCH=amd64
endif

.PHONY: all build clean install uninstall deps whisper test help

all: build ## Build the project

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

deps: ## Download Go dependencies
	$(GO) mod download
	$(GO) mod verify

whisper: ## Clone and build whisper.cpp
	@if [ ! -d "$(WHISPER_CPP_DIR)" ]; then \
		echo "Cloning whisper.cpp..."; \
		git clone --depth 1 --branch $(WHISPER_CPP_VERSION) https://github.com/ggerganov/whisper.cpp.git $(WHISPER_CPP_DIR); \
	else \
		echo "whisper.cpp already exists, cleaning before rebuild..."; \
		cd $(WHISPER_CPP_DIR) && $(MAKE) clean || true; \
	fi
	@echo "Building whisper.cpp..."
ifeq ($(WHISPER_METAL),1)
	@echo "Building with Metal acceleration for macOS ARM64..."
	cd $(WHISPER_CPP_DIR) && WHISPER_METAL=1 $(MAKE) libwhisper.a
else
	cd $(WHISPER_CPP_DIR) && $(MAKE) libwhisper.a
endif
	@echo "whisper.cpp built successfully"

build: deps whisper ## Build the CLI tool
	@echo "Building $(BINARY_NAME) for $(PLATFORM)/$(ARCH)..."
	@mkdir -p $(BUILD_DIR)
	C_INCLUDE_PATH=$(PWD)/$(WHISPER_CPP_DIR) \
	LIBRARY_PATH=$(PWD)/$(WHISPER_CPP_DIR) \
	CGO_ENABLED=1 \
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-darwin-amd64: ## Build for macOS (Intel)
	@echo "Cross-compiling for darwin/amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
	C_INCLUDE_PATH=$(PWD)/$(WHISPER_CPP_DIR) \
	LIBRARY_PATH=$(PWD)/$(WHISPER_CPP_DIR) \
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)

build-darwin-arm64: ## Build for macOS (Apple Silicon)
	@echo "Cross-compiling for darwin/arm64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
	C_INCLUDE_PATH=$(PWD)/$(WHISPER_CPP_DIR) \
	LIBRARY_PATH=$(PWD)/$(WHISPER_CPP_DIR) \
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)

build-linux-amd64: ## Build for Linux (AMD64)
	@echo "Cross-compiling for linux/amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
	C_INCLUDE_PATH=$(PWD)/$(WHISPER_CPP_DIR) \
	LIBRARY_PATH=$(PWD)/$(WHISPER_CPP_DIR) \
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)

install: build ## Install the binary to /usr/local/bin
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/
	@echo "Installation complete"

uninstall: ## Uninstall the binary from /usr/local/bin
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_PATH)..."
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstallation complete"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@if [ -d "$(WHISPER_CPP_DIR)" ]; then \
		cd $(WHISPER_CPP_DIR) && $(MAKE) clean; \
	fi
	@echo "Clean complete"

clean-all: clean ## Clean everything including whisper.cpp
	@echo "Removing whisper.cpp directory..."
	@rm -rf $(WHISPER_CPP_DIR)
	@echo "Deep clean complete"

test: ## Run tests
	$(GO) test -v ./...

fmt: ## Format Go code
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

lint: fmt vet ## Run linters

download-model: ## Download Whisper large-v3 model
	@echo "Downloading Whisper large-v3 model..."
	@mkdir -p ~/.cache/whisper
	@if [ ! -f ~/.cache/whisper/ggml-large-v3.bin ]; then \
		curl -L -o ~/.cache/whisper/ggml-large-v3.bin \
		https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin; \
		echo "Model downloaded to ~/.cache/whisper/ggml-large-v3.bin"; \
	else \
		echo "Model already exists at ~/.cache/whisper/ggml-large-v3.bin"; \
	fi

.DEFAULT_GOAL := help
