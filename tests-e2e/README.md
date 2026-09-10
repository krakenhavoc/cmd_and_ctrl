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
    ├── auth.spec.ts           # admin login, logout, invite-URL paste
    ├── board-layout.spec.ts   # HTML/CSS board: zones, pile buttons, draw-to-hand
    ├── deck-fixture.ts        # minimal legal Commander deck (Kenrith + 99 Plains)
    ├── entry.spec.ts          # invite page: table preview, full table, spectator link
    ├── env.ts                 # shared constants (admin token, etc.)
    ├── full-game.spec.ts      # 2-player game: join → upload → start → mulligan → turns
    ├── game.spec.ts           # game route WS handshake
    ├── join.spec.ts           # invite-link → join flow
    ├── lobby-api.ts           # thin HTTP wrapper for test setup shortcuts
    ├── lobby.spec.ts          # admin lobby create / list / refresh / table cards
    ├── mulligan.spec.ts       # keep / mulligan dialog + the opening-hand roll call
    ├── players.ts             # shared joinAsPlayer (invite flow → seated on the game route)
    ├── s19-deck-fixture.ts    # two 100-card decks built around the S19 trigger cards
    ├── s19-helpers.ts         # S19 game setup via admin WS (join, decks, keep_hand)
    ├── s19-triggers.spec.ts   # S19 triggered abilities: chooser modal, targets, resolution
    ├── smoke.spec.ts          # router + entry pages + healthz
    └── zone-browser.spec.ts   # graveyard / exile modal behind the pile chips
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

The `full-game.spec.ts`, `board-layout.spec.ts`, and S19 trigger
suites import decks against the Scryfall bulk dump. Make sure
`<repo>/data/scryfall/default-cards.json` exists before running
them — see `scripts/scryfall-refresh.sh` if it's not there yet. The
remaining tests run without card data.

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
- **Scryfall data needed only by the deck-driven suites.** Most tests
  avoid paths that depend on the `default-cards.json` dump, and deck
  validation itself is covered by the server's Go tests. But
  `full-game.spec.ts` and `board-layout.spec.ts` each import a deck,
  and the S19 trigger suite uploads two 100-card decks against the
  bulk index — those need `data/scryfall/default-cards.json` present.
- **Rate limits are relaxed for the spawned server.** The config sets
  `CMDCTRL_DEV_RELAX_RATE_LIMITS=1` so serial suites that mint many
  sessions (S19 spins up a fresh 2-player game per test) don't trip
  the lobby join/login limiter between tests. If you attach the suite
  to an already-running dev server instead, start that server with the
  same variable set.

## Adding tests

- Keep tests independent — every test should set its own session up
  via the login UI or by seeding `localStorage`, and should never
  rely on state from a previous test.
- Use `{ page, request }` together when a test needs both the browser
  and a side-channel (e.g., read an invite token that the UI only
  flashes briefly).
- Prefer role-based selectors (`getByRole("button", { name: "create" })`)
  over CSS selectors; the styles move, the labels rarely do. Better
  still, prefer a selector the app already had to get right for
  accessibility: a modal's `aria-labelledby` heading, a panel's
  `role="region"` + `aria-label`, a pile chip's `aria-label` count.
  Those are contracts; class names and copy are not.
- Watch for accidental substring matches. `getByRole`'s `name` is a
  case-insensitive SUBSTRING match by default, so `{ name: "Sol Ring" }`
  inside the zone browser also matches "move Sol Ring to hand" and its
  two siblings — pass `exact: true` when you mean the card itself.
  More than one thing on the table carries `role="dialog"` (the prompt
  modal, the zone browser, the targeting banner), so always name the
  dialog you mean.
- **The admin WebSocket is not a player's browser.** The S19 helpers
  observe the authoritative snapshot, which is ahead of every page. A
  click on something that is on screen either way — a battlefield card,
  a pile chip — has no actionability wait to save it: if the page has
  not entered targeting mode yet the click is silently swallowed and
  the test hangs until it times out. Gate board clicks on a UI signal
  from that player's own page (`waitForPickTarget` waits for the
  targeting banner). Clicks on modal buttons are fine — the button
  does not exist until the modal renders.
- **Seeding a card by drawing changes the library.** `seedHandWithCard`
  draws until the named card surfaces, because library contents are
  redacted on the wire (CR 400.2). On an unlucky shuffle that empties
  the deck. Any assertion downstream of "search your library" or
  "draw N" must restock first (`returnToLibrary`) or it is really
  testing deck order.
