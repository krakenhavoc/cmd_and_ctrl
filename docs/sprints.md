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
| S06 | Hand, battlefield, zones UI | 3 | [#6](https://github.com/krakenhavoc/cmd_and_ctrl/issues/6) | 2026-07-03 | **in progress** |
| S07 | Turn/phase UI + chat + manual priority | 3 | [#7](https://github.com/krakenhavoc/cmd_and_ctrl/issues/7) | 2026-07-17 | planned |
| S08 | First playable sandbox (2-player, milestone) | 4 | [#8](https://github.com/krakenhavoc/cmd_and_ctrl/issues/8) | 2026-07-31 | planned |
| S09 | Polish I — animations + VFX | 5 | [#9](https://github.com/krakenhavoc/cmd_and_ctrl/issues/9) | 2026-08-14 | planned |
| S10 | Polish II — Commander UX (cmd damage, politics) | 5 | [#10](https://github.com/krakenhavoc/cmd_and_ctrl/issues/10) | 2026-08-28 | planned |
| S11 | Polish III — hover preview, undo, spectator | 5 | [#11](https://github.com/krakenhavoc/cmd_and_ctrl/issues/11) | 2026-09-11 | planned |
| S12 | Deploy + 4-player go-live with friends | 6 | [#12](https://github.com/krakenhavoc/cmd_and_ctrl/issues/12) | 2026-09-25 | planned |
| S13+ | **B→C rules graft track** (ongoing) | 7 | TBD at S12 retro | rolling | not started |

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

## S07 — Turn/phase UI + chat + manual priority
**Phase:** 3 · **Goal:** the fiddly but essential parts of a manually-driven game.

- [ ] Turn / phase / step indicator
- [ ] "Pass priority" button per player
- [ ] "Pass until end of turn" shortcut
- [ ] Any-player chat, with player colors and timestamps
- [ ] Quick-action buttons: draw, shuffle, mulligan, untap all, change life
- [ ] Priority indicator ("any responses?") visible across all 4 seats

**Exit criteria:** two players can play a complete turn cooperatively, passing priority manually, with chat, and it feels like a game — not a demo.

---

## S08 — First playable sandbox (milestone)
**Phase:** 4 · **Goal:** complete a 2-player sandbox game in a browser, end-to-end.

- [ ] Mulligan flow (take 7, keep/mulligan, scry for extra mulligans — or simplified to "take N")
- [ ] Combat UI: declare attackers, declare blockers, manually resolve damage
- [ ] Life tracker with history
- [ ] Win/loss screen (manual; player clicks "I lose")
- [ ] Play against yourself across two tabs, record a video

**Exit criteria:** a recorded 2-player sandbox game, start to finish, with no manual state intervention outside the client.

---

## S09 — Polish I — animations + VFX
**Phase:** 5 · **Goal:** make the table feel alive.

- [ ] GSAP integration
- [ ] Card draw / play / tap / untap animations
- [ ] Damage number popups, combat arrows
- [ ] Sound pass (draw, play, tap, damage, turn change)
- [ ] Particle effects on ETB / death

**Exit criteria:** side-by-side video vs. Cockatrice shows a clear "this feels better" delta.

---

## S10 — Polish II — Commander UX
**Phase:** 5 · **Goal:** the Commander-specific differentiators from PLAN.md §5.

- [ ] Command zone as a first-class UI element
- [ ] Commander damage 4×4 grid, always visible
- [ ] Life tracker starting at 40 with history, poison/infect/energy
- [ ] Monarch / initiative / goad markers
- [ ] Politics UI scaffold (deal buttons, promise tokens)
- [ ] Council's dilemma / voting UI

**Exit criteria:** all Commander-specific affordances from PLAN.md §5 are at least rough-rendered.

---

## S11 — Polish III — hover preview, undo, spectator
**Phase:** 5 · **Goal:** the Arena-feel features that Cockatrice and XMage both lack.

- [ ] Hover a card → detail panel with full Oracle text and current game context
- [ ] One-click undo / rewind to last server snapshot
- [ ] Spectator mode (read-only connection to a game)
- [ ] Auto-saved replays (server writes every delta to a game log file)

**Exit criteria:** hover preview works for every card, undo works for the last action, spectating a live game works without interfering.

---

## S12 — Deploy and go live with friends
**Phase:** 6 · **Goal:** real games, real feedback.

- [ ] VPS provisioning script (Docker Compose or plain systemd units)
- [ ] Deploy Go game server + static client
- [ ] Basic observability (structured logs, uptime ping, crash dumps)
- [ ] Play a real 4-player game with friends
- [ ] Triage top 10 pain points from the real game into the S13+ backlog

**Exit criteria:** a real 4-player Commander game happens on the deployed stack, and a prioritised S13+ backlog exists.

---

## S13+ — B→C rules graft (ongoing)

After S12, every sprint adds one slice of rules enforcement to the Go server.
Priorities come from the S12 pain-point triage and subsequent real games.

Likely early wins:
- Auto-untap at the start of each player's untap step
- Auto-draw one card at draw step
- Auto-tap lands for their mana type (mana pool UI)
- Auto-resolve combat damage to life totals
- Auto-move dead creatures to graveyard (state-based action, subset)
- Triggered-ability ETB hooks for the most common cards in the playgroup

**Design rule:** every rule is a pure function of game state. Manual override
remains the permanent fallback. No sprint ever leaves the game in a state where
a player cannot manually resolve an unhandled interaction.

Sprint cadence stays at 2 weeks. Track sprints S13–Sxx in new issues once the
S12 retro happens.
