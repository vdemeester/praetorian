BINARY      := praetorian
BIN_DIR     := bin
PKG         := github.com/vdemeester/praetorian

VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE        ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X $(PKG)/version.Version=$(VERSION) \
	-X $(PKG)/version.Commit=$(COMMIT) \
	-X $(PKG)/version.Date=$(DATE)

GO       ?= go
GOFLAGS  ?=

.PHONY: all
all: check build

.PHONY: build
build:
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' -o $(BIN_DIR)/$(BINARY) .

.PHONY: install
install:
	CGO_ENABLED=0 $(GO) install $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' .

.PHONY: test
test:
	$(GO) test -race -cover ./...

.PHONY: coverage
coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: check
check: fmt vet lint test

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) coverage.out dist

.PHONY: snapshot
snapshot:
	goreleaser release --snapshot --clean

.PHONY: release-check
release-check:
	goreleaser check
