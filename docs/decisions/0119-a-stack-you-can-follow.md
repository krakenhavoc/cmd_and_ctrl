# ADR 0119 — A stack you can follow

**Status:** Accepted · 2026-10-04 · S60 — Table clarity: a stack you can follow (tracker [#2204](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2204))
**Issues:** [#2204](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2204), which is both this change and the sprint's tracker. No separate S60 tracker exists.
**Owner decisions:** the request of 2026-10-04 and the three answers given the same day, quoted under [Owner request and answers](#owner-request-and-answers-2026-10-04). They are binding. This ADR also makes sixteen smaller calls the answers did not cover. They are listed under [Calls made here](#calls-made-here) so the owner can overturn any of them in review, before the PR that implements each one lands.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-04. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 40 remote heads, including `origin/develop`, `origin/main`, the S59 branches (`feat/s59-2201-lone-ability-click`, `feat/s59-2202-commander-strip`, `fix/s59-2203-land-drop-highlight`) and the card, chore, docs, feat, fix, repro and wip branches. The highest number on any of them is 0117. **0118 is reserved** for strict mana payment by default ([#2188](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2188), S59 tracker [#2189](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2189)) and is not written yet. No branch has 0119, so this ADR takes **0119**.
**Amends:** [ADR 0009](0009-smart-priority-autopass.md), the #1307 amendment's "Considering a response…" section (the 800 ms delay assumed an automatic pass lands in one round trip; §2 adds a hold before it). [ADR 0111](0111-action-dock.md) §4 (the stack card leaves the attention strip when the pile is the style) and §1's status line (it also shows the hold's countdown). [ADR 0075](0075-table-settings-and-host-controls.md) §2.2 (a bot pace preset also sets a stack hold).
**Builds on:** [ADR 0007](0007-stack-foundation.md) (the stack, and the caster keeping priority), [ADR 0018](0018-triggers-on-the-stack.md) §4 and §5 (wire order by `Seq`; autopass through triggers), [ADR 0037](0037-unimplemented-card-signal.md) (the `manual` chip), [ADR 0053](0053-combat-damage-beats.md) (display-only sequencing from cached geometry), [ADR 0104](0104-gaining-control-of-a-spell.md) §6 (the taken-from chip), [ADR 0110](0110-remember-me.md) §4 (synced settings).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

The stack is a small card in the top-left corner. An opponent's spell that the viewer cannot answer is passed by the viewer's own client the moment it arrives, so it is often on screen for less than a second. When it resolves, it vanishes on the next frame with nothing to say what happened. The owner asked for MTG Arena's presentation: a pile of large, readable cards, arrows to targets, and time to read them.

Every claim below was checked on `origin/develop` at `92a7e019`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`).

### Owner request and answers (2026-10-04)

The request, verbatim:

> we need to make the stack and resolution much more clear. See how arena does it? … things are passing by too quickly that it is not obvious what is happening some times

The owner attached an Arena screenshot. It shows a fanned pile of full-size cards at the battlefield's right edge, with the top item in front and its whole text readable. While a target is chosen, the source permanent glows and a curved arrow runs from it to the target, under the prompt "Choose any target." Undo, Cancel and Submit sit at the bottom right.

The answers:

1. **Look.** An Arena-style pile becomes the default. It shows full-size card images, not art crops, fanned, with the top card in front and readable without hovering (its full oracle text). It shows whose item it is, an arrow to each target, and the lower cards peeking out behind. It goes **on the left side of the battlefield, not the right, "so hover on cards still can show on the right side"** (the hover zoom is pinned top-right). The old styles (compact, fan, spotlight, ribbon) stay selectable in Settings.
2. **Pacing: hold, and a linger on resolution.**
   - (a) An opponent's or bot's stack item stays on the stack for at least about 2 s before the viewer's auto-pass lets it resolve. Bots likewise wait before passing on other players' items. This is a per-player setting, and 0 turns it off.
   - (b) When anything resolves, its card lingers for about 1.5 s, marked resolved, countered or fizzled, and moves toward where it went. This does not slow the game further. It is display only, like ADR 0053's combat beats.
3. **Also.** Arrows while targeting: the spell or source glows and a curved arrow follows to each picked target (Arena's "Choose any target" view). Targeted permanents get a ring while the item is on the stack. A **fuller game log** records triggers going on the stack and every activated ability, not only resolutions. **Not chosen:** opening the log by default, and a phone layout. Phones keep today's behaviour (compact); a phone layout is out of scope.

### The rules

- **CR 117.3c:** "If a player has priority when they cast a spell, activate an ability, or take a special action, that player receives priority afterward." A bot's own spell is passed by the bot first, so the hold that matters is on the other seats.
- **CR 117.3d:** a player with priority who "chooses not to take any actions" passes. The rules say nothing about how long a player may take before passing. The hold is the player taking time, and it changes nothing in the rules.
- **CR 117.4 and 405.5:** "If all players pass in succession …, the spell or ability on top of the stack resolves." Every seat's pass is needed, so one seat's hold delays the resolution for the whole table. That is why the hold does not need to be added to every seat in turn (§2).
- **CR 405.1 and 405.2:** an activated or triggered ability goes on the stack "without any card associated with it", and each new object is put "on top of all objects already there." The pile shows an ability with its source's card, captioned with the ability.
- **CR 113.7:** the source of an activated ability on the stack "is the object whose ability was activated"; for a triggered ability, "the object whose ability triggered." That is the card that glows while targeting (§4).
- **CR 115.1 and 601.2c:** targets "are declared as part of the process of putting the spell or ability on the stack", and "can't be changed except by another spell or ability that explicitly says it can do so." The arrows drawn while choosing are the ones the item will carry.
- **CR 603.3 and 603.3b:** a triggered ability goes on the stack "the next time a player would receive priority"; several at once are put there in APNAP order (CR 101.4), each player choosing the order of their own. The trigger-order sheet and the pending-trigger rows stay as they are.
- **CR 605.3b:** "An activated mana ability doesn't go on the stack." The fuller log leaves mana abilities out (§5).
- **CR 608.2b:** if all of an item's targets are illegal, "the spell or ability doesn't resolve. It's removed from the stack and, if it's a spell, put into its owner's graveyard." The log already calls this "countered by game rules (no legal targets)". The linger calls it *fizzled*, the word players use.
- **CR 608.2n:** an instant or sorcery is put into its owner's graveyard "as the final part" of its resolution, and an ability "is removed from the stack and ceases to exist." **CR 608.3a:** a permanent spell "becomes a permanent and enters the battlefield." These are the linger's destinations.
- **CR 701.6a:** "A countered spell is put into its owner's graveyard"; a countered ability is removed from the stack.

### What exists

**Four stack styles, one model.** `settings.display.stackStyle` is `compact | fan | spotlight | ribbon` (`client/src/lib/stackLane.ts:59-62`), default `compact` (`settings.ts:353`), chosen in Settings under "Stack *experimental*" (`Settings.svelte:602-622`). They came from #1467 (PRs #1515, #1521, #1523) without an ADR. One model, `buildStackLane` (`stackLane.ts:324`), gives every style the verbatim title, the art card and preview card, the caster and their seat colour, the targets and `targetText`, `isTop` and `position` (the wire is bottom to top, so it is reversed), the chips (`chipsFor`, `:484`, including `manual` at `:509-518`), and an effect summary that never invents rules text (`stackLane.ts:33-50`, `summarize` at `:639`).

**Compact (the default)** is `StackOverlay.svelte`, mounted first in the attention strip (`Board.svelte:2615-2633`). The strip is `region "attention"`, top-left, `width: min(512px, calc(50% - 132px))`, z 40 (`Board.svelte:3167-3178`), and moves to the middle column in the row layout with three opponents (`:3184-3187`). Each row is a 40×56 thumbnail (`StackOverlay.svelte:387-388`), the title, the caster, `targetText`, the chips and a Counter button that is disabled without priority (`:233-240`). It shows no oracle text and no summary. Hovering a row opens the hover zoom, which is pinned top-right: `right: var(--zoom-right, 10px); top: 10px; width: min(332px, calc(50% - 132px)); z-index: 300` (`HoverZoomOverlay.svelte:315-334`).

**The floating lanes.** `StackLaneHost.svelte` mounts when the style is not compact (`Board.svelte:2600-2610`). It owns the header (count and who holds priority), the summary line, the controls, an aria-live announcer (`:254`) and the label `stack: N on the stack` (`:222`). It picks the design from `STYLE_BODIES` (`:71-75`), centres on the seam between the opponents' row and the viewer's row (`measure`, `:155-205`), and sits at z 38 (`:262-268`). **Fan** draws `art_crop` tiles (`StackLaneFan.svelte:217`) up to 252px tall (`:404`) and its own arrows (`stackArrows.ts` `planArrows` `:96`, `measureStackArrows` `:356`), and outlines what it points at with `data-stack-lane-target` (`syncTargetMarks`, `:419`). **Spotlight** draws the top item as a `normal` image at 150×210 (`StackLaneSpotlight.svelte:55`, `:223-224`). **Ribbon** is a numbered row.

**Arrows.** While a target is being chosen there are none. Legal targets get a green ring and picked ones a gold ring (`Card.svelte:1770-1789`, `.card.targetable` and `.card.picked`), and the question is in the dock ("Select target for X", `targetingDock.ts:181`). The targeting store knows the source: `TargetingState.card` is the hand card being cast or the permanent whose ability it is (`targeting.ts:242-243`). Once the item is on the stack, `CombatArrows.svelte` draws gold dashed arrows from `[data-stack-item-id]` rows to permanents and players (`:111`, `:178-196`) for compact, spotlight and ribbon. Fan draws its own, and CombatArrows stands down for it (`Board.svelte:2583-2588`). A graveyard or exile target gets no arrow (`stackArrows.ts:12-13`). Nothing glows on the source.

**Resolution.** There is no animation, toast or marker. The item disappears on the next frame. A permanent gets `etbPulse` on the battlefield (`animations.ts:157`). An instant or sorcery goes to the graveyard silently. Countering and fizzling are only in the log: "X countered Y" and "… was countered by game rules (no legal targets)" (`server/internal/protocol/log.go:1823-1835`). `animations.ts` has no cross-zone flight; its comment says that would need a portal or FLIP step (`:119-125`).

**What the log carries about resolution.** A spell's `EventResolve` and `EventFizzle` carry the card's instance ID (`game/mutations.go:3114-3142`), which is also the stack item's ID for a spell (`CombatArrows.svelte:187-190`), so the log entry names the item. An ability's carry only its source and label (`mutations.go:3470-3483`), so two items with the same label from one source cannot be told apart. A countered ability's `EventCounterSpell` does carry the item's ID in `Target` (`game/effect_api.go:1931-1937`), but the projection clears it in favour of the label (`log.go:939-955`). Instance IDs survive zone moves: the engine mints a new one only for new objects (tokens, copies, emblems, spawns), never on a move.

**Pacing.**
- **The server** resolves when every seat has passed in succession. Nothing delays a broadcast.
- **The client's auto-pass** is a pure gate chain, `autopassDecision.ts` (precedence in its header comment, `:18-50`). Its defaults are `autoPassPriority` true, `smartAutoPass` true, `alwaysStopOpponentStack` false and `autoPassOwnStack` true (`settings.ts:377-405`). With those, an opponent's item the viewer cannot answer passes at once (ADR 0009, the #1307 amendment). The pass is sent from a `$effect` in `Game.svelte` (`:445-485`), de-duplicated by snapshot seq.
- **The timed bluff** is the precedent for a delayed pass. It holds for a random 1.5–4 s (`settings.ts:396-400`) on a timer that is called off by a new frame or by any action the viewer sends (`Game.svelte:327-367`), and the dock's status line counts it down (`ActionDock.svelte:357-367`, `:429-431`).
- **The "considering a response…" chip** shows on another human's seat once they have held priority for 800 ms in a response window, on the assumption that "automatic passes land in about one round trip" (`Board.svelte:336-379`; ADR 0009, "Considering a response…").
- **Bots** pace every move, passes included, by the table's bot speed: `botPacePresets` gives fast 0 / 2 s, normal 700 ms / 2 s and slow 2 s / 8 s (`server/internal/aiseat/runner.go:129-133`), re-read before every decision (`pacingNow`, `:866`). A bot's spell that the viewer cannot answer is on screen for about MinThink plus a round trip, roughly 0.7 s at normal. The runner already holds a pass for a reason other than thinking: `BlockGrace` and `holdForBlockers` (`runner.go:51-62`, `:908`). A stepped (lockstep) runner switches off every such hold (`aiseat/stepped.go:38-60`). The table settings panel calls the setting "Bot speed" (`TableSettingsPanel.svelte:235-251`, choices in `tableSettings.ts:146`).

**Triggers** wait in `pending_triggers` and are drained in APNAP order (ADR 0018 §3). Compact lists their labels, and the dock's trigger-order sheet asks "Order your triggers".

**The log.** `GameLogPanel.svelte` is a right-edge drawer, mounted only while open (`Game.svelte:1662-1664`), and closed by default. It is rendered by the server (`server/internal/protocol/log.go`), in a ring of `PublicLogMax = 200` entries (`log.go:83`). It records casts, resolutions, counters and fizzles. It does not record an activation, except one made on another player's permanent (`LogActivateAcross`, `log.go:1240-1253`), and it does not record a trigger (`EventTrigger` is a listed silence, `log_event_kind_gate_test.go:81`). `EventTrigger` is emitted in three places: when a harvested trigger is queued (`game/triggers.go:953-958`) and when one is announced by hand (`game/mutations.go:3900-3905`), both with the label, and as a label-less breadcrumb on every activation (`game/activated.go:1824-1829`), which is followed by `EventActivateAbility` with the label and the item's ID (`:1854`). Cycling emits `EventCycle` from its cost (`game/discard_cost.go:306-311`) and then the same activation events.

**Contracts.**
- `board-layout.spec.ts:108-112` needs exactly one `region "attention"`, and no `stack: N on the stack` label while the stack is empty.
- `s19-triggers.spec.ts:140-145` needs the verbatim label "Mulldrifter — draw two cards" visible.
- The s19 helper seeds `alwaysStopOpponentStack: true` (`s19-helpers.ts:634-644`), and `resolveStack` clicks `next`.
- The tutorial's step 6 detour says "Your creature is on the stack, at the top of the screen." (`tutorialSteps.ts:250-254`), and `region "attention"` is step 9's anchor (ADR 0076 §2.4).
- The vitest files `stackLane*.test.ts`, `stackArrows.test.ts`, `stackHover.test.ts` and `stackPreview.test.ts` cover these styles.
- ADR 0111 §8 sets the phone break at `max-width: 599px`.

---

## Decision

### 1. The pile

**A fifth style, `pile`, and the new default.** It is one more entry in `STACK_STYLES` and in `StackLaneHost`'s `STYLE_BODIES`, so it reads the same model, header, controls, summary and announcer as the other lanes. Nothing about the stack is derived twice. The host gains a per-style placement. Fan, spotlight and ribbon keep the centred lane; the pile anchors left.

**Placement.** The pile's left edge is 12px from the board's left edge, the same inset as the attention strip. It is vertically centred on the seam between the opponents' row and the viewer's row, which the host already measures. It is then clamped:

- below the attention strip's measured bottom plus 8px, whenever the strip has content (the bot feed, a reveal, a toast);
- above any bottom-left fixture that is showing (the tutorial's coach card, the dev dock), and inside the board.

It floats over the left part of both rows and never reflows the grid, like the other lanes. It takes pointer events only on its own cards. The hover zoom stays top-right, uncovered, which is the reason for the left side (owner answer 1). The dock stays bottom-right.

**Size.**
- **The top card** is the Scryfall `normal` image. Its width is `clamp(200px, 30% of the board's height, 280px)`, and never more than 24% of the board's width. At 270px wide on a 900px-tall board, a card's rules text is readable without hovering. The image's `alt` is the card's name and its oracle text from the `/cards` cache, so a screen reader reads what the eye reads.
- **An ability** has no card of its own (CR 405.1). It is drawn as its source's card image with a caption band across the text box. The band shows the ability's verbatim label and a kind tag ("triggered ability" or "activated ability").
- **A face-down spell, or a source the viewer may not see** (`previewCard` null, #697), is drawn as a card back with the name the model gives.

**Depth.** The top card is in front. Up to four lower items peek out above it, each offset 34px up so its name line shows. 34px is the name bar of a `normal` image at the widest size. Lower items show the caster's colour and the title, and do not zoom on their own; hovering one opens the hover zoom, as on compact (#322). A fifth item or deeper shows a **"+N more"** chip on the top edge. The chip opens the whole stack as a scrolling column of compact rows over the pile, top first, and closes on Escape or a second click. Pending triggers are listed under the pile as compact rows headed "waiting to go on the stack", as compact lists them today.

**What each item carries.** Every chip the model gives, `manual` included (ADR 0037), and the taken-from chip (ADR 0104 §6). The caster's name in their seat colour above the title, and the seat colour as a band on the card's left edge. The verbatim title as real text, never only inside the image, because e2e specs read it. The target line ("→ your Birds of Paradise"). **Counter** on every item, on the top card's caption and on each peeking name line, disabled without priority (ADR 0111 §7). The header (count, priority holder, split second) and the summary line come from the host, unchanged.

**Arrows and rings.** The pile uses the fan's machinery (`planArrows`, `measureStackArrows`, `syncTargetMarks`). Gold arrows run from the top card and seat-colour arrows from each peeking item, and every board target gets the `data-stack-lane-target` ring. CombatArrows stands down for the pile as it does for the fan.

**While the viewer is choosing a target,** the pile shrinks to its top card at 160px wide and drops its peeking items, so the board under it can be reached. It grows back when the choice ends. A **collapse** button on the header folds the pile to a slim tab ("Stack · N") at the same place, for the rest of the stack's life. The tab expands it again, and a new item on an empty stack starts expanded.

**The attention strip** keeps everything else: the bot feed, reveals, the roll-call, toasts and the game-over text. It stays `region "attention"`, so exactly one exists and tutorial step 9's anchor holds. It draws no stack card while the pile is the style, the same rule as today's lanes (`Board.svelte:2616`).

**Phones and small boards.** Phones keep today's behaviour (owner answer 3). At `max-width: 599px`, ADR 0111 §8's break, the board renders the compact card whatever the setting says, and the stored setting is untouched. It does the same on a board too short for a 200px top card between the strip and the bottom.

**Accessible names.** Nothing is renamed (AGENTS.md §5).
- The pile's section keeps `stack: N on the stack`, so `board-layout.spec.ts` holds.
- Counter keeps its name, `Counter`.
- New names: the collapse button `hide the stack` / `show the stack`; the depth chip `show N more on the stack`; the pending list `group "waiting to go on the stack"`.

**Settings.** The select loses "experimental" for the pile and lists it first: "Pile (default): large cards on the left". The other four keep their labels, marked experimental as today.

**Migration.** `SETTINGS_VERSION` goes up. A stored blob below that version whose `stackStyle` is `compact` becomes `pile`. A stored `fan`, `spotlight` or `ribbon` is kept, because nobody reaches those without choosing them. `compact` was the default, and the shallow merge has always written defaults into the stored blob (`settings.ts:615-630`), so an untouched `compact` and a chosen `compact` look the same. The v2 → v3 migration met the same problem and moved the untouched case (`settings.ts:657-668`). From the new version on, a stored `compact` is honoured. The account copy (ADR 0110 §4) goes through the same `migrate` with its own version (`applySyncedCopy`, `settings.ts:589`), so a synced copy moves the same way. `isStackStyle` accepts `pile`, and an unknown value falls back to `pile` instead of `compact`.

### 2. The hold

**The setting.** `gameplay.stackHoldMs`, synced (ADR 0110 §4). The choices are 0 (off), 1, 2 and 3 seconds, and the default is 2000. The control is a select in Settings → Gameplay, under the auto-pass options: "Let other players' spells sit on the stack for at least … before auto-pass lets them resolve". A stored value outside 0–3000 is clamped where it is read, as the bluff bounds are (`bluff.ts`).

**What it holds.** An automatic pass, only, when the **top** stack item is controlled by someone other than the viewer. The pass waits until that item has been on the viewer's screen for `stackHoldMs`, measured from the first frame in which this client saw its ID. It covers every automatic pass: smart auto-pass (rule 6 of `autopassDecision.ts`) and the session autopass toggle (rule 4) alike, because the owner asked that the item stay visible "before the viewer's auto-pass lets it resolve". It never holds:
- the viewer's explicit `next` (or its keyboard shortcut), which passes at once as today;
- the viewer's own top item (`autoPassOwnStack` passes it at once, as today);
- an empty stack.

**Where it lives.** `autopassDecision` stays a pure precedence and is not changed. A new pure module, `stackHold.ts`, answers how much longer to wait, from the stack, the viewer, the first-seen times, `stackHoldMs` and now. The `$effect` in `Game.svelte` turns a `pass` verdict into a timer for that remainder, through the same machinery as the timed bluff. The hold is called off by a new frame, by any action the viewer sends, and by the verdict changing. A timed bluff and the hold combine as the longer of the two delays, not their sum. While the hold runs, the dock's status line counts it down ("auto-pass in 1.4 s"), so the table does not look stalled and the player knows `next` still works. That status line is the one ADR 0111 §5 gave the bluff.

**The bots.** A bot cannot read a person's settings, so the bot side is part of the table's **bot speed** (ADR 0075 §2.2), which the host already sets and the runner already re-reads before every decision. Each `botPacePresets` entry gains a `StackHold`: fast 0, normal 2 s, slow 3 s. When the runner's chosen move is a pass and the top stack item is controlled by another seat, it holds until that item has been on the stack for `StackHold`, measured from the commit at which the runner first saw the item's ID. It keeps a first-seen map, pruned as items leave. The hold re-checks every 100 ms, as `holdForBlockers` does. If the top item changes while it waits, the runner re-enumerates instead of sending a pass decided on the old frame. The hold overlaps MinThink rather than adding to it: the pass goes at whichever is later. `FollowTablePace` off (every hand-built test `Config`) or a stepped runner means no stack hold, so tests, arenas and the soak keep their speed. The "Bot speed" hints say so ("normal: bots think for about a second and leave others' spells on the stack for 2 s").

**Why it is not added up seat by seat.** Every seat measures from the moment the item appeared, and every seat sees it at about the same moment. So at a table of three bots and one person, a bot's spell sits for about the hold once, plus the later seats' MinThink, not four holds (CR 117.4). A seat with its hold at 0 still waits for the others, because the item resolves only when every seat has passed.

**The "considering" chip.** It assumed an automatic pass clears within 800 ms. A human seat holding for up to 3 s would wrongly read as "considering a response". In a stack window, the chip's delay becomes the largest allowed hold (3 s) plus 800 ms. Outside a stack window it stays 800 ms. This keeps ADR 0009's property that an automatic pass, a real hold and a bluff look the same to the others. The hold is the same length whether or not the viewer could answer, so it reveals nothing either. This amends the "Considering a response…" section of ADR 0009's #1307 amendment.

### 3. The linger on resolution

**Display only.** Nothing waits for it. The board already shows the frame's end state, and input stays live under it, as in ADR 0053.

**The signal.** The server's public log is authoritative: a `resolve`, `fizzle` or `counter` entry newer than the last one the client handled, naming an item that left the stack between two frames. A spell's entry names it by `card_id`, which is the item's ID. For an ability the log needs one additive field:
- `LogEvent.stack_item_id`, set on `resolve`, `fizzle` and `counter` entries for both spells and abilities.
- The engine stamps `StackItemID` on the ability path's `EventResolve` and `EventFizzle` (`mutations.go:3470-3483`). The ability counter already carries the ID in `Target` (`effect_api.go:1931-1937`), and the projection copies it.
- `events[].stack_item_id` is already in the snapshot shape (`game/testdata/snapshot_shape/v7.txt:433`), so this writes a value into an existing field. There is no schema change and no protocol version change.
- PR 4 checks that no watcher keys on `StackItemID` for these kinds before stamping it.

**Amendment 2026-10-04:** the id travels in a new omitempty event field, `Event.ResolvedStackItemID`, because stamping `Event.StackItemID` would change frozen v7 fixture values (AGENTS.md §5). The bullets above that say the engine stamps `StackItemID` and that the field is already in the shape are superseded: the new key is recorded additively under v7, and `LogEvent.stack_item_id` on the wire is unchanged.

An item that leaves with no such entry (an undo, an effect that returns a spell to its owner's hand, a reconnect) gets **no** linger. It fades out in 200 ms, because a badge would be a claim the client cannot back. The first frame after mounting, a reconnect or a replay toggle primes the tracker without lingering, as `beatsPrimeKey` does for combat beats.

**What the player sees.** The departed card stays where it was drawn in the pile (or the compact row, or the lane tile) for 1.5 s, with an outcome badge:
- **Resolved** (green), for a resolve entry;
- **Countered** (red), for a counter entry, with the counter's name under it ("by Counterspell");
- **Fizzled** (grey), for a fizzle entry, titled "countered on resolution: no legal targets (CR 608.2b)".

The lingering copy is `aria-hidden` and outside the labelled section, so `board-layout.spec.ts`'s "no stack label on an empty stack" still holds. The announcer reads the same news ("Lightning Bolt resolved").

**Then it goes where it went.** In the next frame, the client finds the card's instance ID:
- on the battlefield (a permanent spell, CR 608.3a): it flies to its new slot `[data-instance-id]`, where `etbPulse` then plays;
- in a graveyard (an instant or sorcery, CR 608.2n; a countered spell, CR 701.6a; a fizzled spell, CR 608.2b): it flies to its owner's graveyard pile button;
- in exile (flashback): it flies to the exile strip;
- in the command zone: it flies to the command zone.

The flight is 400 ms × `animations.speed`, from cached geometry, with the GSAP portal-and-FLIP step `animations.ts` deferred (`:119-125`) built for this one case. Pile buttons gain a `data-pile` attribute for it. An ability (CR 608.2n, "ceases to exist"), a copy, or a card that went somewhere this viewer cannot see fades in place instead.

**Gating.** The badge is information, so it shows even with animations off, without motion. The flight is gated by `animations.cardPlay` under the master switch, so reduced motion means no flight. The 1.5 s does not scale with `animations.speed`. It is reading time, not motion.

**A burst.** When several items leave between two frames (bots passing fast), they linger in resolution order. Each gets `max(400 ms, 1.5 s / n)`, and at most three are queued. Older ones are dropped, and the log keeps the record. A new item arriving on the stack goes on top; the lingering card moves out to the pile's right, so it never covers the live top card.

**All five styles.** The linger is one overlay, anchored to the style's own element for the departed item. So compact rows and the experimental lanes linger too.

### 4. Arrows and rings while choosing targets

**The source glows.** While `targeting` is set, the element showing `TargetingState.card` gets a `targeting-source` glow in gold, outside the green "legal" ring's colour. That is the hand card being cast, the permanent whose ability is being activated, or the source of a trigger choosing its target (CR 113.7). It is found by `[data-instance-id]` wherever it is drawn: the hand fan, the battlefield, the command zone. If the source is not on screen (it left the battlefield, or it is in a hidden zone), the arrows start from the dock's question line instead.

**An arrow per pick.** Each picked target gets a curved gold arrow from the source, drawn with `stackArrows.ts`'s `boardCurve` and `edgePoint` in CombatArrows' board-sized layer (z 36). Picks are read from `targeting.picked`, so a multi-target spell shows one arrow per pick as the picks are made. A graveyard or exile pick gets no arrow, as on the stack today.

**An arrow that follows the pointer.** With a mouse or pen, while a pick is still open, one more arrow runs from the source to the pointer, throttled to animation frames. It snaps to a legal target under the pointer and turns green there. Touch and keyboard targeting get no follow arrow; they have no hover position.

**Rings on the stack, in every style.** Today only the fan rings what an item targets. `syncTargetMarks` moves into the host, so compact, spotlight, ribbon and pile all ring targeted permanents and player headers while the item is on the stack: gold for the top item, the caster's colour for the rest. Compact needs this too, since it is now the phone style.

The dock's question line is unchanged ("Select target for X", `s19-triggers` reads it). The arrows add to it and replace nothing.

### 5. A fuller game log

Two new log kinds, rendered by the server with the redaction the resolve line already uses (`projectAbilityItem` and `abilityName`, `log.go:606`, `:1775`, so a face-down source stays hidden in a trigger's line too).

- **`trigger`:** "Alice's trigger: Mulldrifter — draw two cards". It is projected from the two labelled `EventTrigger` emits: the harvested queue (`triggers.go:953-958`) and the manual announce (`mutations.go:3900-3905`). It is not projected from the activation's label-less breadcrumb (`activated.go:1824-1829`). PR 7 picks the discriminator, for example stamping the queued item's ID on the two trigger emits, and pins it with a test, rather than relying on an empty label. The line is written when the ability triggers. That is the same frame in which it reaches the stack, except while its controller is ordering triggers (CR 603.3b). Consecutive identical trigger lines (same controller, source, label and step) collapse into one with "×N", as draws do. `EventTrigger` leaves the gate test's silence list.
- **`activate`:** "Alice activated Prodigal Sorcerer — deal 1 damage to any target", from `EventActivateAbility` (`activated.go:1854`). The across case keeps its own line (`activate_across`). An activation already told by its own line gets no second one: cycling (`LogCycle`) is the one today, and the fold skips the activation that follows a cycle of the same card, the way it skips a sacrifice already told. An ability activated from a hand is revealed by the rules (CR 602.2a), so naming it leaks nothing.

**Mana abilities are not logged.** They do not use the stack (CR 605.3b), every land tap would be a line, and 200 entries would hold a fraction of a turn. `EventManaAbilityActivated` stays silent.

**The ring.** `PublicLogMax` stays 200. PR 7 measures, on a recorded four-seat game, how much of a turn cycle the ring covers with the new lines. It raises the cap only if that is less than one full cycle, and says so in the PR.

`docs/protocol.md`'s list of log kinds gains `trigger` and `activate`, and the `activate_across` entry's last sentence ("An activation by the permanent's own controller has no entry.") is replaced. The log is still closed by default (owner answer 3).

### 6. Tests and contracts

**Labels that stay.** `region "attention"` (exactly one). `stack: N on the stack` (absent on an empty stack). The verbatim item labels as visible text. `Counter`. `next`. The dock's "Select target for X". The new names are listed in §1.

**The tutorial.** The step 6 detour's "at the top of the screen" becomes "on the left of the table". That is true for the pile and for compact, whose card is top-left. The experimental centred lanes are opt-in, and the player who chose one knows where it is. The anchor (`DOCK`) and the completion are unchanged.

**Playwright.** Every client PR runs the nightly E2E on its branch before it merges (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`), and the nightly runs on `develop` after the last client PR.
- The s19 helper's seed gains `stackHoldMs: 0` beside `alwaysStopOpponentStack: true`. It holds by hand anyway, and resolves with `next`, which is never delayed.
- Any other spec whose waits are tighter than a 2 s hold, or a 2 s bot hold at normal speed, gets the same seed or a fast table. It does not get a looser timeout.
- One new spec covers the hold: an opponent's spell the viewer cannot answer is still on the stack 1.5 s after it appears, and gone within the hold plus 2 s.
- The linger is never in a labelled region, so no existing assertion sees it.

**Vitest.**
- The pile's placement as pure functions: the seam, the strip clamp, the coach-card clamp, the size clamp and the compact fallback.
- The `+N more` threshold.
- The migration: an untouched `compact` becomes `pile`; a chosen `fan` is kept; a new-version `compact` is kept; an unknown value becomes `pile`. The synced copy moves the same way.
- `stackHold.ts`: the viewer's own top item, an opponent's top item, an opponent's item under the viewer's own, the setting at 0, the bluff and hold combining as the larger delay, and a new frame resetting it.
- The considering chip's delay in and out of stack windows.
- The linger tracker as pure functions: matching log entries to departed items by `stack_item_id` or `card_id`; no entry, no linger; priming; burst scheduling and the cap of three; the destination lookup.
- The targeting arrows' plan: a source on and off screen, one arrow per pick, no arrow to a graveyard pick, and the follow arrow only for a pointer.
- Rings in every style.
- Render tests for the pile: the names above, the title as text, chips including `manual`, Counter disabled without priority, and the shrink while targeting.

**Go.**
- `aiseat`: the presets carry `StackHold`. A runner holds a pass on another seat's top item until first-seen plus the hold, tested with a short hold. It does not hold on its own item. It re-enumerates when the top item changes. A stepped runner and `FollowTablePace` off never hold.
- `protocol`: `stack_item_id` on spell and ability resolve, fizzle and counter entries. The `trigger` line for a harvested and a manual trigger, redacted for a face-down source, and none for the activation breadcrumb. Identical trigger lines collapse. The `activate` line, none after a cycle, and the across case unchanged.
- The kind gate passes with `EventTrigger` off the silence list.

---

## Delivery

Each PR goes into `develop`, Sprint S60, Issue #2204.

| PR | What | Needs | Parallel with |
|---|---|---|---|
| 1 | **This ADR,** the S60 section and index row in `docs/sprints.md`, and the AGENTS.md §3 ADR range line. Docs only. | — | — |
| 2 | **Client: the pile** (§1). `pile` in `STACK_STYLES` and `STYLE_BODIES`; the host's per-style placement and clamps; `StackLanePile.svelte`; the shrink while targeting and the collapse tab; the compact fallback at phone width and on short boards; the default, the Settings row and the migration; the tutorial copy (§6); its vitest; the nightly E2E on the branch. | 1 | 3, 4, 7 |
| 3 | **Hold, client and bots** (§2). `gameplay.stackHoldMs` and its Settings row; `stackHold.ts` and the timer in `Game.svelte`; the dock's countdown; the considering chip's delay; `StackHold` in `botPacePresets` and the runner's hold; the "Bot speed" hints; the e2e seed and the new hold spec; the vitest and Go tests; the nightly E2E on the branch. | 1 | 2, 4, 7 |
| 4 | **Server: `stack_item_id` on resolve, fizzle and counter log entries** (§3, the signal). The two ability emits, the projection, `docs/protocol.md`, and the Go tests. | 1 | 2, 3, 7 |
| 5 | **Client: the linger** (§3). The tracker, the overlay for every style, the badges, the announcer line, the flight and `data-pile`; its vitest; the nightly E2E on the branch. | 2, 4 | 6, 7 |
| 6 | **Client: targeting arrows and rings** (§4). The source glow, an arrow per pick, the follow arrow, and `syncTargetMarks` in the host for every style; its vitest; the nightly E2E on the branch. | 2 | 5, 7 |
| 7 | **Server and client: the fuller log** (§5). The `trigger` and `activate` kinds and the collapse, the gate test, `docs/protocol.md`, the client's log icons for the two kinds, the ring measurement, and the Go tests. | 1 | 2–6 |

Settings fields land in PRs 2 and 3. Whichever merges first takes the next `SETTINGS_VERSION`, and the other takes the one after. The pile migration is keyed on the version PR 2 introduces.

After PR 7: run the nightly E2E on `develop`, then close #2204 with evidence: the run, and a cmd-dev game against a bot at normal speed in which the bot's spell sits on the stack for about 2 s, lingers marked resolved, and appears in the log with its trigger.

## Consequences

- The stack is a pile of readable cards on the left. The hover zoom keeps the top right, and the old styles are one setting away.
- An opponent's or a bot's spell stays up for at least about 2 s before it resolves, even when nobody can answer it. Against one bot at normal speed that is about 1.3 s longer than today's 0.7 s. At a four-seat table the extra is 0.6–1.3 s, depending on where the person sits in turn order, because the bots' own think times already cover part of it. A player who wants the old speed sets the hold to 0, and a host sets the bots to fast.
- Resolving, countering and fizzling are visible for 1.5 s, and where the card went is shown, without slowing the game.
- Choosing a target looks like the result: the source glows and the arrows it will carry are drawn as it is chosen. What an item targets is ringed in every style, not only the fan.
- The log tells triggers and activations as they happen, not only when they resolve. A face-down source stays hidden.
- The "considering a response…" chip appears later in a stack window (3.8 s instead of 0.8 s).
- The pile covers part of the left side of both rows while the stack is live. The shrink while targeting and the collapse tab are the ways past it.
- One additive log field and two log kinds. No protocol version, no snapshot schema change, and no new persisted state beyond two settings.

## Out of scope

- **A phone layout** for the stack (owner answer 3). Phones keep compact.
- **Opening the log by default** (owner answer 3).
- **Undo, Cancel and Submit** in Arena's corner. The dock already has Undo, Done and Cancel (ADR 0111 §6).
- **Logging mana abilities** (§5).
- **Changing the trigger-order sheet** or the APNAP drain (ADR 0018). The pile shows pending triggers; it does not order them.
- **A server-side hold.** The server does not delay broadcasts or resolutions. Every hold is a seat taking its time before passing.
- **Removing the experimental lanes.** They stay selectable, as the owner asked. Retiring any of them is a later decision.

## Calls made here

The owner's answers did not settle these. Each is decided above, and each can be overturned in review:

1. **The pile is a style inside `StackLaneHost`,** not its own host (§1), so the header, summary, controls and announcer stay single.
2. **Placement:** 12px from the left, centred on the opponents/viewer seam, clamped below the attention strip's content and above the coach card and dev dock (§1).
3. **Size:** the top card is the `normal` image at `clamp(200px, 30% of board height, 280px)`, at most 24% of the board's width. An ability is its source's card with a caption band (§1).
4. **Depth:** four lower items peek at 34px each, and deeper stacks show "+N more", which opens a scrolling list (§1).
5. **The pile shrinks while the viewer chooses a target,** and has a collapse tab, so the cards under it can be reached (§1).
6. **Migration:** an untouched pre-version `compact` becomes `pile`, because the stored blob cannot tell untouched from chosen. A stored fan, spotlight or ribbon is kept (§1).
7. **Phones (≤599px) and short boards render compact** without changing the setting (§1).
8. **The hold setting** is `gameplay.stackHoldMs`, synced, with choices 0–3 s and default 2 s. It lives under gameplay, not display, because it changes when the client passes (§2).
9. **The hold covers every automatic pass,** the session autopass toggle included. It is keyed on the top item not controlled by the viewer, and measured from when this client first saw it. A bluff and the hold combine as the larger delay (§2).
10. **The considering chip waits 3.8 s in a stack window** instead of 0.8 s, so a hold does not read as thinking (§2; amends ADR 0009).
11. **The bot hold is part of the bot speed preset** (fast 0, normal 2 s, slow 3 s), not a new table setting and not a guess at the people's settings (§2).
12. **The linger's signal is the public log** plus an additive `stack_item_id`. An item that leaves without a resolve, counter or fizzle entry gets no linger (§3).
13. **The linger badge shows even with animations off.** Only the flight is gated, by `animations.cardPlay`. A burst caps at three lingering items (§3).
14. **The follow arrow is for a pointer only,** not touch or keyboard (§4).
15. **The trigger line is written when the ability triggers.** Identical consecutive lines collapse. Mana abilities are not logged. The ring stays 200 unless PR 7's measurement says otherwise (§5).
16. **The tutorial copy says "on the left of the table",** which is true for the pile and for compact (§6).
