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

The three hard problems that remain are: **rules engine**, **multiplayer state sync**, and **the UX polish that is actually the point**.

---

## 2. The central architectural decision

One decision dominates everything else: where does the rules engine come from?

### Option A — New client on the XMage backend (recommended)

Reuse [XMage](https://github.com/magefree/mage) as the rules engine. XMage is a Java/Maven project under an MIT license, actively developed (v1.4.58 released October 2025), supports Commander with up to 10 players, and ships **full rules enforcement for over 28,000 unique cards**. Replace only the client: write a thin protocol bridge and a new web frontend.

Pros:
- The hardest problem in the space — the rules engine — is already solved.
- Commander is natively supported.
- MIT license allows anything.
- Inherits 15 years of rules work and new-set support for free.

Cons:
- Coupled to XMage's protocol and release cadence.
- Java server is a deployment wart (fine for personal VPS).
- XMage's abstractions may not expose everything a nice UI wants (animation hooks, intent preview, undo).

MVP effort: **~3–6 months part-time** to a playable, prettier client.

### Option B — Fresh sandbox client (fallback)

Build a Cockatrice-style sandbox from scratch. No rules enforcement. Players move their own cards, tap and untap manually, track life manually. Build Commander-specific UX on top (command zone, commander damage grid, monarch, politics, vote tracking).

Pros:
- Total freedom over stack, UX, and animations.
- Any card "works" instantly — no per-card implementation bottleneck.
- The playgroup already knows the rules; enforcement is not the value-add.

Cons:
- Cannot achieve the Arena "click card → it just works" feel.
- No automatic triggers, stack management, or combat phase automation.

MVP effort: **~2–4 months part-time** to a working Cockatrice-beater.

### Option C — Full custom rules engine (do not do this)

XMage took 15+ years and 50,000+ commits with a team of volunteers to reach 28k cards. A solo developer will not out-build this.

### Recommendation: Option A

"As nice as MTG Arena" implies the "it just works" feel, which requires rules enforcement. XMage is the only realistic path to that as a solo hobbyist. The hybrid move: fork XMage's server, write a thin protocol bridge, and pour all energy into the client and Commander-specific UX layer. That is where the gap in the space actually is.

A safety valve is **Option A with a manual-override escape hatch**: allow players to manually move cards or fix state when the engine gets a weird interaction wrong, so games do not get stuck.

---

## 3. Competitive landscape

| Tool | Rules | Multiplayer | UX | Commander-tailored | Notes |
|---|---|---|---|---|---|
| MTG Arena | Full | 1v1 | Excellent | No (by design) | Unity, WotC. Historic Brawl is the closest; no 4-player. |
| MTGO | Full | Up to 4p | Poor (2004-era) | Yes | The only "official" Commander client; universally complained about. |
| **XMage** | Full (28k) | Up to 10p | Poor (Java/Swing) | Yes | Technically excellent, visually dated. **This project's leverage point.** |
| Cockatrice | Sandbox | Up to 10p | Dated | Partial | C++/Qt, GPLv2. Server-authoritative but no enforcement. Still active (2.10.3, Feb 2026). |
| Forge | Full-ish | Local mostly | Dated | Partial | Data-driven card scripts, strong AI, weak multiplayer. |
| Spelltable | None (webcam) | 4p | OK | Yes | WotC-owned. Paper play over video. |
| Tabletop Simulator | None | 4p | Clunky | Partial | Generic VTT with MTG mods. |

**The gap:** nobody combines rules enforcement, good UX, and Commander-first design. XMage has the first and third; the second is missing. That is the wedge.

---

## 4. Tech stack (assuming Option A)

- **Backend rules engine:** XMage server (Java 21, Maven, MIT). Hosted on a $5–10/mo VPS.
- **Protocol bridge:** thin Go or TypeScript/Node service. Speaks XMage's protocol on one side, WebSocket + JSON on the other. Isolates Java ugliness from the client.
- **Client:** web app. TypeScript + React (or Svelte) + PixiJS for the play area. React/Svelte for menus, lobby, deck manager, chat. PixiJS canvas for the play area.
- **Card data and images:** [Scryfall](https://scryfall.com/docs/api) bulk data ("default cards" JSON, ~200MB), refreshed weekly via cron. Images downloaded on-demand, cached locally.
- **Deck import:** Moxfield and Archidekt exports (open text formats), `.cod`, `.dek`, plain text.
- **Auth:** single shared password or magic-link email. It is 4–8 people. Do not over-engineer.
- **State sync:** already handled by XMage's protocol. Server-authoritative by default.
- **Animations and juice:** GSAP + PixiJS. Card flips, tap/untap, damage numbers, combat arrows, stack visualization.
- **Later packaging:** Tauri or Electron wrap for desktop; Capacitor for mobile. Same codebase.

Alternative stack for Option B: SvelteKit + [Colyseus](https://colyseus.io) (Node.js real-time server). Simpler, far less capable.

---

## 5. Commander-specific UX goals (the differentiator)

XMage has rules but does not have *Commander feel*. Design wins to pursue:

- **4-player table layout** that gives proper spatial sense of opponents. Not a 1v1 view bodged sideways.
- **Command zone** as a first-class prominent UI element, not an afterthought.
- **Commander damage grid** (4×4 matrix) always visible.
- **Life-total ticker** starting at 40, with change history, poison, infect, energy.
- **Politics UI:** deal buttons, promise tokens, voting UI for council's dilemma, goad / monarch / initiative markers.
- **Priority visualization:** clear indicator of who has priority, "any responses?" across 3 opponents.
- **Intent preview:** hover a card to preview what it would do before committing. Huge Arena feature; XMage lacks it.
- **Undo / take-back:** one-click rewind to last priority pass for casual play.
- **Deck stats sidebar:** lands left, cards in hand, average CMC drawn.
- **Chat and reactions:** Discord-tier chat in-game. Table-talk is half of Commander.
- **Replays and screenshots:** auto-saved game logs and a "share this turn" button.
- **Spectator mode:** friends can watch and narrate on voice.

---

## 6. Phased roadmap

Assumes ~10 hours/week part-time.

| Phase | Scope | Rough duration |
|---|---|---|
| **0 — Discovery spike** | Clone XMage, build it locally, play a game, capture protocol traffic. Decision gate: confirm Option A or pivot. | 2–3 weeks |
| **1 — Protocol bridge** | Minimal Go/TS service speaking XMage protocol on one side, WebSocket/JSON on the other. CLI test client plays basic actions. | 3–4 weeks |
| **2 — Lobby and deck import** | React app: login, create game, invite by link, import decks from Moxfield. Scryfall data pipeline. | 3 weeks |
| **3 — Core play UI** | PixiJS play area: hand, battlefield, libraries, graveyards, exile, command zone. 4-player layout. Functional, not pretty. | 6–8 weeks |
| **4 — First playable** | Two-player game end-to-end in browser. Play yourself across two tabs. | milestone |
| **5 — Polish pass 1** | Animations, sound, particles, hover previews, commander damage grid, life tracker. Make it feel good. | 4–6 weeks |
| **6 — 4-player and go live with friends** | Deploy to VPS. Play real games. Iterate on pain points. | 2 weeks |
| **7+** | Ongoing polish driven by what annoys us in real sessions. | ongoing |

Realistic "first real game with friends": **4–6 months part-time**. First "this feels premium": **9–12 months**.

---

## 7. Open decisions

Before writing protocol or UI code, the following need calls:

1. **Option A (XMage backend) vs B (sandbox)?** Strong recommendation: A.
2. **Delivery target:** web browser only, or also desktop (Tauri) or mobile (Capacitor)? Recommendation: web-only first.
3. **Framework preference:** React, Svelte, Vue? Familiarity beats theoretical best.
4. **Comfort with Java:** shapes how thick the protocol bridge needs to be.
5. **Art direction:** Scryfall art as-is (traditional look), or stylize the table, backgrounds, and VFX?
6. **Playgroup shape:** 4 fixed friends or rotating 6–8? Affects whether we bother with a real lobby system.

---

## 8. Immediate next step — Phase 0

1. Project skeleton in place (this document, devcontainer, research notes directory).
2. Clone `magefree/mage` locally and build it.
3. Play one game in XMage's native client; capture network traffic.
4. Write a technical note on XMage's protocol surface. That note becomes the bridge design doc.
5. Commit to Option A vs B with real information.
