GO ?= go
GORELEASER ?= goreleaser
VERSION ?= dev
BINARY := bin/codex-sess
LDFLAGS := -X main.version=$(VERSION)
GO_BIN_DIR := $(shell $(GO) env GOPATH)/bin

.PHONY: build install test vet fmt fmt-check check bench release-check snapshot

build:
	CGO_ENABLED=0 $(GO) build -buildvcs=false -trimpath -ldflags '$(LDFLAGS)' -o $(BINARY) .

install:
	install -d "$${GOBIN:-$(GO_BIN_DIR)}"
	CGO_ENABLED=0 $(GO) build -buildvcs=false -trimpath -ldflags '$(LDFLAGS)' -o "$${GOBIN:-$(GO_BIN_DIR)}/codex-sess" .

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w *.go

fmt-check:
	@test -z "$$(gofmt -l *.go)" || { gofmt -l *.go; exit 1; }

check: fmt-check test vet

bench:
	$(GO) test -run '^$$' -bench . -benchmem ./...

release-check:
	$(GORELEASER) check

# Build and package every release target locally; never upload artifacts.
snapshot: check release-check
	$(GORELEASER) release --snapshot --clean --parallelism 2
