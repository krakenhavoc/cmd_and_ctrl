.PHONY: docker-test docker-lint all buildbuild-bot build-boteval run dev dev-skip-validation test vet lint fmt fmt-check tidy clean

BINARY := bin/cmd_and_ctrl-server
BOT_BINARY := bin/cmd_and_ctrl-bot
BOTEVAL_BINARY := bin/boteval
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
# version go.mod declares (ADR 0122 §1.1); keep it in step with
# scripts/go-docker.sh's LINT_PKG. First invocation downloads + builds
# the tool; later runs hit the Go build cache. Unlike the old target,
# this can never silently skip linting.
GOLANGCI_LINT := github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0

lint:
	go run $(GOLANGCI_LINT) run $(PKGS)

# No local Go? The same checks through the shared-cache Docker wrapper (#2035).
docker-test:
	../scripts/go-docker.sh test $(PKGS)

docker-lint:
	../scripts/go-docker.sh lint $(PKGS)

tidy:
	go mod tidy

clean:
	rm -rf bin
