# ADR 0120 — Expand a player's board on top of the table

**Status:** Accepted · 2026-10-04 · S60 — Table clarity: a stack you can follow
**Issues:** [#2208](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2208) (this change). The S60 stack work is [#2204](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2204) (ADR 0119).
**Owner decisions:** the request and the three answers of 2026-10-04, quoted under [Owner request and answers](#owner-request-and-answers-2026-10-04). They are binding. This ADR also makes eight smaller calls the answers did not cover. They are listed under [Calls made here](#calls-made-here) so the owner can overturn any of them in review before PR 3 lands.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-04. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 39 remote heads. The highest number on any of them is 0117, on `origin/develop`. 0118 is reserved for strict mana by default ([#2188](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2188)). 0119 is being written at the same time for the stack (#2204, branch `docs/s60-adr-0119-stack-you-can-follow`, not pushed when I swept). This ADR takes **0120**.
**Amends:** [ADR 0077](0077-opponent-board-summary.md) §3, the "Pinned" row. The pin becomes this overlay, so a summary is no longer pinned open in place (§6 below).
**Builds on:** [ADR 0077](0077-opponent-board-summary.md) (opponent summaries, and the rule that the table never moves under a click), [ADR 0111](0111-action-dock.md) §4 (the dock's corner, and what must not cover it), [ADR 0117](0117-click-to-act-and-a-per-colour-mana-stepper.md) §1 (the ability popover store, keyed by instance ID), [ADR 0076](0076-tutorial.md) §2.4 (the tutorial's anchors), [ADR 0047](0047-keyboard-shortcuts.md) (Escape belongs to the prompt on screen).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

On a busy four-player table, an opponent's board is small. In the quadrant layout the top row is 0.7fr of 2fr, and the across seats cap their cards at 168px. With summaries on, there are no cards at all. The owner wants a larger, readable copy of one player's board drawn over the table, which still accepts clicks.

Every claim below was checked on `origin/develop` at `92a7e019`. No Comprehensive Rules section matters here: the overlay shows and routes exactly what the table already does.

### Owner request and answers (2026-10-04)

The request, verbatim:

> an additional nicety that expands a players board when you hover over their avatar that draws on top showing an expanded view that is easier to read on busy boards with an option to click and pin it so it stays overlaid while you can then hover over their cards to read still

The answers:

1. **Act through it.** Cards in the overlay behave as on the board: hover zooms them, and while choosing targets or attacking you can click a card, or the player, in the overlay to pick it.
2. **Every seat, yours included.**
3. **Hover opens, click pins.** Hovering the avatar for about 0.4 s opens the overlay. It closes when the pointer leaves both the avatar and the overlay. Clicking the avatar, or a pin button, pins it; clicking again, the pin or Escape unpins it. One pinned board at a time.

### What exists

**The ADR 0077 pin.** `Board.svelte` owns `pinnedSeatID` (`:2389`), a per-table `$state` that is not persisted. It is dropped when its seat leaves (`:2401-2403`). `renderingFor` (`:2405-2421`) feeds it to `decideSeatRendering`, where a pin outranks every other reason (`expansion.ts:111`). A summary's ⤢ button, labelled "Show <name>'s full board" (`SeatSummary.svelte:236-243`), and a click on a creature pip that is not a target or a block (`:203-211`) call `onExpand`, which toggles the pin through `nextPinnedSeat` (`Board.svelte:2489`, `expansion.ts:208`). The pinned seat then renders as a full `PlayerPanel` in its own small slot, with a collapse button over it (`Board.svelte:2494-2507`, CSS `:3143-3159`). That slot is the size the summary was, so the cards are no bigger. ADR 0077 §3's table says the avatar pins, but the code pins only from ⤢ and the pips. `display.expandStyle` (`settings.ts:123`) is still inert: ADR 0077 §6 never picked reflow or overlay.

**The avatar.** `PlayerIdentity.svelte` renders the avatar disc with `data-seat-id` (`:258-273`). The same component is used by full panels (`PlayerPanel.svelte:789`) and summaries (`SeatSummary.svelte:222`). A click calls `handleAvatarClick` (`:82-89`). It picks the player as a target when a targeting prompt allows it, else it declares an attack on that player when the seat can be attacked, else it does nothing. The disc is a focusable `button` only while one of those applies (`:206`, `:261-262`). Otherwise it is a `group` named "<name>, N life". The life, poison and energy steppers are separate buttons beside it, on your own seat only.

**The panel.** `PlayerPanel.svelte` is a `region` named "your board" or "<name> board" (`:650-651`). The e2e suite finds boards by that name. Three call sites in two specs match it as a substring rather than exactly (`s19-triggers.spec.ts:221`, `:303`, `zone-browser.spec.ts:105`, `name: "Opponent board"`), so any second element whose name contains "Opponent board" breaks them. Card height is a share of the panel's height with a clamp: 240px cap on self (`:868`), 200px on an upright opponent (`:951`), 168px on a flipped one (`:979`). A taller panel already gets bigger cards up to its cap, with no new rule. Its click router (`handleCardClick`, `:491`) sends targeting, combat and ability clicks up to Board's handlers through props. It keeps its own token-group modal as local state (`:819-833`).

**Things found by `data-*` attributes.** Card tiles carry `data-instance-id` (`Card.svelte:672`). Avatars carry `data-seat-id`. These readers take the **first** match:

| Reader | Where | Scope |
|---|---|---|
| Combat and stack-target arrows | `CombatArrows.svelte:152-161` (`rectIn`), `:245-262` (`measureTiles`, "first match per ID"), `:314-318` | `boardEl` |
| The fan lane's target arrows and rings | `stackArrows.ts:297-312` (`findTarget`, first one with a size) | the board minus the lane |
| Tutorial spotlights | `tutorialAnchor.ts:33-39` | `document` |
| Popover and picker anchors | `PlayerPanel.svelte:627-640` (`anchorFor`, falls back to `document.querySelector`) | `document` |

A summary's pips carry `data-card-id`, not `data-instance-id` (`SeatSummary.svelte:317`). So today no arrow can reach a summarised seat's creatures.

**One popover per card, drawn by "the one Card" with that ID.** ADR 0117 keeps the open ability popover in `abilityPopover.ts` keyed by instance ID. Every `Card` whose ID matches draws it (`Card.svelte:300`, `BattlefieldRow.svelte:214`). With two copies of a card mounted, both would draw it.

**The layers over the board** (all inside `.board`, which is `position: relative`):

| Layer | Where | z |
|---|---|---|
| Panel chrome, the collapse button | `PlayerPanel.svelte`, `Board.svelte:3147` | 3–10 |
| Combat and stack-target arrows (SVG, no pointer events) | `CombatArrows.svelte:599`, `:638` | 35–36 |
| Floating stack lane (#1467; not the default) | `StackLaneHost.svelte:268`, centred at 40% height | 38 |
| Attention strip, top left | `Board.svelte:3176` | 40 |
| Action dock, bottom right | `ActionDock.svelte:728` | 55 |
| Card-local menus, the mana picker | `ManaSourcePicker.svelte:235` and others | 60–70 |
| Hover zoom, top right, `aria-hidden`, no pointer events | `HoverZoomOverlay.svelte:309-334` | 300 |

The zoom stops above the dock by `--dock-zoom-clear`. It moves left of an open dock sheet by `--zoom-right` (`Game.svelte:2331-2356`). ADR 0119 plans a stack pile at the left edge. Its width is not settled yet.

**Hover timing.** A card writes `hoveredCard` after `display.hoverDelayMs`, default 300 ms (`Card.svelte:492-532`, `settings.ts:366`). The zoom reads that store, so it works for any `Card` anywhere on the board.

**Escape.** Escape belongs to the prompt on screen. The dock's one key handler cancels targeting or a combat selection with it (`dock.ts:388-410`, `dockKeyFor`), and it stands down while a foreign modal layer is open (`modalLayers.ts:60`). Card-local menus close on their own Escape.

**Phones.** Under 600px the dock moves below the board (`Game.svelte:2276`, `:2348`). The board itself has no phone layout.

---

## Decision

### 1. What opens it, and what closes it

The state is Board's, per table, and is never persisted. It replaces `pinnedSeatID`: `expanded = { seatID, pinned } | null`. The transitions are a pure module, `lib/boardExpand.ts`, so they can be unit-tested. Board runs the timers.

- **Hover opens a peek.** The pointer rests on a seat's avatar (the disc, in a panel or a summary) for `max(400 ms, display.hoverDelayMs)`. A peek opens for that seat. It replaces an open peek for another seat. **It does not open while a board is pinned:** there is never more than one overlay on screen.
- **A press consumes the hover.** A `pointerdown` on the avatar cancels the pending peek, and the avatar does not start another until the pointer has left it. Without this, clicking a player to attack or target them would leave the pointer resting there, and a large overlay would open over the board 0.4 s later. No hover timer runs while a button is held, so dragging a card from the hand across avatars opens nothing.
- **A peek closes** when the pointer has been outside both the avatar and the overlay for 250 ms. It also closes on Escape.
- **Click pins.** A click on the avatar first tries `handleAvatarClick`'s intercepts, in their current order: target the player, then attack them. Only when neither applies does it toggle the pin. Picking a player as a target always targets, even with an overlay open (owner answer 1). Pinning another seat moves the pin. A pinned board stays open when the pointer leaves.
- **Unpin closes.** Clicking the avatar again, pressing the pin button, or Escape closes the overlay. The avatar then does not peek again until the pointer leaves it and comes back.
- **Gone seat, gone overlay.** As with ADR 0077's pin, a state whose seat is no longer at the table is dropped by a derivation, not an effect.
- **Mouse only for hover.** Without a hover-capable pointer (`(hover: hover)` false), nothing peeks. The pin button still works.

The avatar keeps its role, its tab stop and its accessible name. The pin-on-click is a pointer behaviour. The keyboard route is the button in §4.

### 2. Where it sits, and how big

The overlay is `position: absolute` inside `.board`, after the slots in DOM order, at **z 32**. That is above all panel chrome (≤ 10) and below the arrows (35–36), the floating stack (38), the attention strip (40), the dock (55), card-local menus (60+), modals (200) and the zoom (300). This is the rule `Board.svelte:3140` already uses for the collapse button: a prompt covering the overlay is better than the overlay covering a prompt.

- **Beside the avatar, never over it.** The overlay takes the wider of the two horizontal spans beside the avatar it was opened from, with an 8px gap. The left span runs from the board's left edge (plus `--stack-pile-clear`, which ADR 0119's pile sets, default 0) to the avatar. The right span runs from the avatar to the board's right edge. Each is capped at 1200px wide and hugs the avatar. Rails sit on the right of each panel, so in practice the overlay opens to the left: a left-hand seat gets about half the board, a right-hand seat nearly all of it. Because the overlay never covers the avatar, the second click lands on the avatar and not on a card, and the pointer crosses only the 8px gap.
- **Full height, above the dock.** The overlay runs from `top: 10px` to `bottom: calc(10px + var(--dock-zoom-clear, 0px))`, the same clearance the zoom uses. The dock, at z 55, is never covered, and nothing in the overlay is ever under it.
- **No scrim.** The table stays visible and clickable around the overlay (owner answer 1). The overlay has an opaque surface, a border and the card shadow, like the zoom.
- **The zoom stays on top.** It is z 300 with no pointer events. Hovering a card in the overlay zooms it as on the table. When a right-hand seat's overlay reaches under the zoom column, the zoom covers that corner while a card is hovered. The top-right seat's own board has the same overlap today.
- **The stack.** The floating lane (38) and ADR 0119's pile sit above the overlay. The pile's left clearance comes from `--stack-pile-clear`.

### 3. What is drawn in it, and how clicks reach the table

The overlay renders a **second `PlayerPanel`** for the seat. It gets the same props Board gives that seat's table panel: the same handlers, combat mode, `legal` and `legalGate`. It also gets a new prop, `expanded`.

- **Cards.** With `expanded`, the panel is upright (never `flipped`) and its `--card-h-max` is 240px for every seat, as on a spectator's panel (`PlayerPanel.svelte:973`). The existing height ramp then sizes the cards. A 4-player across seat goes from at most 168px to 240px on a 900px-tall board. A summarised seat goes from pips to cards. `docked` and `coached` are false in the overlay, because the dock and the coach card belong to the table panel.
- **Act-through, for free.** Clicks in the overlay go through the panel's own router to Board's handlers, exactly as on the table. That covers a target, an attacker, a block, the ability popover, a mana tap, life ± on your own seat, and drag-casting from your own hand. The overlay shows nothing the table does not: it reads the same filtered `GameView`.
- **The overlay's copy is the anchor.** A new helper, `lib/boardAnchor.ts`, answers "the element for this selector" by looking inside an open overlay (`[data-board-expanded]`) first and then in the rest of the root. All four readers in the table under What exists switch to it. While a seat is expanded, arrows, stack-target rings, tutorial spotlights and popover anchors use the overlay's copy. The covered table copy is never measured. When the overlay closes, they fall back to the table. The arrows (z 35–36) draw over the overlay, so an arrow to a card in it is visible. An arrow whose other end is a card covered by the overlay starts from under it. That is accepted. `CombatArrows` and the fan lane also re-measure when the overlay opens, closes, changes seat or resizes, not only on a snapshot tick.
- **One popover.** `openAbilityPopover` records the surface it was opened from, `table` or `expanded`, read from a Svelte context that the overlay sets. A `Card` draws the popover only when both the ID and the surface match. The token-group modal is already local to each panel, so it needs nothing.
- **Names stay unique.** With `expanded`, the panel is a `group` with no name instead of a `region` named "<name> board". The overlay itself is the `region`, named "<name>'s board, expanded", or "your own board, expanded" for your seat. Neither name contains "<name> board" or "your board", so the e2e substring matches still find exactly one board. No existing name changes.

### 4. Controls, keyboard and motion

- **An expand button on each full panel.** It is new, sits at the slot's top-right corner where the collapse button was, and is labelled "Expand <name>'s board" ("Expand your own board" for yours). It shows on slot hover and on focus. It is the keyboard route, and on a touch screen it is the only route: it opens the overlay **pinned** and moves focus into it. A summary already has this button: its existing ⤢, "Show <name>'s full board", does the same (§6).
- **A pin button on the overlay,** top right: "Pin <name>'s expanded board", with `aria-pressed`. On a peek it pins. On a pinned board it unpins, which closes the overlay. When the overlay opened from the expand button, focus returns to that button on close.
- **Escape** closes the overlay, peek or pinned, but only when nothing else owns Escape. The handler acts only when the event is not `defaultPrevented`, no foreign modal layer is open, no ability popover or mana picker is open, and `dockKeyFor` gives the key no action. That keeps cancelling a target pick and closing a card menu ahead of closing the board, as ADR 0047 requires.
- **Motion.** The overlay fades in over 120 ms, scaling from 0.98, only when animations are on and reduced motion is off. Otherwise it appears and disappears at once.
- **Peeks are not announced.** A peek is a pointer aid and takes no focus. The region is in the accessibility tree while it is open, so a screen reader can reach it.

### 5. Spectators and phones

- **Spectators get it too,** read-only, because the panel already handles `spectator`. The spectator grid's slots use the same avatar and expand button.
- **Phones.** Under 600px the overlay fills the board area (inset 0, still above the dock's clearance), and there is no peek. The expand button and the pin button are the whole feature there. A phone layout for the board itself is out of scope.

### 6. ADR 0077's pin becomes this overlay

One pin at a time (owner answer 3) cannot mean two different pins. The in-place pin is retired:

- A summary's ⤢ and its fall-through pip click open the overlay **pinned** for that seat. Their labels do not change.
- `pinnedSeatID`, the `"pinned"` expansion reason, `isPinned`, `nextPinnedSeat` and the slot's collapse button are deleted. `expansion.ts`'s header comment changes from "(b) a pin the viewer asked for" to the overlay, which does not move the table at all.
- The other expansion reasons (targeting, blocking, attacking, active player) are unchanged. They still expand a summary in place.

This keeps ADR 0077 §3's safety rule: nothing on the table opens the overlay, only the viewer's pointer or a key. It does not settle ADR 0077 §6. `display.expandStyle` stays inert. It is not this overlay.

### 7. Tests

**Vitest, new.**

- `boardExpand.ts`: hover past the delay opens a peek. Leaving inside the grace keeps it open, and leaving past the grace closes it. A press consumes the hover. No peek opens while a board is pinned. An intercept beats the pin. Clicking the pinned seat closes. Another seat's click moves the pin. A seat that leaves drops the state. `hoverDelayMs` above 400 raises the open delay.
- `boardAnchor.ts`: the overlay copy is preferred, the table copy is used without an overlay, and an element with no size is skipped.
- Rendered (Board with a 4-seat view): the overlay's region name, a single "<name> board" region while it is open, a card click in the overlay during targeting sends the target, the avatar click during targeting targets and does not pin, Escape with a live targeting prompt cancels the prompt and leaves the overlay, one popover drawn for a card opened from the overlay, the expand button pins and focuses, and the placement never covers the avatar's box.
- `expansion.test.ts` loses its `isPinned` and `nextPinnedSeat` cases.

**Playwright.** One new spec, `board-expand-2208.spec.ts`, on the two-player fixture. It hovers the opponent's avatar and the expanded region appears. It moves away and the region goes. It clicks the avatar and the region stays. It picks a creature in the overlay as a spell's target. It presses Escape and the region goes. The overlay changes what is under a resting pointer, so the nightly E2E runs on PR 3's branch before it merges (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`), and on `develop` after.

---

## Delivery

Each PR goes into `develop`, Sprint S60, Issue #2208.

| PR | What | Needs | Parallel with |
|---|---|---|---|
| 1 | **This ADR** and the AGENTS.md §3 ADR range line. Docs only. | — | anything |
| 2 | **Client plumbing, no visible change.** `lib/boardAnchor.ts` and its four call sites. The popover's surface (`abilityPopover.ts`, `Card.svelte`, `BattlefieldRow.svelte`). `PlayerPanel`'s `expanded` prop (role, card cap, upright). The vitest for these. | 1 | ADR 0119's PRs |
| 3 | **Client: the overlay** (§1–§6). `lib/boardExpand.ts`, the overlay in `Board.svelte`, the avatar's pointer handlers in `PlayerIdentity.svelte`, the expand and pin buttons, Escape, the arrows' re-measure, ADR 0077's pin retired, and a pointer line in ADR 0077 §3. The rest of the vitest in §7, the Playwright spec, and the nightly E2E on the branch. | 2 | — |

If ADR 0119's pile lands first, PR 3 reads its width through `--stack-pile-clear`. If it lands after, ADR 0119 sets that variable. After PR 3: run the nightly E2E on `develop`, check a 4-player table on cmd-dev, and close #2208 with evidence.

## Consequences

- Any seat's board can be read at up to 240px cards without changing the table's layout, and acted on in place.
- A summarised opponent becomes a full board on hover, without the slot growing. ADR 0077's in-place pin, which never made the cards bigger, is gone.
- While a seat is expanded, arrows and spotlights point at the overlay's copy of its cards. A screenshot taken mid-combat can show an arrow start under the overlay.
- Resting the pointer on an avatar now has an effect. The press rule and the 0.4 s delay keep it from firing after a click on the avatar.
- One more full `PlayerPanel` is mounted while the overlay is open. It is torn down on close.

## Out of scope

- **ADR 0077 §6** (`display.expandStyle`, reflow or overlay for the automatic expansions). Unchanged and still inert. *(2026-10-16: reflow is now built, [ADR 0077's reflow amendment](0077-opponent-board-summary.md#amendment-2026-10-16-reflow-4065), and its avatar cards open this overlay.)*
- **A board layout for phones.** §5 gives only the minimum.
- **Expanding the stack, the command zone or a graveyard on their own.** The zone browser and ADR 0119 cover those.
- **Remembering a pin** across games or reloads. It is per table by design.
- **Keyboard shortcuts** for expanding a seat (ADR 0047's dispatcher). Can follow if wanted.

## Calls made here

The owner's answers did not settle these. Each is decided above, and each can be overturned in review:

1. **ADR 0077's in-place pin is retired** (§6). The summary's ⤢ and pip click pin the overlay instead, so "one pinned board at a time" means one thing.
2. **The overlay opens beside the avatar, never over it** (§2), taking the wider side and the full height above the dock. It is not centred, so the second click always lands on the avatar.
3. **No peek while a board is pinned** (§1). There is never more than one overlay on screen.
4. **A press on the avatar consumes the hover** (§1). Clicking a player to attack or target them does not open their board 0.4 s later.
5. **Open delay is `max(400 ms, hoverDelayMs)`, close grace is 250 ms** (§1). A player who slowed the card zoom also slows this. A zero zoom delay does not make the board flash.
6. **Unpinning closes** (§1, §4). There is no separate close button: the pin toggle, the avatar and Escape all close it.
7. **z 32, under the arrows, the stack, the strip and the dock; no scrim** (§2). Arrows draw over the overlay, and prompts are never covered.
8. **Spectators get it, phones get the pin only** (§5).

### Amendment (2026-10-18): even card sizes, a seat strip and a header (#4122)

**Status:** Proposed · 2026-10-18 · S73 — Big boards: layout and per-action cost.
**Issue:** [#4122](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4122). Written
plan-first; no code changed with it. Checked on `origin/develop` at `c2fe92bd1`.

**Owner scope (2026-10-18), binding.**

1. **Layout and card sizes.** The overlay's groups read at an even size, tapped
   cards don't eat the row, there is no stray scrollbar, and the seat's hand,
   avatar and piles sit in a proper strip.
2. **A header naming the seat**: avatar, name, life, the reflow turn chip if the
   seat has one, and the existing pin control.

Out of scope until the owner has seen the result: combat and target arrows,
owner-only chrome such as "+ Promise", and the overlay's position and size on the
table (§2 is unchanged).

**What goes wrong.** In the owner's screenshot, Bot 3's board is expanded at
about 930px tall. Bot 3 has two creatures, one of them tapped, and its lands are
two tapped Plains.

- **The card size follows the overlay's height alone.** The overlay's panel is a
  `.panel.opponent.expanded`, never flipped, so the upright opponent's formula
  sizes its cards: `--card-h: clamp(90px, 52cqh - 33px, --card-h-max)`. §3 capped
  that at 240px. #3032 (for #3028) raised the cap to `100cqh`, which no card ever
  reaches (`PlayerPanel.svelte`, `.panel.opponent.expanded`). So a 930px overlay
  draws a 450px creature. The formula was made for a table panel 300 to 400px
  tall, where its ratios are fine. At 450px they no longer are.
- **The groups use the table's steps.** The back row is `--card-h-sm`, half a
  creature (225px). The commander is `0.42 × --card-h`, capped by the room beside
  an avatar (`--seat-side`). The piles' columns clamp at 52px. At a 450px
  creature that gives the 1 : 0.5 : 0.42 : 0.12 spread in the screenshot.
- **The tapped card takes its turned width at that size.** A tapped pile is
  padded `(h − w) / 2` on each side (`BattlefieldRow.svelte`, #2442), so a turned
  450px creature is 450px wide. `rowFit` shrinks a row only when it overflows its
  width. One untapped and one tapped card (about 770px) fit in about 880px, so
  nothing shrinks, and the two cards fill the row.
- **The scrollbar is a land-strip bug that the overlay's size makes visible.**
  When every land is tapped, the first pile gets `margin-left: (h − w) / 2`
  (`.row.strip .pile.tapped:first-child`). `fitRow` measures the strip's spread
  from `first.offsetLeft`, which leaves that margin out. It then caps the row at
  `--strip-spread`. The row's content is wider than its cap by the margin, 32px
  at a 225px land, and the row's `overflow: auto` shows a scrollbar. The table
  has the same bug, but its smaller lands give a smaller overflow.
- **The hand's edge is a table panel's edge.** ADR 0204 seats the identity at the
  middle of the hand's edge, over a fan whose backs are drawn at the creature
  size, 450px, peeking 30%. The fan's arc lifts its outer backs above the zone
  (`--hand-headroom`, 0.3 of a card), into the back row. The piles squeeze into
  `--seat-side` beside the avatar. The commander is tucked at the right end.
- **Nothing names the seat on screen.** The overlay's name ("Bot 3's board,
  expanded") is only its `aria-label`. The only visible identity is the seated
  avatar at the bottom, so the overlay starts with "Creatures 2".

**The design.** The expanded panel gets a layout of its own, chosen by the
existing `expanded` prop (`.panel.expanded`, from 600px up; §5's phone rule is
unchanged). It is still the same `PlayerPanel` with the same props and handlers,
so act-through, the single popover, `boardAnchor` and the region names of §3
don't change.

```
┌ header ─ (avatar) Bot 3 [BOT] · 40 · markers ········ NEXT UP ··· [pin] ┐
│ Creatures 2                                                             │
│  [card] [card, turned]                                                  │
│ Lands 2        Artifacts 1        Command 1                             │
│  [card][card]   [card]             [commander]                          │
├ strip ─ [lib 90][gy 0][ex 0] ······ (fan of backs) 5 ······ promises ───┤
```

*1. One card size.*

- Every battlefield group draws at one size: creatures, lands, each other
  permanent type, and the commander. In `.panel.expanded`, `--card-h-sm` is
  `--card-h`, so a `compact` row is no smaller than the creature row.
- The size comes from the room the two battlefield rows share, with a cap:
  `--card-h: clamp(96px, (100cqh − header − strip − 96px) / 2 × --card-scale, 240px × --card-scale)`,
  where 96px is the two rows' labels, padding and gaps. That gives about 200px at
  1024×768, and the 240px cap at 1440×900 and 1920×1080. This restores §3's
  240px cap for the overlay only. #3028's uncapped table panels keep their own
  rules.
- The overlay uses the viewer's `--card-scale` for every seat, not
  `--card-scale-opponent`. It is a reading surface, as a spectator's panel is.
- Width is still `rowFit`'s job. A wide row shrinks its cards, wraps them within
  its height budget (#3044), or overlaps them, exactly as on the table. The
  budgets (`allocateRowHeights`) are fed the overlay panel's own height, as now.
  An empty creature row still holds one card's line (#3893), so the back row
  doesn't jump up when the last creature dies.

*2. Tapped cards.* A tapped card stays turned 90° (the same signal as on the
table). At an even size, capped at 240px, it is 1.4 times a card's width, not a
450px slab. `rowFit` already counts the turned width (`tappedRoom`), so a row
with many tapped cards shrinks or wraps sooner. Nothing reorders on a tap,
because moving a card under the pointer is worse than a wide row.

*3. The command zone joins the back row.* In the overlay, `CommandStrip` (and the
revealed companion, if any) is drawn as the last group of the back row, labelled
like the other zones and at the even size. It keeps its accessible name
(`L.commandZone`) and its tax badge. It is no longer tucked half behind the hand's
edge, which is what a table panel needs and the overlay doesn't.

*4. The seat strip.* The bottom row is a strip of fixed height,
`--x-strip-h: clamp(64px, 10cqh, 96px)`, set off from the battlefield by a
hairline (`--border`):

- **Piles** at the left: library, graveyard and exile, with their mono labels
  showing, because there is room. The thumbnails are sized to the strip.
- **The hand** in the middle. The fan's cards are sized so that their peek is
  the strip's height (`--card-h` on the hand zone is
  `--x-strip-h / --hand-peek`), and the fan's headroom is clipped to the strip,
  so at rest no back rises above the hairline. An opponent's face-down fan shows
  its count beside it. Your own fan in your own overlay lifts on hover, as on
  the table (z above the rows), and drag-casting from it still works.
- **Promises** keep their current anchor at the strip's right end. Their chrome
  is out of scope.
- **No avatar.** The seat's one identity is in the header (call 12). The
  ADR 0204 seating is turned off in the overlay: no `--seat-notch` band in the
  back row, and no `seatRoom` measuring. `fieldHeight` measures the panel minus
  the header and the strip.

*5. The header.* A row of about 52px across the overlay's top, on a 9% seat-colour
wash (the `.seat-panel` rule), so whose board it is reads before the name:

- `PlayerIdentity` gets a `header` variant: a 40px disc with its turn and
  priority rings, the name and bot or agent tag on one line, the life total in
  the display face, and the markers (poison, commander damage, monarch,
  initiative) and, on your own seat, the life steppers in a line after it.
  It is the same component, so it keeps `data-seat-id`, the target and attack
  intercepts and the peek handlers. A click on it during targeting still targets
  the player (owner answer 1 of 2026-10-04), and arrows to the player land on it
  through `boardAnchor`.
- **The turn chip**: when the seat is a reflow avatar card, the header shows the
  same mono chip as the card (NEXT UP, AFTER YOU, AFTER <NAME>). It comes from
  `reflowChips`, which Board passes to the overlay's panel as a new `turnChip`
  prop. Otherwise there is no chip.
- **The pin toggle** moves from floating over the panel into the header's right
  end, with the same name, `aria-pressed` and behaviour (§4; call 6 stands:
  unpinning closes). No new control is added.
- The header is not a heading in the accessibility tree. The region's name
  already says whose board it is, and a second name would break §3's uniqueness
  rule for the e2e matches.

**Calls made here** (each can be overturned in review; numbering continues from
the eight above):

9. **One card size for every group, capped at 240px** times the card-size setting.
   The back row and the commander are no longer a step smaller in the overlay,
   and #3032's uncapped ceiling no longer applies to it.
10. **The command zone is a back-row group in the overlay**, at the even size,
    not tucked beside the hand.
11. **Tapped cards stay turned**, at the even size. The other way would be
    upright cards with a "tapped" badge in the overlay only, which keeps rows
    narrow but tells the player something different from the table.
12. **One avatar, in the header.** The strip has none. Two avatars for one seat
    in the overlay would give `boardAnchor`'s first-match readers two
    candidates, and two click targets for picking the player.
13. **The strip is fixed and short**: `clamp(64px, 10cqh, 96px)`, piles, hand and
    promises. The fan never rises above it at rest.
14. **The pin stays the only control**, now in the header. There is no separate
    close button.
15. **The land-strip scrollbar is fixed everywhere.** `fitRow`'s spread includes
    the first pile's left margin, so the table loses the same scrollbar.
16. **The overlay reads `--card-scale` for every seat**, not the opponent scale.

**Product question for the owner.** The scope lists the avatar in both the strip
and the header. Should the overlay show it once, in the header (call 12), or
twice? *Recommended: once, in the header.* A second disc would be decoration, and
it would make the player's own click target and arrow anchor ambiguous.

**Delivery.** One client PR into `develop` (S73, closes #4122), once the owner has
accepted this amendment:

- `PlayerPanel.svelte`: the `.panel.expanded` grid (`head`, `creatures`,
  `middle`, `strip`), the size tokens, the command zone placed in the back row
  when expanded, seating and `--seat-notch` off, and `fieldHeight` less the
  header and the strip.
- `PlayerIdentity.svelte`: the `header` variant.
- `Board.svelte`: `turnChip` passed to the overlay's panel, and the pin moved
  into the header's slot. `.board-expanded` loses the floating pin's placement.
- `CommandStrip.svelte` and `Hand.svelte`: a size that follows the `--card-h`
  they are given (strip size, back-row size), if either still assumes the hand's
  edge.
- `lib/rowFit.ts`: the strip spread measured from the list's content edge, so it
  includes the first pile's margin.
- DESIGN.md §5: one line on the expanded board (a header, even groups, a strip).

Tests:

- **Vitest, unit.** `rowFit.test.ts`: the spread of an all-tapped strip includes
  the first pile's turned overhang, through a pure helper split out of `fitRow`.
- **Vitest, rendered** (`boardExpand.render.test.ts`): the header shows the
  seat's name and life. It shows the reflow chip for an avatar-card seat and none
  for a full-board seat. Inside the overlay there is exactly one
  `[data-seat-id]` for the seat, and it is in the header. The command zone's
  group is in the back row when expanded and beside the hand when not. The pin
  is in the header with an unchanged name. Every existing case (region names,
  act-through, Escape, the single popover, placement) still passes.
- **Stylesheet**, read the way `boardArtPip.test.ts` resolves its cascade: in
  `.panel.expanded`, `--card-h-sm` is `--card-h` and the ceiling is 240px times
  the scale.
- **Playwright** (`board-expand-2208.spec.ts`, extended): with three and with four
  players joined, at 1024×768, 1440×900 and 1920×1080, open an opponent's overlay
  and check:
  - the header names the seat;
  - all of the overlay's battlefield tiles in one row have the same height
    (±1px), and rows that didn't shrink match each other;
  - no element in the overlay has `scrollWidth > clientWidth + 1`;
  - no hand tile at rest has its top above the strip's top.

  The spec attaches a screenshot per size and seat count. The same six
  screenshots, with a tapped creature, all lands tapped and a commander out,
  go in the PR body.
- **The nightly E2E on the branch before merge**
  (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`), because the overlay's
  layout changes what is under a resting pointer. Then on `develop` after
  merge, with a look at a 4-player table on cmd-dev before #4122 is closed.
