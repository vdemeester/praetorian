BINARY      := praetorian
BIN_DIR     := bin
PKG         := github.com/vdemeester/praetorian

VERSION     ?= $(shell cat VERSION 2>/dev/null || echo dev)
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

# Cut a release: bump the VERSION file, commit it, and tag the commit so the
# tag matches VERSION (the release workflow enforces v$(cat VERSION) == tag).
# Does NOT push — prints the explicit refspec commands to run.
#   make release V=2.0.0-rc3
.PHONY: release
release:
	@test -n "$(V)" || { echo "usage: make release V=<version> (e.g. 2.0.0-rc3)"; exit 1; }
	@test -z "$$(git status --porcelain)" || { echo "error: working tree not clean"; exit 1; }
	@git rev-parse -q --verify "refs/tags/v$(V)" >/dev/null && { echo "error: tag v$(V) already exists"; exit 1; } || true
	@printf '%s\n' "$(V)" > VERSION
	@git add VERSION
	@git commit -m "release: v$(V)"
	@git tag -a "v$(V)" -m "v$(V)"
	@echo
	@echo "Tagged v$(V). To publish:"
	@echo "    git push origin $$(git rev-parse --abbrev-ref HEAD):$$(git rev-parse --abbrev-ref HEAD)"
	@echo "    git push origin v$(V):refs/tags/v$(V)"
