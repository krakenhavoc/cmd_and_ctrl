.PHONY: all build run dev test vet lint fmt tidy clean

BINARY := bin/cmd_and_ctrl-server
PKGS := ./...

all: fmt vet test build

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

fmt:
	gofmt -s -w .

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
