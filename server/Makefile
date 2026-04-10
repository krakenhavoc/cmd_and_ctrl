.PHONY: all build run dev test vet lint fmt fmt-check tidy clean

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
