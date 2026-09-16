# client

The `cmd_and_ctrl` web client. TypeScript + Vite + Svelte 5 + (eventually)
PixiJS. See the [top-level PLAN.md](../PLAN.md) and [AGENTS.md](../AGENTS.md).

## Running

```bash
# install deps (first time)
npm install

# dev server on :5173, proxies /ws to ws://localhost:8080
npm run dev

# build and type-check
npm run build

# lint + format check
npm run lint

# auto-format
npm run format

# unit tests
npm test
```

The Vite dev server proxies `/ws` to the Go game server, so run the server
in another terminal (`cd ../server && make dev`) and open
[http://localhost:5173](http://localhost:5173) to drive the ping round-trip.

## Layout

```
client/
├── src/
│   ├── main.ts             # app entry point, mount(App, ...)
│   ├── App.svelte          # S01 demo UI: send ping, show log
│   ├── app.css             # global styles
│   └── lib/
│       ├── protocol.ts     # hand-maintained mirror of server protocol types
│       └── ws.ts           # WebSocket client + Svelte stores
├── index.html              # Vite entry HTML
├── vite.config.ts          # Vite + /ws proxy to Go server
├── svelte.config.js        # Svelte preprocessor
├── tsconfig.json
├── eslint.config.js        # ESLint 9 flat config
├── .prettierrc
├── .prettierignore
└── package.json
```

## Architecture notes

- State from the server lives in Svelte stores exposed by `GameClient`
  (`lib/ws.ts`). Components subscribe to stores; they never call the
  WebSocket directly.
- The protocol types in `lib/protocol.ts` are a hand-maintained mirror of
  `server/internal/protocol/protocol.go`. When the spec in
  `docs/protocol.md` changes, update both sides in lockstep.
- Svelte 5 runes (`$state`, etc.) are used throughout.

## Card art

Build art URLs with `lib/cardImage.ts` (`lib/catalog.ts` for the public
catalogue), and put `use:cardArt={url}` (`lib/cardArt.ts`) on every
`<img>` that shows card art — the same string as its `src`. The action
retries a failed load once after ~2 s, then inserts a small
"Art failed to load" pip that retries on click, and records the failure
in `lib/clientErrors.ts` once per URL. The logic is in
`lib/cardArtRetry.ts` and is unit-tested there; the action is DOM
wiring only.

- The img's parent must be positioned (`position: relative`); set
  `--art-error-top` / `--art-error-right` / `--art-error-left` (with
  `-right: auto`) on it to move the pip off a corner the tile already
  uses, and `--art-error-z` to lift it over the tile's own overlays.
  `Card.svelte` puts it on the left edge below the top badge row,
  above the counter column.
- Inside a tile that has its own click / double-click / Enter handling
  nothing extra is needed — the pip stops propagation. Inside a
  `pointer-events: none` or `aria-hidden` surface, pass
  `{ url, interactive: false }`.
- Keyboard and screen readers: where nothing above the img is a
  control, the pip is a real button (a tab stop, Enter/Space retry).
  Inside a control or image — a `<button>`, or `role="button"` /
  `"img"` and the other roles whose children are presentational — a
  nested button would be a tab stop with no role or name, so the pip
  is pointer-only (`aria-hidden`), the outermost such control (the
  pile `<button>` around a `role="img"` `Card`, not the `Card`) gets
  "Art failed to load" through `aria-describedby`, and keyboard focus
  on it retries the art.
- Not for card backs (`/card-back*.jpg`), which are bundled assets.
  Nor for the seat avatar in `PlayerIdentity.svelte`: a commander art
  crop that fails there falls through to the seat-colour disc, which
  is already its error state.
- `fetchpriority="high"` is reserved for the viewer's own hand
  (`Card`'s `priority` prop, set by `Hand.svelte`) and the hover-zoom
  scan. Don't add it elsewhere: on every tile it means nothing.

## PWA

The client installs as a progressive web app. See
[ADR 0031](../docs/decisions/0031-progressive-web-app.md) for the full
rationale; the short version:

- `public/manifest.webmanifest` + `public/icons/` make it installable.
  Regenerate the PNGs from the SVG sources with `tools/gen-icons.sh`.
- `src/sw/service-worker.js` is a **template**. `vite-plugin-sw.ts` stamps a
  content-derived build id and the precache list into it at build time and
  emits `/sw.js`. It is deny-by-default: `/ws` and every `@api` route in
  `deploy/Caddyfile` are network-only, and only the app shell, hashed
  `/assets/*` and immutable card art (`/cards/{id}/image`) are cached.
- The worker is **not registered in dev**. To exercise it, run
  `npm run build && npm run preview`.
- A new build never reloads the page by itself: it waits, and
  `lib/components/UpdatePrompt.svelte` asks the player. Do not add
  `skipWaiting()` anywhere else.
