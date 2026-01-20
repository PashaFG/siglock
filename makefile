# siglock/Makefile

SHELL := /bin/bash
NAME := siglock
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILDDATE ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

OUT_DIR := dist
BIN_NAME := $(NAME)
BIN_PATH := $(OUT_DIR)/$(BIN_NAME)

GO := go
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.builddate=$(BUILDDATE)
BUILD_FLAGS := -ldflags="$(LDFLAGS)" -trimpath

HOST_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
HOST_ARCH := $(shell uname -m)
ifeq ($(HOST_ARCH),x86_64)
	GOARCH := amd64
else ifeq ($(HOST_ARCH),arm64)
	GOARCH := arm64
else
	GOARCH := $(HOST_ARCH)
endif

.PHONY: all build build-fat release clean help

all: build

build: ## build binary for current arch
	@echo "🔧 Building $(BIN_NAME) for $(HOST_OS)/$(GOARCH)..."
	@mkdir -p $(OUT_DIR)
	$(GO) build $(BUILD_FLAGS) -o $(BIN_PATH)

build-fat: check-darwin ## build universal2 binary
	@echo "🔧 Building universal binary for macOS..."
	@mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -o $(OUT_DIR)/$(BIN_NAME)-darwin-arm64
	GOOS=darwin GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o $(OUT_DIR)/$(BIN_NAME)-darwin-amd64
	lipo -create -output $(BIN_PATH) $(OUT_DIR)/$(BIN_NAME)-darwin-arm64 $(OUT_DIR)/$(BIN_NAME)-darwin-amd64
	@rm -f $(OUT_DIR)/$(BIN_NAME)-darwin-*

install: build ## install build in /usr/local/bin and add in PATH
	@echo "📥 Installing to /usr/local/bin/$(BIN_NAME)..."
	@sudo install -m 0755 $(BIN_PATH) /usr/local/bin/$(BIN_NAME)

clean: ## remove building binaries
	@echo "🧹 Cleaning..."
	@rm -rf $(OUT_DIR)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

check-darwin:
	@if [ "$(HOST_OS)" != "darwin" ]; then \
		echo "❌ build-fat only supported on macOS"; \
		exit 1; \
	fi
