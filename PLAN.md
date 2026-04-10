# cmd_and_ctrl — Project Plan

A high-polish digital client for **Magic: The Gathering Commander (EDH)** multiplayer, built for personal use with friends.

---

## 1. Project goal and constraints

**Goal.** Build something that feels as good as MTG Arena but is tailored to 4-player Commander — a format Arena has never supported and MTGO serves poorly.

**Hard constraint: personal use only.** This is a private tool for the author and friends. Not a public product. Not distributed, not monetised, not advertised. This reframes the project significantly:

- Legal risk is low. Cockatrice and XMage have existed for 15+ years in this same territory. Scryfall card images and data, card names, and mana symbols are all acceptable for private use.
- Scale is trivial. Target is ≤8 concurrent users on a single small VPS.
- No content treadmill. New cards can lag official releases by days or weeks.
- Realistic hobby-project scope. Not a startup.

The three hard problems that remain are: **game state and multiplayer sync**, **the UX polish that is actually the point**, and — on a long horizon — **incremental rules enforcement**.

---

## 2. The central architectural decision

One decision dominates everything else: how do we handle the rules engine?

### The three options

**Option A — Reuse XMage as the backend.** Thin protocol bridge to XMage's Java server. Inherits ~28,000 implemented cards, full rules enforcement, commander support, multiplayer sync. *Rejected after S01 discovery* — see §2.2.

**Option B — Sandbox-style client.** Server-authoritative game state, but no rules enforcement. Players tap their own cards, resolve their own triggers, track their own mana. Like Cockatrice, but built around 4-player Commander with better UX.

**Option C — Custom rules engine.** Build rules enforcement from scratch. Forge has been at this 15+ years; XMage the same. A solo hobbyist cannot out-build either.

### 2.1 The decision: **B first, C grafted on incrementally**

Build Option B fully — a polished sandbox client with authoritative multiplayer state and world-class Commander UX. Ship it, play real games on it, and then **layer rules enforcement on top, card by card and mechanic by mechanic, driven entirely by what actually appears at our table.**

This is a deliberate inversion of PLAN.md v1's recommendation. The original plan chose Option A (reuse XMage) on the reasoning that rules enforcement is the hardest problem and should be delegated. That's still true in the abstract — but the coupling cost is real (see §2.2), and a manual-first, rules-grafted-in-later path has properties the original plan undervalued:

- **First playable in ~3 months** instead of ~6. Real games with friends drive every subsequent priority decision.
- **No "stuck" state, ever.** Manual override is the fallback for any interaction the engine doesn't understand yet. Games don't halt because of an unimplemented card — players resolve it by hand and move on.
- **Rules work is prioritised by pain.** You implement the cards and interactions that actually annoyed you last weekend, not a speculative coverage goal.
- **The long tail is small in practice.** A playgroup's 8 decks share maybe 400–800 unique cards. That's a tractable target over years — not 28,000.
- **All original stack.** No JBoss Remoting. No Maven wart on the VPS. Go backend, TypeScript client, one deployable per service.

The risk this accepts: **full rules enforcement may never arrive**, in the sense that XMage-level coverage is years away. That's fine if the sandbox itself is the compelling product — and Option B's analysis already makes that case: the playgroup knows the rules, so enforcement is not the value-add.

### 2.2 Why not Option A (the S01 discovery)

S01 was originally scoped as "clone XMage, play a game, capture protocol, confirm Option A". The clone-and-read-source portion surfaced a blocker before any protocol capture happened:

- XMage's wire protocol is **JBoss Remoting 2.5.4** (dependency declared in `Mage.Common/pom.xml:31`) with **Java object serialization** (`Connection.java:36`: `?serializationtype=java`).
- The "protocol" is remote Java method calls on interfaces like `Session`, `GamePlay`, `PlayerActions` (see `Mage.Common/src/main/java/mage/remote/`), invoked via `TransporterClient.createTransporterClient(...)` in `SessionImpl.java:371`.
- JBoss Remoting 2 was deprecated around 2013. The library is unmaintained. It is not reimplementable in Go or Node as a client — the only practical client is the original Java library.

