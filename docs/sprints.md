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
| S07 | Turn/phase UI + chat + manual priority | 3 | [#7](https://github.com/krakenhavoc/cmd_and_ctrl/issues/7) | 2026-07-17 | planned |
| S08 | First playable sandbox (2-player, milestone) | 4 | [#8](https://github.com/krakenhavoc/cmd_and_ctrl/issues/8) | 2026-07-31 | planned |
| S08.5 | Game-logic cleanup pass (wave 1: gating + room + import) | 4 | [#43](https://github.com/krakenhavoc/cmd_and_ctrl/issues/43) | 2026-05-03 | planned |
| S09 | Polish I — animations + VFX | 5 | [#9](https://github.com/krakenhavoc/cmd_and_ctrl/issues/9) | 2026-08-14 | planned |
| S10 | Polish II — Commander UX (cmd damage, politics) | 5 | [#10](https://github.com/krakenhavoc/cmd_and_ctrl/issues/10) | 2026-08-28 | **done** |
| S11 | Polish III — hover preview, undo, spectator | 5 | [#11](https://github.com/krakenhavoc/cmd_and_ctrl/issues/11) | 2026-09-11 | planned |
| S11.5 | Per-user settings and preferences (mini) | 5 | [#82](https://github.com/krakenhavoc/cmd_and_ctrl/issues/82) | 2026-09-18 | planned |
| S12 | Deploy + 4-player go-live with friends | 6 | [#12](https://github.com/krakenhavoc/cmd_and_ctrl/issues/12) | 2026-09-25 | planned |
| S12.5 | Discord identity for players (OAuth + bot + presence) | 6 | [#59](https://github.com/krakenhavoc/cmd_and_ctrl/issues/59) | 2026-10-09 | planned |
| S13 | Priority foundation (rules graft kickoff) | 7 | [#62](https://github.com/krakenhavoc/cmd_and_ctrl/issues/62) | 2026-05-17 | planned |
| S13.1 | Stack: cast/resolve/target/counter/trigger/SBA | 7 | [#63](https://github.com/krakenhavoc/cmd_and_ctrl/issues/63) | 2026-06-14 | planned |
| S13.2 | Counter mechanics (SBAs + player counters + UI) | 7 | [#79](https://github.com/krakenhavoc/cmd_and_ctrl/issues/79) | 2026-06-28 | planned |
| S13.3 | Client-side timing affordance (greyed illegal actions) | 7 | [#99](https://github.com/krakenhavoc/cmd_and_ctrl/issues/99) | 2026-07-04 | planned |
| S13.4 | Interactive cleanup discard + per-player MaxHandSize | 7 | [#103](https://github.com/krakenhavoc/cmd_and_ctrl/issues/103) | 2026-07-11 | planned |
| S14 | Card-effect catalog foundation | 7 | [#64](https://github.com/krakenhavoc/cmd_and_ctrl/issues/64) | 2026-07-12 | planned |
| S15 | Mana pool, cost model, and auto-tapper | 7 | [#65](https://github.com/krakenhavoc/cmd_and_ctrl/issues/65) | 2026-08-09 | planned |
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

## S08.5 — Game-logic cleanup pass (wave 1)
**Phase:** 4 · **Goal:** fix the three frictions that bit hardest in S08 playtest.

Originally deferred post-S08 retro because most candidate items would be subsumed by the S13+ rules graft. Re-activated for **wave 1** when 2-player play surfaced three items that don't overlap with rules work — they're UX/authorization issues that hurt now and the rules engine wouldn't help with later. The broader cleanup (turn-1 skip-draw, commander damage attribution, command-zone tax, token API, mulligan penalty, etc.) stays deferred to S13+.

**Wave 1 scope** ([issue #43](https://github.com/krakenhavoc/cmd_and_ctrl/issues/43)):

- [ ] **Controller-only card interactions.** Server + client gate on `caller == card.controller` for `tap`, `untap`, `move_card`, `add_counter`, `set_battlefield_position`, `declare_attacker`, `declare_blocker`. Admins still bypass.
- [ ] **Battlefield real-estate.** Remove the chat panel from the in-game UI (players use Discord); slim turn bar + toolbar; canvas grows from ~78%×85% to ~99%×92% of viewport. Chat wire-protocol stays intact for future re-mount.
- [ ] **In-game deck import.** Players who land in `Game.svelte` via invite see a deck-import modal when the game is in lobby state and their library is empty — no need to navigate back to the lobby.

**Still deferred to S13+** (rules-graft would reshape these): turn-1 skip-draw, commander damage attribution, partner / companion deck imports, commander returns + tax, the Stack zone, token API, comprehensive move-card cleanup, mulligan penalty.

**Exit criteria:** all three wave-1 items shipped; a second 2-player playtest confirms (a) no cross-player card mutations, (b) the canvas feels uncramped, (c) a brand-new player can join and import without leaving the game route.

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

- [ ] Hover a card → detail panel with full Oracle text and current game context
- [ ] One-click undo / rewind to last server snapshot
- [ ] Spectator mode (read-only connection to a game)
- [ ] Auto-saved replays (server writes every delta to a game log file)

**Exit criteria:** hover preview works for every card, undo works for the last action, spectating a live game works without interfering.

---

## S11.5 — Per-user settings and preferences (mini)
**Phase:** 5 · **Goal:** one canonical settings panel that lets each player customise sound, animations, display, gameplay, and accessibility. Preferences persist in `localStorage`, respect OS-level accessibility hints, and unify the ad-hoc toggles otherwise scattered across S09/S10/S11/S12.5/S13. Lands before S12 deploy so friends' first real games open with configurable defaults.

**Slot rationale.** Ships *after* the three polish sprints (S09 animations + sound, S10 Commander UX, S11 hover / undo / spectator) so the settings panel has real toggles to surface, and *before* S12 deploy so the go-live build ships with settings in place. 1-week mini-sprint scope — the infrastructure is small (single Svelte store + JSON schema + panel UI); most of the surface is wiring existing features up to toggles.

**Storage decision.** Client-only `localStorage`. Matches project ethos (hobby scale, 4–8 users, no DB). Cross-device sync becomes feasible post-S12.5 using Discord ID as the key; documented as a future follow-up, not this sprint.

### Tasks

**Settings infrastructure (client):**
- [ ] `client/src/lib/settings.ts` — Svelte `writable` store, typed `Settings` interface, schema versioning (`cmdctrl.settings.v1`) with a migration function shape ready for future bumps
- [ ] Default values honour `window.matchMedia('(prefers-reduced-motion: reduce)')` at first load; subsequent OS changes are live-observed
- [ ] Reactive apply — changes take effect without page reload
- [ ] Export / import as JSON (copy-to-clipboard + paste) for moving settings between devices without a server
- [ ] Reset-to-defaults button per section + one global
- [ ] Absorb S13's "per-step stops" `localStorage` preferences (issue [#62](https://github.com/krakenhavoc/cmd_and_ctrl/issues/62)) into the same schema with a one-time migration

**Settings panel UI (Svelte):**
- [ ] New `client/src/lib/components/Settings.svelte` — modal panel with sidebar tabs (Audio, Animations, Display, Gameplay, Accessibility, Keybindings, Advanced)
- [ ] Gear icon in `Game.svelte` header + keyboard shortcut `,` (comma); also reachable from `Lobby.svelte` and `Login.svelte`
- [ ] Each setting = label + control (slider / toggle / select / key-capture) + short help text
- [ ] Changes apply instantly (no "Save" button); a subtle "saved ✓" indicator flashes beside the changed control

**Audio (S09 dependency):**
- [ ] Master volume slider (0–100)
- [ ] Effects volume (draw, play, tap, damage, turn change — everything S09 adds)
- [ ] Music volume — disabled with a "no music yet" note if S09 doesn't ship music
- [ ] Mute-all toggle; keyboard shortcut `M`
- [ ] Test button per category that plays a sample sound

**Animations (S09 dependency):**
- [ ] Animations master toggle — auto-off when `prefers-reduced-motion: reduce` on first load; user override persists
- [ ] Per-animation toggles: card draw / play / tap / untap / flip
- [ ] Particle effects on ETB / death (on/off)
- [ ] Damage-number popups + combat arrows (on/off)
- [ ] Animation speed multiplier select: 0.5× / 1× (default) / 1.5× / 2×

**Display:**
- [ ] Theme select — dark (default), light, high-contrast (scaffolded; full theming is a follow-up, but the select + CSS-variable plumbing ships here)
- [ ] Card size on battlefield: small / medium (default) / large
- [ ] Hand layout: fan (default) / stacked
- [ ] Auto-rotate battlefield to viewer POV (resolves issue [#38](https://github.com/krakenhavoc/cmd_and_ctrl/issues/38) if still open — setting defaults to on)
- [ ] Card tooltip hover delay (0–1000 ms, default 300)
- [ ] Show opponent hand count (default on)
- [ ] Show mana pip icons vs. text (icons default)

**Gameplay:**
- [ ] Per-step stops grid (absorbs S13's localStorage prefs). 4 opponents × 10 steps checkbox grid; "stop on my upkeep / opponent's end step / …"
- [ ] Auto-pass priority when the stack is empty and I have nothing playable (convenience — still overridable by holding Shift)
- [ ] Confirm before exiting an active game (default on)
- [ ] Default targeting: "always prompt" vs. "auto-pick if only one legal target"
- [ ] Chat panel visibility — setting exists but no-ops until chat UI returns post-S08.5 removal; wire-protocol is still live
- [ ] Discord Rich Presence toggle — **deferred**: S12.5 ([#59](https://github.com/krakenhavoc/cmd_and_ctrl/issues/59)) adds the feature and will plug the toggle into this panel when it ships

**Accessibility:**
- [ ] Respect `prefers-reduced-motion` (on by default; off = user is overriding OS)
- [ ] Increased text size: 0.9× / 1.0× / 1.2× / 1.5× (applies a root `--font-scale` CSS var)
- [ ] High-contrast mode (overrides theme when active)
- [ ] Color-blind-friendly seat colors (alternate palette in [colors.ts](../client/src/lib/colors.ts))
- [ ] Focus indicators always visible (default on) — bypasses `:focus-visible` suppression

**Keybindings:**
- [ ] Read-only list of default bindings + "Change" button per row — but full re-binding UI (conflict detection, capture) **deferred to a follow-up** unless the sprint has spare budget; ship a JSON-editable `keybindings` setting and a `Reset to defaults` button at minimum
- [ ] Defaults: pass priority (`Space`), untap all (`U`), draw (`D`), settings (`,`), mute (`M`), cancel targeting (`Esc`)

**Advanced:**
- [ ] Export settings to clipboard (JSON)
- [ ] Import settings from clipboard (JSON, validates against schema; rejects with a toast on mismatch)
- [ ] "Copy my settings hash" — short fingerprint for bug reports
- [ ] "Reset all settings" (with confirm)

**Docs:**
- [ ] `docs/decisions/0006-per-user-settings.md` — ADR. Why client-only localStorage (simplicity, hobby scale, no PII leak across the network). Future path: server-synced settings keyed on Discord ID once S12.5 lands.
- [ ] Update [README.md](../README.md) (or add a client-level one) with a "Customising your client" section

**Tests:**
- [ ] Schema migration test: v0 (no prior settings) → v1 loads with defaults; S13-style per-step prefs migrate into the unified schema
- [ ] Default settings load correctly when `localStorage` is empty
- [ ] `prefers-reduced-motion: reduce` at page load disables animations on first open
- [ ] Export → import round-trip produces identical store state
- [ ] Keybinding defaults render without conflicts

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
**Phase:** 7 · **Goal:** model priority + turn-based actions per CR 117 / 502 / 504 / 514.

- [ ] Untap and cleanup steps don't grant priority (sentinel `Turn.PriorityHolder == -1`)
- [ ] Auto turn-based actions (untap, draw, cleanup-discard fire automatically)
- [ ] Eliminated-player skip in priority rotation
- [ ] Turn-1 skip-draw rule (CR 103.7c)
- [ ] Per-step stops UI (client-side localStorage preferences)
- [ ] "Pass to my next stop" client action

**Exit criteria:** a 4-player game runs through full turns with auto turn-based actions; per-player stops work; eliminated seats correctly skipped.

**Out of scope (S13.1+):** real stack, state-based actions, hold-priority modifier, per-commander damage tracking.

---

## S13.1 — Stack: cast / resolve / target / counter / trigger / SBA
**Phase:** 7 · **Goal:** all stack-adjacent items in one sprint.

- [ ] Type helpers (IsLand, IsInstant, IsSorcery, IsPermanent, …)
- [ ] `cast_spell` action; lands route to battlefield; spells to stack
- [ ] Auto-resolve on full priority pass
- [ ] Targeting (announce + re-check on resolve)
- [ ] Modes / X / distribution capture
- [ ] Manual `announce_trigger` + APNAP-ordered queue
- [ ] Activated abilities + loyalty (sorcery-speed, once-per-turn)
- [ ] `counter_spell` and `counter_ability`
- [ ] Hold-priority modifier; split-second flag
- [ ] Commander cast tax (per-commander instance ID — fixes partner-pair collapse)
- [ ] Commander zone replacement (CR 903.9)
- [ ] State-based actions: lethal damage, 0 toughness, 0 life, 21 commander damage, draw from empty library
- [ ] Leaving-game stack cleanup with target scrubbing

**Exit criteria:** a Commander player can cast Lightning Bolt, opponent counters with Counterspell on the stack, full priority/SBA loop works end-to-end.

**Bright line — out of scope (S14+):** auto-fire of triggered abilities from card events, auto-validation of target legality at announce, auto-resolution of spell effects, mana pool, replacement effect engine generally, static abilities, combat keyword effects.

---

## S13.2 — Counter mechanics (SBAs + player counters + UI)
**Phase:** 7 · **Goal:** close out the "counters" surface the engine still has to resolve manually. S13.1 ships the four canonical Commander SBAs; this sprint adds the counter-specific SBAs (planeswalker loyalty, battle defense, +1/+1/-1/-1 cancel, poison player-loss, saga final-chapter), first-class UI treatment of counters, and a shared registry of MTG counter types. Sits between S13.1 and S14 so the effect catalog (S14) can rely on counters being fully modelled.

Gap analysis behind this sprint: `Card.Counters` exists today ([server/internal/game/card.go](../server/internal/game/card.go)) and `CurrentPower()` already applies +1/+1 / -1/-1 to P/T. S13.1 covers the Commander SBAs but explicitly not the counter SBAs. S14's `AddCounters` primitive and S17's replacement engine assume counter data is a first-class shape. Nothing in the roadmap (as of issues [#62](https://github.com/krakenhavoc/cmd_and_ctrl/issues/62)–[#70](https://github.com/krakenhavoc/cmd_and_ctrl/issues/70)) covers player-level counters, counter-specific SBAs, or client pip rendering.

### Tasks

**Counter-specific state-based actions (server):**
- [ ] CR 704.5i — planeswalker with 0 loyalty counters → owner's graveyard (extend `Game.StateBasedActions` from S13.1)
- [ ] CR 704.5p — battle with 0 defense counters → owner's graveyard (covers post-MoM battle cards)
- [ ] CR 704.5q — `+1/+1` and `-1/-1` on same creature: remove N of each, where N = `min(count(+1/+1), count(-1/-1))`
- [ ] CR 704.5c — player with ≥10 poison counters loses the game
- [ ] CR 704.5u — saga with final-chapter lore counter is sacrificed (the SBA half; the lore-counter advance trigger ships in S14+ with the effect catalog)

**Player-level counters (server):**
- [ ] Add `Player.Counters map[string]int` alongside the existing `Life` int
- [ ] `Game.AddPlayerCounter(playerID, name, delta)` mutation with existing zero-drops-key semantics
- [ ] Wire action `add_player_counter` (payload: `{ player_id, name, delta }`)
- [ ] Protocol: extend `PlayerView` with `counters` map; visibility unchanged (counters are public)
- [ ] Snapshot/replay format carries `Player.Counters`

**Counter-type registry (server + client):**
- [ ] `server/internal/game/counter_types.go` — canonical list from MTG comprehensive rules (approx 80 types: +1/+1, -1/-1, loyalty, charge, defense, poison, energy, experience, rad, level, lore, shield, stun, age, charge, flood, ice, time, verse, …). Used for UI iconography and structured-log tagging. Unknown counter names remain accepted — they render as text-only pips.
- [ ] Shared TypeScript mirror in `client/src/lib/counterTypes.ts` (type list + per-type color + optional icon key)

**Marked-damage cleanup (CR 514.2):**
- [ ] `Card.MarkedDamage int` — distinct from counters, separate storage
- [ ] Combat damage step writes `MarkedDamage` instead of mutating counters (currently damage routes through counters or direct toughness inspection)
- [ ] Cleanup-step turn-based action (already auto-fires from S13) clears `MarkedDamage` on every creature
- [ ] Lethal-damage SBA (already in S13.1) reads `MarkedDamage >= Toughness` instead of whatever the S13.1 impl uses — one-line refactor, worth doing here so S14+ don't build on the old shape

**Client UI (Svelte):**
- [ ] Counter pip overlay on battlefield cards: stacked chips at top-right of `CardTile`; chip shows counter name abbreviation + count; color from the type registry
- [ ] Counter inventory popover: right-click card → "Counters" menu → add/remove via per-type rows (common types pinned, free-text name entry for unknowns)
- [ ] Player-level counter panel near `PlayerHeader.svelte`: poison (green/purple drop), energy (yellow bolt), experience (star), rad (radiation icon)
- [ ] `client/src/lib/protocol.ts` — `CardView` + `PlayerView` gain optional `counters` field
- [ ] Animated counter placement (borrow from the S09 GSAP primitives) — pip fade-in on add, fade-out on remove

**Docs:**
- [ ] `docs/decisions/0005-counters.md` — ADR covering counter taxonomy sources, rationale for putting player counters on `Player` vs a separate store, why saga SBA ships here but lore-advance triggers wait for S14, marked-damage split from counters
- [ ] `docs/protocol.md` — document `add_player_counter` action + `counters` fields on views

**Tests:**
- [ ] Go unit tests for each new SBA: planeswalker loyalty 0 → graveyard; battle defense 0 → graveyard; +1/+1 and -1/-1 cancel correctly (including 3×+1 + 2×-1 → 1×+1 + 0×-1); poison ≥10 → loss; saga final chapter → sacrifice
- [ ] Regression test: marked damage clears in cleanup; no carry-over between turns
- [ ] Test that unknown counter names round-trip through protocol without validation errors

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

**Client legality helpers (`client/src/lib/timing.ts` — new file):**
- [ ] `canCastFromHand(card, snap, viewerID): { legal, reason? }` — checks viewer-has-priority + correct step/phase + stack-empty (sorcery) + split-second clear
- [ ] `canActivateAbility(card, ability, snap, viewerID)` — sorcery-speed gate for non-mana activations
- [ ] `canPlayLand(card, snap, viewerID)` — checks land-drops-remaining + main phase + stack empty
- [ ] `canActivateLoyalty(card, snap, viewerID)` — sorcery-speed + once-per-turn from `LoyaltyActivatedThisTurn`
- [ ] Reasons are plain English (not CR citations): "Not your turn", "Stack isn't empty", "Already activated this turn", "Already played a land this turn", "Split second spell on stack"

**Hand component:**
- [ ] Each card runs `canCastFromHand` in a `$derived`; when illegal, applies `.timing-disabled` (opacity 0.55, pointer-events: none, slight grayscale)
- [ ] Hover/zoom still works — only the click affordance is disabled
- [ ] Tooltip on hover shows the legality reason

**Battlefield + ability dialog:**
- [ ] Each ability row in the right-click ability dialog ([S13.1](#s131--stack-castresolvetargetcountertriggersba)) renders disabled with reason if illegal
- [ ] Loyalty abilities greyed with "Already activated this turn" / "Stack isn't empty" / "Not your main phase"

**Toolbar:**
- [ ] `pass_priority` greys when sentinel says no priority OR when viewer doesn't hold priority
- [ ] `play_land` greys when used / not main / stack non-empty
- [ ] `advance_step` / `pass_to_next_stop` follow existing S13 sentinel rules

**Tests:**
- [ ] `client/src/lib/__tests__/timing.test.ts` — table-driven, one scenario per reason string, fake snapshots
- [ ] Manual smoke: 2-tab playtest covering all 8 verification scenarios (sorceries on opponent turn, instants OK, split-second blocking, loyalty re-activation, second land drop, etc.)

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
- [ ] `Player.MaxHandSize int` field (default 7; `-1` sentinel = no maximum)
- [ ] `Turn.DiscardPending map[playerID]int` — cleanup-step-entry sets pending counts for any player over their max
- [ ] `Turn.advance` refuses to leave cleanup while pending is non-empty; re-checks on `discard_selection` resolve
- [ ] `discard_selection` action (`{ card_ids }`) — gated to the pending discarder; validates count + ownership; moves to graveyard; clears pending
- [ ] `set_max_hand_size` action (`{ player_id, value }`) — sandbox helper; admin-only or self-only
- [ ] Replace S13's "discard from hand top" placeholder with the prompt pathway
- [ ] APNAP order for multi-player simultaneous discard (Mindslicer, Painful Quandary)

**Server — wire:**
- [ ] `PlayerView.max_hand_size int`
- [ ] `GameView.discard_pending map<playerID, count>`
- [ ] `KindDiscardPrompt` frame (`{ player_id, count }`)

**Client:**
- [ ] `DiscardPromptModal.svelte` — opens when `discard_pending[viewerID] > 0`; multi-select exactly N cards; non-dismissible; auto-closes when pending drops to 0
- [ ] `PlayerHeader` shows non-default `MaxHandSize` next to hand-count ("8 / ∞" for Reliquary, "3 / 2" for Null Profusion)
- [ ] Optional sandbox numeric input for `set_max_hand_size` (slot into [S11.5](#s115--per-user-settings-and-preferences-mini) settings UI if that lands first)

**Tests:**
- [ ] Server: over-max prompts / selection resolves / wrong count rejected / non-hand card rejected / `-1` skips / admin & self gating
- [ ] Client: `DiscardPromptModal` opens when pending, submits correct payload
- [ ] Manual 2-tab smoke through 5 exit scenarios

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

## S14 — Card-effect catalog foundation
**Phase:** 7 · **Goal:** ~30 of the most-played Commander cards work end-to-end with no manual intervention.

- [ ] Event log infrastructure (typed events, append-only per-Game) — Arena pattern
- [ ] Listener registry pre-wired for S19
- [ ] Card-effect catalog in `server/internal/cards/effects/` keyed by Scryfall ID
- [ ] 15 effect primitives (DealDamage, DrawCards, GainLife, DestroyTarget, …)
- [ ] ~30 starter Commander cards (Lightning Bolt, Sol Ring, Cultivate, Counterspell, Wrath of God, Eternal Witness, Birds of Paradise, Demonic Tutor, …)
- [ ] Client "auto" badge on catalog cards

**Architectural decisions:** Forge-style declarative DSL in Go structs; Cockatrice fallback for unimplemented cards; pull-based event dispatch; Scryfall data for display only (don't parse oracle text into effects).

**Exit criteria:** Cast Lightning Bolt → opponent's life drops by 3 automatically (no manual change_life).

---

## S15 — Mana pool, cost model, and auto-tapper
**Phase:** 7 · **Goal:** server-side mana pool with backtracking auto-tapper that beats Arena on correctness.

- [ ] Cost parser (`{1}{R}`, `{W/U}`, `{X}`, `{W/P}`, `{S}`)
- [ ] Scryfall `produced_mana` + `mana_cost` ingestion
- [ ] `ManaPool` as multiset of tokens; end-of-step empty (CR 106.4)
- [ ] `tap_for_mana` action (distinct from generic `tap`)
- [ ] Cost validation in `cast_spell`
- [ ] Auto-tapper algorithm: backtracking + constraint propagation, <1ms p99
- [ ] Tiebreaker scoring (avoid pain, restriction-bearing mana, utility activation)
- [ ] Auto-tap-and-cast UX with preview, confirm, ESC cancel
- [ ] Lock-tap override (clicking a land first locks it in)
- [ ] Commander tax modifier + static-modifier registry

**Out of scope:** filter lands (Mystic Gate sub-payment), Cavern of Souls tribe-locking, alternative costs.

**Exit criteria:** Cast Cyclonic Rift overload from a 38-land Bant manabase → auto-tap finds plan in microseconds → preview → confirm → resolves.

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

