# E2E suite review — 2026-09-08

Scope: `tests-e2e/` (13 files, 33 tests, ~2.5 min on the runner) and the
two workflows that run it. Reviewed at `main` = `704cbea` (post #207).

## Verdict

The suite is well built for its size: real server, real client, no
mocks, assertions on authoritative WS snapshots, helpers that document
*why* rather than *what*. The S19 helper layer (`s19-helpers.ts`) is the
strongest part and should become the pattern for every future
gameplay suite.

The problems are structural, not in the tests themselves:

1. **E2E never runs on PRs.** It fires on push to `main`, nightly, and
   manual dispatch. `b8ec24b` merged green (CI/CD only runs `tsc` on
   tests-e2e) and broke E2E on `main` for four days. This is the single
   most valuable fix.
2. **`mode: "serial"` in the S19 suite hides failures.** One failure
   skips the remaining tests (the Node 20 run reported "1 failed, 23
   passed" — 9 tests never ran). Every S19 test does its own setup, so
   serial mode buys nothing. Remove it.
3. **The core gameplay loop has zero e2e coverage.** Casting a spell
   through the UI (mana, auto-tap, targeting), combat, undo, mulligan,
   reconnect — none are exercised. All S19 tests place cards via admin
   `move_card`, so `cast_spell` (the most-used client action) is only
   covered by Go unit tests and vitest.

## CI / workflow

| Issue | Recommendation |
|---|---|
| E2E not on PRs | Add `pull_request:` to `e2e-nightly.yml` (rename to `e2e.yml`). Concurrency group already cancels superseded runs. Cost: ~5 min of runner time per PR push; the Scryfall cache makes it cheap. |
| No job timeout | Add `timeout-minutes: 20` to the `playwright` job. A wedged `go run` currently blocks the single runner for 6 h. |
| Failures don't annotate PRs | CI reporter: `[["list"], ["github"], ["html", { open: "never" }]]`. |
| `retries: 1` in CI masks flakes silently | Keep the retry, but grep the list output for `flaky` in a follow-up step and warn (or set `failOnFlakyTests: true` once the seed-hand issue below is fixed). |
| Node version lives in three places | Add `.nvmrc` = `22` at repo root; `setup-node` → `node-version-file: .nvmrc`. `client/package.json` engines still says `>=20.11` — bump to match. |
| `playwright install --with-deps` needs sudo on the runner | Works today; if the runner is ever rebuilt without passwordless apt it will break. Bake the deps into the runner image and drop `--with-deps`. |

## Suite-level

**Duplication.** `joinAsPlayer` exists three times (`full-game`,
`board-layout`, `s19-helpers`) with drifting wait strategies (URL
regex vs. localStorage poll). `COMMANDER_NAME` is defined in both deck
fixtures. Consolidate into `tests/fixtures.ts` using Playwright's
`test.extend` so a test can ask for `twoPlayerGame` and get
`{ admin, caster, opponent, game }` — this is what `setupS19Game`
already does; promote it, drop the ad-hoc versions.

**Selector policy is inconsistent.** README says prefer roles; the
suite uses `p.error`, `ul.games > li`, `.hand-slot`,
`[aria-label='Opponent board']`, `form.filter({hasText:"open"})`.
The aria-label ones should be `getByRole("region", { name: "Opponent
board" })`; the CSS ones need either a role or a `data-testid` (the
client currently has zero test ids). Not urgent, but every one of
these is a future silent breakage when styles move.

**Seeding by draw-pumping is the main flake source.**
`seedHandWithCard` fires `draw_card` up to 105 times to surface a
named card from a redacted library. It is slow (dominates the 3–5 s
S19 setups), it fires draw triggers as a side effect (the Tithe test
has to work around it), and its 2 s per-draw `waitFor` is where a
loaded runner will lose races. Two options, in order of preference:

- Server: honour a `CMDCTRL_DEV_SEED` env var so shuffles are
  deterministic, and order the fixture decks so the cards under test
  are on top. No new admin surface; deterministic games also make
  replay tests possible.
- Server: a dev-only `tutor` admin action (`move_card` from library by
  name, admin role only, gated on `CMDCTRL_DEV_*`). Simpler, but adds
  an action that must never ship enabled.

**Stale comments and docs** (cheap, do in one pass):

- `game.spec.ts` header still describes a PixiJS canvas.
- `full-game.spec.ts` claims "the lobby upload UI has its own coverage
  in lobby.spec.ts" — it doesn't; there is no deck-upload UI test.
- `full-game.spec.ts` prerequisites say "server running with make
  server-dev"; the config spawns it.
- `s19-triggers.spec.ts` describe is "S19 ETB triggers" but now holds
  the Smothering Tithe draw trigger and pick-target flows. Rename to
  "S19 triggers" or split by trigger kind.
- README's `--with-deps` / apt list is fine; add the Node ≥ 22
  requirement and the `WebSocket` reason.

**Misplaced test.** `game.spec.ts` › "admin token wiring" is a smoke
canary and belongs in `smoke.spec.ts`.

**Session seeding.** `lobby.spec.ts` logs in through the UI in
`beforeEach` (4× per run); `game.spec.ts` seeds via `addInitScript`.
Use the init-script approach everywhere except `auth.spec.ts`, which
is the one place the login UI is the subject.

## Per-file

- `smoke.spec.ts` — fine. Fast, no state.
- `auth.spec.ts` — fine. `p.error` → give the error a `role="alert"`
  and use `getByRole("alert")`.
- `join.spec.ts` — good; the "real game + bogus token" case is the
  right shape.
- `lobby.spec.ts` — the "start absent with zero seats" assertion is
  weak (asserts on absence only). Add the positive path: two seats +
  two decks → start button enabled → click → game state active. That
  is the DeckUploadForm coverage that's currently missing.
- `game.spec.ts` — the WS-handshake test with the framereceived
  listener comment is excellent. Keep.
- `board-layout.spec.ts` — overlaps `full-game` setup entirely; fold
  its zone assertions into `full-game` (after keep) or move both onto
  the shared fixture.
- `full-game.spec.ts` — the raw-WS final assertion is fine. Missing:
  an actual mulligan (both tests click Keep), and the turn/phase
  display asserting step names as turns pass.
- `s19-helpers.ts` — strong. `resolveStack` is best-effort by design;
  its 20-attempt bound is right. `adminMoveByName`'s `_srcHint`
  parameter is dead — remove it.
- `s19-triggers.spec.ts` — the Yes/No pairs are the correct shape.
  Assertions on `pending_choices` length being 0 are repeated in
  every test; a `expectNoPendingChoices(v)` helper would cut noise.
- `deck-fixture.ts` / `s19-deck-fixture.ts` — merge into `decks.ts`.
- `lobby-api.ts` — fine. The 503 retry loop is the right place for
  index-loading tolerance.

## Coverage gaps, prioritised

1. **Cast a spell through the UI.** Hand card → cast → auto-tap
   preview → stack → resolve → battlefield. This is the product. One
   test with a Forest + a green creature covers `cast_spell`,
   `AutoTapPreviewModal`, `StackOverlay`, `pass_priority`.
2. **E2E on PRs** (above) — not coverage, but it's what makes the
   rest matter.
3. **Lobby deck upload via the UI** (`DeckUploadForm`) and the start
   gate. Today decks only go in via the admin HTTP helper.
4. **Combat.** Declare attacker → declare blocker → damage → life
   totals. `CombatArrows`, `declare_attacker`, `declare_blocker` have
   no e2e.
5. **Mulligan** (not keep): hand size 7 → 6, bottom a card.
6. **Undo** — `undos_remaining` decrement and state restore.
7. **Reconnect** (#202): kill the WS from the page, assert the client
   re-attaches and receives a fresh snapshot. This is exactly the kind
   of behaviour that regresses silently.
8. **3–4 players.** S12 go-live is four players; every test is two.
   One 4-seat smoke (join, start, pass around the table) would catch
   seat-ordering and priority-rotation bugs the 2-player tests can't.
9. **Spectator** — joins without a seat, sees redacted hands.
10. Lower: concede, votes/monarch/initiative, settings persistence,
    bug-report modal (has vitest coverage).

Deliberately not recommended: visual/screenshot tests (layout is
still moving), and per-card e2e tests for every S19/S20 card — the Go
catalog tests own card semantics; e2e should cover one card per
*prompt shape* (mandatory, optional Yes/No, pick-target, pay-unless),
which is what the suite already does.

## Suggested order

1. `pull_request` trigger + `timeout-minutes` + github reporter + drop
   `mode: "serial"` — one small PR, no test changes, biggest payoff.
2. Stale-comment pass + `.nvmrc` + move the misplaced test.
3. Shared fixture (`test.extend`) replacing the three `joinAsPlayer`s
   and both deck files.
4. Cast-a-spell test and lobby upload/start test on the new fixture.
5. Deterministic shuffle seed on the server; rewrite
   `seedHandWithCard` to rely on it.
6. Combat, mulligan, undo, reconnect, 4-player — one test each, in
   that order.
