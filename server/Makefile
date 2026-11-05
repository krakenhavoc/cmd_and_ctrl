.PHONY: docker-test docker-lint vulncheck tool-pins all buildbuild-bot build-boteval build-mcpseat run dev dev-skip-validation test vet lint fmt fmt-check tidy clean

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

# Pinned golangci-lint, run via `go run` so no locally-installed
# binary is needed and the version can't drift from the v2-format
# config in .golangci.yml. v2.14.0 supports and runs on Go 1.27, the
# version go.mod declares (ADR 0122 §1.1). This line is the one pin:
# scripts/go-docker.sh reads GOLANGCI_LINT_VERSION from it, and CI's
# tool cache is keyed on `make tool-pins` (#4958). First invocation
# downloads + builds the tool; later runs hit the Go build cache.
# Unlike the old target, this can never silently skip linting.
GOLANGCI_LINT_VERSION := v2.14.0
GOLANGCI_LINT := github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

# Pinned govulncheck (#3335): known vulnerabilities that the code can
# actually reach, in the Go standard library of the toolchain running it
# and in the modules go.mod requires. A finding exits non-zero. A
# standard-library finding is fixed by the newest patch release, which
# CI's setup-go "1.27" picks up on its own (locally: docker pull
# golang:1.27). A module finding is fixed with `go get <module>@<fixed>`.
GOVULNCHECK := golang.org/x/vuln/cmd/govulncheck@v1.9.0

# Prebuilt tools (#4958). Neither tool is in go.mod, so on a cold
# hosted runner `go run` downloaded and compiled both before lint could
# start. With TOOLS_DIR set, each is `go install`ed once into a
# directory named for the Go release and the pin, and run from there:
# CI caches TOOLS_DIR, and a pin or Go bump lands in a new directory
# rather than reusing a stale binary. The install runs on that Go
# release (GOTOOLCHAIN): `go install pkg@version`, like `go run`, would
# otherwise pick the toolchain from the tool's go.mod and the local Go,
# not from ours. Unset (the default), both targets are plain `go run`
# as before.
ifdef TOOLS_DIR
TOOLS_GO := $(shell go env GOVERSION)
LINT_BIN := $(abspath $(TOOLS_DIR))/$(TOOLS_GO)/$(notdir $(GOLANGCI_LINT))/golangci-lint
VULN_BIN := $(abspath $(TOOLS_DIR))/$(TOOLS_GO)/$(notdir $(GOVULNCHECK))/govulncheck

$(LINT_BIN):
	GOBIN=$(@D) GOTOOLCHAIN=$(TOOLS_GO)+auto go install $(GOLANGCI_LINT)

$(VULN_BIN):
	GOBIN=$(@D) GOTOOLCHAIN=$(TOOLS_GO)+auto go install $(GOVULNCHECK)

lint: $(LINT_BIN)
	$(LINT_BIN) run $(PKGS)

vulncheck: $(VULN_BIN)
	$(VULN_BIN) $(PKGS)
else
lint:
	go run $(GOLANGCI_LINT) run $(PKGS)

vulncheck:
	go run $(GOVULNCHECK) $(PKGS)
endif

# The pins, one per line, for CI's tool cache key (#4958).
tool-pins:
	@echo $(GOLANGCI_LINT)
	@echo $(GOVULNCHECK)

# No local Go? The same checks through the shared-cache Docker wrapper (#2035).
docker-test:
	../scripts/go-docker.sh test $(PKGS)

docker-lint:
	../scripts/go-docker.sh lint $(PKGS)

tidy:
	go mod tidy

clean:
	rm -rf bin
