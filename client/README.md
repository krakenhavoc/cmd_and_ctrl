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
├── vitest.config.ts        # test-only Vite config (see Tests)
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
- **Build every store with `guardedWritable` / `guardedDerived` from
  `lib/guardedStore.ts`, never `writable` / `derived` from
  `svelte/store`.** ESLint enforces it. `svelte/store` keeps ONE
  module-global `subscriber_queue` and never resets it after a callback
  throws, so a single throwing subscriber — or a throwing `derived`
  callback, or a `localStorage` write that hits a quota inside one —
  stops notifications on _every_ store in the app for the life of the
  page. That is #266's "state freeze": socket green, frames arriving,
  nothing reaching the DOM. `guardedStore.ts` has the full mechanism;
  `subscriberQueue.test.ts` pins that it is still unfixed upstream.
- A throw while _rendering_ is the other half, and Svelte 5 batches
  those onto a microtask rather than through the queue.
  `<svelte:boundary>` around `<Board>` in `routes/Game.svelte` catches
  them, records them via `clientErrors.ts` and offers the player a
  redraw. Do not widen it to the whole route: the header, the
  bug-report button and "back to lobby" have to survive a dead table.
- The protocol types in `lib/protocol.ts` are a hand-maintained mirror of
  `server/internal/protocol/protocol.go`. When the spec in
  `docs/protocol.md` changes, update both sides in lockstep.
- Svelte 5 runes (`$state`, etc.) are used throughout.

## Tests

`npm test` runs vitest against `src/**/*.test.ts`. The config is
`vitest.config.ts`, deliberately separate from `vite.config.ts` so the
dev proxy and the build-only service-worker plugin stay out of a test
run; read the comments there before changing it.

**Default to a pure-helper test.** Pull the decision out of the
component into `lib/<thing>.ts` and test that. It is faster, it reads
as a specification of the rule rather than of the markup, and it
survives a redesign. `lib/choiceRejection.ts`, `lib/cardArtRetry.ts`,
`lib/combatBeats.ts` and `lib/priorityStops.ts` are all this pattern,
and it should stay the common case.

**Write a render test only for behaviour that does not exist until the
markup does** (#689):

- roles, ARIA and accessible names — `role="alert"` on a refusal,
  `aria-describedby` on a tile whose art failed
- focus: what is a tab stop, where focus lands, what `:focus-visible`
  triggers
- event propagation — a control inside a control, which is most of the
  board
- conditional rendering: that a branch is reachable at all with real
  props

If a test would only re-assert what a helper already pins, it belongs
in the helper's file instead.

### Writing one

Add `// @vitest-environment jsdom` as the **first line** of the file.
The suite defaults to the `node` environment and that is on purpose:
most tests are pure logic, and several of them (`settings.ts`,
`session.ts`) exist to pin the branch taken when there is no
`localStorage`, which a global DOM would quietly stop testing.

Mount through `lib/test/render.svelte.ts` — Svelte's own `mount`, the
same entry point `main.ts` uses, no testing-library layer:

```ts
// @vitest-environment jsdom
import { afterEach, expect, it } from "vitest";
import Thing from "./components/Thing.svelte";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(cleanup);

it("announces the refusal", () => {
  const view = render(Thing, { value: 1, onDone: () => {} });
  click(view.container.querySelector("button")!);
  view.setProps({ error: "nope" }); // reactive: props are a $state proxy
  expect(view.container.querySelector('[role="alert"]')?.textContent).toContain("nope");
});
```

- `render` flushes the first render for you; `setProps` flushes after
  merging. Mutate state by any other route and call `flushSync()`
  yourself.
- `cleanup()` in an `afterEach` is not optional. A leaked mount's
  `$effect`s keep running against the next test's document.
- `click` dispatches a real bubbling `MouseEvent`, because propagation
  is usually the thing under test.
- Faking timers: fake `setTimeout` only
  (`vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] })`).
  Svelte flushes updates on a microtask, so faking `queueMicrotask`
  strands every render behind a tick nobody advances.

`choicePromptModal.render.test.ts` and `cardArtPip.render.test.ts` are
the two worked examples.

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
  above the counter column, and lower on a tapped card, whose turned
  tile the next tapped land overlaps. `BattlefieldRow.svelte` moves an
  Aura's or Equipment's pip to its bottom-left corner, the part its
  host leaves uncovered whether either is tapped
  (`boardArtPip.test.ts` checks both).
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
  on it retries the art. That needs the control to take focus: a
  `role="img"` `Card` with nothing focusable around it (a zone browser
  tile with no target prompt, say) is not a tab stop, so its retry is
  pointer-only.
- Not for card backs (`/card-back*.jpg`), which are bundled assets.
  Nor for the seat avatar in `PlayerIdentity.svelte`: a commander art
  crop that fails there falls through to the seat-colour disc, which
  is already its error state.
- `fetchpriority="high"` is reserved for the viewer's own hand
  (`Card`'s `priority` prop, set by `Hand.svelte`) and the hover-zoom
  scan. Don't add it elsewhere: on every tile it means nothing.

## Combat damage beats

When a combat had a first-strike step, the log tags each combat damage
entry with `combat_step`, and the board plays the two steps one after
the other ([ADR 0053](../docs/decisions/0053-combat-damage-beats.md)).
`lib/combatBeats.ts` holds every rule and is unit-tested there: which
log entries are new (with priming and undo), beat membership, arrow
IDs from the log, live vs ghost, the mode and the schedule.
`components/board/CombatArrows.svelte` only measures, draws and
renders.

- Beats are overlays. Never hold the board or input for one: the frame
  already shows its end state.
- The motion gate is `beatMode`, fed `animations.enabled`,
  `animations.damagePopups` and `accessibility.reduceMotion` from the
  settings store. `animations.ts` never sees `reduceMotion`, so
  `gatedDuration` is not enough for a timer or a GSAP tween. With
  motion off, the text cue still plays on the same schedule.
- Anything that should not replay old beats when it changes (today, a
  reconnect and the replay toggle) belongs in `Game.svelte`'s
  `beatsPrimeKey`.

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
