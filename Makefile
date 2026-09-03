BINARY  := aido
PKG     := github.com/mikepea/aido-linux-unlocker
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.1.0-dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X $(PKG)/internal/version.Version=$(VERSION) \
	-X $(PKG)/internal/version.Commit=$(COMMIT) \
	-X $(PKG)/internal/version.BuildDate=$(DATE)

# The laptop is Arch on x86_64; everything ships as a static-ish CGO-free build
# so there is nothing to install alongside it.
TARGET_OS   ?= linux
TARGET_ARCH ?= amd64

.PHONY: all
all: check build

.PHONY: build
build: ## build for the host, into bin/
	go build -ldflags '$(LDFLAGS)' -o bin/$(BINARY) ./cmd/$(BINARY)

.PHONY: run
run: build ## build and play
	./bin/$(BINARY)

.PHONY: dist
dist: ## cross-compile the laptop build into dist/
	CGO_ENABLED=0 GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) \
		go build -trimpath -ldflags '$(LDFLAGS)' \
		-o dist/$(BINARY)-$(TARGET_OS)-$(TARGET_ARCH) ./cmd/$(BINARY)

.PHONY: release
release: dist ## build dist/ and write the update manifest
	VERSION='$(VERSION)' BINARY='$(BINARY)' \
	TARGET_OS='$(TARGET_OS)' TARGET_ARCH='$(TARGET_ARCH)' \
		./scripts/release.sh

.PHONY: check
check: fmt vet test ## everything CI would run

.PHONY: test
test:
	go test ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: fmt
fmt:
	@out="$$(gofmt -l . )"; \
	if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: clean
clean:
	rm -rf bin dist

.PHONY: help
help:
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
