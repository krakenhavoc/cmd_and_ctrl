.PHONY: help dev server-dev server-dev-skip-validation client-dev server-test client-check client-test test lint cost-usage cost-push cost-report

help:
	@echo "cmd_and_ctrl — top-level targets"
	@echo ""
	@echo "  make dev           Run server and client dev loops (two terminals recommended)"
	@echo "  make server-dev    Run the Go game server (:8080)"
	@echo "  make server-dev-skip-validation  Like server-dev, but bypasses deck validation"
	@echo "                     (lets you import a 5-card test deck; never use in prod)"
	@echo "  make client-dev    Run the Vite dev server (:5173, proxies /ws)"
	@echo "  make test          Run server tests, client typecheck, and client unit tests"
	@echo "  make lint          Lint server and client"
	@echo "  make cost-usage    Measure your Claude Code usage on this repo (local ledger)"
	@echo "  make cost-push     ...and publish the ledger for the weekly cost report"
	@echo "  make cost-report   Print what the project has cost (human / API / subscriptions)"
	@echo ""
	@echo "Per-service targets live in server/Makefile and client/package.json."

dev:
	@echo "Run 'make server-dev' in one terminal and 'make client-dev' in another."
	@echo "Then open http://localhost:5173."

server-dev:
	$(MAKE) -C server dev

server-dev-skip-validation:
	$(MAKE) -C server dev-skip-validation

client-dev:
	cd client && npm run dev

server-test:
	$(MAKE) -C server test

client-check:
	cd client && npm run check

client-test:
	cd client && npm test

test: server-test client-check client-test

lint:
	$(MAKE) -C server vet
	cd client && npm run lint

cost-usage:
	python3 scripts/cost/claude_usage.py

cost-push:
	python3 scripts/cost/claude_usage.py --push

cost-report:
	python3 scripts/cost/cost_report.py
