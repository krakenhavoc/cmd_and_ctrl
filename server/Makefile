.PHONY: docker-test docker-lint vulncheck tools tools-key all buildbuild-bot build-boteval build-mcpseat run dev dev-skip-validation test vet lint fmt fmt-check tidy clean

BINARY := bin/cmd_and_ctrl-server
BOT_BINARY := bin/cmd_and_ctrl-bot
BOTEVAL_BINARY := bin/boteval
MCPSEAT_BINARY := bin/cmd_and_ctrl-mcpseat
PKGS := ./...

# Default target never mutates the working tree: it verifies formatting,
# runs vet and tests, and builds the binary. Use `make fmt` explicitly
# when you want to rewrite files.
all: fmt-check vet test build build-bot

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/server

# S12.5 bot binary — second systemd unit on the prod VPS. Lives
# in the same Go module as the server so config + types can be
# shared; runs as its own process for crash isolation and
# independent restart.
build-bot:
	mkdir -p bin
	go build -o $(BOT_BINARY) ./cmd/bot

# S31 bot evaluation harness (#837). Not a deployed binary: a local
# tool for the runs that need a model endpoint, a Scryfall dump or
# minutes of wall clock, none of which CI has. `boteval probe` is the
# one that matters today — it says whether the local endpoint is
# truncating the prompt or thinking instead of answering.
build-boteval:
	mkdir -p bin
	go build -o $(BOTEVAL_BINARY) ./cmd/boteval

# ADR 0122 MCP seat: a local stdio binary Claude Code or Codex launches to
# sit in a guest seat at a table. Not deployed. With no local Go, from the
# repo root: scripts/go-docker.sh go build -o bin/cmd_and_ctrl-mcpseat ./cmd/mcpseat
build-mcpseat:
	mkdir -p bin
	go build -o $(MCPSEAT_BINARY) ./cmd/mcpseat

run: build
	./$(BINARY)

dev:
	CMDCTRL_ADMIN_TOKEN="$${CMDCTRL_ADMIN_TOKEN:-dev-admin-token-not-for-production}" \
	CMDCTRL_DATA_DIR="$${CMDCTRL_DATA_DIR:-../data}" \
	go run ./cmd/server

# S14 manual-test convenience: spin the server with deck validation
# bypassed so a throwaway 5-card list (one commander, a few catalog
# cards, a basic or two) imports without tripping the 100-card /
# singleton / color-identity checks. Boot logs WARN when the flag is
# active — never use in production.
dev-skip-validation:
	CMDCTRL_ADMIN_TOKEN="$${CMDCTRL_ADMIN_TOKEN:-dev-admin-token-not-for-production}" \
	CMDCTRL_DATA_DIR="$${CMDCTRL_DATA_DIR:-../data}" \
	CMDCTRL_DEV_SKIP_DECK_VALIDATION=1 \
	go run ./cmd/server

test:
	go test -race -cover $(PKGS)

vet:
	go vet $(PKGS)

# Rewrite source in place. Explicit opt-in.
fmt:
	gofmt -s -w .

# Non-destructive: exits non-zero if any file would be rewritten.
fmt-check:
	@unformatted=$$(gofmt -s -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "These files are not gofmt-clean:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

# Pinned golangci-lint, so no locally-installed binary is needed and
# the version can't drift from the v2-format config in .golangci.yml.
# v2.14.0 supports and runs on Go 1.27, the version go.mod declares
# (ADR 0122 §1.1); keep it in step with scripts/go-docker.sh's LINT_PKG.
# Without a prebuilt binary (`tools`, below) lint runs it through
# `go run`: the first invocation downloads + builds the tool, later runs
# hit the Go build cache. Unlike the old target, this can never silently
# skip linting.
GOLANGCI_LINT_VERSION := v2.14.0
GOLANGCI_LINT := github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

# Pinned govulncheck (#3335): known vulnerabilities that the code can
# actually reach, in the Go standard library of the toolchain running it
# and in the modules go.mod requires. A finding exits non-zero. A
# standard-library finding is fixed by the newest patch release, which
# CI's setup-go "1.27" picks up on its own (locally: docker pull
# golang:1.27). A module finding is fixed with `go get <module>@<fixed>`.
GOVULNCHECK_VERSION := v1.9.0
GOVULNCHECK := golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

# Prebuilt lint and vulncheck binaries (#4958). Neither tool is in
# go.mod, so `go run` downloads and compiles it wherever the module and
# build caches are cold, as on every hosted CI run. `make tools` installs
# both under TOOLS_DIR/<go version>/<tool>-<version>/, built by exactly
# that Go version, so a pin bump or a new toolchain is a fresh directory
# and a stale binary is never run. lint and vulncheck use the binary when
# it is there and fall back to `go run` when it is not. CI caches
# TOOLS_DIR keyed on `make tools-key` (ci-cd.yml, server-check), so the
# two *_VERSION pins above stay the one place the versions are set.
#
# Both paths build with the Go version go.mod selects ($gv below), pinned
# through GOTOOLCHAIN: on a host whose go is older than go.mod's,
# `go run pkg@version` and `go install pkg@version` otherwise switch to
# the go line of the TOOL's go.mod (go1.26 for both today), and a tool
# built by 1.26 refuses to load this go 1.27 module.
TOOLS_DIR ?= $(CURDIR)/bin/tools
# $(call tool_bin,<tool>,<version>): the prebuilt path, as one shell word.
tool_bin = "$(TOOLS_DIR)/$$gv/$(1)-$(2)/$(1)"

tools:
	@set -e; gv=$$(go env GOVERSION); \
	build_tool() { \
	  if [ -x "$$1" ]; then echo "tools: $$1 is already built"; return 0; fi; \
	  echo "tools: building $$2 with $$gv"; \
	  GOTOOLCHAIN="$$gv" GOBIN="$$(dirname "$$1")" go install "$$2"; \
	}; \
	build_tool $(call tool_bin,golangci-lint,$(GOLANGCI_LINT_VERSION)) $(GOLANGCI_LINT); \
	build_tool $(call tool_bin,govulncheck,$(GOVULNCHECK_VERSION)) $(GOVULNCHECK)

# What CI's tool cache key holds besides the OS, the arch and the Go version.
tools-key:
	@echo "golangci-lint-$(GOLANGCI_LINT_VERSION)_govulncheck-$(GOVULNCHECK_VERSION)"

lint:
	@gv=$$(go env GOVERSION); bin=$(call tool_bin,golangci-lint,$(GOLANGCI_LINT_VERSION)); \
	if [ -x "$$bin" ]; then set -x; "$$bin" run $(PKGS); \
	else set -x; GOTOOLCHAIN="$$gv" go run $(GOLANGCI_LINT) run $(PKGS); fi

vulncheck:
	@gv=$$(go env GOVERSION); bin=$(call tool_bin,govulncheck,$(GOVULNCHECK_VERSION)); \
	if [ -x "$$bin" ]; then set -x; "$$bin" $(PKGS); \
	else set -x; GOTOOLCHAIN="$$gv" go run $(GOVULNCHECK) $(PKGS); fi

# No local Go? The same checks through the shared-cache Docker wrapper (#2035).
docker-test:
	../scripts/go-docker.sh test $(PKGS)

docker-lint:
	../scripts/go-docker.sh lint $(PKGS)

tidy:
	go mod tidy

clean:
	rm -rf bin