This forces any XMage-based bridge onto the JVM (Java or Kotlin), which in turn forces a JVM runtime on the VPS alongside Go/Node. The coupling isn't catastrophic, but it imposes a permanent Java dependency on a project whose whole appeal is a clean, modern stack. Combined with the B-then-C path being more playable sooner, Option A is no longer the recommendation.

**Keep as a fallback:** if Option B stalls on game-state or sync problems we haven't anticipated, re-evaluating Option A remains on the table.

### 2.3 Why not Option C (straight custom rules engine)

Same reason the original plan rejected it: 28,000 cards × hand-written rules × 15 years of edge cases is not reachable by one person at 10 hrs/week. Attempting it means zero playability for 2+ years with no user feedback — the classic infra-first trap. **Option C as the starting point is rejected.** Option C *as the destination* — incrementally grown on top of Option B — is the plan.

---

## 3. Competitive landscape

| Tool | Rules | Multiplayer | UX | Commander-tailored | Notes |
|---|---|---|---|---|---|
| MTG Arena | Full | 1v1 | Excellent | No (by design) | Unity, WotC. Historic Brawl is the closest; no 4-player. |
| MTGO | Full | Up to 4p | Poor (2004-era) | Yes | The only "official" Commander client; universally complained about. |
| XMage | Full (~28k) | Up to 10p | Poor (Java/Swing) | Yes | Technically excellent, visually dated. Considered and rejected as a backend in S01. |
| Cockatrice | Sandbox | Up to 10p | Dated | Partial | C++/Qt, GPLv2. Server-authoritative but no enforcement. Closest sibling to what we're building. |
| Forge | Full-ish | Local mostly | Dated | Partial | Data-driven card scripts, strong AI, weak multiplayer. |
| Spelltable | None (webcam) | 4p | OK | Yes | WotC-owned. Paper play over video. |
| Tabletop Simulator | None | 4p | Clunky | Partial | Generic VTT with MTG mods. |

**The gap we're targeting:** a Cockatrice-class sandbox with an Arena-class UX, built from the ground up around 4-player Commander, with a credible long-term path to rules enforcement. Nobody is in that quadrant.

---

## 4. Tech stack

