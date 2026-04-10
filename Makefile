.PHONY: help dev server-dev client-dev server-test client-check test lint

help:
	@echo "cmd_and_ctrl — top-level targets"
	@echo ""
	@echo "  make dev           Run server and client dev loops (two terminals recommended)"
	@echo "  make server-dev    Run the Go game server (:8080)"
	@echo "  make client-dev    Run the Vite dev server (:5173, proxies /ws)"
	@echo "  make test          Run server tests"
	@echo "  make lint          Lint server and client"
	@echo ""
	@echo "Per-service targets live in server/Makefile and client/package.json."

dev:
	@echo "Run 'make server-dev' in one terminal and 'make client-dev' in another."
	@echo "Then open http://localhost:5173."

server-dev:
	$(MAKE) -C server dev

client-dev:
	cd client && npm run dev

server-test:
	$(MAKE) -C server test

client-check:
	cd client && npm run check

test: server-test client-check

lint:
	$(MAKE) -C server vet
	cd client && npm run lint
