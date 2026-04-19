# tests-e2e — Playwright end-to-end tests

End-to-end tests for the cmd_and_ctrl sandbox client. Drives the real
Go server (`server/cmd/server`) and the real Vite dev server
(`client/`) through a headless Chromium.

## What's here

```
tests-e2e/
├── playwright.config.ts   # spawns server + client, shared fixtures
├── package.json
├── tsconfig.json
└── tests/
    ├── env.ts             # shared constants (admin token, etc.)
    ├── lobby-api.ts       # thin HTTP wrapper for test setup shortcuts
    ├── deck-fixture.ts    # minimal legal Commander deck (Kenrith + 99 Plains)
    ├── smoke.spec.ts      # router + healthz
    ├── auth.spec.ts       # admin login, logout, invite-URL paste
    ├── lobby.spec.ts      # admin lobby create / list / refresh
    ├── join.spec.ts       # invite-link → join flow
    ├── game.spec.ts       # game route WS handshake
    └── full-game.spec.ts  # 2-player game: join → upload → start → mulligan → turns
```

## Setup

First-time setup installs Playwright, downloads the Chromium binary
(~150 MB), and pulls the system shared libraries Chromium needs:

```bash
cd tests-e2e
npm install
npx playwright install --with-deps chromium   # requires sudo
```

If you don't have sudo, swap `--with-deps chromium` for plain
`chromium` and install the system packages yourself — on Debian /
Ubuntu that's:

```bash
sudo apt-get install -y libnspr4 libnss3 libasound2t64 \
  libatk1.0-0t64 libatk-bridge2.0-0t64 libcups2t64 \
  libxkbcommon0 libxcomposite1 libxdamage1 libxfixes3 \
  libxrandr2 libgbm1 libpango-1.0-0 libcairo2
```

Go is expected to be on `PATH` — the config runs `go run ./cmd/server`
rather than a prebuilt binary.

The `full-game.spec.ts` test imports a deck against the Scryfall
bulk dump. Make sure `<repo>/data/scryfall/default-cards.json`
exists before running that test — see `scripts/scryfall-refresh.sh`
if it's not there yet. Every other test runs without card data.

## Running

```bash
cd tests-e2e
npm test                    # headless
npm run test:headed         # watch the browser
npm run test:ui             # Playwright's UI mode (best for debugging)
npm run report              # open the last HTML report
```

The config starts both the Go server (:8080) and the Vite client
(:5173) as `webServer` entries. With `reuseExistingServer: true`
(default outside CI) it'll attach to anything already running on
those ports, so you can iterate against a long-lived dev stack.

## Design notes

- **Admin token matches `make server-dev`.** Both `playwright.config.ts`
  and `tests/env.ts` use `dev-admin-token-not-for-production` — the
  same default `server/Makefile` picks when `CMDCTRL_ADMIN_TOKEN` is
  unset. Run the dev server normally and the tests attach to it.
- **Single worker.** Lobby state is a process-global map on the
  server; parallel workers would race on game IDs. Once the server
  gains per-test namespacing we can raise this.
- **API helpers for setup, UI assertions for the test.** The
  `lobby-api.ts` wrapper mints admin sessions and creates games over
  HTTP; the actual assertions still run through the browser. This
  keeps tests fast without sacrificing UI coverage.
- **No Scryfall data required.** Tests avoid paths that depend on the
  `default-cards.json` dump (image routes, deck upload happy-path).
  Full deck-upload coverage lives in the server's Go tests, where
  validation is exercised directly.

## Adding tests

- Keep tests independent — every test should set its own session up
  via the login UI or by seeding `localStorage`, and should never
  rely on state from a previous test.
- Use `{ page, request }` together when a test needs both the browser
  and a side-channel (e.g., read an invite token that the UI only
  flashes briefly).
- Prefer role-based selectors (`getByRole("button", { name: "create" })`)
  over CSS selectors; the styles move, the labels rarely do.