- **Game server:** **Go** (stdlib + [gorilla/websocket](https://github.com/gorilla/websocket) or similar). Authoritative game state, WebSocket+JSON protocol to clients, room/lobby management. Hosted on a $5–10/mo VPS.
- **Protocol:** WebSocket per client, JSON frames (our schema, versioned). Server-authoritative; clients submit actions, server emits state deltas.
- **Client:** TypeScript + Vite. [React](https://react.dev) or [Svelte](https://svelte.dev) for menus, lobby, deck manager, chat. [PixiJS](https://pixijs.com) canvas for the play area and animations.
- **Card data and images:** [Scryfall](https://scryfall.com/docs/api) bulk data ("default cards" JSON, ~200 MB), refreshed weekly via cron. Images downloaded on-demand, cached locally.
- **Deck import:** Moxfield and Archidekt exports (open text formats), `.cod`, `.dek`, plain text.
- **Auth:** single shared password or magic-link email. It is 4–8 people. Do not over-engineer.
- **State storage:** in-memory during a game, JSON snapshots to disk for replays and crash recovery. No database until there's a reason for one.
- **Animations and juice:** GSAP + PixiJS. Card flips, tap/untap, damage numbers, combat arrows.
- **Rules enforcement (B→C path, from S13+):** grafted onto the Go game server incrementally. Each rule lives as a pure function of game state; mechanics ship one at a time. Manual override is the permanent fallback for anything the engine doesn't know yet.
- **Later packaging:** Tauri or Electron wrap for desktop; Capacitor for mobile. Same codebase.

---

## 5. Commander-specific UX goals (the differentiator)

Where MTGO and XMage have Commander but no feel, and Cockatrice has feel but no Commander-specific affordances, the differentiator is doing both.

- **4-player table layout** that gives proper spatial sense of opponents. Not a 1v1 view bodged sideways.
- **Command zone** as a first-class prominent UI element, not an afterthought.
- **Commander damage grid** (4×4 matrix) always visible.
- **Life-total ticker** starting at 40, with change history, poison, infect, energy.
- **Politics UI:** deal buttons, promise tokens, voting UI for council's dilemma, goad / monarch / initiative markers.
- **Priority visualization:** clear indicator of who has priority, "any responses?" across 3 opponents. (In Option B: priority is advanced manually via a pass button until S13+ rules work automates it.)
- **Intent preview:** hover a card to preview what it would do. Starts as "show full Oracle text and current state" in Option B; becomes true intent simulation as the rules engine grows in the B→C phase.
- **Undo / take-back:** one-click rewind to last snapshot for casual play.
- **Deck stats sidebar:** lands left, cards in hand, average CMC drawn.
- **Chat and reactions:** Discord-tier chat in-game. Table-talk is half of Commander.
- **Replays and screenshots:** auto-saved game logs and a "share this turn" button.
- **Spectator mode:** friends can watch and narrate on voice.

---

## 6. Phased roadmap

Assumes ~10 hours/week part-time.

| Phase | Scope | Rough duration |
|---|---|---|
| **0 — Server + client architecture spike** | Pick Go WS stack, scaffold `server/` and `client/`, echo an action through both. Decision gate on framework and protocol schema v0. | 2 weeks |
| **1 — Core game server** | Authoritative in-memory game state: players, zones, turn/phase/step state machine, action protocol, state deltas over WebSocket. No rules enforcement — all actions valid. | 4 weeks |
| **2 — Lobby, auth, Scryfall pipeline** | Create/join game, shared-password auth, weekly Scryfall pull, card image cache. | 2 weeks |
| **3 — Core play UI** | PixiJS play area: hand, battlefield, libraries, graveyards, exile, command zone. 4-player layout. Functional, not pretty. Deck import. | 6–8 weeks |
| **4 — First playable sandbox** | 2-player game end-to-end in browser, played across two tabs. No rules enforcement; players resolve everything manually. Milestone. | — |
| **5 — Polish pass 1** | Animations, sound, hover previews, commander damage grid, life tracker, politics UI, monarch/initiative markers. Make it feel good. | 4–6 weeks |
| **6 — 4-player and go live with friends** | Deploy to VPS. Real 4-player Commander games. Triage pain points into the S13+ backlog. | 2 weeks |
| **7 — B→C rules graft (ongoing)** | Incremental rules enforcement. Start with auto-untap, auto-draw step, auto-tap lands for mana. Grow into combat resolution, target validation, triggered abilities, the long tail of cards. Priorities come from real games. | indefinite |

Realistic "first real game with friends on the sandbox": **~3 months part-time**. First "this feels premium": **~6 months**. First "the engine is taking over meaningful chunks of the rules": **~12+ months**.

---

## 7. Open decisions

Before writing serious code, the following need calls:

1. **Go WebSocket library:** stdlib `net/http` + `gorilla/websocket`, or something higher-level like `nhooyr.io/websocket`. Recommendation: gorilla — boring, battle-tested.
2. **Client framework:** React, Svelte, or Vue? Familiarity beats theoretical best.
3. **Delivery target:** web browser only, or also desktop (Tauri) or mobile (Capacitor)? Recommendation: web-only first.
4. **Art direction:** Scryfall art as-is (traditional look), or stylize the table, backgrounds, and VFX?
5. **Playgroup shape:** 4 fixed friends or rotating 6–8? Affects whether we bother with a real lobby system.

---

## 8. Immediate next step — Phase 0 (S01)

1. Scaffold `server/` (Go module) and `client/` (Vite + TS) with lint, format, test.
2. Draft `docs/protocol.md` — v0 schema for action frames and state delta frames.
3. Echo demo: client sends a `ping` action over WebSocket; server broadcasts a state delta; client renders a change. Proves the seam.
4. Resolve the framework decisions in §7 items 1 and 2, record them in `docs/decisions/`.
5. End of S01: a repo with both services scaffolded, one round-trip working, and a written v0 protocol.
