# ADR 0111 — Table UX: one action dock, bottom right

**Status:** Accepted · 2026-10-02 · S56 — Table UX: one action dock, bottom right (tracker [#1958](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1958))
**Owner decisions:** 2026-10-02. All five questions are answered, each with the recommended option (a). See [Owner decisions](#owner-decisions-2026-10-02) at the end. The sections and the Delivery plan below are written as decided; the options not chosen are kept as considered options.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-02. I ran `git fetch --all --prune` and read the `docs/decisions/` file names on all 36 remote branches: `origin/develop`, `origin/main`, `origin/docs/adr-0110-remember-me` and 33 chore, docs, feat, fix, repro and wip branches. I also listed every ADR name ever committed on any ref (`git log --all --name-only -- docs/decisions/`) and the files of every open PR. 0110 is taken by the open PR [#1956](https://github.com/krakenhavoc/cmd_and_ctrl/pull/1956) (remember me), which has not merged yet. Nothing anywhere uses 0111, so this one takes **0111**.
**Amends:** [ADR 0076](0076-tutorial.md) §1 (the phase widget and stack bullets) and §2.3 (why the coach card goes bottom-left). [ADR 0009](0009-smart-priority-autopass.md)'s #1307 amendment, "Bluffing" (where the bluff button lives, and when it shows). A pointer line goes into each with the first PR that touches it.
**Builds on:** [ADR 0009](0009-smart-priority-autopass.md) (autopass, the bluff verdict and the "considering" chip), [ADR 0047](0047-keyboard-shortcuts.md) (one dispatcher; Escape and Enter belong to the prompt on screen), [ADR 0105](0105-legal-action-highlights.md) (the ready highlights and their live region) and [ADR 0075](0075-table-settings-and-host-controls.md) (table settings).

This ADR was written plan-first. No client code changed with it. "Decision N" below means the owner's answer N in [Owner decisions](#owner-decisions-2026-10-02). The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

On 2026-10-02 the owner asked for this:

> The bluff feature should be visible in the main screen bottom right. And any button the player should click should also show in the bottom right as well. Not mousing all over the screen.

Everything below was audited on `origin/develop` at `bba24546`.

### Where the buttons are today

The table has one fixed bottom-right widget. It is `PhaseDisplay.svelte`, mounted at `PlayerPanel.svelte:723-737` in the self panel's bottom row (`.grid-bottom`), to the right of the hand and the exile strip and to the left of the 112px player rail (`PlayerIdentity` plus `PileBar`). It holds the turn line, the phase track and four buttons: `next`, `hold`, `bluff` and `autopass` (`PhaseDisplay.svelte:240-302`). The self panel is always in the bottom row: the bottom-right quadrant with three opponents, the full bottom row otherwise (`Board.svelte:2924-2986`). So that widget is always near the bottom right of the screen, one rail's width in from the edge.

Almost everything else a player presses during a turn is somewhere else:

- **The attention strip, top left** (`Board.svelte:2407-2427`, `.strip`, `position: absolute; top: 10px; left: 12px`, z 40). It holds the stack card with its `Pass` and `hold` buttons (`StackOverlay.svelte:172-183`, `:263-271`), then Game.svelte's `attention` snippet (`Game.svelte:1597-1964`): the targeting banner's Done and Cancel (`TargetingBanner.svelte:152-166`), attack with all, Choose attackers… and the declaration Undo (`:1617-1720`), Done blocking / No blocks (`:1726-1749`), the combat-selection Cancel (`:1751-1774`), the insufficient-mana toast's Auto-tap & cast and Cast anyway (`:1832-1853`), the attack-tax and attack-limit toasts' Choose attackers… (`:1854-1918`), and the game-over Back to lobby (`:1942-1957`).
- **The command bar, top right** (`Game.svelte:1205-1509`). Pass turn (`:1240-1249`), mute, game log, settings, the bug button (moved here from the menu by #1952 / PR #1953) and the ⋯ menu (`:1293-1445`). The menu holds Draw, Untap all, Shuffle, Mulligan to N, **Undo** (`:1339-1357`), Life history, Table settings, Spawn, Link Discord, My games, Back to lobby and **Concede** (`:1430-1442`, confirmed by a dialog at `:1448-1462`).
- **The middle of the screen.** Every pending choice in `ChoicePromptModal.svelte`, every cost picker Board.svelte mounts (`:2429-2774`), `DiscardPromptModal`, `AutoTapPreviewModal`, `AttackDeclarationModal` and the mulligan dialog (`Game.svelte:2004-2049`) is a centred overlay. All but the mulligan dialog use the shared `.prompt-backdrop` shell (`app.css:452-560`): fixed, full-viewport, blurred, z 200, swallowing every click. None of them can be minimised to look at the board. `VotingPanel` floats top-centre (z 50) and its launcher sits top-left (z 30).

A normal turn therefore goes: cast from the hand (bottom), confirm the target top-left, pay in the centre, pass bottom-right, attack top-left, pass bottom-right, end the turn top-right, and undo through a menu top-right. That is the mousing the owner describes.

### Bluff, and why the owner has not seen it

Bluffing came with smart autopass's #1307 amendment ([ADR 0009](0009-smart-priority-autopass.md), "Bluffing"). Smart autopass passes a window you cannot answer in one round trip, and holds only when you have a response. So the length of a pause is a tell: a snap pass means "nothing", a pause means "something". A bluff removes the tell. When you have **nothing**, it pauses anyway, so the table sees the same pause either way.

- **Where it pauses.** Two settings pick the windows, and both are off by default (`settings.ts:186-203`, `:394-400`). `gameplay.bluffCounterspell` ("represent a counter") pauses on an opponent's spell or ability on the stack. `gameplay.bluffInstant` ("represent an instant") pauses there too, and also in the combat window and on an opponent's end step.
- **How it pauses.** `gameplay.bluffMode` is `timed` by default: it holds for a random delay drawn uniformly between `bluffDelayMinMs` and `bluffDelayMaxMs` (1500 and 4000 ms), clamped to 500-15000 ms (`bluff.ts:22-39`), then passes. `manual` holds until you click `next`, like a real hold, with no ceiling.
- **Armed.** `bluffArmed` (`bluff.ts:41-64`) is a session switch on top of the settings. At game mount it starts on if either setting is on (`initBluffArmed`, called at `Game.svelte:333`). The `bluff` button flips it for the rest of the game. A bluff needs a setting **and** the switch: `bluffCounter: gp.bluffCounterspell && $bluffArmed` (`Game.svelte:398-401`).
- **The decision.** `autopassDecision.ts` returns a `bluff` verdict in exactly three places (rules 4, 6 and 7, `:168-197`). A bluff only ever replaces a pass. It never beats a guard, a pin, a real answer or any other hold.
- **The timer.** `Game.svelte:283-336` rolls the delay once per frame (`seq`) and shows "bluffing — passes in Ns" or "bluffing — click next" (`bluffStatusText`, `bluff.ts:80-85`). Any other verdict, `next`, any action sent and leaving the game all cancel it (`:412-423`, `:485`, `:748`). When it fires it re-checks the frame and the action count, then passes only if the answer is still a timed bluff or a pass (`:316-326`).
- **What others see.** Nothing that tells them it is a bluff. `considering.ts` shows "{name} is considering a response…" for any human holding more than 800 ms in a response window, from public state and elapsed time only. A real hold, a timed bluff, a manual bluff and a player who has stepped away all look the same (`Board.svelte:317`, `PlayerIdentity.svelte:239`, the stack card's hint at `StackOverlay.svelte:157-160`).

The bluff button renders **only when one of the two settings is on** (`PhaseDisplay.svelte:155-160`, `:272-285`). The Settings fieldset that turns them on is disabled while smart autopass is off (`Settings.svelte:806`). Both settings default off. So on a default install the bluff control is not on the screen at all, which is why the owner asks for it to be visible.

Bluffing is moot with smart autopass off. Rule 6 then holds on every opponent stack item and the empty-stack key windows follow the stops grid, so a pause tells nobody anything.

### Keyboard

`shortcuts.ts:105-208` binds: Space pass priority (it defers to a focused button, `chordDefersToFocus`), `t` pass turn, `h` hold, `Shift+p` autopass, `a` attack with all (inert with more than one opponent), `u` undo, `d` draw, `l` log, `m` mute, `,` settings, `?` help. **There is no bluff key.** Escape, Enter and Tab are reserved (`RESERVED_CHORDS`, `:371-377`): every prompt owns Escape and Enter. Game.svelte's window handler cancels targeting on Escape and confirms a multi-target pick on Enter, but only while no modal layer is open (`Game.svelte:2135-2160`, `modalLayers.ts`). PhaseDisplay and the command bar print the bound key in each button's tooltip (`keyHint`).

### Tests that click these controls

Playwright runs nightly only (`.github/workflows/e2e-nightly.yml`, job `playwright`, at 1280×720 Desktop Chrome). PR CI typechecks `tests-e2e` and runs the client vitest suite. No e2e test checks a position, a size or a phone viewport. What a move can break:

- **Class selectors.** `attack-all-318.spec.ts:139,148,165` finds the cluster by `.att.attack-all` and its buttons inside it. `autopass-blockers-328.spec.ts:105` uses `button.action.autopass` plus `aria-pressed`. `state-freeze-266.spec.ts:81,83` and `autopass-blockers-328.spec.ts:189` read `.step-label` and `.turn-no` with `.first()`, so a second copy earlier in the DOM would be read silently.
- **Dialog scoping.** `mulligan.spec.ts:63,71,80` finds Mulligan and Keep hand inside `getByRole("dialog", {name: /keep or mulligan/i})`. `state-freeze-266.spec.ts:95-104` finds `.prompt-count`, `button.card-pick` and Discard inside the "Discard N card" dialog. `s19-triggers.spec.ts:510-511` finds "Take" inside the search dialog.
- **Names are a contract.** `next` is matched exactly (`state-freeze-266.spec.ts:137,156`, `s19-helpers.ts:751`); "pass turn", "Keep hand" and "Mulligan" by substring; `/^Yes$/`, `/^No$/`, `/^Pay \{2\}$/` and `/^Don't pay$/` by anchored regex (`s19-triggers.spec.ts`).
- **Disabled, not removed.** `full-game.spec.ts:87-109`, `mulligan.spec.ts:87-91` and the `next` checks call `isEnabled()` / `toBeDisabled()`. `next` and Pass turn must be rendered disabled for a seat that cannot press them, not hidden.
- **Duplicates fail.** The page-wide locators run in strict mode, so a second "next" or "Keep hand" anywhere throws. `s19-triggers.spec.ts:608-613,694` also asserts that Yes, No and Pay are **absent** on the opponent's page.
- **Vitest, on every PR.** `gameToolbarBug.render.test.ts:130-156` pins settings, the bug button and the ⋯ menu inside `.bar-icons` and their order. `phaseDisplay.extraTurn.render.test.ts` and `targetingBanner.render.test.ts` mount the components this ADR moves.

Untested today: hold, bluff, the phase pins, declare blockers, damage assignment, auto-tap confirm, the targeting buttons, undo, concede, the ⋯ menu and the keyboard shortcuts.

### The tutorial

Of [ADR 0076](0076-tutorial.md)'s six sub-PRs, two have merged: the event bus (#1077, PR #1933) and the practice table (#1078, PR #1940). The coach card, the scrim and the anchor resolver ([#1079](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1079)) and the nine middle steps ([#1081](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1081)) are open, and no step script exists on any branch. Spotlights are to anchor to the `aria-label`s that `board-layout.spec.ts` asserts (ADR 0076 §2.4). Step 8 ("move the turn along") anchors to "the phase widget" with no selector yet. Step 9 anchors to "the attention strip", which has no `aria-label`. §2.3 puts the coach card bottom-left *because* "the phase widget owns bottom-right". `tutorialBus.ts`'s three events (hand hover, pile hover, ability menu) are untouched by anything here. So the dock changes no tutorial code. It changes two anchors that are not built yet, and it should give them stable labels before #1079 lands.

### What overlaps the bottom right

- **The hand** is to the left of PhaseDisplay in the same row, and lifts upward on hover (`Hand.svelte`). The drag-to-cast zone is a fixed full-board rect (`Hand.svelte:1011-1025`) and is unaffected.
- **The player rail** (`PlayerPanel.svelte:742-772`) is to the right, full panel height: identity at the top and the four piles (library, graveyard, exile, command zone, `PileBar.svelte`) at the bottom. It already scrolls on short panels (`:998-1001`).
- **The hover zoom** (`HoverZoomOverlay.svelte:308-330`) is pinned top-right, up to `calc(100% - 20px)` tall, z 300, `pointer-events: none`. A long card's panel reaches the bottom right.
- **The game log drawer** (`GameLogPanel.svelte:116-124`) is fixed to the right edge from the command bar to the bottom of the window, z 30.
- **The dev dock** (`DevDock.svelte:117-121`) is fixed bottom-left, dev deployments only.

### Phones

The board has no phone layout. Below 600px only the command bar changes (#1953: the wordmark, separator and seq counter hide), plus the mulligan grid and the long attack-limit toast. Card heights floor at 168px in the self panel (`PlayerPanel.svelte:823`). Modals cap at `min(760px, 100vw - 32px)`. Building the board a phone layout is not this ADR's job. The dock must still fit at 390px and must not hide anything it sits over.

---

## Inventory

**Primary** means the game waits for you to answer: a pending choice owed by you, a mulligan, a discard to hand size, a block declaration you owe, or `next` when you hold priority. **Flow** means a step of something you started: a cast's targets, costs and payment, or a combat selection. **Secondary** means an option you may take. "Phone" is the same place at 390px unless the row says otherwise, because nothing else has a phone layout.

| Control | Where today (component; region) | When it shows | Kind | Goes where |
|---|---|---|---|---|
| `next` (pass priority) | PhaseDisplay `:241-252`; self panel bottom row, left of the rail | Always; disabled without priority | Primary | Dock action bar, primary slot |
| Stack `Pass` | StackOverlay `:263-271` and StackLaneHost `:148`; top-left strip or floating lane | Stack non-empty and you hold priority | Primary (duplicate of `next`) | Removed. The dock's `next` is the one pass button |
| `hold` | PhaseDisplay `:259-271`; mirrored on the stack card, StackOverlay `:172-183` | Always; stack mirror while the stack is live | Secondary toggle | Dock toggles row. The stack mirror is removed |
| `autopass` | PhaseDisplay `:286-301` (paused state and loop notice `:328-333`) | Always | Secondary toggle | Dock toggles row; loop notice in the dock's status line |
| `bluff` | PhaseDisplay `:272-285`; status line `:304-306` | Only when a bluff setting is on | Secondary toggle | Dock toggles row, **always shown** (§5) |
| Phase-icon pins (one-time stops) | PhaseDisplay track `:200-226` | Always | Secondary | Dock header (moves with PhaseDisplay) |
| Step stops grid, smart autopass, respond categories | Settings `:709-804` | Settings | Configuration | Stay in Settings |
| Pass turn (`pass_turn`) | Command bar `Game.svelte:1240-1249`; top right | Always; disabled unless active | Secondary | Dock action bar, secondary slot |
| `advance_step` / pass-until | Not sent by the client (only `gamecli`) | — | — | Nothing to move |
| Attack with all / per-opponent buttons | Strip, `Game.svelte:1617-1708` | Declare attackers, something can attack | Secondary (the step continues with `next`) | Dock step row |
| Choose attackers… | Strip `:1657-1705`; refusal toasts `:1854-1918` | Tax or limit applies; refusal | Secondary | Dock step row; the picker becomes a dock sheet |
| Declaration Undo | Strip `:1709-1718` | You declared, undo budget left | Secondary | Merged into the dock's one Undo |
| Done blocking / No blocks | Strip `:1726-1749` | You owe a block declaration | Primary | Dock action bar, primary slot |
| Combat-selection Cancel | Strip `:1751-1774` | An attacker or blocker is selected | Flow | Dock action bar, secondary slot, with the hint as the prompt line |
| Attack / block targets (seat medallion, attacker card) | `PlayerIdentity.svelte:265`, the board | During a selection | Board answer | Stay on the board |
| Combat damage assignment | ChoicePromptModal `damage_assignment` `:1636-1727`; centre modal | Multi-blocker damage | Primary | Dock sheet; "Deal damage" in the action bar |
| Keep hand / Mulligan | Mulligan dialog `Game.svelte:2004-2049`; centre | Mulligans open, your hand not kept | Primary | Dock action bar; the hand shows in a dock sheet (decision 2) |
| Bottom-N | Not implemented (`Game.svelte:2023`, `server/internal/game/mutations.go:8772`) | — | — | When it lands, a dock sheet |
| Opening-hands roll-call | Strip `:1776-1806` | Mulligans open | Status | Stays in the strip |
| Yes / No family: `trigger_prompt`, `optional_replacement`, `confirm`, `may_cast`, `entry_pay_life` | ChoicePromptModal `:1380-1447`, `:1538-1553`; centre modal | Pending choice | Primary | **Inline in the dock** (§2) |
| `pay_unless` without card or tap picks | ChoicePromptModal `:1554-1635` | Pending choice | Primary | Inline: "Don't pay" / "Pay N" |
| `pay_unless` with `pay_cards` or a waterbend tap list | Same | Pending choice | Primary | Dock sheet for the picks, Pay / Don't pay in the action bar |
| `coin_call` | `:1260-1283` | Pending choice | Primary | Inline: Heads / Tails / Stop |
| `loop_shortcut` (CR 732) | `:1284-1318` | Pending choice | Primary | Inline: number field, Stop here / Resolve N more |
| `option_pick`, `entry_controller` | `:1448-1490` | Pending choice | Primary | Inline when every option is a short label; a sheet when an option embeds cards |
| `mana_pick`, `choose_color` | `:1223-1259` | Pending choice | Primary | Inline symbol row (keys 1-9 kept) |
| `choose_creature_type`, `choose_card_name` | `:1319-1379` | Pending choice | Primary | Dock sheet (filter field), submit in the action bar |
| `mode_pick` | `:1491-1537` | Pending choice | Primary | Dock sheet, Choose in the action bar |
| `scry`, `surveil`, `look_at_top`, `put_in_library` | `:1071-1222` | Pending choice | Primary | Dock sheet, Done in the action bar |
| `trigger_order`, `replacement_order` | `:1728-1776` | Pending choice | Primary | Dock sheet |
| Card-grid kinds: sacrifice, search, copy, choose cards, untap, reveal, entry discard / sacrifice, permanents, source | `:1777-1957` | Pending choice | Primary | Dock sheet, the kind's verb in the action bar |
| `pick_target`, `legend_rule`, `choose_protector`, `retarget` | Board-answered (`boardAnsweredChoice.ts`), TargetingBanner | Pending choice | Primary | Board clicks; Done in the dock |
| Discard to hand size | DiscardPromptModal `:79-124`; centre | Cleanup, over hand size | Primary | Dock sheet, Discard in the action bar |
| Targeting Done / Cancel | TargetingBanner `:152-166`; strip, top left | `$targeting` set | Flow | Dock prompt line and action bar; the board still answers |
| Cost pickers: X, Phyrexian, sacrifice, discard, exile, tap, crew, teamwork, blight, delve, counters, alternative costs, alt-cost payment, face, modes, divide | Board.svelte `:2429-2774`; centre modals | During a cast or activation | Flow | Dock sheet; Cancel / confirm in the action bar |
| Insufficient mana: Auto-tap & cast / Cast anyway | Strip toast `:1832-1853` | A strict cast refused | Flow | Dock prompt; the auto-tap preview becomes a dock sheet |
| Auto-tap preview Cast / Cancel | AutoTapPreviewModal `:147-243`; centre | After Auto-tap & cast | Flow | Dock sheet, Cast in the action bar |
| Card right-click menu, mana-ability menu, mana-source picker | CardContextMenu, ManaAbilityMenu, ManaSourcePicker; at the card | A click on a card | Card-local | Stay at the card (decision 5) |
| Token-group actions | TokenGroupModal `PlayerPanel.svelte:773`; centre | A click on a token group | Card-local | Stay as they are; revisit once the sheets (Delivery PR 6) have landed |
| Stack `Counter` | StackOverlay `:251-262`, lanes | Stack item | Item-local | Stays on its stack item |
| Life, poison, energy +/−; monarch, initiative | PlayerIdentity `:297-433`; rail | Always (sandbox) | Seat-local | Stay on the seat |
| Promises | PromisesRow on opponent panels | Always (sandbox) | Seat-local | Stay |
| Library draw, pile browse | PileBar; rail | Always | Zone-local | Stay |
| Vote options, end vote | VotingPanel `:70-131`; top centre | A vote is open | Primary | Inline in the dock (option buttons) |
| Vote launcher | VotingPanel `:143-147`; top left | Always (sandbox) | Secondary | Dock overflow (⋯) |
| Undo (global) | ⋯ menu `Game.svelte:1339-1357` | Always; disabled without budget | Secondary | Dock toggles row, with the count left |
| Sandbox: draw, untap all, shuffle, mulligan to N, life history | ⋯ menu `:1309-1364` | Always | Secondary | Dock overflow (decision 3) |
| Table settings, Spawn, Link Discord, My games, Back to lobby | ⋯ menu `:1366-1428` | Always | Secondary | Dock overflow (decision 3) |
| Concede | ⋯ menu `:1430-1442`, confirm `:1448-1462` | Always | Secondary, irreversible | Dock overflow, last, with its confirm |
| Back to lobby (game over) | Strip `:1954-1956` | Game ended | Primary | Dock action bar, primary slot; the banner stays in the strip |
| Mute, game log, settings, bug, Lobby link | Command bar `:1206`, `:1251-1288` | Always | App chrome | Stay in the command bar |
| Rejection, rewind and reveal toasts (dismiss ×) | Strip | On an event | Status | Stay in the strip |
| "Considering a response…" | PlayerIdentity, StackOverlay hint | Another human holds > 800 ms | Status (others) | Stays on their seat and the stack |

---

## Options

### A. Move the buttons, keep the containers

Re-home each control's markup into PhaseDisplay one at a time, and leave the modals centred. It is cheap and keeps the board's DOM. It also fails the ask: every pending choice and cost still opens in the middle of the screen with its buttons there, and PhaseDisplay grows into a pile of conditionals that four components write into through props.

### B. One dock that every component asks for space in — chosen

A new `ActionDock.svelte` owns the bottom-right corner and renders whatever the table is asking the viewer right now. The components that ask (Game.svelte's combat and mulligan state, Board.svelte's cost pickers, ChoicePromptModal, TargetingBanner, DiscardPromptModal) stop drawing buttons themselves. Each one registers a **dock request** in a small store, `client/src/lib/dock.ts`, the same way every modal registers a `ModalLayer` today. The dock picks one request by precedence and draws it. The bodies that need room (scry, a card grid, damage assignment) render as a sheet that grows up out of the dock, with their buttons in the dock's bottom row.

### C. A radial or cursor-following menu

Put the buttons wherever the pointer is. No mousing at all, but nothing stays in one place, it hides the board under the cursor, it does not work by keyboard or touch, and the owner asked for a place: bottom right.

---

## Decision

### 1. The dock

`ActionDock.svelte` is mounted by Game.svelte as a sibling of `Board`, inside `.play-area`, and absolutely positioned at its bottom-right corner (decision 1). It is not rendered for spectators, while the dev replay scrubber shows a past frame, or before the first snapshot. It is one `role="region"` named **"actions"**. That label is the tutorial anchor and an e2e contract (§10).

From top to bottom:

1. **Header.** Today's PhaseDisplay, moved whole: turn number, active player, extra-turn marks, the phase track with its pins, the step label, "no priority", the turn-rule lines and the ready-actions live region. It keeps `aria-label="turn and phase indicator"`, `.turn-no` and `.step-label`. It is the only element with those classes, so the `.first()` reads in the e2e suite read the right one.
2. **Status line.** One line, shown only when it has something to say: the running bluff ("bluffing — passes in 3s"), the CR 732 loop notice, or the hint for what `next` will do ("passing lets Lightning Bolt resolve", "passing moves to combat"). The hint comes from public state only: the top of the stack and the next step.
3. **Toggles row.** Small, quiet buttons, always in the same order: `hold`, `autopass`, `bluff` (§5), Undo with the count left, and ⋯ (decision 3). Each keeps its accessible name, `aria-pressed` and, for autopass, its `action autopass` class.
4. **Prompt area.** Present only when a request is open. It has a question line ("Lightning Bolt — choose a target, 1 of 1") and, for an inline request, the choice itself (§2). A sheet grows upward from here (§3).
5. **Action bar.** One row, always the bottom of the dock, and its slots never move: secondary buttons on the left, **one primary on the right**. The primary is the gold button, which is the convention PhaseDisplay already uses for `next` (`PhaseDisplay.svelte:535-544`). The primary sits in the very corner, so the pointer's resting place is always the button that moves the game on.

The action bar's primary is picked by precedence, strongest first:

1. A pending choice you owe, the mulligan or a discard to hand size: its confirm ("Yes", "Pay 2", "Keep hand", "Discard", "Deal damage").
2. A flow you started: targeting's Done, a cost picker's confirm, the auto-tap Cast. Its secondary is Cancel.
3. A block declaration you owe: Done blocking or No blocks.
4. The game has ended: Back to lobby.
5. Otherwise: `next`. Its secondary is Pass turn.

When a stronger request takes the bar, `next` and Pass turn give way to it, and they come back when it closes. While nothing outranks them they are always rendered and are **disabled, not hidden**, when you cannot press them (the e2e suite and the tutorial both rely on that).

**Width and height.** About 340px on a desktop: `clamp(300px, 26vw, 380px)`. PhaseDisplay is 220-300px today. Collapsed height (header, toggles and action bar) is about 150px. The dock publishes its live size as `--dock-w` and `--dock-h` on `.play-area` (a `ResizeObserver`), and the layout keeps that rectangle clear. See §4.

**Steps.** What the dock shows as the turn goes:

| Moment | Prompt area | Action bar |
|---|---|---|
| You hold priority, empty stack | — (hint: "passing moves to …") | Pass turn · **next** |
| Opponent's spell on the stack, you hold priority | — (hint: "passing lets X resolve") | Pass turn · **next** |
| Bluff running | status: "bluffing — passes in 3s" | Pass turn · **next** |
| Declare attackers, your turn | "attack": N ready · per-opponent Attack all buttons · Choose attackers… | Pass turn · **next** |
| An attacker selected | "Attacking with Grizzly Bears — click an opponent" | Cancel (no primary: the click on the board commits) |
| You owe blocks | "block": N declared | **No blocks** / **Done blocking** |
| Casting, targets | "Lightning Bolt — choose a target" | Cancel · **Done** |
| A cost picker | sheet with the picks | Cancel · **Sacrifice** (the picker's verb) |
| A "may" trigger | "Mulldrifter — draw two cards?" | No · **Yes** |
| Mulligan | sheet with your seven cards | Mulligan · **Keep hand** |
| Game over | — | **Back to lobby** |

The attack row is a secondary row, not the primary: in this engine you keep declaring while you hold priority, and `next` is what ends the declaration. That is the same rule the strip follows today.

**Keyboard focus.**

- Space stays "pass priority" and still defers to a focused button.
- **Enter activates the action bar's primary and Escape its Cancel** (with the yes/no exception below), while focus is on the board or the body (not in a text field) and no card-local menu has the keyboard. This generalises today's targeting handler (`Game.svelte:2135-2160`) from one flow to every request. Enter and Escape stay in `RESERVED_CHORDS`: the dock is the prompt on screen now, so it is the thing that owns them.
- Y / N, H / T / S and 1-9 keep working for the inline kinds that have them today (`ChoicePromptModal.svelte:883-908`). **A yes/no question does not take Enter**: a stray Enter must not accept an optional effect or pay a cost, so those answer by Y / N or a click, as they do today. Enter is for confirms that commit what the player already picked (a target, a sheet's selection, Keep hand, Done blocking). It never presses `next`; Space does.
- When a **primary** request opens (rule 1 or 3 above) and focus is on the body, the primary button takes focus, so a keyboard player can answer it at once and a screen reader lands on it. Focus never moves on an ordinary priority frame: that would steal it every pass. The ready-actions live region still speaks once per arrival, as ADR 0105 §7 says.
- Every dock button prints its key in its tooltip, as PhaseDisplay does today (`keyHint`), and also carries `aria-keyshortcuts`. The primary shows a small `⏎` or `Space` cap on the button itself.
- A new shortcut, **`toggleBluff`**, default `b` (unbound today), in the priority group, `kind: "view"`, `scope: "game"`.

### 2. Prompts: inline in the dock, or a sheet

A pending choice is **inline** when its whole answer fits the prompt area and the action bar: a question plus at most about six short buttons, or a single number. Everything else is a **sheet** (§3).

- **Inline:** `trigger_prompt`, `optional_replacement`, `confirm`, `may_cast`, `entry_pay_life`, `pay_unless` without card or tap picks, `coin_call`, `loop_shortcut`, `mana_pick`, `choose_color`, `option_pick` and `entry_controller` when every option is a short label, and an open vote's options.
- **Sheet:** the scry family, every card-grid kind, `damage_assignment`, the two order kinds, `mode_pick`, `choose_creature_type`, `choose_card_name`, `pay_unless` with `pay_cards` or a tap list, the discard to hand size, the mulligan hand (decision 2), every cost picker, the attack picker and the auto-tap preview.

An inline prompt is not a modal. It does not blur or block the board, so you can hover and read cards before you answer, which no prompt allows today. It still registers a `ModalLayer` while it owns keys (Y / N), so the global shortcuts stand down exactly as they do now.

Each request, inline or sheet, renders inside **one non-modal `role="dialog"`** (no `aria-modal`) that contains its question, its body and its action-bar buttons, named by the same text the modal is named by today. That keeps the existing dialog-name locators working ("keep or mulligan your hand", "Discard N card", a trigger's reason, "Select target for X") and keeps buttons scoped inside their dialog. The toggles row and the header are outside it.

Refusals: ChoicePromptModal's "Not accepted" alert (`:1959-1964`) moves with it into the prompt area.

### 3. Sheets

A sheet is a panel that grows upward out of the dock, right-aligned with it, as wide as it needs up to `min(720px, 100% - 24px)` and as tall as it needs up to 60% of the play area, scrolling inside. Its buttons are the dock's action bar, so a sheet's confirm is in the same corner as everything else. It has a dim edge, not a full-screen blur: the left of the board stays visible and unblurred. It has a **minimise** control on its top edge that folds it down to its question line. That control is new. Today nothing can be minimised to look at the board, and a scry over a busy table is exactly when you want to.

*Amendment (2026-10-05, #2200):* the opening hand is the one sheet that is wider than 720px. It asks for `SHEET_HAND_WIDTH` (1500px, still capped at the screen less 24px), so the seven cards sit in one row at full size (about 170px wide at 1280 and 200px at 1920) and their text reads without hovering. Everything else in decision 2 holds: it is still a sheet that grows up out of the dock, with no scrim, the roll call and "Hand size: 7. Keep or mulligan?" above the cards, a minimise control to peek at the board, and **Keep hand** and **Mulligan** in the dock bar where they were. The dialog name and button names are unchanged. At 390px it is the bottom sheet it was, three cards to a row. The hover zoom moves left of an open sheet, but never past `100vw - 360px`, so a wide sheet cannot push it off screen.

The sheet bodies are today's modal bodies, moved. The shared `.prompt-modal` shell becomes a `.dock-sheet` shell in `app.css`. Each picker keeps its own body and validation and stops rendering its own footer. Its confirm and cancel become a dock request. Each picker's Enter / Escape `$effect` is deleted, because the dock's handler does it once. The z-index ladder gets one new rung: the dock and its sheets sit at 55. That is above the log drawer (30), the strip and the command bar (40) and the vote panel (50). It is below the card-local menus (`ManaAbilityMenu` 60, `ManaSourcePicker` 70, `CardContextMenu` 3000), the full-screen modals that remain (200: settings, table settings, the bug form, deck import, spawn, the zone browser) and the hover zoom (300).

### 4. Placement, and what it must not cover

- **The self panel keeps the corner clear.** Under decision 1, the rail spans the creature and middle rows only. The bottom row is the hand, the exile strip and an empty cell `var(--dock-w)` wide under the rail. The rail's piles end above the dock. The hand keeps the rest of the row, so nothing is covered at rest. The hand's hover lift goes up, not right, so it does not touch the dock.

  *Amendment (2026-10-04, #2202):* the exile strip is now the castable-from-other-zones strip. It also holds the viewer's own commanders while they sit in the command zone, first, nearest the hand, with the same card size, hover zoom, ready ring, price tag and drag to cast. The row's layout is unchanged: still the hand, the strip and the dock's cell. The command zone panel keeps showing the commander and its own cast. Added accessible names (none renamed): the strip's region is `castable from other zones` while it holds a commander (`castable from exile` otherwise, as before); the narrow-panel chip reads `N commander(s) you may cast` or `N commander(s) and M exiled card(s) you may cast`, plus `, K ready`; a commander's tag is `costs {…} to cast from the command zone`.

- **A sheet may cover the self panel's right side** while it is open. That is a picker you opened, and it replaces a full-screen blur, so it covers less than today.
- **The hover zoom** caps at `calc(100% - 20px - var(--dock-h))` so a long card stops above the dock. While a sheet is open, the zoom moves left of the sheet (`right: calc(var(--dock-w) + 20px)`), so hovering a card in a scry shows its text beside the sheet.
- **The game log drawer** ends at `bottom: var(--dock-h)` and sits above the dock's top edge, so the log and the dock are both usable.
- **The attention strip** stays top-left and keeps only what asks nothing: the stack card (without Pass and hold), the bot feed, reveals, the roll-call, the toasts that only dismiss, and the game-over banner's text. It gets `aria-label="attention"` for tutorial step 9.
- **The dev dock** stays bottom-left. At phone width it stacks above the dock bar.
- **The tutorial's coach card** stays bottom-left (ADR 0076 §2.3). It now goes there because the action dock owns the bottom right.

### 5. Bluff in the dock

The `bluff` toggle is **always shown**, armed or not, set up or not. It is a split button: the left part toggles, the right part (`▾`) opens a small popover.

- **One click arms or disarms** (`toggleBluffArmed`), and so does `b`. It reads `bluff` off, `bluff ✓` armed, and `bluffing 3s` or `bluffing` while a bluff is running. It uses the magenta "user override" colour `hold` uses. The running-bluff line moves from PhaseDisplay's row into the dock's status line.
- **The popover (decision 4)** holds the two kinds as checkboxes ("Represent a counterspell", "Represent an instant") and the style ("Timed" / "Manual"). They write the same `gameplay.bluff*` settings the Settings panel writes, so the two never disagree. The pause range stays in Settings, with a link to it from the popover.
- **Nothing set up yet.** A click on a bluff with neither kind chosen arms it **and** turns on "Represent a counterspell", the narrower one, and the button says `bluff ✓ counter` so the player can see what they asked for. This is the only place the dock changes a stored setting without the popover, and it does so because the alternative is a button that does nothing.
- **Smart autopass off.** The button is shown disabled, titled "bluffing needs smart auto-pass: with it off you stop at every opponent spell, so a pause gives nothing away".
- **The game-mount default does not change.** `initBluffArmed` still arms the switch at mount when either kind is on.
- **Nobody else sees any of this.** The chip and its line are the viewer's own, and the "considering" chip other seats see is unchanged (`considering.ts`).

### 6. Targeting

Board targeting stays on the board: highlights, rings and the click that picks. TargetingBanner's text becomes the dock's question line ("Lightning Bolt — choose a target, 1 of 2"). Its Done and Cancel become the action bar, with Done as the primary (disabled until `canConfirm`) and Cancel as the secondary. Choice-driven targeting (`pick_target`, `retarget`, which has no Cancel today) keeps that rule. Enter and Escape behave as today, through the dock's one handler. The banner leaves the strip, so nothing duplicates.

The combat-selection hint (`Game.svelte:1751-1774`) works the same way: question line plus Cancel.

### 7. What leaves its old place

Nothing is drawn twice. Each control below is deleted where it was in the PR that adds it to the dock:

- PhaseDisplay leaves `PlayerPanel.svelte`'s bottom row. The Board → PlayerPanel → PhaseDisplay props (`onPassPriority`, `onToggleAutopass`, `autopassEnabled`, `loopNotice`) are removed. The dock reads Game.svelte's state directly.
- The stack card's `Pass` and `hold` and the floating lanes' pass (`StackOverlay.svelte:172-183`, `:263-271`; `StackLaneHost.svelte:148`) are removed. `Counter` stays on its item.
- Pass turn leaves the command bar.
- The strip's attack, block, combat-hint, mana-override and refusal clusters and the targeting banner leave the strip.
- Undo leaves the ⋯ menu and the attack cluster. There is one Undo, in the dock.
- Under decision 3, the ⋯ menu leaves the command bar. The bar keeps Lobby, the game crumb, the status, mute, log, settings and the bug button.
- Every centred `.prompt-backdrop` used during play is replaced by the dock. The full-screen modals that stay are the ones that are not about the turn: Settings, table settings, the bug report, deck import, spawn, the zone browser and the concede confirm.

The one deliberate duplicate is the keyboard. Every dock button keeps its shortcut, so a player can still keep their hands off the mouse.

### 8. Phones and narrow screens

At `max-width: 599px`, the same break #1953 uses for the command bar:

- The dock is a **full-width bar** fixed to the bottom of the play area, with the 16px side gutter. `.play-area` gets `padding-bottom: var(--dock-h)`, so the board ends above the bar instead of under it.
- The header folds to one line (`T7 · Alice · DECLARE BLOCKERS`). A `▴` opens the phase track and its pins.
- The toggles row becomes icon chips with their names as `aria-label`s (hold, autopass, bluff, undo, ⋯).
- The action bar's buttons are at least 44px tall. The primary takes the right half, and secondaries share the left half. Three or more secondaries move into a second row above.
- A sheet is a bottom sheet, full width, up to 70% of the play area tall, above the bar.
- The command bar's #1953 rules are unchanged, and it loses Pass turn and, under decision 3, the ⋯ button, so it has more room.
- Between 600px and about 1000px the desktop dock is used at its 300px floor.

### 9. Sketches

Desktop, 1440×900, three opponents (quadrant layout):

```
┌ ‹ Lobby │ CMD & CTRL / 1a2b3c4d ·························· ● connected · seq 412   🔊 📜 ⚙ 🐞 ┐
├─────────────────────────────────────────────┬─────────────────────────────────────────────┤
│ ┌ attention ──────────────┐                  │                       ┌ hover zoom ───────┐ │
│ │ stack: Lightning Bolt    │                  │                       │  card text         │ │
│ │ reveals · toasts         │   across         │       across_next     │                    │ │
│ └──────────────────────────┘                  │                       └────────────────────┘ │
├─────────────────────────────────────────────┼─────────────────────────────────┬───────────┤
│                                             │  your creatures                 │ identity  │
│                                             │                                 │ life 40   │
│   next                                      │  lands · artifacts / enchant.   │ piles     │
│                                             │                                 ├───────────┤
│                                             │  hand ▔▔▔▔▔▔▔▔▔▔▔▔  ┌──────────────────────────┐
│                                             │                     │       ACTION DOCK        │
└─────────────────────────────────────────────┴─────────────────────┴──────────────────────────┘
```

The dock, close up, while you cast a targeted spell:

```
┌──────────────────────────────────────────┐
│ T7  ● Alice                               │  header: turn, phase track (click to pin),
│ ∘∘∘ ∘ ●∘∘∘∘∘∘ ∘ ∘∘                        │  step label
│ PRECOMBAT MAIN                            │
├──────────────────────────────────────────┤
│ [hold] [autopass] [bluff ✓|▾] [↶ 1] [⋯]   │  toggles row (secondary, quiet)
├──────────────────────────────────────────┤
│ Lightning Bolt — choose a target  1 of 1  │  prompt area (only when asked)
├──────────────────────────────────────────┤
│ [ Cancel  Esc ]          [   Done  ⏎   ]  │  action bar: secondary · PRIMARY
└──────────────────────────────────────────┘
```

The same dock with a sheet open (scry 2), growing up out of it:

```
                 ┌───────────────────────────────────────────┐
                 │ Scry 2                                  ▾  │  minimise
                 │ ┌──────┐ ┌──────┐                          │
                 │ │ card │ │ card │   ↑  To bottom  Keep on top│
                 │ └──────┘ └──────┘                          │
                 ├───────────────────────────────────────────┤
                 │ T7 ● Alice · PRECOMBAT MAIN               │
                 │ [hold] [autopass] [bluff|▾] [↶ 1] [⋯]     │
                 │                           [    Done  ⏎   ] │
                 └───────────────────────────────────────────┘
```

Phone, 390px wide:

```
┌ ‹ Lobby / 1a2b3c4d   ●    🔊 📜 ⚙ 🐞 ┐
│                                      │
│   board (as today; ends above the    │
│   dock bar, nothing underneath it)   │
│                                      │
├──────────────────────────────────────┤
│ T7 Alice · PRECOMBAT MAIN        ▴   │
│ (hold)(auto)(bluff✓)(↶1)(⋯)          │
│ ┌───────────────┐┌─────────────────┐ │
│ │   Pass turn   ││      next  ⎵    │ │  44px tall
│ └───────────────┘└─────────────────┘ │
└──────────────────────────────────────┘
```

### 10. Accessibility, the e2e suite and the tutorial

- **Labels are contracts.** The dock is `region "actions"`. Its header keeps `turn and phase indicator`. The toggles row is `group "priority controls"`, PhaseDisplay's current label. The attack row is `group "declare attackers"` and the block row `group "declare blockers"`, the clusters' current labels. The strip becomes `region "attention"`. `board-layout.spec.ts` asserts each of them, so ADR 0076 §2.4's contract covers them. AGENTS.md gets the line ADR 0076 asked for: these `aria-label`s are a user-facing contract (tutorial anchors and e2e selectors), and renaming one is a breaking change.
- **Names stay.** `next` (exact), "Pass turn", "Keep hand", "Mulligan", "Attack <name> with all N", Yes, No, "Pay N", "Don't pay", "Discard", "Take" and every dialog name keep their text. A dock request renders only for its chooser, so the "absent on the opponent's page" checks hold.
- **Each PR fixes the selectors its move breaks, in the same PR,** and runs the nightly Playwright job on its branch before merge (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`), because PR CI does not run Playwright.
- **Tutorial.** Step 8 anchors to `region "actions"` (it teaches `next`, `hold` and `autopass`, which all live there). Step 9 anchors to `region "attention"`. Step 10's attack stays the creature row and the opponent medallion, and the attack-all button is now in the dock's attack row if a step wants it. #1079 and #1081 are not built, so they target these labels from the start. ADR 0076 §1's "Counter and Pass sit on its top item" becomes "Counter sits on its top item; Pass is the dock's `next`".

---

## Delivery

Each PR ships on its own, leaves no control drawn twice, works at 390px for what it moves, updates the e2e selectors it breaks, and says in its body what it changes for the tutorial anchors. The bluff control and the priority buttons land first.

| PR | What lands | e2e and vitest | Tutorial anchors |
|---|---|---|---|
| 1 | **Bluff, always visible.** The split `bluff` button is always shown in PhaseDisplay's row, with the `▾` popover (kinds and style), the "nothing set up" rule and the disabled smart-autopass-off state (§5). The `toggleBluff` shortcut on `b`. Settings' Bluff help text names the button. ADR 0009's pointer line. | New `bluffChip.render.test.ts`. No e2e selector changes. `board-layout.spec.ts` asserts the bluff button. | None |
| 2 | **The dock, with priority.** `ActionDock.svelte` at the bottom-right (decision 1). PhaseDisplay moves in as the header. `next` and Pass turn (out of the command bar) in the action bar; hold, autopass and bluff in the toggles row. The status line and the `next` hint. Stack `Pass` and `hold` mirrors removed. `--dock-w` / `--dock-h`, the rail change, the hover-zoom and log-drawer clearances. The phone bar (§8). Enter / Escape / focus rules for `next`. | `.step-label` and `.turn-no` stay unique. `button.action.autopass` keeps its class and `aria-pressed`. "pass turn" moves but keeps its name and its disabled state. `gameToolbarBug.render.test.ts` drops Pass turn and keeps its order checks. `phaseDisplay.extraTurn.render.test.ts` mounts the header. `board-layout.spec.ts` asserts `region "actions"`. | Step 8 → `region "actions"`. ADR 0076's pointer line (§1, §2.3) |
| 3 | **Combat.** `dock.ts`, the request store. The attack row (attack all, per-opponent, Choose attackers…), Done blocking / No blocks as primary, the combat-selection Cancel, the attack-tax and attack-limit refusal actions. **Undo** joins the toggles row with its count, and leaves both the ⋯ menu and the attack cluster in the same PR, so there is one Undo. The strip clusters go. | `attack-all-318.spec.ts:139,148,165` moves from `.att.attack-all` to `group "declare attackers"` inside `region "actions"`, and its Undo to the dock's one Undo. `gameToolbarBug.render.test.ts`'s menu-item list drops Undo. A new e2e for No blocks. | Step 10's attack-all lives in the dock |
| 4 | **Targeting and payment.** TargetingBanner → dock question line plus Done / Cancel. The insufficient-mana toast's Auto-tap & cast and Cast anyway become a dock request. One Enter / Escape handler for every request; the per-modal key effects these flows used are deleted. The strip gets `region "attention"`. | `targetingBanner.render.test.ts` rewritten against the dock. `s19-triggers.spec.ts:68`'s "Select target for X" dialog name kept. | Step 9 → `region "attention"` |
| 5 | **Inline choices.** The yes/no family, `pay_unless` without picks, `coin_call`, `loop_shortcut`, `mana_pick`, `choose_color`, short `option_pick` / `entry_controller`, vote options and game-over Back to lobby (§2). ChoicePromptModal keeps only the sheet kinds. | `s19-triggers.spec.ts` dialog names and Yes / No / Pay / Don't pay unchanged. | None |
| 6 | **Sheets.** The `.dock-sheet` shell and the minimise control. Every remaining ChoicePromptModal kind, DiscardPromptModal, every Board.svelte cost picker, AttackDeclarationModal, AutoTapPreviewModal and the mulligan hand (decision 2, with Keep hand / Mulligan in the action bar) render as sheets with their buttons in the action bar. Their own footers and key effects are deleted. Hover zoom moves left of an open sheet. | `state-freeze-266.spec.ts:95-104` (`.prompt-count`, `button.card-pick`, Discard) and `s19-triggers.spec.ts:510-511` ("Take") keep their dialog scope, and so does `mulligan.spec.ts:63,71,80` (the dock request is the dialog). The classes they read are kept or the selectors move to roles. A new e2e answers a scry from the dock. | None |
| 7 | **Overflow and close-out.** The ⋯ menu moves into the dock (decision 3): sandbox (draw, untap all, shuffle, mulligan to N, life history), table, vote launcher and Concede with its confirm. The AGENTS.md labels-are-a-contract line. The S56 section of `docs/sprints.md`. This ADR's delivery note. | `gameToolbarBug.render.test.ts`'s ⋯ checks move to the dock. A new e2e opens the dock's ⋯ and cancels a concede. | None |

PRs 1 and 2 are the owner's two sentences: bluff visible bottom right, and the buttons pressed every turn in the same corner. PRs 3-6 bring each remaining kind of prompt into it. PR 7 is cleanup.

### Delivery note (2026-10-02)

All seven PRs shipped on 2026-10-02, into `develop`: PR 1 [#1969](https://github.com/krakenhavoc/cmd_and_ctrl/pull/1969), PR 2 [#1976](https://github.com/krakenhavoc/cmd_and_ctrl/pull/1976), PR 3 [#1978](https://github.com/krakenhavoc/cmd_and_ctrl/pull/1978), PR 4 [#1980](https://github.com/krakenhavoc/cmd_and_ctrl/pull/1980), PR 5 [#1981](https://github.com/krakenhavoc/cmd_and_ctrl/pull/1981), PR 6 [#1982](https://github.com/krakenhavoc/cmd_and_ctrl/pull/1982) and PR 7 [#1983](https://github.com/krakenhavoc/cmd_and_ctrl/pull/1983). The sprint is S56 in [docs/sprints.md](../sprints.md#s56--table-ux-one-action-dock-bottom-right).

Where PR 7 went past what this ADR wrote down:

- **Spectators keep a ⋯ menu on the command bar.** §1 gives a spectator no dock, and decision 3 moved the menu into it, so the ADR left a spectator without one. Before PR 7 a spectator had no ⋯ menu at all: the command-bar menu needed a seat. Now a viewer with no dock (a spectator, or an admin with no seat at this table) gets a smaller one on the command bar, opening downward: life history, table settings (read-only unless they may manage the table), spawn where both gates allow it, My games and Back to lobby. Nothing in it acts on a seat. At 390px that bar also hides the game id and the status word (the dot stays, and the word stays for a screen reader), so it fits beside the "spectating" tag.
- **The vote launcher is a form inside the menu.** "Call a vote…" swaps the menu for the launcher's two fields (`vote topic`, `vote options`) and its start / cancel, where the menu was. The launcher left the board's top-left corner, and VotingPanel now draws only an open vote for a viewer with no dock. A spectator never could call a vote: the launcher needed a seat, so it was a dead button for them.
- **Concede's confirm opens where the menu was**, inside the dock, rather than as a popover under the command bar. It keeps its name (`concede the game?`), stays `aria-modal`, and puts focus on Keep playing.
- **Life history** still opens under the command bar, where it always was. It is a read-out, not a prompt.
- **Keys:** the menu closes on Escape (focus returns to ⋯), on a press outside it and when focus tabs out. The arrow keys, Home and End move between its entries. While it is open it holds a ModalLayer, as the command-bar menu did, so the global shortcuts and the dock's Enter / Escape stand down.

Known follow-ups, also listed in the S56 section:

- [#1977](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1977), the phone board layout, which predates the dock (Out of scope, above).
- PR 4's nuance: with keyboard focus on a board card during targeting, Enter picks the card rather than pressing Done (§1, "Enter … while focus is on the board or the body").
- The `zone-browser.spec.ts` cross-socket flake ([#1468](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1468)), which PR 6's E2E run hit once and passed on retry.

---

## Consequences

- Every button the game waits on is in one corner, the primary is always in the same slot, and Enter / Escape do the same thing for every prompt.
- Prompts stop blurring the board. A player can read the table while answering, and can minimise a sheet to look at it.
- About twenty modal footers and their key handlers go. The dock's request store is a new shared seam: a new prompt registers a request instead of drawing buttons. A prompt that forgets is visible (its confirm is missing), not silent.
- The self panel's rail loses its bottom cell under decision 1. On short panels the piles scroll sooner.
- The dock covers part of the self panel while a sheet is open.
- The e2e suite's class and dialog selectors move to the labels this ADR names, which is what `docs/e2e-review.md` asked for anyway.

## Out of scope

- A phone layout for the board itself.
- London bottom-N (not implemented; when it lands it is a dock sheet).
- Moving card-local menus off the card (decision 5) or seat-local controls off the seat.
- A read-only dock for spectators.
- Any server change. Nothing here touches the protocol or `legal_moves`.

## Questions for the owner (answered)

All five were answered on 2026-10-02 with option (a). The (b) options are kept as considered options. See [Owner decisions](#owner-decisions-2026-10-02).

1. **Which corner, exactly?**
   - (a) **Recommended. Chosen.** The screen's bottom-right corner. The self panel's rail (your identity and piles) stops above the dock, so the piles sit higher and the rail scrolls sooner on short screens.
   - (b) Considered, not chosen. Where the phase widget is today, in the self panel's bottom row to the left of the rail. The rail keeps its full height, and the dock sits about 112px in from the right edge.
2. **Big pickers (scry, search, card grids, damage assignment, cost pickers, the mulligan hand).**
   - (a) **Recommended. Chosen.** They become sheets that grow up out of the dock, with their buttons in the dock's bottom row. The rest of the board stays visible, and a sheet can be minimised.
   - (b) Considered, not chosen. They stay centred modals that blur the board. Only the small prompts (yes/no, pay, coin, simple confirms) move into the dock.
3. **The ⋯ menu (undo, sandbox draw/untap/shuffle, life history, table settings, concede).**
   - (a) **Recommended. Chosen.** It moves into the dock as a ⋯ button on the toggles row, opening upward. The command bar keeps lobby, status, mute, log, settings and the bug button. Undo is in the dock either way.
   - (b) Considered, not chosen. It stays in the command bar. Only Undo comes out into the dock.
4. **Bluff options.**
   - (a) **Recommended. Chosen.** The dock's bluff button arms and disarms in one click. Its `▾` sets which bluff (counterspell, instant) and how (timed, manual), writing the same settings the Settings panel uses. A click with nothing set up arms "represent a counterspell". The pause range stays in Settings.
   - (b) Considered, not chosen. The dock's button only shows and toggles the armed state. Which bluff and how stay in Settings. With nothing set up, a click opens Settings at the Bluff section.
5. **Card menus (right-click, a land's mana choices, an ability popover).**
   - (a) **Recommended. Chosen.** They stay at the card. They open under the pointer, so there is no travel, and everything after them (targets, costs, payment, confirm) is in the dock.
   - (b) Considered, not chosen. Clicking a card also lists its actions in the dock, so every button, card actions included, can be pressed in the corner.

---

## Owner decisions, 2026-10-02

The owner answered the five questions on 2026-10-02. Every answer was the recommended option (a).

1. **The corner.** The dock sits in the screen's true bottom-right corner. The self panel's rail (identity and piles) spans the creature and middle rows only and stops above the dock, so the piles sit higher and the rail scrolls sooner on short panels (§1, §4). It lands in Delivery PR 2, with the dock itself.
2. **Big pickers.** Scry and its family, search, every card grid, damage assignment, the order kinds, every cost picker, the attack picker, the auto-tap preview and the mulligan hand become sheets that rise out of the dock. Their buttons are in the dock's action bar, the rest of the board stays visible and unblurred, and a sheet can be minimised (§2, §3). It lands in Delivery PR 6. The small prompts are inline in Delivery PR 5.
3. **The ⋯ menu.** It moves into the dock as a ⋯ button on the toggles row, opening upward, with the sandbox entries, life history, table settings, spawn, the vote launcher and Concede with its confirm. The command bar keeps Lobby, the status, mute, log, settings and the bug button (§1, §7). It lands in Delivery PR 7. Undo comes out of the menu earlier, in Delivery PR 3.
4. **Bluff options.** The dock's bluff button arms and disarms in one click, and so does `b`. Its `▾` sets which bluff (counterspell, instant) and how (timed, manual), writing the same `gameplay.bluff*` settings the Settings panel writes. A click with nothing set up arms "represent a counterspell". The pause range stays in Settings (§5). It lands in Delivery PR 1.
5. **Card menus.** The right-click menu, a land's mana choices and the ability popover stay at the card, because they open under the pointer. Everything after them (targets, costs, payment, confirm) is in the dock (§6, Out of scope). No PR moves them. Delivery PRs 4 and 6 move what follows them.
