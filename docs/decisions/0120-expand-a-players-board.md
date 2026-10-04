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

- **ADR 0077 §6** (`display.expandStyle`, reflow or overlay for the automatic expansions). Unchanged and still inert.
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
