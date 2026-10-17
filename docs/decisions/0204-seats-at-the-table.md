# ADR 0204 — Seats at the table

**Status:** Accepted · 2026-10-17 · S73 — Big boards: layout and per-action cost
**Issue:** [#4104](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4104)
**Numbering:** `git fetch --all --prune`, then every `docs/decisions/` file name in the history of all 14 remote heads (`git log --remotes --name-only -- docs/decisions`). The highest in use is **0203**. Reserved on #4104 before any code was written.
**Amends:** [ADR 0120](0120-expand-a-players-board.md) (the expanded board opens beside the avatar; the avatar moves, so "beside" follows it) and [ADR 0176](0176-a-movable-stack-panel.md)'s "Known cost" (the pile on the seam no longer covers avatars).
**Builds on:** [ADR 0077](0077-opponent-board-summary.md) (the reflow avatar card), [ADR 0111](0111-action-dock.md) (the dock owns the bottom right), [ADR 0119](0119-a-stack-you-can-follow.md) and its 2026-10-16 amendment (the pile, the arrival trace), and DESIGN.md §5.

This ADR was written plan-first. No code changed with it. The owner accepted it the same day, calls 1–6 included.

---

## Context

### Owner request and answers (2026-10-17)

While deciding where an opponent's arrival trace should start when the stack pile covers their avatar (ADR 0119 amendment §D), the owner asked: *"should we change the location of the avatars so it also simulates how a player would be in the middle of the 'table' holding their hand in front of them even though monitors are wider than they are tall typically?"* The answers, which are binding:

1. **Everyone moves.** Opponents sit at the centre of their hand's edge; the viewer sits at the centre of their bottom edge.
2. **The avatar is in front of the hand fan.** The fan spreads out behind it, like cards held behind a player. The viewer's own hand stays fully clickable, with the avatar just above its arc.
3. **ADR first, then build.**

### What exists

- **A top-row opponent** (every opponent in the row layout, both with two opponents) is drawn `flipped`: hand fan along the top edge, commander at the top left, piles at the top right, battlefield toward the centre. Its identity (`PlayerIdentity`: portrait, life, counters, keyword badges) is in the panel's `rail` area at the **bottom left**, right on the seam between the rows (`PlayerPanel.svelte`, `.panel.opponent.flipped`).
- **An unflipped opponent** (the quadrant layout's `next` beside the viewer, the one-opponent `across`) has its hand along its bottom edge and its identity at the **top of a right-hand rail**.
- **The viewer** has their hand along the bottom, the action dock at its right end (ADR 0111), piles at the bottom left, and their identity at the **top right** of the panel (`.panel.self.docked`, `grid-area: rail`).
- **The stack pile** docks on the top row's seam (ADR 0176, 2026-10-15). ADR 0176 records as a known cost that it covers the inner corners of the seats beside the seam, which is where the top-row avatars are. ADR 0119's arrival trace treats an avatar under the pile as off screen, so an opponent's spell often gets no trace.
- No ADR fixes where an avatar goes. DESIGN.md §5 fixes the corners: the dock bottom right, the hover zoom top right, the coach card bottom left, the pile on the left or the seam.

---

## Decision

### 1. One rule: the avatar sits at the centre of its hand

Every seat's identity is placed **centred on the edge its hand fan is on**, in front of the fan. It's one rule for every layout, so a seat reads the same way wherever it sits: the player is at the middle of their side, holding their hand.

| Seat | Hand edge | Avatar |
|---|---|---|
| Top-row opponent (`flipped`) | top | top centre, over the card-back fan |
| Unflipped opponent (quadrant `next`, one-opponent `across`) | bottom | bottom centre, over the card-back fan |
| Viewer | bottom | bottom centre, just above the hand's arc |

The identity component itself doesn't change: the same 84px portrait, life chip, counters and keyword badges, the same `data-seat-id` hover and pin target, the same accessible names.

### 2. An opponent's avatar in front of their fan

- The card-back fan stays where it is and keeps its count. The avatar is drawn over its centre, in front of it (above it in z-order inside the panel), so the fan reads as cards held behind the player.
- The fan's centre cards are under the avatar. The hand-count badge moves to the avatar's side so the count is never hidden.
- The commander strip stays at the hand edge's left end, and the piles at its right end. The `rail` area the avatar used is freed, so the battlefield gets that width back.
- The avatar overhangs the hand band toward the battlefield by at most half its height. The battlefield's top-centre keeps a clear notch that size, so no permanent sits under the avatar.

### 3. The viewer's avatar above their hand

- The viewer's hand stays fully clickable: the avatar never covers a card in the fan.
- The avatar sits centred above the fan's arc, straddling the line between the battlefield and the hand band. The battlefield's bottom-centre keeps a clear notch the avatar's size, the same rule as §2.
- The dock keeps the bottom right (ADR 0111) and the piles the bottom left. The top-right `rail` is freed, so the viewer's battlefield gets that width back.
- When the hand is wide enough to reach the centre under the avatar, the fan's arc dips at its centre by the avatar's overhang, so the cards there stay fully visible and clickable.

### 4. What follows the avatar

- **Expanding a board (amends ADR 0120).** The expanded board still opens beside the avatar, never over it. For a top-row opponent "beside" now means **below** the avatar, toward the table's centre, in the wider half of the panel's span, with the same 8px gap and `--stack-pile-clear` rule.
- **The stack pile (amends ADR 0176's known cost).** The pile's default spot is unchanged, but it no longer covers an avatar there. That paragraph of ADR 0176 is superseded.
- **The arrival trace (ADR 0119 amendment §D).** No change in code: an opponent's spell now traces from their avatar at the top of their panel, which the pile no longer covers. The "no trace when covered" rule stays for a pile the viewer has dragged over an avatar.
- **Arrows** to and from players (combat, targeting, the stack) already aim at `data-seat-id`, so they follow the avatar.
- **The attention strip** keeps its left inset clearing the commander strip.
- **The focus layout's summaries and the reflow avatar cards** (ADR 0077, SeatSummary) are unchanged: they are already centred cards.

### 5. Sizes

The layout targets tablet (820px) as the smallest screen (DESIGN.md §1). With three opponents in a row at 820px each top-row panel is about 270px wide. The avatar plus its hand-count badge must fit inside the fan's width there. If a panel is narrower than the avatar plus 2 × 52px (room for the commander strip and the piles either side), the avatar keeps its centre but the commander strip and piles tuck under the fan's ends, never under the avatar.

### 6. Tests and contracts

- Nothing is renamed; the identity keeps its accessible names and `data-seat-id`.
- `board-layout.spec.ts` and any spec that asserts where an identity sits are updated in the same PR; specs that find the identity by role or name are untouched.
- The tutorial's coach copy and anchors that point at "your life" or an avatar are checked and reworded if they name a corner.
- Screenshots at 1024×768 and 1440×900, for two, three and four seats in the row layout and the quadrant layout, with the stack pile live, under War room, Classic and High contrast.

## Delivery

One client PR: the placement in `PlayerPanel.svelte` (and Board's seat templates if needed), the notch, the viewer's arc dip, ADR 0120's expand anchor, and the spec and tutorial updates. A branch E2E before merge.

## Consequences

- Every player sits at the middle of their side, holding their hand, in every layout.
- The seam between the rows is clear of avatars, so the stack pile, the expanded board and the arrival trace stop competing with them.
- Each battlefield gets back the width of the old rail, and loses a small centred notch on its hand side.
- An opponent's fan hides its centre cards behind the avatar; the hand count stays visible.

## Out of scope

- A phone layout (DESIGN.md: tablet is the minimum).
- Changing the avatar's size, the life chip or the counters.
- Moving the dock, the hover zoom, the coach card or the pile's default spot.
- The focus layout's summaries.

## Calls made here

Each can be overturned in review:

1. **One rule for every layout:** the avatar centres on the edge its hand is on (§1), so an unflipped opponent's avatar is at its bottom centre.
2. **The hand count moves to the avatar's side** so the fan's hidden centre never hides it (§2).
3. **A clear notch on the battlefield** under each avatar's overhang, rather than letting permanents slide under it (§2, §3).
4. **The viewer's fan dips at its centre** rather than splitting, so every card stays clickable (§3).
5. **An expanded top-row board opens below its avatar** (§4).
6. **Narrow panels tuck the commander strip and piles under the fan's ends**, never under the avatar (§5).
