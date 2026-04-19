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
| S13.2 | Counter mechanics (SBAs + player counters + UI) | 7 | [#79](https://github.com/krakenhavoc/cmd_and_ctrl/issues/79) | 2026-06-28 | planned |
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
- [ ] Sound pass (draw, play, tap, damage, turn change) — **deferred** until SFX pack lands; prompts + wiring plan in [s09-sound-pass.md](s09-sound-pass.md)
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
