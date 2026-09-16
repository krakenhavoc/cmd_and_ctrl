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

| #        | Name                                                                 | Phase | Issue                                                          | Due        | Status      |
| -------- | -------------------------------------------------------------------- | ----- | -------------------------------------------------------------- | ---------- | ----------- |
| S01      | Go server + client scaffold                                          | 0     | [#1](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1)     | 2026-04-24 | **done**    |
| S02      | Core game state: zones + turns                                       | 1     | [#2](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2)     | 2026-05-08 | **done**    |
| S03      | Action protocol + state deltas                                       | 1     | [#3](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3)     | 2026-05-22 | **done**    |
| S04      | Lobby, auth, Scryfall pipeline                                       | 2     | [#4](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4)     | 2026-06-05 | **done**    |
| S05      | Deck import + 4-player table layout                                  | 3     | [#5](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5)     | 2026-06-19 | **done**    |
| S06      | Hand, battlefield, zones UI                                          | 3     | [#6](https://github.com/krakenhavoc/cmd_and_ctrl/issues/6)     | 2026-07-03 | **done**    |
| S06.5    | Dynamic deck import from URLs (mini)                                 | 3     | [#37](https://github.com/krakenhavoc/cmd_and_ctrl/issues/37)   | 2026-07-10 | **done**    |
| S07      | Turn/phase UI + chat + manual priority                               | 3     | [#7](https://github.com/krakenhavoc/cmd_and_ctrl/issues/7)     | 2026-07-17 | **done**    |
| S08      | First playable sandbox (2-player, milestone)                         | 4     | [#8](https://github.com/krakenhavoc/cmd_and_ctrl/issues/8)     | 2026-07-31 | **done**    |
| S08.5    | Game-logic cleanup pass (wave 1: gating + room + import)             | 4     | [#43](https://github.com/krakenhavoc/cmd_and_ctrl/issues/43)   | 2026-05-03 | **done**    |
| S09      | Polish I — animations + VFX                                          | 5     | [#9](https://github.com/krakenhavoc/cmd_and_ctrl/issues/9)     | 2026-08-14 | **done**    |
| S10      | Polish II — Commander UX (cmd damage, politics)                      | 5     | [#10](https://github.com/krakenhavoc/cmd_and_ctrl/issues/10)   | 2026-08-28 | **done**    |
| S11      | Polish III — hover preview, undo, spectator                          | 5     | [#11](https://github.com/krakenhavoc/cmd_and_ctrl/issues/11)   | 2026-09-11 | **done**    |
| S11.5    | Per-user settings and preferences (mini)                             | 5     | [#82](https://github.com/krakenhavoc/cmd_and_ctrl/issues/82)   | 2026-09-18 | **done**    |
| S12      | Deploy + 4-player go-live with friends                               | 6     | [#12](https://github.com/krakenhavoc/cmd_and_ctrl/issues/12)   | 2026-09-25 | **done**    |
| S12.5    | Discord identity for players (OAuth + bot + presence)                | 6     | [#59](https://github.com/krakenhavoc/cmd_and_ctrl/issues/59)   | 2026-10-09 | planned     |
| S13      | Priority foundation (rules graft kickoff)                            | 7     | [#62](https://github.com/krakenhavoc/cmd_and_ctrl/issues/62)   | 2026-05-17 | **done**    |
| S13.1    | Stack: cast/resolve/target/counter/trigger/SBA                       | 7     | [#63](https://github.com/krakenhavoc/cmd_and_ctrl/issues/63)   | 2026-06-14 | **done**    |
| S13.2    | Counter mechanics (SBAs + player counters + UI)                      | 7     | [#79](https://github.com/krakenhavoc/cmd_and_ctrl/issues/79)   | 2026-06-28 | **done**    |
| S13.3    | Client-side timing affordance (greyed illegal actions)               | 7     | [#99](https://github.com/krakenhavoc/cmd_and_ctrl/issues/99)   | 2026-07-04 | **done**    |
| S13.4    | Interactive cleanup discard + per-player MaxHandSize                 | 7     | [#103](https://github.com/krakenhavoc/cmd_and_ctrl/issues/103) | 2026-07-11 | **done**    |
| S13.5    | Card visibility + known-by tracking                                  | 7     | [#108](https://github.com/krakenhavoc/cmd_and_ctrl/issues/108) | 2026-07-25 | **done**    |
| S13.6    | Intelligent priority auto-pass (smart skip)                          | 7     | [#168](https://github.com/krakenhavoc/cmd_and_ctrl/issues/168) | —          | **done**    |
| S14      | Card-effect catalog foundation                                       | 7     | [#64](https://github.com/krakenhavoc/cmd_and_ctrl/issues/64)   | 2026-07-12 | **done**    |
| S15      | Mana pool, cost model, and auto-tapper                               | 7     | [#65](https://github.com/krakenhavoc/cmd_and_ctrl/issues/65)   | 2026-08-09 | **done**    |
| S16      | Continuous effects + layer system (CR 613)                           | 7     | [#66](https://github.com/krakenhavoc/cmd_and_ctrl/issues/66)   | 2026-09-06 | **done**    |
| S16.5    | Layer follow-ups: copy effects (CR 613.8 dependency → #668)          | 7     | [#159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/159) | —          | **done**    |
| S17      | Replacement effects engine (CR 614)                                  | 7     | [#67](https://github.com/krakenhavoc/cmd_and_ctrl/issues/67)   | 2026-10-04 | **done**    |
| S18      | Combat keywords                                                      | 7     | [#68](https://github.com/krakenhavoc/cmd_and_ctrl/issues/68)   | 2026-11-01 | **done**    |
| S18.5    | Zone browser + library search (mini)                                 | 7     | [#179](https://github.com/krakenhavoc/cmd_and_ctrl/issues/179) | 2026-11-05 | **done**    |
| S19      | Auto-fire triggered abilities                                        | 7     | [#69](https://github.com/krakenhavoc/cmd_and_ctrl/issues/69)   | 2026-11-29 | **done**    |
| S20      | Auto-target legality + smart cast UI                                 | 7     | [#70](https://github.com/krakenhavoc/cmd_and_ctrl/issues/70)   | 2026-12-27 | **done**    |
| S21      | Tokens, sacrifice, aristocrats                                       | 7     | [#73](https://github.com/krakenhavoc/cmd_and_ctrl/issues/73)   | 2027-01-24 | **done**    |
| S22      | Card draw + library manipulation                                     | 7     | [#74](https://github.com/krakenhavoc/cmd_and_ctrl/issues/74)   | 2027-02-21 | **done**    |
| S23      | Mass removal + boardwipes                                            | 7     | [#75](https://github.com/krakenhavoc/cmd_and_ctrl/issues/75)   | 2027-03-21 | partial     |
| S24      | Equipment, auras, attachments                                        | 7     | [#76](https://github.com/krakenhavoc/cmd_and_ctrl/issues/76)   | 2027-04-18 | planned     |
| S25      | Voltron / commander damage focus                                     | 7     | [#77](https://github.com/krakenhavoc/cmd_and_ctrl/issues/77)   | 2027-05-16 | partial     |
| S26      | Tribal / creature type matters                                       | 7     | [#78](https://github.com/krakenhavoc/cmd_and_ctrl/issues/78)   | 2027-06-13 | planned     |
| S27      | Card-type completeness (planeswalkers, sagas, vehicles, battles)     | 7     | [#92](https://github.com/krakenhavoc/cmd_and_ctrl/issues/92)   | 2027-07-04 | **done**    |
| S28      | Cost modification + alternative casts                                | 7     | [#93](https://github.com/krakenhavoc/cmd_and_ctrl/issues/93)   | 2027-07-25 | planned     |
| S29      | Alt-cast paths from non-hand zones (flashback, suspend, foretell, …) | 7     | [#94](https://github.com/krakenhavoc/cmd_and_ctrl/issues/94)   | 2027-08-15 | partial     |
| S30      | Damage prevention, cloning, face-down, deferred protection keywords  | 7     | [#95](https://github.com/krakenhavoc/cmd_and_ctrl/issues/95)   | 2027-09-05 | partial     |
| Post-S30 | Rolling deck-driven catalog growth                                   | 7     | TBD at S30 retro                                               | rolling    | not started |
| S31      | AI bot seat (legal-move enumeration + tiered policy)                 | 8     | [#89](https://github.com/krakenhavoc/cmd_and_ctrl/issues/89)   | 2027-09-26 | partial     |
| S33      | Surviving a deploy: reconnect, resume, and schema safety             | 6     | [#515](https://github.com/krakenhavoc/cmd_and_ctrl/issues/515) | 2027-10-10 | planned     |

### How to read the status column

`**done**` means the sprint's own **exit criteria** are met, not that every checklist box is ticked
— most sprints leave a primitive or a card-count line behind, and a sprint is not held open by one.
`partial` means load-bearing work has shipped under that sprint's name but the exit criteria are
demonstrably unmet. `planned` means nothing has shipped, and it is the row that does real damage
when it is wrong: a `planned` row is read by agents as "this mechanic is unbuilt", which is how
several of them concluded that shipped features were blocked.

**S22 moved `partial` → `**done**` when its checklist was reconciled against the code**, and it is
worth saying why, because the row looked wrong in both directions for a while. Both of S22's exit
criteria have been demonstrably met since #507, which makes `partial` — "the exit criteria are
demonstrably unmet" — the one reading the legend rules out. What was really keeping the row down
was the tracking issue's checklist, every box of which is still unticked, including four
primitives and two wire frames that shipped under names the checklist does not use. The rule above
is the rule: a card-count line and a smoke test do not hold a sprint open, and [#74](https://github.com/krakenhavoc/cmd_and_ctrl/issues/74)
stays open on its own for those.

**The `feat(sNN)` commit scope has drifted away from this table, and S22 is where it shows.** Nine
commits carry `feat(s22)` — attack triggers (#254), flicker and delayed triggers (#255),
alternative cast costs (#257), the Hashaton deck (#258), staples batches (#261, #267), shocklands
(#268), airbend (#269) and convoke (#271) — and **none of them is card draw or library
manipulation**, which is what S22 actually is. They belong to S19, S22, S28 and rolling catalog
growth respectively. When the two disagree, the sprint *section* below is the scope and the commit
tag is just a label someone typed. Retitling S22–S30 as themed epics, so PRs can reference the
sprint they are really in, is tracked on [#281](https://github.com/krakenhavoc/cmd_and_ctrl/issues/281).

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

**Slot rationale.** Ships _after_ the three polish sprints (S09 animations + sound, S10 Commander UX, S11 hover / undo / spectator) so the settings panel has real toggles to surface, and _before_ S12 deploy so the go-live build ships with settings in place. 1-week mini-sprint scope — the infrastructure is small (single Svelte store + JSON schema + panel UI); most of the surface is wiring existing features up to toggles.

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

- S09 / S10 / S11 are in flight. This sprint assumes the polish work has landed; if S09 slips, S11.5's audio/animation toggles are wiring toggles into features that don't exist yet. Mitigation: order S11.5 strictly _after_ S11 rather than overlapping; scope the toggles that refer to unshipped features as no-ops with a "feature not yet available" disabled state.
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
- [x] Basic observability (structured logs, crash dumps) — server emits JSON slog, crash-recovery snapshot dumps in `CMDCTRL_DATA_DIR`, `systemctl status` as live health signal. The "uptime ping" half was never built: nothing watches `/healthz` from outside, which is why an Aug 6 – Sep 8 crash loop went unnoticed for a month. Split out to [#598](https://github.com/krakenhavoc/cmd_and_ctrl/issues/598)
- [ ] Play a real 4-player game with friends — real games are happening on the deployed stack, but 2-player so far. The 4-player table carries to S32 ([#277](https://github.com/krakenhavoc/cmd_and_ctrl/issues/277)) exit criterion 6
- [x] Triage top 10 pain points from the real game into the S13+ backlog — delivered as a sprint of its own, S32 ([#277](https://github.com/krakenhavoc/cmd_and_ctrl/issues/277)), from the 2026-09-10/11 games

**Exit criteria:** a real 4-player Commander game happens on the deployed stack, and a prioritised S13+ backlog exists.

**Status:** done. The deployed stack is live and in real use — `https://cmd.labxp.io` serves the client, `/healthz` returns 200, auto-deploy on merge to `main`, scheduled scryfall refresh, nightly e2e. Real games produce a steady stream of in-app bug reports, and their triage became S32.

Closed with one exit criterion moved rather than met: the table has not yet been a 4-player one. That box is S32's, so the tracker closes rather than sitting open behind a playtest that a later sprint already owns. Infrastructure follow-ups the go-live exposed — host addresses that can rot, unbounded unit restarts, no external health check, a deploy that depends on one runner's own SSH key — are [#598](https://github.com/krakenhavoc/cmd_and_ctrl/issues/598). The Discord bot is still not installed on the prod host ([#249](https://github.com/krakenhavoc/cmd_and_ctrl/issues/249)).

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

- [x] New top-level directory `bot/` or a cmd under `server/cmd/bot/` — picked `server/cmd/bot/` + `server/internal/bot/` in ADR 0004 (shared module, separate binary). Go, using `bwmarrin/discordgo`. Allow-list gate via `CMDCTRL_DISCORD_GUILD_IDS`.
- [x] Slash command `/cc-invite [name]` — calls server `POST /games` with admin credentials (bot holds `CMDCTRL_ADMIN_TOKEN` via env), posts the invite link back to the channel (channel-visible embed; ephemeral-toggle deferred).
- [ ] Slash command `/cc-invite-dm @user [name]` — deferred (needs invite-side pre-bind of DiscordID; not in MVP).
- [x] Slash command `/cc-games` — ephemeral list of active/lobby games (invite tokens already stripped by `Lobby.List`). `/cc-end <id>` deferred (destructive, wants confirmation UX).
- [x] Bot deploys as a second systemd unit on the same VPS (S12 infra). Unit at `deploy/cmd-and-ctrl-bot.service`; env file separate from the server's (ADR 0004 §6).

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

- [x] `docs/decisions/0004-discord-identity.md` — ADR. Why OAuth + bot together; PKCE + state handling; bot deployment shape; gateway-vs-webhook rationale; secret-handling split (Rich Presence deferred, documented as out-of-scope).
- [ ] `docs/lobby.md` — document the new auth routes + the `SeatInfo` field additions.
- [x] `AGENTS.md` §5 — env vars (`CMDCTRL_DISCORD_CLIENT_ID` / `_SECRET` + bot token / app ID / guild IDs / server + client base URLs). RPC client ID lands with Rich Presence.

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

**Phase:** 7 · **Goal:** model priority + turn-based actions per CR 117 / 502 / 504 / 514. The Untap and Cleanup steps stop granting priority. Untap and Draw fire automatically on step entry (with the Turn-1 skip-draw exception per CR 103.8a). Priority rotation skips eliminated seats. The client adds per-step "stops" preferences and a "Pass to my next stop" button so a 4-player turn doesn't require 30+ manual `pass_priority` clicks per cycle.

This sprint is the _foundation_ for the entire S13.x rules-graft track: S13.1 (stack), S13.2 (counter SBAs), S13.3 (greyed illegal actions), S13.4 (interactive cleanup discard), and S13.5 (card visibility) all assume the priority + turn-based-action shape lands here. Cleanup-discard is intentionally deferred to S13.4 — S13's cleanup step just auto-advances to the next turn without granting priority, and S13.4 will inject the discard pause into that gap.

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
  - `StepDraw`: auto-draw 1 card for `ActiveSeat`, **except** when `Turn.Number == 1 && ActiveSeat == Game.StartingSeat` (CR 103.8a).
  - `StepCleanup`: in S13, auto-advance to the next turn's Untap so the cleanup cursor never sits idle. (S13.4 injects the discard pause here.)
- [ ] Manual `TypeUntapAll` / `TypeDrawCard` actions stay dispatchable as sandbox overrides; gate them so they no-op (or return a soft error) during their auto-fire steps to avoid double-fire.

**Eliminated-player skip in priority rotation (server):**

- [ ] `Game.PassPriority`: replace `next = (PriorityHolder + 1) % numSeats` with a loop that walks past `Eliminated == true` seats (reuse the iteration shape from `advancePastEliminatedLocked` at [mutations.go:517](../server/internal/game/mutations.go)).
- [ ] When wrapping back to the active seat through skips, trigger the same step-advance branch as the all-passed case.

**Turn-1 skip-draw rule, CR 103.8a (server):**

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

**Post-ship fix (`refactor/major-wip`, 2026-07):** stack items now resolve strictly LIFO by an insertion `Seq` (CR 608.1) instead of map-iteration order, so an ability activated/triggered in response to a spell resolves first; also closed a stack overflow where `runStateChecksLocked` recursed on itself instead of draining `PendingTriggers` (a manually announced trigger + any SBA check could blow the stack). Commit: `fix(game): resolve stack items LIFO by insertion sequence`.

**Exit criteria:** a Commander player can cast Lightning Bolt, opponent counters with Counterspell on the stack, full priority/SBA loop works end-to-end.

**Bright line — out of scope (S14+):** auto-fire of triggered abilities from card events, auto-validation of target legality at announce, auto-resolution of spell effects, mana pool, replacement effect engine generally, static abilities, combat keyword effects.

---

## S13.2 — Counter mechanics (SBAs + player counters + UI)

**Phase:** 7 · **Goal:** close out the "counters" surface the engine still has to resolve manually. S13.1 ships the four canonical Commander SBAs; this sprint adds the counter-specific SBAs (planeswalker loyalty, battle defense, +1/+1/-1/-1 cancel, poison player-loss, saga final-chapter), first-class UI treatment of counters, and a shared registry of MTG counter types. Sits between S13.1 and S14 so the effect catalog (S14) can rely on counters being fully modelled.

Gap analysis behind this sprint: `Card.Counters` exists today ([server/internal/game/card.go](../server/internal/game/card.go)) and `CurrentPower()` already applies +1/+1 / -1/-1 to P/T. S13.1 covers the Commander SBAs but explicitly not the counter SBAs. S14's `AddCounters` primitive and S17's replacement engine assume counter data is a first-class shape. Nothing in the roadmap (as of issues [#62](https://github.com/krakenhavoc/cmd_and_ctrl/issues/62)–[#70](https://github.com/krakenhavoc/cmd_and_ctrl/issues/70)) covers player-level counters, counter-specific SBAs, or client pip rendering.

### Tasks

**Counter-specific state-based actions (server):**

- [x] CR 704.5i — planeswalker with 0 loyalty counters → owner's graveyard
- [x] CR 704.5v — battle with 0 defense counters → owner's graveyard
- [x] CR 704.5q — `+1/+1` and `-1/-1` cancel 1-for-1; runs BEFORE the lethal-damage / 0-toughness destruction passes (CR 704.3)
- [x] CR 704.5c — player with ≥10 poison counters loses the game
- [ ] CR 704.5s — saga final-chapter sacrifice: deferred to S14+ alongside the effect catalog (needs per-card final-chapter metadata)

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
- **Infect / wither damage as counter placement** — depends on combat keyword layer (S18 [#68](https://github.com/krakenhavoc/cmd_and_ctrl/issues/68)); poison-counter SBA ships here but the _creature-inflicts-poison_ mechanic lives with other combat keywords
- **Persist / undying counter-conditional triggers** — S19 (needs event log + trigger framework)
- **Proliferate** — S14 effect primitive; ships alongside the effect catalog
- **Counter animations beyond simple fade** — S09 polish work; not this sprint

### Risks / gotchas

- `Card.CurrentPower()` already encodes layer-7d math; S16's layer system will eventually want to own this. Don't over-invest in refactoring `CurrentPower()` now — S16 will absorb it cleanly, and duplicating the +1/+1 math elsewhere will just create churn.
- The canonical counter-type list is long and rarely-used types outnumber common types. Ship the list structured but small (top ~20 types get icons; rest are text-only) — over-designing the iconography is a trap.
- `+1/+1 / -1/-1` SBA ordering matters: the cancel-SBA runs _before_ the lethal-damage SBA, so a 2/2 with a -1/-1 counter and 1 marked damage shouldn't die if there's also a +1/+1 counter to cancel. Snapshot-test the ordering explicitly.
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

After S13 + S13.1, the server enforces timing: who holds priority, sorcery-speed window, split-second blocks, once-per-turn loyalty, etc. Today the only UI signal is an error toast _after_ an illegal click. S13.3 closes the loop: hand cards grey out when not castable, ability buttons disable with a reason tooltip, the play-land button greys after the one-per-turn is used. Server enforcement stays authoritative — UI greyness is best-effort UX.

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
- Replacement effects on discard (Library of Leng's "to top/bottom of library instead") — [S17](#s17--replacement-effects-engine-cr-614). *Corrected 2026-09-16:* Leng puts a card discarded by an **effect** on **top** of the library; it doesn't apply to costs or to this cleanup discard, which is a turn-based action, not an effect (true even when a later effect like Null Profusion gives Leng's controller a maximum hand size again). No discard reaches the replacement pipeline yet: [#650](https://github.com/krakenhavoc/cmd_and_ctrl/issues/650), after [#651](https://github.com/krakenhavoc/cmd_and_ctrl/issues/651). See [ADR 0013 §10a](decisions/0013-replacement-effects.md)
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
- [x] `ShuffleLibrary` clears KnownBy on every library card (CR 701.24)
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
- [x] Mixed hand rendering (face-up for revealed-by-Thoughtseize cards) — shipped in S14 (#136), together with the opponent-hand-per-card filter it depended on. `keepKnownInHandZone` (`server/internal/protocol/view.go`) sends a seated viewer only the opponent hand cards they know, and `Hand.svelte` draws those face up with backs for the rest of the count. Ticked at the #170 closeout; this line had stayed "deferred" since S13.5

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

## S13.6 — Intelligent priority auto-pass (smart skip)

**Phase:** 7 · **Goal:** make priority passing feel intelligent. S13 ships the per-step stops grid and `autoPassPriority`; S13.3 ships the client legality predicates. S13.6 joins them: the stops grid now means _"stop when there's something to consider"_, not _"stop every time."_ In practice an empty hand + empty board sees the cursor walk Untap → Untap with zero pointless clicks.

### Tasks

**Client — aggregate-legality predicate (`client/src/lib/priority.ts`):**

- [x] `hasAnyLegalResponse(snap, viewerID, snapSeq?): boolean` — folds `canCastFromHand` + `canActivateAbility` across the viewer's hand, command zone, and viewer-controlled battlefield cards. Returns true as soon as any predicate returns legal.
- [x] Memoised by `snap.seq` so the autoPassPriority effect + future consumers share one scan per snapshot window.
- [x] Conservative by design: treats every viewer-controlled permanent as _potentially_ activatable (client doesn't carry per-card ability lists). False-positive-stop > false-negative-skip per ADR 0009 §3.

**Client — smart-skip setting:**

- [x] Settings schema v5: `gameplay.smartAutoPass: boolean` (default `true`); migration note added to `migrate()` hook.
- [x] Settings.svelte gameplay tab: add the toggle with help text under the stops grid.
- [x] `Game.svelte` autoPassPriority effect folds the predicate in — if the step is a configured stop but `hasAnyLegalResponse` is false and `smartAutoPass` is on, auto-pass anyway.

**Client — pared-down priority toolbar (`next` + `autopass` + `pass turn`):**

- [x] Removed the `next step`, `⇥ pass step`, and `→ next stop` buttons (`advanceStep`, `passRound`, `passToNextStop` functions deleted along with them). The intelligent auto-pass effect (stops grid + smartAutoPass + manual pins) makes the batch-pass buttons redundant.
- [x] `next` = one-shot pass_priority (renamed from "pass priority"). Only active when the viewer holds priority.
- [x] `autopass` = session toggle (not a one-shot). When on, every priority window the viewer holds auto-passes regardless of stepStops / smartAutoPass / manual pins / `settings.autoPassPriority`. Persists until toggled off or reload; the effect still requires actual viewer priority so opponents' turns don't generate "you do not hold priority" rejections.
- [x] Toggle button uses amber active-state styling + aria-pressed so the on/off state reads at a glance without reading the label.

**Client — manual one-time stops (click-to-pin on phase icons):**

- [x] `client/src/lib/priorityStops.ts` — Svelte store of pinned `StepID`s with `toggleManualStop` / `hasManualStop` / `consumeManualStop` / `canManuallyStop` helpers. Rejects Untap + Cleanup (no priority).
- [x] PhaseDisplay.svelte: priority-granting phase icons become `<button>` elements; click toggles a pin. Pinned icons show a small top-right dot + accent tint. No-priority steps stay non-interactive with a "no priority" tooltip.
- [x] Game.svelte autoPassPriority effect: manual pins take precedence over everything else (stops grid, smartAutoPass) when autopass-mode is off. "Fake a game action" — viewer wants the cursor to hold even when the engine has nothing to offer. Autopass-mode overrides pins too (deliberately — the point of autopass is "no more asking").
- [x] Game.svelte consumer: on step transition, consume the pin on the prior step (one-time semantics).

**Tests:**

- [x] `client/src/lib/priority.test.ts` — 14 vitest cases covering hand / command / battlefield / no-priority / split-second / memoisation paths.
- [x] `client/src/lib/priorityStops.test.ts` — 10 vitest cases: toggle, no-op on no-priority steps, multi-pin, consume, subscriber observability.
- [ ] Manual smoke: 4-player game with all stops configured, nobody has instants → cursor walks without clicks; one player casts Counterspell → cursor stops for opposing seats with the mana to respond.
- [ ] Manual smoke: click a phase icon mid-game → cursor holds there next cycle; clears after one pass; multi-pin stacks.

**Docs:**

- [x] ADR `0009-smart-priority-autopass.md`.
- [x] This sprint entry.

### Out of scope (explicit handoffs)

- **Mana affordability in the predicate.** Requires `/auto-tap-preview` on the hot path; perf cliff. Future sprint will precompute affordability once per snapshot if the value shows up in playtest.
- **Triggered-ability responses.** S19 auto-fires them; the viewer doesn't dispatch them, so they don't gate priority-window UX.
- **Opponent "thinking" / "nothing to do" indicators.** Accurate only post-S13.5 hand visibility across all seats (not granted today).
- **Per-player granularity on the toggle.** Global for now; split on demand.

### Risks / gotchas

- **False-negative skip eats a response.** If `hasAnyLegalResponse` returns false but the viewer _could_ have acted, smart-skip passes their priority window. Mitigation: the predicate is deliberately permissive (viewer-controlled permanent → true) so the realistic failure mode is the opposite (stop when there's nothing to do). Turning the toggle off reverts to strict stops.
- **Server / client legality divergence.** The client predicate doesn't know about every effect the server knows about. Smart-skip dispatches a real `pass_priority`; worst case is the same "viewer passed through a priority window they could have used" outcome as a manual click — survivable, undo still works.
- **Toggle surprise on migration.** Default-on changes behaviour for existing users. Self-explanatory under the stops grid's help text; if a player finds their stops are getting skipped, the toggle is two clicks away.

### Exit criteria

1. Four-player game, all seats with `smartAutoPass` on, empty hands + empty boards → cursor walks Untap → Untap with zero `pass_priority` clicks.
2. Seat 0 casts Lightning Bolt targeting seat 2. Seat 1 holds Counterspell + U available → cursor stops for seat 1. Seat 3 holds no instants → cursor auto-passes through seat 3.
3. Toggle `smartAutoPass` off → every configured stop blocks on a manual click regardless of hand state (strict pre-S13.6 behaviour).
4. Click `autopass` → button lights amber; every time priority lands on the viewer, cursor auto-passes (including through stepStops and manual pins). Click again → reverts to the intelligent default path.
5. Click an upkeep icon on PhaseDisplay with autopass off → next time priority lands on the viewer at upkeep, cursor holds even with `smartAutoPass` on and nothing legal to do; pin clears on the pass; next turn's upkeep is unpinned.
6. ADR `0009-smart-priority-autopass.md` exists and explains the predicate's conservative stance, the "smartAutoPass default on" decision, the mana-affordability punt, manual-stops as the highest-precedence override, and autopass-mode as the session-scoped override on top of everything.

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

| Family                                      | Cards                                                                                                                                            | Primitives                             |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------- |
| Direct damage                               | Lightning Bolt, Shock, Lightning Helix, Pyroclasm                                                                                                | DealDamage + GainLife iteration        |
| Mass removal                                | Wrath of God, Damnation, Day of Judgment                                                                                                         | DestroyTarget iteration                |
| Counter magic                               | Counterspell, Negate, Swan Song                                                                                                                  | CounterTarget + CreateToken            |
| Draw                                        | Divination, Harmonize, Sign in Blood                                                                                                             | DrawCards + ChangePlayerLife           |
| Mill / discard                              | Glimpse the Unthinkable (mill 10), Thoughtseize (random — UI deferred to S22), Mind Rot                                                          | MillCards + DiscardCards               |
| Targeted removal                            | Swords to Plowshares, Path to Exile (enters-tapped deferred)                                                                                     | ExileTarget + GainLife + SearchLibrary |
| Bounce                                      | Unsummon                                                                                                                                         | BounceToHand                           |
| Mana rocks (vanilla permanents this sprint) | Sol Ring, Arcane Signet                                                                                                                          | none — S15 wires mana abilities        |
| Tutors                                      | Cultivate (both lands → hand — enters-tapped deferred to S17), Demonic Tutor, Vampiric Tutor                                                     | SearchLibrary (to hand / library-top)  |
| Recursion                                   | Eternal Witness (`OnETB` direct-call hook → S19 migrates to listener), Regrowth                                                                  | ReturnFromGraveyard                    |
| ETB creatures                               | Solemn Simulacrum (enters-tapped land deferred), Acidic Slime                                                                                    | `OnETB` composition                    |
| Vanilla creature placeholder                | Birds of Paradise                                                                                                                                | none — mana ability lands in S15       |
| **Planeswalker**                            | **The Wandering Emperor** (`OnETB` stamps starting loyalty 3 via `AddCounter`; activated abilities remain manual via S13.1's `activate_loyalty`) | **AddCounter on ETB**                  |

**Primitive coverage summary:** 13 of 15 primitives exercised by catalog cards; TapTarget / UntapTarget ship as primitives with unit-test-only coverage so S15 has the hooks ready.

### Out of scope (explicit handoffs)

- **Mana pool + cost validation** — S15. Sol Ring / Arcane Signet / Birds of Paradise ship as vanilla permanents; their mana abilities plug into `effects/` as activated-ability specs in S15.
- **Continuous effects + layers** — S16. No S14 catalog card has a static ability.
- **Replacement effects (ETB-tapped, "if-would-die-exile-instead")** — S17. Cultivate's "enters tapped" half, Path to Exile's tapped land, Solemn Simulacrum's tapped land all defer.
- **Combat keywords (trample, lifelink, flying)** — S18.
- **Auto-fire triggered abilities from catalog cards** — S19. ETB hooks use the direct-call path in S14 and migrate for free.
- **Scry / surveil / look-at-top-N + Brainstorm's "put 2 back" half** — S22 library-manipulation polish. Scry, surveil and look-at-top-N shipped in S22; Brainstorm's "put two cards from your HAND back" is a different shape (it picks from a hand, not from a looked-at library slice) and is still NOT in the catalog.
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
- **Alternative costs** — Force of Will's pitch cost, suspend, flashback, madness, overload-as-alt-cost. Overload-AS-a-mode-choice is in-scope (Cyclonic Rift is an exit-criterion card) because overload is still a mana cost — just a _different_ one announced at cast time via `Modes`.
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
- **Synthetic basic-land ability vs. custom-printed lands with basic subtypes.** A "Dryad Arbor" (creature-land with type Forest) should get the Forest synthetic ability but not _only_ that. For S15 the synthetic path keys on TypeLine containing a basic subtype AND no catalog entry existing; a catalog `ManaAbilities` slice overrides it wholesale. The test matrix covers basic Forest, basic Island, Sol Ring (catalog), and Birds of Paradise (catalog + creature).
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

**Phase:** 7 · **Goal:** ship the 7-layer continuous-effect engine (CR 613) with timestamp-only ordering. Add `Spec.Static []StaticAbility` to the S14 catalog spec, populated by 4 starter cards proving the in-scope layers: Glorious Anthem (7c), Mycosynth Lattice (Layer 4 type-add), Lord of Atlantis (Layer 6 ability-grant + 7c), Tarmogoyf (Layer 7a CDA). Wire-side `CardView.power`, `toughness`, `type_line`, and the new `abilities` field reflect post-layer _effective_ characteristics; printed values stay server-only and feed the recompute. Recompute runs lazily via a `Game.LayerVersion` counter bumped by listener hooks; the snapshot path resolves stale versions before serialising.

The hard sprint of the rules-engine arc — continuous effects are the second-most complex MTG subsystem after the stack. After this sprint, the catalog can include cards whose effects exist _while a permanent is on the battlefield_, not just on resolution. Co-requisite for S18 (combat keywords are layer-6 grants) and S17 (replacement effects need the layered characteristic snapshot).

### Architectural decisions

- **Recompute-from-scratch with a version counter.** No incremental diffs. `Game.LayerVersion` bumps on every event that could affect continuous effects (battlefield zone changes, counter changes, control changes, step advance). `RecomputeLayersIfStaleLocked` runs only when `LayerVersion > LastResolvedVersion`. Snapshot path is the canonical caller — recomputes once per "settled" game state, not per individual mutation. Mirrors XMage / Forge.
- **Two characteristic shapes**: `Card.printedCharacteristic()` (immutable, derived from existing fields) and `Card.Effective()` (cached, post-layer). Wire ships `Effective`; rules logic that needs printed values (mana cost, owner) keeps reading `Card` directly.
- **Static abilities declared on `Spec.Static`** — new field on the existing S14 catalog spec, parallel to `OnETB`, `OnResolve`, `ManaAbilities`. Each entry is `{Layer, SubLayer, AppliesTo func, Apply func}`. Registered through the existing `Register()` flow; consumed via a 6th function-var hook (`CatalogStaticAbilities`) following the cycle-break pattern S14 established.
- **Counter math stays in `CurrentPower()` / `CurrentToughness()`.** Layer 7d _calls_ them rather than subsuming them — keeps S13.2's SBA suite green; minimal blast radius. Layer engine starts with anthems + CDAs + types; counter integration is one delegating Apply function in 7d.
- **Skip dependency detection (CR 613.8).** Pure timestamp ordering works for ~95% of real cards. Opalescence + Humility is the canonical pathological case and never shows up in casual EDH. S16.5 follow-up if a real card surfaces.
- **Layer 7 ships full sub-layer support** (7a CDA, 7b setting, 7c modifying, 7d counters, 7e switch) — Tarmogoyf needs 7a, anthems need 7c, counters need 7d. 7b and 7e ship as no-op-capable stubs.
- **Battlefield-only** — CR 113.6 default. Continuous effects from non-battlefield zones (Yixlid Jailer in graveyard) deferred. Niche cards opt in later.
- **`EnteredBattlefieldAt int64`** stamped on every battlefield-entry path. Drives layer ordering. New helper `g.stampBattlefieldEntryLocked(cardID)` called by every existing entry site (cast-spell land branch, ETB resolution, MoveCard).
- **Replace S15 commander-identity proxy.** `commanderIdentityFor` (mutations.go:2089) hand-rolled `distinctColorsInManaCost` was the explicit S15→S16 hand-off. Now reads the commander's effective characteristic colors. Proxy stays as fallback for placeholder commanders the catalog doesn't know.

### Tasks

**Server — characteristic snapshot:**

- [x] New [server/internal/game/characteristic.go](../server/internal/game/characteristic.go) with `Characteristic{Power, Toughness, Loyalty, Types, Subtypes, Supertypes, Colors, Abilities, Name}`. Field choice mirrors what `viewOfCard` projects to the wire.
- [x] `Card.printedCharacteristic() Characteristic` reads immutable printed fields off existing struct (no new storage on `Card` for printed; printed values ARE the existing fields).
- [x] `Card.effective` private cache field (nil-able). `Card.Effective()` returns it or falls back to printed when nil.

**Server — layer engine:**

- [x] New [server/internal/game/layers.go](../server/internal/game/layers.go) with `Layer`/`SubLayer` enums, `ContinuousEffect` interface (`Layer`, `Timestamp`, `AppliesTo`, `Apply`).
- [x] `Game.LayerVersion uint64` + `Game.LastResolvedVersion uint64`.
- [x] `g.RecomputeLayersLocked()` — reset effective = printed; collect active continuous effects; for each layer in 1..7 (sub-layers in 7a..7e for Layer7PT): filter, sort by timestamp, apply.
- [x] `g.RecomputeLayersIfStaleLocked()` — fast-path no-op when `LayerVersion == LastResolvedVersion`. Called from snapshot path.
- [x] Layer 7d Apply delegates to `CurrentPower()` / `CurrentToughness()` for counter math.

**Server — `Spec.Static` + catalog hook:**

- [x] `Spec.Static []StaticAbility` field on [server/internal/cards/effects/spec.go](../server/internal/cards/effects/spec.go). `StaticAbility{Layer, SubLayer, AppliesTo func, Apply func}`.
- [x] `CatalogStaticAbilities func(oracleID string) []StaticAbility` — sixth function-var hook in [game/effect_hooks.go](../server/internal/game/effect_hooks.go); populated by [effects/wire.go](../server/internal/cards/effects/wire.go).
- [x] `g.activeStaticAbilitiesLocked()` — walks battlefield, looks each card's oracle ID up via the hook, adapts `StaticAbility` into a `ContinuousEffect` bound to source card's `EnteredBattlefieldAt` timestamp.
- [x] `Card.EnteredBattlefieldAt int64` (Unix-nano timestamp) + `g.stampBattlefieldEntryLocked(cardID)` helper called by every existing battlefield-entry site.

**Server — event hooks + listeners:**

- [x] **Audit + emit `EventLTB`** on every battlefield-leave path. Mirrors existing `EventETB` emission at [mutations.go:336-341](../server/internal/game/mutations.go#L336-L341). Today only ETB fires; layer system needs LTB to invalidate.
- [x] Built-in listener: `EventETB`, `EventLTB`, `EventCounterPlaced` (battlefield card), `EventControlChanged` (S16-new), step advance — all bump `g.LayerVersion`. Registered at game-start.
- [x] Snapshot path: `ReadSnapshot` calls `g.RecomputeLayersIfStaleLocked()` inside the read closure before building views. Recompute uses a separate `sync.Mutex` so the read lock isn't promoted; double-check version after acquiring the recompute mutex to avoid duplicate work. *(Superseded in S24 — that mutex did not exclude read-lock holders and the recompute raced them. `ReadSnapshot` now upgrades to the write lock; see the amendment to [ADR 0012](decisions/0012-layer-system.md) decision 11.)*

**Server — wire projection:**

- [x] `viewOfCard` ([protocol/view.go:954-1001](../server/internal/protocol/view.go#L954-L1001)) reads from `c.Effective()` for `Power`, `Toughness`, `TypeLine`, `Colors`, and the new `abilities []string` field.
- [x] `redactCardForViewer` ([view.go:899-914](../server/internal/protocol/view.go#L899-L914)) — same redaction shape (fields are the same; values are now effective).
- [x] `CardView.abilities []string` new wire field — drives S18's keyword renderer; ships now so S18 can plug in without a wire bump.
- [x] **No protocol bump.** `power`/`toughness`/`type_line` keep the same key names; values become effective.

**Server — replace S15 commander-identity proxy:**

- [x] `commanderIdentityFor` ([mutations.go:2089-2131](../server/internal/game/mutations.go#L2089-L2131)) replaced with layer-aware computation: read commander's effective characteristic colors. `distinctColorsInManaCost` proxy stays as fallback for lobby-seeded placeholder commanders.
- [x] Existing S15 tests (Arcane Signet commander-identity filtering) stay green — layer system produces same identity for test commanders.

**Catalog cards (4):**

- [x] **glorious_anthem.go** — Layer 7c. `AppliesTo`: target.IsCreature() && target.Controller == source.Controller. `Apply`: `c.Power++; c.Toughness++`.
- [x] **mycosynth_lattice.go** — Layer 4. `AppliesTo`: every permanent. `Apply`: append `"Artifact"` to `c.Types` (idempotent). Lattice's other clauses (lands tap for any color, no mana ability adds non-colorless) are S17 territory — out of scope for S16.
- [x] **lord_of_atlantis.go** — Two static abilities on one card:
  - **7c**: AppliesTo = `target.IsCreature() && hasSubtype("Merfolk") && target != source && target.Controller == source.Controller`. Apply: +1/+1.
  - **6**: AppliesTo same predicate. Apply: append `"flying"` and `"islandwalk"` to `c.Abilities`. Behavior of flying/islandwalk lands with S18 — S16 just exposes the keyword grant on the wire.
- [x] **tarmogoyf.go** — Layer 7a CDA. AppliesTo: target == source. Apply: `n := distinctCardTypesInAllGraveyards(g); c.Power = n; c.Toughness = n + 1`. Helper unions `Types` across every graveyard zone.

**Tests:**

- [x] `server/internal/game/layers_test.go` (new):
  - `TestSingleAnthemAddsPlusOne`, `TestTwoAnthemsStack`, `TestAnthemRespectsControllerBoundary`, `TestRemoveAnthemRevertsCreatures`.
  - `TestMycosynthLatticeAddsArtifact`, `TestLordOfAtlantisGrantsFlyingToMerfolkOnly`, `TestLordOfAtlantisDoesNotPlusItself`.
  - `TestTarmogoyfCDA` (4 distinct types in graveyards → 4/5).
  - `TestCounterAndAnthemStack` (2/2 + +1/+1 counter + anthem → 4/4 — counter via 7d delegating to CurrentPower; anthem via 7c).
- [x] `TestLayerVersionInvalidationFastPath` — two consecutive snapshot reads with no events between invoke `RecomputeLayersLocked` exactly once (instrumentation counter on `Game`).
- [x] `TestEventLTBFiresOnZoneExit` — regression guard for the LTB emission audit.
- [x] Snapshot regression: `viewOfCard` for 2/2 + anthem → wire `power=3 toughness=3`; round-trip through `ViewOfGameFor` confirms redaction doesn't strip effective values for knowers.
- [x] S15 commander-identity regression stays green after proxy replacement.

**Docs:**

- [x] [docs/decisions/0012-layer-system.md](decisions/0012-layer-system.md) — ADR. Topics: recompute-from-scratch + version cache vs incremental diffs, timestamp-only ordering before dependency detection, Layer 7d delegating to `CurrentPower`/`CurrentToughness`, Layer 5 stub with no card, Layer 1/3 deferred, `Spec.Static` declarative-DSL consistency, commander-identity proxy replacement.
- [x] [docs/protocol.md](protocol.md) — note `CardView.power`, `toughness`, `type_line`, `abilities` are now _effective_ (post-layer); document new `EventLTB` event kind.
- [x] [docs/sprints.md](sprints.md) S16 section — this expansion lives here.
- [x] [AGENTS.md](../AGENTS.md) §7 — extend catalog-card recipe with "Adding a static ability" sub-recipe (parallel to S15 mana-ability recipe), worked example using Glorious Anthem.

### Out of scope (explicit handoffs)

- **Dependency detection (CR 613.8)** — Opalescence + Humility pathological case. S16.5 follow-up if a real card surfaces. **Still deferred after S16.5** — the trigger has not fired and no pair of statics in the catalog can produce an observable dependency; see [ADR 0043](decisions/0043-copy-effects.md) §5 for the reasoning and the cost. **Corrected 2026-09-16:** the "no pair" claim is false. Layer-4 pairs have reproduced on develop since 2026-09-13 (Urborg, Tomb of Yawgmoth + Song of the Dryads, Maskwood Nexus + a crewed Vehicle, and three more; listed in ADR 0043 §5's amendment). Dependency ordering is tracked in [#668](https://github.com/krakenhavoc/cmd_and_ctrl/issues/668).
- **Layer 1 copy effects** — Clone, Phyrexian Metamorph, Spark Double. Defer to S16.5. **Landed in S16.5** ([#159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/159), closes [#335](https://github.com/krakenhavoc/cmd_and_ctrl/issues/335)) — as a copiable-value baseline rewrite rather than a `Layer1Copy` `ContinuousEffect`; see [ADR 0043](decisions/0043-copy-effects.md).
- **Layer 3 text-changing effects** — Mind Bend, Glamerdye. Engine ships layer 3 stub; no execution path.
- **Layer 5 color-changing effects** — Painter's Servant. Engine ships the layer 5 stub with no card.
- **Continuous effects from non-battlefield zones** — Yixlid Jailer, command-zone commander effects. Battlefield-only is the S16 simplification (CR 113.6 default).
- **Aura / Equipment attachment infrastructure** — Mind Control needs the aura's controller to grant control over the enchanted creature, requires aura-attaching state the engine doesn't model. Bundled with S17.
- **Mycosynth Lattice's other clauses** — "lands tap for any color" + "no land's mana ability adds non-colorless." Defer to S17 (S15 mana-ability override needed).
- **Combat keyword behavior** — flying / islandwalk / first strike / trample / vigilance show on the wire after S16 but combat behavior lands with S18. Explicit S18 hand-off.
- **Continuous activated abilities** — non-mana activated abilities (planeswalker +1/-1, equip, cycling) are S19. S16 is static-ability-only.

### Risks / gotchas

- **`EventLTB` audit miss.** Layer system correctness depends on LTB firing every battlefield-leave. Mitigation: grep audit every `g.Battlefield` removal site; regression test walks every public battlefield-leave mutation and asserts an `EventLTB`.
- **Recompute under SBA loop.** SBA mutations emit zone-change events that bump `LayerVersion`. Recompute must NOT run inside the SBA loop (would interleave with SBA evaluation). Mitigation: only call `RecomputeLayersIfStaleLocked` inside `ReadSnapshot`'s read closure — never inside a write mutation.
- **Layer 7d counter integration via `CurrentPower`.** 7d's Apply reads `c.CurrentPower() - c.Power` to derive counter delta. With a 7c anthem + 7d counter delta, effective power == `printed + 1 + counter_delta` — matches CR 613.5 layering. Test: `TestCounterAndAnthemStack`.
- **Tarmogoyf CDA timestamp.** CDAs are layer 7a — they SET P/T before any +1/+1 (7c) or counter (7d) apply. Tarmogoyf with a +1/+1 counter and Glorious Anthem is `(N+2) / (N+3)` where N = distinct types in graveyards. The layer engine must apply 7a's "set" before 7c's "modify" arithmetic.
- **`Card.EnteredBattlefieldAt` not stamped on test fixtures.** `pushBattlefieldForTest` bypasses cast/move paths. Mitigation: extend the helper to stamp the field.
- **`Game.LayerVersion` race over snapshot reads.** Two competing readers might both see "stale" and try to recompute. Mitigation: separate `sync.Mutex` for the recompute, double-check version after acquiring (avoid promoting the read lock).
- **No protocol bump but a behavior change.** Pre-S16 replays decoded with post-S16 servers re-resolve all layers from scratch. No-op for legacy replays (no static abilities ⇒ effective == printed). Confirm via replay round-trip test.

### Exit criteria

1. **Anthem buffs creatures on the wire.** Cast Glorious Anthem; every controlled battlefield creature shows `power+1`/`toughness+1` in the next snapshot.
2. **Removing the anthem reverts.** Move Glorious Anthem to graveyard; next snapshot shows printed P/T.
3. **Anthems stack.** Two anthems → +2/+2.
4. **Mycosynth Lattice adds artifact.** Cast Lattice; a Forest's wire `type_line` includes `"Artifact"`.
5. **Lord of Atlantis grants flying.** Lord + one Merfolk + one non-Merfolk; only the Merfolk's wire `abilities` contains `"flying"`.
6. **Lord doesn't pump itself.** Single Lord on board with no other Merfolk shows printed 2/2.
7. **Tarmogoyf CDA reflects graveyards.** Tarmogoyf + 4 distinct card types across graveyards → wire 4/5; mill a 5th type → 5/6.
8. **Counter + anthem stack correctly.** 2/2 with a +1/+1 counter + 1 anthem → wire 4/4 (counter via 7d, anthem via 7c).
9. **No regression in S15 commander identity.** Arcane Signet filtering tests stay green after proxy replacement.
10. **Snapshot fast-path is free.** Two consecutive snapshot reads with no intervening mutation invoke `RecomputeLayersLocked` exactly once (instrumentation counter confirms).
11. **Replay round-trip clean.** Pre-S16 replay JSONL decodes and re-snapshots with effective == printed for every card.
12. **ADR 0012 + protocol/AGENTS docs + sprint flip** all land in the docs sub-PR.

### Sub-PR split

Stop-and-show for manual testing at each boundary, mirroring S15.

1. **Sprint plan expansion (sub-PR 0).** This expansion in `docs/sprints.md` + GitHub issue #66 body. Pure docs.
2. **Characteristic + layer-engine skeleton (sub-PR 1).** `Characteristic`, `Card.effective` cache, `Card.Effective()`, `g.LayerVersion`, `g.RecomputeLayersLocked` (no-op pass), `g.RecomputeLayersIfStaleLocked`, snapshot integration. **No static abilities yet** — engine runs but does nothing observable. Stop-and-show: server tests prove `viewOfCard` reads from `Effective()`; a 2/2 stays 2/2.
3. **`EventLTB` + listener registration (sub-PR 2).** Audit + emit `EventLTB` on every battlefield-leave path; built-in listener bumps `LayerVersion`. Stop-and-show: zone-leave → events log shows `EventLTB`; `LayerVersion` ticks.
4. **`Spec.Static` + `CatalogStaticAbilities` hook + Glorious Anthem (sub-PR 3).** Adds declaration shape, the 6th function-var hook, the first card. Layer engine actually runs effects. Stop-and-show: cast Anthem → all your creatures gain +1/+1.
5. **Mycosynth Lattice + Lord of Atlantis (sub-PR 4).** Layer 4 + Layer 6 cards. Stop-and-show: cast each, verify type / ability grant lands on the wire; Lord doesn't pump itself.
6. **Tarmogoyf CDA + commander-identity proxy replacement (sub-PR 5).** Layer 7a + S15 hand-off. Stop-and-show: 4 card types in graveyards → Tarmogoyf 4/5; Arcane Signet filtering still works.
7. **ADR 0012 + protocol + AGENTS recipe + sprint flip (sub-PR 6).** Docs leg, mirrors S15's sub-PR 6 shape.

### Critical files

- **Server new:** [server/internal/game/characteristic.go](../server/internal/game/characteristic.go), [server/internal/game/layers.go](../server/internal/game/layers.go), [server/internal/game/layers_test.go](../server/internal/game/layers_test.go), [glorious_anthem.go](../server/internal/cards/effects/glorious_anthem.go), [mycosynth_lattice.go](../server/internal/cards/effects/mycosynth_lattice.go), [lord_of_atlantis.go](../server/internal/cards/effects/lord_of_atlantis.go), [tarmogoyf.go](../server/internal/cards/effects/tarmogoyf.go).
- **Server modified:** [card.go](../server/internal/game/card.go) (`effective` cache, `Effective()`, `EnteredBattlefieldAt`), [game.go](../server/internal/game/game.go) (`LayerVersion`, `LastResolvedVersion`, recompute mutex), [effect_hooks.go](../server/internal/game/effect_hooks.go) (6th hook), [events.go](../server/internal/game/events.go) (`EventLTB` if missing), [listeners.go](../server/internal/game/listeners.go) (built-in version-bump), [mutations.go](../server/internal/game/mutations.go) (LTB audit, entry timestamp helper, `commanderIdentityFor` replacement), [spec.go](../server/internal/cards/effects/spec.go) (`Static` field), [wire.go](../server/internal/cards/effects/wire.go) (populate hook), [view.go](../server/internal/protocol/view.go) (read from `Effective()`, new `abilities` field).
- **Client (small):** [protocol.ts](../client/src/lib/protocol.ts) (mirror new `CardView.abilities` for S18 readiness).
- **Docs:** new [docs/decisions/0012-layer-system.md](decisions/0012-layer-system.md), [docs/protocol.md](protocol.md), [docs/sprints.md](sprints.md), [AGENTS.md](../AGENTS.md) §7.

### Branch + commit conventions

- Branches: `feat/s16-sprint-plan-expansion`, `feat/s16-layer-engine`, `feat/s16-event-ltb`, `feat/s16-anthem`, `feat/s16-type-and-ability-grant`, `feat/s16-cda-and-identity`, `feat/s16-docs-and-sprint-flip`.
- Footer:
  ```
  Sprint: S16 — Continuous effects + layer system (CR 613)
  Issue: #66
  ```

---

## S16.5 — Layer system follow-ups (rolling)

**Phase:** 7 · **Goal:** land the two things [ADR 0012](decisions/0012-layer-system.md) deferred, when and only when a real card asks for them. Tracked at [#159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/159); no deadline, rolling on demand.

### Layer 1 copy effects ✅

Triggered by a Clone-class card reaching the catalog — and by [#335](https://github.com/krakenhavoc/cmd_and_ctrl/issues/335), where Clone "does not allow the selection of any usable target and enters the battlefield as a 0/0". Both halves of that report are one missing feature: the CR 614 pipeline could not ask a question as a permanent entered, and the layer engine could not say "this permanent's printed values are somebody else's".

- [x] `game.PrintedValues` + `Card.PrintedSelf` — the CR 707.2 copiable-value set, and the revert the battlefield-leave path restores (CR 400.7). Carried by the snapshot, deep-copied by `clone.go`.
- [x] `ReplacementEffect.CopySelector` + `PendingChoiceCopyTarget` — an as-it-enters picker, answered with the shared `{choice_id, card_ids}` payload; an empty list declines. The copy lands **before** `EventETB`, so every ETB trigger sees the copied characteristics.
- [x] Stack resolution is now an `entryResumable` entry site, carrying its `StackItem` so the Aura attach (CR 303.4a) and evoke's sacrifice trigger (CR 702.74a) survive the pause. This also fixed a latent bug nothing had hit: **any** permanent spell whose entry queued a CR 616 ordering prompt was dropped instead of pushed.
- [x] Catalog: Clone, Phyrexian Metamorph, Spark Double, Sakashima the Impostor — between them the three kinds of "except" clause (add a type, override a name, remove a supertype + change how it enters).
- [x] [ADR 0043](decisions/0043-copy-effects.md), AGENTS.md §7 "Adding a copy effect", client `copy_target` prompt branch.

**Declared gaps:** duration-scoped copy effects (Mirage Mirror, Cytoshape) — the `Layer1Copy` bucket survives for them; an "except" clause that GRANTS an ability (Sakashima's own return-to-hand ability); a declined Clone surviving as a 0/0, which is the engine's pre-existing printed-0-toughness SBA convention, not a copy bug.

### Dependency detection (CR 613.8) — not built; moved to [#668](https://github.com/krakenhavoc/cmd_and_ctrl/issues/668)

*As written when #414 merged:* the trigger #159 named for this half — a playtester hitting an Opalescence + Humility-class pathology — has not fired, and no pair of statics in the catalog can produce one: dependency is only observable when one static changes whether another static APPLIES, and every static in the catalog is an anthem, a keyword grant, a type-add or a CDA, all commutative under timestamp order. The cost is a required dependency declaration on every `StaticAbility` present and future, plus a topological sort with cycle detection on the hottest path in the engine. Reasoning in full at [ADR 0043](decisions/0043-copy-effects.md) §5. The trigger stays armed.

**Corrected 2026-09-16: that paragraph's premise is false.** Type-adds don't commute with type-sets when the add's "applies to" reads the type the set writes. The catalog has had such pairs since a few hours after #414 merged, when Maskwood Nexus (#404) and Arixmethes (#480) arrived, and Song of the Dryads (#563) followed on 2026-09-14. Reproduced on develop `7c1ae9b`, all in layer 4, each coming out wrong in one entry order: Urborg, Tomb of Yawgmoth + Song of the Dryads; Urborg + Arixmethes, Slumbering Isle; Maskwood Nexus + any Vehicle crewed after it entered (the usual order, since crew is stamped when it resolves); Maskwood Nexus + The Warring Triad; Maskwood Nexus + a slumbering Arixmethes. The trigger fired through catalog pairs, not a table report. The consequences are wrong mana and wrong tribal pumps, not Opalescence + Humility loops. The engine also already resolves one CR 613.8 case, ability removal, by iterating to a fixed point ([ADR 0046](decisions/0046-layer-6-authoritative.md) §4), and part of that is itself wrong. The details are in the dated notes on ADR 0043 §5, ADR 0012 §4 and ADR 0046 §4-5.

**Status: done.** The copy-effects half shipped (#414). The dependency half did **not** ship, and it no longer belongs to this rolling sprint: it needs an ADR and a change to the hottest path in the engine, not an on-demand follow-up. [#159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/159) closes with this reconciliation, and the tail has its own issues. Until they land, AGENTS.md §7 "When NOT to add a catalog entry" holds back the cards that would add more pairs.

**What is left, and where it went** (decided 2026-09-16):

- [#668](https://github.com/krakenhavoc/cmd_and_ctrl/issues/668): ADR, then CR 613.8 dependency ordering in layer 4. Detection is judged on each effect's instruction rather than by diffing outputs, with CR 613.8c re-checks. Blocks Arcane Adaptation, Leyline of Transformation, Encroaching Mycosynth, Yavimaya, Cradle of Growth and Prismatic Omen.
- [#669](https://github.com/krakenhavoc/cmd_and_ctrl/issues/669): ability removal drops a silenced source from every layer (CR 613.6), and Song of the Dryads wipes abilities that other effects granted (CR 305.7). Comes before #668, and blocks Magus of the Moon.
- [#670](https://github.com/krakenhavoc/cmd_and_ctrl/issues/670): "every creature type" is stored as a keyword, so layer-6 removal wipes a layer-4 grant (Maskwood Nexus). A storage bug that belongs to the tribal area, not a dependency.
- [#675](https://github.com/krakenhavoc/cmd_and_ctrl/issues/675), found at closeout: CR 704.5p isn't implemented, so an Aura or Equipment that Song of the Dryads turned into a Forest stays attached and works again once the Song leaves. Today only the gather-time silence hides it, so it has to land before #669.
- Open PR [#644](https://github.com/krakenhavoc/cmd_and_ctrl/pull/644) declares the five pairs and the CR 305.7 gap as caveats on Urborg, Song of the Dryads, Arixmethes, Maskwood Nexus and The Warring Triad, and pins each as a skipped test for #668 and #669 to un-skip.

---

## S17 — Replacement effects engine (CR 614)

**Phase:** 7 · **Goal:** ship a pre-event replacement pipeline (CR 614/616) layered onto the five rules-visible mutation functions, plus step-transition hooks for skip-step. `ReplacementEvent` tagged-union value type is mutated or canceled by registered `ReplacementEffect`s declared on `Spec.Replacements`. CR 616.1 iterative apply-loop, CR 614.5 once-per-event tracking, CR 616 affected-player-chooses-order, and CR 614.10 "may" optional replacements all enforced centrally so card files stay declarative. S13.1's hand-rolled commander-zone replacement is refactored into a built-in that fires for EVERY commander move (spell-driven, SBA-driven, admin-driven) — closing a pre-existing gap.

The second-most structural rules subsystem after the stack. Unblocks S18 combat keywords (replaceable triggers), S21 tokens, S22 draw manipulation, and S30 damage prevention. After this sprint, the catalog can describe cards whose effects intercept events _before_ they happen.

**Shipped 2026-04-23** via PRs [#158](https://github.com/krakenhavoc/cmd_and_ctrl/pull/158), [#163](https://github.com/krakenhavoc/cmd_and_ctrl/pull/163), [#172](https://github.com/krakenhavoc/cmd_and_ctrl/pull/172), [#166](https://github.com/krakenhavoc/cmd_and_ctrl/pull/166), [#167](https://github.com/krakenhavoc/cmd_and_ctrl/pull/167), [#171](https://github.com/krakenhavoc/cmd_and_ctrl/pull/171). Closes [#67](https://github.com/krakenhavoc/cmd_and_ctrl/issues/67) and [#164](https://github.com/krakenhavoc/cmd_and_ctrl/issues/164).

**Catalog cards shipped (6 new + 3 finishers):** Doubling Season, Hardened Scales, Branching Evolution, Stasis, Kismet, Fog. Plus TappedOnEntry finishers for Cultivate / Path to Exile / Solemn Simulacrum.

### Architectural decisions

See [ADR 0013](decisions/0013-replacement-effects.md). Abbreviated:

- **Pre-event pipeline with a tagged-union value.** `ReplacementEvent` carries per-kind mutable payload (draw / move / counter / life / damage / step). The five mutation functions construct one, call `applyReplacementsLocked`, and branch on the result.
- **`ReplacementEffect` is a declarative struct** — `{Watches, AppliesTo, Replace, Controller, SelfReplacement, Label}` — sibling of S16's `StaticAbility`. Registered via `Spec.Replacements`, consumed through a 7th function-var hook `CatalogReplacements`.
- **CR 616.1 iterative apply-loop, centrally enforced.** Engine owns iteration, the once-per-event map, and the CR 616 ordering prompt. Cards declare `AppliesTo` + `Replace`.
- **Once-per-event tracking applies to all fired replacements**, not only `SelfReplacement`-flagged ones. Per-call scope, keyed on `(ReplacementEventID, ReplacementEffectID)`.
- **CR 616 order-choose via new `replacement_order` PendingChoice kind.** Queue-and-return pattern reused from S13/S15 (synchronous, no goroutines). Wire surface is additive — no protocol bump.
- **CR 614.10 optional replacements** via new `optional_replacement` PendingChoice kind (yes/no prompt). Added mid-sprint in [sub-PR 6](#s17-sub-pr-6) after manual testing surfaced the CR 903.9 auto-route gap. `ReplacementEffect.Optional bool` flags effects that need the owner's opt-in before firing.
- **Six integration points**: `AddCounter`, `MoveCardByID`, `DrawCard`, `ChangePlayerLife`, `MarkDamage`, plus `runStepEntryHooksLocked` for skip-step. Parallel `*ForEffect` hooks in `effect_api.go` so catalog-driven mutations route through replacements too.
- **Battlefield-entry pipeline routing** lives inline at each entry site (cast resolution, land play, `MoveCardByIDAsCommander`, `SearchLibrary` battlefield branch). The planned `enterBattlefieldLocked` shared helper refactor was scoped out in favor of targeted pipeline calls at each site — less risk, same behavior.
- **S13.1 commander-zone replacement refactored** into a built-in in `builtin_replacements.go`, registered at `NewGame`. Bespoke `applyCommanderZoneReplacementLocked` deleted. Sub-PR 6 widened the `AppliesTo` (dropped the `asCommanderMove` gate) and flipped the effect to `Optional=true` so commanders dying to spells / SBAs / wrath now prompt the owner instead of going to graveyard silently.
- **Fetched-land enters-tapped uses a primitive flag**, not the generic pipeline. `SearchLibrary.TappedOnEntry` closes the Cultivate / Path to Exile / Solemn Simulacrum deferrals cleanly.
- **`routeBattlefieldCardToOwnerGraveyardLocked` routes through the pipeline** so SBA deaths / wrath destroys fire the commander-zone replacement. A new `executeBattlefieldLeaveLocked` helper runs the physical move post-pipeline.
- **Damage prevention hook only**; Fog ships as a turn-scoped atomic cancel via `Game.TurnScopedReplacements` cleared at `StepCleanup`. Full shield mechanic with charges stays in S30.

### Tasks

**Server — engine core:**

- [x] New `server/internal/game/replacements.go` — `ReplacementEffect`, `ReplacementEvent`, `ReplacementEventKind`, `ReplacementEventID`, `ReplacementEffectID`, `applyReplacementsLocked`, iterative apply-loop with per-call once-per-event map.
- [x] New `server/internal/game/builtin_replacements.go` — `commanderZoneReplacement` built-in; `Game.BuiltinReplacements []ReplacementEffect` field populated at `NewGame`.
- [x] `server/internal/game/pending_choice.go` — `PendingChoiceReplacementOrder` kind + `ReplacementEffectIDs` field + server-only `replacementResume` frame + `ResolveReplacementOrder` method.
- [x] `server/internal/game/pending_choice.go` — `PendingChoiceOptionalReplacement` kind + `ResolveOptionalReplacement` method (sub-PR 6, CR 614.10 yes/no).
- [x] `server/internal/game/effect_hooks.go` — 7th function-var hook `CatalogReplacements`.
- [x] `server/internal/cards/effects/spec.go` — `Replacements []game.ReplacementEffect` field.
- [x] `server/internal/cards/effects/wire.go` — populate `game.CatalogReplacements`.
- [x] `server/internal/game/game.go` — step-transition hook at top of `runStepEntryHooksLocked`, `TurnScopedReplacements` cleared at `StepCleanup`.

**Server — pipeline integration:**

- [x] `drawCardLocked` + `MoveCardByIDAsCommander` + `ChangePlayerLife` + `MarkDamage` (+ new `MarkCombatDamage` wrapper) + `AddCounter` — construct `ReplacementEvent`, call pipeline, branch.
- [x] Parallel hooks in `effect_api.go`: `AddCounterForEffect`.
- [x] Inline pipeline call at each battlefield-entry site (cast resolve, land play). Shared-helper refactor deferred — targeted per-site calls were lower risk.
- [x] `routeBattlefieldCardToOwnerGraveyardLocked` routes through the pipeline + new `executeBattlefieldLeaveLocked` for the physical move (sub-PR 6).
- [x] Delete `applyCommanderZoneReplacementLocked`; built-in takes over (sub-PR 2).
- [x] Widen commander-zone built-in AppliesTo + flip to `Optional=true` (sub-PR 6). Fires on every commander move, not just admin-flagged.
- [x] `SearchLibrary` primitive gains `TappedOnEntry bool`.
- [x] Cast-time target validation: `CastSpell` rejects `ErrInvalidParam` when `target_mode` is non-empty and `params.Targets` is empty (sub-PR 6, guards against client-side targeting skip).

**Server — wire + dispatcher:**

- [x] `server/internal/protocol/view.go` — `ReplacementOptionView{id, label, source_card_id}` + `PendingChoiceView.ReplacementOptions`.
- [x] `server/internal/actions/actions.go` — `TypeResolveChoice` dispatcher legs for `order []string` → `ResolveReplacementOrder` and `apply bool` → `ResolveOptionalReplacement`.
- [x] `viewOfCard` sends `CurrentPower()` / `CurrentToughness()` so the wire P/T includes the +1/+1 counter delta (previously missed; on-card pip showed printed P/T after counters landed).

**Client:**

- [x] `client/src/lib/components/board/ChoicePromptModal.svelte` — `replacement_order` click-to-order branch.
- [x] `client/src/lib/components/board/ChoicePromptModal.svelte` — `optional_replacement` yes/no branch (sub-PR 6).
- [x] `client/src/lib/protocol.ts` — mirror `PendingChoiceView.replacement_options`, `ReplacementOptionView`, `optional_replacement` kind.
- [x] `CounterPips.svelte` — abbr + tabular count rendering; fixed sub-pixel-ambiguous `·` separator.
- [x] Shift+click battlefield card → add +1/+1 counter (debug affordance until right-click admin menu from [#170](https://github.com/krakenhavoc/cmd_and_ctrl/issues/170) ships).
- [x] `Game.svelte.sendAction` stashes last cast_spell payload per card so `castAnyway` / `confirmAutoTap` retries replay original `targets` / modes / X / distribution (bug fix during sub-PR 6 manual testing).

**Catalog cards shipped (6 new + 3 finishers):**

- [x] **doubling_season.go** — RepEventCounter replacement; `ev.CounterDelta *= 2` for any counter placed on a permanent you control.
- [x] **hardened_scales.go** — RepEventCounter replacement; `+1/+1` on creatures you control; `ev.CounterDelta += 1`.
- [x] **branching_evolution.go** — same predicate shape as Hardened Scales but `*= 2`. Narrower Doubling Season.
- [x] **stasis.go** — RepEventStepTransition replacement; cancels `StepUntap` for any seat. Upkeep-sacrifice trigger deferred to S19.
- [x] **kismet.go** — RepEventMove replacement; opponents' creatures / artifacts / lands enter tapped.
- [x] **fog.go** — turn-scoped RepEventDamage replacement; cancels combat damage only (gated on `IsCombatDamage`).
- [x] **cultivate.go** — `TappedOnEntry = true` on the battlefield fetch.
- [x] **path_to_exile.go** — `TappedOnEntry = true` on the basic-land fetch.
- [x] **solemn_simulacrum.go** — `TappedOnEntry = true` on the ETB fetch.

**Deferred cards (not shipped in S17):**

- Hangarback Walker — needs X-cost stack plumbing; future sub-PR.
- Champion of Lambholt — both halves are triggered/static, not replacements; re-homed to S19 (triggers) + S18 (block-restriction). Dropped from S17.
- Library of Leng — cleanup-discard picker UI has its own design surface; re-scoped to a later sub-PR. Gap tracked as [#160](https://github.com/krakenhavoc/cmd_and_ctrl/issues/160) for the voluntariness nuance. *Corrected 2026-09-16:* there is no voluntariness nuance. Leng replaces effect-caused discards only, never the cleanup discard, and #160 closes as not planned. Leng is still not in the catalog; it waits on [#650](https://github.com/krakenhavoc/cmd_and_ctrl/issues/650) and [#651](https://github.com/krakenhavoc/cmd_and_ctrl/issues/651) ([ADR 0013 §10a](decisions/0013-replacement-effects.md)).
- Mycosynth Lattice's "lands tap for any color" + "no mana ability adds non-colorless" clauses — sub-PR 6 pivoted to the CR 903.9 commander-zone fix instead. Lattice clauses re-homed to S18 (mana-ability rewrite) per [#68](https://github.com/krakenhavoc/cmd_and_ctrl/issues/68).

**Tests:**

- [x] `server/internal/game/replacements_test.go` — zero-replacement passthrough; commander-zone built-in routes to command zone on yes, graveyard on no; CR 903.9 SBA death queues prompt; once-per-event tracking; iteration cap; `ResolveReplacementOrder` permutation validation.
- [x] `server/internal/cards/effects/doubling_season_test.go` — sprint exit criterion: `[HS, DS]` → 4 counters, `[DS, HS]` → 3 counters. Three-replacement single-prompt (no re-prompting after applying each).
- [x] `server/internal/cards/effects/stasis_kismet_test.go` — Stasis skip; Kismet opponent-only; fetched-land TappedOnEntry for Cultivate + Path.
- [x] `server/internal/cards/effects/fog_test.go` — combat damage canceled; non-combat damage unaffected; turn-scoped clear at cleanup.

**Docs:**

- [x] [docs/decisions/0013-replacement-effects.md](decisions/0013-replacement-effects.md) — ADR 0013 shipped with sub-PR 1.
- [x] [docs/sprints.md](sprints.md) S17 section — this entry.
- [x] [AGENTS.md](../AGENTS.md) §7 — "Adding a replacement effect (S17+)" subsection.
- [ ] [docs/protocol.md](protocol.md) — document `replacement_order` + `optional_replacement` PendingChoice kinds. Rolling follow-up.

**Tracking discipline (sub-PR 1 + 6):**

- [x] GitHub issue updates: [#67](https://github.com/krakenhavoc/cmd_and_ctrl/issues/67) (S17), [#68](https://github.com/krakenhavoc/cmd_and_ctrl/issues/68) (S18), [#76](https://github.com/krakenhavoc/cmd_and_ctrl/issues/76) (S24), [#93](https://github.com/krakenhavoc/cmd_and_ctrl/issues/93) (S28), [#95](https://github.com/krakenhavoc/cmd_and_ctrl/issues/95) (S30).
- [x] S16.5 milestone + tracking issue [#159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/159).
- [x] Library of Leng voluntariness follow-up [#160](https://github.com/krakenhavoc/cmd_and_ctrl/issues/160).
- [x] CR 903.9 commander-zone gap [#164](https://github.com/krakenhavoc/cmd_and_ctrl/issues/164) — filed + closed by sub-PR 6.
- [x] Admin context-menu mini-sprint [#170](https://github.com/krakenhavoc/cmd_and_ctrl/issues/170) — filed; shift+click counter is the stopgap.

### Out of scope (explicit handoffs)

- **Aura / Equipment attachment infrastructure** (Mind Control) → **S24** [#76](https://github.com/krakenhavoc/cmd_and_ctrl/issues/76). S16 doc bundled this into S17; re-homed for aura + combat-state coupling.
- **Cost-replacement effects** (Trinisphere, Thalia, Spellshift, Kambal) → **S28** [#93](https://github.com/krakenhavoc/cmd_and_ctrl/issues/93). S17 hooks touch the event-path; cost replacement touches the S15 cost engine — separate surface.
- **Damage prevention shields with charges** (CR 615) → **S30** [#95](https://github.com/krakenhavoc/cmd_and_ctrl/issues/95). Fog in S17 is atomic cancel; stateful shields land in S30.
- **Dependency detection** (CR 613.8) + **Layer 1 copy effects** → **S16.5** [#159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/159). *(2026-09-16: copy effects shipped in #414. Dependency ordering moved to [#668](https://github.com/krakenhavoc/cmd_and_ctrl/issues/668) at the S16.5 closeout.)*
- **Library of Leng** (including strict voluntariness CR 701.9) → later sub-PR. Tracked at [#160](https://github.com/krakenhavoc/cmd_and_ctrl/issues/160).
  - *Corrected 2026-09-16:* the voluntariness framing was wrong ([ADR 0013 §10a](decisions/0013-replacement-effects.md)); #160 closes as not planned, and the work is [#650](https://github.com/krakenhavoc/cmd_and_ctrl/issues/650) after [#651](https://github.com/krakenhavoc/cmd_and_ctrl/issues/651).
- **Hangarback Walker** — needs X-cost stack plumbing; future on-demand PR.
- **Champion of Lambholt** — counter half is a trigger (S19), block-restriction is S18. Not a replacement.
- **Mycosynth Lattice clauses** — re-homed to S18 mana-ability rewrite.
- **Right-click admin context menu** → [#170](https://github.com/krakenhavoc/cmd_and_ctrl/issues/170) standalone mini-sprint. Shift+click counter is the S17-era stopgap. *Shipped in #256 ([ADR 0028](decisions/0028-admin-context-menu.md)), which removed the Shift+click chord. #170 closed out on 2026-09-16; its two unshipped items moved to [#671](https://github.com/krakenhavoc/cmd_and_ctrl/issues/671) (reveal a hand card) and [#672](https://github.com/krakenhavoc/cmd_and_ctrl/issues/672) (take one creature out of combat).*

### Exit criteria (all met)

1. **CR 616 ordering works.** Doubling Season + Hardened Scales on battlefield; add +1/+1 counter; `replacement_order` modal pops; submit `[HS, DS]` → 4 counters land; retry with `[DS, HS]` → 3 counters land. Triple-stack with Branching Evolution → single prompt, no re-prompting per applied effect.
2. **CR 903.9 commander-zone prompt.** Any commander move (spell-driven, SBA-driven, wrath-driven, admin-driven) to graveyard / exile / hand / library queues the `optional_replacement` yes/no prompt. Owner says yes → command zone. Owner says no → proceeds to destination.
3. **Kismet taps opponents' permanents.** Opponent plays a creature / artifact / land → enters with `Tapped == true`. Your own permanents unaffected.
4. **Stasis skips untap.** Advance through own turn; `StepUntap` skipped; `StepUpkeep` entered directly. Tapped permanents stay tapped across turn cycles.
5. **Fog cancels combat damage only.** Cast Fog pre-combat; combat damage step lands 0 damage. Non-combat damage still lands. Clears at cleanup so next turn resolves normally.
6. **Fetched lands enter tapped.** Cultivate / Path to Exile / Solemn Simulacrum → fetched land has `Tapped == true`.
7. **No regression in S14–S16 catalog tests.** Every existing catalog card passes.
8. **ADR 0013 + sprints.md + AGENTS.md §7 + GitHub issue updates** all landed.

### Sub-PR split (as shipped)

1. **[#158](https://github.com/krakenhavoc/cmd_and_ctrl/pull/158) — ADR + sprints.md + AGENTS.md + tracking.** Zero code. Filed follow-up issues [#159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/159), [#160](https://github.com/krakenhavoc/cmd_and_ctrl/issues/160). Branch: `feat/s17-adr-plan`.
2. **[#163](https://github.com/krakenhavoc/cmd_and_ctrl/pull/163) — Engine skeleton.** `replacements.go` + `builtin_replacements.go` + PendingChoice extensions + `ResolveReplacementOrder` + protocol view + dispatcher leg + 6 pipeline hooks + commander-zone built-in + deleted `applyCommanderZoneReplacementLocked`. Zero catalog replacements — byte-for-byte identical behavior. Branch: `feat/s17-engine-skeleton`.
3. **[#172](https://github.com/krakenhavoc/cmd_and_ctrl/pull/172) — Prompt modal + counter cards.** `ChoicePromptModal.svelte` `replacement_order` branch + Doubling Season + Hardened Scales + Branching Evolution + exit-criterion test. Branch: `feat/s17-prompt-and-counter-cards`.
4. **[#166](https://github.com/krakenhavoc/cmd_and_ctrl/pull/166) — Enters-tapped + skip-step.** Stasis + Kismet + `SearchLibrary.TappedOnEntry` + Cultivate / Path / Solemn finishers. Hangarback + Champion of Lambholt dropped from scope. Branch: `feat/s17-enters-tapped-skip-step`.
5. **[#167](https://github.com/krakenhavoc/cmd_and_ctrl/pull/167) — Fog + turn-scoped damage prevention.** `Game.TurnScopedReplacements` slot cleared at `StepCleanup`. Library of Leng dropped to a later sub-PR. Branch: `feat/s17-fog-damage-prevention`.
6. <a id="s17-sub-pr-6"></a>**[#171](https://github.com/krakenhavoc/cmd_and_ctrl/pull/171) — CR 903.9 commander-zone optional replacement (closes [#164](https://github.com/krakenhavoc/cmd_and_ctrl/issues/164)).** Widened commander-zone AppliesTo, added `ReplacementEffect.Optional` + `optional_replacement` prompt kind, routed `routeBattlefieldCardToOwnerGraveyardLocked` through the pipeline. Pivoted away from the originally-planned Mycosynth Lattice clauses after manual testing surfaced a higher-priority CR 903.9 gap. Branch: `feat/s17-commander-zone-prompt`.

   Sub-PR 6 also gathered bug fixes that surfaced during manual testing: cleanup-step discard scoped to the active player only (CR 514.1); wire-side `CardView.power` / `.toughness` now source from `CurrentPower()` / `CurrentToughness()` so on-card P/T includes counter delta; CounterPips rendering disambiguation; `Game.svelte.sendAction` stashes cast_spell payloads so `castAnyway` / `confirmAutoTap` retries replay targets; server guards cast with empty targets when `target_mode` is non-empty.

7. **#174 (this PR) — sprint flip + ADR consequences.** Flip S17 to `done` in sprints.md, fill in ADR 0013 Consequences, drop MEMORY entry for the arc. Branch: `feat/s17-docs`.

### Critical files (as shipped)

- **Server new:** [server/internal/game/replacements.go](../server/internal/game/replacements.go), [server/internal/game/builtin_replacements.go](../server/internal/game/builtin_replacements.go), [server/internal/game/replacements_test.go](../server/internal/game/replacements_test.go), [doubling_season.go](../server/internal/cards/effects/doubling_season.go), [hardened_scales.go](../server/internal/cards/effects/hardened_scales.go), [branching_evolution.go](../server/internal/cards/effects/branching_evolution.go), [stasis.go](../server/internal/cards/effects/stasis.go), [kismet.go](../server/internal/cards/effects/kismet.go), [fog.go](../server/internal/cards/effects/fog.go), [doubling_season_test.go](../server/internal/cards/effects/doubling_season_test.go), [stasis_kismet_test.go](../server/internal/cards/effects/stasis_kismet_test.go), [fog_test.go](../server/internal/cards/effects/fog_test.go).
- **Server modified:** [mutations.go](../server/internal/game/mutations.go) (pipeline hooks, `MarkCombatDamage` wrapper, `routeBattlefieldCardToOwnerGraveyardLocked` pipeline routing, `executeBattlefieldLeaveLocked` helper, target-required cast guard), [game.go](../server/internal/game/game.go) (`BuiltinReplacements`, `TurnScopedReplacements`, step-transition hook, cleanup-step discard scope), [pending_choice.go](../server/internal/game/pending_choice.go) (two new prompt kinds + resume methods), [effect_hooks.go](../server/internal/game/effect_hooks.go) (7th hook), [effect_api.go](../server/internal/game/effect_api.go) (`AddCounterForEffect` routing, `SearchLibraryForEffectWithOptions`), [events.go](../server/internal/game/events.go) (`EventStepTransition` sentinel + slog for effect errors), [spec.go](../server/internal/cards/effects/spec.go) (`Replacements` field), [wire.go](../server/internal/cards/effects/wire.go) (populate hook), [primitives.go](../server/internal/cards/effects/primitives.go) (`TappedOnEntry`), [cultivate.go](../server/internal/cards/effects/cultivate.go) + [path_to_exile.go](../server/internal/cards/effects/path_to_exile.go) + [solemn_simulacrum.go](../server/internal/cards/effects/solemn_simulacrum.go) (finishers), [view.go](../server/internal/protocol/view.go) (`ReplacementOptionView`, `CurrentPower`/`CurrentToughness` on the wire), [actions.go](../server/internal/actions/actions.go) (`TypeResolveChoice` legs).
- **Client:** [ChoicePromptModal.svelte](../client/src/lib/components/board/ChoicePromptModal.svelte) (two new branches), [CounterPips.svelte](../client/src/lib/components/board/CounterPips.svelte) (abbr + count layout), [PlayerPanel.svelte](../client/src/lib/components/board/PlayerPanel.svelte) (shift+click counter), [protocol.ts](../client/src/lib/protocol.ts) (mirror fields), [Game.svelte](../client/src/routes/Game.svelte) (cast-payload stash).
- **Docs:** new [docs/decisions/0013-replacement-effects.md](decisions/0013-replacement-effects.md), [docs/sprints.md](sprints.md), [AGENTS.md](../AGENTS.md) §7.

### Branch + commit conventions

- Branches shipped: `feat/s17-adr-plan`, `feat/s17-engine-skeleton`, `feat/s17-prompt-and-counter-cards`, `feat/s17-enters-tapped-skip-step`, `feat/s17-fog-damage-prevention`, `feat/s17-commander-zone-prompt`, `feat/s17-docs`.
- Footer:
  ```
  Sprint: S17 — Replacement effects engine (CR 614)
  Issue: #67
  ```

---

## S18 — Combat keywords

**Phase:** 7 · **Goal:** 12 keyword effects with full combat behavior + summoning sickness + CR 510.1c damage-assignment prompt. See [ADR 0014](decisions/0014-combat-keywords.md).

### Architecture

- **Keywords stay strings in `Characteristic.Abilities`.** S16's Lord of Atlantis pattern (Layer 6 `StaticAbility` appends `"flying"`) is the representation. S18 adds consumers, not a new ability type. [ADR 0014 §1](decisions/0014-combat-keywords.md#1-keywords-are-strings-in-characteristicabilities-read-via-one-helper).
- **One helper file** — [server/internal/game/keywords.go](../server/internal/game/keywords.go) with `HasKeyword`, `HasSummoningSickness`, `CanBlock`, `BlockerCountValid`. All combat consumers route through this; no inline `for _, a := range c.Effective().Abilities` loops.
- **Summoning sickness** — `Card.SummonedThisTurn bool`, set on battlefield entry, cleared at controller's untap step. Haste is a read-time bypass in `HasSummoningSickness`, not a clear-on-ETB. CR 302.6 / 702.10.
- **Flash reads off-battlefield** — new `Spec.PrintedKeywords []string` slot so `HasKeyword` works on cards in hand. Dedicated catalog hook `CatalogPrintedKeywords`. [ADR 0014 §8](decisions/0014-combat-keywords.md#8-flash-needs-a-keyword-reader-that-works-on-cards-in-hand).
- **Combat damage rewrite** — `resolveCombatDamageLocked` splits into first-strike (CR 510.2) and regular (CR 510.3) substeps. SBA fires between; dead creatures exit. Double-strike participates in both.
- **Damage assignment prompt** — new `PendingChoiceDamageAssignment` kind for multi-blocker lethals; single-stage (blocker order + amounts + trample-to-player all in one payload). `ResolveDamageAssignment` validator enforces at-least-lethal prefix rule.
- **Deathtouch via `Card.MarkedLethalByDeathtouch`** flag, read by SBA. Cleaner than short-circuiting `DamageMarked >= Toughness`.
- **Lifelink applies universally** — CR 702.15 covers all damage from the source, not combat only. Routed through the mark-damage code path so `DealDamage*ForEffect` catalog helpers credit life too.
- **Menace enforced at declare-blockers close-out**, not per-decl. Single blocker on a menace attacker is silently reverted (blocker effectively didn't block); attacker becomes unblocked.
- **Post-ship fix (`refactor/major-wip`, 2026-07):** combat damage dealt by a commander now accrues toward the 21-damage loss SBA (CR 903.10a), including trample overflow routed through the damage-assignment prompt. The prompt frame caches the source's commander flag + owner so the damage still counts when the attacker dies to blocker damage before the prompt resolves. Commit: `fix(game): accrue commander combat damage toward the 21-damage SBA`.

### Sub-PR breakdown

1. **`feat/s18-adr-plan`** — this PR. ADR 0014, sprints.md expansion, AGENTS.md §7 "Adding a combat-keyword card" subsection, tracking updates on #68/#69/#76/#95, new umbrella issue for deferred keywords.
2. **`feat/s18-helpers-sickness`** — `keywords.go` helpers; `Card.SummonedThisTurn` + set/clear hooks; `DeclareAttacker` summoning-sickness gate; tap-cost activation gate; `Spec.PrintedKeywords` slot + `CatalogPrintedKeywords` hook; auto-generated layer-6 `StaticAbility` stamping printed keywords onto battlefield cards. Unit tests. Zero behaviour change without catalog entries.
3. **`feat/s18-combat-rewrite`** — split `resolveCombatDamageLocked` into two substeps; `assignAndDealCombatDamageLocked`; `Card.MarkedLethalByDeathtouch` + SBA branch; `PendingChoiceDamageAssignment` + `ResolveDamageAssignment` + wire `DamageAssignmentView` + dispatcher leg; client `ChoicePromptModal.svelte` `damage_assignment` branch. Engine-level tests via manufactured battlefield state; zero catalog cards.
4. **`feat/s18-vanilla-keyword-cards`** — batch ① (8 cards): Serra Angel, Colossal Dreadmaw, Giant Spider, Typhoid Rats, Youthful Knight, Fencing Ace, Lightning Elemental, Wall of Stone. Exercises flying, vigilance, trample, reach, deathtouch, first strike, double strike, haste, defender end-to-end.
5. **`feat/s18-multi-keyword-cards`** — batch ② (4 cards): Vampire Nighthawk, Baneslayer Angel, Ambush Viper, Boggart Brute. Exercises lifelink, menace, flash + compound interactions. (Dreg Mangler swapped out for Boggart Brute after oracle-text review showed Dreg Mangler has scavenge + haste, not menace.)
6. **`feat/s18-keyword-badges`** — client `CardTile.svelte` / `Card.svelte` keyword-badge row (FLY/VIG/DS/FS/TR/RE/DT/LL/MEN/DEF/HST/FLS). Reads existing `CardView.Abilities`. Presentational only.
7. **`feat/s18-docs`** — AGENTS.md polish; ADR 0014 Consequences populated with actually-shipped behavior; this sprint section flipped to `done`; MEMORY entry `s18_arc_complete.md`; close #68.

### Tasks

- [x] **sub-PR 1** ([#177](https://github.com/krakenhavoc/cmd_and_ctrl/pull/177)) — ADR 0014, sprints.md S18 expansion, AGENTS.md §7 combat-keyword subsection, update #68/#69/#76/#95, open umbrella issue #176 for deferred keywords
- [x] **(shipped early via S16 hotfix [#157](https://github.com/krakenhavoc/cmd_and_ctrl/pull/157))** Bare-bones blocked-creature combat damage. Replaced by the two-substep rewrite in sub-PR 3.
- [x] Keyword detection helpers (`HasKeyword`, `HasSummoningSickness`, `CanBlock`, `BlockerCountValid`) — sub-PR 2
- [x] Summoning sickness (`Card.SummonedThisTurn`), cleared at controller's untap step — sub-PR 2
- [x] `Spec.PrintedKeywords` slot + `CatalogPrintedKeywords` hook (for off-battlefield flash gating) — sub-PR 2
- [x] Combat damage flow rewrite — first-strike substep (CR 510.2) + regular substep (CR 510.3) — sub-PR 3
- [x] Damage assignment order prompt (`PendingChoiceDamageAssignment`, CR 510.1c) — sub-PR 3
- [x] Deathtouch (`Card.MarkedLethalByDeathtouch` flag → SBA) — sub-PR 3
- [x] Lifelink (life gain on any damage from source, not just combat) — sub-PR 3
- [x] Trample (overflow to defending player, respects blocker at-least-lethal) — sub-PR 3
- [x] Vigilance (skip tap on `DeclareAttacker`) — sub-PR 3 (via vigilance keyword consumed in DeclareAttacker tap-skip)
- [x] Flying / reach (`CanBlock` gate) — sub-PR 2 (helper), consumed sub-PR 3
- [x] Menace (≥2 blockers; enforced at close-out, single blocker silently reverted) — sub-PR 3
- [x] Defender (can't attack, returns `ErrDefender`) — sub-PR 2
- [x] Haste (bypass summoning sickness read) — sub-PR 2
- [x] Flash (`HasKeyword` reads off-battlefield via `PrintedKeywords`) — sub-PR 2
- [x] First strike + double strike (first-substep participation) — sub-PR 3
- [x] 12 catalog cards: Serra Angel, Colossal Dreadmaw, Giant Spider, Typhoid Rats, Youthful Knight, Fencing Ace, Lightning Elemental, Wall of Stone (sub-PR 4); Vampire Nighthawk, Baneslayer Angel, Ambush Viper, Boggart Brute (sub-PR 5)
- [x] Client keyword-badge row on `Card.svelte` via new `KeywordBadgeRow.svelte` (abbreviated 3-letter tokens) — sub-PR 6
- [x] Docs polish + sprint flip + MEMORY arc-complete entry — sub-PR 7 (this PR)

**Explicitly deferred** (tracked on umbrella issue + cross-referenced sprints):
- **Champion of Lambholt** — both halves land in S19 (counter half is a trigger). See [ADR 0014 §10](decisions/0014-combat-keywords.md#10-champion-of-lambholt-deferred-to-s19).
- **Protection** (CR 702.16) — S24 alongside Mind Control. Baneslayer Angel ships without protection clauses. *(Update 2026-09-16: neither S24 nor S30 delivered it; tracked in [#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662).)*
- **Indestructible** (CR 702.12), **damage-prevention shields with charges** (CR 615) — S30. *(Update 2026-09-16: both shipped. Indestructible came in S25 ([#380](https://github.com/krakenhavoc/cmd_and_ctrl/pull/380)) and charged prevention shields in S30 ([#420](https://github.com/krakenhavoc/cmd_and_ctrl/pull/420)).)*
- **Hexproof, shroud, ward, banding, rampage, flanking, fear, intimidate, shadow, exalted, annihilator, persist, undying, tribute, prowess, cascade** — umbrella issue; land when a real card needs them. *(Update 2026-09-16: the umbrella, [#176](https://github.com/krakenhavoc/cmd_and_ctrl/issues/176), is closed. Shipped: hexproof and shroud ([#353](https://github.com/krakenhavoc/cmd_and_ctrl/pull/353), [ADR 0038](decisions/0038-protection-style-keywords.md)); ward ([#421](https://github.com/krakenhavoc/cmd_and_ctrl/pull/421), recovered by [#433](https://github.com/krakenhavoc/cmd_and_ctrl/pull/433), plus life and sacrifice costs in [#647](https://github.com/krakenhavoc/cmd_and_ctrl/pull/647)); cascade ([#426](https://github.com/krakenhavoc/cmd_and_ctrl/pull/426), recovered by [#433](https://github.com/krakenhavoc/cmd_and_ctrl/pull/433)). Tracked separately: prowess [#706](https://github.com/krakenhavoc/cmd_and_ctrl/issues/706), landwalk [#705](https://github.com/krakenhavoc/cmd_and_ctrl/issues/705), regeneration [#667](https://github.com/krakenhavoc/cmd_and_ctrl/issues/667), protection [#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662). The rest (banding, rampage, flanking, fear, intimidate, shadow, exalted, annihilator, persist and undying as keywords, tribute, and also infect, phasing and totem armor) have no tracker and get built on demand, when a card needs one.)*
- **Flash-in-hand badge** — server-side gating is sufficient for S18; hand-view keyword surface follows with S20 smart-cast UI.
- **Mycosynth Lattice mana-ability clauses** — re-homed from S17's provisional sub-PR 6 to the general backlog (not S18; not mana-system work that fits here).
- **Combat-damage substep animation** — engine runs CR 510.2 / 510.3 as two substeps but the client collapses them into one frame, so double strike is visually indistinguishable from first strike against blockers that die in substep 1. Polish work tracked on [#187](https://github.com/krakenhavoc/cmd_and_ctrl/issues/187).

**Exit criteria:** A 1/1 deathtouch attacker takes down a 5/5 blocker (Typhoid Rats blocks Colossal Dreadmaw → Dreadmaw dies; Rats takes 6 damage and dies). Lifelink attackers gain life (Vampire Nighthawk deals 2 → +2 life). Trample carries over (6/6 Dreadmaw vs 3/3 → 3 to blocker, 3 to player). Vigilance keeps attackers untapped (Serra Angel attacks → still untapped post-combat). Multi-blocker damage-assignment prompt queued and resolved for 5/5 Baneslayer Angel vs 2/2+3/3.

---

## S18.5 — Zone browser + library search (mini) ✅

**Phase:** 7 · **Goal:** close the S06/S11 promise of a clickable zone-browser modal for public zones, and give the tutor cards a player-facing pick UI to the extent the wire protocol supports it today.

Mini sprint slotted after S18 started, to ship two pieces of long-promised client UX that never made it into their parent sprints:

- S06 line 172 — "Graveyard / exile / command — clickable modal browser" deferred to S07, never shipped.
- S11 line 173 — "Library — searchable / reorderable" deferred, never shipped.
- S14 shipped the `SearchLibrary` primitive (Demonic/Vampiric Tutor, Cultivate) but the server auto-picks the first library match; there's no player-facing search UI today.

- [x] `ZoneBrowserModal.svelte` — view-only grid for graveyard / exile / command / stack, opens from the pile chips and command-zone "browse" affordance; hover-zoom works unchanged via the shared `hoveredCard` store.
- [x] Any seated player can open any other player's public zones (graveyard / exile / command / stack).
- [x] Owner actions: per-card "hand / field / lib" buttons fire `move_card` through the existing wire path (no new action kind). Non-owners see view-only.
- [x] Client-side `canManageZone` guard mirrors the server's `requireCardController` gate; `buildMovePayload` rejects speculative moves client-side so we never fire an action the server would bounce.
- [ ] **Library search modal — deferred.** The SearchLibrary primitive today auto-picks server-side (`SearchLibraryForEffectWithOptions` picks the first library match and returns). There's no `PendingChoice` emitted, no `search_library` choice kind, no wire frame for the client to render against. Landing the search modal requires a new pending-choice kind (picker options = controller-known library slice filtered by predicate) on the server side — outside the "client UI only" fence this mini-sprint set up, and it's closer in scope to S22 (library manipulation — Scry/Surveil/Explore) than to S18.5. Captured in [ADR 0015](decisions/0015-zone-browser-library-search.md) so S22's kick-off picks it up.

**Deferred:** library-bottom destination from the owner action cluster (ZoneKind doesn't distinguish top/bottom today; Library of Leng and Vampiric Tutor are rare enough that top-of-library is the useful default); stack browser "counter selected" integration (StackOverlay already owns that flow); library-search modal (see above).

**Exit criteria:** Click graveyard / exile / command pile on any seat → modal opens listing cards with images; hover-zoom works; owner sees move cluster; non-owner sees view-only. ✅

---

## S19 — Auto-fire triggered abilities

**Phase:** 7 · **Goal:** ETB / dies / upkeep / cast / combat triggers fire automatically for catalog cards.

- [x] **Sub-PR 1/8 — `s19-dispatcher`**: `TriggeredAbility` type + `triggerHarvester` listener + LKI snapshot (CR 603.10) — [#193](https://github.com/krakenhavoc/cmd_and_ctrl/pull/193)
- [x] **Sub-PR 2/8 — `s19-prompt`**: optional trigger prompts via existing prompt frame (modal prompts deferred — no catalog card needs one yet)
- [x] **Sub-PR 3/8 — `s19-etb-triggers`**: 5 ETB cards (Mulldrifter, Eternal Witness migration, Reclamation Sage, Acidic Slime, Solemn Simulacrum)
- [x] **Sub-PR 4/8 — `s19-dies-triggers`**: 4 dies-trigger cards (Solemn Simulacrum dies half, Filigree Familiar, Doomed Traveler, Wurmcoil Engine) — [#200](https://github.com/krakenhavoc/cmd_and_ctrl/pull/200)
- [x] **Sub-PR 5/8 — `s19-upkeep-triggers`**: 4 upkeep cards (Phyrexian Arena, Bitterblossom, Sulfuric Vortex, Awakening Zone) via `EventBeginUpkeep` — [#201](https://github.com/krakenhavoc/cmd_and_ctrl/pull/201)
- [x] **Sub-PR 5.5/8 — `s19-triggers-on-stack`**: triggers actually use the stack. Sub-PRs 3–5 applied their effect inline from `Build` and returned nil — no stack item, no response window, no CR 608.2b re-check. `StackItem.Effect` + `NewTriggeredItem` + `resolveTopAbilityLocked` running the effect; all 12 cards migrated; prompt-yes / sandbox-move / play-land now drain onto the stack; wire order by `Seq`; client auto-pass stops for ability items (smart-skip escape). [ADR 0018](decisions/0018-triggers-on-the-stack.md).
- [x] **Sub-PR 6/8 — `s19-cast-triggers`**: 5 cast / opponent-draws cards — Rhystic Study, Smothering Tithe, Esper Sentinel, Beast Whisperer, Consecrated Sphinx. Brings the **pay-unless prompt** (`PendingChoicePayUnless` / `PayUnless` primitive: the taxed player pays from pool or auto-tap, or the "unless" consequence fires), `Game.SpellsCastThisTurn` for "first noncreature spell each turn", and `EventCast` / `EventDrawCard` on the harvester. `CastSpell` and the sandbox `draw_card` now drain triggers onto the stack at their own boundary. Treasure tokens are inert until S21 (sacrifice-cost mana abilities).
- [x] **Sub-PR 7/8 — `s19-combat-triggers`**: 4 combat-damage cards — Edric, Spymaster of Trest; Bident of Thassa; Coastal Piracy; Scroll Thief. `Event.Combat` flags the four combat-damage emit sites (attacker→player, attacker→blocker, blocker→attacker, trample overflow) and combat-damage events carry the dealing creature's controller in `Actor`, so Edric's "its controller may draw" prompts and pays the right player even if the creature traded. `combatDamageToPlayerBy` helper for the "creature you control → player" shape.
- [x] **Sub-PR 8/8 — `s19-trigger-ordering`**: CR 603.3b same-controller ordering — a seat with ≥2 differing simultaneous triggers gets a `trigger_order` prompt (identical triggers, e.g. two Bident draws, never ask); the APNAP drain holds the whole queue until every such seat answers; submitted order = resolution order. Client reorder modal reuses the CR 616 list; stack overlay now shows ability items with their source card's art. Tests: ordering via a real Damnation double-kill, LKI under Glorious Anthem, non-catalog manual-announce canary.

**Manual fallback preserved:** `announce_trigger` from S13.1 stays for unimplemented cards and "hidden info" triggers.

**Deferred** (detail in [ADR 0018 → Out of scope](decisions/0018-triggers-on-the-stack.md)): Skullclamp — its trigger keys off the *equipped* creature dying, so it needs the attachment layer ([S24](#s24--equipment-auras-attachments)); Sylvan Library and Mana Crypt — draw-step life-payment / library reorder and a coin flip, both better after the [S20](#s20--auto-target-legality--smart-cast-ui) choice UI; Treasure's sac-for-mana is inert until [S21](#s21--tokens-sacrifice-aristocrats) ships sacrifice as a cost; modal triggers ("draw a card or gain 3 life") wait on a `ModePrompt` slot until a catalog card needs one.

**Exit criteria:** Cast Mulldrifter → on resolve, you draw 2 cards automatically; advance to upkeep with Phyrexian Arena → trigger goes on stack automatically.

---

## S20 — Auto-target legality + smart cast UI

**Phase:** 7 · **Goal:** capstone sprint — per-card targeting predicates; modal/X UI; structured cast dialog.

- [x] **Sub-PR 1 — `s20-target-predicates`**: predicate library (`Creature`, `NonBlack`, `Noncreature`, `PowerLE(n)`, `OpponentControls`, `YouOwn`, …) + `And`/`Or`/`Not`; `game.TargetSpec` (zones, player/card predicates, Min/Max); `LegalTargetsFor`; `ErrIllegalTarget` at announce; `CardView.legal_targets` on the viewer's own hand; client picker highlights the legal set, "No legal target" greys the card; 13 `TargetMode` cards migrated + Doom Blade. [ADR 0019](decisions/0019-structured-targeting.md).
- [x] **Sub-PR 2 — `s20-trigger-target-picker`**: targeted triggers choose on the board. `TriggeredAbility.Targets` → the harvester computes the legal set at trigger time (empty → trigger removed, CR 603.3d), asks "you may" first, then queues a `pick_target` prompt the controller answers by clicking the board (or a card in the zone browser). Chosen ref stamped on the item + re-checked at resolution. Reclamation Sage / Acidic Slime / Eternal Witness off the auto-picker.
- [x] **Sub-PR 5 — `s20-multi-target`**: multi-target clauses. `TargetSpec.Min/Max` do real work (`WithCount(2, 2)` "two target creatures", `(0, 2)` "up to two", `(1, 0)` "any number"); announce rejects duplicates unless `AllowSame` (CR 115.3); spell items remember their spec so resolution checks each slot — a partly-illegal spell does what it can (`Context.LegalTargets()`, `IsTargetLegal` per slot), all-illegal still fizzles. `pick_target` prompts take several refs (`resolve_choice {targets}`); `min`/`max` ride `legal_targets` / `pick_target` on the wire. Client: clicks toggle into a pick list (gold ring) and the banner's Done / Enter fires at ≥ min; `canCastFromHand` needs ≥ min candidates. Arc Trail (positional), Ashes to Ashes (exactly two), Sylvan Reclamation (up to two).
- [x] **Sub-PR 3 — `s20-x-costs`**: X spells end to end. Clicking an `{X}` card opens an X prompt with live affordability from the auto-tap preview (pool + untapped sources) before targeting / cast; `x_value` rides the payload, the S15 gate charges it, negative X rejected. `Context.X()` / `Context.Opponents()` for effects. Blaze, Exsanguinate, Stroke of Genius.
- [x] **Sub-PR 4 — `s20-modes`**: modal spells. `game.ModeSpec{Prompt, Options[{Label, Targets}], Min, Max}` on `Spec.Modes` (`ChooseOne` / `ChooseN` + `Mode(label, targets?)`); announce validates the choice (distinct, in range, count) and derives the cast's target clause from the chosen option, so the S20 legality gate and CR 608.2b re-check apply per mode; `CardView.modes` (owner-only, per-option legal sets); `ModePickerModal` between the X prompt and targeting, greys options with no legal target; `Context.HasMode(i)`. Rakdos Charm, Izzet Charm, Austere Command. Limit: one targeted option per cast (per-mode target slots ride with multi-target).
- [x] Cast dialog: X prompt → mode picker → board targeting, each its own step (sub-PRs 3 + 4); a single consolidated dialog isn't needed while every step has one surface
- [x] Resolution-time re-check using same predicates (CR 608.2b) — sub-PR 1
- [x] Catalog updates: `TargetSpec` on the S19 trigger cards (Reclamation Sage, Acidic Slime, Eternal Witness) — sub-PR 2

**Free-form fallback preserved:** cards without structured predicates use S13.1's free-form picker.

**Deferred out of S20:** per-mode target slots for "choose two" cards with several targeted options (Cryptic Command, Kolaghan's Command); divided damage / `Distribution` UI (Forked Bolt, Fireball's per-target surcharge — the X prompt can't price targets picked after it); Cyclonic Rift overload is an alternative cost (S28); hexproof / shroud / protection as target-legality modifiers.

**Exit criteria:** Cast Doom Blade → picker shows only non-black creatures; cast Cyclonic Rift overload → no picker, mass effect; cast Fireball → X input with live mana-pool validation.

---

## S21 — Tokens, sacrifice, aristocrats

**Phase:** 7 · **Goal:** an aristocrats Commander deck plays end-to-end.

- [x] **Sub-PR 1 — `s21-sacrifice`**: sacrifice as an engine operation (CR 701.21). `EventSacrifice` fires while the permanent is still on the battlefield, then it takes the ordinary route to the graveyard, so dies-triggers and the CR 903.9 commander replacement keep working; not destruction, so indestructible / regeneration never apply. `SacrificePermanentForEffect` + the `SacrificePermanent` primitive + a `sacrifice_permanent` action. Sacrifice-cost mana abilities are live, closing the S19 Treasure deferral. Tokens carry their own rules: `Card.Keywords` / `Card.ManaAbilities` (a token has no oracle ID for the catalog hooks), so flying tokens fly, Wurmcoil makes one deathtouch and one lifelink Wurm as printed, and Treasure / Eldrazi Spawn crack for mana through the existing right-click menu with no client change.
- [x] Token catalog (Food, Clue, Blood, Powerstone, generic creatures) — shipped in sub-PR 4; Map is not modelled
- [x] `Proliferate`, `CreateTokenAdvanced` primitives — **sub-PR 7 — `s21-remaining`**. `Proliferate` (CR 701.34) splits the rule in two: `Game.ProliferateForEffect` takes the chosen permanents and players and gives each another counter of every kind already there (through `AddCounterForEffect`, so Doubling Season doubles a proliferated counter); the catalog primitive makes the choice, today via a deterministic beneficial auto-pick — everything of yours a counter helps, everything of theirs a counter hurts — because "any number of permanents and/or players" is a free-form multi-select the client has no picker for. `CreateTokenAdvanced` takes a structured `TokenSpec` (printed template + `EntersTapped` / `WithCounters` / `WithKeywords`) and creates through `Game.CreateTokensForEffect`, which returns the new instance IDs; the older idiom of mutating a template before handing it to `CreateToken` still works and is what Mary Read's tapped Treasure uses. Declared gap: a token's entry counters do not run the CR 614 counter-replacement pipeline (token creation has no zone move to replace).
- [x] **Sub-PR 2 — `s21-activated`**: activated abilities in the catalog (CR 602) — the fifth way a card does something, after spells, triggers, statics and replacements. `Spec.Activated` with a struct cost (`Tap` / `SacrificeSelf` / `SacrificeOther` / `Mana` / `Life`), an optional target clause validated like a spell's, and the same `Effect` closure S19 gave triggers, so resolution needed no new code. Costs are validated in full before any is paid, and paid at announce — so a creature sacrificed to Goblin Bombardment puts its dies-trigger on the stack ABOVE the ability, which is what makes aristocrats work. Tap costs enforce summoning sickness (`CardView.summoning_sick` on the wire for the affordance). Cards: Goblin Bombardment, Carrion Feeder, Krenko Mob Boss. Client: activated abilities join the right-click menu; a sacrifice cost opens a picker before targeting. [ADR 0020](decisions/0020-activated-abilities.md).
- [x] **Sub-PR 3 — `s21-aristocrats`**: the payoffs. Blood Artist (any creature's death, including its own, targeted drain), Zulaport Cutthroat (your creatures only, every opponent, gain exactly 1), Mayhem Devil (the first `EventSacrifice` consumer — any player, any permanent, and pointedly NOT destruction), Midnight Reaper (nontoken only, the first card to care about the token split). `diedCreature` / `IsToken` helpers. **Exit criteria met and pinned by a test.**
- [x] **Sub-PR 4 — `s21-token-abilities`**: the last catalog hook a token can't reach. `Card.ActivatedAbilities` carries them on the card object (a token has no oracle ID), `ActivatedAbilitiesForCard` prefers it over the lookup, and clone deep-copies it. Food, Clue, Blood and Powerstone then work through sub-PR 2's machinery with no special-casing — Food, Clue and Blood ARE their activated ability. Producers: Thraben Inspector (its Clue is crackable) and Ichor Wellspring, the catalog's first non-creature dies-trigger. Declared gaps: Blood's discard cost (no discard component on `AbilityCost` — see sub-PR 5) and the Powerstone's spend restriction.
- [x] **Pirates decklist batch 1** (out of the sub-PR sequence, on a real Izzet Pirates list): Reckless Fireweaver, Ingenious Artillerist, Marauding Mako, Glint-Horn Buccaneer, Corsair Captain (first card to pair a trigger with a static), Impulsive Pilferer, Angrath's Marauders, Faithless Looting, Mary Read and Anne Bonny. Found a real engine bug: `DealDamageToPlayerForEffect` skipped the CR 614 replacement pipeline entirely, so no damage doubler or prevention shield could ever see a spell's damage to a player. [Triage of all 65 nonland cards](decklists/pirates-mary-read-anne-bonny.md).
- [x] **Sub-PR 5 — `s21-additional-costs`**: "As an additional cost to cast this spell, discard a card" (CR 601.2f) — the third kind of cost, after a spell's mana cost and an activated ability's. `Spec.AdditionalCost` / `CastSpellParams.DiscardIDs`, validated with the announce-time choices (so a rejection costs nothing) and paid once the spell is on the stack, which is the point: the discard payoff triggers ABOVE the spell and resolves first, and a countered spell still costs the card. Cards: Thrill of Possibility, Big Score, Unexpected Windfall. Client: a discard picker opens before the X / mode / target prompts, mirroring the sacrifice-cost picker. [ADR 0021](decisions/0021-additional-costs.md).
- [x] **Sub-PR 6 — `s21-impulse-exile`**: "exile the top card of that player's library — until end of turn, you may cast that card". The first time a card can be played from a zone that isn't the player's own, and exile is shared and public, so the permission had to live on the card: `Card.ExilePlay` names a holder (usually not the owner), a turn it expires on, whether it permits playing a land, and whether mana may be spent as any colour. `CastSpell` grows an `exile` source zone gated on that grant. Cards: Ragavan, Nimble Pilferer (Treasure + cast-only steal) and Breeches, Brazen Plunderer (play, plus the colour relaxation). Client: the exile browser's first *play* affordance. [ADR 0022](decisions/0022-impulse-exile.md).
- [x] ~40 cards: more token producers, sacrifice outlets, proliferate cards — the last 14 landed with sub-PR 7. Proliferate: Steady Progress, Karn's Bastion, Contagion Clasp, Flux Channeler, Evolution Sage, Inexorable Tide. Token producers: Dragon Fodder, Hordeling Outburst, Grave Titan (enters-or-attacks, on S22's `EventAttack`), Pawn of Ulamog, Stern Lesson (the first tapped token through `CreateTokenAdvanced`). Sacrifice outlets and payoffs: Bloodflow Connoisseur, Korvold Fae-Cursed King, Mazirek Kraul Death Priest (High Market was in this batch until a sibling branch landed it first). New token templates: 2/2 black Zombie, 1/1 Thopter.
- [x] Theme-deck smoke test (Korvold-style aristocrats deck plays 3 turns) — `TestS21ThemeDeckAristocratsPlaysThreeTurns` (`server/internal/cards/effects/theme_deck_aristocrats_test.go`). A real 40-card library with a stacked opening hand plays three of its own turns through the turn engine: land → Dragon Fodder's Goblins → Blood Artist + Bloodflow Connoisseur, feed a Goblin to the outlet → Korvold, whose entry eats the second Goblin and draws → Karn's Bastion proliferates both creatures. Written as a Go test rather than a gamecli script because every assertion is game state, which belongs where CI runs it; gamecli would additionally need a live server, a deck upload and a Scryfall dump.

**Exit criteria:** Cast Goblin Bombardment + Blood Artist + Krenko, Mob Boss; sacrifice tokens to Bombardment one at a time → opponent's life ticks down (Blood Artist + Bombardment damage); your life ticks up (Blood Artist gain). — **met** in `TestS21ExitCriteriaAristocratsCombo` (`server/internal/cards/effects/aristocrats_test.go:211`).

**Status: done, and now complete.** All six sub-PRs merged (#216, #217, #219, #225, #230, #232) plus the Pirates batch; sub-PR 7 (`s21-remaining`) closes the three items that were left behind — both primitives, the last 15 cards, and the theme-deck smoke test, which is the first such harness in the repo and is reusable by any later sprint that wants one.

Two engine findings recorded on [#73](https://github.com/krakenhavoc/cmd_and_ctrl/issues/73) were re-verified while doing that work and are **both stale**: `EventAttack` exists (`server/internal/game/events.go:194`, S22) and Grave Titan's attack half rides on it; and a catalog replacement CAN fire on its own source's entry — `gatherActiveReplacementsLocked` grew a third gathering block for exactly that (`server/internal/game/replacements.go:517`), which is what the Temple cycle's enters-tapped uses.

---

## S22 — Card draw + library manipulation

**Phase:** 7 · **Goal:** a draw-heavy Commander deck plays end-to-end.

- [x] `ScryN` (as the `Scry` primitive), `SurveilN` (`Surveil`), `MillToZone`, `DrawAndScry` (as `Scry.Then`) primitives
- [x] ~~`Explore` primitive with branching reveal~~ — **struck, not deferred.** No card in the catalog performs the CR 701.44 keyword action and none of the roadmap's named blockers is an explore card; the only `Explore` in the tree is Horizon Explorer's *name*. A checklist line nothing needs makes the sprint look permanently incomplete, which is the failure mode the status legend above describes. It comes back as a line the day a card wants it.
- [x] ~~`RevealAndChoose` primitive for cascade-style effects~~ — **struck: the use case it was written for shipped without it.** Cascade is S28's `server/internal/game/cascade.go` plus the one-line `effects.Cascade()` / `GrantsCascade()` constructors — Shardless Agent, Maelstrom Wanderer, Maelstrom Nexus, Kozilek, Hydroid Krasis, Satoru — with its own exile loop, random bottoming and "you may cast it" prompt. The reveal half landed separately in the reveal frame below, and the controller-side choose half is `PendingChoiceChooseCards`. What is genuinely left is a *different* thing than this line names, and it has its own home: **an opponent's non-mana choice at resolution**, a named cluster in [the coverage roadmap](decklists/card-coverage-roadmap.md) (Torment of Hailfire, Combustible Gearhulk, Charismatic Conqueror, Painful Quandary, Chain of Vapor — Fact or Fiction is a sixth).
- [x] Delayed-trigger mechanism (`Game.DelayedTriggers`) for "at the next end step" patterns — the name on the plan is the name in the code, `server/internal/game/delayed.go` ([ADR 0026](decisions/0026-delayed-triggers.md), #255)
- [x] `KindLookAtCards` wire frame (controller-only with redacted view) — shipped as the `scry` / `surveil` / `look_at_top` PendingChoice kinds, answered by the existing `resolve_choice` action rather than a `KindLookAtCardsResponse` of its own
- [x] `KindRevealCards` wire frame (broadcast to all) — shipped as `GameView.reveals`, the one field on the view that is byte-identical for every seat
- [x] ~40 cards: passive draw engines, top-of-library manipulation, tutoring, mill, big draw payoffs — **count met, five named exemplars still out.** Roughly 36 registrations across #405, #507, #548, #550 and #552, on top of S21's Preordain and Viscera Seer. Each of the five buckets clears eight. The holdouts are named below: Mystic Remora, Bruvac, Fact or Fiction, Dark Confidant — and **Niv-Mizzet, Parun**, which the tracker names and which is *not* in the catalog; `niv_mizzet_the_firemind.go` is a different card and does not satisfy the line.
- [ ] Theme-deck smoke test (blue draw deck plays 3 turns)

**Exit criteria:** Activate Sensei's Divining Top → personal-only modal shows top 3 cards → reorder → confirm. Necropotence: activate to exile a card → advance to end step → card moves to hand automatically.

**Status: both exit criteria are met, every primitive and wire frame on the plan is built or struck, and the row above reads `**done**` on the strength of that.** What is left is a card-count tail and the smoke test, and [the status legend](#how-to-read-the-status-column) is explicit that neither holds a sprint open. The tracking issue stays open for the tail; the row is not the issue.

**Where the plan and the code parted company — the short version.** Of the nine primitives and frames the tracker scoped, **five shipped under a different name**, **two shipped as written**, and **two were struck as phantom work**. Across the whole 14-line checklist that is **11 shipped, 2 struck, 1 open** (the smoke test), with five named cards still out inside the shipped card-bucket lines. Not one of the nine engine names on the tracker returns a grep hit for the thing it describes, which is exactly how two earlier passes over this issue concluded that shipped machinery was unbuilt. The mapping, once, in one place:

| Tracker name | What is actually in the tree |
| --- | --- |
| `ScryN` | `effects.Scry` (`primitives.go`) + `Game.ScryThenForEffect` |
| `SurveilN` | `effects.Surveil` + `Game.SurveilThenForEffect` |
| `MillToZone` | `effects.MillToZone` + `Game.MillToZoneForEffect` — name as planned |
| `DrawAndScry` | `Scry.Then`, a **field**, not a second primitive |
| `Explore` | nothing, deliberately — struck above |
| `RevealAndChoose` | `game/cascade.go` for the use case; `RevealCards` + `PendingChoiceChooseCards` for the halves — struck above |
| `Game.DelayedTriggers` | `Game.DelayedTriggers` — name as planned |
| `KindLookAtCards` / `…Response` | `PendingChoiceScry` / `PendingChoiceSurveil` / `PendingChoiceLookAtTop`, over `resolve_choice` |
| `KindRevealCards` | `GameView.reveals`, a projection of `Game.Events` |

**`DrawAndScry` is the one the code got right and the plan got wrong.** A composite primitive that scried and then drew would have drawn before the player answered the prompt — `Scry.Apply` only *queues* it — and worse, it would have pulled the card out of the library the reorder was about to put back. Making the continuation a field on `Scry` is the only shape in which "scry 2, **then** draw a card" is Preordain rather than a different card; the comment above the field says so, because getting it wrong is silent.

**`KindRevealCards` is the other.** The plan asked for a wire frame; what shipped is a projection of `Game.Events`, so nothing new sits on `game.Game` and the window survives an undo, a snapshot restore and a deploy without anyone writing code for it — `snapshot_drift_test.go` needed no new classification.

The **first exit criterion is met.** Sensei's Divining Top is in the catalog and both its abilities work: `{1}` opens the controller-only reorder modal on the top three, and `{T}` draws then tucks the Top itself, in that order.

The **second exit criterion is now met too.** Necropotence is in the catalog with all three of its clauses — the skip-draw-step replacement, the discard→exile trigger, and "Pay 1 life: exile the top card face down, put it into your hand at the beginning of your next end step". The two missing pieces were built for it: `Game.ExileTopFaceDownForEffect` is the only exile in the engine that does NOT mark the table as knowers (it clears the card's knowledge set and sets `Card.FaceDown`), and `BounceToHandForEffect` clears the face-down flag on the way into hand, per CR 400.7. The delivery rides the existing CR 603.7 delayed-trigger queue with `ControllerTurnOnly` set, because the current Oracle text says "YOUR next end step" — an opponent's end step in between does not hand the card over.

Shipped:

- **The scry family.** One mechanism, three `PendingChoice` kinds that differ only in where the cards that leave the top go: `scry` (bottom of library, CR 701.22, S21), `surveil` (graveyard, CR 701.25) and `look_at_top` ("put them back in any order"). All three are "look at", not "reveal" — only the chooser is marked a knower, and `FilterViewFor` projects them as backs for every other seat. The count stays public, which is right: "scry 2" is a printed number. The `resolve_choice` dispatcher routes them **by the choice's kind**, not by payload shape, because all three carry `top_order`.
- **`MillToZone`**, generalising `MillCards` along the three axes printed cards need: a destination zone (exile, which is not a mill and does not emit `EventMill`), a stopping predicate, and the list of what moved.
- **`SearchLibrarySpec.ToTop`** — "shuffle, then put that card on top", which retires Vampiric Tutor's S14 to-hand simplification. The delay is the discount on a one-mana tutor; modelling it as to-hand was printing a strictly better card.
- **`TuckToLibraryForEffect`** and **`EventBeginDrawStep`**.
- **~20 cards**: the three surveil lands now surveil; Sensei's Divining Top, Ponder, Crystal Ball, Otherworldly Gaze, Thought Scour, Reliquary Tower; Enlightened / Worldly / Mystical Tutor, Imperial Seal, Fabricate, Beseech the Queen; Psychosis Crawler, The Locust God, Howling Mine, Hedron Crab.
- **Face-down exile** (`ExileTopFaceDownForEffect` / the `ExileTopFaceDown` primitive) and **`SkipYourDrawStep`**, the CR 500.11 step-skip replacement, seat-scoped to its controller in a way Stasis's table-wide untap skip is not.
- **The broadcast reveal** (`RevealForEffect`, the `RevealCards` / `RevealTopOfLibrary` primitives, `GameView.reveals`, and the board's reveal banner). The mirror image of face-down exile, and the other end of the scry family: a look-at marks one knower and the wire hides the cards from everyone else; a reveal marks every seat AND announces itself, because knowledge alone is silent — it changes what a client may see without telling anyone that anything happened, and a shuffle takes it away again before the next frame. The frame carries printed identity and **no instance IDs at all**, for the revealed cards or for the card that revealed them: that is the "a UUID is a correlation handle" argument the public log settled at build time and #513 had to settle again for pending-choice options, arriving a third time and pointed at a card that is about to go back into a hidden zone. It retrofits every card that already printed the word — 59 tutors with `SearchLibrarySpec.Reveal`, Hermit Druid (whose reveal only marked the land it kept), and Consuming Aberration (which declared the reveal unmodelled).
- **The life-for-cards draw engines**: Necropotence, Yawgmoth's Bargain, Griselbrand, Vilis, Broker of Blood, Bloodgift Demon. All five declared `full`.
- **Prompt-after-a-prompt** (#552): `server/internal/game/chained_choice.go`, and Sylvan Library and Ponder with it. See the struck line below.
- **The untap step becomes observable** (#548) and **`{X}` on an activated ability** (#550). Both were blockers on this checklist's mill and top-of-library cards; both are struck below with what actually shipped.
- **The bot no longer kills itself paying a life cost** (#547). Found while writing Necropotence and fixed on its own branch rather than bundled. At exactly 7 life, Griselbrand's `Pay 7 life: Draw seven cards` is a **legal** move — CR 119.4 lets you pay down to zero — so the enumerator offered it, the engine accepted it, and the seat was gone. Silent and total. The cost now rides `legal.Move.Cost *MoveCost` carrying `Life` and `Loyalty`, **beside** `Params` rather than inside it: `Params` is contractually exactly the ActionPayload that performs the move, and the dispatcher reads a life cost off the ability, so a field nothing on the receiving end reads has no business in there. A pointer, so a cost-free move is byte-identical on the wire. This is the half of a draw sprint that only shows up once bots sit at the table, and it is worth naming as S22 work rather than filing under bots.

Still open, each blocked on something specific rather than on time:

- ~~**`Explore`** and **`RevealAndChoose`** — no card in the catalog needs them yet.~~ **Struck rather than carried** — reasoning on the checklist lines above. "Nothing needs it yet" was the honest read three status passes running; a line that is true and unfinishable is a line that should not be on the list.
- **Fact or Fiction** — the reveal frame it was waiting on is built; it is not registered, and it is the only card on this checklist with real engine work still in front of it. "An opponent separates those cards into two piles" is an **opponent's** non-mana choice at resolution, and #552 drew that fence deliberately: `choose_cards` has no redaction pass for a choice made over someone else's zone, which is the whole difficulty. The coverage roadmap already carries the cluster — Torment of Hailfire, Combustible Gearhulk, Charismatic Conqueror, Painful Quandary, Chain of Vapor — so Fact or Fiction is a sixth card on a seam worth one PR, not a card worth a bespoke primitive. Its oracle id is `437b2dab-15e0-4b9a-a204-58622d37a3b3`, so the id half of the blocker below does not apply to it.
- **Dark Confidant** — blocked on a mechanical detail rather than on engine work, and shared with any other card added blind: the catalog keys on Scryfall `oracle_id`, which comes out of the local bulk dump at `$CMDCTRL_DATA_DIR/scryfall/default-cards.json`, and a wrong id silently fails to bind (or, worse, collides with another card — the registry's duplicate panic catches only the second case). Landing under the sibling PR that also takes Psychosis Crawler's invalidation and the smoke test.
- ~~**Sylvan Library** — needs a repeated per-card "pay 4 life or put it back" prompt; the choice queue has no prompt-after-a-prompt composition. The same gap blocks Ponder's "you may shuffle".~~ **Closed under #74.** The queue gained composition: a choice's continuation runs under `g.mu` and may queue the next prompt itself, plus two general links — `PendingChoiceConfirm` ("do A, or do B", with the card's own branch labels) and `PendingChoiceChooseCards` ("choose N of these cards", bounds on the choice so the enumerator offers exactly what the resolver accepts). Both cards are `full`; both new kinds have a case in `internal/legal` and in the heuristic policy, because a prompt `legal` cannot answer is a bot asleep on a live table (#544). See `server/internal/game/chained_choice.go`.
- **Mystic Remora** — cumulative upkeep (CR 702.24), and it is the whole job. The Rhystic half is already `PayUnless`, which Rhystic Study itself uses. What is absent is the keyword's own machinery: an age counter added at the upkeep trigger, a cumulative cost scaled by the counters, and the sacrifice on decline. Nothing about it is draw-specific, which is why it has outlived three status passes on this issue — it is an upkeep mechanic that happens to be printed on a draw engine.
- ~~**Soothsaying** and **Helm of Obedience** — `AbilityCost` has no `{X}` component.~~ **Shipped** (#550). `game.AbilityCost` reads the `{X}` out of its mana string rather than carrying a flag — `DemandsX()` / `XSlots()` / `FloorX()` are all derived, so the view, the enumerator and the client cannot disagree with the card about whether there is an X to announce. The one new field is `MinX`, the printed "X can't be 0" floor, and `Register` panics at boot on a floor with no `{X}` beside it. X is validated at CR 602.2b, locked onto `StackItem.XValue`, and read back through `ctx.X()` exactly as an X spell's `OnResolve` does; `{X}{X}` is two slots, which is what makes Treasure Vault cost six at X=3. The enumerator offers **one** X (largest affordable, at or above the floor) rather than a range, because a range would have spent Helm's whole expansion budget on near-identical activations pointed at the first opponent — #544 verbatim. Helm declares one caveat: a milled card whose move pauses on a prompt is not counted as milled, so a milled *commander* creature stops the run and reanimates nothing while its owner is still being asked about the command zone. Weaker than printed, never stronger.
- ~~**Mesmeric Orb**~~ — shipped. `untapAllForLocked` cleared `Tapped` in a bare loop and announced nothing; every untap in the engine now routes through one primitive (`game/untap.go`) that emits `EventUntapCard` for a permanent that was actually tapped. The step-scoped half of the same seam — "untap all permanents you control during each other player's untap step" — landed alongside it as `Spec.UntapStep`, a contribution to the CR 502.3 turn-based action rather than a trigger: Seedborn Muse, Unwinding Clock, Drumbellower and Bender's Waterskin, and Quest for Renewal's untap caveat dropped.
- **Bruvac the Grandiloquent** — needs a replacement kind for the mill **amount**. ~~`MillToZoneForEffect` moves cards directly and never enters the replacement pipeline the way `drawCardLocked` does.~~ **That half of the sentence was wrong, and it had spread.** Every milled card routes through `routeCardToZoneLocked`, which builds a `RepEventMove` and pushes it through `applyReplacementsLocked` — so Leyline of the Void and the CR 903.9 commander redirect both apply to a mill *today*, and #529 went further and chose the batch up front precisely so a commander's prompt pauses one card without shortening the rest of the run. What is missing is narrower and differently shaped: "if a player would mill one or more cards, they mill twice that many instead" is not a per-card zone move at all, and `ReplacementEventKind` has `draw` / `move` / `counter` / `life` / `damage` / `step` and no `mill`. The doubling itself is proven — Doubling Season is `ev.CounterDelta *= 2` — so this is one event kind plus its emit site, not a pipeline. **Where the wrong version still reads**, found by searching every PR and comment that names `MillToZoneForEffect`: #548's PR body, and the #405, #507 and #548 status comments on #74. Those are merged artefacts and are left alone; this section is the durable record and now says the right thing. Note that **#539's PR body already documented the correct behaviour** — "a paused commander stays on top of the library" is only a sentence anyone can write if the mill is running the CR 903.9 redirect — so the two have been contradicting each other on `main` since #539 merged.
- **Psychosis Crawler's P/T goes stale between layer recomputes** — the invalidation list does not include hand-size changes, and two tests in `internal/game` (`TestLayerVersionDoesNotBumpOnZoneMoveOutsideBattlefield`, `TestLayerVersionDoesNotBumpOnIrrelevantEvent`) deliberately guard against widening it, because a bump on every hand move makes the recompute much hotter. The fix that respects both is conditional invalidation — bump on a hand change only while a hand-size CDA is on the battlefield — which is a layer-engine change, not a card change. Declared on the card. **In flight** on the same sibling PR as Dark Confidant.
- **Theme-deck smoke test (blue draw deck plays 3 turns)** — **in flight**, on the sibling PR above. It was gated on Sylvan Library and Mystic Remora for three status passes; Sylvan Library landed in #552, and the deck is worth simulating without Remora. The shape to copy is `theme_deck_smoke_test.go` (S27) and `theme_deck_aristocrats_test.go` (S21): cards come off the real `deck.ToGameCard` import road, every cast pays through Strict + AutoTap, and the turns are real turns rather than teleported steps — the three properties that make a smoke test catch what the per-primitive fixtures cannot.

**This section is the record for #74's checklist; the tracking issue's own boxes are unticked and are not maintained.** The reconciled, line-by-line version was posted as a comment on the issue rather than edited into its body, per the standing rule.

The nine older `feat(s22)` commits on `main` shipped attack triggers, flicker, alternative cast costs (which is S28 scope — see that section) and roughly fifty cards; they are good work under a misleading tag. See the note under the sprint index.

---

## S23 — Mass removal + boardwipes

**Phase:** 7 · **Goal:** mass-effect cards work; boardwipes wipe correctly across decks.

- [x] `DestroyAllMatching`, `ExileAllMatching`, `BounceAllMatching`, `ReturnAllToHand` primitives
- [x] Predicate-driven mass effects with non-X exclusions
- [x] ~30 cards: Wrath of God (extended), Damnation, Toxic Deluge, Vandalblast, Austere Command, Farewell, Merciless Eviction, Cyclonic Rift overload (extended), …
- [ ] Theme-deck smoke test (control deck plays 3 turns including a boardwipe)

**The plan, written just-in-time and then executed.** Four sub-parts, in dependency
order, because each one is what makes the next honest.

**1. Simultaneity first, because every card depends on it.** A board wipe is ONE
event (CR 700.4), and this engine destroyed permanents one at a time. The trigger
harvester answers each `EventLTB` by walking the live battlefield, so a Blood
Artist wiped alongside three other creatures was found only for the deaths that
happened to be processed after it — a Zulaport Cutthroat at the front of the
battlefield slice drained for **one** death instead of four, and the number
depended on insertion order. The fix is `server/internal/game/simultaneous.go`:
publish the batch's pre-move copies before the first move, and let the harvester
scan those alongside the live zone, skipping anything still on the battlefield
(`harvestFromZone` has it) and the card whose own death is being reported
(`harvestLTB` has it). The moves stay sequential; the *observation* becomes
simultaneous, which is the half with rules consequences. The same batch wraps the
state-based-action sweep (`mutations.go` §704.3), so damage wipes and `-X/-X`
wipes get it too. `TestWrathDeathsAreSimultaneousForAristocratsPayoffs` and
`TestShrinkWipeDeathsAreSimultaneousToo` both fail by exactly this margin with the
hook removed.

**2. The four primitives** (`server/internal/cards/effects/mass.go`), each a
predicate plus a batched mover plus a `Then` that receives the count — because
"for each creature destroyed this way" (Fumigate, Deadly Tempest, Bane of
Progress) is unanswerable from a fire-and-forget loop. `ReturnAllToHand` is the
explicit-set sibling of `BounceAllMatching` for sets that are not battlefield
predicates (Aetherize's attackers live in combat state).

**3. The exclusion vocabulary**, built on S20's predicate library rather than
beside it — `Except`, `Subtype`, `AnySubtype`, `ControlledBy`, `Multicolored` in
`targets.go`. "All creatures except for Krakens, Leviathans, Octopuses, and
Serpents" is `Except(Creature(), AnySubtype(…))` in the printed order; "you don't
control" is the existing `OpponentControls()`; "non-Dragon" is
`Except(Creature(), Subtype("Dragon"))`, built from the same predicate object as
Crux of Fate's other mode so the two provably partition the board.

**4. The cards.** Twelve existing wipes rewritten onto the primitives (Wrath,
Damnation, Day of Judgment, Blasphemous Act, Damn, Pyroclasm, Austere Command,
Farewell, Merciless Eviction, Cyclonic Rift, Vandalblast, Aetherize) and eighteen
new ones. (Chandra's Ignition was a nineteenth until the roadmap batch landed it
on `main` mid-flight; S23's copy was dropped at rebase and only its test kept.)
Two pieces of cost machinery came with them: `AdditionalCost.PayLifeX`
for Toxic Deluge's "pay X life" (CR 601.2f, with CR 119.4 enforced at announce,
and a third source for the client's X prompt), and `Spec.CantBeCountered` for
Supreme Verdict, honoured at the counter choke point so a Counterspell aimed at it
*resolves* and does nothing rather than fizzling (CR 701.6a).

**One real bug fixed in passing.** Blasphemous Act was registered as "destroy all
creatures" since S14; the card deals 13 damage. Damage is survivable by
indestructible, stoppable by prevention, profitable for lifelink, and kills on the
SBA rather than immediately. It now deals damage.

**Deliberately not registered**, per [ADR 0037 §5](decisions/0037-unimplemented-card-signal.md)
("a card joins the catalog when its whole printed text is carried out, or it does
not join") — recorded so the next person does not re-derive the blocker:

| Card | Blocked on |
|---|---|
| Star of Extinction, Brotherhood's End | "damage to each planeswalker". `DealDamageToCreatureForEffect` only increments `DamageMarked`, and the planeswalker SBA reads loyalty counters (`mutations.go` 704.5i) — damage to a planeswalker removes no loyalty anywhere in the engine. |
| Akroma's Vengeance, Rout | Cycling and "cast as though it had flash for {2} more" — both alternative cast paths that `AlternativeCost` does not carry (it pairs a price with a text rewrite, not a timing change). |
| Vanquish the Horde, Hour of Revelation | Cost reduction, which has no `Spec` hook (S28). Blasphemous Act's pre-existing exception is not a licence to add more. |
| Devastation Tide | Miracle. |
| The Meathook Massacre | Its ETB reads the X paid when it was cast; `Card` does not carry the announced X past resolution. |
| Bontu's Last Reckoning | "Lands you control don't untap during your next untap step" — `untapAllForLocked` has no hook, the same gap ADR 0037 records for Ty Lee. |
| Hallowed Burial | "On the bottom of their owners' libraries" needs a battlefield→library mass move, a fifth primitive outside this sprint's four. |
| Wildfire | "Each player sacrifices FOUR lands"; `EachPlayerSacrifices` has no count. |
| Sunfall, Anger of the Gods, Kindred Dominance, Slaughter the Strong, Living Death, Culling Ritual, Urza's Ruinous Blast | Incubate, a die-replacement, a creature-type prompt, a multi-select prompt, a three-phase mass reanimation, a mana-colour prompt, and the legendary-sorcery cast restriction, respectively. |

**Known gaps this sprint did NOT close**, and did not pretend to: indestructible,
regeneration and totem armor are still absent from the engine, so "destroy all
creatures" really does destroy all creatures and "they can't be regenerated" is
still cosmetic. `DestroyAllMatching` is where all three will be honoured when they
land. The theme-deck smoke test is the one checklist item still open.

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

- [x] `BoostUntilEOT`, `GiveKeywordUntilEOT`, `HexproofUntilEOT`, `IndestructibleUntilEOT` primitives — shipped as **two** primitives, not four. `BoostUntilEOT` and `GrantKeywordUntilEOT` landed with #314 (`server/internal/cards/effects/until_end_of_turn.go`); the hexproof and indestructible variants are that second primitive with a different token, since a duration is orthogonal to which keyword is granted. What was actually missing was the CONSUMER for `indestructible` — see below.
- [x] Until-end-of-turn effect lifecycle (cleanup-step removal of one-shot continuous effects) — shipped with #314: `server/internal/game/turn_scoped_statics.go`, swept by `ClearExpiredTurnScopedStaticsLocked` at the cleanup-step hook. [ADR 0035](decisions/0035-until-end-of-turn-effects.md).
- [x] **Indestructible (CR 702.12)** — not on the original checklist, and the actual blocker behind two of its items. `server/internal/game/indestructible.go` gates `DestroyPermanentForEffect` and the two damage-driven creature SBAs. Un-blocks Boros Charm's mode 1 and Darksteel Citadel, both of which were shipped with declared-inert grants betting on exactly this. Closes the indestructible line of [#176](https://github.com/krakenhavoc/cmd_and_ctrl/issues/176).
- [x] Per-commander damage UX from S13.1 exercised heavily — and found broken. `Player.CommanderDamage` was keyed by opposing PLAYER id while the client's `HoverZoomOverlay` bar chart read it by commander INSTANCE id, so every row rendered 0. Rekeyed to the commander (CR 903.10a), which also fixes partner commanders pooling two 21-damage clocks into one.
- [x] Cards (13): Heroic Intervention, Blossoming Defense, Ranger's Guile, Snakeskin Veil, Tamiyo's Safekeeping, Make a Stand, Titanic Growth, Avacyn Angel of Hope, Bastion Protector, Selfless Spirit, Dauntless Escort, Zurgo Helmsmasher, Uril the Miststalker, Sigarda Host of Herons, Gladecover Scout, Slippery Bogle.
- [x] **Uril, the Miststalker is complete**, Aura count included. It was written with that clause declared as a deferral — there was no attachment relation when this sprint started — and S24's [#374](https://github.com/krakenhavoc/cmd_and_ctrl/pull/374) landed the relation on `main` partway through. The deferral note had predicted "a one-line predicate over that field the day it lands"; it was. Layer 7c over `Card.AttachedTo`, filtered to `IsAura()` so Equipment on the same field doesn't double-count.
- [ ] **Bruna, Light of Alabaster** — still deferred, but no longer on the relation. Her trigger attaches *any number* of Auras drawn from the battlefield, the graveyard AND the hand in one resolution, which needs a multi-select prompt spanning three zones that the attachment UX does not yet have. Belongs with the S24 attachment-cards work ([#76](https://github.com/krakenhavoc/cmd_and_ctrl/issues/76)) rather than here.
- [ ] Big-equipment cards — belong to S24 ([#76](https://github.com/krakenhavoc/cmd_and_ctrl/issues/76)), which owns the relation and is shipping the Equipment/Aura card batch.
- [ ] Theme-deck smoke test (voltron deck deals 21 commander damage in 3 turns) — waits on the Equipment/Aura card batch; a voltron deck with neither is not the deck under test.
- [ ] "ramp + protection" — the protection half shipped (above). Ramp was already covered by S14–S22 (Sol Ring, Cultivate, Arcane Signet, the signet/rock suite) and needed nothing here.

---

## S26 — Tribal / creature type matters

**Phase:** 7 · **Goal:** tribal Commander decks (Goblins, Merfolk, Slivers, etc.) work end-to-end.

- [x] Creature-type vocabulary — `game.AllCreatureTypes` / `IsCreatureType` / `SharesCreatureType` (CR 205.3m)
- [x] Changeling as an enforced keyword (CR 702.73a) — works in **all zones**, reaches non-catalog cards
- [x] `ChooseCreatureTypeOnETB` primitive + per-permanent `Card.NamedTribe` (carried by clone and snapshot)
- [x] `GrantAllCreatureTypesUntilEOT` (layer 4, turn-scoped) and the `TypeFilter` family of predicates
- [x] Dynamic mana-spend restrictions (`ManaAbility.RestrictionsFunc`) for Cavern of Souls
- [x] 14 cards: Cavern of Souls, Door of Destinies, Vanquisher's Banner, Adaptive Automaton, Coat of Arms, Maskwood Nexus, Elvish Archdruid, Elvish Champion, Goblin King, Goblin Chieftain, Death Baron, Irregular Cohort, Shields of Velis Vel, + a Lord of Atlantis oracle fix
- [ ] Theme-deck smoke test (tribal deck plays 3 turns with type-locked Cavern + lord buffs)

### Plan (written for #78)

**Design decision — one mechanism for "is every creature type".** CR 702.73a
makes changeling a characteristic-defining ability that works in every zone.
Rather than stamping ~330 subtypes onto `Characteristic.Subtypes` (which would
blow up the wire `type_line` and every subtype loop), "every creature type" is
carried as the `"changeling"` token in the ability list, and `Card.HasSubtype`
answers `true` for any CR 205.3m creature type when it is present. That one
change gets the all-zones rule for free, because `HasKeyword` already falls
back to the card's printed `Keywords` off the battlefield — which is also why
a vanilla changeling (Woodland Changeling, Universal Automaton) needs **no
catalog entry at all** since #330 put Scryfall's keyword array on every
imported card.

**Order of work (one commit each):**

1. **Engine — creature types + changeling.** `game/creature_types.go`
   (vocabulary + `SharesCreatureType`), `"changeling"` joins the closed
   `canonicalKeywords` table, `Card.HasSubtype` consults it.
2. **Engine — named-tribe state + prompt.** `Card.NamedTribe`, the
   `choose_creature_type` pending choice with its resolve path and wire view,
   and the client picker.
3. **Engine — dynamic mana restrictions.** `ManaAbilityShape.RestrictionsFunc`,
   plus a changeling-aware `ManaSpendContext`.
4. **Catalog — primitives and cards.** `effects/tribal.go` holds the shared
   builders (`ChooseCreatureTypeOnETB`, `NamedTribeAnthem`, `TypeFilter`,
   `GrantAllCreatureTypesUntilEOT`); one file per card.

**Cards (14):** Cavern of Souls, Door of Destinies, Vanquisher's Banner,
Adaptive Automaton (named tribe); Coat of Arms, Maskwood Nexus (type payoffs);
Elvish Archdruid, Elvish Champion, Goblin King, Goblin Chieftain, Death Baron
(lords); Irregular Cohort, Shields of Velis Vel (changeling); plus an oracle
fix to the existing Lord of Atlantis, which was restricted to its controller's
Merfolk and should not be.

**Deferred, declared:** Cavern of Souls' "and that spell can't be countered"
stays inert for the reason Path of Ancestry and Delighted Halfling already
record — the counter path has no per-spell uncounterable flag. Maskwood Nexus'
"creature cards you own that aren't on the battlefield" half is not modelled:
the layer engine maintains `effective` for battlefield cards only, and
`HasSubtype` off the battlefield has no `*Game` to ask. Metallic Mimic, Icon of
Ancestry and Obelisk of Urd wait on enters-with-counters-on-others, look-at-top-N
and convoke respectively.

---

## S27 — Card-type completeness (planeswalkers, sagas, vehicles, battles)

**Phase:** 7 · **Goal:** the four card types missing or only partially modeled by S13–S26 become full citizens.

- [x] `LoyaltyAbility{Cost, Effect}` — proper stack-item activation (CR 606); replaces S13.1's thin "delta-on-action" model — shipped as **`AbilityCost.Loyalty`** (`server/internal/game/activated.go`, written `LoyaltyCost(n)` in card files) on the S21 activated-ability pipeline, with CR 606.2 / 606.6 / 606.3 derived from the component in `ActivateCatalogAbility` rather than asked of each card (#347 / #371, [ADR 0032](decisions/0032-planeswalkers.md) §8). `LoyaltyActivatedThisTurn` is the gate for both paths. One departure from the line: the S13.1 `ActivateLoyalty` action was **hardened, not retired** — it stays as the manual path for planeswalkers with no catalog entry (ADR 0032 §8 says why)
- [x] Planeswalker uniqueness SBA (CR 704.5j) — controller chooses which to keep — shipped as **the legend rule** (`server/internal/game/legend_rule.go`, the `legend_rule` pending choice, #418). CR 704.5j *is* the legend rule; the subtype-based "planeswalker uniqueness rule" was deleted when planeswalkers became legendary, and building it as written would have stopped two different Teferis coexisting
- [x] `declare_attacker` polymorphic target — attack player OR planeswalker OR battle (CR 506.4, 508.1d) — `server/internal/game/attack_target.go` (#415, which also closed #406: damage to a planeswalker removed no loyalty)
- [x] Saga chapter triggers + lore counter advance as turn-based action (CR 714) — `ChapterTrigger` / `ChapterTriggerTargeting` (`server/internal/cards/effects/saga.go`); the CR 714.3 entry and precombat-main lore counters and the CR 704.5s sacrifice in `server/internal/game/sagas.go` (#378)
- [x] `CrewCost{N}` cost component + `BecomeCreatureUntilEOT(P, T)` effect (CR 702.122) — `AbilityCost.Crew` plus `CrewCost` / `BecomeCreatureUntilEOT` / `CrewEffect` in `server/internal/cards/effects/vehicles.go` (#407)
- [x] `BattleSpec{Defense, Subtype}` + `Card.ProtectorPlayerID` + ETB protector prompt + defeating triggers (CR 310) — `BattleSpec` / `DefeatedTrigger` / `SiegeDefeated` in `server/internal/cards/effects/battles.go`; `ProtectorPlayerID` and the `choose_protector` prompt in `server/internal/game/battle.go` (#415). Two pieces the line did not name were needed: `Card.StartingDefense`, stamped at deck import (without it every battle entered with zero defense and died to CR 704.5v — #274's battle twin), and a per-instance face on the exile-play grant so a Siege casts its back face (#574)
- [ ] ~30 cards: 8 planeswalkers, 8 sagas, 6 vehicles, 4 battles, 4 counter-payoff bridges — **29 of 30 registered; one battle out.** Planeswalkers 8/8 (#418, #572; Teferi, Hero of Dominaria replaces "Atraxa", which is not a planeswalker). Sagas 8/8 with four substituted (#378). Vehicles 6/6 (#407). Battles 3/4 (#415, #574). Counter-payoff bridges 4/4 (#418; Hardened Scales and Doubling Season were already in). Invasion of New Phyrexia is the one left, under *What is left* below. Heart of Kiran's alternative crew cost shipped with #625
- [x] Theme-deck smoke test (Superfriends/sagas/vehicles deck plays through 4 turns) — `TestThemeDeckPlaysFourTurns` (`server/internal/cards/effects/theme_deck_smoke_test.go`, #506). Cards come off the real `deck.ToGameCard` import road, casts pay through Strict + AutoTap, and the turns are real turns

**Status: done.** The exit criterion — the theme-deck smoke test — is met (#506), and every primitive on the list is built, two of them under the names the rules actually use. The one unticked line is a card-count tail, and [the status legend](#how-to-read-the-status-column) is explicit that it does not hold a sprint open. [#92](https://github.com/krakenhavoc/cmd_and_ctrl/issues/92) closes with this reconciliation; the tail is tracked in its own issues, listed below.

**What is left, and what each piece is waiting on.** Neither is a card-file follow-up; each needs machinery whose shape was undecided at closeout. Decided 2026-09-16: emblems get an ADR first ([#623](https://github.com/krakenhavoc/cmd_and_ctrl/issues/623)); `ChooseCardsPrompt` gets a set-level `Validate` following `SearchLibrary` ([#624](https://github.com/krakenhavoc/cmd_and_ctrl/issues/624) — **built**: the hook, the enumerator reaching it, and the rejection shown inside the open prompt); Heart of Kiran's alternative becomes a second ability with a real remove-a-loyalty-counter cost ([#625](https://github.com/krakenhavoc/cmd_and_ctrl/issues/625)); Invasion of New Phyrexia stays out until all three back-face abilities work ([#626](https://github.com/krakenhavoc/cmd_and_ctrl/issues/626)).

- **Invasion of New Phyrexia** (`480ea052-87d7-4fea-ac97-b6121b78f863`). The front is Hydroid Krasis's read-X-in-`OnResolve` pattern and is not the problem. The back face, Teferi Akosa of Zhalfir, is: its +1 is a conditional discard ("discard two cards unless you discard a creature card"), which `ChooseCardsPrompt.Validate` can now express (#624), so that ability is no longer the blocker; its −2 makes an **emblem**, and emblems have no shape (see [the seam registry](engine-seams.md)) (#623); its −3 is a reflexive trigger. Those two still hold the card out. Shipping it with two of three abilities omitted would be a 4-loyalty blank, which the #574 status comment on #92 rejected.
- ~~**Heart of Kiran's alternative crew cost**~~ — **done in #625.** The alternative is a second ability entry, "Crew — remove a loyalty counter from a planeswalker you control", paid with the new `AbilityCost.RemoveCounters` component ([ADR 0020](decisions/0020-activated-abilities.md) addendum). The card is `complete`. The same component landed the missing counter-cost abilities on Dragon's Hoard, Fain, the Broker, Mikaeus, the Lunarch and Benevolent Hydra.
- **Emblem ultimates** on Elspeth, Sun's Champion, Wrenn and Six and Teferi, Hero of Dominaria, plus the other omitted ultimates #572 lists, stay omitted rather than stubbed, per ADR 0032. Not on this checklist, but the same emblem seam as the New Phyrexia back face. Elspeth was the one planeswalker whose omission the catalog page did not publish — #418 left it `unreviewed` — and it now declares `caveats` like the others.

Detailed plan: `/home/node/.claude/plans/s27-card-type-completeness.md`. Builds on S13.1, S13.2 (counters + counter SBAs already shipped), S14, S15, S16, S18, S19, S25.

**Re-scoped 2026-09-11 against [ADR 0032](decisions/0032-planeswalkers.md) (#285).** Two corrections to the list above, neither of which shrinks the sprint's card count but both of which change what the engine work actually is:

- **`LoyaltyAbility` is much smaller than it reads.** It is written as though a whole activation pipeline has to be built; it does not. The stack item, the targeting clause, the `Effect` closure and the sorcery-speed flag all shipped with S21's activated abilities (`server/internal/game/activated.go:65`, `:84`). The **entire** remainder is a `Loyalty int` component on `AbilityCost` (`activated.go:35-59` — today exactly `Tap`, `SacrificeSelf`, `SacrificeOther`, `Mana`, `Life`; [ADR 0020](decisions/0020-activated-abilities.md) excluded loyalty on purpose), its payment step, and moving the once-per-turn gate into `ActivateAbility`. After that the S13.1 `ActivateLoyalty` action is **retired**, not extended. ADR 0032 §7 states this in as many words.
- **Half of the `LoyaltyCost` item is already done.** `LoyaltyActivatedThisTurn` exists (`server/internal/game/game.go:141-146`), enforces CR 606.3 in `ActivateLoyalty` (`mutations.go:1546`), is flushed on turn advance (`game.go:554-563`) and is carried through `Clone` / `RestoreFrom` (`clone.go:67-70`, `:392`). It needs re-wiring, not writing.
- **A prerequisite this sprint assumed was fixed underneath it.** Planeswalkers now enter with their printed loyalty (`Card.StartingLoyalty`, `card.go:96`, stamped at deck import) instead of only when a catalog entry supplied it, which is what made every planeswalker but one die instantly to CR 704.5i. That was #274, not S27, but S27's eight planeswalkers were unplayable without it.

Everything else in the list is untouched and verified absent on `f26c961`: CR 704.5j has zero hits; `ChapterTrigger`, `CrewCost`, `BattleSpec` and `ProtectorPlayerID` have zero hits; `Card.AttackingTarget` (`card.go:162`) is still a player ID, so the polymorphic `declare_attacker` is unstarted.

**Battles, follow-up 2026-09-14.** S27 shipped one battle of four (#415) and named the seam the other three were behind: *"a per-instance face on `ExilePlayPermission` and `faceOnResolve` honouring it"*. Both landed, plus the rule that a face-naming grant narrows the card's own offer rather than widening it — see the S32 addendum on [ADR 0034](decisions/0034-multi-face-cards.md). **Three of four battles** are now in (Innistrad, Karsus, Tarkir), each with its back face registered under `"<oracle_id>#1"`, and the Siege reminder text finally does both halves: exile it, *then cast it transformed*. Invasion of New Phyrexia is the one left; its back face is Teferi Akosa of Zhalfir, whose three loyalty abilities need a conditional discard, an emblem and a reflexive shuffle-into-library, none of which have a shape. The declared simplification for all three is timing, not faces — the free cast is a grant bounded to the turn of the defeat, cascade's trade for cascade's reason.

---

## S28 — Cost modification + alternative casts

**Phase:** 7 · **Goal:** the cost engine that the auto-tapper hooks before pool validation.

- [ ] `CostModifier interface { Modify(*Cost, *Card, *Game, uuid.UUID) *Cost }` registered per static ability
- [ ] Modifier ordering per CR 601.2f: alternative cost → additional costs → increasers → reducers, floor at {1} (CR 117.13)
- [ ] Alternative-cost slots in cast dialog (Force of Will pitch, Fierce Guardianship "if you control a commander")
- [ ] Additional-cost slots (Snuff Out's "pay 4 life", Cabal Therapy's "sacrifice")
- [x] Cascade primitive (CR 702.85) — exile-until-CMC-less, may cast for free — `server/internal/game/cascade.go` plus the `effects.Cascade()` / `GrantsCascade()` constructors ([#426](https://github.com/krakenhavoc/cmd_and_ctrl/pull/426), which merged into a deleted base branch and reached `main` through the recovery PR [#433](https://github.com/krakenhavoc/cmd_and_ctrl/pull/433))
- [ ] ~25 cards: 6 reducers, 6 increasers (Stax pieces), 4 free-cast, 4 additional-cost, 5 cascade
- [ ] Theme-deck smoke test (Maelstrom Wanderer cascade chain with Goblin Electromancer + Trinisphere on table)

**Partly shipped under the S22 tag, 2026-09-11.** The **alternative-cost half of this sprint is live on `main`** and was landed as `feat(s22)` by #257: `Spec.AlternativeCosts` (`server/internal/cards/effects/spec.go:257`) with `Overload` / `Evoke` / `Cleave` builders, plus [ADR 0025](decisions/0025-alternative-costs.md), which also labels itself S22. The "alternative-cost slots in cast dialog" item above is therefore substantially done — for the keyword forms; Force of Will's pitch and Fierce Guardianship's commander condition are not among them. S21 had already delivered the **additional**-cost slot as `Spec.AdditionalCost` ([ADR 0021](decisions/0021-additional-costs.md)).

The **cost-modification half — which is this sprint's actual goal — is untouched**: `CostModifier` returns zero hits across `server/internal/`, there is no CR 601.2f ordering, no reducers, no increasers and no cascade. Do not read the shipped alternative costs as progress on the cost engine; they bypass it rather than hook it. *(Update 2026-09-16: "no cascade" is stale. Cascade shipped the same day this note was written, through #426 and its recovery PR #433; the checklist line above is ticked. The rest of the paragraph still describes the cost-modification half.)*

Detailed plan: `/home/node/.claude/plans/s28-cost-modification.md`. Builds on S14, S15, S20. Convoke / Improvise / Delve / Affinity stay deferred (state-scanning at cast time + different UX).

---

## S29 — Alt-cast paths from non-hand zones

**Phase:** 7 · **Goal:** spells cast from graveyard, exile, or hand-with-special-marker.

- [x] `CastableZones []Zone` per card (default `[Hand]`); cast dialog walks all legal zones and surfaces all legal cast paths as separate buttons with their costs — shipped as **`Spec.CastableZones`** plus `AlternativeCost.FromZone` (`server/internal/game/cast_zones.go`, #409): hand is implicit, a declared zone adds a path, a zone with a bound offer can only be cast from by claiming it, and `castable_here` is the flag the zone browser's graveyard cast button keys off (exile keeps its per-instance `exile_play` grant). The card-level `ZoneExile` declaration is **not** the shape suspend and foretell need, although comments said it was: a card-level declaration makes every exiled copy castable. Open PR #638 corrects those comments; [#659](https://github.com/krakenhavoc/cmd_and_ctrl/issues/659) removes the path
- [x] Flashback (CR 702.34) — cast from graveyard for flashback cost; exile after — `Flashback(cost)` in `server/internal/cards/effects/alternative_cost.go` bundles the price, `FromZone: ZoneGraveyard` and `ExileOnLeavingStack`, so a fizzled or countered flashback is exiled too (#411). **4 of the 8 cards** on develop: Faithless Looting, Lingering Souls, Cackling Counterpart (#411), Increasing Vengeance (#419). Open PR #642 adds Deep Analysis (the life component already existed) and gives Otherworldly Gaze the flashback it printed but never had. Snapcaster Mage and Past in Flames grant flashback to *other* cards → [#652](https://github.com/krakenhavoc/cmd_and_ctrl/issues/652). Memory's Journey needs a set-level target check (a player, plus cards in that player's graveyard); no issue, it waits on the `TargetSpec` analogue of [#624](https://github.com/krakenhavoc/cmd_and_ctrl/issues/624)
- [ ] Madness (CR 702.35) — replacement on discard: the card is exiled **face up**, with no marker (CR 702.35a; this line used to say face down), and its owner may cast it for its madness cost while the trigger resolves → [#657](https://github.com/krakenhavoc/cmd_and_ctrl/issues/657). Blocked three ways: discards don't pass through the CR 614 window ([#650](https://github.com/krakenhavoc/cmd_and_ctrl/issues/650)), a trigger during cleanup waits for the next player's upkeep (CR 514.3a, [#661](https://github.com/krakenhavoc/cmd_and_ctrl/issues/661)), and no grant can name the `"madness"` key or ignore timing ([#652](https://github.com/krakenhavoc/cmd_and_ctrl/issues/652))
- [ ] Foretell (CR 702.143) — a **special action** from hand, taken any time its owner has priority **during their own turn** (CR 116.2h, 702.143a-b), not a sorcery-speed action: pay {2}, exile face down; cast on a later turn for foretell cost → [#658](https://github.com/krakenhavoc/cmd_and_ctrl/issues/658). Needs the hand special action ([#655](https://github.com/krakenhavoc/cmd_and_ctrl/issues/655)), a face-down exile the owner can see ([#656](https://github.com/krakenhavoc/cmd_and_ctrl/issues/656)), and a per-card grant keyed `"foretell"` that leaves the spell foretold (CR 702.143c, [#652](https://github.com/krakenhavoc/cmd_and_ctrl/issues/652))
- [x] Escape (CR 702.138) — cast from graveyard, exile N **other** cards from your graveyard. The exile is part of the **escape cost itself** (`AlternativeCost.ExileFromGraveyard`), not an `AdditionalCost`: CR 702.138a makes "[cost]" one cost paid *rather than* the mana cost, and modelling it as an additional cost would have charged the printed mana cost too. Also ships CR 702.138c's "escapes with N +1/+1 counters". 7 cards (#510): Voracious Typhon, Loathsome Chimera, Ox of Agonas, Fruit of Tizerus, Glimpse of Freedom, Sweet Oblivion, Cling to Dust. The escape rule was cited as 702.144 (demonstrate) here and in the code until the S29 closeout
- [x] Warp (CR 702.185) — not on the original list; shipped under S29 as `Warp(cost)`: a discount from hand, the permanent exiled at the next end step, and an exile-play grant with a `NotBeforeTurn` floor for the later cast (#413). Weftstalker Ardent. [#324](https://github.com/krakenhavoc/cmd_and_ctrl/issues/324) stays open: Anticausal Vestige also puts a card from hand onto the battlefield ([#654](https://github.com/krakenhavoc/cmd_and_ctrl/issues/654))
- [ ] Suspend (CR 702.62) — a **special action** from hand, available whenever its owner could begin to cast the card (CR 116.2f, 702.62c), so instant speed for a card with flash and barred by split second: exile with N time counters; a trigger in exile removes one each upkeep; when the last is removed, cast it without paying its mana cost, and a creature cast that way has haste **until its controller loses control of it**, not until end of turn (CR 702.62a) → [#659](https://github.com/krakenhavoc/cmd_and_ctrl/issues/659). Needs the hand special action and exile-zone triggers ([#655](https://github.com/krakenhavoc/cmd_and_ctrl/issues/655)), the CR 118.6 fix for cards with no mana cost (open PR #648), and counters cleared when a card leaves exile (CR 400.7, open PR #649)
- [ ] Cycling (CR 702.29) — an **activated ability** that works only in hand ("[Cost], Discard this card: Draw a card", CR 702.29a), not a cast path; "when you cycle this card" triggers from **whatever zone the card ends up in**, normally the graveyard, not from hand (CR 702.29c) → [#660](https://github.com/krakenhavoc/cmd_and_ctrl/issues/660). Blocked on a discard cost component and on activation from hand ([#655](https://github.com/krakenhavoc/cmd_and_ctrl/issues/655); both are rows in [the seam registry](engine-seams.md)). Six catalog cards ship without their cycling: Raffine's Tower, Ketria Triome, Marauding Mako, Scrounging Skyray and Magmakin Artillerist declare it in `Caveats`; Sylvan Reclamation (basic landcycling) is still unreviewed
- [ ] Splice (CR 702.47) — addon to instants/sorceries — **deferred, not planned**, decided at closeout (2026-09-16) with no issue filed. About two niche cards, none in the catalog waiting on it, and it needs multi-clause targeting (the "Modal / multi-target clause machinery" seam) plus a second card revealed from hand at cast time
- [ ] ~30 cards: 8 flashback, 4 madness, 4 foretell, 4 escape, 4 suspend, 4 cycling, 2 splice — **12 registered, all of them flashback, escape and warp.** Flashback 4/8 (6/8 with open PR #642). Escape 7, replacing the list: Underworld Breach grants escape rather than having it ([#652](https://github.com/krakenhavoc/cmd_and_ctrl/issues/652)), Uro and Kroxa need cast provenance ([#653](https://github.com/krakenhavoc/cmd_and_ctrl/issues/653)), and Klothys, God of Destiny has no escape. Warp 1. Madness, foretell, suspend and cycling 0, each with its card picks corrected in its own issue: Augury Adept and Reflections of Littjara have no foretell; New Perspectives has no cycling; Greater Gargadon and Hypergenesis wait on other seams
- [ ] Theme-deck smoke test (Muldrotha-style graveyard deck casts via flashback + escape) — open PR #645, `TestGraveyardThemeDeckCastsFromTheGraveyardAcrossTurns`: a **scripted** engine test over three real turns, Strict + AutoTap, cards imported through `deck.ToGameCard`. Not bot-driven, because the bot enumerator never offers a graveyard or exile cast or an alternative cost ([#673](https://github.com/krakenhavoc/cmd_and_ctrl/issues/673)), so a bot game would pass without casting from the graveyard

**Status: partial, and S29's tracking issue closes anyway.** Casting from other zones, flashback, escape and warp are on develop. The exit criterion, the graveyard theme-deck test, is written but still in open PR #645. When it merges, the exit criterion is met and this row moves to **done**: the unticked lines are separate mechanics, and [the status legend](#how-to-read-the-status-column) says a tail does not hold a sprint open. [#94](https://github.com/krakenhavoc/cmd_and_ctrl/issues/94) closes with this reconciliation (as #92 did with S27), and the rest moves to its own issues.

**What is left, and what each piece waits on.** Decided 2026-09-16: ADRs first for granted cast permissions ([#652](https://github.com/krakenhavoc/cmd_and_ctrl/issues/652)) and for hand special actions and activations from hand ([#655](https://github.com/krakenhavoc/cmd_and_ctrl/issues/655)); foretell, suspend and cycling are separate implementation issues on top of that ADR.

- **Granted cast permissions** ([#652](https://github.com/krakenhavoc/cmd_and_ctrl/issues/652), ADR): Snapcaster Mage, Past in Flames and The Grim Captain's Locker grant flashback or escape to specific cards until end of turn; Underworld Breach grants escape through a static. Madness, foretell and suspend's free cast reuse the same grant.
- **Cast provenance** ([#653](https://github.com/krakenhavoc/cmd_and_ctrl/issues/653)): a permanent remembers which alternative cost paid for it (CR 400.7d), for "unless it escaped" (CR 702.138b). Phlage, Uro, Kroxa, Tizerus Charger, Skyway Robber, Pharika's Spawn.
- **Hand-to-battlefield move** ([#654](https://github.com/krakenhavoc/cmd_and_ctrl/issues/654)): Uro's land from hand, and the 15-card "Put-from-hand onto the battlefield" seam.
- **Hand special actions and activations from hand** ([#655](https://github.com/krakenhavoc/cmd_and_ctrl/issues/655), ADR), then **foretell** ([#658](https://github.com/krakenhavoc/cmd_and_ctrl/issues/658), also needs face-down objects, [#656](https://github.com/krakenhavoc/cmd_and_ctrl/issues/656)), **suspend** ([#659](https://github.com/krakenhavoc/cmd_and_ctrl/issues/659)) and **cycling** ([#660](https://github.com/krakenhavoc/cmd_and_ctrl/issues/660)).
- **Madness** ([#657](https://github.com/krakenhavoc/cmd_and_ctrl/issues/657)), after discard routing ([#650](https://github.com/krakenhavoc/cmd_and_ctrl/issues/650)), CR 514.3a ([#661](https://github.com/krakenhavoc/cmd_and_ctrl/issues/661)) and [#652](https://github.com/krakenhavoc/cmd_and_ctrl/issues/652).
- **Bugs found by the closeout audit:** cost modifiers priced a graveyard cast as a hand cast (open PR #638); a card with no mana cost could be cast from hand for free, against CR 118.6 (open PR #648); counters and exile-play grants survived a card leaving exile, against CR 400.7 (open PR #649); a trigger during cleanup waits for the next player's upkeep, against CR 514.3a ([#661](https://github.com/krakenhavoc/cmd_and_ctrl/issues/661)).
- **Bots** never cast from the graveyard or exile, or for an alternative cost ([#673](https://github.com/krakenhavoc/cmd_and_ctrl/issues/673)).
- **Splice** is deferred, not planned (above), alongside dredge, retrace, rebound and aftermath.

Detailed plan: `/home/node/.claude/plans/s29-alt-cast-paths.md`. Builds on S14, S17, S19, S20, S22, S28. Dredge/retrace/rebound/aftermath stay deferred.

---

## S30 — Damage prevention, cloning, face-down, deferred protection keywords

**Phase:** 7 · **Goal:** engine-completeness capstone. After S30 there are no major missing primitives.

- [x] Damage prevention shields (CR 615) — replacement subtype with charges; integrates with S17 pipeline — shipped as charged shields in `server/internal/cards/effects/prevention.go` on the S17 turn-scoped replacement registry; a shield smaller than the damage prevents part of it and is used up (#420). The same PR routed `DealDamageToCreatureForEffect` through the replacement pipeline, which it had been skipping. Fog itself is S17's (#167)
- [x] Cloning (CR 707, not 706) — ETB replacement that captures copyable values; populates S16 layer 1 copy slot — `game.PrintedValues` and `Card.PrintedSelf` (`server/internal/game/copy.go`), `effects.EntersAsCopyOf` (`copy_effects.go`), [ADR 0043](decisions/0043-copy-effects.md) (#414). As ADR 0043 argues, the copy is materialised printed values rather than a Characteristic-level layer 1 effect. What it can't do yet is an "except" clause that **grants an ability** or adds a subtype (CR 707.9a / 707.9b): [#665](https://github.com/krakenhavoc/cmd_and_ctrl/issues/665)
- [x] Spell copies (CR 707.10, not 706.10) — `CopyTopOfStack`; controller chooses new targets _before_ copy hits stack — shipped as `Game.CopySpellForEffect` / `effects.CopySpell` (`server/internal/game/spell_copy.go`), which copies **any** spell on the stack by ID, not only the top (#419). A copy is neither cast nor a card. A copy of a **permanent** spell is refused with an `EventEffectError` at resolution instead of becoming a token (CR 608.3f): [#666](https://github.com/krakenhavoc/cmd_and_ctrl/issues/666). Copying an ability is the engine-seams "Ability-copy" row
- [ ] Morph / manifest (CR 702.37, 701.40, 708 — the line cited 702.36, 701.34 and 707) — face-down zone state on `Card`; reuses S22 hidden-info wire frame — **not started, and deferred.** Only `Card.FaceDown` exists, set for face-down exile (Necropotence). There is no face-down wire frame to reuse: the per-viewer `KnownBy` redaction exists, but a face-down permanent needs a 2/2 projection and catalog hooks that go silent (CR 708.2). One ADR covers foretell's owner-known exile and the battlefield half; only the exile half is built next: [#656](https://github.com/krakenhavoc/cmd_and_ctrl/issues/656). The identity leak through the redacted view is open PR #646
- [ ] Protection (CR 702.16) — predicate-based guard at four DEBT hook points (damage / enchant-equip SBA / block / target) — **not started.** ADR first; it supersedes the protection paragraph of [ADR 0038](decisions/0038-protection-style-keywords.md) §7: [#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662). Two corrections to the line: protection is an engine rule read from `Effective().Abilities`, not a predicate a card opts into (ADR 0038 §1 decided the same for hexproof); and block and the attachment check already receive both objects, so the refactor is targeting plus a last-known-information source for damage. The cost-option views have to stop applying the targeting gate first: open PR #639
- [x] Hexproof (CR 702.11) — targeting-by-opponent guard — `game.CanBeTargetedBy` (`server/internal/game/keywords.go`), with shroud (CR 702.18a) (#353, [ADR 0038](decisions/0038-protection-style-keywords.md))
- [x] Ward (CR 702.21) — `TriggeredAbility` on `EventBecomesTarget` via S19; pay-or-counter prompt — `effects.Ward(WardMana(…))` (`server/internal/cards/effects/ward.go`) resolving into `QueuePayUnlessForEffect`. It was written in #421, which merged into a deleted base branch, and reached `develop` through the recovery PR #433. Life and sacrifice wards (Refraction Elemental, Sedgemoor Witch, Vein Ripper) followed in #647, built on #552's Confirm and ChooseCards prompts with no new prompt kind
- [x] Indestructible (CR 702.12) — flag short-circuits lethal-damage SBA + destroy-effect rejection — `server/internal/game/indestructible.go` (#380). The mass destroy path (`DestroyAllMatching`) honoured it only from #502
- [ ] ~25 cards: 4 fog, 4 cloning, 4 spell copies, 4 morph, 6 protection/indestructible, 3 ward — **14 of the 25 named cards are registered, plus five off the list; the rest have their own issues or open PRs.** Fog 3/4: Fog (#167), Holy Day and Tangle (#420); Tangle's doesn't-untap rider is a no-op, declared in `Caveats` since the closeout; Constant Mists waits on optional additional costs, and its buyback is *sacrifice a land* ([#664](https://github.com/krakenhavoc/cmd_and_ctrl/issues/664)). Mending Hands (#420) is the card that actually prints a charged shield; the four on the list are all uncharged. Cloning 2/4: Clone and Sakashima the Impostor (#414, Sakashima without its granted return ability, declared), plus Phyrexian Metamorph and Spark Double off-list; Stunt Double is open PR #641; Phantasmal Image is [#665](https://github.com/krakenhavoc/cmd_and_ctrl/issues/665). Spell copies 3/4: Reverberate, Twincast, Increasing Vengeance (#419), plus Dualcaster Mage (#437) and Thousand-Year Storm (#445) off-list; Doublecast waits on "when you next cast" triggers ([#663](https://github.com/krakenhavoc/cmd_and_ctrl/issues/663)). Morph 0/4, deferred with [#656](https://github.com/krakenhavoc/cmd_and_ctrl/issues/656); morph alone doesn't unblock Willbender (stack retarget), Backslide (cycling) or Ixidron (turn face down plus a face-down count). Protection/indestructible 4/6: Avacyn, Angel of Hope (#380), Darksteel Forge (#422), Lightning Greaves and Swiftfoot Boots (#379); True-Name Nemesis is [#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662); Asceticism was on the wrong list, since its hexproof half has worked since #353 and it waits on regeneration ([#667](https://github.com/krakenhavoc/cmd_and_ctrl/issues/667)). Ward 2/3 substituted: the listed Sheoldred the Apocalypse, Ledger Shredder and Esika print no ward, so Hulking Raptor and Rimeshield Frost Giant shipped instead (#433); Refraction Elemental (#574) shipped with its life ward declared missing, and #647 made the ward live
- [ ] Theme-deck smoke test (Avacyn + Reverberate + Clone + a fog plays 4 turns end-to-end) — **not written.** Nothing blocks it; every card it names is registered: [#678](https://github.com/krakenhavoc/cmd_and_ctrl/issues/678)

**Status: partial.** Every primitive on the list except morph and protection has shipped. But S30's one exit criterion, the theme-deck test, has not been written, and the goal line ("no major missing primitives") is demonstrably unmet. By [the status legend](#how-to-read-the-status-column) that is `partial`, not `**done**`: unlike S22, where the smoke test sat beside exit criteria that were met, here it *is* the exit criterion. [#95](https://github.com/krakenhavoc/cmd_and_ctrl/issues/95) closes with this reconciliation (decided 2026-09-16, as #92 did at the S27 closeout), and the tail is tracked in its own issues:

- **Protection (CR 702.16):** ADR first, then the DEBT checks, then Baneslayer Angel, both Swords and Animar go Full and True-Name Nemesis ships: [#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662)
- **Face-down objects (CR 406.3a / 708):** one ADR, shared with foretell; the owner-known exile half is built first, morph and manifest are deferred: [#656](https://github.com/krakenhavoc/cmd_and_ctrl/issues/656)
- **"When you next cast" delayed triggers:** Doublecast and Galvanic Iteration; reverses ADR 0026 §1-2, so an amendment or a turn-scoped trigger registry comes first: [#663](https://github.com/krakenhavoc/cmd_and_ctrl/issues/663)
- **Optional additional costs:** kicker and buyback, mana and non-mana; Constant Mists: [#664](https://github.com/krakenhavoc/cmd_and_ctrl/issues/664)
- **Copy effects that grant an ability or add a subtype:** Phantasmal Image; Sakashima loses its caveat: [#665](https://github.com/krakenhavoc/cmd_and_ctrl/issues/665)
- **A copy of a permanent spell becomes a token:** Double Major: [#666](https://github.com/krakenhavoc/cmd_and_ctrl/issues/666)
- **Regeneration (CR 701.19):** Asceticism and the regeneration batch skips: [#667](https://github.com/krakenhavoc/cmd_and_ctrl/issues/667)
- **The exit criterion:** [#678](https://github.com/krakenhavoc/cmd_and_ctrl/issues/678)

Also in flight from the same audit, as PRs rather than issues: Stunt Double (#641), Blood Money and The Battle of Bywater as one simultaneous sweep (#640), life and sacrifice ward (#647, merged to `develop` 2026-09-16), the cost-option views' targeting gate (#639), and the face-down view leak (#646). Each engine seam above also has a row in [the seam registry](engine-seams.md).

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

## S31 — AI bot seat (legal-move enumeration + tiered policy)

**Phase:** 8 · **Goal:** any player at an unstarted table can add up to four bot seats that play the game. Pulled forward from its post-S30 slot; architecture per [ADR 0033](decisions/0033-ai-bot-seat.md), which supersedes the heuristic-only design this section previously carried.

**The bar is Forge's bar,** not tournament strength: makes legal moves, makes locally-sensible decisions, doesn't deadlock, uses removal on threats. Play quality is capped by catalog coverage (a few hundred cards), not by the policy — see ADR 0033 §7 on which archetypes are actually buildable today.

**Four deliverables, in dependency order:** a public game log (the engine has none today), the legal-move enumerator (which the client also consumes, deleting its duplicated timing predicates), the virtual-seat runner with a heuristic policy, and the model-backed tiers on top.

**Prerequisites — three, and one of them is not on this sprint's list.** Verified against `f26c961`:

1. **The legal-move enumerator** — `server/internal/legal/` (sub-PR 1, #286). **Shipped.** This is the closed move list the whole design rests on; nothing bot-shaped is safe to build before it.
2. **`protocol.LogEvent` and a bounded public log on `GameView`** (sub-PR 0). **Shipped** in [#431](https://github.com/krakenhavoc/cmd_and_ctrl/pull/431), after this paragraph was written — `GameView.log` carries the last 200 table-visible events through the same `isKnower` filter that redacts every `CardView`. All four of sub-PR 0's boxes are ticked below. ADR 0033 §4 argues this is a prerequisite rather than a refinement: a policy handed only a snapshot cannot see that the seat to its left wiped the board last turn, and that is most of what a Commander player reasons about. Players want it independently.
3. ~~**Attachments — equipment and auras**~~ ([#280](https://github.com/krakenhavoc/cmd_and_ctrl/issues/280)). **Shipped** on 2026-09-11, after this paragraph was written: the relation in [#374](https://github.com/krakenhavoc/cmd_and_ctrl/pull/374) (`game/attach.go`, `Card.AttachedTo`, the CR 704.5m/n SBA), then the first attachments in [#379](https://github.com/krakenhavoc/cmd_and_ctrl/pull/379) and [#380](https://github.com/krakenhavoc/cmd_and_ctrl/pull/380). This no longer blocks sub-PR 5 — see the note there for what the deck list ended up being and why Voltron still isn't one of them.

Nothing is now blocked on #280. Sub-PR 5's deck list was the only thing that ever was, and it shipped four decks without needing it.

**Read every tick below with one caveat: sub-PRs 5, 6, 7 and 8 shipped their packages but not their wiring.** Verified against `b555289`, and filed as [#501](https://github.com/krakenhavoc/cmd_and_ctrl/issues/501). `server/internal/aiseat/decks` and `server/internal/aiseat/tiers` — the latter being the factory that builds Layer A, the heuristic and the model funnel — are imported by their own tests and by nothing else. `main.go` still wires `BotDecks: aiseat.PlaceholderDecks()`, and `Manager.StartBots` still builds through `aiseat.NewPolicy`, which returns `ErrTierUnavailable` for `heuristic`, `assisted` and `strong`. So a bot seated from the lobby today is a `RandomPolicy` holding `placeholder-mono-red` — a commander and ninety-nine Mountains — and `aiseat.Improviser` has no non-test implementer, which leaves sub-PR 8's path live in the runner and unreachable in play. There are now **two tier registries carrying the same four names**: `aiseat.Tier`, which the API serves and which can build one tier, and `tiers.Tier`, which can build all four and which nothing imports. Every measurement recorded below is real and was taken against packages constructed directly in tests; none of it describes what a player gets from the lobby. That is why the index row stays `partial`.

### Sub-PR 0 — public game log

The engine has no event history: `GameView` carries none, the client has none, and `PlayerView.LifeHistory` is the only past tense anywhere. A policy reasoning from a bare snapshot cannot see a boardwipe that already happened. Players have wanted this since S07 shipped chat without it, and bug reports get materially better when the log ships alongside the replay.

- [x] `protocol.LogEvent` + bounded public log on `GameView` (few hundred entries), through the same visibility filter as everything else
- [x] Emit on: zone changes, casts, resolutions, combat declarations, life changes, step boundaries
- [x] Client game-log panel
- [x] Attach the log to bug-report artifacts in `bugstore`

### Sub-PR 1 — `internal/legal`: enumerate a seat's legal moves

**Shipped** ([#286](https://github.com/krakenhavoc/cmd_and_ctrl/pull/286)) — the prerequisites paragraph above has said so since that PR merged, but this checklist was never reconciled. Verified line by line below; two lines describe something that was not built and stay unticked.

- [x] `legal.Move` — one fully-specified move, wire-ready. **Three corrections to the field list.** `Group` shipped as `Kind Kind`; the struct carries a sixth field this line omits, `Player`; and `Type` is a plain `string`, not an `actions.Type`, because `legal` deliberately does not import `actions` — that is what lets `protocol` import `legal` in sub-PR 2 without a cycle, and `TestWireTypesMatchActions` asserts every constant equal to its `actions.Type*` counterpart instead. `Params` is a `json.RawMessage` built from private mirror structs rather than a `protocol.ActionPayload`, for the same reason. The claim the line is really making — that `Params` is exactly what performs the move — is held executably by `dispatchAll`, which sends **every** enumerated move against a cloned game through `actions.Dispatch` as a seated caller and requires acceptance
- [x] `legal.EnumerateFor(g *game.Game, seat uuid.UUID) []Move` — signature exact. **The lock is the game's, not the room's, and the function takes it itself.** The package doc is explicit that callers must *not* hold `g.mu`; `EnumerateLocked` is the variant for a caller already under the read lock (`ViewOfGame` is one), with the re-entrant-`RLock`-with-a-writer-queued deadlock spelled out next to it. `EnumerateForWithOptions` is the third entry point
- [ ] Coverage — most of it, with three deliberate exclusions and one gap that is not deliberate. **Shipped:** land drop (one per turn, which the enumerator enforces and the engine still does not), `cast_spell` from hand and from the command zone (with commander tax), `activate_ability`, `activate_mana_ability`, `pass_priority`, mulligan/keep, and — not on this line, and the table wedges without it — `discard_selection` for the S13.4 cleanup pause. **Deliberately not enumerated:** `advance_step`, `sacrifice_permanent` and `activate_loyalty` are sandbox affordances rather than game actions, and the package doc names them as such. Note the nuance sub-PR 2's note flattens: catalog *loyalty abilities* **are** offered, as `activate_ability` with an `ability_index` under a CR 606 gate; it is only the raw loyalty-delta verb that is absent. **Not deliberate:** "every live `PendingChoice` kind" was 16 of 20 — `may_cast` (S28 cascade), `legend_rule`, `choose_creature_type` (S26) and `choose_protector` (battles) had no case in `choiceMoves`, and because a seat that owes a choice is enumerated that choice's answers **and nothing else**, a seat owing one of the four got an empty list — no answer, and no pass either. Filed as [#499](https://github.com/krakenhavoc/cmd_and_ctrl/issues/499). **Two of the four closed on 2026-09-14** with the S27 battle follow-up: `choose_protector` and `legend_rule` both answer with the pick_target payload and now ride its case, and the `switch` grew a `default` that logs the kind instead of wedging in silence. `may_cast` and `choose_creature_type` are still open on #499
- [ ] Target and mode expansion — the cap shipped and its default really is 12; **the field name is wrong and the ordering was never built.** The knob is `legal.Options.MaxExpansionPerSource` (`MaxTargetExpansion` appears nowhere in the tree but this line and [ADR 0033](decisions/0033-ai-bot-seat.md) §1), and a second undocumented knob rides with it: `Options.MaxX`, default 20, capping the X an `{X}` spell is tried at. The budget is per **source** across the whole crossed product — modes × targets × discards × sacrifices — not per target/mode, and `TestExpansionCapIsHonoured` pins it. **There is no threat ordering**: no sort and no threat weight exists anywhere in `legal`, candidates arrive in `LegalTargetsForEffect` order (players first, then cards) and subsets come out lexicographically, so which targets survive the cap is an accident of enumeration order. Threat ranking lives one layer up in `aiseat/heuristic/rank.go`, which is arguably where it belongs — `legal` holds no policy opinions — but ADR 0033 §1 promises both the ordering and an "explicit *choose target* move" for the remainder, and neither exists. The remainder past the cap is simply dropped
- [x] Combat enumerates **per creature**, not per assignment — one `declare_attacker` per (eligible creature × legal attack target) and one `declare_blocker` per (eligible blocker × attacker attacking this seat). Not priority-gated, matching the engine, so a defender can block while the active seat still holds priority
- [x] Table-driven tests against the existing `timing.ts` predicates, both directions — `internal/legal/testdata/timing_agreement.json` carries eight hand-declared scenarios; `agreement_test.go` asserts the enumerator matches them and diffs the fixture, and `client/src/lib/timingAgreement.test.ts` asserts `canCastFromHand` matches the same declarations against the same real filtered frames, plus the converse sweep over every card in hand. Neither side is the other's oracle

### Sub-PR 2 — serve the move list to the client, delete the TS duplication

**Shipped** ([#429](https://github.com/krakenhavoc/cmd_and_ctrl/pull/429)). Notes on where the checklist and the code parted company are inline below.

- [x] `GameView` grows `legal_moves` for the viewer's own seat only, populated when the seat holds priority or owes a choice. Enumerated per seat into an unexported map by `ViewOfGame`; `FilterViewFor` hands back the viewer's own entry and nothing else, so the unfiltered frame that reaches the crash dump and the replay log carries no seat's moves at all
- [x] `client/src/lib/timing.ts` becomes a lookup over `legal_moves`. **One correction to the scope line:** `canActivateLoyalty` gained a caller in #334 / #371 (`contextMenu.logic.ts`) and — more to the point — `activate_loyalty` is a sandbox verb `internal/legal` deliberately does not enumerate, so there is no move list to look it up in. It stays, as the file's only surviving rules derivation. `canPassPriority` and `canActivateAbility` were deleted as written
- [x] Delete the now-dead client-side rules reimplementation; keep `targeting.ts`'s presentation logic. Gone: `castTimingForTypeLine`, `castableFaceTypeLines`, `canActivateAbility`, `canPassPriority`, and `priority.ts`'s whole hand/command/battlefield walk. Kept: the snapshot readers and `canCastFromHand`'s target / mode / additional-cost branches, which read server-stamped fields and supply the tooltip sentence behind the server's verdict
- [x] Fixes a live bug in passing: `sorcery_speed` ships on the wire and the client never reads it. **Half of it fixed itself while this was in flight** — the S24 equip work (#379, #380) taught `contextMenu.logic.ts`'s `abilityBlocked` to honour the flag and landed `canActivateSorcerySpeedAbility`, which this branch had independently written under another name and now adopts. The half fixed here is `ManaAbilityMenu.svelte`'s own copy of `abilityBlocked`, which is the DEFAULT right-click popover (`adminOverrides` is off by default) and so the path most players were actually hitting
- [x] View-size check: measure the frame growth on a full four-player board and gate on it staying under budget. **14.0 KB / +38–55%** on the worst realistic frame (39 moves); 0 B on any frame where the seat owes no decision. The enumerator's cap is per SOURCE, which does not bound a board, so the wire projection adds a 48-move global cap that degrades to one move per `(source, kind)` rather than truncating. Gated by `TestLegalMovesFrameBudget` at 24 KiB — **coordinate with sub-PR 0 before spending the rest**
- [x] Both-directions agreement: `server/internal/legal/testdata/timing_agreement.json` carries eight hand-declared scenarios plus the real filtered wire frames. `agreement_test.go` asserts the enumerator matches the declarations; `client/src/lib/timingAgreement.test.ts` asserts `canCastFromHand` matches the same declarations against the same frames. Neither side is the other's oracle

### Sub-PR 3 — `Room` observers + `internal/aiseat` runner

**Shipped** ([#287](https://github.com/krakenhavoc/cmd_and_ctrl/pull/287)), and never reconciled. Verified below; one line names fields that do not exist and is corrected to the shipped names, and one describes a guard that was not built and stays unticked.

- [x] `Room.Subscribe()` — all four properties hold. Edge-trigger (`chan struct{}`, no payload; the doc tells subscribers to re-read the room on wake because there is none), capacity 1, non-blocking send (`select { case ch <- struct{}{}: default: }`), and fired **after** `r.mu` is released. The last one is structural rather than incidental: every mutator is split into an exported wrapper and an unexported `mu`-holding core, and all four `r.notify()` call sites — in `Apply`, `ApplyBundle`, `ApplyExternal` and `Undo` — sit in the wrapper after the core has returned. The subscriber map has its own `subMu`, documented as "never `mu`", precisely so an observer may call `Apply` from its wake without deadlocking. **One correction to the signature:** it returns `(<-chan struct{}, func())`, not a bare channel — the second value is the unsubscribe, and there is no `Unsubscribe` method and no close on game end. A runner exits on its own state check with `defer unsubscribe()`
- [x] Runner goroutine per bot seat — wake → decide → `Room.Apply(seat, fn)` calling `actions.Dispatch`, as written. **Where the four windows live is not where this line puts them.** `Runner.loop` asks no question about priority, choices, mulligans or blockers; it calls `legal.EnumerateFor` and treats an empty list as "not my turn to act". All four tests are inside `legal.enumerateLocked` — `MulligansOpen`, `choiceMoves`, `holdsPriority`, and `combatMoves`, which is deliberately *not* priority-gated so a defender can block while the active seat still holds priority. There is a fifth the line omits, the CR 514.1 cleanup discard. Keeping the windows in one place is the better arrangement — the client's `legal_moves` gets the same answer — but the runner is not where to look for them
- [x] `RandomPolicy` — uniform over `Moves`, `math/rand/v2`, seeded per seat from a PCG in `NewPolicy`; a nil source falls through to the process generator. It is the `random` tier and the engine fuzzer, and it is deliberately **not** wrapped in the Layer A filter: a filter that answered the trivial windows correctly would narrow the fuzzer for no gain
- [x] Improvisation bundles apply as `uuid.Nil` so any seated human can undo them — confirmed, and the caller gate needed no change: `caller != uuid.Nil && top.caller != uuid.Nil && top.caller != caller` short-circuits on a nil-stamped entry. **The open question was decided in sub-PR 8: cleaning up after a bot is free.** `ws.undoEntry.freeUndo` is set by `ApplyBundle` from `Bundle.FreeUndo`, which improvisation is the only caller to set, and `Room.undo` gates both the budget peek and the `SpendUndo` debit on one `spendBudget := caller != uuid.Nil && !top.freeUndo`. Reasoning is recorded under sub-PR 8 and in [ADR 0033](decisions/0033-ai-bot-seat.md) §8
- [x] Pacing and hard context deadline — shipped, under different names. **The fields are `MinThink` / `MaxThink` on `aiseat.Config` and they are `time.Duration`, not millisecond ints**; `MinThinkMs` / `MaxThinkMs` appear nowhere in the tree (ADR 0033 §10 has the same stale names and wants the same correction). Defaults: `MinThink` 700ms, `MaxThink` 2s, 5s for `strong` via `ConfigFor`. The deadline is a real `context.WithTimeout` around `Policy.Decide`, and expiry, error or an out-of-range index all fall back to the pass move, or to move 0 when no pass is on offer. `MinThink` is measured from enumeration start rather than added on top, so the pacing is a floor on the whole decision, not a sleep after it. Three fields ship alongside that this line did not scope and the runner depends on: `MaxActionsPerWake` (64), `MaxConsecutiveRejects` (3) and `BlockGrace` (4s) — the last holds an active bot's pass out of declare-blockers while any defender still has a legal block, so a bot cannot end the step before a human has clicked
- [ ] **Loop guard: not built.** The runner keeps no memory of what it last played — `mv` is a local and nothing outlives the iteration — so "same non-pass move twice in one priority window" is not detected anywhere in `aiseat`. What shipped instead is keyed on **rejection, not repetition**: after `MaxConsecutiveRejects` (3) consecutive `actions.Dispatch` errors the runner takes the pass if one is on offer and otherwise sleeps, logging `forced pass after repeated rejections`. A bot repeating a legal, *accepted* non-pass move forever never trips it, and `MaxActionsPerWake` is only a soft bound because the runner's own `Apply` notifies the runner's own wake channel. In practice the four-`random` and four-`heuristic` games do not livelock and the soak harness prints a seed when one stalls, so this is a missing safety rail rather than an observed failure — but #89's own "safety rails over cleverness" asked for it, and it is not there

### Sub-PR 4 — lobby + client integration

- [x] `POST /games/{id}/seats/bot` `{tier, deck}` — any seated player at an unstarted table, plus admin; `DELETE .../seats/bot/{player_id}` while unstarted. A spectator's session carries the game ID and is refused: watching is not seating. `GET /bot/options` backs the picker.
- [x] Bot seats carry real decks, so `Lobby.Start`'s `DeckUploaded` gate needs no special case — verified by a test that starts a human + bot table with no gate change
- [x] Seat arithmetic: bots take real seats, so a seated human can add at most **three**; the all-bot table is reachable via the admin path only, and is the test harness rather than a player flow
- [x] `PlayerView.is_bot` / `bot_tier` / `bot_deck`; protocol.ts and `docs/protocol.md` updated. The three fields also ride the engine snapshot and the persisted lobby metadata, so a bot seat survives a deploy and the lobby's restore path relaunches its runner.
- [x] Lobby UI: "Add bot" on an open seat → tier + deck picker → seat appears with a BOT chip and a remove control
- [x] Bot-seat treatment in `PlayerIdentity.svelte` (the identity-first rebuild that replaced `PlayerHeader.svelte`, which no longer exists in the tree): chip, robot avatar mark, thinking pulse while the bot holds priority (respects the S11.5 animations-off setting and `prefers-reduced-motion`; the chip text carries the information when motion is off)
- [x] Tier registry: all four ADR 0033 §6 names declared from the first PR so the API never reshapes, `available:false` on the three with no policy, and an unavailable tier is a 422 rather than a silent downgrade to `random`
- [x] `deploy/Caddyfile` `@api`, `client/vite.config.ts` and the service worker's `API_PATH` all carry the new `/bot` prefix

**Deck source is an interface, not a list.** `aiseat.DeckSource` —
`List() []DeckInfo` plus `Decklist(id) (DeckInfo, string, bool)`
returning decklist TEXT — is what sub-PR 5 implements; this sub-PR
ships one honest placeholder behind it and `main.go` swaps the one
wiring line. Text rather than resolved cards is deliberate: the bot
seat then runs the identical parse → resolve → validate pipeline a
human's upload runs.

### Sub-PR 5 — curated decks + coverage test

- [x] `internal/aiseat/decks/` — **four** decks, one more than this section originally scoped. Aggro (Izzet, Mary Read and Anne Bonny), ramp-stompy (Simic, Tatyova), spell-based control (Esper, Hashaton) — plus **aristocrats** (mono-black, Syr Konrad), which this section deferred on the grounds that S21/S23 were missing. They landed: the sacrifice outlets and death payoffs are in the catalog, and so are the board wipes.
- [x] Build-failing test: every card in every bot deck resolves to a registered `effects.Spec`, with the MDFC back-face keys (`<oracle_id>#1`) explicitly rejected — a card is not covered because its back face registered
- [x] `decks.Lookup(id)` / `decks.All()` / `decks.Load(idx, id)` — sub-PR 4's interface. `Load` runs the same `ParseText` → `Resolve` → `Validate` pipeline a player upload runs, so a bot deck is held to exactly the rules a human deck is
- [x] Offline tests for count, singleton and colour identity; a `CMDCTRL_SCRYFALL_DUMP`-gated test runs the full `deck.Validate` against the real dump (all four decks pass against 117,738 printings)
- [ ] **Voltron / Equipment / Aura: still deferred, but the reason has changed.** The attachment layer is no longer the blocker — the relation shipped in #374 and the first attachments in #379 and #380. There are seven of them (Bonesplitter, Lightning Greaves, Skullclamp, Swiftfoot Boots, Sword of Feast and Famine, Sword of Fire and Ice, Rancor). A Voltron deck wants fifteen to twenty-five. This is now a card-count problem that clears as the catalog grows, with no engine work in front of it
- [ ] Combo: still out, per ADR 0033 §7 — bad idea for a bot regardless of coverage

### Sub-PR 6 — heuristic policy (Layer B)

- [x] `score(view, perspective) float64` — life, hand, board (power + toughness + keyword table), non-creature permanents, untapped mana, commander tax
- [x] Threat ranking across three opponents; aggression rotation after three ineffective turns on one target
- [x] Move selection by `Δscore`; blocks minimise incoming damage subject to not trading up
- [x] Concede heuristic, deliberately conservative
- [x] This is also the fallback under every model failure — it must stand alone
- [x] Import test: no package under `aiseat/` may import `internal/game` (ADR 0033 §3's type gate, promised by `aiseat/policy.go`'s package doc)

**"Δscore on a cloned game" was dropped, deliberately.** Cloning a game
needs a `*game.Game`, and a `Policy` is handed a `protocol.GameView`
and a `[]legal.Move` precisely so that it cannot hold one — that is
the whole hidden-information guarantee in ADR 0033 §3, and the import
test above fails the build over it. The two requirements are
incompatible and the guarantee is the more important of the pair. The
delta is **estimated** in the same units the score uses, from what the
move does to the visible board, rather than simulated. The honest
limitation: with no oracle text on the wire for a single-faced card,
the policy reads a spell's intent from what it may target and what it
costs, not from what it says. `Input.Oracle` is what closes that gap,
and it lands with the model tiers.

**Two additive hooks on `aiseat`**, both because the `Decision{Index}`
contract alone cannot express them:

- `aiseat.Decline` (`Index == -1`) — "none of these". A defender is
  offered blocks *before* priority reaches it, so a blocks-only window
  has no pass to take, and without a decline a bot would have to keep
  declaring blocks until it ran out of creatures. Honoured only when
  the seat does not hold priority; the runner turns it into the pass
  whenever one is on offer, so it can never stall a table.
- `aiseat.Conceder` — conceding is not a legal *move*. `legal`
  deliberately does not enumerate it (a random policy that could
  concede would scoop out of its own fuzzer), so a policy expresses it
  out of band and the runner dispatches it through `actions.Dispatch`
  like everything else.

### Sub-PR 7 — Layer A rules filter + Layer C model policy

- [x] Layer A: `internal/aiseat/rules/` resolves forced/trivial windows with no model call — three rules, each defensible as a fact about the game rather than an opinion about the board: one legal move, only floating mana on offer (casts auto-tap, CR 106.4 empties the pool), and interchangeable copies of one land. **Measured absorption: 91.3%** over 5,207 windows across two four-bot games (90.1% over 7,810 across three), against ADR 0033 §5's 80% floor — `mana-only` carries ~51% of it and `forced` ~38%. The rate is asserted, not just logged, and the same games cross-check every absorbed window against the heuristic: 4,753 windows, 0 disagreements
- [x] Prompt assembly from `Input` only — the board half of the prompt reads `aiseat.Input` and nothing else; the static half is a `DeckProfile` handed to the seat at construction. The existing import test in `aiseat/heuristic` walks the whole subtree, so `rules/`, `model/` and `tiers/` are covered without a second one
- [x] Prompt-cache the static block; per-decision delta is board state + move list. Three tests hold it: the cache breakpoint sits on the last system block and nowhere else, the static half is byte-identical across a whole game's calls, and a reordered decklist does not change a byte
- [x] Cheap model for routine, frontier model on escalation, with all five triggers from ADR 0033 §5. **Measured escalation rate: ~60–70% of surviving windows**, an order above the ADR's "~20%" estimate and driven almost entirely by the combat trigger — every attack and block declaration escalates by design. Recorded rather than tuned away; the thresholds are all in `model.Config`
- [x] Model returns a `Moves` **index**, never an action; out-of-range, malformed, timed out, errored or absent → Layer B. Exercised as a unit table covering all six shapes, and in a whole game by the outage drill
- [x] Tiers: `random`, `heuristic`, `assisted`, `strong` in `internal/aiseat/tiers/`, each with its ADR 0033 §10 `MaxThink`. **`strong`'s "1-ply sim on top-K" is not implemented** and will not be: simulating a move needs a `*game.Game`, which a policy may not hold — the same collision sub-PR 6 hit with "Δscore on a cloned game", resolved the same way. What `strong` buys is the frontier model on every surviving window and a wider candidate list
- [x] Per-decision instrumentation: layer, rule, escalation reasons, model id, latency, model latency, tokens (including cache read/write) and the fallback cause, as a bounded ring plus aggregate counters

### Sub-PR 8 — improvisation, announced

- [x] When the chosen line needs an effect the catalog cannot execute, the bot may use `move_card` / `change_life` / `add_counter` / `mark_damage`, emitted as one bundle. `aiseat.Improviser` is the opt-in policy hook (out of band for the same reason `Conceder` is: a `Decision` names an INDEX into the closed move list, and improvising is by definition doing something that list does not contain). The four verbs are an allow-list — a bundle carrying `concede` or `discard_selection` is refused
- [x] Chat line naming the card, the intended effect, and that it was improvised. Enforced in code, not left to the policy: `Improvisation.Validate` refuses a bundle that will not name its card and its effect, **before** anything is dispatched
- [x] Replay-log tag so bot improvisations are greppable — `SnapshotPayload.annotation` with `tag: "bot_improvisation"`, written inside the same commit as the bundle. Replay-only; no WS frame carries one
- [x] Setting: "show bot reasoning" surfaces `Decision.Reason` in chat. `settings.gameplay.showBotReasoning` (schema v9), server side gated by `aiseat.Config.Narrate`

**`ApplyBundle`, because `Apply` cannot promise atomicity.** `Room.Apply` stashes a pre-mutation clone and drops it when `fn` errors — correct for one dispatch, wrong for four: the third verb failing leaves the first two applied, with no undo entry pointing at them. `Room.ApplyBundle` runs every step under a single hold of the room lock and restores the pre-bundle clone on any failure, so the bundle commits as one seq, one replay line, one undo entry, or not at all.

**The `UndosRemaining` question ADR 0033 §8 left open: cleaning up after a bot is free.** `undoEntry.freeUndo` exempts improvisation bundles from the undoing player's budget. The budget polices the social cost of taking back *your own* move; an improvisation is the bot asserting a rules interpretation the engine could not execute, and a human correcting it is doing maintenance on a catalog gap. Charging for it would make the careful response cost more than the lazy one, in a feature whose whole safety argument is reversibility — and it would leave a player who had spent their turn's undos stuck with someone else's mistake.

**The caller-nil stamp is a power grant.** `uuid.Nil` bypasses `requireCardController` and `ErrPlayerCallerMismatch` — it must, since improvised removal reaches an opponent's board. That is why the announcement is validated ahead of the commit and the verb list is closed: this is the one path where a bot acts with admin authority.

**Left to [#427](https://github.com/krakenhavoc/cmd_and_ctrl/pull/427):** the bot-seat identity surface. `PlayerView.is_bot` / `bot_tier` / `bot_deck`, the BOT chip, and the seat treatment are all still that PR's. This one deliberately does **not** prefix bot chat with `[BOT]`: the `kind` field marks the line as a bot's, cannot be spoofed by a player who names themselves `[BOT] Kess`, and is the thing the seat treatment should key off. Chat also has no UI since S08.5 wave 1, so an announcement had nowhere to land — sub-PR 8 ships a minimal read-only `BotFeed` in the board's attention strip, which a real chat panel subsumes whenever one returns.

### Tests

- [x] Enumerator agrees with `timing.ts` across a table-driven state matrix, both directions (sub-PR 2: `internal/legal/testdata/timing_agreement.json`, asserted from Go and from vitest against one hand-written expectation table)
- [ ] Visibility: policy `Input.View` byte-identical to a human view at that seat; import test enforces the type gate
- [ ] Four `random` bots play to a winner across 20 consecutive unattended runs — no deadlocks, no illegal actions, replays captured
- [x] Four `heuristic` bots play to a winner within 50 turns — 20/20 seeds, 13–20 turns each (`AISEAT_HEURISTIC_GAMES=20`)
- [x] Zero engine-rejected actions across a 100-game randomized run — and across 60 four-`heuristic` and 60 mixed games through the same soak harness (`AISEAT_SOAK_POLICY=heuristic|mixed`)
- [x] `heuristic` beats `random` head-to-head: 40/40 decided games, alternating seats
- [x] Heuristic decision latency: p50 8.5µs, p99 52µs, max 276µs over 2,715 decisions — four orders of magnitude inside the 2s `MaxThink`
- [x] Model-outage drill: Layer C hard-fails, game completes on Layer B, no frozen table. `TestModelOutageDrill` — the endpoint answers 25 calls and then fails forever, the four-bot game plays to a single survivor, the runner's own fallback counter stays at zero, and 2,382 of 2,407 windows were answered without a usable model. A sibling proves the other half of §10: a model that never answers costs its budget and hands over, and the runner still never force-passes
- [x] Human undo of a bot improvisation succeeds without the admin token, and the bundle reverts as one entry — `TestHumanUndoOfBotImprovisationRevertsTheWholeBundle`: a seated caller (their own seat ID, never `uuid.Nil`) pops a three-verb bundle, all three revert, the second `Undo` returns `ErrNothingToUndo` proving it was one entry, and the caller's `UndosRemaining` is unchanged. `TestImprovisationBundleIsAtomic` holds the other direction: a bundle whose last step fails leaves no state change, no seq bump, no chat line and no replay tag

### Exit criteria

1. A player at an unstarted table adds a bot, picks tier and deck, and the seat appears with a BOT chip; Start enables with bot seats counting as deck-satisfied.
2. Four bots play an unattended game to completion with replays captured — the rules-engine fuzz harness exists and runs.
3. A 1v1 human-vs-bot game at `assisted` shows recognisable play: lands, profitable casts, profitable attacks, removal on the biggest threat.
4. Layer A absorbs >80% of priority windows; per-game model spend is measured and recorded, not estimated.
5. Killing the model endpoint mid-game degrades to `heuristic` without a stall.
6. `client/src/lib/timing.ts` no longer reimplements server rules.

**Where they stand**, verified against `b555289`. **(6) is met** — sub-PR 2 deleted the duplication and the two-way agreement test holds it. **(4) is half met**: Layer A absorbs 91.3%, comfortably past the floor, but per-game spend still needs a live endpoint and there is no key in CI, exactly as [ADR 0033](decisions/0033-ai-bot-seat.md) §5's update says. **(1), (2), (3) and (5) are all provable in tests and none of them is reachable from the lobby**, because of the wiring gap noted at the top of this section ([#501](https://github.com/krakenhavoc/cmd_and_ctrl/issues/501)): the seat, the chip and Start all work, but only at `random` and only with the placeholder deck, so "picks tier and deck" is a one-item menu; the unattended four-bot game runs from `TestRandomBotSoak` and `TestFourHeuristicBotsPlayToAWinner` rather than from a table anyone can create; `assisted` cannot be selected at all, which is (3) and the model half of (5). The index row stays `partial` until those are wired — the work to close them is small and none of it is new design.

### Out of scope

- **MCTS / deep lookahead** — the `Policy` slot accepts it later.
- **Any-deck bots** — improvisation from oracle text on uncatalogued cards is a later tier, gated on sub-PR 8 proving itself on curated decks first.
- **Politics, deal-making, table talk** — bots speak only to disclose improvisations and, behind a setting, their reasoning.
- **Learning across games** — stateless between games; no opponent modelling.
- **Bot deckbuilding** — curated only.

### Prior art references

- [Forge AI wiki](https://github.com/Card-Forge/forge/wiki/AI) — rule-based heuristics, ~95% of cards scripted. The "playable but dumb" bar.
- [XMage](https://github.com/magefree/mage) — `ComputerPlayer` target-score evaluation; inspiration for threat-weighted targeting.
- Cowling / Ward / Powley, ["Ensemble Determinization in MCTS for Magic: The Gathering"](https://eprints.whiterose.ac.uk/id/eprint/75050/1/EnsDetMagic.pdf), 2012 — canonical MCTS-for-MTG; shapes the future `MCTSPolicy` slot.

---

## S33 — Surviving a deploy: reconnect, resume, and schema safety

**Phase:** 6 · **Goal:** finish what [ADR 0041](decisions/0041-game-persistence.md) started — a routine deploy costs a live table a page refresh, not the game. Design per [ADR 0044](decisions/0044-surviving-a-deploy.md).

**S32 built the state persistence and it works.** The engine snapshot is schema-versioned and round-trip-exact, `snapshot_drift_test.go` fails CI on an unclassified field, the RNG's stream position rides the file, the invite tokens are persisted, `RestoreFromDisk` runs before the first route is registered and relaunches bot runners, and the data directory is a host path on a dedicated disk that the deploy never touches. A `systemctl restart` really does leave the board on disk.

**Nobody can get back to it.** ADR 0041 rested its client story on one sentence — *"the client already auto-reconnects and resyncs by `seq`, so nothing on the client needs to change"* — and the S33 audit found every clause of it false:

- **The client does not reconnect on a deploy.** `ws.ts` treats close **1000 and 1001 alike as terminal**. That was right when `hub.EvictGame`'s 1000 ("game deleted") was the only deliberate server close; `hub.Shutdown` now sends **1001 on every restart**, and the game is not gone. The backoff ladder immediately below that branch is dead code on the one path it exists for. Worse, `hub.Shutdown` writes the close frame and closes the socket with no flush, so whether the browser sees 1001 or 1006 is a race — **the same binary behaves differently on different deploys**, which is why this read as a flaky network and not as a bug.
- **There is no resync by `seq`.** No resume frame, no `since` field; the client never tells the server what it has seen. `seq` is an ordering value, not a cursor.
- **The session dies with the process anyway.** `auth.MemoryAuthenticator` is an in-process map, so every stored token 401s on the upgrade — which a browser reports as 1006, indistinguishable from a blip.
- **And the documented escape does not exist.** ADR 0041 says players "re-authenticate through their invite link". `Lobby.JoinWithIdentity` refuses any game that has left the lobby, and the route table has no seat-reclaim endpoint. A player whose credentials died can spectate their own game; they cannot play it.

Meanwhile the restore point is staler than the ADR implies. The census fires on **any card with intrinsic abilities and no oracle ID to look them up by**, which is every true token — and `TreasureToken()` sets four pure-data fields and no closures at all. One Treasure on any battlefield suppresses every restore-point write for as long as it sits there. Treasure, Food, Clue and Blood are ordinary catalog output, so in practice many real games have no usable restore point at all.

**Net:** the state survives and the return path was never built. This sprint builds it, and takes the one slice of ADR 0041's phase 3 that decides whether the restore point is minutes stale or turns stale.

**Not a drain.** Production is one systemd unit on one VM binding one port; `systemctl restart` is stop-then-start, so there is no rolling replacement to coordinate with and nowhere to drain to. Durability is already per-`Apply` and a shutdown-time write would be skipped for exactly the games whose restore point is stale. The decision is recorded in ADR 0044 §1 so it is not re-opened as an oversight. What shutdown *does* gain is a log line.

### Sub-PR 0 — ADR 0044 and the sprint docs ([#516](https://github.com/krakenhavoc/cmd_and_ctrl/issues/516))

- [ ] `docs/decisions/0044-surviving-a-deploy.md` — the eight decisions, with rejected alternatives
- [ ] This section and its index row
- [ ] Forward pointer on ADR 0041 recording that its client/`seq` claim is superseded in part

### Sub-PR 1 — durable sessions ([#517](https://github.com/krakenhavoc/cmd_and_ctrl/issues/517))

`auth.MemoryAuthenticator`'s own doc comment has said since S04 that "session is lost on server restart", and ADR 0041 named the swap as a one-line change because `Authenticator` is the only seam the rest of the server sees. It is still one line.

- [ ] `auth.HMACAuthenticator` — Principal signed into the credential, constant-time verify, expiry honoured, no server-side store
- [ ] Key from env; **absent key falls back to `MemoryAuthenticator` with a loud warning** rather than booting with a default
- [ ] `Revoke` semantics under a stateless backend, documented on the type rather than left as a silent `return nil`
- [ ] Wired in `main.go` — rebase against S31's bot wiring ([#514](https://github.com/krakenhavoc/cmd_and_ctrl/pull/514))

### Sub-PR 2 — reconnect when the server restarts ([#518](https://github.com/krakenhavoc/cmd_and_ctrl/issues/518))

**Must not ship ahead of sub-PR 1.** Reconnecting without durable sessions converts a silent `disconnected` into a silent infinite 401 loop, which is strictly worse than today.

- [ ] Split the terminal branch: 1000 stays terminal, 1001 reconnects. Replace the comment with what is now true
- [ ] `hub.Shutdown` flushes the close frame deterministically, so a deploy stops being a coin-flip between 1001 and 1006
- [ ] Re-read the session on each dial instead of replaying the URL captured at mount
- [ ] Expose the attempt count so sub-PR 3 can show it
- [ ] `ws.test.ts` covers 1000 / 1001 / 1006 — it currently tests only `reconnectDelayMs` arithmetic, which is why this regressed unseen

### Sub-PR 3 — the board must admit it is stale ([#519](https://github.com/krakenhavoc/cmd_and_ctrl/issues/519))

While the socket is down the board stays fully rendered and fully clickable, and every click lands in a log store only the bug-report modal reads. No toast, no banner — the one-word status pill is the whole UI.

- [ ] Connection banner for `reconnecting` / `disconnected`, with attempt/retry information and a manual retry
- [ ] Input gated on connection state instead of silently swallowing clicks
- [ ] `sendAction` failures reach `lastError` like every other error
- [ ] `aria-live` on connection-state transitions
- [ ] Wording deliberately distinct from the #266 freeze case, so a real freeze stays reportable as one

### Sub-PR 4 — seat reclaim through the invite link ([#520](https://github.com/krakenhavoc/cmd_and_ctrl/issues/520))

The backstop, not the mechanism — sub-PR 1 handles the common case. This covers cleared storage, a new device, an expired token, a closed tab.

- [ ] A reclaim path for a **started** game: invite token plus proof of seat identity → re-seated in your own seat with a fresh session
- [ ] Seat identity from `Player.DiscordID` / `SeatInfo`, or a per-seat secret in `GameMeta` (already written `0600` because it holds secrets)
- [ ] `Join.svelte` offers "return to your seat" instead of the spectator message, to callers who can prove one
- [ ] Rate-limited and constant-time, like the other invite-token routes. **`JoinWithIdentity` is not widened** — a brand-new seat in a started game is still wrong

### Sub-PR 5 — tokens get catalog keys ([#521](https://github.com/krakenhavoc/cmd_and_ctrl/issues/521))

The parallel track: no dependency on the reconnect chain, and the item that decides how far a deploy rewinds a real table. ADR 0041 called this phase 3's top priority; it is smaller than that framing suggests.

- [ ] A synthetic key namespace (`token:treasure` …) that cannot collide with a real oracle ID — a token *copy* carries the copied card's real ID under CR 707.2 and must keep resolving to the printed card
- [ ] Token templates registered so `CatalogManaAbilities` / `CatalogActivatedAbilities` re-derive their abilities, exactly as they already do for a printed card
- [ ] Census narrowed to abilities the registry genuinely cannot return — **narrowed, not deleted**
- [ ] A board holding a Treasure, a Food, a Clue and a Blood captures a restorable snapshot, and all four come back working

### Sub-PR 6 — nothing tests that yesterday's snapshot decodes in today's binary ([#522](https://github.com/krakenhavoc/cmd_and_ctrl/issues/522))

`snapshot_drift_test.go` forces **classification**, never **compatibility** — both sides of its round-trip property are the same binary. Rename a mirror field and the drift test passes, the round trip passes, and yesterday's file decodes wrong. There is no snapshot fixture anywhere in the tree: `ws/testdata` holds one `protocol.GameView` golden and nothing else.

- [ ] A committed corpus of real restore-point files, restored on every CI run, **appended to and never rewritten**
- [ ] A test that fails when the mirror changes shape without a `SnapshotSchemaVersion` decision
- [ ] Compare `ManaAbilityCount` / `ActivatedAbilityCount` against what the catalog actually returned — capture already records them and restore never reads them, so a card whose entry moved between binaries comes back silently stripped
- [ ] The compatibility rule written down next to the constant

### Sub-PR 7 — rewind is a protocol event ([#523](https://github.com/krakenhavoc/cmd_and_ctrl/issues/523))

`docs/protocol.md` promises `seq` is "monotonically non-decreasing". A restored room resumes from the last clean boundary, so it hands clients a `seq` **lower than the one they already rendered**. It works today only because the client zeroes its watermark on every socket open — an accident, and the "fix" of making that watermark durable would wedge every rewound game permanently.

- [ ] A restore generation on the room, carried in the restore-point file and exposed on the snapshot frame
- [ ] Client discards and re-renders on a generation change; the `seq` guard applies within a generation and becomes safe to trust
- [ ] `docs/protocol.md` corrected — non-decreasing **within a generation**
- [ ] `restorePointFile` sits outside `GameSnapshot` and the drift test does not guard it: extend the guard or say plainly why not

### Sub-PR 8 — deploy observability ([#524](https://github.com/krakenhavoc/cmd_and_ctrl/issues/524))

- [ ] On SIGTERM, one line per live game: clean boundary or not, the census labels if not, and how far back its restore point is. **No new writes** — the shutdown path deliberately stays write-free
- [ ] Boot: restore duration and restore-point age alongside the existing summary
- [ ] Measure the restart window from `journalctl` — both `"shutdown signal received"` and `"server listening"` already exist and are timestamped. The hypothesis is that startup dominates and is mostly the Scryfall index load before bind; confirm with a number and record it on [#515](https://github.com/krakenhavoc/cmd_and_ctrl/issues/515)

**Exit criteria:**

1. A `systemctl restart` with a 4-player game in progress: every client reconnects on its own, re-authenticates with the session it already had, and re-renders the table. No refresh, no invite link, no lost seat.
2. During the restart window every client shows an unambiguous reconnecting state and refuses input.
3. A player who clears `localStorage` mid-game follows their invite link and lands in **their own seat**.
4. A board with a Treasure, a Food, a Clue and a Blood is restorable — the restore point is written on the action after the token is made, not withheld until it is spent.
5. A snapshot committed to `testdata/` before a change still restores after it, in CI.
6. A card whose catalog entry disappears between binaries causes a reported failure, not a silently stripped permanent.
7. `journalctl -u cmd-and-ctrl` answers what a deploy cost without consulting any other source.

### Out of scope

- **ADR 0041 phase 3 in full** — `StackItem.Effect`, `DelayedTrigger.Effect`, the resume frames, the turn-scoped statics and replacements. S33 takes only the token slice, because a token blocks *every* capture after it while the others block only the window they are open in. The rest wants sub-PR 8's census numbers to sequence it.
- **A drain, blue/green, or a second node** — ADR 0044 §1. Revisit only if the topology changes.
- **Persisting the undo stack** — ADR 0044 §6. 32 game clones per room, to preserve the one thing a player can trivially redo.
- **Cutting the cold-start window** — measured in sub-PR 8, attacked later; the fix interacts with `RestoreFromDisk`'s deliberate ordering after the card assets.
- **Reaping abandoned restore points** — they accumulate and re-`ERROR` on every boot forever. Adjacent, separate.

Also found during the audit and filed separately, not sprint scope: [#525](https://github.com/krakenhavoc/cmd_and_ctrl/issues/525) — `pruneOrphanMeta` deletes the lobby metadata of an abandoned game, so ADR 0041's documented roll-back recovery destroys the game (and its replay log) rather than returning it.

---
