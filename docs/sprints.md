# Sprint plan

Work on `cmd_and_ctrl` is organised into **2-week sprints**, budgeted at ~10
hours/week (~20 hours/sprint). Sprints map to the phases in
[PLAN.md](../PLAN.md#6-phased-roadmap) but break them down into concrete,
trackable units of work.

**Every commit and PR must reference its sprint and issue.** See
[AGENTS.md](../AGENTS.md#4-sprint-and-tracking-discipline) for the format.

**Project board:** https://github.com/users/krakenhavoc/projects/5
**Milestones:** https://github.com/krakenhavoc/cmd_and_ctrl/milestones

Each sprint has a single tracking issue in this repo; sub-tasks live in that
issue's checklist. Commits reference the sprint issue number.

## Path: Option B first, then B→C

S01–S12 build **Option B** — a polished sandbox client with a Go backend,
TypeScript client, and no rules enforcement. S12 ends with a real 4-player
Commander game on a deployed stack.

S13+ is the **B→C phase**: incremental rules enforcement grafted onto the
Go server, one mechanic and one card at a time, driven entirely by what
real games at our table demand. That track is intentionally open-ended and
planned just-in-time from the S12 pain-point triage.

---

## Sprint index

| # | Name | Phase | Issue | Due | Status |
|---|---|---|---|---|---|
| S01 | Go server + client scaffold | 0 | [#1](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1) | 2026-04-24 | **done** |
| S02 | Core game state: zones + turns | 1 | [#2](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2) | 2026-05-08 | **done** |
| S03 | Action protocol + state deltas | 1 | [#3](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3) | 2026-05-22 | **done** |
| S04 | Lobby, auth, Scryfall pipeline | 2 | [#4](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4) | 2026-06-05 | **done** |
| S05 | Deck import + 4-player table layout | 3 | [#5](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5) | 2026-06-19 | **done** |
| S06 | Hand, battlefield, zones UI | 3 | [#6](https://github.com/krakenhavoc/cmd_and_ctrl/issues/6) | 2026-07-03 | **done** |
| S06.5 | Dynamic deck import from URLs (mini) | 3 | [#37](https://github.com/krakenhavoc/cmd_and_ctrl/issues/37) | 2026-07-10 | **done** |
| S07 | Turn/phase UI + chat + manual priority | 3 | [#7](https://github.com/krakenhavoc/cmd_and_ctrl/issues/7) | 2026-07-17 | **done** |
| S08 | First playable sandbox (2-player, milestone) | 4 | [#8](https://github.com/krakenhavoc/cmd_and_ctrl/issues/8) | 2026-07-31 | **done** |
| S08.5 | Game-logic cleanup pass (wave 1: gating + room + import) | 4 | [#43](https://github.com/krakenhavoc/cmd_and_ctrl/issues/43) | 2026-05-03 | **done** |
| S09 | Polish I — animations + VFX | 5 | [#9](https://github.com/krakenhavoc/cmd_and_ctrl/issues/9) | 2026-08-14 | **done** |
| S10 | Polish II — Commander UX (cmd damage, politics) | 5 | [#10](https://github.com/krakenhavoc/cmd_and_ctrl/issues/10) | 2026-08-28 | **done** |
| S11 | Polish III — hover preview, undo, spectator | 5 | [#11](https://github.com/krakenhavoc/cmd_and_ctrl/issues/11) | 2026-09-11 | **done** |
| S11.5 | Per-user settings and preferences (mini) | 5 | [#82](https://github.com/krakenhavoc/cmd_and_ctrl/issues/82) | 2026-09-18 | **done** |
| S12 | Deploy + 4-player go-live with friends | 6 | [#12](https://github.com/krakenhavoc/cmd_and_ctrl/issues/12) | 2026-09-25 | planned |
| S12.5 | Discord identity for players (OAuth + bot + presence) | 6 | [#59](https://github.com/krakenhavoc/cmd_and_ctrl/issues/59) | 2026-10-09 | planned |
| S13 | Priority foundation (rules graft kickoff) | 7 | [#62](https://github.com/krakenhavoc/cmd_and_ctrl/issues/62) | 2026-05-17 | planned |
| S13.1 | Stack: cast/resolve/target/counter/trigger/SBA | 7 | [#63](https://github.com/krakenhavoc/cmd_and_ctrl/issues/63) | 2026-06-14 | planned |
| S13.2 | Counter mechanics (SBAs + player counters + UI) | 7 | [#79](https://github.com/krakenhavoc/cmd_and_ctrl/issues/79) | 2026-06-28 | planned |
| S13.3 | Client-side timing affordance (greyed illegal actions) | 7 | [#99](https://github.com/krakenhavoc/cmd_and_ctrl/issues/99) | 2026-07-04 | planned |
| S13.4 | Interactive cleanup discard + per-player MaxHandSize | 7 | [#103](https://github.com/krakenhavoc/cmd_and_ctrl/issues/103) | 2026-07-11 | planned |
| S13.5 | Card visibility + known-by tracking | 7 | [#108](https://github.com/krakenhavoc/cmd_and_ctrl/issues/108) | 2026-07-25 | planned |
| S14 | Card-effect catalog foundation | 7 | [#64](https://github.com/krakenhavoc/cmd_and_ctrl/issues/64) | 2026-07-12 | **done** |
| S15 | Mana pool, cost model, and auto-tapper | 7 | [#65](https://github.com/krakenhavoc/cmd_and_ctrl/issues/65) | 2026-08-09 | **done** |
| S16 | Continuous effects + layer system (CR 613) | 7 | [#66](https://github.com/krakenhavoc/cmd_and_ctrl/issues/66) | 2026-09-06 | planned |
| S17 | Replacement effects engine (CR 614) | 7 | [#67](https://github.com/krakenhavoc/cmd_and_ctrl/issues/67) | 2026-10-04 | planned |
| S18 | Combat keywords | 7 | [#68](https://github.com/krakenhavoc/cmd_and_ctrl/issues/68) | 2026-11-01 | planned |
| S19 | Auto-fire triggered abilities | 7 | [#69](https://github.com/krakenhavoc/cmd_and_ctrl/issues/69) | 2026-11-29 | planned |
| S20 | Auto-target legality + smart cast UI | 7 | [#70](https://github.com/krakenhavoc/cmd_and_ctrl/issues/70) | 2026-12-27 | planned |
| S21 | Tokens, sacrifice, aristocrats | 7 | [#73](https://github.com/krakenhavoc/cmd_and_ctrl/issues/73) | 2027-01-24 | planned |
| S22 | Card draw + library manipulation | 7 | [#74](https://github.com/krakenhavoc/cmd_and_ctrl/issues/74) | 2027-02-21 | planned |
| S23 | Mass removal + boardwipes | 7 | [#75](https://github.com/krakenhavoc/cmd_and_ctrl/issues/75) | 2027-03-21 | planned |
| S24 | Equipment, auras, attachments | 7 | [#76](https://github.com/krakenhavoc/cmd_and_ctrl/issues/76) | 2027-04-18 | planned |
| S25 | Voltron / commander damage focus | 7 | [#77](https://github.com/krakenhavoc/cmd_and_ctrl/issues/77) | 2027-05-16 | planned |
| S26 | Tribal / creature type matters | 7 | [#78](https://github.com/krakenhavoc/cmd_and_ctrl/issues/78) | 2027-06-13 | planned |
| S27 | Card-type completeness (planeswalkers, sagas, vehicles, battles) | 7 | [#92](https://github.com/krakenhavoc/cmd_and_ctrl/issues/92) | 2027-07-04 | planned |
| S28 | Cost modification + alternative casts | 7 | [#93](https://github.com/krakenhavoc/cmd_and_ctrl/issues/93) | 2027-07-25 | planned |
| S29 | Alt-cast paths from non-hand zones (flashback, suspend, foretell, …) | 7 | [#94](https://github.com/krakenhavoc/cmd_and_ctrl/issues/94) | 2027-08-15 | planned |
| S30 | Damage prevention, cloning, face-down, deferred protection keywords | 7 | [#95](https://github.com/krakenhavoc/cmd_and_ctrl/issues/95) | 2027-09-05 | planned |
| Post-S30 | Rolling deck-driven catalog growth | 7 | TBD at S30 retro | rolling | not started |
| S31 | AI bot seat (heuristic policy) | 8 | [#89](https://github.com/krakenhavoc/cmd_and_ctrl/issues/89) | 2027-09-26 | planned |

---

## S01 — Go server + client scaffold
**Phase:** 0 · **Goal:** stand up both services and prove one round-trip.

- [ ] `server/` Go module (`go mod init`), with lint (`golangci-lint`), format (`gofmt`), test (`go test`), `make dev` target
- [ ] `client/` TypeScript + Vite app with ESLint, Prettier, Vitest
- [ ] Pick WebSocket library (gorilla recommended); decision recorded in `docs/decisions/0001-ws-library.md`
- [ ] Pick client framework (React or Svelte); decision recorded in `docs/decisions/0002-client-framework.md`
- [ ] WebSocket server on :8080 accepting JSON frames
- [ ] Client sends a `ping` action frame; server broadcasts a state delta; client logs it
- [ ] Draft `docs/protocol.md` — v0 schema for action frames and state delta frames

**Exit criteria:** `make dev` in `server/` + `pnpm dev` (or equivalent) in `client/` produces a working round-trip, and the protocol schema is written down.

---

## S02 — Core game state: zones + turns
**Phase:** 1 · **Goal:** authoritative in-memory game state covering 4-player Commander structure.

- [ ] Domain types: `Game`, `Player`, `Zone`, `Card`, `Turn`, `Phase`, `Step`
- [ ] Zones: library, hand, battlefield, graveyard, exile, command zone, stack
- [ ] Turn structure: untap, upkeep, draw, main1, combat phases, main2, end, cleanup
- [ ] Game lifecycle: create, join (up to 4), start, end
- [ ] Unit tests for state mutations (pure functions where possible)
- [ ] Life totals start at 40, commander damage tracked as 4×4 matrix per game

**Exit criteria:** a Go test starts a 4-player game, advances through a full turn cycle, and asserts the expected state at each step.

---

## S03 — Action protocol + state deltas
**Phase:** 1 · **Goal:** clients drive the game state via the action protocol.

- [ ] Action types: `draw_card`, `play_card`, `move_card` (between zones), `tap`, `untap`, `untap_all`, `pass_priority`, `pass_turn`, `mulligan`, `shuffle_library`, `change_life`, `add_counter`, `set_commander_damage`
- [ ] Server applies action → computes state delta → broadcasts to all players in the game
- [ ] Delta format versioned in `docs/protocol.md`
- [ ] `cmd/gamecli/` — CLI test client that drives a game via the protocol
- [ ] Snapshot tests of a recorded game transcript (golden files)
- [ ] Crash recovery: server writes JSON snapshots after every action

**Exit criteria:** the CLI plays a full scripted turn (draw, play land, cast creature, attack, pass turn) end-to-end against the server and the resulting state matches a golden file.

---

## S04 — Lobby, auth, Scryfall pipeline ✅
**Phase:** 2 · **Goal:** the non-play parts of the app.

- [x] HTTP endpoints: `POST /games`, `GET /games/:id`, `POST /games/:id/join`, `POST /games/:id/start`, `GET /me`
- [x] Pluggable `auth.Authenticator` interface; S04 default is an in-memory invite store (admin token + per-game invite tokens). Swap for stateless HMAC later without touching handlers.
- [x] Invite links (`#/games/<id>/join?t=<token>`); session cookie + bearer + query-param transports
- [x] RoomManager for multi-game routing; hub broadcasts scoped per game
- [x] Per-connection visibility filter (opponent `hand.cards` + `library.cards` hidden; counts preserved)
- [x] Scryfall bulk download script (`scripts/scryfall-refresh.sh`) + weekly cron line
- [x] Streaming Scryfall index loader + disk-backed image cache (sharded layout, per-id dedup)
- [x] Client lobby: hash router, login / lobby / join / game routes, localStorage-persisted session

**Exit criteria:** two browser tabs log in (admin creates, second tab opens invite), create/join a game, and see each other in the lobby; any card can be looked up by ID and served as an image.

**Deferred:** lobby persistence across server restart (S12 deploy work); TLS (reverse proxy); login rate limiting; per-invite revocation. See [ADR 0003](decisions/0003-auth-and-lobby.md).

---

## S05 — Deck import + 4-player table layout ✅
**Phase:** 3 · **Goal:** bring decks into games and render a 4-player table.

- [x] Moxfield export parser (JSON)
- [x] Plain-text parser (`1x Sol Ring` / `SB:` / `*CMDR*` dialect)
- [x] Validation: 100-card singleton, commander legality, color identity, format legality
- [x] `POST /games/{id}/decks` endpoint — RolePlayer may only set own deck; admin may set any
- [x] Banlist snapshot: commander legality comes from the Scryfall index loaded at server start; mid-game rotations don't invalidate existing decks
- [x] Partner / companion surfaced as `ErrUnsupportedMechanic` — decks using them fail loudly rather than silently drop the second commander
- [x] PixiJS play area canvas mounted in the Game route
- [x] 4-player table layout: viewer at bottom, opponents clockwise at left / top / right
- [x] Placeholder card "backs" (solid rectangles) — Scryfall's card back is Wizards' IP, deferred to custom asset work
- [x] Client lobby: per-seat deck-status badges, inline upload form, Start gated on all-uploaded

**Deferred:** Archidekt JSON parser, `.cod`, `.dek`, `.txt`-file-upload parsers — all follow-ups; Moxfield + plain-text covers ~90% of users.

**Exit criteria:** import a Moxfield or plain-text decklist, create a game, and see your seat at a 4-player table with the right zones laid out.

---

## S06 — Hand, battlefield, zones UI
**Phase:** 3 · **Goal:** render and manipulate the core game zones.

- [x] Hand: fan layout, hover lift, drag to battlefield
- [x] Battlefield: free card placement, tap/untap via click
- [ ] Battlefield: token / copy stacking — deferred (tokens don't enter the game until the rules graft track; S13+)
- [x] Library — clickable (draws top card on tap)
- [ ] Graveyard / exile / command — clickable modal browser (S07 alongside turn/phase UI)
- [ ] Library — searchable / reorderable — deferred to S11 (hover preview / polish)
- [x] All zone state driven by server state deltas (no local truth)
- [x] Card art rendered from the Scryfall cache

**Exit criteria:** a player draws 7 cards, plays a land, casts a creature, taps a permanent — all visible, all driven by server-authoritative deltas. ✅

---

## S06.5 — Dynamic deck import from URLs (mini)
**Phase:** 3 · **Goal:** skip the copy-paste step — paste a Moxfield / Archidekt deck URL and have the server fetch, parse, and validate.

One-week mini sprint slotted between S06 and S07 to address UX friction surfaced during S06 smoke testing. The paste-based flow works but is tedious for 100-card decks, and every paste-era bug on the S06 branch (single-slash MDFCs, art-series name collisions) would have been avoided by fetching the structured JSON directly from the source of truth.

- [x] Moxfield API client — `GET https://api2.moxfield.com/v3/decks/all/{id}` (public decks only)
- [x] Archidekt API client — `GET https://archidekt.com/api/decks/{id}/`, new parser for its JSON shape
- [x] `POST /games/{id}/decks` grows a `format: "url"` source type; handler dispatches on host
- [x] Outbound HTTP hygiene: 10s timeout, descriptive User-Agent
- [ ] Short in-process cache of (URL → parsed result) — deferred as polish; the external APIs are fast enough that first-call latency is acceptable at one-user scale
- [x] New structured violations (`external_api_unavailable`, `deck_not_found`, `deck_private`, `unknown_source`) rendered via the existing 422 shape
- [x] Client lobby: auto-detect URL-vs-JSON-vs-text on paste; spinner while fetching (shows target hostname)

**Deferred:** TappedOut / Deckstats / EDHREC parsers (add incrementally); authenticated imports of private decks; Scryfall `/cards/collection` batch name resolution; URL-result cache (flagged above).

**Exit criteria:** paste `https://moxfield.com/decks/<id>` or `https://archidekt.com/decks/<id>/…` into the deck-upload textarea, click upload, see the parsed + validated deck install on the seat. Same UX as paste, zero manual conversion. ✅

---

## S07 — Turn/phase UI + chat + manual priority
**Phase:** 3 · **Goal:** the fiddly but essential parts of a manually-driven game.

- [x] Turn / phase / step indicator
- [x] "Pass priority" button per player
- [x] "Pass until end of turn" shortcut
- [x] Any-player chat, with player colors and timestamps
- [x] Quick-action buttons: draw, shuffle, mulligan, untap all, change life
- [x] Priority indicator ("any responses?") visible across all 4 seats

**Exit criteria:** two players can play a complete turn cooperatively, passing priority manually, with chat, and it feels like a game — not a demo.

**Status:** done. Server-tracked priority + chat broadcast landed in [`ef211be`](https://github.com/krakenhavoc/cmd_and_ctrl/commit/ef211be); client turn bar, pass button, chat panel, and quick-action toolbar in [`0d2d979`](https://github.com/krakenhavoc/cmd_and_ctrl/commit/0d2d979). Note: the chat panel was later removed from the in-game view in S08.5 (wire protocol stays live).

---

## S08 — First playable sandbox (milestone)
**Phase:** 4 · **Goal:** complete a 2-player sandbox game in a browser, end-to-end.

- [x] Mulligan flow (take 7, keep/mulligan, scry for extra mulligans — or simplified to "take N")
- [x] Combat UI: declare attackers, declare blockers, manually resolve damage
- [x] Life tracker with history
- [x] Win/loss screen (manual; player clicks "I lose")
- [x] Play against yourself across two tabs, record a video

**Exit criteria:** a recorded 2-player sandbox game, start to finish, with no manual state intervention outside the client.

**Status:** done. End-to-end 2-player flow is covered by [tests-e2e/tests/full-game.spec.ts](../tests-e2e/tests/full-game.spec.ts) (join → deck upload → mulligan → combat → turn passing), landed in [`d8ee7f2`](https://github.com/krakenhavoc/cmd_and_ctrl/commit/d8ee7f2).

---

## S08.5 — Game-logic cleanup pass (wave 1)
**Phase:** 4 · **Goal:** fix the three frictions that bit hardest in S08 playtest.

Originally deferred post-S08 retro because most candidate items would be subsumed by the S13+ rules graft. Re-activated for **wave 1** when 2-player play surfaced three items that don't overlap with rules work — they're UX/authorization issues that hurt now and the rules engine wouldn't help with later. The broader cleanup (turn-1 skip-draw, commander damage attribution, command-zone tax, token API, mulligan penalty, etc.) stays deferred to S13+.

**Wave 1 scope** ([issue #43](https://github.com/krakenhavoc/cmd_and_ctrl/issues/43)):

- [x] **Controller-only card interactions.** Server + client gate on `caller == card.controller` for `tap`, `untap`, `move_card`, `add_counter`, `set_battlefield_position`, `declare_attacker`, `declare_blocker`. Admins still bypass. ([`e8a70ff`](https://github.com/krakenhavoc/cmd_and_ctrl/commit/e8a70ff))
- [x] **Battlefield real-estate.** Remove the chat panel from the in-game UI (players use Discord); slim turn bar + toolbar; canvas grows from ~78%×85% to ~99%×92% of viewport. Chat wire-protocol stays intact for future re-mount. ([`36c66ea`](https://github.com/krakenhavoc/cmd_and_ctrl/commit/36c66ea))
- [x] **In-game deck import.** Players who land in `Game.svelte` via invite see a deck-import modal when the game is in lobby state and their library is empty — no need to navigate back to the lobby. ([`dbbb2ab`](https://github.com/krakenhavoc/cmd_and_ctrl/commit/dbbb2ab), [`f596057`](https://github.com/krakenhavoc/cmd_and_ctrl/commit/f596057))

**Still deferred to S13+** (rules-graft would reshape these): turn-1 skip-draw, commander damage attribution, partner / companion deck imports, commander returns + tax, the Stack zone, token API, comprehensive move-card cleanup, mulligan penalty.

**Exit criteria:** all three wave-1 items shipped; a second 2-player playtest confirms (a) no cross-player card mutations, (b) the canvas feels uncramped, (c) a brand-new player can join and import without leaving the game route.

**Status:** done. All three wave-1 items shipped despite the `88d5cee` commit briefly marking S08.5 deferred; the deferral was reversed when the items were pulled in ahead of S13.

---

## S09 — Polish I — animations + VFX
**Phase:** 5 · **Goal:** make the table feel alive.

- [x] GSAP integration (#58)
- [x] Card draw / play / tap / untap animations (#58, #61)
- [x] Damage number popups, combat arrows (#72, #81)
- [x] Sound pass (draw, play, tap, damage, turn change) — assets (#106) + Svelte-seam wiring + toolbar mute toggle (#107). See [s09-sound-pass.md](s09-sound-pass.md) for the event → cue mapping.
- [x] Particle effects on ETB / death — shipped as a scale-pulse in #72; full particle system folded back in if/when needed

**Exit criteria:** side-by-side video vs. Cockatrice shows a clear "this feels better" delta.

---

## S10 — Polish II — Commander UX
**Phase:** 5 · **Goal:** the Commander-specific differentiators from PLAN.md §5.

- [x] Command zone as a first-class UI element (#86)
- [x] Commander damage 4×4 grid, always visible (#88)
- [x] Life tracker with history, poison/infect/energy controls (#87)
- [x] Monarch / initiative / goad markers (#85, #87)
- [x] Politics UI scaffold (promise tokens) (#91)
- [x] Council's dilemma / voting UI (#91)

**Exit criteria:** all Commander-specific affordances from PLAN.md §5 are at least rough-rendered.

---

## S11 — Polish III — hover preview, undo, spectator
**Phase:** 5 · **Goal:** the Arena-feel features that Cockatrice and XMage both lack.

- [x] Hover a card → detail panel with full Oracle text and current game context ([#98](https://github.com/krakenhavoc/cmd_and_ctrl/pull/98))
- [x] One-click undo / rewind to last server snapshot — per-player per-turn budget (default 1, admin-configurable via `set_undo_limit`), caller-gated so players can only undo their own actions ([#98](https://github.com/krakenhavoc/cmd_and_ctrl/pull/98))
- [x] Spectator mode (read-only connection to a game) — separate `SpectatorInvite` token, `ReadOnly` binding at the WS layer, uniform grid layout without perspective rotation ([#101](https://github.com/krakenhavoc/cmd_and_ctrl/pull/101))
- [x] Auto-saved replays (server writes every delta to a game log file) — append-only JSONL at `<dumpDir>/replays/<game-id>.jsonl`, lobby "download replay" link gated by admin or seated player ([#102](https://github.com/krakenhavoc/cmd_and_ctrl/pull/102))

**Exit criteria:** hover preview works for every card, undo works for the last action, spectating a live game works without interfering.

**Status:** done.

---

## S11.5 — Per-user settings and preferences (mini)
**Phase:** 5 · **Goal:** one canonical settings panel that lets each player customise sound, animations, display, gameplay, and accessibility. Preferences persist in `localStorage`, respect OS-level accessibility hints, and unify the ad-hoc toggles otherwise scattered across S09/S10/S11/S12.5/S13. Lands before S12 deploy so friends' first real games open with configurable defaults.

**Slot rationale.** Ships *after* the three polish sprints (S09 animations + sound, S10 Commander UX, S11 hover / undo / spectator) so the settings panel has real toggles to surface, and *before* S12 deploy so the go-live build ships with settings in place. 1-week mini-sprint scope — the infrastructure is small (single Svelte store + JSON schema + panel UI); most of the surface is wiring existing features up to toggles.

**Storage decision.** Client-only `localStorage`. Matches project ethos (hobby scale, 4–8 users, no DB). Cross-device sync becomes feasible post-S12.5 using Discord ID as the key; documented as a future follow-up, not this sprint.

### Tasks

**Settings infrastructure (client):**
- [x] `client/src/lib/settings.ts` — Svelte `writable` store, typed `Settings` interface, schema versioning (`cmdctrl.settings.v1`) with a migration function shape ready for future bumps ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))
- [x] Default values honour `window.matchMedia('(prefers-reduced-motion: reduce)')` at first load; subsequent OS changes are live-observed ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))
- [x] Reactive apply — changes take effect without page reload ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))
- [x] Export / import as JSON (copy-to-clipboard + paste) for moving settings between devices without a server ([#119](https://github.com/krakenhavoc/cmd_and_ctrl/pull/119))
- [x] Reset-to-defaults button — global only; per-section reset deferred ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116) + [#119](https://github.com/krakenhavoc/cmd_and_ctrl/pull/119))
- [ ] Absorb S13's "per-step stops" `localStorage` preferences — **deferred to S13.3** ([#99](https://github.com/krakenhavoc/cmd_and_ctrl/issues/99)) when the per-step-stops mechanism actually lands. Schema has the empty `gameplay.stepStops` slot reserved.

**Settings panel UI (Svelte):**
- [x] New `client/src/lib/components/Settings.svelte` — modal panel with sidebar tabs ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))
- [x] Gear icon in `Game.svelte` header + keyboard shortcut `,` (comma); reachable from `Lobby.svelte` ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))
- [x] Each setting = label + control + short help text ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))
- [x] Changes apply instantly with a "saved ✓" flash beside the changed control ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))

**Audio:**
- [x] Master volume slider (0–100) ([#117](https://github.com/krakenhavoc/cmd_and_ctrl/pull/117))
- [x] Effects volume — folded with master into a single multiplier in `sounds.ts` ([#117](https://github.com/krakenhavoc/cmd_and_ctrl/pull/117))
- [x] Music volume — slider exists but no-op until a music track ships ([#117](https://github.com/krakenhavoc/cmd_and_ctrl/pull/117))
- [x] Mute-all toggle ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))
- [ ] `M` keyboard shortcut + per-category test buttons — deferred (low value; mute toggle in modal works)

**Animations:**
- [x] Animations master toggle — auto-off when `prefers-reduced-motion: reduce` on first load ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116) + [#117](https://github.com/krakenhavoc/cmd_and_ctrl/pull/117))
- [x] Per-animation toggles: card draw / play / tap / untap / flip ([#117](https://github.com/krakenhavoc/cmd_and_ctrl/pull/117))
- [x] Particle effects on ETB ([#117](https://github.com/krakenhavoc/cmd_and_ctrl/pull/117))
- [x] Damage-number popups (gates `floatUp`/`fadeOut`) ([#117](https://github.com/krakenhavoc/cmd_and_ctrl/pull/117))
- [x] Animation speed multiplier select: 0.5× / 1× / 1.5× / 2× ([#117](https://github.com/krakenhavoc/cmd_and_ctrl/pull/117))

**Display:**
- [~] Theme select — CSS palette scaffolded for dark / light / high-contrast, but the select is **disabled** until per-component `var()` migration ships. ([#119](https://github.com/krakenhavoc/cmd_and_ctrl/pull/119))
- [x] Card size on battlefield: small / medium / large — drives `--card-scale` (self) + `--card-scale-opponent` (gentler curve to avoid opponent-row overflow) ([#118](https://github.com/krakenhavoc/cmd_and_ctrl/pull/118))
- [x] Hand layout: fan (default) / stacked ([#118](https://github.com/krakenhavoc/cmd_and_ctrl/pull/118))
- [ ] Auto-rotate battlefield to viewer POV — already on by design; the toggle isn't surfaced because no use-case has appeared
- [x] Card tooltip hover delay (0–1000 ms) ([#118](https://github.com/krakenhavoc/cmd_and_ctrl/pull/118))
- [x] Show opponent hand count ([#118](https://github.com/krakenhavoc/cmd_and_ctrl/pull/118))
- [ ] Show mana pip icons vs. text — deferred (no toggle implemented)

**Gameplay:**
- [ ] Per-step stops grid — **deferred to S13.3** ([#99](https://github.com/krakenhavoc/cmd_and_ctrl/issues/99))
- [~] Auto-pass priority — wired for opponents' turns + empty stack. Full "nothing playable" check waits on S13.3's legality engine ([#118](https://github.com/krakenhavoc/cmd_and_ctrl/pull/118))
- [x] Confirm before exiting an active game (back button + browser `beforeunload`) ([#118](https://github.com/krakenhavoc/cmd_and_ctrl/pull/118))
- [ ] Default targeting — deferred (no targeting UI yet to plug into)
- [ ] Chat panel visibility — deferred (chat UI removed in S08.5, no-op until it returns)
- [ ] Discord Rich Presence toggle — **deferred to S12.5**

**Accessibility:**
- [x] Respect `prefers-reduced-motion` (live media query observer) ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))
- [x] Increased text size: 0.9× / 1.0× / 1.2× / 1.5× via `--font-scale` ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))
- [~] High-contrast mode — palette in CSS scaffold; needs the per-component `var()` migration before the toggle works
- [x] Colour-blind palette toggle persists ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116)); the alternate palette in `colors.ts` is a follow-up
- [x] Focus indicators always visible toggle ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))

**Keybindings:**
- [ ] **Deferred wholesale.** Adding any kind of rebind UI without a conflict-detection layer gives users a way to lock themselves out of the app. The two existing shortcuts (`,` for settings, mute via the modal toggle) are hardcoded for now. Revisit once we have a meaningful set of bindings to manage.

**Advanced:**
- [x] Export settings to clipboard (JSON) ([#119](https://github.com/krakenhavoc/cmd_and_ctrl/pull/119))
- [x] Import settings from clipboard (JSON, validates against schema) ([#119](https://github.com/krakenhavoc/cmd_and_ctrl/pull/119))
- [x] "Copy my settings hash" — FNV-1a fingerprint ([#119](https://github.com/krakenhavoc/cmd_and_ctrl/pull/119))
- [x] "Reset all settings" with confirm ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))

**Docs:**
- [ ] ADR — deferred. Decision rationale captured in PR descriptions + commit messages; an ADR sweep can fold them in later.
- [ ] Client-level "Customising your client" README section — deferred.

**Tests:**
- [x] Schema migration test (v0 → v1 defaults; corrupt-blob fallback; partial-blob field merge; legacy `cmdctrl.muted` absorption) ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))
- [x] Default settings load correctly when `localStorage` is empty ([#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116))
- [ ] `prefers-reduced-motion: reduce` at page load — manually verified, no automated test (jsdom doesn't ship `matchMedia`; not worth a polyfill for one assertion)
- [x] Export → import round-trip produces identical store state ([#119](https://github.com/krakenhavoc/cmd_and_ctrl/pull/119))
- [x] Fingerprint determinism + sensitivity ([#119](https://github.com/krakenhavoc/cmd_and_ctrl/pull/119))
- [ ] Keybinding defaults — N/A, keybindings deferred

### Out of scope (explicit handoffs)
- **Server-backed settings sync across devices** — future post-S12.5 work; feasible using Discord ID as the key once S12.5 ships
- **Full keybinding re-bind UI** — deferred to a follow-up mini-sprint unless the 1-week scope has spare budget (JSON-edit + reset-to-defaults ships here)
- **Custom themes beyond dark / light / high-contrast** — scaffolding ships, full theme catalog is a follow-up
- **Card-art customisation** — separate concern; Scryfall assets only for personal use
- **Admin-wide / server-wide settings** (deck cap, game timeout, lobby limits) — operator domain, not per-user
- **Per-game settings** ("music only in lobbies") — one level too deep for a hobby tool

### Risks / gotchas
- S09 / S10 / S11 are in flight. This sprint assumes the polish work has landed; if S09 slips, S11.5's audio/animation toggles are wiring toggles into features that don't exist yet. Mitigation: order S11.5 strictly *after* S11 rather than overlapping; scope the toggles that refer to unshipped features as no-ops with a "feature not yet available" disabled state.
- `prefers-reduced-motion` is a **live** media query — settings must observe changes mid-session, not just at load, or a user who toggles OS reduced-motion mid-game won't see the effect.
- `localStorage` is per-origin, per-browser. Switching browsers or clearing site data wipes settings. Document this and ship the export/import path as the mitigation.
- Keybindings are the deepest rabbit hole. Time-box aggressively — `Reset to defaults` + JSON-edit is a full backstop; polished per-key rebinding UI waits.
- S13's per-step-stops `localStorage` prefs land on a different track (rules graft, due 2026-05-17 — well before S11.5). Ship a migration so players don't lose their stops config when S11.5's schema takes over.

### Exit criteria
1. A gear icon opens the settings panel from the game header, the lobby, and the login screen. Keyboard `,` opens it from anywhere.
2. Every setting persists across page reloads, tab closures, and re-opens of the browser.
3. Muting sound effects makes the S09 card-tap sound silent immediately — no reload required.
4. Disabling animations makes the S09 card-tap animation static immediately.
5. Booting with OS `prefers-reduced-motion: reduce` auto-disables non-essential animations at first load.
6. Changing card size re-flows the battlefield tiles without a reload.
7. Export settings → paste into another browser → same config applies after import.
8. S13's per-step-stops preferences are visible and editable from this panel (once S13 has shipped); existing S13 users don't lose their stops config on migration.

### Status
**Done** across four PRs: [#116](https://github.com/krakenhavoc/cmd_and_ctrl/pull/116) infrastructure + modal shell, [#117](https://github.com/krakenhavoc/cmd_and_ctrl/pull/117) audio + animation consumers, [#118](https://github.com/krakenhavoc/cmd_and_ctrl/pull/118) display + gameplay consumers, [#119](https://github.com/krakenhavoc/cmd_and_ctrl/pull/119) export/import + fingerprint + theme palette scaffold.

Deferred, with concrete pickup points:
- **Per-step stops grid (Gameplay tab)** — waits on S13.3's priority-aware stops mechanism; schema slot reserved.
- **Theme live-apply (Display tab)** — CSS palette scaffolded; needs a sweep to migrate hardcoded hex colours in `PlayerPanel.svelte` / `Card.svelte` / `Game.svelte` / etc. to `var(--bg)`/`var(--fg)`. Select is disabled in the UI until then.
- **Keybinding rebind UI** — deferred wholesale; revisit once more shortcuts exist.
- **ADR + README "Customising your client"** — decision context lives in the four PR descriptions for now.

---

## S12 — Deploy and go live with friends
**Phase:** 6 · **Goal:** real games, real feedback.

- [x] VPS provisioning script (Docker Compose or plain systemd units) — systemd unit on a LAN box at cmd.labxp.io
- [x] Deploy Go game server + static client — GitHub Actions CI/CD landed in [#111](https://github.com/krakenhavoc/cmd_and_ctrl/pull/111); scryfall bulk-dump refresh in [#112](https://github.com/krakenhavoc/cmd_and_ctrl/pull/112) / [#113](https://github.com/krakenhavoc/cmd_and_ctrl/pull/113)
- [x] Basic observability (structured logs, uptime ping, crash dumps) — server emits JSON slog, crash-recovery snapshot dumps in `CMDCTRL_DATA_DIR`, `systemctl status` as live health signal
- [ ] Play a real 4-player game with friends
- [ ] Triage top 10 pain points from the real game into the S13+ backlog

**Exit criteria:** a real 4-player Commander game happens on the deployed stack, and a prioritised S13+ backlog exists.

**Status:** partial. Deploy infrastructure is live (auto-deploy on merge to main, scheduled scryfall refresh, nightly e2e). The 4-player playtest + pain-point triage remain open.

---

## S12.5 — Discord identity for players
**Phase:** 6 · **Goal:** the playgroup uses Discord for coordination already (S08.5 removed in-game chat in favour of it); lean all the way in. Invite links opened from Discord land the player in the game with their Discord display name and avatar; a bot posts/DMs invites from Discord itself; the player's presence reflects what they're doing in the game.

Every "out of scope" deferral from the initial planning pass is pulled into this sprint — the user decided the full integration is worth a single ~2-week sprint rather than chaining three mini-sprints.

### Tasks

**Discord OAuth (core):**
- [ ] Register the Discord application; `CMDCTRL_DISCORD_CLIENT_ID` / `CMDCTRL_DISCORD_CLIENT_SECRET` env vars (dev + prod). Callback URLs for both environments registered with Discord.
- [ ] Server routes: `GET /auth/discord/start?game=<id>&t=<invite>` builds the Discord authorize URL with PKCE + state; `GET /auth/discord/callback` exchanges the code, calls `/users/@me`, completes `Lobby.Join` on the bound invite, mints the session.
- [ ] Server-side state store (`game_id`, `invite_token`, `pkce_verifier`, 5-minute TTL) — reuse the pattern from `auth.MemoryAuthenticator`; no schema migration.
- [ ] Extend `auth.Principal` with optional `DiscordID`, `DiscordUsername`, `DiscordAvatarHash`, `DisplayName`. `MemoryAuthenticator` preserves them across `Validate`.
- [ ] Extend `lobby.SeatInfo` with `display_name` (Discord `global_name`, fallback username, final fallback manual name) and `avatar_url`.
- [ ] Client: `Join.svelte` gets a primary "Sign in with Discord" button; manual name-entry kept as fallback.

**Avatar rendering:**
- [ ] `client/src/lib/protocol.ts` — `PlayerView` / `SeatInfo` grow the two optional fields.
- [ ] `PlayerHeader.svelte` — render a 24px circular avatar before `.seat-dot` when present.
- [ ] Lobby seat list, any S10 politics / commander-damage UI that names seats — pick up `displayName` / `avatarUrl`.
- [ ] Server-side avatar cache (`$CMDCTRL_DATA_DIR/avatars/<discord_id>/<hash>.png`): on first fetch hit `cdn.discordapp.com`, cache with immutable headers, re-fetch when hash changes. Client requests `/avatars/<discord_id>`; server serves from cache or proxies on miss. (Avoids embedding Discord CDN URLs directly in the state stream.)

**Discord bot (`cmd_and_ctrl-bot`):**
- [ ] New top-level directory `bot/` or a cmd under `server/cmd/bot/` — pick one in the ADR. Go, using `bwmarrin/discordgo`.
- [ ] Slash command `/cc-invite [name]` — calls server `POST /games` with admin credentials (bot holds `CMDCTRL_ADMIN_TOKEN` via env), posts the invite link back to the channel (ephemeral or channel-visible, configurable).
- [ ] Slash command `/cc-invite-dm @user [name]` — creates the game and DMs the invite to `@user`, pre-binding the invite's `DiscordID` so the OAuth round-trip on click is a no-op if they're already signed in.
- [ ] Slash command `/cc-games` — lists active/lobby games known to the server; `/cc-end <id>` admin-only shutdown.
- [ ] Bot deploys as a second systemd unit on the same VPS (S12 infra). Shares the server's VPS data dir via env only; no DB.

**Rich Presence:**
- [ ] Opt-in toggle in the client ("Show this game on Discord"); stored in `localStorage` alongside the session.
- [ ] When enabled, client uses Discord's RPC over the local IPC socket (`discord-rpc` from npm or a thin WebSocket wrapper) to publish presence: "In a Commander game — Turn 5, 3 opponents alive".
- [ ] Presence updates on phase change + life change (throttled to 1 update / 15 s to stay inside Discord's rate limits).
- [ ] Presence clears on game end / browser close.

**Re-link after the fact:**
- [ ] `GET /auth/discord/link` — already-signed-in player reopens the invite/OAuth loop to attach (or swap) their Discord identity onto an existing seat without leaving the game.
- [ ] Client surfaces a "Link Discord" row in a session-settings panel (new — probably a small menu in the corner of `Game.svelte`).
- [ ] Server merges Discord fields onto the existing `Principal` + broadcasts a `SeatInfo` update delta so opponents immediately see the avatar/name swap.

**Docs:**
- [ ] `docs/decisions/0004-discord-identity.md` — ADR. Why OAuth + bot + presence all together; PKCE + state handling; bot deployment shape; opt-in scopes; what happens when a user revokes the Discord token.
- [ ] `docs/lobby.md` — document the new auth routes + the `SeatInfo` field additions.
- [ ] `AGENTS.md` §5 — env vars (`CMDCTRL_DISCORD_CLIENT_ID` / `_SECRET`, bot token, RPC client ID).

### Risks / gotchas
- OAuth callback URLs must be whitelisted per environment — document dev, staging, prod in the ADR.
- Bot token compromise gives admin access to the server (it holds `CMDCTRL_ADMIN_TOKEN`). Store it in `systemd` credential, not a plain `.env`.
- Discord Rich Presence over IPC requires the Discord desktop client running on the user's machine; it silently no-ops on web-only / mobile. Document the gap, don't treat it as a failure mode.
- Avatar hash changes when a user updates their Discord avatar; our cache must key on `discord_id + hash`, not `discord_id` alone, or we'll serve stale avatars for hours.
- Ratelimits: `/users/@me` is generous (once per join); bot's channel posts can hit per-guild limits if the sprint expands the command set later — budget headroom.

### Exit criteria
1. A friend clicks an invite link shared in Discord, clicks "Sign in with Discord" once, and lands in the game with their Discord name + avatar already on the seat. No manual name prompt.
2. `/cc-invite` in a Discord channel produces a game + pastes an invite link the playgroup can click.
3. `/cc-invite-dm @alice` DMs Alice a link she can click for one-tap onboarding.
4. With Rich Presence toggled on and the Discord desktop client running, the player's Discord profile shows "In a Commander game" while they play, and clears within seconds of leaving.
5. A player who joined with manual name entry can later click "Link Discord" and have their avatar appear at all four seats without leaving the game.
6. All four of the above work against the deployed VPS from S12.

---

## S13 — Priority foundation (rules graft kickoff)
**Phase:** 7 · **Goal:** model priority + turn-based actions per CR 117 / 502 / 504 / 514. The Untap and Cleanup steps stop granting priority. Untap and Draw fire automatically on step entry (with the Turn-1 skip-draw exception per CR 103.7c). Priority rotation skips eliminated seats. The client adds per-step "stops" preferences and a "Pass to my next stop" button so a 4-player turn doesn't require 30+ manual `pass_priority` clicks per cycle.

This sprint is the *foundation* for the entire S13.x rules-graft track: S13.1 (stack), S13.2 (counter SBAs), S13.3 (greyed illegal actions), S13.4 (interactive cleanup discard), and S13.5 (card visibility) all assume the priority + turn-based-action shape lands here. Cleanup-discard is intentionally deferred to S13.4 — S13's cleanup step just auto-advances to the next turn without granting priority, and S13.4 will inject the discard pause into that gap.

### Tasks

**Priority sentinel + no-priority steps (server):**
- [ ] Define `game.NoPriority = -1` constant in [server/internal/game/turn.go](../server/internal/game/turn.go); document on `Turn.PriorityHolder`.
- [ ] `Turn.advance()` sets `PriorityHolder = NoPriority` when the new step is `StepUntap` or `StepCleanup`; sets it to `ActiveSeat` for every other step.
- [ ] `Game.PassPriority` returns an error (`ErrNoPriority`) if called while `PriorityHolder == NoPriority`. Existing `< 0` defensive guard at [actions.go:138](../server/internal/actions/actions.go) stays as the wire-level catch.
- [ ] `protocol.ViewOfGame` round-trips `PriorityHolder == -1` to clients unchanged.

**Auto turn-based actions on step entry (server):**
- [ ] Extract a private `untapAllForLocked(seat)` from `UntapAll` so the step hook and the existing manual action share code.
- [ ] Extract a private `drawCardLocked(playerID)` from `DrawCard` for the same reason.
- [ ] Extend `runStepEntryHooksLocked` ([server/internal/game/game.go:327](../server/internal/game/game.go)):
  - `StepUntap`: untap all permanents controlled by `ActiveSeat`, then auto-advance the cursor to Upkeep — Untap grants no priority.
  - `StepDraw`: auto-draw 1 card for `ActiveSeat`, **except** when `Turn.Number == 1 && ActiveSeat == Game.StartingSeat` (CR 103.7c).
  - `StepCleanup`: in S13, auto-advance to the next turn's Untap so the cleanup cursor never sits idle. (S13.4 injects the discard pause here.)
- [ ] Manual `TypeUntapAll` / `TypeDrawCard` actions stay dispatchable as sandbox overrides; gate them so they no-op (or return a soft error) during their auto-fire steps to avoid double-fire.

**Eliminated-player skip in priority rotation (server):**
- [ ] `Game.PassPriority`: replace `next = (PriorityHolder + 1) % numSeats` with a loop that walks past `Eliminated == true` seats (reuse the iteration shape from `advancePastEliminatedLocked` at [mutations.go:517](../server/internal/game/mutations.go)).
- [ ] When wrapping back to the active seat through skips, trigger the same step-advance branch as the all-passed case.

**Turn-1 skip-draw rule, CR 103.7c (server):**
- [ ] Add `Game.StartingSeat int` field; set in `Game.Start()` to whichever seat is `ActiveSeat` at that moment.
- [ ] `protocol.GameView` carries `starting_seat` so spectators / reconnects see the same skip-draw decision.
- [ ] Used by the `StepDraw` auto-action above.

**Per-step stops UI (client):**
- [ ] Extract the 12-step list + labels from [Game.svelte](../client/src/routes/Game.svelte) (`STEP_LABELS`) into a shared `client/src/lib/turn.ts` constant so Settings and Game render identically.
- [ ] Replace the empty `gameplay.stepStops: {}` seed in [client/src/lib/settings.ts](../client/src/lib/settings.ts) with sensible defaults (typical MTGO opt-ins: upkeep off, draw off, precombat_main on, declare_attackers on, declare_blockers on, end on, rest off). Bump `__version`; add a migration that fills defaults for users with an empty `stepStops`.
- [ ] Settings.svelte gameplay tab: add a "Step stops" section — checkbox per step. Wires to `updateSettings('gameplay', 'stepStops', {...})`.

**Auto-pass through unstopped steps (client):**
- [ ] Extend the existing `autoPassPriority` effect in Game.svelte: fire `pass_priority` whenever the viewer holds priority, the stack is empty, and `settings.gameplay.stepStops[turn.step] === false`. Reuse the existing dedup-by-snapshot-seq pattern.

**"Pass to my next stop" button (client):**
- [ ] Add `passToNextStop()` next to `passToEnd()` in Game.svelte — same shape (24-iteration cap, await snapshot between sends), with the exit condition: stop when active seat changes, when stack changes, or when `settings.gameplay.stepStops[turn.step] === true`.
- [ ] Render a "→ Next stop" button in the priority controls toolbar between "pass priority" and "pass until end of turn".

**Hide priority indicator when no one holds priority (client):**
- [ ] Priority pills row in Game.svelte: render only when `turn.priority_holder >= 0`; show a muted "—" marker during Untap / Cleanup so the bar doesn't visually pop.
- [ ] PlayerHeader badge: verify `hasPriority` is false when `priority_holder === -1` (it already should be — `-1 === seatIndex` is always false).

**Docs:**
- [ ] `docs/decisions/0006-priority-foundation.md` — ADR. Topics: why `-1` sentinel vs a separate enum, why auto-fire lives in step entry hooks (vs a tick loop), why per-step stops are client-only (no protocol bump), why `Game.StartingSeat` is a separate field rather than inferred, why `TypeUntapAll` / `TypeDrawCard` actions are kept (sandbox / replay).
- [ ] [docs/protocol.md](protocol.md): document `starting_seat` on `GameView` and the `PriorityHolder == -1` sentinel.

**Tests:**
- [ ] Server unit tests next to the existing `TestFullFourPlayerTurnCycle` pattern in [game_test.go](../server/internal/game/game_test.go):
  - Full 4-player turn cycle with all auto-actions firing; cursor lands on the next active seat's Upkeep without manual `untap_all` / `draw_card`.
  - Turn 1 starting-seat draw is skipped; turn 1 next seat draws normally; turn 2 starting-seat draws normally.
  - PassPriority skips eliminated seats (4 seats, eliminate seats 1+3, expect rotation 0 → 2 → 0 → step advance).
  - PassPriority during Untap / Cleanup returns an error (priority sentinel guard).
  - Cleanup auto-advances to next turn's Untap without granting priority.
- [ ] Snapshot/replay round-trip test for `StartingSeat` and `PriorityHolder == -1`.

### Out of scope (explicit handoffs)
- **Stack, cast actions, SBAs, targeting, modes/X, hold-priority, split-second, commander cast tax + zone replacement** — S13.1 [#63](https://github.com/krakenhavoc/cmd_and_ctrl/issues/63).
- **Counter SBAs (planeswalker loyalty 0, battle defense 0, +1/+1/-1/-1 cancel, poison, saga final chapter), player-level counters, marked-damage cleanup** — S13.2 [#79](https://github.com/krakenhavoc/cmd_and_ctrl/issues/79).
- **Greyed illegal-action affordance** — S13.3 [#99](https://github.com/krakenhavoc/cmd_and_ctrl/issues/99).
- **Interactive cleanup discard + per-player MaxHandSize** — S13.4 [#103](https://github.com/krakenhavoc/cmd_and_ctrl/issues/103). S13's auto-advance on cleanup is the placeholder S13.4 will hook into.
- **Card visibility / known-by tracking** — S13.5 [#108](https://github.com/krakenhavoc/cmd_and_ctrl/issues/108).
- **Per-commander damage tracking** — explicit S13 non-goal.

### Risks / gotchas
- **Auto-action / manual-action collision.** `TypeUntapAll` and `TypeDrawCard` remain dispatchable. If the server has already auto-fired on step entry, a client-sent `draw_card` would draw again. Mitigation: gate the manual actions on `Turn.Step != StepUntap / StepDraw` so they no-op during the auto-fire window. Document in the ADR.
- **Step-entry recursion.** Untap auto-untaps then auto-advances the step, which fires the next step's hook. The Cleanup → next-Untap → auto-untap → Upkeep chain runs synchronously inside one mutation. Verify the existing lock semantics in `AdvanceStep` ([game.go:301](../server/internal/game/game.go)) tolerate the chain (likely fine — `runStepEntryHooksLocked` already runs under the same write lock).
- **Stops + auto-pass races.** The client effect that auto-passes on unstopped steps must dedup against the snapshot seq the same way the existing `autoPassPriority` effect does; otherwise a single step entry can fire two `pass_priority` actions.
- **StartingSeat in old replays.** Replays recorded before this sprint won't carry `StartingSeat`. Decode default of `0` matches the only seat games start on today, so replays still work — document the assumption in the ADR and the protocol doc.
- **Stop defaults are opinionated.** Pre-seeding step stops changes behaviour for existing users on first load after the migration. The migration must only fill defaults when the existing `stepStops` is empty; never overwrite user-configured stops.
- **No stack yet ⇒ "stops" are partial UX.** Without S13.1's stack, "stop on upkeep" only lets the player see an empty upkeep. That's the intended hand-off — the affordance lands now, the value compounds when S13.1 ships.

### Exit criteria
1. A 4-player game runs through full turns with **zero manual `untap_all` or `draw_card` clicks**. Cursor walks Untap → Upkeep → Draw → Main → ... → Cleanup → next seat's Untap.
2. Turn 1's starting seat does not draw at its draw step. Turn 2 onwards, every seat draws normally.
3. With seats 1 and 3 eliminated, priority rotation across non-eliminated seats works (`pass_priority` from seat 0 lands on seat 2, not seat 1).
4. The priority pill / badge is hidden during Untap and Cleanup; players cannot send `pass_priority` during those steps (server returns an error).
5. With "stop on declare attackers" off, the client auto-passes through the declare-attackers step. With it on, the client stops and waits for an explicit click.
6. "→ Next stop" button advances to the next configured stop (or the next seat's turn if no further stops are configured), bounded by the same 24-iteration safety cap as `passToEnd`.
7. ADR `0006-priority-foundation.md` exists and explains the sentinel, auto-fire architecture, and the kept-but-gated manual actions.

---

## S13.1 — Stack: cast / resolve / target / counter / trigger / SBA
**Phase:** 7 · **Goal:** all stack-adjacent items in one sprint.

- [x] Type helpers (IsLand, IsInstant, IsSorcery, IsPlaneswalker, IsBattle, IsArtifact, IsEnchantment, IsPermanent)
- [x] `cast_spell` action; lands route to battlefield; spells to stack
- [x] Auto-resolve on full priority pass
- [x] Targeting (announce + re-check on resolve, CR 608.2b)
- [x] Modes / X / distribution capture
- [x] Manual `announce_trigger` + APNAP-ordered queue
- [x] Activated abilities + loyalty (sorcery-speed, once-per-turn)
- [x] `counter_spell` (with destination override) and `counter_ability`
- [x] Hold-priority modifier; split-second flag
- [x] Commander cast tax (per-commander instance ID via Player.CommanderCasts)
- [x] Commander zone replacement (CR 903.9 — `move_card.as_commander` flag)
- [x] State-based actions: lethal damage, 0 toughness, 0 life, 21 commander damage, draw from empty library
- [x] Leaving-game stack cleanup (CR 800.4a — spells exile, abilities + triggers vanish)
- [ ] Heavier announce-time UX: CastDialog (target picker / X / modes / distribution), AbilityDialog, inline mark-damage on creature tiles — follow-up; the wire shape already accepts the data so the basic cast loop and counter affordance are testable today

**Exit criteria:** a Commander player can cast Lightning Bolt, opponent counters with Counterspell on the stack, full priority/SBA loop works end-to-end.

**Bright line — out of scope (S14+):** auto-fire of triggered abilities from card events, auto-validation of target legality at announce, auto-resolution of spell effects, mana pool, replacement effect engine generally, static abilities, combat keyword effects.

---

## S13.2 — Counter mechanics (SBAs + player counters + UI)
**Phase:** 7 · **Goal:** close out the "counters" surface the engine still has to resolve manually. S13.1 ships the four canonical Commander SBAs; this sprint adds the counter-specific SBAs (planeswalker loyalty, battle defense, +1/+1/-1/-1 cancel, poison player-loss, saga final-chapter), first-class UI treatment of counters, and a shared registry of MTG counter types. Sits between S13.1 and S14 so the effect catalog (S14) can rely on counters being fully modelled.

Gap analysis behind this sprint: `Card.Counters` exists today ([server/internal/game/card.go](../server/internal/game/card.go)) and `CurrentPower()` already applies +1/+1 / -1/-1 to P/T. S13.1 covers the Commander SBAs but explicitly not the counter SBAs. S14's `AddCounters` primitive and S17's replacement engine assume counter data is a first-class shape. Nothing in the roadmap (as of issues [#62](https://github.com/krakenhavoc/cmd_and_ctrl/issues/62)–[#70](https://github.com/krakenhavoc/cmd_and_ctrl/issues/70)) covers player-level counters, counter-specific SBAs, or client pip rendering.

### Tasks

**Counter-specific state-based actions (server):**
- [x] CR 704.5i — planeswalker with 0 loyalty counters → owner's graveyard
- [x] CR 704.5p — battle with 0 defense counters → owner's graveyard
- [x] CR 704.5q — `+1/+1` and `-1/-1` cancel 1-for-1; runs BEFORE the lethal-damage / 0-toughness destruction passes (CR 704.3)
- [x] CR 704.5c — player with ≥10 poison counters loses the game
- [ ] CR 704.5u — saga final-chapter sacrifice: deferred to S14+ alongside the effect catalog (needs per-card final-chapter metadata)

**Player-level counters (server):**
- [x] `Player.Counters map[string]int` alongside legacy Poison/Energy ints (kept synchronised)
- [x] `Game.AddPlayerCounter(playerID, name, delta)` mutation with zero-clamp + map-sparse semantics
- [x] Wire action `add_player_counter` (payload: `{ name, delta }` — player-scoped)
- [x] Protocol: `PlayerView.counters` map (sparse / omitempty)
- [x] Clone/restore round-trip carries `Player.Counters`

**Counter-type registry (server + client):**
- [x] `server/internal/game/counter_types.go` — named constants for engine-referenced types + KnownCardCounters / KnownPlayerCounters slice forms. Unknown names round-trip without validation.
- [x] `client/src/lib/counterTypes.ts` — TS mirror with iconography (per-type color + glyph + abbr); `counterStyle()` falls back to neutral for homebrew

**Marked-damage cleanup (CR 514.2):**
- [x] `Card.DamageMarked int` (already shipped in S13.1 — feeds the lethal-damage SBA)
- [x] Cleanup-step turn-based action zeros `DamageMarked` on every creature before the auto-advance
- [x] Lethal-damage SBA reads `DamageMarked >= CurrentToughness` (S13.1)

**Client UI (Svelte):**
- [x] `CounterPips.svelte` — stacked pip overlay at top-right of every battlefield card; colour from the registry
- [x] PlayerHeader counter row — poison/energy keep dedicated chips, generic loop renders other non-zero counters (experience ⭐, rad ☢, homebrew •); +/- routes through `add_player_counter`
- [x] `damage_marked` badge on battlefield cards (red, bottom-right)
- [x] `protocol.ts` — `CardView.damage_marked`, `PlayerView.counters`, `PlayerView.commander_casts`
- [ ] Counter-inventory popover (right-click → "Counters" menu) — deferred; the CounterPips + PlayerHeader counter row are enough to verify counters at the table
- [ ] Animated counter placement — deferred; static rendering is enough for the foundation

**Docs:**
- [x] `docs/decisions/0008-counter-mechanics.md` — ADR
- [x] `docs/protocol.md` — `add_player_counter` action + `counters` / `damage_marked` field documentation

**Tests:**
- [x] SBAs: planeswalker loyalty 0, battle defense 0, +1/+1 -1/-1 cancel (5-case table), poison ≥ 10 loss
- [x] AddPlayerCounter clamps at zero; legacy Energy/Poison ints stay in sync
- [x] Cleanup-step clears `DamageMarked`

### Out of scope (explicit handoffs)
- **Counter-placement replacement effects** (Doubling Season, Hardened Scales, Branching Evolution) — S17 [#67](https://github.com/krakenhavoc/cmd_and_ctrl/issues/67)
- **Counter-generating triggered abilities** (e.g. "when this ETBs, put a +1/+1 counter on target") — auto-fire in S19 [#69](https://github.com/krakenhavoc/cmd_and_ctrl/issues/69); manual `announce_trigger` from S13.1 still works here
- **Infect / wither damage as counter placement** — depends on combat keyword layer (S18 [#68](https://github.com/krakenhavoc/cmd_and_ctrl/issues/68)); poison-counter SBA ships here but the *creature-inflicts-poison* mechanic lives with other combat keywords
- **Persist / undying counter-conditional triggers** — S19 (needs event log + trigger framework)
- **Proliferate** — S14 effect primitive; ships alongside the effect catalog
- **Counter animations beyond simple fade** — S09 polish work; not this sprint

### Risks / gotchas
- `Card.CurrentPower()` already encodes layer-7d math; S16's layer system will eventually want to own this. Don't over-invest in refactoring `CurrentPower()` now — S16 will absorb it cleanly, and duplicating the +1/+1 math elsewhere will just create churn.
- The canonical counter-type list is long and rarely-used types outnumber common types. Ship the list structured but small (top ~20 types get icons; rest are text-only) — over-designing the iconography is a trap.
- `+1/+1 / -1/-1` SBA ordering matters: the cancel-SBA runs *before* the lethal-damage SBA, so a 2/2 with a -1/-1 counter and 1 marked damage shouldn't die if there's also a +1/+1 counter to cancel. Snapshot-test the ordering explicitly.
- Poison counters currently have no home (`Player.Counters` doesn't exist). Adding the field is a protocol change — old client builds will silently drop it on decode. Fine for the friends-only deployment; document the bump in `docs/protocol.md`.

### Exit criteria
1. Cast a planeswalker, activate its minus ability to 0 loyalty → the card moves to its owner's graveyard without manual `move_card`.
2. Infect a player to 10 poison counters via `add_player_counter` → they lose immediately on the next SBA check.
3. Put a +1/+1 and a -1/-1 counter on the same creature → both clear on the next SBA check; a 2/2 with +1/+1 and -1/-1 is still a 2/2 afterward.
4. Cast a battle, let opponents damage it to 0 defense counters → it moves to graveyard.
5. Every counter on every card renders as a visible pip on the battlefield tile, with distinct colors for the top-20 known types.
6. Right-clicking a card opens a Counters popover; players can add/remove any type without typing into chat.
7. Player-level counters (poison, energy, experience, rad) render near each `PlayerHeader` and update live when the server broadcasts a delta.

---

## S13.3 — Client-side timing affordance (greyed illegal actions)
**Phase:** 7 · **Goal:** the UI pre-disables illegal cast / activation affordances so players don't try actions the server will only reject after the click. Pure client work; consumes priority/step/stack/split-second/loyalty-this-turn state that S13 + S13.1 already publish.

After S13 + S13.1, the server enforces timing: who holds priority, sorcery-speed window, split-second blocks, once-per-turn loyalty, etc. Today the only UI signal is an error toast *after* an illegal click. S13.3 closes the loop: hand cards grey out when not castable, ability buttons disable with a reason tooltip, the play-land button greys after the one-per-turn is used. Server enforcement stays authoritative — UI greyness is best-effort UX.

### Tasks

**Client legality helpers (`client/src/lib/timing.ts`):**
- [x] `canCastFromHand(card, snap, viewerID): { legal, reason? }` — checks priority + sorcery-speed window for non-instants + split-second clear; lands gated identically to sorceries (CR 305.3)
- [x] `canActivateAbility(card, snap, viewerID)` — instant-speed activation predicate (priority + split-second clear)
- [x] `canActivateLoyalty(card, snap, viewerID, alreadyActivated?)` — sorcery-speed + battlefield-presence; the once-per-turn flag is best-effort callee-tracked since `LoyaltyActivatedThisTurn` is server-only
- [x] `canPassPriority(snap, viewerID)` — false during Untap / Cleanup (NoPriority sentinel) and when viewer isn't the holder
- [x] Reasons are plain English: "Not your priority", "Stack isn't empty", "Only at sorcery speed", "Split second on the stack", "Lands only on your main phase", "Already activated this turn"

**Hand component:**
- [x] Each card runs `canCastFromHand` in a `$derived`; illegal cards get `.timing-disabled` (opacity 0.55 + grayscale 0.4) and the click handler is detached
- [x] Hover/zoom still works — only the click affordance is muted
- [x] Tooltip on the slot exposes the legality reason

**Toolbar / battlefield / ability dialog:**
- [ ] Toolbar pass-priority button still uses pre-S13.3 `viewerHasPriority` derivation (matches `canPassPriority` for the common case); CastDialog / AbilityDialog UX integration follows when those land alongside S13.1's deferred dialogs
- [ ] Loyalty greyness in the ability dialog ships with the dialog itself (S13.1 deferred UX batch)

**Tests:**
- [x] `client/src/lib/timing.test.ts` — 23 vitest cases covering every reason string + the spectator path
- [ ] Manual smoke: 2-tab playtest (sorceries on opponent turn, instants OK, split-second blocking, etc.) — covered by the 5 exit criteria below

### Out of scope (stays for later)
- **Mana-source prediction** ("can't afford this") — depends on [S15](#s15--mana-pool--auto-tapper-from-cost-vector) mana pool
- **Target-predicate greyness** ("no legal targets") — [S20](#s20--auto-target-legality--smart-cast-ui) territory
- **Alt-cast paths from non-hand zones** ("flashback not yet eligible") — [S29](#s29--alt-cast-paths-from-non-hand-zones)
- **Animated un-grey transitions** — S11-ish polish

### Exit criteria
1. On opponent's turn, viewer's hand sorceries are visibly greyed; instants render normally; tooltip on a greyed sorcery says "Not your turn".
2. Cast a sorcery on your own main phase → other sorceries in your hand grey out with "Stack isn't empty"; un-grey when the stack resolves.
3. Activate a planeswalker's loyalty → that planeswalker's ability list greys remaining options with "Already activated this turn"; resets next turn.
4. Play a land → second land in hand greys with "Already played a land this turn".
5. Force a race condition (skip the client check) → server still rejects with the same error toast that worked pre-S13.3 (server enforcement remains authoritative).

Detailed plan: `/home/node/.claude/plans/s13-3-client-timing-affordance.md`. Builds on S13 (priority sentinel) + S13.1 (sorcery-speed + split-second + loyalty-activated state). ~1 week, 2 sub-PRs.

---

## S13.4 — Interactive cleanup discard + per-player MaxHandSize
**Phase:** 7 · **Goal:** close out S13's deferred interactive-discard item + lay the per-player `MaxHandSize` groundwork for hand-size-modifying effects (Reliquary Tower, Thought Vessel, Spellbook, Library of Leng, Null Profusion, Venser's Journal) to plug into later via S14/S16.

S13's cleanup step auto-discards from the hand top as a placeholder — S13.4 replaces that with an interactive prompt: the cleanup step pauses, the active player picks which cards to discard, then the turn advances. Separately, the hardcoded "7" becomes a per-player `MaxHandSize int` (default 7, `-1` = no maximum) so later catalog cards can modify it. The catalog declarations themselves (Reliquary Tower et al.) ship via S14's effect catalog and S16's layer system writing to the field this sprint creates.

### Tasks

**Server — engine + actions:**
- [x] `Player.MaxHandSize int` field (default 7; `-1` = no max)
- [x] `Game.DiscardPending map[playerID]int` populated by the cleanup-step-entry hook
- [x] Cleanup hook auto-advances only when `DiscardPending` is empty; re-fires on `discard_selection` drain so the cursor resumes
- [x] `discard_selection` action (`{ card_ids }`) — count-validated, hand-ownership-validated, idempotent for callers not in the pending map
- [x] `set_max_hand_size` action (`{ value }`, player-scoped) — sandbox helper for now; effect-catalog work in S14+ writes to it via the layer pipeline
- [ ] APNAP order for multi-player simultaneous discard — single-prompt-at-a-time is fine for the friends-only deployment; revisit if/when Mindslicer / Painful Quandary land at the table

**Server — wire:**
- [x] `PlayerView.max_hand_size int` (always present so clients know the cap)
- [x] `GameView.discard_pending map<playerID, count>` (omitempty)

**Client:**
- [x] `DiscardPromptModal.svelte` — opens when `discard_pending[viewerID] > 0`; multi-select exactly N cards; non-dismissible; auto-closes when the wire drains the entry
- [ ] `PlayerHeader` shows non-default `MaxHandSize` next to hand-count — deferred until set_max_hand_size has a UI consumer
- [ ] Settings panel toggle for `set_max_hand_size` — deferred, sandbox helper only

**Tests:**
- [x] Server: over-max prompts, selection resolves + advances, wrong count rejected, NoMaxHandSize bypass, set_max_hand_size validation
- [ ] Manual 2-tab smoke through the 5 exit scenarios below

### Out of scope
- **Static-ability declarations** for Reliquary Tower, Thought Vessel, Spellbook, Library of Leng, Null Profusion, Venser's Journal — these land as [S14](#s14--card-effect-catalog-foundation) catalog cards plugged into [S16](#s16--continuous-effects--layer-system-cr-613) layer pipeline, writing to the `MaxHandSize` field S13.4 creates
- Replacement effects on discard (Library of Leng's "to top/bottom of library instead") — [S17](#s17--replacement-effects-engine-cr-614)
- Triggers on discard (Asylum Visitor, madness) — [S19](#s19--auto-fire-triggered-abilities) + [S29](#s29--alt-cast-paths-from-non-hand-zones)
- Hand-reveal UX (Telepathy, Bottled Cloister)
- Multi-player simultaneous discard ordering UI (server handles APNAP correctly; only one modal at a time)

### Exit criteria
1. End turn with 9 cards → modal opens on active-player tab prompting "discard 2 cards"; cannot close without selecting.
2. Select 2 + submit → cards move to graveyard, modal closes, turn advances.
3. Set `MaxHandSize = -1` → end turn with 15 cards, no modal, cleanup passes through.
4. Set `MaxHandSize = 2` → end turn with 5 cards, modal prompts "discard 3 cards".
5. Opponent tab doesn't see the prompt; hand contents stay private.

Detailed plan: `/home/node/.claude/plans/s13-4-hand-size-and-interactive-discard.md`. Builds on S13 (cleanup turn-based action) + S13.1 (pause-and-resume prompt pattern). ~1 week, 2 sub-PRs.

---

## S13.5 — Card visibility + known-by tracking
**Phase:** 7 · **Goal:** Arena-style per-instance sticky visibility. Every `Card` carries a `KnownBy` set — who currently knows this specific instance's identity. Zone moves / reveals / shuffles mutate the set; hub filter redacts printed characteristics for non-`KnownBy` viewers; client renders card back with hover-reveal for viewers who previously saw the card face-up.

Today's filter is zone-default-only — no way to express Thoughtseize reveals, bounced-but-known creatures, scry persistence, or morph-face-up-then-down with correct opponent knowledge. S13.5 replaces the zone-default heuristic with authoritative per-card `KnownBy`. [S22](#s22--card-draw--library-manipulation)'s transient reveal frames become animation sugar on top.

### Tasks

**Server — card + mutations:**
- [x] `Card.KnownBy map[uuid.UUID]bool` field; `Card.FaceDown bool` field
- [x] Helpers: `AddKnower`, `AddKnowersAll`, `ClearKnown`, `IsKnownTo`
- [x] `Game.Start` initialiser — library no knowers, starting hand → owner, command zone → all seated
- [x] `Game.markCardKnownInZoneLocked` post-move hook called from PlayCard / CastSpell / resolution / counter / battlefield-to-graveyard SBA paths; public zones grant all, hand grants owner only
- [x] `ShuffleLibrary` clears KnownBy on every library card (CR 701.20)
- [x] `Mulligan` clears hand + library knowledge, re-grants owner on the new opening hand
- [ ] Token / Clone ETB / spell-copy creation initializes `KnownBy = {all seated}` — deferred until those paths exist (S14+ effect catalog)

**Server — wire:**
- [x] `CardView.face_down bool` + `CardView.known_by_you bool` (per-viewer)
- [x] Printed characteristics (name, type_line, scryfall_id, power, toughness, counters, is_commander) zeroed by FilterViewFor for non-knowers; instance_id, owner, controller, tapped, position, face_down, damage_marked always sent

**Hub filter:**
- [x] `FilterViewFor` walks every visible zone and per-card decides redaction via the unexported `knowers` map carried from `viewOfCard`. Opponent hand + library still get the S04 zone-wholesale hide; visible-zone cards get per-card redaction.
- [x] `known_by_you` populated per viewer

**Client:**
- [x] `Card.svelte` renders the back when `face_down` or `known_by_you === false`; face-up otherwise
- [x] `protocol.ts` mirrors `face_down` + `known_by_you`
- [ ] Hover-reveal on face-down + known_by_you cards — deferred; the existing hover-zoom already shows the face for known cards via the imgSrc path
- [ ] Mixed hand rendering (face-up for revealed-by-Thoughtseize cards) — deferred; depends on the opponent-hand-per-card filter, which is also deferred

**Tests:**
- [x] Server: Start initialises KnownBy correctly per zone; shuffle clears library; public-zone move grants all seats; draw adds owner only
- [x] Hub filter: redaction on non-knower viewer keeps instance_id but zeroes printed chars; known cards keep characteristics + known_by_you=true
- [ ] Client: face-down + known_by_you renders back with hover — manual smoke covered by exit criteria

### Out of scope
- Long-term "opponent-X-saw-card-Y in the past" advisory UI — pure engine tracking is enough
- Undo / take-back logic — falls out naturally from instance-sticky `KnownBy`
- Rules-lawyering on face-down concealment (paper says no-characteristics; we match Arena's sticky model and document the deviation)
- Replay "what did I know when" historical tracking

### Exit criteria
1. **Scry persistence.** Scry 1, keep on top, next turn draw → card stays visible to you all along.
2. **Bounce-known.** Cast Lightning Bolt (public on stack), opponent Unsummons it → renders face-up for everyone in your hand.
3. **Thoughtseize-style reveal.** Opponent reveals your hand → cards stay known to opponent after the reveal; next turn's drawn cards unknown again.
4. **Morph face-up-then-down.** Morph cast face-down, flipped face-up by paying morph cost, then Ixidroned back face-down → opponents still hover-reveal the true card.
5. **Shuffle clears.** Scry 3, see all, then shuffle → top-of-library knowledge cleared.
6. **Commander path.** Public from game start through cast / battlefield / graveyard / command-zone replacement. All steps all-known. Shuffle into library (rare effect) clears.
7. **Token.** Creates with `KnownBy = {all}` immediately.

Detailed plan: `/home/node/.claude/plans/s13-5-card-visibility-knownby.md`. Builds on S13 (turn/step primitives) + S22 (transient reveal frames become animation sugar). ~1.5 weeks, 3 sub-PRs.

---

## S14 — Card-effect catalog foundation
**Phase:** 7 · **Goal:** ~30 of the most-played Commander cards resolve end-to-end with zero manual intervention — cast Lightning Bolt at an opponent, the stack resolves on a full priority pass, opponent's life drops by 3, the card routes to graveyard, no `change_life` click required. Unimplemented cards keep today's Cockatrice-style manual posture; the catalog is opt-in per Scryfall ID.

S14 lays the rules-engine infrastructure the rest of Phase 7 hangs off: a per-game append-only event log, a pull-based listener registry (pre-wired for S19 triggered abilities), a declarative Forge-style card-effect catalog, 15 composable effect primitives, and 30 starter cards picked to exercise the primitives without depending on mana (S15), layers (S16), replacement effects (S17), or combat keywords (S18).

### Architectural decisions

- **Forge-style declarative DSL in Go structs.** No oracle-text parser. Each card file is ~5-15 lines: an `init()` that calls `effects.Register(Spec{...})`.
- **Cockatrice-style fallback for unimplemented cards.** Non-catalog cards keep 100% of today's manual sandbox posture — the auto-resolve path is opt-in per Scryfall ID. Regression-pinned.
- **Pull-based event dispatch.** Listeners query the log on demand; the engine walks listeners synchronously after every `EmitEvent` under the write lock. S14 registers zero production listeners — the registry is infrastructure for S19.
- **Scryfall data is display-only.** `server/internal/cards` resolves names / images / type lines / mana costs; it does NOT parse Oracle text into effects.
- **Catalog membership is opt-in.** `effects.Lookup(scryfallID).OnResolve != nil` is the auto-trigger bit; absent ⇒ manual.
- **Effects run under the existing resolution write lock.** Primitives call `*Locked` helpers only — never public locking mutators — to avoid deadlock.
- **One file per card** under `server/internal/cards/effects/`. Low merge-conflict surface, clean `git blame`.
- **`Spec.OnETB` is a direct-call shim in S14.** S19 re-routes it through the listener registry with no per-card changes.

### Tasks

**Event log infrastructure (server):**
- [x] New `server/internal/game/events.go` — `Event` tagged struct (`{Kind, Actor, Source, Target, Amount, CardID, OldZone, NewZone, Seq}`), `EventKind` constants (Cast, Resolve, Fizzle, DealDamage, ChangeLife, DrawCard, DiscardCard, Mill, ZoneMove, TapCard, UntapCard, CounterPlaced, TokenCreated, SearchLibrary, CounterSpell, Concede, ETB, LTB, EffectError).
- [x] `Game.Events []Event` field; `Game.EmitEvent(Event)` append point (caller holds `g.mu`).
- [x] `server/internal/game/clone.go` — deep-copy `Events` slice; shallow-copy `Listeners`.
- [x] Instrument every rules-visible mutation to call `EmitEvent`: `drawCardLocked`, `routeBattlefieldCardToOwnerGraveyardLocked`, `routeStackCardToGraveyardLocked`, `MarkDamage`, `ChangePlayerLife`, `AddCounter`, `TapCard`, `ShuffleLibrary`, `MoveCardByID`, `CastSpell`, `resolveTopOfStackLocked`, `CounterSpell`, `CounterAbility`, `eliminatePlayerLocked`, `AnnounceTrigger`.
- [x] `protocol.GameView.events` (last-N window, omitempty) for client debugging / auto annotations.

**Listener registry (server):**
- [x] New `server/internal/game/listeners.go` — `Listener interface { OnEvent(*Game, Event) }`, `Game.RegisterListener(Listener)`, `Game.notifyListenersLocked(Event)` fired immediately after `EmitEvent`.
- [x] Zero production listeners in S14. `noOpListener` in tests exercises the wiring.

**Effect engine (server):**
- [x] New subpackage `server/internal/cards/effects/` (separate from `cards` — keeps Scryfall-index concerns disentangled from effect concerns).
- [x] `effects/registry.go` — `Spec` struct, `Register(Spec)`, `Lookup(scryfallID) (Spec, bool)`, `All() []Spec`. Duplicate-ScryfallID registration panics at package init.
- [x] `effects/context.go` — `Context` wrapping `*game.Game` under-lock, helper accessors (`CreatureIDs`, `PlayerByID`, `ZoneOf`, `IsTargetLegal`).
- [x] `effects/primitives.go` — 15 primitives, each a struct with `Apply(ctx *Context) error` that emits the matching `Event`.
- [x] `effects/spec.go` — `Spec{ScryfallID, Name, OnResolve, OnETB, StartingLoyalty}`.
- [x] `effects/tokens.go` — `TokenSpec` + `CreateToken`. Fresh `InstanceID`, empty `ScryfallID`, `KnownBy = {all seated}`.
- [x] 30 card files under `effects/` (one per card).

**Resolution integration (server):**
- [x] `resolveTopOfStackLocked` (mutations.go) — after target-legality short-circuit, before zone routing: `if spec, ok := effects.Lookup(top.ScryfallID); ok && spec.OnResolve != nil { spec.OnResolve(item, ctx) }`. Errors emit `EventEffectError`, don't wedge resolution.
- [x] ETB hook: after `MoveCard(Stack → Battlefield, …)`, call `spec.OnETB(&card, ctx)` if non-nil.
- [x] Starting loyalty: `StartingLoyalty > 0` on the spec stamps `CounterLoyalty` counters in the ETB branch.

**Wire protocol (server + client):**
- [x] `protocol.CardView.auto bool` — set by `viewOfCard` when `effects.Lookup(c.ScryfallID).OnResolve != nil || .OnETB != nil`.
- [x] `client/src/lib/protocol.ts` — mirror `CardView.auto` + `GameView.events`.
- [x] `docs/protocol.md` — document new fields + `EventKind` constants.

**Client (Svelte):**
- [x] `client/src/lib/components/board/Card.svelte` — gold-leaf "auto" badge bottom-right when `card.auto === true`. Hover tooltip.
- [x] `client/src/lib/components/board/StackOverlay.svelte` — "auto-resolve" chip next to the controller name on catalog stack items.
- [x] `client/src/lib/events.ts` — derived store pulling `GameView.events`, surfaces last ~5 as 2s toast notifications ("Alice took 3 from Lightning Bolt"). Soft nudge against double-applying after an auto-resolve.

**Tests:**
- [x] `server/internal/game/events_test.go` — emit points + clone/restore round-trip + listener notification order.
- [x] `server/internal/cards/effects/primitives_test.go` — each primitive in isolation.
- [x] `server/internal/cards/effects/cards_test.go` — table-driven, one case per catalog card.
- [x] `mutations_test.go` extensions — `TestCastLightningBoltAutoResolves`, `TestCastNonCatalogSpellStaysSandbox` (opt-in regression canary), `TestCastSpellFizzleRoutesToGraveyard`.
- [x] Vitest — `Card.svelte` auto badge, `StackOverlay.svelte` auto chip.

**Docs:**
- [x] `docs/decisions/0010-card-effect-catalog.md` — ADR covering the decisions above.
- [x] `docs/protocol.md` — wire docs for new fields.
- [x] `AGENTS.md` — "how to add a new catalog card" recipe.

### Starter card list (31 cards)

| Family | Cards | Primitives |
|---|---|---|
| Direct damage | Lightning Bolt, Shock, Lightning Helix, Pyroclasm | DealDamage + GainLife iteration |
| Mass removal | Wrath of God, Damnation, Day of Judgment | DestroyTarget iteration |
| Counter magic | Counterspell, Negate, Swan Song | CounterTarget + CreateToken |
| Draw | Divination, Harmonize, Sign in Blood | DrawCards + ChangePlayerLife |
| Mill / discard | Glimpse the Unthinkable (mill 10), Thoughtseize (random — UI deferred to S22), Mind Rot | MillCards + DiscardCards |
| Targeted removal | Swords to Plowshares, Path to Exile (enters-tapped deferred) | ExileTarget + GainLife + SearchLibrary |
| Bounce | Unsummon | BounceToHand |
| Mana rocks (vanilla permanents this sprint) | Sol Ring, Arcane Signet | none — S15 wires mana abilities |
| Tutors | Cultivate (both lands → hand — enters-tapped deferred to S17), Demonic Tutor, Vampiric Tutor | SearchLibrary (to hand / library-top) |
| Recursion | Eternal Witness (`OnETB` direct-call hook → S19 migrates to listener), Regrowth | ReturnFromGraveyard |
| ETB creatures | Solemn Simulacrum (enters-tapped land deferred), Acidic Slime | `OnETB` composition |
| Vanilla creature placeholder | Birds of Paradise | none — mana ability lands in S15 |
| **Planeswalker** | **The Wandering Emperor** (`OnETB` stamps starting loyalty 3 via `AddCounter`; activated abilities remain manual via S13.1's `activate_loyalty`) | **AddCounter on ETB** |

**Primitive coverage summary:** 13 of 15 primitives exercised by catalog cards; TapTarget / UntapTarget ship as primitives with unit-test-only coverage so S15 has the hooks ready.

### Out of scope (explicit handoffs)
- **Mana pool + cost validation** — S15. Sol Ring / Arcane Signet / Birds of Paradise ship as vanilla permanents; their mana abilities plug into `effects/` as activated-ability specs in S15.
- **Continuous effects + layers** — S16. No S14 catalog card has a static ability.
- **Replacement effects (ETB-tapped, "if-would-die-exile-instead")** — S17. Cultivate's "enters tapped" half, Path to Exile's tapped land, Solemn Simulacrum's tapped land all defer.
- **Combat keywords (trample, lifelink, flying)** — S18.
- **Auto-fire triggered abilities from catalog cards** — S19. ETB hooks use the direct-call path in S14 and migrate for free.
- **Scry / surveil / look-at-top-N + Brainstorm's "put 2 back" half** — S22 library-manipulation polish. Brainstorm is NOT in the S14 catalog.
- **Target-picker UX for Thoughtseize (pick-from-revealed-hand)** — S22. S14 ships random-pick sandbox.
- **Target legality validation at announce time** — S20 smart-cast UI. Effects re-check at resolve via CR 608.2b.

### Risks / gotchas
- **Sandbox-override double-apply.** Player clicks "-3" manually after Lightning Bolt auto-resolves. Mitigation: recent life-change toast + gold-leaf "auto" badge as visual nudges. No server-side de-dup.
- **Non-catalog-card regression.** Opt-in invariant is load-bearing; every non-catalog test must assert manual-only behaviour is unchanged. Canary: `TestCastNonCatalogSpellStaysSandbox`.
- **Effect resolution under write lock.** Primitive calling a public locking mutator ⇒ deadlock. Enforce via godoc + code review (Go can't encode "locked context" at the type level).
- **Target cardinality mismatch.** Effect with missing target silently no-ops + emits `EventEffectError`. Debug via `GameView.events`. Announce-time validation is S20 territory.
- **Undo through resolution.** `pass_priority` that triggers an auto-resolve is one `Apply` frame on the undo stack — undoing rewinds everything. Matches S13.1 resolution-undo behaviour.
- **Pre-S14 replay snapshots.** No `events` field ⇒ decodes to empty log ⇒ correct lossy restore.
- **Event log size.** ~2-5k events per game × 32 undo frames × ~64 B each ≈ 10 MB per room. Acceptable for friends-only deployment.
- **ETB hook migration to S19.** Direct-call path in S14 becomes listener-sugar in S19 — catalog cards untouched. Documented in ADR.

### Exit criteria
1. **Lightning Bolt** — cast at opponent, pass priority round, opponent's life drops by 3, card in caster's graveyard. No manual clicks.
2. **Counterspell** — opponent's Bolt on the stack, cast Counterspell, pass priority: Bolt in opponent's graveyard, no damage, Counterspell in caster's graveyard.
3. **Cultivate** — cast, resolve: library shuffled, two basics in caster's hand, Cultivate in graveyard.
4. **Wrath of God** — with 4 creatures on the board, cast Wrath, resolve: all 4 in owners' graveyards, Wrath in caster's graveyard.
5. **Eternal Witness** — cast, resolves to battlefield, caster clicks ETB hook, picks a graveyard card, card returns to hand.
6. **The Wandering Emperor** — cast, enters battlefield with 3 loyalty counters automatically. Loyalty abilities remain manual in S14.
7. **Non-catalog card regression** — cast a card not in the catalog, passes priority, routes per S13.1, no auto-effect.
8. **Auto badge** — catalog cards in hand show the gold-leaf badge; non-catalog cards don't.
9. **Event log inspection** — `GameView.events` includes Cast → Resolve → DealDamage → ZoneMove for the Bolt case.
10. **Undo** — cast Bolt, pass priority (auto-resolves), undo: life restored, Bolt back on stack.

### Sub-PR split
Stop-and-show for manual testing at each.
1. Event log + listener registry infrastructure. Zero visible changes; regression is "nothing breaks."
2. `effects/` package skeleton + 15 primitives + unit tests. Still no visible change.
3. Resolution integration + `auto` bit on CardView + auto badge in `Card.svelte` + StackOverlay chip. First visible change: catalog stubs (empty-body Lightning Bolt) render with the badge.
4. Catalog cards in 2-3 sub-PRs grouped by family (damage + counters, draw + mill + discard + removal, bounce + tutors + recursion + ETB creatures + Wandering Emperor).
5. Event-toast client polish.
6. ADR + protocol docs final pass.

### Critical files
- Server: [events.go](server/internal/game/events.go) (new), [listeners.go](server/internal/game/listeners.go) (new), [game.go](server/internal/game/game.go), [clone.go](server/internal/game/clone.go), [mutations.go](server/internal/game/mutations.go), [view.go](server/internal/protocol/view.go), new `server/internal/cards/effects/` subpackage (registry, context, primitives, spec, tokens, 30 card files).
- Client: [Card.svelte](client/src/lib/components/board/Card.svelte), [StackOverlay.svelte](client/src/lib/components/board/StackOverlay.svelte), [protocol.ts](client/src/lib/protocol.ts), new [client/src/lib/events.ts](client/src/lib/events.ts).
- Docs: new [docs/decisions/0010-card-effect-catalog.md](docs/decisions/0010-card-effect-catalog.md), [docs/protocol.md](docs/protocol.md), [AGENTS.md](AGENTS.md).

### Branch + commit conventions
- Branch: `feat/s14-card-effect-catalog` (sub-PRs use `-s14-<slice>` suffixes).
- Footer:
  ```
  Sprint: S14 — Card-effect catalog foundation
  Issue: #64
  ```

---

## S15 — Mana pool, cost model, and auto-tapper
**Phase:** 7 · **Goal:** bolt a real mana economy onto the sandbox — a server-side `ManaPool` per player, a cost parser that understands `{1}{R}`, `{W/U}`, `{X}`, `{W/P}`, `{S}`, and `{C}`, cost validation in `cast_spell` that defaults to warn-and-proceed but gates strictly under an opt-in setting, and a backtracking auto-tapper that returns a tap plan in microseconds for any real Commander manabase. The Sol Ring / Arcane Signet / Birds of Paradise vanilla catalog entries seeded in S14 light up this sprint with their activated mana abilities, tapping through the catalog on an `activate_mana_ability` action that also covers the synthetic "basic land → one mana" path.

Mana abilities do NOT use the stack (CR 605.3); they resolve synchronously on activation and emit the same Event log the rest of the engine reads. This is distinct from S19's stack-aware activated-ability pipeline. Permissive-by-default keeps the Cockatrice feel for anyone who just wants to deal with mana on paper; strict opt-in gives the Arena-style "can't cast that yet" gate to players who want the engine to enforce it.

### Architectural decisions

- **Greenfield, no migration debt.** No mana code exists today (confirmed: zero hits on `ManaPool` / `ManaCost` / `ProducedMana` / `strict_mana` outside this sprint's stub). Every field, action, and wire entry lands fresh.
- **Permissive by default, strict opt-in.** `gameplay.strictMana` ships as a client setting (schema v4 migration in [client/src/lib/settings.ts](../client/src/lib/settings.ts), same shape as the S13→S13-hotfix `autoPassPriority` / `stepStops` seeding pattern). Default `false`: an insufficient-mana `cast_spell` emits an `EventCostWarning` and proceeds. `true`: the cast is rejected with a structured error.
- **Commander tax is enforced only under strict mode.** The `CommanderCasts` counter S13.1 already increments stays the informational affordance everyone sees; the `+{2}` per prior cast is folded into the parsed cost only when strict mode gates the call.
- **Mana abilities live on `Spec.ManaAbilities`.** New slice: each entry is `{Cost ManaAbilityCost, Produced string, Label string}`. Sol Ring / Arcane Signet / Birds of Paradise grow `init()` blocks declaring theirs. Basic lands get a synthetic ability derived from their printed land subtype (Forest → `"{G}"`, Plains → `"{W}"`, etc.) so the dispatcher has one code path for every mana source.
- **Produced-mana syntax uses pipe for pick-one.** `"{W|U|B|R|G}"` means "one mana of any of these colors, controller picks." Birds of Paradise emits that full set. Arcane Signet emits the subset restricted to the controller's commander's color identity, resolved server-side when the ability fires (commander-identity lookup lives on `Player` / `Card.IsCommander` already). `{C}{C}` (Sol Ring) has no pipe; it just drops two colorless tokens in.
- **ANY-color picks reuse the S14 `PendingChoice` queue.** A mana ability whose production string contains a pipe queues a `PendingChoiceMana` entry (new kind), the controller picks via `resolve_choice`, and the picked color lands in the pool. Tap fires immediately; the pool update is deferred only when a pick is needed. Matches the Thoughtseize / Mind Rot precedent for server-driven choice dispatch.
- **One action, `activate_mana_ability`.** The stub's proposed `tap_for_mana` is subsumed. The action carries `{card_id, ability_index}` (0 for basic lands, the index into `Spec.ManaAbilities` for rocks). Basic lands get a synthetic default ability lazily generated from `Card.TypeLine`. Keeps the dispatcher flat and matches S14's "one path for every catalog card" shape.
- **Pool is a slice of tokens preserving insertion order.** `ManaPool []ManaToken` where `ManaToken = {Color string; Source uuid.UUID; Restrictions []string}`. Slice over multiset so the S15 auto-tapper lock-in preview ("you already tapped this Forest") renders the same order in sub-PR 5, and so pain-land / any-color insertion order is observable for future filter-land work (S17+).
- **Pool empties on step change (CR 106.4).** Hook into the existing step-advance path in `turn.go`; emit `EventManaPoolEmptied` per token cleared. Undo through a step-change rolls the pool back like it does for every other rules-visible mutation.
- **Auto-tapper is iterative-deepening backtracking with restriction-first heuristic.** Public API: `Game.AutoTapForCost(controller uuid.UUID, cost ParsedCost) (plan []uuid.UUID, ok bool)`. Restricted sources (Birds of Paradise pick, pain-source) tried before unrestricted (basics) so unrestricted mana is preserved for the next cast. Detailed algorithm in ADR 0011.
- **Cost parser is regexp-free, token-by-token.** `{1}{R}{W}{W/U}{X}{C}` walks char-by-char; produces `ParsedCost{Generic int, Required []ColorRequirement, XSlots int, HasPhyrexian bool, HasSnow bool}`. ~80 LoC, single-pass. Mirrors the S14 "declarative in Go structs, no parser library" posture.

### Tasks

**Card data + Scryfall ingestion (server):**
- [x] `cards.Card.ManaCost string` — raw Scryfall `mana_cost` string ("{1}{R}"); empty for lands.
- [x] `cards.Card.ProducedMana []string` — Scryfall's `produced_mana` array (WUBRGC entries; empty for non-producers). Tracks which sources the auto-tapper is allowed to consider.
- [x] `cards.index.go` JSON tags: `mana_cost` and `produced_mana` added to the struct; `Load()` streams them into the in-memory index.
- [x] `game.Card.ManaCost string` + `game.Card.ProducedMana []string` — stamped alongside the existing `OracleID` copy at `deck.toGameCard`; sandbox placeholder cards keep zero values.
- [x] `effects.ParseCost(string) (ParsedCost, error)` in a new `server/internal/cards/effects/cost.go`. Handles `{N}` generic, `{W|U|B|R|G|C}` colored, `{W/U}` hybrid (Guildmages), `{W/P}` phyrexian (parsed but self-pay life cost out-of-scope — see below), `{X}` (records `XSlots`, caller supplies `XValue`), `{S}` (snow — parsed but no snow-source routing this sprint). Table-driven test over every cost string in a 100-card sample.
- [x] `effects.ParseProducedMana(string) []ManaToken` helper for the synthetic basic-land ability (Forest → `[{Color:"G"}]`).

**`ManaPool` + mana abilities (server):**
- [x] `game.ManaToken struct { Color string; Source uuid.UUID; Restrictions []string }` in a new [server/internal/game/mana.go](../server/internal/game/mana.go).
- [x] `Player.ManaPool []ManaToken` field with `AddMana`, `SpendMana(ParsedCost) (ok bool)`, `EmptyPool()` helpers. `SpendMana` is dry-run-safe (returns a boolean + the post-spend slice; caller commits).
- [x] Step-change hook: `g.emptyAllManaPoolsLocked()` called from the step-advance boundary; one `EventManaPoolEmptied` per non-empty pool.
- [x] `Spec.ManaAbilities []ManaAbility` new field on [effects.Spec](../server/internal/cards/effects/spec.go). `ManaAbility{Cost ManaAbilityCost, Produced string, Label string}`; `ManaAbilityCost` is a tiny struct expressing `{Tap bool; Sacrifice bool}` (sacrifice reserved for future mana rocks — not exercised this sprint; documented).
- [x] `Game.ActivateManaAbility(playerID, cardID uuid.UUID, abilityIdx int) error` in [mutations.go](../server/internal/game/mutations.go). Dispatches through the catalog; falls back to the synthetic basic-land ability when `effects.Lookup(oracleID)` misses but the card's TypeLine contains a basic-land subtype. Emits `EventManaAbilityActivated` + `EventManaAdded` + (if tap cost) `EventTapCard`.
- [x] Pipe-syntax production queues a `PendingChoiceMana` for the controller: options list is the pipe set intersected with the controller's identity (Arcane Signet) or unrestricted (Birds of Paradise). `resolve_choice` drains the pick and adds the resulting `ManaToken` to the pool.
- [x] [sol_ring.go](../server/internal/cards/effects/sol_ring.go), [arcane_signet.go](../server/internal/cards/effects/arcane_signet.go), [birds_of_paradise.go](../server/internal/cards/effects/birds_of_paradise.go) grow their `Spec.ManaAbilities` declarations. Sol Ring: `{Tap}` → `"{C}{C}"`. Arcane Signet: `{Tap}` → `"{W|U|B|R|G}"` (server narrows to identity). Birds: `{Tap}` → `"{W|U|B|R|G}"`.

**Cost validation in `cast_spell` (server):**
- [x] New `CastSpellParams.ForceCast bool` — strict-mode override hatch ("Cast anyway"). Default false.
- [x] `CastSpellParams.AutoTap bool` — when true, the server runs `AutoTapForCost` before the cost-check, taps the returned plan, and drops the produced mana into the pool. Per-source preview (sub-PR 5) is client-driven; the server path is a single atomic frame under `g.mu`.
- [x] `Game.effectiveCostLocked(p *Player, card Card, params CastSpellParams) ParsedCost` — parses `card.ManaCost`, adds `{2}` per prior commander cast if `FromZone == "command"`.
- [x] Strict gate: if `gameplay.strictMana` is set on the calling seat's session AND `!params.ForceCast` AND `!p.ManaPool.CanPay(cost)` → return `ErrInsufficientMana` carrying the missing-symbols slice. Permissive gate: emit `EventCostWarning`, proceed.
- [x] On successful cast, deduct from `Player.ManaPool` (strict mode only; permissive mode leaves the pool alone so the affordance matches paper tracking).

**Auto-tapper (server):**
- [x] `Game.AutoTapForCost(controller uuid.UUID, cost ParsedCost) (plan []uuid.UUID, ok bool)` in new `server/internal/game/autotap.go`. Iterative-deepening backtracking over controller's untapped permanents with `ProducedMana`-non-empty. Restriction-first heuristic: permanents whose produced set is a strict subset of another's are tried first, so generic basics survive for the next cast.
- [x] Budget cap: 10k node expansions (p99 benchmark target < 1 ms for a 38-land Bant manabase). Above budget → `(nil, false)` and the client falls back to manual tapping.
- [x] Pre-tapped / lock-tap integration: `AutoTapForCostExcluding(controller, cost, excluded set)` where the lock-in sources are removed from the search set but their contribution is pre-credited to the cost.
- [x] Unit tests: Cyclonic Rift overload `{1}{U}{U}{U}{U}{U}{U}` from Hallowed Fountain + Breeding Pool + 4 Islands → plan of 6 sources covers 6 blue, 1 generic; Lightning Bolt `{R}` from 3 basics + 1 Signet → picks the Signet first (pipe restriction), cheapest generic survives.

**Wire protocol (server + client):**
- [x] `PlayerView.mana_pool []string` — each entry a serialised token ("W", "U", "C"). Slice, not map. `omitempty` when empty.
- [x] `CardView.mana_cost string` — raw Scryfall cost string for hand-zone cost-chip rendering.
- [x] `CardView.mana_abilities []ManaAbilityView` — `{Index, Label, Cost, Produced}` for activation buttons on the battlefield.
- [x] New `PendingChoiceKind = "mana_pick"` in [pending_choice.go](../server/internal/game/pending_choice.go) + the wire mirror in `PendingChoiceView`.
- [x] New `EventKind` constants: `EventManaAdded`, `EventManaSpent`, `EventManaAbilityActivated`, `EventManaPoolEmptied`, `EventCostWarning`.
- [x] `protocol.ErrorFrame` (new or extend existing) carrying `{code: "insufficient_mana", message, missing: ["{R}", "{R}"]}` — structured so the client can render the strict-mode override affordance without string-matching.

**Client:**
- [x] [client/src/lib/protocol.ts](../client/src/lib/protocol.ts) — mirror `PlayerView.mana_pool`, `CardView.mana_cost`, `CardView.mana_abilities`, the new event kinds, and the structured error.
- [x] [client/src/lib/settings.ts](../client/src/lib/settings.ts) — `gameplay.strictMana: boolean` field (default false), schema bump to v4, migration preserves prior v3 state and seeds strictMana=false.
- [x] Settings panel toggle ("Strict mana enforcement") under the existing Gameplay group.
- [x] `client/src/lib/components/board/ManaPoolPips.svelte` — horizontal WUBRGC pip row rendered inside [PlayerHeader.svelte](../client/src/lib/components/board/PlayerHeader.svelte) next to the life total. Zero-count colors hidden; any-pending choice shows a pulsing "?" pip.
- [x] `Card.svelte` — mana-cost chip rendered bottom-left in hand-zone presentations, hidden on battlefield. Uses the same gold-leaf/badge visual treatment as the S14 auto badge.
- [x] `client/src/lib/components/board/ManaAbilityMenu.svelte` — right-click / long-press on a battlefield permanent surfaces the ability list from `card.mana_abilities`; click fires `activate_mana_ability`. Basic lands with the synthetic ability get a convenience single-click tap-for-mana.
- [x] Auto-tap-and-cast UX: cast-targeting flow detects mana-short state, calls a new `GET /games/:id/auto-tap-preview?card=<instanceID>` endpoint (read-only, no state mutation) returning the candidate plan, renders a confirm modal listing which permanents will tap. Enter confirms (fires `cast_spell` with `auto_tap: true`); ESC cancels.
- [x] Lock-tap override: clicking a land BEFORE the cast locks it into the auto-tap excluded set; repeated clicks toggle. The preview modal highlights locked sources.
- [x] Strict-mode error surface: `insufficient_mana` error frame renders a toast with an "Override strict mode for this cast" button; re-fires the action with `force_cast: true`.
- [x] New `PendingChoiceKind: "mana_pick"` branch in `ChoicePromptModal.svelte` — renders a 5-button color picker for Arcane Signet / Birds of Paradise. Color options filtered server-side per the produced-string pipe set intersection; client just renders the buttons the wire sends.

**Tests:**
- [x] Server: `effects/cost_test.go` — parse table over 100 cost strings including every pipe / `{X}` / `{W/P}` / `{S}` shape.
- [x] Server: `game/mana_test.go` — `ManaPool.AddMana` / `SpendMana` / `EmptyPool` round-trips; step-change clears; undo rolls back.
- [x] Server: `game/autotap_test.go` — Cyclonic Rift overload from Bant manabase; Lightning Bolt from 3 basics + Signet; unsolvable cost returns `(nil, false)`; lock-tap excluded set honoured.
- [x] Server: `cards/effects/cards_test.go` extensions — Sol Ring tap adds `{C}{C}`; Arcane Signet tap queues PendingChoice with the commander's identity subset; Birds of Paradise tap queues unrestricted 5-color pick.
- [x] Server: `mutations_test.go` — `TestCastSpellStrictModeRejectsShort`, `TestCastSpellForceCastOverrides`, `TestCastSpellPermissiveWarnsAndProceeds`, `TestCastSpellCommanderTaxStrictIncludesSurcharge`.
- [x] Client: vitest on `ManaPoolPips.svelte` (pip rendering), `settings.ts` v3→v4 migration (strictMana seeded false, existing keys preserved), `ChoicePromptModal.svelte` mana_pick branch.
- [x] Manual 2-tab smoke through the exit-criteria list below.

**Docs:**
- [x] `docs/decisions/0011-mana-pool-and-auto-tapper.md` — ADR covering the pool-as-slice choice, pipe syntax, single-action dispatcher, auto-tapper algorithm, permissive-default rationale, strict-mode override hatch.
- [x] `docs/protocol.md` — `mana_pool`, `mana_cost`, `mana_abilities`, new EventKinds, `insufficient_mana` error frame shape, `mana_pick` PendingChoice kind.
- [x] `AGENTS.md` — "how to add a mana ability to a catalog card" recipe addendum.

### Out of scope (explicit handoffs)
- **Filter lands** (Mystic Gate's "`{1}`, pay sub-mana: add `{W}{W}` or `{U}{U}`") — the sub-payment mechanic is an S17 replacement-effect collaboration. S15 treats filter lands as vanilla dual-produce sources for the auto-tapper (good enough until filter-only costs show up).
- **Cavern of Souls tribe-locking** (the produced mana carries "spend this on a creature type X spell only") — the `ManaToken.Restrictions` slot is scaffolded, but the catalog card ships in a later sprint.
- **Alternative costs** — Force of Will's pitch cost, suspend, flashback, madness, overload-as-alt-cost. Overload-AS-a-mode-choice is in-scope (Cyclonic Rift is an exit-criterion card) because overload is still a mana cost — just a *different* one announced at cast time via `Modes`.
- **Phyrexian mana self-pay life cost.** `{W/P}` parses and the auto-tapper treats `W/P` symbols as payable by white mana; paying 2 life instead is NOT implemented. Announce-time "2 life or `{W}`" picker lands with replacement effects (S17) or when a phyrexian-mana card enters the catalog, whichever comes first. Documented in ADR.
- **Hybrid mana preferences.** `{W/U}` parses; the auto-tapper picks one half greedily (whichever the restriction-first heuristic scores higher). No user-facing "I want to pay this with blue specifically" affordance.
- **Snow mana sources.** `{S}` parses but every mana source counts as snow — no Snow-Covered Forest distinction. Real snow routing lands with the S16 layers work.
- **`X` as a cost in the auto-tapper.** `{X}` parses at announce time via `params.XValue`; the auto-tapper uses the announced value. Mid-cast X sliders with live recompute are out.
- **Cost replacement effects** — Trinisphere, Thalia, Spellshift, Kambal's life-gain-on-cast. All S17 territory.
- **Activated non-mana abilities** — the catalog non-mana activated pipeline is S19. S15 special-cases only mana abilities (CR 605).

### Risks / gotchas

- **Auto-tapper complexity vs. correctness.** Backtracking can explode on pathological manabases (12+ dual lands, filter lands, tri-lands). Mitigation: 10k-node budget cap + fallback to manual tapping. Benchmark target documented in ADR; exit criterion #8 enforces it.
- **Settings migration race.** v3→v4 merges in-place; a user on multiple tabs with a v3 cache and a v4 cache can flip-flop. Same shape seen with S13→S13-hotfix `autoPassPriority`. Mitigation: the migration is idempotent, strictMana defaults false, and the store subscribe overwrites localStorage on every mutation — convergence on next settings touch.
- **Commander tax double-count.** The informational `CommanderCasts` counter is incremented in `CastSpell` after zone move (S13.1). The strict-mode surcharge reads `CommanderCasts[cardID]` at `effectiveCostLocked` time BEFORE the increment — so the first cast pays `cost+0`, second `cost+2`, matching CR 903.8. Unit-test the boundary: `TestCastCommanderTwiceStrictMode`.
- **Mana pool clearing on step-change undo.** Step-advance is already one undo frame; the pool-empty event rides inside it, so undo rolls the pool back automatically. Explicit test: `TestManaPoolSurvivesStepAdvanceUndo`.
- **Strict-mode-on with incomplete cost wedges the caster.** If a player toggles strict, gets their cost rejected, and doesn't know how to override, they're stuck. Mitigation: the `insufficient_mana` error frame carries a `missing` slice and the client surfaces an "Override strict mode for this cast" button (force_cast=true). Documented + tested.
- **Synthetic basic-land ability vs. custom-printed lands with basic subtypes.** A "Dryad Arbor" (creature-land with type Forest) should get the Forest synthetic ability but not *only* that. For S15 the synthetic path keys on TypeLine containing a basic subtype AND no catalog entry existing; a catalog `ManaAbilities` slice overrides it wholesale. The test matrix covers basic Forest, basic Island, Sol Ring (catalog), and Birds of Paradise (catalog + creature).
- **Pipe-produce inside auto-tap-and-cast.** If the auto-tapper plans to tap an Arcane Signet for one of the cost's requirements, the mana-pick PendingChoice would ordinarily fire mid-cast — deadlocking the cast flow. Mitigation: the auto-tapper materialises a specific color choice inline (it already has the cost it's solving for, so the pick is determined); no PendingChoice is queued when the ability is fired via the auto-tap path.
- **Pre-S15 replay snapshots.** No `mana_pool` / `mana_cost` fields ⇒ decodes as empty pool + empty costs ⇒ sandbox-equivalent behaviour. Confirmed non-breaking by replay round-trip test.

### Exit criteria

1. **Parser table** — 100-row cost-string table all parse without error; every symbol class is represented.
2. **Basic-land tap** — click Forest, Forest taps, pool gains `G`, `mana_pool` wire field reflects; advance step → pool empties.
3. **Sol Ring** — right-click Sol Ring → "Add {C}{C}" menu entry → click fires, pool gains two colorless, Sol Ring tapped, no PendingChoice.
4. **Arcane Signet in a Bant deck** — activate → PendingChoice opens with W / U / G buttons (not B, not R — commander identity enforced server-side); pick blue → pool gains `U`.
5. **Birds of Paradise** — activate → PendingChoice opens with all 5 colors; pick any → pool gains that color.
6. **Cyclonic Rift overload from 38-land Bant manabase** — cast with `auto_tap: true` → auto-tapper returns a 7-source plan (6 blue + 1 generic) in under 1 ms p99; preview modal lists the plan; Enter confirms; Rift resolves; battlefield cleared.
7. **Strict-mode gate** — toggle `strictMana = true`, cast Lightning Bolt with zero mana → error frame with `missing: ["{R}"]`; toast shows "Override strict mode for this cast" button; click → re-fires with `force_cast: true` → cast proceeds.
8. **Auto-tapper budget** — 10-permanent pathological manabase (filter lands + duals + basics) resolves or times out cleanly under 1 ms p99 with no goroutine wedge.
9. **Commander tax under strict** — cast commander from command zone once (cost `{2}{W}{U}{G}`), put it back, cast again → effective cost is `{4}{W}{U}{G}`; cast third time → `{6}{W}{U}{G}`.
10. **Lock-tap override** — pre-click a specific Island, then cast Counterspell: preview modal shows that Island pre-checked (locked); auto-tapper picks a second blue source from the remaining pool.
11. **Step-change empties pool** — add `{U}{U}` to pool manually, advance to combat step → pool clears; undo → pool restored.
12. **Non-catalog cast regression** — cast a card not in the catalog with permissive mode → cost warning event emitted, cast proceeds, sandbox posture unchanged. Confirmed the opt-in invariant still holds.

### Sub-PR split
Stop-and-show for manual testing at each boundary.
1. **Catalog scaffold + Scryfall ingestion.** `cards.Card.ManaCost` / `ProducedMana` + `game.Card` mirror + `effects.ParseCost` + `effects.ParseProducedMana`. No behavior change; cost strings render on hand-zone cards as inert chips.
2. **`ManaPool` + manual mana abilities.** `Player.ManaPool` + `AddMana`/`SpendMana`/`EmptyPool` + step-change hook + `activate_mana_ability` action + the three catalog mana abilities + `PendingChoiceKind: "mana_pick"`. Stop-and-show: right-click Sol Ring, see pool fill; step advance, see it empty.
3. **Cost validation in `cast_spell`.** Permissive default + `strictMana` setting migration + `ForceCast` override + commander-tax surcharge under strict. Stop-and-show: toggle strict, try to cast Bolt with no mana, see the override toast.
4. **Auto-tapper algorithm.** `AutoTapForCost` + `AutoTapForCostExcluding` + `/games/:id/auto-tap-preview` read-only endpoint + 10k-node budget. Stop-and-show: server-side benchmark hits the p99 bar; manual UI path still unchanged.
5. **Auto-tap-and-cast UX.** Preview modal + Enter/ESC + lock-tap integration + `cast_spell` with `auto_tap: true`. Stop-and-show: cast Cyclonic Rift overload from a Bant board with one keyboard shortcut.
6. **ADR 0011 + protocol docs + sprint-sheet cleanup.** ADR, protocol updates, AGENTS recipe, S15 section rolled up.

### Critical files
- Server: new [server/internal/game/mana.go](../server/internal/game/mana.go) (pool + tokens), new [server/internal/game/autotap.go](../server/internal/game/autotap.go), [server/internal/game/player.go](../server/internal/game/player.go) (ManaPool field), [server/internal/game/mutations.go](../server/internal/game/mutations.go) (CastSpell gate + ActivateManaAbility), [server/internal/game/turn.go](../server/internal/game/turn.go) (step-change pool empty), new [server/internal/cards/effects/cost.go](../server/internal/cards/effects/cost.go), [server/internal/cards/effects/spec.go](../server/internal/cards/effects/spec.go) (ManaAbilities slot), [sol_ring.go](../server/internal/cards/effects/sol_ring.go) / [arcane_signet.go](../server/internal/cards/effects/arcane_signet.go) / [birds_of_paradise.go](../server/internal/cards/effects/birds_of_paradise.go), [server/internal/cards/index.go](../server/internal/cards/index.go) (new JSON fields), [server/internal/deck/deck.go](../server/internal/deck/deck.go) (stamp onto game.Card), [server/internal/actions/actions.go](../server/internal/actions/actions.go) (new type), [server/internal/protocol/view.go](../server/internal/protocol/view.go) (PlayerView.ManaPool, CardView.ManaCost/ManaAbilities, error frame), [server/internal/game/pending_choice.go](../server/internal/game/pending_choice.go) (mana_pick kind).
- Client: new `client/src/lib/components/board/ManaPoolPips.svelte`, new `client/src/lib/components/board/ManaAbilityMenu.svelte`, [PlayerHeader.svelte](../client/src/lib/components/board/PlayerHeader.svelte) (pip row), [Card.svelte](../client/src/lib/components/board/Card.svelte) (cost chip), [ChoicePromptModal.svelte](../client/src/lib/components/board/ChoicePromptModal.svelte) (mana_pick branch), [settings.ts](../client/src/lib/settings.ts) (v4 migration + strictMana), [protocol.ts](../client/src/lib/protocol.ts) (mirrors).
- Docs: new [docs/decisions/0011-mana-pool-and-auto-tapper.md](decisions/0011-mana-pool-and-auto-tapper.md), [docs/protocol.md](protocol.md), [AGENTS.md](../AGENTS.md).

### Branch + commit conventions
- Branch: `feat/s15-mana-pool-and-auto-tapper` (sub-PRs use `-s15-<slice>` suffixes — e.g. `feat/s15-catalog-ingest`, `feat/s15-auto-tapper`).
- Footer:
  ```
  Sprint: S15 — Mana pool, cost model, and auto-tapper
  Issue: #65
  ```

---

## S16 — Continuous effects + layer system (CR 613)
**Phase:** 7 · **Goal:** ship the 7-layer skeleton with timestamp ordering; ~10 catalog cards exercising each layer.

- [ ] `Characteristic` snapshot type (printed vs. effective)
- [ ] Layer engine (1 copy, 2 control, 3 text, 4 type, 5 color, 6 abilities, 7 P/T with sub-layers 7a-7e)
- [ ] Recompute-from-scratch on state change, cached by version number
- [ ] `StaticAbility` declarations on `CardImpl` (placeholder from S14, now populated)
- [ ] ~10 catalog cards: Mycosynth Lattice, Conspiracy, Lord of Atlantis, Glorious Anthem, Crusade, Honor of the Pure, Tarmogoyf, Mind Control

**Skip dependency detection (CR 613.8) initially** — pure timestamp ordering works for ~95% of real cards. Add when a problem card surfaces (Opalescence + Humility — niche in casual EDH).

**Exit criteria:** Glorious Anthem in play → all your creatures show +1/+1 on the wire; remove anthem → reverts.

---

## S17 — Replacement effects engine (CR 614)
**Phase:** 7 · **Goal:** effects that watch for events and substitute different events before they happen.

- [ ] Replacement engine with iterative apply-loop
- [ ] Affected-player-chooses-order prompt (CR 616) via paused server prompt
- [ ] Self-replacement once-per-event tracking (CR 614.5)
- [ ] Pipeline integration in `AddCounter`, `MoveCardByID`, `DrawCard`, `ChangePlayerLife`, `MarkDamage`
- [ ] Built-in replacements: commander zone (refactored from S13.1), enters-tapped, skip-step
- [ ] ~10 catalog cards: Doubling Season, Hardened Scales, Branching Evolution, Champion of Lambholt, Hangarback Walker, Stasis, Kismet
- [ ] Damage prevention sub-category (CR 615)

**Exit criteria:** Doubling Season + Hardened Scales in play → cast a counter-placing card → prompt for order → 4 counters land per chosen order.

---

## S18 — Combat keywords
**Phase:** 7 · **Goal:** 12 keyword effects with full combat behavior + summoning sickness.

- [ ] Keyword detection helpers (`HasKeyword`, `IsFlyingBlockable`, `BlockerCountValid`)
- [ ] Combat damage flow rewrite (first-strike + regular sub-steps)
- [ ] Lifelink, deathtouch, trample, vigilance
- [ ] Flying / reach (block restriction)
- [ ] Menace (≥2 blockers)
- [ ] Defender, haste, flash
- [ ] First strike + double strike
- [ ] Summoning sickness (`Card.SummonedThisTurn`)
- [ ] Damage assignment order prompt (CR 510.1c)
- [ ] 12 catalog cards demonstrating each keyword

**Out of scope:** protection, indestructible, hexproof, shroud, ward, banding, rampage, flanking, fear, intimidate, shadow.

**Exit criteria:** A 1/1 deathtouch attacker takes down a 5/5 blocker; lifelink attackers gain life; trample carries over; vigilance keeps attackers untapped.

---

## S19 — Auto-fire triggered abilities
**Phase:** 7 · **Goal:** ETB / dies / upkeep / cast / combat triggers fire automatically for catalog cards.

- [ ] Auto-fire dispatcher: register listeners on zone change; LKI snapshot at trigger time (CR 603.10)
- [ ] Optional / modal trigger prompts via existing prompt frame
- [ ] ~25 catalog cards: Mulldrifter, Eternal Witness, Reclamation Sage, Acidic Slime, Solemn Simulacrum, Smothering Tithe, Esper Sentinel, Edric Spymaster of Trest, Phyrexian Arena, Sylvan Library, …

**Manual fallback preserved:** `announce_trigger` from S13.1 stays for unimplemented cards and "hidden info" triggers.

**Exit criteria:** Cast Mulldrifter → on resolve, you draw 2 cards automatically; advance to upkeep with Phyrexian Arena → trigger goes on stack automatically.

---

## S20 — Auto-target legality + smart cast UI
**Phase:** 7 · **Goal:** capstone sprint — per-card targeting predicates; modal/X UI; structured cast dialog.

- [ ] Predicate library (`AnyTarget`, `Creature`, `NonBlackCreature`, `Spell`, `PowerLE(n)`) + composers (`And`/`Or`/`Not`)
- [ ] Extended `TargetSpec` (predicate + AllowSelf + AllowSameTarget)
- [ ] `ModeSpec` for modal spells (Min/Max choose)
- [ ] Cast dialog rewrite: filter target candidates by predicate; structured mode picker; X-cost live validation
- [ ] Resolution-time re-check using same predicates (CR 608.2b)
- [ ] Catalog updates: add `TargetSpec` predicates to all S14/S17/S18/S19 catalog cards

**Free-form fallback preserved:** cards without structured predicates use S13.1's free-form picker.

**Exit criteria:** Cast Doom Blade → picker shows only non-black creatures; cast Cyclonic Rift overload → no picker, mass effect; cast Fireball → X input with live mana-pool validation.

---

## S21 — Tokens, sacrifice, aristocrats
**Phase:** 7 · **Goal:** an aristocrats Commander deck plays end-to-end.

- [ ] Token catalog (Treasure, Food, Clue, Blood, Map, Powerstone, generic creatures)
- [ ] `SacrificePermanent`, `Proliferate`, `CreateTokenAdvanced` primitives
- [ ] Sacrifice as a cost component
- [ ] ~40 cards: token producers, sacrifice outlets, aristocrats payoffs, proliferate cards
- [ ] Theme-deck smoke test (Korvold-style aristocrats deck plays 3 turns)

**Exit criteria:** Cast Goblin Bombardment + Blood Artist + Krenko, Mob Boss; sacrifice tokens to Bombardment one at a time → opponent's life ticks down (Blood Artist + Bombardment damage); your life ticks up (Blood Artist gain).

---

## S22 — Card draw + library manipulation
**Phase:** 7 · **Goal:** a draw-heavy Commander deck plays end-to-end.

- [ ] `ScryN`, `SurveilN`, `Explore`, `RevealAndChoose`, `MillToZone`, `DrawAndScry` primitives
- [ ] Delayed-trigger mechanism (`Game.DelayedTriggers`) for "at the next end step" patterns
- [ ] `KindLookAtCards` / `KindRevealCards` wire frames (controller-only with redacted view)
- [ ] ~40 cards: passive draw engines, top-of-library manipulation, tutoring, mill, big draw payoffs
- [ ] Theme-deck smoke test (blue draw deck plays 3 turns)

**Exit criteria:** Activate Sensei's Divining Top → personal-only modal shows top 3 cards → reorder → confirm. Necropotence: activate to exile a card → advance to end step → card moves to hand automatically.

---

## S23 — Mass removal + boardwipes
**Phase:** 7 · **Goal:** mass-effect cards work; boardwipes wipe correctly across decks.

- [ ] `DestroyAllMatching`, `ExileAllMatching`, `BounceAllMatching`, `ReturnAllToHand` primitives
- [ ] Predicate-driven mass effects with non-X exclusions
- [ ] ~30 cards: Wrath of God (extended), Damnation, Toxic Deluge, Vandalblast, Austere Command, Farewell, Merciless Eviction, Cyclonic Rift overload (extended), …
- [ ] Theme-deck smoke test (control deck plays 3 turns including a boardwipe)

Detailed plan TBD; lands just-in-time after S22 ships.

---

## S24 — Equipment, auras, attachments
**Phase:** 7 · **Goal:** equipment + auras work as attached state on creatures.

- [ ] `Card.AttachedTo *uuid.UUID` field + wire shape
- [ ] `Attach`, `Detach`, `EquipPay`, `EnchantTarget` primitives
- [ ] ~30 cards: Sword of Feast and Famine, Sword of Fire and Ice, Lightning Greaves, Swiftfoot Boots, Skullclamp (extended), Rancor, Curse-style auras, …
- [ ] Theme-deck smoke test (equipment / voltron-adjacent deck plays 3 turns)

Detailed plan TBD; lands just-in-time after S23.

---

## S25 — Voltron / commander damage focus
**Phase:** 7 · **Goal:** "make commander big and swing" decks work end-to-end.

- [ ] `BoostUntilEOT`, `GiveKeywordUntilEOT`, `HexproofUntilEOT`, `IndestructibleUntilEOT` primitives
- [ ] Until-end-of-turn effect lifecycle (cleanup-step removal of one-shot continuous effects)
- [ ] Per-commander damage UX from S13.1 exercised heavily
- [ ] ~30 cards: Bruna Light of Alabaster, Uril the Miststalker, voltron commanders + support, ramp + protection, big-equipment cards, …
- [ ] Theme-deck smoke test (voltron deck deals 21 commander damage in 3 turns)

Detailed plan TBD; lands just-in-time after S24.

---

## S26 — Tribal / creature type matters
**Phase:** 7 · **Goal:** tribal Commander decks (Goblins, Merfolk, Slivers, etc.) work end-to-end.

- [ ] `ChooseCreatureTypeOnETB` primitive (per-permanent persistent state for Cavern of Souls' named tribe)
- [ ] `GrantTypeUntilEOT`, `TypeFilter` predicate
- [ ] ~30 cards: Cavern of Souls, Door of Destinies, Vanquisher's Banner, Coat of Arms, Adaptive Automaton, tribal lords, changeling creatures, …
- [ ] Theme-deck smoke test (tribal deck plays 3 turns with type-locked Cavern + lord buffs)

Detailed plan TBD; lands just-in-time after S25.

---

## S27 — Card-type completeness (planeswalkers, sagas, vehicles, battles)
**Phase:** 7 · **Goal:** the four card types missing or only partially modeled by S13–S26 become full citizens.

- [ ] `LoyaltyAbility{Cost, Effect}` — proper stack-item activation (CR 606); replaces S13.1's thin "delta-on-action" model
- [ ] Planeswalker uniqueness SBA (CR 704.5j) — controller chooses which to keep
- [ ] `declare_attacker` polymorphic target — attack player OR planeswalker OR battle (CR 506.4, 508.1d)
- [ ] Saga chapter triggers + lore counter advance as turn-based action (CR 714)
- [ ] `CrewCost{N}` cost component + `BecomeCreatureUntilEOT(P, T)` effect (CR 702.122)
- [ ] `BattleSpec{Defense, Subtype}` + `Card.ProtectorPlayerID` + ETB protector prompt + defeating triggers (CR 310)
- [ ] ~30 cards: 8 planeswalkers, 8 sagas, 6 vehicles, 4 battles, 4 counter-payoff bridges
- [ ] Theme-deck smoke test (Superfriends/sagas/vehicles deck plays through 4 turns)

Detailed plan: `/home/node/.claude/plans/s27-card-type-completeness.md`. Builds on S13.1, S13.2 (counters + counter SBAs already shipped), S14, S15, S16, S18, S19, S25.

---

## S28 — Cost modification + alternative casts
**Phase:** 7 · **Goal:** the cost engine that the auto-tapper hooks before pool validation.

- [ ] `CostModifier interface { Modify(*Cost, *Card, *Game, uuid.UUID) *Cost }` registered per static ability
- [ ] Modifier ordering per CR 601.2f: alternative cost → additional costs → increasers → reducers, floor at {1} (CR 117.13)
- [ ] Alternative-cost slots in cast dialog (Force of Will pitch, Fierce Guardianship "if you control a commander")
- [ ] Additional-cost slots (Snuff Out's "pay 4 life", Cabal Therapy's "sacrifice")
- [ ] Cascade primitive (CR 702.85) — exile-until-CMC-less, may cast for free
- [ ] ~25 cards: 6 reducers, 6 increasers (Stax pieces), 4 free-cast, 4 additional-cost, 5 cascade
- [ ] Theme-deck smoke test (Maelstrom Wanderer cascade chain with Goblin Electromancer + Trinisphere on table)

Detailed plan: `/home/node/.claude/plans/s28-cost-modification.md`. Builds on S14, S15, S20. Convoke / Improvise / Delve / Affinity stay deferred (state-scanning at cast time + different UX).

---

## S29 — Alt-cast paths from non-hand zones
**Phase:** 7 · **Goal:** spells cast from graveyard, exile, or hand-with-special-marker.

- [ ] `CastableZones []Zone` per card (default `[Hand]`); cast dialog walks all legal zones and surfaces all legal cast paths as separate buttons with their costs
- [ ] Flashback (CR 702.34) — cast from graveyard for flashback cost; exile after
- [ ] Madness (CR 702.35) — replacement on discard: exile face-down with marker; may cast for madness cost
- [ ] Foretell (CR 702.143) — sorcery-speed `foretell_card` action; cast on a later turn for foretell cost
- [ ] Escape (CR 702.144) — cast from graveyard, exile N cards from graveyard as additional cost
- [ ] Suspend (CR 702.62) — exile with N time counters; auto-fire removes one each upkeep; cast for free with haste-until-EOT when last removed
- [ ] Cycling (CR 702.32) — activated ability of cards in hand; cycling triggers fire from hand
- [ ] Splice (CR 702.47) — addon to instants/sorceries
- [ ] ~30 cards: 8 flashback, 4 madness, 4 foretell, 4 escape, 4 suspend, 4 cycling, 2 splice
- [ ] Theme-deck smoke test (Muldrotha-style graveyard deck casts via flashback + escape)

Detailed plan: `/home/node/.claude/plans/s29-alt-cast-paths.md`. Builds on S14, S17, S19, S20, S22, S28. Dredge/retrace/rebound/aftermath stay deferred.

---

## S30 — Damage prevention, cloning, face-down, deferred protection keywords
**Phase:** 7 · **Goal:** engine-completeness capstone. After S30 there are no major missing primitives.

- [ ] Damage prevention shields (CR 615) — replacement subtype with charges; integrates with S17 pipeline
- [ ] Cloning (CR 706) — ETB replacement that captures copyable values; populates S16 layer 1 copy slot
- [ ] Spell copies (CR 706.10) — `CopyTopOfStack`; controller chooses new targets *before* copy hits stack
- [ ] Morph / manifest (CR 702.36, 701.34, 707) — face-down zone state on `Card`; reuses S22 hidden-info wire frame
- [ ] Protection (CR 702.16) — predicate-based guard at four DEBT hook points (damage / enchant-equip SBA / block / target)
- [ ] Hexproof (CR 702.11) — targeting-by-opponent guard
- [ ] Ward (CR 702.21) — `TriggeredAbility` on `EventBecomesTarget` via S19; pay-or-counter prompt
- [ ] Indestructible (CR 702.12) — flag short-circuits lethal-damage SBA + destroy-effect rejection
- [ ] ~25 cards: 4 fog, 4 cloning, 4 spell copies, 4 morph, 6 protection/indestructible, 3 ward
- [ ] Theme-deck smoke test (Avacyn + Reverberate + Clone + a fog plays 4 turns end-to-end)

Detailed plan: `/home/node/.claude/plans/s30-damage-cloning-protection.md`. Builds on S14, S16, S17, S18 (picks up its deferred set), S19, S22. Phasing / banding / shroud-as-distinct stay deferred.

---

## Post-S30 — Rolling deck-driven catalog growth
**Phase:** 7 · **Status:** rolling, not started.

After S30 the engine is feature-complete for major Commander mechanics and the catalog (~600 cards) is mature enough that incremental work fits in 1-2 day batches. Remaining mechanics (MDFCs, adventures, mutate, energy, day/night, monarch, vehicles-with-saddle, phasing) drop to on-demand work.

- **Deck-import audit:** when the user imports a deck, the lobby UI shows "X% of cards in catalog." If <80%, suggest filing a "missing cards" issue.
- **Small PR cadence:** 5-15 cards per PR, 1-2 day turnaround. No sprint scaffolding.
- **Engine work as needed:** when a card surfaces a missing primitive, ship a tiny engine PR + the card together.
- **Quarterly catalog review:** every 12 weeks, audit "what cards have come up in real games but aren't in the catalog?" Prioritize by appearance count.

Triaged just-in-time from real-play feedback.

---

## S31 — AI bot seat (heuristic policy)
**Phase:** 8 · **Goal:** fill an empty Commander seat with a bot good enough for solo practice and 1–3-friend games. After S27–S30 close the engine gaps (card-type completeness, cost modification, alt-cast paths, damage/cloning/face-down/protection), the engine is Arena-parity with a ~600-card catalog and structured `Effect` descriptors on every card — for the first time in the project, a competent AI is actually buildable. S31 ships that competence as a tiered, swappable policy framework.

**Not in scope: tournament-strength AI.** Forge has spent 15+ years on rule-based heuristics and still plays "dumb but playable." XMage's MCTS variant takes minutes per turn. Academic MCTS + RL work (Cowling-Ward-Powley 2012; MageZero AlphaZero-style) is multi-year research. This sprint's bar is Forge's bar: makes legal moves, makes locally-sensible decisions, doesn't deadlock, uses removal on threats. A learning bot is a separate multi-sprint arc (S40+ or never).

### Design decisions

1. **Heuristic-first.** Rule-based evaluation + greedy action selection, informed by the structured `Effect` descriptors the S14+ catalog already provides. No per-card scripts (the big Forge tax) — the bot reasons over effect types.
2. **Bot = virtual seat.** A bot is a `Player` with no WS connection; a goroutine subscribes to the same game stream, runs through the same visibility filter as a human client, and emits actions via the existing action protocol. Zero protocol forking.
3. **Swappable `Policy` interface.** `Policy interface { DecideAction(view *ViewOfGame, legal []Action) Action }`. S31 ships `HeuristicPolicy`. Future sprints can plug in MCTS or LLM-backed policies — contract stays the same.
4. **4-player-first targeting.** Multi-opponent Commander shapes every targeting decision: threat ranking across 3 opponents, aggression rotation so the bot doesn't tunnel on one seat, concede heuristic so hopeless bots don't drag games out.
5. **Safety rails over cleverness.** Loop detection, illegal-action filter, unknown-card graceful pass. A bot that occasionally passes on a suboptimal play is fine; a bot that crashes the game is not.

### Tasks

**Bot infrastructure (server/`bot/`):**
- [ ] New package `server/internal/bot/` with `Policy interface`, `RandomPolicy` (baseline for tests), `HeuristicPolicy` (S31's deliverable)
- [ ] `bot.Seat{PlayerID, Policy, Difficulty, MinThinkMs}` — the "virtual seat" abstraction
- [ ] Goroutine-per-bot: subscribes to `ws.Room` state deltas using the same `FilterViewFor(playerID)` a client would; emits `Action`s via `actions.Dispatch`
- [ ] Visibility guarantee test: bot only receives the same `ViewOfGame` a human seat at the same seat would, never raw game state
- [ ] `Game.AddBotPlayer(name, deck, policy, difficulty)` seating API
- [ ] Illegal-action filter: every emitted action runs through the same pre-dispatch validation the WS handler uses; on failure the bot silently falls back to `pass_priority`
- [ ] Loop detector: if the bot emits the same non-`pass_priority` action twice in one priority window, force-pass
- [ ] Unknown-card handler: action choices involving cards without `Effect` specs are scored as "unknown, low priority"; never hard-fail

**Heuristic policy (`server/internal/bot/heuristic/`):**
- [ ] Board-state evaluation `score(view, perspective) float64`:
  - Life total × w_life (default 1.0); clamps to [-∞, 0] on death
  - Cards in hand × w_hand (default 3.0 — hand is resources)
  - Creatures on battlefield: `Σ (power + toughness + keyword_bonus)` × w_creatures; keyword_bonus from a static table (flying +2, deathtouch +3, lifelink +2, vigilance +1, etc.)
  - Non-creature permanents × w_utility (default 1.5 per permanent, +2 per planeswalker, +3 per battle controlled)
  - Untapped mana sources × w_mana (default 1.0)
  - Commander in command zone: 0 penalty if castable, −(current_tax) otherwise
- [ ] Threat ranking across opponents: each opponent gets a `threat_score = weighted(life_rev, board, hand, combo_signals)`. Bots prefer attacking the highest-threat opponent; targeting removal hits the highest-threat opponent's best permanent
- [ ] **Aggression rotation**: if bot has attacked the same opponent 3 turns in a row and hasn't reduced their life, swap to the next-highest threat — stops tunnel-vision into a dead lane
- [ ] Action-selection branches:
  - Untap / upkeep / draw: defer to engine (S13 auto-turn-based actions handle these)
  - **Main phase**: (1) play a land if one in hand, preferring colors that unlock most spells; (2) iterate castable spells, pick the one that maximizes `Δscore = score(view_after_cast) − score(view_before)`; skip any with negative Δscore unless life is critical
  - **Declare attackers**: simulate each possible attack set, pick the one that maximizes damage-to-threat after expected blocks; bias toward commander damage on the threat leader
  - **Declare blockers**: minimize incoming damage subject to "never trade a 5+ power creature for a 2/2"; prefer chump-block at low life
  - **Priority windows with empty stack**: pass unless an instant-speed castable improves score by > threshold
  - **Priority windows with a spell on the stack**: cast counter/removal if `score(view_if_counter) − score(view_if_resolve)` > threshold; otherwise pass
- [ ] Modes / X / targets: greedy per-dimension; for X, cap at `min(available_mana, threat_hp_for_lethal)`
- [ ] Concede heuristic: if `score(self) < concede_threshold` for N consecutive turns AND no upswing potential in hand, bot emits a `concede` action (the sandbox concede path)

**Difficulty tiers:**
- [ ] `Easy`: `RandomPolicy` wrapped in legality filter — picks a random legal action per priority window. Baseline for testing and for new human players.
- [ ] `Medium`: `HeuristicPolicy` with default weights. Default choice.
- [ ] `Hard`: `HeuristicPolicy` with 1-ply lookahead (simulate each top-K candidate action, re-score). Latency allowance 1500ms; bounded by action set size.

**Latency discipline:**
- [ ] `MinThinkMs` per bot — default 600ms; a fast decision is artificially held so the game doesn't feel like the bot is precognitive
- [ ] `MaxThinkMs` hard cap — default 2000ms (Easy), 2000ms (Medium), 4000ms (Hard); scoring budget above that forces a fallback to "best so far"
- [ ] Per-decision instrumentation: log p50/p99/p999 decision latencies; surface in admin tools
- [ ] Test: 4-bot game completes 20 turns in under 20 minutes wall-clock on dev hardware

**Lobby integration:**
- [ ] `POST /games/{id}/seats/bot` (admin-only) — adds a bot seat with `{deck_source, difficulty, personality_hint?}`. Counts toward `MaxPlayers`.
- [ ] Curated bot decks in `server/internal/bot/decks/` — 6 starter decks covering the major Commander archetypes:
  - **Aggro** — Isshin, Two Heavens as One-style wide beatdown
  - **Control** — Kess, Dissident Mage-style spell-based control
  - **Combo** — Zur-light combo with reliable pieces (no infinite-combo-of-the-week)
  - **Ramp-stompy** — Omnath, Locus of Mana-style big creatures
  - **Aristocrats** — Meren of Clan Nel Toth-style sac/recurse
  - **Voltron** — Sram / Rafiq-style commander damage
  - Each deck is catalog-only: every card has an `Effect` spec from S14/S27-S30 so the bot can reason about it
- [ ] Lobby UI (`Lobby.svelte`): "Add bot" button → deck-archetype picker + difficulty dropdown → seat appears. Start gate counts bot seats as satisfied.
- [ ] Admin kick-bot action (`DELETE /games/{id}/seats/bot/{seat}`) for mid-lobby changes.

**Client UI:**
- [ ] Bot-seat visual treatment: `PlayerHeader.svelte` grows a small "BOT" chip with tooltip showing difficulty + deck archetype
- [ ] Avatar slot shows a distinctive bot avatar (simple abstract mark; deliberate visual separation from Discord human avatars shipped in S12.5)
- [ ] "Bot thinking…" indicator: during the bot's decision window, its `PlayerHeader` pulses with a subtle `animate-thinking` class (auto-disabled when S11.5 animations are off)
- [ ] Optional [S11.5 setting] "show bot reasoning" — when on, the bot announces its scored-action summary in chat (debug mode for tuning)
- [ ] Chat messages from bots are prefixed `[BOT]` and colored slightly differently from human messages
- [ ] `client/src/lib/protocol.ts` — `PlayerView` grows `is_bot bool`, `bot_difficulty string?`, `bot_archetype string?`

**Docs:**
- [ ] New ADR `docs/decisions/00NN-bot-architecture.md` covering: why rule-based not MCTS/LLM (Forge-level bar justification), tiered `Policy` interface shape, bot-as-virtual-seat rationale, safety rails (loop detection / legality filter / unknown-card handling), 4-player targeting design (threat ranking, aggression rotation, concede)
- [ ] `docs/bot.md` — user-facing: how to add a bot, difficulty tiers, curated deck list, known limitations (will make "dumb" plays sometimes, no politics / bluffing / deal-making, no inter-turn memory)
- [ ] `AGENTS.md` §5 — env vars (if any) and bot package location
- [ ] `docs/protocol.md` — document the `is_bot` / `bot_difficulty` fields on `PlayerView`

**Tests:**
- [ ] Bot-vs-bot smoke: 4 `HeuristicPolicy` bots play to a winner within 50 turns across 20 consecutive runs — no deadlocks, no infinite loops, no illegal actions
- [ ] Visibility enforcement: bot receives filtered view only; golden test asserts the bot's `view` is byte-identical to what a human at the same seat would see
- [ ] Illegal-action regression: 100-game randomized run emits zero engine-rejected actions (the loop fallback to `pass_priority` catches everything)
- [ ] Latency: p99 decision latency under 2000ms at Medium difficulty, under 4000ms at Hard
- [ ] Concede path: bot at 1 life with empty hand + empty board for 3 turns concedes
- [ ] Aggression rotation: bot that attacks player A three ineffective turns in a row switches to player B on turn four
- [ ] Catalog-coverage gap: deck containing a card without `Effect` spec doesn't break the bot — unknown-card path exercised in test

### Out of scope (explicit handoffs)

- **MCTS / deep-lookahead policies** — the `Policy` interface accepts them; a future sprint (not on this roadmap yet) can ship `MCTSPolicy`. Dependencies are non-trivial (game-state deep copy, parallel simulation, ensemble determinization for hidden info).
- **LLM-backed policy** — same `Policy` slot. Big open questions not resolved by this sprint: API cost at 4× bots × N priority windows, latency, steering against hallucinated actions, running local vs. remote models, moderation.
- **Bot politics / deal-making** — "I'll attack X if you attack Y" is a whole design surface. Bots in S31 don't make deals and don't respond to human chat. A "political layer" (message parsing + deal tracking + deal honouring) is a legit later sprint.
- **Learning / memory across games** — bots are stateless between games; no ELO, no opening-move memory, no opponent-modelling. Makes deterministic testing possible and avoids the "bot learned to beat me specifically" dynamic.
- **Bot deckbuilding** — curated decks only. The Forge `CardRanker` approach is non-trivial and solves a different problem (draft / deck construction) than in-game decisionmaking.
- **Bots in Commander pod-selection / turn-order UX** — normal seat flow; no "spectator bots" or "bot-only games" UI beyond what the admin endpoint enables.

### Risks / gotchas

- **Catalog gaps fall through to the bot.** Every card with `Effect` = nil forces the bot into unknown-card pass. If the catalog is <85% coverage when S31 lands, the bot will feel worse than Forge (which has per-card hand-written logic). Mitigation: curated bot decks are 100% catalog-covered at all times.
- **Hidden-info leak is the biggest correctness trap.** If the bot goroutine accidentally reads raw `Game` state instead of the filtered view, it's cheating undetectably. Hard-gate via type: `HeuristicPolicy.DecideAction` takes `*protocol.ViewOfGame`, never `*game.Game`.
- **Priority loops with mixed human+bot tables** can stall if a bot keeps passing and humans keep passing — nobody advances. Current auto-resolve (S13.1 full-priority-pass) handles this, but a bot that refuses to act on its turn will freeze the table. Force-progress rule: bot must emit a non-`pass_priority` action at least once per main phase when castable spells exist.
- **Concede heuristic is loss-aversion-tuned.** A bot that concedes at the first sign of trouble denies human players the catharsis of finishing them off. Default concede threshold is intentionally conservative — only triggers in truly hopeless positions.
- **Difficulty expectations are emotional, not technical.** "Hard" should feel hard; if 1-ply lookahead is indistinguishable from greedy, rename the tier or widen the weight gap. Tune after the first real playtest.
- **Bot-vs-bot games are a silent debugging gold mine** but also a way to discover engine bugs nobody sees in human play. Capture replays of every test bot game; surface regressions loudly.

### Exit criteria

1. Admin clicks "Add bot", picks a deck archetype and difficulty, a bot seat appears with a "BOT" chip; the Start button enables once all seats are filled (bot seats count as deck-satisfied).
2. A 4-bot game plays to a winner in under 30 minutes of wall-clock time, with no illegal actions, no deadlocks, and no unknown-card hard-fails across a 20-run automated sample.
3. A 1v1 human-vs-bot game at Medium difficulty produces recognisable "playing the game" behaviour: bot plays lands, casts spells when profitable, attacks when profitable, uses removal on high-threat permanents, concedes only in genuinely hopeless positions.
4. Over a 100-game randomized regression the bot produces zero engine-rejected actions.
5. p99 per-decision latency under 2000ms at Medium difficulty on dev hardware; 4-bot game feels paced, not frozen.
6. Swapping `HeuristicPolicy` for `RandomPolicy` in a unit test requires touching exactly one line; confirms the `Policy` interface shape is clean enough for a future MCTS/LLM swap.
7. Documentation covers: how to add a bot, difficulty tiers, curated deck list, what the bot explicitly does not do (politics, learning, deckbuilding).

### Prior art references
- [Forge AI wiki](https://github.com/Card-Forge/forge/wiki/AI) — rule-based heuristics + per-card hints via `CardRanker`; ~95% of cards scripted, not hardcoded. Bar to match on "playable but dumb."
- [XMage (magefree/mage)](https://github.com/magefree/mage) — `ComputerPlayer` + `ComputerPlayerMCTS` variants, target-score evaluation, reworked targeting logic. Inspiration for threat-weighted targeting.
- [MageZero](https://github.com/WillWroble/MageZero) — AlphaZero-style RL over XMage as a gym. Long-horizon work; informs the `Policy` interface shape so this path stays open.
- Cowling / Ward / Powley, ["Ensemble Determinization in Monte Carlo Tree Search for the Imperfect Information Card Game Magic: The Gathering"](https://eprints.whiterose.ac.uk/id/eprint/75050/1/EnsDetMagic.pdf), IEEE Transactions on Computational Intelligence and AI in Games, 2012 — the canonical MCTS-for-MTG paper; argues for ensemble determinization over hidden info. Out of scope for S31 but shapes the future `MCTSPolicy`.

---

