.PHONY: all build run dev dev-skip-validation test vet lint fmt fmt-check tidy clean

BINARY := bin/cmd_and_ctrl-server
PKGS := ./...

# Default target never mutates the working tree: it verifies formatting,
# runs vet and tests, and builds the binary. Use `make fmt` explicitly
# when you want to rewrite files.
all: fmt-check vet test build

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/server

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

lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not installed; skipping. Install from https://golangci-lint.run/"; \
		exit 0; \
	fi
	golangci-lint run $(PKGS)

tidy:
	go mod tidy

clean:
	rm -rf bin
