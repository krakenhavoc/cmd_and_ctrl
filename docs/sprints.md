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

---

## Sprint index

| # | Name | Phase | Issue | Due | Status |
|---|---|---|---|---|---|
| S01 | Discovery spike | 0 | [#1](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1) | 2026-04-24 | planned |
| S02 | Protocol bridge foundations | 1 | [#2](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2) | 2026-05-08 | planned |
| S03 | Protocol bridge — game actions | 1 | [#3](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3) | 2026-05-22 | planned |
| S04 | Lobby, auth, Scryfall pipeline | 2 | [#4](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4) | 2026-06-05 | planned |
| S05 | Deck import + play area skeleton | 2 / 3 | [#5](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5) | 2026-06-19 | planned |
| S06 | Hand, battlefield, zones | 3 | [#6](https://github.com/krakenhavoc/cmd_and_ctrl/issues/6) | 2026-07-03 | planned |
| S07 | Stack, priority, turn structure | 3 | [#7](https://github.com/krakenhavoc/cmd_and_ctrl/issues/7) | 2026-07-17 | planned |
| S08 | First playable — 2-player in browser | 3 / 4 | [#8](https://github.com/krakenhavoc/cmd_and_ctrl/issues/8) | 2026-07-31 | planned |
| S09 | Polish I — animations + VFX | 5 | [#9](https://github.com/krakenhavoc/cmd_and_ctrl/issues/9) | 2026-08-14 | planned |
| S10 | Polish II — Commander UX (cmd damage, politics) | 5 | [#10](https://github.com/krakenhavoc/cmd_and_ctrl/issues/10) | 2026-08-28 | planned |
| S11 | Polish III — intent preview, undo, spectator | 5 | [#11](https://github.com/krakenhavoc/cmd_and_ctrl/issues/11) | 2026-09-11 | planned |
| S12 | Deploy + 4-player go-live with friends | 6 | [#12](https://github.com/krakenhavoc/cmd_and_ctrl/issues/12) | 2026-09-25 | planned |

---

## S01 — Discovery spike
**Phase:** 0 · **Goal:** confirm Option A (XMage backend) with real information.

- [ ] Clone `magefree/mage` as a submodule or fork under `xmage/`
- [ ] Build XMage server + client locally in the devcontainer
- [ ] Play one game against the built-in AI and capture network traffic
- [ ] Write `docs/xmage-protocol-notes.md` documenting the protocol surface
- [ ] Go/no-go decision on Option A vs B, recorded in PLAN.md section 7

**Exit criteria:** a written decision and a protocol sketch good enough to
design a bridge against.

---

## S02 — Protocol bridge foundations
**Phase:** 1 · **Goal:** scaffold the bridge service and establish the seam.

- [ ] Pick bridge language (Go or TS) — decision recorded in `docs/decisions/`
- [ ] `bridge/` skeleton with lint, format, test, and a `make dev` target
- [ ] WebSocket server on :8080 accepting JSON frames
- [ ] XMage client library or hand-rolled protocol client — stub
- [ ] Connect bridge → local XMage server; log handshake

**Exit criteria:** bridge process starts, accepts a WS connection, and logs an
XMage handshake round-trip.

---

## S03 — Protocol bridge — game actions
**Phase:** 1 · **Goal:** a CLI test client can drive a basic game through the bridge.

- [ ] Translate: join game, draw, play land, pass priority
- [ ] Translate inbound state updates to a JSON schema the client will consume
- [ ] `scripts/bridge-cli` minimal CLI test harness
- [ ] Snapshot tests of a recorded game transcript

**Exit criteria:** scripted CLI plays a full turn against XMage through the bridge.

---

## S04 — Lobby, auth, Scryfall pipeline
**Phase:** 2 · **Goal:** the non-play parts of the web app.

- [ ] `client/` skeleton (React or Svelte + Vite) with lint/format/test
- [ ] Shared-password auth (env var, no user accounts)
- [ ] Lobby page: create/join game, invite link
- [ ] Scryfall bulk download job + weekly refresh script
- [ ] Card image cache layer (on-demand download, disk cache)

**Exit criteria:** log in, create a game, see a lobby; card data and images
resolve by name/id.

---

## S05 — Deck import + play area skeleton
**Phase:** 2 / 3 · **Goal:** bring decks into games and render a 4-player table.

- [ ] Moxfield / Archidekt / plain-text deck import
- [ ] Validation: commander legality, 100-card singleton
- [ ] PixiJS play area canvas wired into the client shell
- [ ] 4-player table layout (you + 3 opponents), no real cards yet

**Exit criteria:** import a deck, join a game, see your seat at the table.

---

## S06 — Hand, battlefield, zones
**Phase:** 3 · **Goal:** render and manipulate the core game zones.

- [ ] Hand: fan layout, drag to battlefield
- [ ] Battlefield: tap/untap, stacking tokens
- [ ] Library / graveyard / exile / command zone — clickable, searchable
- [ ] Zone state driven entirely by bridge updates

**Exit criteria:** play a land, cast a creature, attack — all through the real
XMage backend via the bridge.

---

## S07 — Stack, priority, turn structure
**Phase:** 3 · **Goal:** the fiddly but essential parts of a rules-enforced game.

- [ ] Stack visualization (who cast what, in order)
- [ ] Priority indicator (whose turn to respond)
- [ ] Turn / phase / step indicator
- [ ] "Pass priority" and "pass until end of turn" shortcuts

**Exit criteria:** casting a counterspell in response to a creature works and
is legible.

---

## S08 — First playable (milestone)
**Phase:** 3 / 4 · **Goal:** complete a 2-player game in a browser, end-to-end.

- [ ] Mulligan flow
- [ ] Combat UI (declare attackers/blockers)
- [ ] Win/loss screen
- [ ] Play against yourself across two tabs, record a video

**Exit criteria:** a recorded 2-player game, start to finish, with no manual
state intervention.

---

## S09 — Polish I — animations + VFX
**Phase:** 5 · **Goal:** make the table feel alive.

- [ ] GSAP integration
- [ ] Card draw / play / tap / untap animations
- [ ] Damage number popups, combat arrows
- [ ] Sound pass (draw, play, tap, damage)

**Exit criteria:** side-by-side video vs. XMage's Java client shows a clear
"this feels better" delta.

---

## S10 — Polish II — Commander UX
**Phase:** 5 · **Goal:** the differentiator work from PLAN.md section 5.

- [ ] Command zone as a first-class element
- [ ] Commander damage 4×4 grid, always visible
- [ ] Life tracker starting at 40 with history, poison/infect/energy
- [ ] Monarch / initiative / goad markers
- [ ] Politics UI scaffold (deal buttons, promise tokens)

**Exit criteria:** all Commander-specific affordances from PLAN.md section 5
are at least rough-rendered.

---

## S11 — Polish III — intent preview, undo, spectator
**Phase:** 5 · **Goal:** the Arena-feel features that XMage lacks.

- [ ] Hover a card → preview what it would do
- [ ] One-click undo / rewind to last priority pass
- [ ] Spectator mode (read-only connection)
- [ ] Auto-saved replays

**Exit criteria:** intent preview and undo work for a meaningful sample of
cards; spectating a live game works.

---

## S12 — Deploy and go live with friends
**Phase:** 6 · **Goal:** real games, real feedback.

- [ ] VPS provisioning script
- [ ] Deploy XMage + bridge + client
- [ ] Basic observability (logs, uptime ping)
- [ ] Play a real 4-player game with friends
- [ ] Triage top 10 pain points into S13+ backlog

**Exit criteria:** a real 4-player Commander game happens on the deployed stack.

---

## After S12

Phase 7 is ongoing. The backlog from S12 feedback drives subsequent sprints.
Polish work continues until the project feels premium — PLAN.md estimates
9–12 months part-time to that point.
