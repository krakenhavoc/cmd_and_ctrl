# ADR 0018 — Triggered abilities use the stack

**Status:** Implemented · 2026-09-02 · Branch `feat/s19-triggers-on-stack`
**Addendum:** 2026-09-17 · **Accepted** · [Trigger doubling (CR 603.2d)](#addendum-2026-09-17-trigger-doubling-cr-6032d--accepted) · tracked on [#752](https://github.com/krakenhavoc/cmd_and_ctrl/issues/752)
**Amendment:** 2026-09-24 · **Accepted** · [Resolution-time LKI for a permanent that has left (CR 608.2h)](#amendment-2026-09-24--resolution-time-lki-for-a-permanent-that-has-left-cr-6082h--accepted--s38) · [#1379](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1379)
**Amendment:** 2026-09-24 · **Accepted** · [A batch that commutes needs no CR 603.3b order](#amendment-2026-09-24--a-batch-that-commutes-needs-no-cr-6033b-order--accepted--s48) · [#1511](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1511)
**Amendment:** 2026-09-24 · **Accepted** · [A trigger still announcing holds its batch (CR 603.3b / 603.3d)](#amendment-2026-09-24--a-trigger-still-announcing-holds-its-batch-cr-6033b--6033d--accepted--s39) · [#1529](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1529)
**Amendment:** 2026-10-11 · **Accepted** · [The independence check, on the boards people play](#amendment-2026-10-11--the-independence-check-on-the-boards-people-play-3013--accepted) · [#3013](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3013)
**Amendment:** 2026-10-12 · **Accepted** · [An identical batch is asked again when its copies read their subjects](#amendment-2026-10-12--an-identical-batch-is-asked-again-when-its-copies-read-their-subjects-3021--accepted) · [#3021](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3021)
**Amendment:** 2026-10-13 · **Accepted** · [The order prompt says why it asked](#amendment-2026-10-13--the-order-prompt-says-why-it-asked-3015--accepted) · [#3015](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3015)
**Amendment:** 2026-10-13 (second) · **Accepted** · [The six remaining villainous-choice cards](#amendment-2026-10-13-second--the-six-remaining-villainous-choice-cards-3143--accepted--backlog) · [#3143](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3143)
**Amendment:** 2026-10-15 · **Accepted** · [A became-target event says what kind of object targeted](#amendment-2026-10-15--a-became-target-event-says-what-kind-of-object-targeted-3779--accepted) · [#3779](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3779)

## Context

S19 sub-PRs 1–5 shipped the auto-fire trigger pipeline: a per-game
`triggerHarvester` listener walks the battlefield on every event,
matches `Spec.Triggered` declarations, and calls `Build` to produce
a `StackItem`. Twelve catalog cards (ETB, dies, upkeep) ride it.

What they actually did, though, was apply the effect *inside*
`Build` and return `nil`. Mulldrifter drew two cards the instant
the creature resolved; Phyrexian Arena took its life and card as
the upkeep step began; Reclamation Sage destroyed the artifact the
moment its controller clicked "Yes". Nothing ever reached
`PendingTriggers` or `StackMeta`. The card files called this a
sandbox simplification "deferred to a future engine-completeness
sweep", and the engine had the matching gap: `resolveTopAbilityLocked`
deleted the item and emitted `resolve` with no way to run anything.

That shortcut cost three things:

- **No response window.** A trigger that resolves at announce time
  can't be answered — no counter-the-ability, no sacrifice in
  response, no bounce-the-target. S19's own exit criterion
  ("advance to upkeep with Phyrexian Arena → trigger goes on stack
  automatically") wasn't met.
- **No CR 608.2b re-check.** Targets picked at announce time were
  acted on immediately, so "target left before resolution" could
  never happen — and neither could the fizzle a correct engine
  produces.
- **A migration debt that grows per card.** Sub-PRs 6 and 7 add
  eight more trigger cards. Every one written in the inline pattern
  is one more to rewrite later.

The stack itself already supported ability items (S13.1 manual
`AnnounceTrigger`, `ActivateAbility`, the LIFO `Seq` fix in #202,
`CounterAbility`). Only the resolution effect was missing.

## Decisions

### 1. `StackItem.Effect` — a resolution callback on the item

```go
Effect func(g *Game, item *StackItem) error
```

`resolveTopAbilityLocked` now mirrors the spell path: remove the
item from `StackMeta`, run `spellAllTargetsIllegalLocked` (every
targeted slot illegal → `EventFizzle`, no effect), emit
`EventResolve`, then run `Effect`. An error from `Effect` surfaces
as `EventEffectError` and does not wedge the stack — the ability
has ceased to exist either way (CR 608.2n). `nil` keeps the S13.1
manual-sandbox meaning: the players resolve it by hand.

Why a closure on the item rather than routing through the
`EffectResolver` hook by oracle ID + ability index: the trigger's
announce-time context (which event fired it, the CR 603.10 LKI
snapshot, the auto-picked target) is already captured at `Build`
time, and a closure is the cheapest way to carry it to resolution.
The `triggerResumeFrame` on `PendingChoice` set the precedent for
func-valued state in the game struct. Games aren't serialised (the
protocol package projects its own views; undo goes through
`Clone`), so a func field costs nothing on the wire.

Two rules follow from undo: `cloneStackItem` copies `Effect`, and
the callback **takes the live `*Game`** instead of capturing one,
so an item restored from a snapshot resolves against the restored
game. For the same reason it must not capture `*Card` pointers into
zone slices — it reads controller / source / targets off the
`item` it receives. `NewTriggeredItem(source, label, effect)` is
the one-liner that gets the shape right.

### 2. `Build` builds; it does not resolve

`TriggeredAbility.Build` returns the item and **must not mutate
game state**. Targets are chosen here (CR 603.3d — when the
ability is put on the stack) and stamped onto `item.Targets`, so
the stack overlay shows them and the resolve-time re-check applies.
A targeted trigger with no legal target returns `nil` and never
reaches the stack, which is also what CR 603.3d says.

All twelve S19 cards were migrated. The dies-trigger and upkeep
cards were mechanical; Acidic Slime and Reclamation Sage share a
`destroyTargetTrigger` helper; Eternal Witness carries its
top-of-graveyard pick as a target and lets `ReturnFromGraveyard`
report `ErrCardNotFound` if the card was exiled in response.

### 3. Every priority boundary drains — including the ones that
### didn't before

A harvested trigger goes `PendingTriggers → StackMeta` via
`drainPendingTriggersAPNAPLocked`, which ran only from the
priority-wrap and step-advance paths. Two paths queued triggers and
then left them stranded until the next wrap — where an empty stack
would advance the step *before* draining, putting the trigger a
step late:

- `ResolveTriggerPrompt` (answering "Yes") now calls
  `runStateChecksLocked` — answering is the moment the ability is
  put on the stack (CR 603.3).
- `MoveCardByID` (the sandbox move_card verb) and the land branch
  of `CastSpell` now call `runStateChecksLocked`. Both are special
  actions after which the actor keeps priority (CR 116.3), and
  CR 117.5 puts SBAs + the trigger drain at exactly that boundary.
  A catalog creature dropped straight onto the battlefield gets its
  ETB trigger on the stack immediately.

### 4. Wire order is by `Seq`

`viewOfStackItemsInStackOrder` appended ability items in map
iteration order after the spell cards. With abilities on the stack
every turn, that's visibly random. It now sorts by insertion `Seq`
— the same order the engine resolves in — with stable ties keeping
the legacy spell-cards-first behaviour for zero-`Seq` snapshots.

### 5. Client: autopass passes through triggers; autopass off is full control

The S13.6 auto-pass gate treated "stack non-empty" as "a card is
on `Game.Stack`", so ability items didn't count and the
conventional (autopass-off) path would sail past every trigger —
Counterspell in hand or not. The gate now stops for ability items
exactly as it does for spells. The two modes mirror Arena:

- **Autopass on** (the session toggle): every priority window
  passes, triggers included — "get me through this turn".
- **Autopass off**: full control. Every spell *and* every trigger
  on the stack is a window the viewer answers by hand. Step stops
  and the smart-skip predicate still govern empty-stack windows
  as before; they never skip a non-empty stack.

No smart-skip escape for triggers: a player who wants routine
upkeep triggers to fly by turns autopass on.

### 6. "Unless that player pays" is a pending choice queued at resolution

Rhystic Study, Smothering Tithe and Esper Sentinel (sub-PR 6) ask
a *different* player a question when the trigger resolves. That is
`PendingChoicePayUnless`: the trigger's `Effect` queues it for the
taxed player (`PayUnless` primitive) and returns; the ability has
resolved and left the stack. The answer arrives as
`resolve_choice {apply}` like the other yes/no kinds. "Pay" spends
from the chooser's pool, auto-tapping their untapped sources first
via the S15 planner (no exclusions); a "Pay" they can't cover
degrades to a decline. Decline runs the card's consequence
(`OnDecline`) against a fresh `Context` for the original item.

Sandbox looseness, accepted: priority isn't gated on the pending
prompt, so play can continue while a Rhystic tax is unanswered.
It matches how the table plays it in paper ("you paying for that?"
while the next spell is already being cast) and avoids a modal
lockstep across four browsers.

**Amendment (2026-09-17, #730): the looseness is `pay_unless` only.**
Nothing refused `advance_step`, or a priority pass that would advance
the step, while *any* prompt sat in `Game.PendingChoices` — so the
sandbox latitude this section granted Rhystic Study was in practice
granted to every prompt in the engine, and a sacrifice, a scry or a
CR 616 ordering could be walked past and then answered against a game
that had moved on (#701 and #725 each had to add a stale-frame guard
for exactly that; #651 reports the same shape for effect discards).

`Game.blockingChoiceLocked` (`server/internal/game/choice_gate.go`) now
gates `AdvanceStep`, `PassPriority` and `PassTurn`, which return a
`*ChoicePendingError` (wrapping `ErrChoicePending`, naming the choice's
ID and kind) while a blocking prompt is outstanding. It is
**deny-by-default with an explicit allowlist**, so a choice kind added
later blocks without its author having to know this rule exists. The
allowlist, verbatim:

- **`pay_unless` — does not block.** The reason this section already
  gives: the ability has resolved and left the stack, the question is
  addressed to a *different* player, and the answer spends from that
  player's pool or runs `OnDecline` — neither reads the step. Play
  continues, as it does in paper.

Every other kind **blocks**: `discard_from_hand`, `mana_pick`,
`replacement_order`, `optional_replacement`, `damage_assignment`,
`trigger_prompt`, `trigger_order`, `pick_target`, `sacrifice_choice`,
`scry`, `surveil`, `look_at_top`, `search_library`, `may_cast`,
`choose_protector`, `legend_rule`, `choose_color`, `confirm`,
`choose_cards`, `entry_pay_life`, `copy_target`,
`choose_creature_type` — and anything added after this line.

`mana_pick` is the near miss worth recording: it looks like background
bookkeeping, but an unanswered colour pick is mana that has *not*
entered the pool, `internal/legal` already offers a seat nothing else
while one is open, and the auto-tapper refuses to create one precisely
because its contract is "no further player decisions required". It
blocks.

**Amendment (2026-09-18, #567): the latitude is per PROMPT, not per
kind, and cumulative upkeep's does not take it.** CR 702.24's "put an
age counter on this permanent, then sacrifice it unless you pay its
upkeep cost for each age counter on it" is a `pay_unless` — the same
prompt Rhystic Study raises, with the cost rebuilt each upkeep from the
counters. Every reason this section gives for letting the table walk
past one is Rhystic Study's and none of them survives the move: the
question is addressed to the **active player**, during their **own**
upkeep, and what hangs on the answer is whether a permanent is still on
the battlefield for the rest of the turn. So this prompt blocks.

It blocks through `PendingChoice.ForceBlocks`, a per-prompt flag read
by the one predicate `game.ChoicePromptBlocksTable`, which the engine's
gated verbs and `internal/legal` both call — the #794 rule that there is
exactly one answer to "does this stop the table" is preserved, and
`ChoiceBlocksTable(kind)` stays the answer for a KIND and stays what
`TestEveryChoiceKindIsClassifiedAndEnumerated` checks. The override is
**one-way**: it can only make a prompt block, never let one through, so
the deny-by-default direction is intact and the allowlist above is
still the whole of it. `pay_unless` remains the one non-blocking kind,
and Rhystic Study, Smothering Tithe and Esper Sentinel play exactly as
they did.

Queue a blocking one with `Game.QueueBlockingPayUnlessForEffect`
(card side: `PayUnless{Blocking: true}`). The rule of thumb the two
cases give: a pay-unless addressed to somebody ELSE after the ability
has resolved does not block; one addressed to the player whose turn it
is, about their own permanent, does.

What the gate does **not** do: answering a prompt (`resolve_choice`
and its kin), `concede`, chat, undo, and the admin context menu's raw
sandbox moves (`move_card`, `change_life`, `add_counter`,
`mark_damage` — ADR 0033 §8's list) are all untouched, so a table can
always be unstuck by hand. Casting and activating are not gated either;
`internal/legal` declines to offer them, and widening the refusal that
far is a bigger behaviour change than this bug needs.

**Amendment (2026-09-18, #794): one predicate, `game.ChoiceBlocksTable`,
read by the engine and by the bots.** The allowlist above was read only
from inside `choice_gate.go`, and `internal/legal` kept a second,
stricter rule of its own: `anyChoiceOpen` returned an empty move list
for *every* seat while *any* prompt sat in the queue, `pay_unless`
included. The two disagreed for the whole life of the allowlist. The
engine let the table play on through an unanswered Rhystic tax exactly
as this section decided, and the enumerator told every bot seat it had
nothing to do until a human answered it — so a bot idled on a question
that was not addressed to it, which is neither what this section
decided nor what a human at the same table may do.

The allowlist is now a full classification, one row per kind, in
`choiceGateDecisions` (`server/internal/game/choice_gate.go`), read
through the exported `game.ChoiceBlocksTable(kind)`. The gated verbs
ask it (`blockingChoiceLocked`) and so does the enumerator
(`legal.anyBlockingChoiceOpen`), so there is one answer to "does this
prompt stop the table" and no second list to drift from. The decisions
themselves are unchanged: `pay_unless` is still the only kind that does
not block, and an *unclassified* kind still blocks — `ChoiceBlocksTable`
is deny-by-default for exactly the reason this section already gives.

A seat that owes a non-blocking prompt is still offered that prompt's
answers and nothing else. The prompt is a question already in front of
it; answering is one dispatch, and the next window carries the seat's
whole turn. What changed is only what the *other* seats may do.

**The classification of the kind added since (#804).** `loop_shortcut`,
the CR 732 shortcut prompt, **blocks**. It is the one entry worth arguing
about, because ADR 0055 §4 was careful that the loop breaker refuse no
passes: the shortcut is proposed while the loop's trigger is still on the
stack, and what the answer decides is how many times that trigger resolves
next, so a table that could pass through the question would be answering it
by doing. It is never queued to a seat that has left, so blocking cannot
wedge a table.

Adding a kind is now gated rather than trusted:
`TestEveryChoiceKindIsClassifiedAndEnumerated`
(`server/internal/legal/choice_gate_test.go`) reads every
`PendingChoiceKind` constant out of `internal/game` and fails unless it
has a row here *and* a case in `legal.choiceMoves`. Those are the two
silent failures a new kind used to be able to ship with — a gate
decision nobody made, and an empty move list for the seat that owes it
(the #499 / #618 wedge).

**Amendment (2026-09-18, #796 / #568): a resolving effect may ask a
question, and it may ask it of somebody else.**

This section shipped the one prompt a resolving ability could put in
front of another player, and welded it to a *mana payment*. Three
shapes were left with nowhere to go, and each blocked real cards:

- **A free yes/no for the effect's own controller.** "Mill two cards.
  Then you may sacrifice this land" (Eden, Seat of the Sanctum). A
  trigger's CR 603.5 "you may" is asked before the ability goes on the
  stack and is over by the time anything resolves; `pay_unless` speaks
  about mana; search / scry / put-from-hand each carry their own
  decline because declining is part of that instruction.
- **The same yes/no, addressed to an opponent.** "Target opponent may
  have you draw three cards" (Combustible Gearhulk), "they may tap
  that permanent" (Charismatic Conqueror), "loses 5 life unless they
  discard a card" (Painful Quandary).
- **A choice among three or more consequences, addressed to an
  opponent.** "Each opponent loses 3 life unless that player
  sacrifices a nonland permanent of their choice or discards a card"
  (Torment of Hailfire), and the two piles of a Fact or Fiction split.

**One primitive, one new kind, and no new yes/no kind.**

`effects.MayChoice{Player, Question, YesLabel, NoLabel, LifeCost,
OnYes, OnNo}` (`server/internal/cards/effects/may_choice.go`) is the
yes/no, and it is built on the existing `PendingChoiceConfirm` — the
chained-choice two-way prompt from #552, which is already a question
whose branches are plain continuations supplied by the card, already
addressed by `Chooser`, already classified here, already enumerated,
already rendered. A `may` kind would have been a fourth spelling of a
question the queue can ask, and every consumer would have needed a case
for it. `Player` defaults to the effect's controller, which is what
"you may" means; #568's cards set it to another seat and the prompt is
otherwise identical.

`PendingChoiceOptionPick` (`option_pick`,
`server/internal/game/option_pick.go`) is the new kind: "choose one of
the following", addressed to any seat, answered with the INDEX of the
chosen option (`{option_index: N}`, routed by kind because zero is the
commonest answer). Its options are plain data — a label, a life cost,
and optionally the cards the option is about — and what an option MEANS
lives in the frame the queuing effect supplies, which is what keeps it
reusable rather than a second modal-spell system. A modal spell's
"choose one —" is NOT this kind: that choice is made at announce
(CR 601.2b) and lives on `Spec.Modes`.

**A pile split needs no kind at all.** It is two chained prompts to two
different seats (#552): a `choose_cards` addressed to the splitter with
a floor of zero, and an `option_pick` addressed back to the controller
whose two options each carry a pile. `Game.QueuePileSplitForEffect` /
`effects.PileSplit` is that composition and nothing more.

**Legality lives at queue time.** `ResolveOptionPick` validates the
index and nothing else, exactly as `ResolveConfirm` validates nothing
about the board. An effect builds its option list out of what the
chooser can actually do (CR 608.2's "as much as possible") — Torment
drops "sacrifice a nonland permanent" for a player who controls none —
and must put a branch that always works FIRST, because that is the one
`legal.choiceMoves` marks `AlwaysLegal`. A prompt whose every branch
could fail is a seat that can be stuck (#544).

**`option_pick` blocks the table**, like every other resolution-time
decision: the effect that asked it is paused mid-resolution and its
continuation is the rest of the card. `pay_unless` remains the only
kind that does not block, and the reason is unchanged — it is asked
*after* its trigger has left the stack.

**Redaction: one function, in the projection.** A prompt addressed to
seat B over seat A's cards is new, and `PendingChoiceView` shipping
cards a viewer should not see is the leak PR #513 fixed. All of it is
now `protocol.redactChoiceCards`, the only answer to "which of a
prompt's cards may this viewer see", applied to `Options` and to every
option's own `cards`:

1. A viewer who is not a knower of a card does not get it — dropped,
   not redacted to a back, because these lists come out of hidden
   zones and a stable instance ID is a correlation handle.
2. The CHOOSER keeps their whole list as answerable backs only when
   the pool is their OWN material, or when it is another player's
   HAND — a hand's size is public (CR 400.2), and a coercive discard
   that revealed nothing still has to be answerable by picking one of
   the backs. A pool that is neither (Fact or Fiction's five cards off
   the top of a LIBRARY) shows the chooser only what was revealed.

Rule 2 is therefore also a contract on the card: a cross-seat prompt
over a non-hand zone must reveal what it asks about (#549), or its
chooser is handed an empty list. That is the right failure — the
alternative leaks a hidden zone, and every printed card of this family
reveals first.

Two more drains fell out of the sub-PR 6 cards: `CastSpell` (the
caster gets priority right after casting, CR 117.3c — Rhystic's
trigger must be on the stack by then) and the sandbox `draw_card`
verb (Tithe / Sphinx watch draws). `Game.SpellsCastThisTurn` is
bumped before `EventCast` fires so "first noncreature spell each
turn" reads `Noncreature == 1` for the spell that triggered it.

**Amendment (2026-09-18, #914): `advance_step` passes priority until
the step ends. CR 117.4.**

The #730 gate above stops the cursor for a PROMPT. Nothing stopped it
for the STACK, and a step that ends owing something on its stack is
the one thing CR 117.4 forbids: "a step or phase ends when all players
pass in succession with an empty stack." So `AdvanceStep` walked past
a trigger the step had already announced, and the next step's
turn-based action happened first — combat damage before the afflict
out of `declare_blockers`, regular damage before the first-strike
damage triggers out of `first_strike_damage` (ADR 0045's amendment
recorded both as open), an upkeep trigger resolving after the draw.
It was never about combat: the cursor walked past anything.

**The verb now means "pass priority until this step ends."** With an
empty stack that is exactly what it always was — one move of the
cursor, no pass, no extra event. With something on the stack it is the
passes the rules require first: `AdvanceStep` drives `PassPriority`
for the caller until the stack is empty, so each resolution is
followed by state-based actions, the CR 603.3b trigger drain and
another priority round, in the order a table of humans clicking "next"
would produce. It is `PassPriority`'s own body under the lock
(`passPriorityLocked`) — one priority engine, not a second one beside
it — and the cursor moves at the end through the same
`advanceCursorLocked` as before.

**The drive stops on the three things that stop automatic passing
anywhere else**, each with the cursor left in the step that still owes
something and **no error**, because the resolutions it already made
are real and the caller has to see them (`Room.apply` broadcasts
nothing and mints no undo entry for a failed dispatch):

- a **blocking prompt** raised by one of the resolutions — this
  section's own rule, now applied to a prompt the verb itself caused;
- a **CR 732 loop notice** ([ADR 0055](0055-loop-breaker.md)), checked
  AFTER a pass so a standing notice still lets one manual nudge
  through, exactly as the client's "next" button does;
- the **game ending** under a resolution.

A prompt that was already open when the verb was called is still the
`*ChoicePendingError` refusal above: nothing has happened yet, so
there is nothing to broadcast.

**Bots are unaffected** — `internal/legal` has never enumerated
`advance_step`, and a bot reaches the same board by passing priority,
which is what the drive does on its behalf. **Three test expectations
changed** rather than being silenced, each to the rules answer:
Drana's first-strike trigger grows the attackers before regular
damage (2 + 3, not 2 + 2), Professional Face-Breaker's first-strike
Treasure is already made when the cursor reaches the regular step, and
`TestB492…` is flipped to "afflict, then damage". Cyberman Patrol's
caveat — the last of #388 — is gone with them.
**Amendment (2026-09-18, #929): choosing a PLAYER is the same kind
again.**

"Choose a player" / "choose an opponent" (Gluntch, the Bestower;
Skullwinder; Slithermuse; the still-unwritten "an opponent of your
choice gains control") is a closed list of things with exactly one
answer, which is the question `option_pick` already asks. So it gets
**no kind of its own**: `Game.QueueChoosePlayerForEffect`
(`server/internal/game/choose_player.go`) queues a
`PendingChoiceOptionPick` with one option per eligible seat, labelled
with that seat's name, and every consumer — the choice gate, the
enumerator, the wire projection, the client modal — answers it
unchanged. The card-facing shape is
`effects.ChoosePlayer{Chooser, Among, Except, Question, Then}`
(`cards/effects/choose_player.go`), with `Among` one of `Players`,
`Opponents` or `OpponentsOf(id)`.

**It is not a target, and it is not an ETB choice.** A target is named
at announce (CR 601.2c), is public from that moment, and is re-checked
at resolution (CR 608.2b); a chosen player is named while the effect
resolves and nothing may respond to it — which is why a Slithermuse
trigger cannot be fizzled by the opponent it was aimed at. The CR
614.12 form ("AS this enters, choose a player" — True-Name Nemesis) is
a third thing again: a replacement-time choice stored on the permanent
and read for the rest of its life, the shape `ChooseColorAsEnters`
already has for colours (#742, `Card.ChosenColor`).

**It is designed elsewhere and is not built here.** Protection has
since shipped ([ADR 0072](0072-protection.md), #662), and its §7
already specifies this field down to the detail — `Card.ChosenPlayer`
on `Card.ChosenColor`'s pattern, classified `carried`, cleared on every
zone change, written by an as-enters (CR 614.12) sibling of
`QueueChoosePlayerForEffect`, with the protection grammar resolving
"the chosen player" against it so no raw UUID ever reaches a display
string. That work is **#980**, and True-Name Nemesis waits on it; this
amendment deliberately does not restate the design, because two copies
of it would drift. What #929 owes #980 is the resolution-time prompt,
and that is what it ships.

**Amendment (2026-09-19, #980 / #994): the as-enters form landed on
this kind too, and the kind learned to prune itself.**

`Game.QueueChoosePlayerAsEntersForEffect` (`game/choose_player.go`) is
the CR 614.12 sibling, and it is the SAME `option_pick`: same option
list, same gate row, same enumerator case, same wire projection.
Nothing about this amendment's "no kind of its own" changes. The only
difference is where the answer goes — `Card.ChosenPlayer` on the
permanent, rather than a `TargetPlayer` ref on one item's payload —
and that difference is a lifetime, not a prompt. The design is ADR
0072 §7 and its #980 amendment; this one still does not restate it.

What DID change about the kind is worth recording here, because it is a
property of `option_pick` and not of either player prompt.
**`ChoiceOption` can now name a seat** (`ChoiceOption.Player`, #994),
`uuid.Nil` on every option that is not about one. Two consequences:

- **An open option list can SHRINK.** A seat that leaves the game is
  pruned off every open prompt that offers it (CR 800.4a) — by
  `reassignChoiceLocked` when the chooser is the one leaving, and by
  `pruneDepartedSeatOptionsLocked` when anybody else is. Before this,
  eligibility was filtered exactly once, at queue time, and a prompt
  went on offering a player who was not a player. A list pruned to
  nothing is dropped, which is the state `QueueChoosePlayerForEffect`
  refuses to queue in the first place and the departure table's own
  `dropDiscard` for this kind.
- **A seat prompt is therefore answered with its SEAT, not its index.**
  `OptionPickPrompt.ThenSeat` is the continuation for an option list of
  players; `ResolveOptionPick` reads `PickOptions[index].Player` off the
  list the chooser was shown and hands that to the branch. The index
  form (`Then`) is unchanged and is still what every consequence-shaped
  prompt uses — a pile, a Torment branch — because those options do not
  move. A player prompt's do, and a closure holding the candidate slice
  it was built with would record the player one seat along.

**The answer rides `StackItem.Payload`.** #636 already carries "what
the effect that created this item had to tell it" as `[]TargetRef`; a
chosen player is a `TargetPlayer` ref and goes there rather than on a
second payload field. `Targets` was not an option: a `pick_target`
answer replaces `Targets` wholesale, so a player parked there would be
erased by the next re-target prompt. A clause that asks twice appends
twice in ask order, which is what makes Gluntch's "a SECOND player"
expressible — `ctx.ChosenPlayers()` is the exclusion list and
`ctx.ChosenPlayer()` is the most recent answer. A question that could
not be asked records a `TargetNone` marker rather than nothing, so a
later clause reads "nobody was chosen" instead of the previous
clause's player.

**Eligibility and order.** A seat that has left the game is not a
player (CR 800.4a) and is never offered. The remaining seats are
offered most-life-first, ties by seat, and that ordering IS the bot
policy: `legal.choiceMoves` marks an option pick's first branch
always-legal, and `docs/bot.md`'s posture for a prompt from somebody
else's card is "price what you can see and take the first offer
otherwise". Reassigning a prompt whose chooser leaves AFTER it is
queued is `reassignChoiceLocked` (`game/pending_choice.go`, the CR
800.4g/h/i function that landed with the CR 800.4 remainder in #959),
not this primitive's business. What this primitive owns is the other
moment — a chooser who is already gone when the question would be
asked — and there the prompt is simply not queued and the absence is
recorded.

**Amendment (2026-09-19, #951): the latitude is Rhystic Study's SHAPE,
not the word `pay_unless` — a prompt whose decline counters an object
on the stack blocks while that object is there.**

Ward borrows this section's prompt (CR 702.21a is a triggered ability
whose resolution is "counter that spell or ability unless its
controller pays"), and so do Daze, Dazzling Denial, Izzet Charm, Mystic
Confluence and Spell Stutter. Every reason §6 gives for letting the
table walk past a pay-unless is Rhystic Study's, and not one of them
survives the move: the question is not background bookkeeping, it is
*whether the spell underneath it resolves*. A table that resolves that
spell has ANSWERED the question by doing it — for free, against the
payer, with the ward tax skipped. #951 reproduced it with Diffusion
Sliver: any third seat passing priority before the payer answered
killed a warded permanent for nothing. CR 117.4 does not let the top of
the stack resolve while a required action is outstanding, and CR 608.2
makes the counter part of the resolution that asked the question.

So the halt is **derived, not declared**. `PendingChoice.GuardsStackItem`
records what the decline is about — the "that spell" — and
`Game.ChoicePromptBlocksTable` (now a method, because the answer is a
question about the board rather than about the prompt) blocks while
that object is still on the stack. The alternative, a second per-card
"please block" switch beside `ForceBlocks`, was rejected for the reason
the six cards above are the evidence for: every one of them was written
after §6 and every one of them got it wrong the same way. A card says
what its text says; the engine works out what that means for the
cursor.

One door, `Game.QueueCounterUnlessPaidForEffect`
(`server/internal/game/counter_unless_paid.go`), with
`effects.CounterUnlessPaid` as its card-side primitive. It also absorbs
the two checks all six cards had been spelling out for themselves: an
object that has already left the stack raises no prompt at all (there
is nothing to counter and so nothing to charge for, CR 118.12), and the
decline re-checks before countering.

**The halt cannot outlive its question, which is what keeps it from
being a wedge.** The block is a live read, so a guarded spell that
leaves the stack some other way — countered underneath the trigger,
fizzled — takes the halt with it and the table plays on without anybody
answering. The prompt's own chooser is never gated by it; a chooser who
leaves the game is settled by the departure table's `dropDecline`
column (CR 800.4f, the 2026-09-18 #961 amendment to
[ADR 0060](0060-leaving-the-game.md)), which counters the spell and
frees the table; and `internal/legal` reads the same predicate, so a
bot seat is offered the answer and answers it.

**What is unchanged.** `pay_unless` is still the one kind classified
non-blocking, `ChoiceBlocksTable(kind)` is still the answer for a KIND
and still what `TestEveryChoiceKindIsClassifiedAndEnumerated` checks,
and Rhystic Study, Smothering Tithe, Esper Sentinel, Mystic Remora and
Kazuul play exactly as they did — their declines guard nothing. Both
per-prompt narrowings remain **one-way**: they can only make a prompt
block, never let one through.

**Noticed and not changed here.** Stasis ("at the beginning of your
upkeep, sacrifice Stasis unless you pay {U}") and Pact of Negation
("pay {3}{U}{U} … if you don't, you lose the game") are the #567 shape
rather than this one — the active player, their own upkeep, a
consequence that is not a stack object — and neither sets
`PayUnless.Blocking` today. They are a separate judgement about the
same section and are left for one. *(That judgement is the amendment
below, 2026-09-19 / #997, which also retired `PayUnless.Blocking`.)*

**Amendment (2026-09-19, #997): the upkeep shape blocks, and it blocks
because the ENGINE reads the cursor — `PayUnless.Blocking` and
`PendingChoice.ForceBlocks` are gone.**

The judgement the amendment above deferred: **yes, it blocks.** "At the
beginning of your upkeep, pay <cost> or <lose something you cannot get
back>" — Stasis, Pact of Negation, every cumulative upkeep — is asked
of the **active player**, during their **own upkeep**, and what hangs
on the answer is whether a permanent is on the battlefield for the rest
of the turn, or whether the player is still in the game. That is the
reasoning the #567 amendment already accepted for CR 702.24; nothing
about it is specific to cumulative upkeep. CR 117.3 does not pass
priority on with a required action outstanding, and CR 500.4 does not
end a step until it has been taken.

**The interesting half is why the fix is not "set the flag on two more
cards".** #567 shipped exactly that flag. Stasis and Pact of Negation
were both written afterwards, with the same sentence printed on them,
and neither author found it — which is #951's argument, arriving a
second time: a card says what its text says, and the engine works out
what that means for the cursor. So the halt is **derived** here too.

`PendingChoice.OwedInStep` is a `TurnStep` — the cursor's turn number
and step, frozen together when the prompt is queued — and
`Game.ChoicePromptBlocksTable` blocks while the game is still standing
in that step. It sits beside `GuardsStackItem` and is read through the
same one predicate, so #794's rule that "does this stop the table" has
exactly one answer is intact, and `ChoiceBlocksTable(kind)` is still
the answer for a KIND and still what
`TestEveryChoiceKindIsClassifiedAndEnumerated` checks. Both narrowings
are still one-way.

**Both narrowings are now facts about the BOARD, and that is what stops
either being a wedge.** #951's lifts when the guarded object leaves the
stack; this one lifts when the cursor leaves the step. Neither can
outlive the thing it is about, so an undo, a restore or an admin
walking the cursor by hand frees the table without anybody answering.

One door, `Game.QueueUpkeepPayUnlessForEffect`
(`server/internal/game/upkeep_pay_unless.go`), with
`effects.UpkeepPayUnless` as its card-side primitive. Stasis, Pact of
Negation and `CumulativeUpkeep` go through it. `PayUnless.Blocking`,
`Game.QueueBlockingPayUnlessForEffect` and `PendingChoice.ForceBlocks`
are **removed**: with the last declared halt derived, there is no
card-level "please block" switch left to miss, which was the whole
lesson of this amendment and of #951's.

The step is read off the cursor rather than hard-coded to the upkeep,
so a beginning-of-end-step pay-or-else is owed before the end step ends
for the same reason. The primitive is named for the family every
printed card of it belongs to.

**What is unchanged.** `pay_unless` is still the one kind classified
non-blocking, and Rhystic Study, Smothering Tithe, Esper Sentinel,
Mystic Remora's Rhystic half and Kazuul play exactly as they did — they
are asked of somebody else, about nothing the cursor cares about. A
departed payer is still settled by the departure table's `dropDecline`
column (CR 800.4f), and a bot seat answers the upkeep prompt through
the ordinary pay-unless policy.

**Amendment (2026-09-19, #1045): a prompt that blocks the table must
keep an answer the resolver would accept, or be withdrawn — and for
the choose-cards family that prune is ONE function keyed by zone.**

Every amendment above is about which prompts halt the cursor. This one
is the obligation that comes with halting it: a blocking prompt whose
answers have all become illegal holds the table forever, which is the
#544 wedge arriving from the other direction.

`PendingChoiceChooseCards` froze its candidate list when the effect
asked and re-checked each pick against the LIVE zone on submit
(`checkChooseCardsPicksLocked`), with nothing in between. So a discard
prompt whose hand was wheeled away, exiled by madness or bounced to a
library while it was open had no legal answer left and could not be
discharged — and since #1027 its RUN leg never settled either, so the
rest of the printed instruction (Syphon Mind's draw, Archon of
Cruelty's last three clauses) never ran. `pruneSacrificeChoicesLocked`
had had the equivalent prune since S17; `choose_cards` never grew one.

`Game.pruneCardSetChoicesLocked` (`server/internal/game/pending_choice.go`)
is that prune, and it is **keyed by `chooseCardsFrame.zone` rather than
by the verb**. A discard is a pick over the discarding seat's own HAND,
Thoughtseize's is one over somebody else's, Skullwinder's names a
GRAVEYARD and Genesis Wave's a LIBRARY — and "is this candidate still
there" has one answer for all of them, so it is one function and not a
`pruneDiscardChoicesLocked` beside a `pruneGraveyardChoicesLocked`. It
asks the SUBMIT path's own question
(`pickStillInPickZoneLocked`, shared with the resolver) so the offered
list and the accepted list cannot drift, and the bounds move with the
list (`setCardSetCandidates`, shared with the departure prune in
`reassignChoiceLocked`) because a floor no remaining set can reach is
the same wedge one card later — CR 701.8a's "as many as you can",
spelled arithmetically.

An emptied prompt goes out through `dropChoiceLocked`, so the departure
table's second column runs and a prompt that is one leg of a run
settles it with "nothing discarded" (#1016's `dropDefault`,
[ADR 0013](0013-replacement-effects.md) §5y item 5). A pick that is no
run's leg has no continuation to run — the table's existing answer for
the kind, unwidened here.

Three call sites, the ones `pruneSacrificeChoicesLocked` already had:
the shared exit primitive (`zone_route.go`), the battlefield-leave
resume (`mutations.go`) and the departure sweep (`leave_game.go`),
because a queued choice stops priority from passing and the state-check
loop is exactly what does NOT run while one is open.

**What is deliberately not swept.** `untap_choice` carries the same
payload (`isCardSetPickKind`) and is left alone: its continuation is
the rest of the untap step, its departure row is `dropDiscard`, so a
withdrawal would end the question by stranding the step — and CR 502.3's
determination is about the active player's own permanents during their
own untap step, where nothing has priority to move them. A pick with no
zone on its frame is left alone too: it re-checks nothing on submit, so
there is no live list for the engine to be right about.

**Amendment (2026-09-23, #1263): the exclusion above was aspirational,
not enforced.** `pruneCardSetChoicesLocked`'s loop excluded a kind with
no zone on its frame (the sentence just above, correctly) but had no
kind check at all for `untap_choice` or `entry_reveal_from_hand` — and
both of them DO set a real zone (`ZoneBattlefield`, `ZoneHand`), so the
loop's comment claiming untap_choice "sets no zone so it falls out at
the next line" was never true (`queueUntapChoiceLocked` has stamped
`Zone: ZoneBattlefield` since #826) and both kinds were examined and
could be WITHDRAWN exactly like a discard prompt. Neither kind's
departure row survives that: `untap_choice`'s `dropDiscard` runs
nothing, stranding the untap step, and `entry_reveal_from_hand`'s
paused `replacementResume` is settled only by
`dropChoicesForPlayerLocked` (a player's departure), never by the
plain withdrawal `dropChoiceLocked` performs here — so a withdrawal
strands the paused CR 614 entry too, with no path back. The fix is an
explicit kind check (`c.Kind == PendingChoiceUntapChoice ||
PendingChoiceEntryRevealFromHand`) rather than any change to the zone
logic, so this paragraph's claim is now what the code does rather than
what it was supposed to do. `TestAnUntapChoiceIsNotWithdrawnWhen-
ItsCandidateLeavesTheBattlefield` and
`TestAnEntryRevealFromHandIsNotWithdrawnWhenItsCandidateLeavesTheHand`
(`card_set_prune_test.go`) pin both shapes and fail without the kind
check.

**The reachability, stated.** No catalog card empties a hand under
another player's open discard prompt today, and the same is true of the
graveyard and library picks; this is the wedge closed before a card
reaches it, which is what #1027 asked for when it filed the hole rather
than living with it. The remaining uncovered door is a card LEAVING a
hand or library for the BATTLEFIELD, which does not go through the exit
primitive (`battlefield_put.go`'s batch and `mutations.go`'s inline
entry branch) — no prompt family can reach it today, and it is named
here rather than swept blind. *(Closed 2026-09-21 by
[#1069](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1069); see
the amendment below.)*

**Amendment (2026-09-21, #1069): the entry door is watched too, through
ONE funnel rather than three sprinklings.**

The paragraph above named the door and left it open. This is the same
prune at it, and the only judgement in the change is where the call
goes.

**1. Why the exits could not cover it.** `routeDestinationLocked`
refuses a battlefield or stack destination outright — *"not an exit,
the entry path owns these"* — so `executeZoneRouteLocked` never runs
for an arrival, and a graveyard pick whose candidate is reanimated
under the open prompt, or a hand pick whose candidate a Warp World-style
effect puts onto the battlefield, kept offering a card
`checkChooseCardsPicksLocked` would refuse. Zone-keyed as the prune is
(#1045), the question is identical to the discard's: the candidate is
not in the zone the answer is re-checked against. Only the mover
differs.

**2. One funnel.** The entry side has three landings and no shared
finisher: `executeEntryToBattlefieldLocked` (every resumable entry —
the land play, stack resolution, a search, an exile return, a
reanimation, a token), `putOntoBattlefieldFromZoneLocked` (the hand /
library batch and manifest) and `moveCardByRefLocked`'s inline branch
(the sandbox move into the battlefield **or the stack**). The last two
are the two sites that are deliberately not resumable
(`ReplacementEvent.entryResumable`), which is exactly why they cannot
be folded into the first. So the prune gets one named home,
`Game.pruneChoicesAfterArrivalLocked` in `battlefield_entry.go`, that
all three call — the shape `battlefieldExitLocked` already has on the
other side, where a landing added later is covered by the rule rather
than by a code review. (*2026-09-23, #1322:* the batch is resumable now
and lands through the same halves of the finisher — see the
[ADR 0061 amendment](0061-token-creation-and-discard-are-replaceable-events.md) —
but it still announces a whole batch after landing it, so it keeps its own
call to the prune. The sandbox move is the one unresumable landing.)

It takes no card ID: the prune re-reads every open pick against the
live board, so one call answers for a whole batch of arrivals, which is
why the batch site calls it once after phase 3 rather than once per
card. And it runs LAST in each landing — after the zone move, the ETB
event and the AsEnters hook — for `executeZoneRouteLocked`'s ordering
reason: a withdrawal settles a run leg, and that continuation is the
rest of somebody's printed instruction.

**3. No double-run.** By the refusal in item 1, not by a flag: an entry
cannot reach the exit primitive's prune block, so the new call adds a
door rather than doubling one. `entry_prune_test.go` pins that as a
fact about `routeDestinationLocked` alongside the two shapes the issue
named.

**4. What is deliberately still not swept, and it is now the only one.**
`pruneSacrificeChoicesLocked` needs no entry door at all — an entry only
ADDS permanents, so it can invalidate no option on an open sacrifice
prompt. `pruneStaleZoneChangeChoicesLocked` is a different matter: an
entry DOES empty a hand slot or take a card out of a graveyard, so a
queued prompt about moving that same card out of that zone is as stale
as it would be after an exit. It is not called here because #1069
scoped one prune at one door, and because nothing in the catalog queues
such a pair today. Named here rather than swept blind — the posture
#1045 took towards this door, one issue earlier. *(Closed 2026-09-21 by
[#1175](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1175); see
the amendment below.)*

**Amendment (2026-09-21, #1175): the entry funnel runs BOTH prunes, and
the sacrifice prune stays out for good.**

Item 4 above named the second gap and left it open. This closes it, and
the change is one line — `g.pruneStaleZoneChangeChoicesLocked()` after
the card-set prune in `Game.pruneChoicesAfterArrivalLocked`
(`server/internal/game/battlefield_entry.go`) — so what is worth
recording is the three judgements around it.

**1. It is the same question, asked at the other door.**
`pausedZoneChangeStaleLocked` asks whether the card a paused exit would
move is still in the zone that exit would leave. A commander's CR 903.9
prompt over a move out of a GRAVEYARD (CR 903.9 holds "from anywhere",
#539) whose card is reanimated while the question hangs has no answer
its own resume would accept — `dropStaleReplacementResumeLocked` is the
backstop that refuses it — and the reanimation reaches no exit, because
`routeDestinationLocked` refuses a battlefield destination outright.
Only the exile ran the prune; now both doors do.

**2. The order is the exit primitive's, and it is load-bearing.**
Card-set prune first, stale-move prune second, exactly as
`executeZoneRouteLocked:718-721` has it. The card-set drop's
continuation is a RUN LEG settling with nothing (#1016's `dropDefault`);
the stale-move prune's is a ROUTE continuation
(`abandonZoneRouteLocked`, #865) that can start the next move or queue
the next leg's prompt. Reversing them would let a route's next leg be
queued in front of a withdrawal the same arrival already owes.

**3. No double-run, and nothing new to prove it with.** The structural
fact item 3 above rests on covers the second prune unchanged: an
arrival never reaches `executeZoneRouteLocked`'s block, so the funnel
adds a door rather than doubling one. The second prune does bring one
hazard the card-set prune did not — a route continuation is somebody's
printed instruction, so running the funnel twice over one arrival would
run it twice — and the prune is idempotent because it re-reads the live
queue and a withdrawn prompt is no longer in it.
`TestTheArrivalFunnelIsIdempotent` pins that beside the issue's own
shape in `entry_prune_test.go`.

**4. `pruneSacrificeChoicesLocked` is now permanently out, not
deferred.** Its question is "is this candidate still on the battlefield
under its chooser's control", and an entry only ADDS permanents. There
is no board an arrival can produce that invalidates an option on an
open sacrifice prompt, so this is a closed argument rather than an
unscoped one. The entry funnel holds two prunes and will not grow a
third for this reason.

**Reachability, restated.** Still none in the catalog: the prompts in
question block the table (#791's gate), so one resolving effect would
have to hold the prompt open and move the same card onto the
battlefield, and no card does. This is the wedge closed before a card
reaches it, which is the posture #1027, #1045 and #1069 each took in
turn.

**Amendment (2026-09-22, #1198): a card-set pick can be welded to the
ENTRY pipeline, and the three tables say so one row each.**

§6's four tables — the gate (`choiceGateDecisions`), the departure rule
(`choiceDepartureDecisions`), the enumerator's `choiceMoves` and the
heuristic's `valueOfChoice` — are the whole contract for a new
`PendingChoiceKind`, and two tests refuse a kind that skips one
(`TestEveryChoiceKindIsClassifiedAndEnumerated`,
`TestEveryChoiceKindHasAReassignmentDecision`).
`entry_reveal_from_hand` (#1198, "as this land enters, you may reveal an
Island or Swamp card from your hand"; see
[ADR 0013](0013-replacement-effects.md) §5z for the pipeline half) is the
third kind to carry the `choose_cards` PAYLOAD after #826's
`untap_choice`, and the first to carry it welded to a PAUSED EVENT rather
than to a resolving effect. Its four rows, and the one thing each says
that is not obvious:

- **Gate: blocks.** For `entry_pay_life`'s reason rather than
  `choose_cards`': the permanent is mid-entry and the whole CR 614
  pipeline is suspended on the answer. A table that could walk past the
  question would be answering it by entering.
- **Departure: `{}` — dropped, and the drop does nothing**, which is
  `entry_pay_life`'s row verbatim and for its argument. The prompt is
  never reassigned (CR 800.4f: the reveal is the "unless" of the departed
  player's own entering permanent, which CR 800.4a takes out of the game
  in the same breath), and the paused event is not left dangling by the
  empty second column: the frame rides `PendingChoice.replacementResume`,
  so `dropChoicesForPlayerLocked` hands it to
  `finishDroppedReplacementLocked` → `abandonZoneRouteLocked` with no row
  of its own. **`dropDefault` would be wrong here**, not merely
  unnecessary: this kind's continuation is not a card's next sentence, it
  is a replacement event, and settling it twice — once through the drop
  action and once through the frame — is the double-resume the second
  column exists to avoid.
- **Enumerated: on the `choose_cards` / `untap_choice` branch**, a third
  verb and nothing else. The bounds ride on the choice, every set is run
  past `ChooseCardsPickLegalLocked`, and the floor is zero — so "reveal
  nothing" is the `AlwaysLegal` answer and no board can leave this seat
  with an empty move list.
- **Scored: its own branch, not `choose_cards`'.** The sign is the whole
  decision (#798). A card named to a `choose_cards` prompt over a bot's
  own hand is a card GIVEN UP, so that branch scores an answer by what it
  keeps; a card named here is revealed and stays in hand, so naming one
  costs nothing and buys an untapped land. #1028's fuel pricer is not the
  hint either — fuel prices a card SPENT to a cost, and a reveal spends
  nothing. Scored on the `untap_choice` side of the ledger: more is
  better, and the count settles itself.

**And one sweep it is deliberately outside.** `pruneCardSetChoicesLocked`
is not extended to it, for `untap_choice`'s reason plus one of its own:
the prune DROPS a prompt whose candidates have all gone, and dropping
this one would end the question by stranding the entry it is pausing.
It cannot need the prune either — its floor is zero, so a candidate list
emptied under it still has an answer the resolver accepts, which is the
property `untap_choice` does not have and `choose_cards` with a floor of
one does not have. The `pickStillInPickZoneLocked` re-check on submit
still refuses a stale pick, so the worst an un-pruned list can do is cost
one rejected click.

**Amendment (2026-09-22, #1214): three resolution-time picks over
cards and permanents — one payload, three kinds, one continuation.**

The #568 amendment above built the one resolution-time question the
queue could not ask ("choose one of the following", addressed to any
seat) and said a pile split needs no kind at all because it is two
chained prompts. Three rows of `docs/engine-seams.md` were the
questions it did not build, and all three are about CARDS or
PERMANENTS rather than about a labelled branch:

- **an opponent picks from a set you revealed** — Gifts Ungiven
  ("target opponent chooses two of those cards"), Intuition;
- **a non-owner chooses among another player's permanents** — Tragic
  Arrogance ("for each player, you choose from among the permanents
  that player controls …"). `PendingChoiceSacrifice` is queued
  `{Chooser: playerID, FromPlayer: playerID}` at its one queue site and
  `ResolveSacrificeChoice` refuses anything else, so nothing could
  point a permanent pick across the table;
- **"choose N of your own permanents", at resolution** — Scapeshift's
  "sacrifice any number of lands". The existing sacrifice prompt is a
  fixed count asked one permanent at a time, which cannot say "any
  number" and cannot be read for a total.

**One payload.** All three carry the choose-cards payload —
`ChooseCards`, `ChooseMin`, `ChooseMax` and a `chooseCardsFrame` — so
`isCardSetPickKind` (`chained_choice.go`) admits them, beside
`choose_cards`, CR 502.3's `untap_choice` and #1198's
`entry_reveal_from_hand` (the amendment above), and every rule about the
payload is written once and asks it:
`checkChooseCardsPicksLocked` (bounds, candidates, duplicates, the
live-zone re-check and #1017's set-level `Validate`),
`ChooseCardsPickLegalLocked` (the enumerator's window onto that),
`setCardSetCandidates`, and both prunes that keep an open prompt honest
as the board moves under it. `reassignChoiceLocked`'s candidate prune
became one `isCardSetPickKind` arm in the same pass, replacing the
`PendingChoiceChooseCards` case that was the same five lines.

**Three kinds, and each difference is observable.** The test surveil
had to pass to be a kind rather than a flag on scry:

| | `reveal_pick` | `their_permanents` | `own_permanents` |
|---|---|---|---|
| the pool | cards the card REVEALED (CR 701.20) | another seat's battlefield | the chooser's own battlefield |
| what a non-chooser sees | the cards and the bounds — the reveal is public | the cards and the bounds — the battlefield is public | the same |
| how a bot should read it | rank the pool by what matters (`Options.OrderTargets`) | the same | rank by what the seat would miss LEAST (`Options.OrderCostFuel`) |
| CR 800.4g | reassigned | reassigned | never — it is the chooser's own material |

A `choose_cards` over the same IDs would be wrong on every row.
`filterPendingChoices` withholds even a choose_cards prompt's BOUNDS
from a non-chooser, because its pool is usually a hand; running a
public reveal through that would hide from the table a fact it watched
happen. And a policy handed one ordering hook where the other was
wanted keeps its worst land and hands an opponent their bomb — the
mistake `Options.OrderCostFuel` exists as a separate type to prevent.

**One continuation: they are RUNS.** Every one of the three is queued
through `runPromptsLocked` (`prompt_run.go`), the shape #1019 built for
a prompted sacrifice and #1027 generalised for a prompted discard. That
is not reuse for its own sake — Tragic Arrogance is ONE printed
instruction asked as one prompt per player, and its "then each player
sacrifices all other nonland permanents they control" may not run until
the last of them is answered. It also settles the drop path for free:
each prompt is one LEG, so `dropDefault` reaches `settleRunLegLocked`
through `PendingChoice.promptRun` — the branch
`defaultDroppedChoiceLocked` already takes on the RUN LINK rather than
on the kind (#1027) — and none of the three needed a case of its own
anywhere. See ADR 0060's amendment of the same date for the departure
rows.

**Which kind a leg is queued as is decided in one line.**
`PermanentsPickedThenForEffect` asks whether the leg's subject IS the
chooser: yes is an `own_permanents`, anything else is a
`their_permanents`. Tragic Arrogance walks every player including its
own controller, so one printed sentence is both kinds and the card says
nothing about either.

**A pile split still needs no kind of its own — but its first leg now
has a name.** `QueuePileSplitForEffect` is unchanged in shape: a pick
addressed to the splitter, chained to an `option_pick` addressed back
to the controller. What changed is that the first leg is a
`reveal_pick` rather than a `choose_cards`, which is what it always
was — "an opponent picks from a set you revealed" is the sentence Fact
or Fiction prints. Two things follow. The prompt stops hiding its
bounds from the rest of the table. And it fixes a hole the #1006
amendment to ADR 0060 left one leg upstream of where it was looking: a
splitter who left AFTER the question went up hit a `dropDefault` that
found no run and no option frame on the prompt, ran nothing, and Fact
or Fiction put neither pile anywhere — the exact failure that amendment
names. It is now one leg of a run, so the withdrawal settles with
nothing separated and the controller still takes a pile.

**All three block the table**, for `option_pick`'s reason stated once:
the effect that asked is paused mid-resolution and its continuation is
the rest of the card. `their_permanents` is the one worth naming out
loud, because it LOOKS like `pay_unless`'s background question — a
prompt addressed to somebody who is not the permanents' controller —
and is not: nothing about a Rhystic tax is holding a spell
half-resolved.

**Narrated.** `EventCardsChosen` → `LogChooseCards`, for the reason
#1023 gave a chosen colour a line: a decision the table watched
somebody make is a decision the history has to carry. Without it the
only trace of Tragic Arrogance is a column of sacrifices with nobody's
name on them.

**Out of scope, stated.** The seam rows keep two cards each that this
does not reach, and both are a second axis rather than a deeper version
of this one. The Legend of Yangchen's chapter I ("starting with you,
each player chooses up to one permanent … from among permanents your
opponents control") needs several players choosing from ONE SHARED
POOL, with each answer narrowing the next — a run over one pool, not a
run over one pool each. Gluntch, the Bestower needs two target players
before the pick and a different effect per seat. Neither is blocked on a
kind any more; both are card work plus a sequencing shape, and they
stay on their rows.

**Amendment (2026-09-23, #1225): a withdrawn mid-card `choose_cards`
runs the rest of the card. The #1045 amendment's "unwidened here" is
repealed.**

That amendment closed the wedge and then declined the second half of
it, in one sentence: *"A pick that is no run's leg has no continuation
to run — the table's existing answer for the kind, unwidened here."*
It is wrong, and it is wrong about the commonest shape this kind has.

**The shape it missed.** Three things can be waiting on a dropped
prompt, and `defaultDroppedChoiceLocked` knew two of them. A RUN LINK
(#1019 / #1027): the leg settles with nothing moved. An OPTION FRAME
(#1006): the continuation runs with `NoChoiceIndex`. The third is a
`choose_cards` carrying **only a `chooseCardsFrame`** — a pick chained
from inside a resolution that is paused waiting for it, which is what
`effects.SacrificeChoice` is and what its own doc comment says it is
for: *"a sacrifice in the MIDDLE of an effect — Torment of Hailfire
repeats X times, and the next repetition must not be asked until this
one has finished."* `optionPickResume` is nil there,
`runWithNoChoice` returns nil on a nil receiver, and the rest of the
card never ran. **Torment of Hailfire stopped at the victim whose
board emptied and never asked the opponents after them** — the same
sentence the #1006 amendment to ADR 0060 wrote about the `option_pick`
half of the very same card, one prompt further in. The floor of one is
what makes it reachable: a zero-floor pick keeps "choose nothing" when
its list shrinks, so `pruneCardSetChoicesLocked` leaves it standing.

**The decision: the drop runs the frame with NOTHING PICKED.** One
case added to `defaultDroppedChoiceLocked`
(`server/internal/game/option_pick.go`), performed by
`chooseCardsFrame.runWithNoChoice`
(`server/internal/game/chained_choice.go`) — the sibling of
`optionPickFrame.runWithNoChoice`, and the one place "nobody chose" is
handed to this payload.

Three reasons it is the right answer rather than a flag on the frame
or a run wrapped around each caller, which were the two shapes #1225
proposed:

1. **The empty pick is not a new outcome.** `QueueChooseCardsForEffect`
   deliberately has no empty-candidate short-circuit, so every caller
   of it already wrote the "there was nothing to ask" path by hand —
   and every one of them runs *the same closure* with *the same empty
   answer*. `SacrificeChoice` runs `Then` immediately for a player with
   no candidate permanent; `PutFromLibraryOntoBattlefield` and
   `TakeFromLibraryToHand` call `finish(g, nil)`; Ward's sacrifice
   counters the spell. The drop was the one path into those
   continuations that did not reach them. What this fixes is a
   card that behaves one way when the board was empty as the question
   was asked and another way when it emptied while the question was
   open.
2. **`Validate` is untouched, because it was already excluded.**
   `ChooseCardsPrompt.Validate` is documented as never being called for
   an empty pick, since a floor of zero has to keep "choose nothing" as
   an answer the engine cannot refuse. So there is no set rule for the
   drop to be judged against and none to bypass.
3. **#1027's rule survives intact: branch on the PROMPT, not on the
   kind.** The run link is asked FIRST and exclusively, because a run
   leg's own frame settles the run from inside itself (the discard's
   `discardCardsLocked` → `opts.then`) and running both would settle
   the leg twice and pay the run out before its other legs were
   answered — `TestARunWhoseLegWasWithdrawnStillCompletes` is the guard.
   A prompt with neither a run nor a `then` still runs nothing, which
   is Thoughtseize's pick and is unchanged.

**What changes for the catalog, swept site by site.** Twenty catalog
`choose_cards` sites plus the engine's two. Seven move; thirteen
cannot.

- **Two cards stopped halfway and now finish** — Torment of Hailfire
  (`effects.SacrificeChoice`; the next repetition and every later
  victim) and Gluntch, the Bestower (the second and third players were
  never chosen). Both are the #544 shape.
- **Four clauses now run their remainder with nothing chosen**, which
  is byte for byte the call their own empty-candidate path already
  makes: `PutFromHandOntoBattlefield` (`then` with `Entered` unset),
  `PutFromLibraryOntoBattlefield` and `TakeFromLibraryToHand` (both
  `finish(g, nil)`), and Invasion of Tarkir (its reflexive damage
  trigger with X = 0 revealed Dragons).
- **One card was neither paid nor countered and now resolves the
  "unless"** — Vein Ripper's `ward—sacrifice a creature`. A payer
  whose last creature leaves mid-pick has not paid, so the spell is
  countered, which is exactly the branch the queue-time guard takes
  for a payer who had no creature to begin with. This is the one
  behaviour change that is not simply "more of the card happens", and
  it is CR 800.4f's shape reached through `dropDefault` rather than
  `dropDecline`.
- **Thirteen sites are unchanged by construction** — the ones whose
  continuation opens `if len(picked) == 0 { return nil }` or loops
  over the picks and so does nothing with none, and Sylvan Library,
  whose `sylvanLibrarySettle(…, nil)` returns immediately. Sylvan
  Library is pinned by a test of its own rather than by argument,
  because its continuation is the one that recurses into further
  prompts.

**The prompt kinds that are NOT touched, and why the table is still
what decides.** `untap_choice` and `entry_reveal_from_hand` carry the
same payload (`isCardSetPickKind`) and both are `dropDiscard` rows, so
neither reaches this action at all; `entry_reveal_from_hand` also
carries no `then`. The three #1214 picks are always run legs and take
case 1 unchanged. The four-table contract this section states for a
new kind is unchanged: what moved is the meaning of ONE cell of one
table, for one kind.

**Where the departure table is printed**, and the two gates in front
of the action, are ADR 0060's: the row is still
`{reassign: true, onDrop: dropDefault}` and `dropDefault`'s definition
there — *"the frame runs with the outcome the kind reserves for
'nobody chose'"* — needed no change, because an empty pick is that
outcome for this payload. Only the claim that this kind had no such
outcome, made here, did.

## Out of scope (explicit deferrals)

- **Treasure's sac-for-mana** is inert until S21 ships sacrifice
  as a cost component; Smothering Tithe makes countable artifacts.
- **"You may draw"** on Rhystic Study is treated as "draw".

- **Modal triggers** ("draw a card or gain 3 life") still need a
  `ModePrompt` slot; nothing in the catalog needs one yet.
- ~~**Trigger ordering UI** for one player's simultaneous triggers
  (CR 603.3b)~~ — shipped in sub-PR 8: `PendingChoiceTriggerOrder`
  holds the APNAP drain until each seat with ≥2 *differing*
  triggers has ordered them (identical ones auto-order, and since
  #1511 so does a batch that is all prowess — see that amendment; the
  submitted order is resolution order). A trigger arriving while a
  prompt is open re-asks with the full list.
- ~~**Cast / combat-damage triggers**~~ — shipped in sub-PRs 6 and 7
  in the `NewTriggeredItem` shape.
- ~~**Stack overlay art for ability items.**~~ Shipped in sub-PR 8:
  the overlay resolves `source_card_id` against battlefield / exile
  / graveyards for art, title and target names.
- **`PlayCard`** (hand → battlefield sandbox drop, pre-S13.1) still
  emits no ETB at all. Left alone; `move_card` is the verb the
  client uses.

## Consequences

- The test harness `passPriorityAroundTable` now waits for
  `Game.Stack`, `StackMeta`, and `PendingTriggers` to all empty, so
  "cast creature → ETB trigger → effect" settles in one call. Tests
  for optional triggers answer the prompt and call it again.
- The Playwright S19 suite resolves triggers through each player's
  own "next" button (`resolveStack`) rather than an admin
  `pass_priority`, so a browser that auto-passed can't be
  double-passed into a step advance.
- Engine coverage: effect runs on resolution not build; fizzle when
  every target is gone; effect error surfaces and clears the item;
  effect survives `Clone`/`RestoreFrom`; sandbox move drains onto
  the stack. Catalog coverage: countering Mulldrifter's trigger
  draws nothing; Reclamation Sage fizzles when its target leaves in
  response and does not re-pick; two dies triggers from one wrath
  stack APNAP.

## Addendum (2026-09-17): trigger doubling (CR 603.2d) — Accepted

**Status:** Accepted · 2026-09-17 · unscheduled (card-coverage audit, wave 2) · tracked on [#752](https://github.com/krakenhavoc/cmd_and_ctrl/issues/752)
**Numbering:** an addendum to ADR 0018, not a new ADR. No new number, and the ADR range line in AGENTS.md §3 is unchanged.
**Builds on:** this ADR (the harvester, `Build` / `Effect`, the APNAP drain, the `trigger_order` prompt),
[ADR 0049](0049-card-engine-seam-review.md) D2 (`OncePerBatch`, #587) and D4 (`CardDef`, #622),
the reflexive-trigger seam (#636, `game/reflexive.go`), the delayed-trigger queue
([ADR 0026](0026-delayed-triggers.md), `game/delayed.go`), the simultaneous-exit batch
(S23, `game/simultaneous.go`), [ADR 0041](0041-game-persistence.md) (snapshots) and
[ADR 0055](0055-loop-breaker.md) (the loop breaker's per-key count).
**Owner decisions:** three questions, answered 2026-09-17 and recorded at the
[end of this addendum](#decided-2026-09-17). The decisions below apply them:
extra instances are tagged on the stack overlay and on their prompt with no log
line (question 1), the first wave includes Elesh Norn, Mother of Machines and
Veyran, Voice of Duality with declared caveats (question 2), and each doubled
optional trigger gets its own yes/no prompt (question 3).

Line references are to `origin/develop` at `684f2786`.

### Context

#### What the rules ask for

CR 603.2d (Aug 7 2026 edition):

> An ability may state that a triggered ability triggers additional times. In
> this case, rather than simply determining that such an ability has triggered,
> determine how many times it should trigger, then that ability triggers that
> many times. An effect that states that an ability triggers additional times
> doesn't invoke itself repeatedly and doesn't apply to other effects that
> affect how many times an ability triggers. An effect that states a triggered
> ability of an object triggers additional times refers only to triggered
> abilities that object has, not to any delayed or reflexive triggered
> abilities (see rule 603.7 and rule 603.12) that may be created by abilities
> the object has.

The printed rulings on the cards this addendum is for (Scryfall, checked
2026-09-17) settle the rest. They say the same things across Panharmonicon,
Teysa Karlov, Yarok, Elesh Norn, Naban, Katara, Starfield Vocalist, Annie Joins
Up, Roaming Throne, Felix Five-Boots and Drivnod:

- **Not a copy.** "Doesn't copy the triggered ability; it just causes the
  ability to trigger twice." Modes and targets are chosen separately for each
  instance as it goes on the stack, and choices made on resolution are made
  separately too.
- **Doublers add up, they don't multiply.** Two Panharmonicons give three
  instances, three give four.
- **Simultaneity counts in both directions.** A permanent entering at the same
  time as Panharmonicon, Yarok, Elesh Norn, Naban, Katara or Starfield Vocalist
  (including the doubler itself) is doubled. A creature dying at the same time
  as Teysa or Drivnod (including Teysa herself) is doubled. If Teysa and a
  second Teysa meet through the legend rule, "an ability that triggers when it
  dies due to the legend rule triggers three times".
- **The cause is the event, not the words.** Teysa doubles "whenever an
  artifact is put into a graveyard from the battlefield" when an artifact
  creature dies, and doubles a "leaves the battlefield" trigger when the
  creature left by dying. She does not double "whenever you sacrifice a
  creature". Isshin doubles only conditions "directly related to attacking",
  not "whenever this creature becomes tapped". Felix doubles only triggers on
  the damage itself, not Ajani's Pridemate gaining from lifelink.
- **The creature is judged as it exists on the battlefield**, continuous
  effects included (a land that has become a creature dies and Teysa counts it;
  a Runeclaw Bear under Arcane Adaptation naming Wizard is a Wizard for Naban).
- **The count is fixed when the ability triggers.** Making a creature legendary
  after its ability triggered adds no instance under Annie Joins Up, and making
  it nonlegendary afterwards removes none.
- **Controller of the entering or dying object doesn't matter** unless the
  doubler prints "you control" (Naban's "a Wizard you control", Felix's "a
  creature you control"). The controller of the ability's *source* does: "a
  triggered ability of a permanent you control". Cloud, Midgar Mercenary is the
  one card here with no "you control" at all: "a triggered ability of Cloud or
  an Equipment attached to it".
- **Replacement effects and "as enters" choices are unaffected.** They are not
  triggers, and the engine already doesn't route them through the harvester.
- **Linked abilities:** extra instances are linked to the same second ability.
  "The exiled card" means every card any instance exiled.

#### Where triggers are put on the queue today

Every catalog trigger that comes from an event passes through
`dispatchTriggerLocked` (`game/triggers.go:364`). Its callers on develop:

| caller | line | what it harvests |
|---|---|---|
| `harvestFromZone` | `triggers.go:310` | battlefield watchers |
| `harvestCastFromStack` | `triggers.go:345` | `FromStack` abilities of the spell just cast (cascade) |
| `harvestLTB` | `triggers.go:449` | the leaving card's own triggers, off its LKI snapshot |
| `harvestSimultaneousExitLocked` | `simultaneous.go:147` | watchers that left earlier in the same batch |
| `QueueReflexiveTriggerForEffect` | `reflexive.go:175` | **CR 603.12 reflexive triggers** (shipped by #636 after the issue was written) |

Three more places build a triggered item without dispatch:

| site | line | what |
|---|---|---|
| `fireDelayedTriggersLocked` | `delayed.go:161`, queue at `:181` | CR 603.7 delayed triggers |
| `queueAltCostEntryTriggerLocked` | `alternative_cost.go:549`, queue at `:563` | evoke's "when this permanent enters, if its evoke cost was paid, its controller sacrifices it" (CR 702.74a). Warp's exile at `:557` is a delayed trigger. |
| `AnnounceTrigger` | `mutations.go:2093` | the sandbox's manual "trigger" button |

Three facts on develop shape the design:

1. **The issue's proposed check point would double reflexive triggers.**
   #752 suggests counting doublers inside `dispatchTriggerLocked`. Since #636,
   reflexive triggers enter that same function (`reflexive.go:175`), and CR
   603.2d excludes them by name.
2. **`OncePerBatch` would swallow the extra instance.** `dispatchTriggerLocked`
   returns at `triggers.go:365` when `triggerInFlightLocked` (`:554`) finds an
   item with the same source and key on `PendingTriggers`, on `StackMeta`, or
   waiting on a trigger or target prompt. A second dispatch for the same match
   finds the first one and returns.
3. **`harvestLTB` deletes the leaving card's LKI before the simultaneous-exit
   pass runs** (`defer delete` at `triggers.go:405`, and `OnEvent` calls
   `harvestSimultaneousExitLocked` after `harvestLTB`). A dying doubler's
   characteristics must therefore be read before that delete, or out of the
   batch copy (`simultaneous.go:78`).

#### What already exists

- `CardDef` (`game/carddef.go:31`): one precomputed catalog lookup, with a
  per-slot hook variable kept only where a test needs to stub it.
- `TriggeredAbility.Key` and `OncePerBatch` (`triggers.go:177-194`).
- The trigger and target prompts carry a value copy of the source, the LKI and
  the `Build` closure in `triggerResumeFrame` / `pickTargetFrame`
  (`pending_choice.go:604-625`), so each instance can own its own prompt.
- `seatNeedsTriggerOrder` (`mutations.go:2928`) already treats items with the
  same source and label as identical and skips the `trigger_order` prompt for
  them.
- Entries that go through `putOntoBattlefieldFromZoneLocked` move every card
  before announcing any of them (`battlefield_put.go:330-346`), so a watcher or
  doubler in the batch sees the others. Token creation announces each token as
  it is created (`effect_api.go:2098-2109`).
- The one catalog card that already declares the missing seam as a caveat:
  Cloud, Midgar Mercenary (`cards/effects/cloud_midgar_mercenary.go:34`).

### Decision 1 — A doubler is a catalog slot: `Spec.TriggerDoublers`

```go
// game/trigger_doubling.go
type TriggerDoubler struct {
	// Label names the static for tests and the attribution ("Panharmonicon").
	Label string
	// Applies reports whether this doubler adds one instance of the
	// ability described by q. Runs under g.mu in write mode and must not
	// take public locks or mutate state.
	Applies func(g *Game, q TriggerDoublingQuery) bool
}

type TriggerDoublingQuery struct {
	Event Event

	// The doubler as it is now (battlefield) or as it last existed (a
	// batch copy or the LKI of the card whose exit is being reported).
	Doubler    Card
	DoublerLKI Characteristic

	// The ability's source and the characteristics the harvester judged
	// it on (Effective() on the battlefield, the LKI snapshot for an LTB,
	// the batch copy for a simultaneous exit).
	Source    Card
	SourceLKI Characteristic
	// FromSpell is true for a FromStack ability of a spell being cast.
	// Every doubler except Echoes of Eternity prints "of a permanent".
	FromSpell bool
	Ability   *TriggeredAbility

	// Subject is the object the event is about (the entering, dying or
	// attacking permanent, the spell cast) with its characteristics as
	// the rulings require: live on the battlefield for an entry or an
	// attack, the LKI snapshot for an exit. Zero when the event has none.
	Subject    uuid.UUID
	SubjectLKI Characteristic
	HasSubject bool
}
```

`effects.Spec` gains `TriggerDoublers []game.TriggerDoubler`, `CardDef` gains
the same slot, and `buildDef` copies it. A `CatalogTriggerDoublers` hook
variable is added beside `CatalogTriggers`, because the engine tests in
`internal/game` stub doublers without importing the catalog (#622's rule for
when a per-slot variable earns its place).

**Why a predicate over the event rather than a declared cause enum.** The
wording varies more than an enum would (Gandalf's "a legendary permanent or an
artifact entering or leaving", Veyran's "you casting or copying", Felix's "a
creature you control dealing combat damage to a player", Cloud's "as long as
Cloud is equipped"). The common shapes become one-line helpers (Decision 6),
and a card with unusual wording writes its own predicate the way cards already
write `AppliesTo`.

**Why not a `StaticAbility`.** Statics are layer effects that the layer
engine recomputes. A doubler changes nothing in any layer. It is read at one
moment, by the harvester, and putting it into the layer system would make every
recompute carry a slot that is never applied.

**The count is `1 + the number of doublers whose Applies returns true`.** Each
doubler adds exactly one, so two Panharmonicons give three instances. A doubler
is never asked about its own contribution or another doubler's (the second
sentence of CR 603.2d), because doublers are not triggers and the scan never
looks at instances it has already added.

### Decision 2 — The count is taken at the harvest, not inside `dispatchTriggerLocked`

`dispatchTriggerLocked` is split in two:

```go
// harvestMatchLocked: the four harvest callers use this instead of
// dispatchTriggerLocked.
func (g *Game) harvestMatchLocked(pass *harvestPass, source Card, lki Characteristic, t TriggeredAbility, fromSpell bool) {
	if t.OncePerBatch && g.triggerInFlightLocked(source.InstanceID, t.Key) {
		return
	}
	extra := g.triggerDoublersLocked(pass, source, lki, t, fromSpell) // []doublerRef
	g.dispatchTriggerInstanceLocked(pass.ev, source, lki, t, doublerRef{})
	for _, d := range extra {
		g.dispatchTriggerInstanceLocked(pass.ev, source, lki, t, d)
	}
}

// dispatchTriggerLocked keeps its signature and meaning for the
// reflexive path: one instance, with the in-flight check.
func (g *Game) dispatchTriggerLocked(ev Event, source Card, lki Characteristic, t TriggeredAbility) {
	if t.OncePerBatch && g.triggerInFlightLocked(source.InstanceID, t.Key) {
		return
	}
	g.dispatchTriggerInstanceLocked(ev, source, lki, t, doublerRef{})
}
```

`dispatchTriggerInstanceLocked` is today's body from `triggers.go:368`
onward: the CR 603.3d empty-target drop, the optional prompt, the target pick,
`Build`, and the queue.

What this decides:

- **Reflexive triggers are never doubled**, because `QueueReflexiveTriggerForEffect`
  keeps calling `dispatchTriggerLocked`. **Delayed triggers are never doubled**,
  because `fireDelayedTriggersLocked` never reaches dispatch. Both are the CR
  603.2d exclusion, enforced by where the count is taken rather than by a flag
  someone could forget to set.
- **Each instance is a complete, independent dispatch.** It gets its own CR
  603.3d legal-target check, its own "you may" prompt, its own target pick, its
  own `Build` call and its own `EventTrigger`. That is the "choices are made
  separately for each instance" ruling, and it needs no new prompt kind.
- **The instances are queued one after another**, the first before the extras,
  so on `PendingTriggers` they are adjacent and in doubler order.
- **The manual `AnnounceTrigger` button is not doubled.** A player announcing
  by hand announces as many as they mean to.

#### Evoke's sacrifice trigger goes through the same count

CR 702.74a makes evoke's sacrifice "a triggered ability that functions on the
battlefield", triggered by the permanent entering, so Panharmonicon, Yarok,
Katara (for an Ally) and the like double it. `queueAltCostEntryTriggerLocked`
builds a `harvestPass` for a synthetic `EventETB` naming the card (the real
`EventETB` was emitted just before, at `mutations.go:1709` / `entry_choice.go:367`),
asks the same `triggerDoublersLocked`, and queues one sacrifice item per
instance. Nothing a player sees changes except the stack: the second sacrifice
resolves after the permanent is gone, and the existing `Effect` already does
nothing then (`alternative_cost.go:569-578`). Warp's exile stays a delayed
trigger and is not doubled (CR 702.185a, CR 603.2d).

#### Saga chapters are ordinary triggered abilities

Chapter abilities are catalog `TriggeredAbility` values watching
`EventSagaChapter` (`cards/effects/saga.go:51`). They reach the harvester like
any other ability, so a doubler keyed on the *source* (Katara, Annie Joins Up,
Roaming Throne, Echoes of Eternity) doubles them when the Saga fits its filter.
Doublers keyed on an *event* (entering, dying, attacking) do not, because the
chapter's event is a lore counter. For chapter I on a Saga's entry, this is the
engine's reading of CR 714.3a (the counter is put on "as it enters", a
replacement) and CR 714.2b (the trigger condition is the counter). Sub-PR 2
checks it against the rulings before any entry doubler ships. If a ruling says
otherwise, the entry helper gains the `EventSagaChapter` case whose counter came
from the entry, and nothing else moves.

### Decision 3 — Which doublers are asked, and as what

`OnEvent` builds a `harvestPass` once per event and passes it to all four
harvest functions:

```go
type harvestPass struct {
	ev         Event
	subject    uuid.UUID
	subjectLKI Characteristic
	hasSubject bool
	doublers   []doublerCandidate // nil until the first match needs them
	scanned    bool
}
```

1. **The subject is resolved eagerly, at the top of `OnEvent`.** For
   `EventLTB` it is the LKI snapshot of `ev.CardID`, read *before*
   `harvestLTB`'s delete (fact 3 in Context). For `EventETB`, `EventZoneMove`
   into the battlefield, `EventTokenCreated` and `EventAttack` it is the live
   battlefield card's `Effective()`. For `EventCast` it is the spell on the
   stack. For a combat `EventDealDamage` it is `ev.Source`, with `ev.Actor`
   already carrying its controller at emit time (`events.go`, the `Combat`
   field comment). This costs one map lookup or one battlefield find per event,
   and only for those kinds.
2. **Doubler candidates are collected lazily, on the first match of the
   event.** Most events match nothing, and most games have no doubler, so the
   scan runs only when a trigger is about to be dispatched, and then at most
   once per event. The candidates are:
   - every battlefield permanent whose `CatalogAbilityKey` has a non-empty
     `TriggerDoublers` slot (a permanent under a CR 613.1f "loses all
     abilities" effect has no key, so it is skipped: the ability-removal case
     from the issue);
   - every card in the open simultaneous-exit batch (`g.simultaneousExit`) that
     is no longer on the battlefield, read through `CatalogAbilityKey(copy)` on
     the batch copy, exactly as `harvestSimultaneousExitLocked` reads watchers;
   - for `EventLTB`, the leaving card itself when it is not in a batch: its
     `CatalogKey` from its destination zone, skipped when the captured LKI has
     `AbilitiesRemoved`, exactly as `harvestLTB` does.

   Candidates are de-duplicated by instance ID, and a batch copy or LKI wins
   over a destination-zone card.
3. **The candidate list is safe to reuse across the event's matches.** Nothing
   between the first match and the last one can change the battlefield:
   `Build` must not mutate state (Decision 2 of this ADR), and queueing a
   prompt or an item moves no card. A nested event (a listener emitting
   during dispatch) builds its own pass.

What this gives, against the rulings:

- **A doubler entering with the triggering permanents counts** whenever the
  engine's entry path puts them on the battlefield before announcing them
  (`battlefield_put.go:330-346`). Panharmonicon's own entry counts for an
  "whenever an artifact enters" watcher, because Panharmonicon is on the
  battlefield when its own `EventETB` fires. Where an entry path announces one
  permanent at a time (token creation), a doubler that enters later in the
  same effect misses the earlier ones. That is the gap watchers already have
  on that path, and this design adds no new one.
- **A doubler leaving in the same event counts**: Teysa dying alone (her own
  LKI), Teysa in a wipe (her batch copy, still visible when a later death in
  the batch is announced), and the legend-rule case (the dying Teysa's LKI
  plus the surviving one on the battlefield, giving three).
- **The count is taken once, when the ability triggers.** A doubler that
  leaves while an instance's prompt is open does not remove that instance, and
  a doubler that arrives later adds none (the Annie Joins Up ruling).

### Decision 4 — `OncePerBatch`: checked once per match, never per instance

The in-flight check runs once, before any instance is dispatched
(`harvestMatchLocked` above). Every instance of the match, including the extras,
is then queued without consulting it. The next event of the same batch finds
the first instance in flight and is declined as it is today, extras included.
Isshin plus Adeline with three attackers is exactly two Adeline triggers.

The check itself is unchanged, so its purpose for a later batch is unchanged.
Whatever replaces it later (see the follow-up below about a batch that
arrives while the previous batch's item is still on the stack) also runs once per
match, because it sits in the same place.

**Declared limitation: the doubler count comes from the batch's first event.**
The engine has no batch identity on events: "whenever one or more" is modelled
as "the first event of the batch, while nothing from it is in flight". If a
doubler's filter only matches a later member of the batch, the extra instance
is missed. Example: Naban plus a "whenever one or more creatures you control
enter" ability, when a non-Wizard's `EventETB` is announced before a Wizard's
in the same batch. The error only ever undercounts, so it is never stronger than
printed. Every doubler whose filter is the same for every member of a batch
(Isshin and Wulfgar for attacks, Teysa for "one or more creatures die",
Panharmonicon for an all-creature batch) is exact. The implementing PR adds a
test that pins this limitation, so fixing it later shows up as a deliberate test
change.

**Amendment (2026-09-17, #829): the check this decision keeps is gone; the
decision is not.** The follow-up named two paragraphs up has shipped. The
engine now HAS a batch identity on events — `Event.Batch`, advanced at one
boundary (a stack item beginning to resolve, or the turn cursor entering a new
step; see `server/internal/game/event_batch.go` and ADR 0049's #829
amendment) — and `dispatchTriggerLocked`'s in-flight scan is replaced by
`oncePerBatchAllowsLocked(ev.Batch, source, key)`. Two consequences for this
ADR, neither of which changes what it decides:

- The historical code samples above spell the guard `g.triggerInFlightLocked(...)`;
  `harvestMatchLocked` uses the batch guard instead, in exactly the same place. Decision 4 stands as written: the guard runs ONCE per
  match, before any instance is dispatched, and the extras are queued without
  consulting it.
- The declared limitation's premise ("the engine has no batch identity on
  events") is no longer true, so the implementing PR may be able to read the
  doubler count off the whole batch rather than off its first event. If it
  still takes the first-event reading, it should say so on its own terms rather
  than cite a missing identity. The implementation retains that reading: the
  count is fixed when the first matching event is harvested, before later
  members of the batch are emitted. A batch ID alone does not expose those
  future members.

### Decision 5 — An extra instance is an ordinary item with one attribution field

```go
// on StackItem
// DoubledBy names the CR 603.2d effect this instance exists because
// of: the doubler's instance ID and its name as it was when the
// ability triggered. Zero on the first instance and on every item
// that was not doubled.
DoubledBy     uuid.UUID
DoubledByName string
```

- **Not a copy.** `IsCopy` stays false. An extra instance is the ability
  triggering again (the rulings), so anything that asks "was this a copy"
  answers no, and the ability-copy seam (Strionic Resonator) stays separate.
- **`Label` and `Key` are identical to the first instance.** That keeps the
  three things that already key on them correct with no change:
  `triggerInFlightLocked`, `seatNeedsTriggerOrder` (identical instances don't
  prompt for order, as two Bident draws don't today), and
  `TurnTally.Triggered` (`turn_tally.go:349`), which counts every instance
  because the ability did trigger that many times.
- **The frames carry it.** `triggerResumeFrame` and `pickTargetFrame` gain a
  `doubledBy doublerRef`, and the item `Build` returns after the prompt is
  answered gets the field stamped on it, as `targetSpec` is stamped today.
- **Clone and snapshot.** `cloneStackItem` copies both fields.
  `stackItemSnapshot` gains `doubledBy` / `doubledByName` (omitempty), restore
  reads them, and the snapshot round-trip test gains a doubled item. There is
  no new `Game` state: the count is computed and never stored, and the harvest
  pass lives on the call stack. A file written before the fields existed
  decodes as "not doubled", which is correct. No schema bump.
- **Undo.** An extra instance is on `PendingTriggers` or `StackMeta` or in a
  prompt frame, and those are already cloned. Undo restores it the way it
  restores any trigger.
- **Wire (owner-decided, [question 1](#decided-2026-09-17)).** The field
  reaches the client as two additive fields, `doubled_by` / `doubled_by_name`,
  on `StackItemView` and on the `trigger_prompt` / `pick_target`
  `PendingChoiceView`. The stack overlay entry and the prompt show a tag, for
  example "Mulldrifter — draw two cards · additional (Panharmonicon)". There is
  **no log line**: the public log records changes to the game's outcome and
  turn structure, not per-object annotations, and it shows no triggers today.
  There is no reveal-strip cue either. The engine field also serves tests,
  which assert attribution, and the server log.

**Linked abilities.** "The exiled card" of a doubled linked ability means every
card any instance exiled. The catalog's shared record, `b27ExiledWith` (which
reads the event log by source; Angel of Serenity and Duplicant use it),
already returns all of them. Sub-PR 2 greps the catalog for card code that
stores a linked result in a single per-source slot, and each hit is fixed or
declared on its card.

### Decision 6 — Cause helpers make each card one line

`cards/effects/trigger_doubling.go`, a vocabulary file (ADR 0049 D1):

| helper | printed cause | catalog examples |
|---|---|---|
| `DoublesEntering(filter)` | "If a [filter] entering causes a triggered ability of a permanent you control to trigger" | Panharmonicon (artifact or creature), Yarok / Elesh Norn / Starfield Vocalist (permanent), Naban (Wizard you control), Ancient Greenwarden (land), Traveling Chocobo (land or Bird you control) |
| `DoublesLeaving(filter, diesOnly)` | "… dying causes …" / "… leaving the battlefield causes …" | Teysa Karlov, Drivnod (creature, dies only); Gandalf's second half |
| `DoublesAttacking(filter)` | "If a creature [you control] attacking causes …" | Isshin (any creature), Wulfgar (creature you control), Windcrag Siege's Mardu mode |
| `DoublesCasting(filter)` | "If you casting … causes …" | Veyran (instant or sorcery, the cast half) |
| `DoublesCombatDamageToPlayer(filter)` | "If a creature you control dealing combat damage to a player causes …" | Felix Five-Boots |
| `DoublesAbilitiesOf(filter, opts)` | "If a triggered ability of a [filter] you control triggers" | Katara (Ally), Annie Joins Up (legendary creature), Harmonic Prodigy (Shaman or another Wizard), Roaming Throne (another creature of the chosen type, `NamedTribe`), Cloud (Cloud or an Equipment attached to it, while equipped), Echoes of Eternity (colorless spell or another colorless permanent) |

Every helper:

- **checks the controller**: `q.SourceLKI.Controller == q.DoublerLKI.Controller`.
  The one exception is an opt-out for Cloud, whose text has no "you control";
- **refuses `FromSpell`** unless the option allows spells (Echoes of Eternity
  only);
- **combines with `AnyOf`**, for Gandalf's "entering or leaving".

**Every helper path ships with a real card** (owner policy, 2026-09-17). A
helper or option is built in the PR that ships the first card using it, never
ahead of one. In the first wave that is every row above except three pieces,
which wait for their card: `DoublesLeaving`'s "leaving" variant and `AnyOf`
(Gandalf the White), and `DoublesAbilitiesOf`'s spells option (Echoes of
Eternity, waiting on #666). Until then `FromSpell` is refused by every helper
without an option to turn it off.

The event cases are pinned as follows:

- **Entering**: `EventETB`, `EventZoneMove` into the battlefield, and
  `EventTokenCreated`. CR 701.7a defines creating a token as putting it onto
  the battlefield, and one ability watches one of these kinds, never two for
  the same entry.
- **Dying**: `EventLTB` or `EventZoneMove` from the battlefield to a graveyard,
  where the subject's LKI is a creature. `EventSacrifice` does not match (the
  Teysa ruling). A "leaves the battlefield" trigger on a creature that died
  does match, because it watches the same `EventLTB`.
- **Attacking**: `EventAttack` only. `EventTapCard` from attacking does not match
  (the Isshin ruling). Tokens created "tapped and attacking" were never
  declared as attackers and emit no `EventAttack`, which is correct.
- **Combat damage to a player**: a combat `EventDealDamage` whose `Target` is a
  seat and whose `Actor` is the doubler's controller. Life gained from
  lifelink is `EventChangeLife` and does not match (the Felix ruling).
- **Casting**: `EventCast` whose `Actor` is the doubler's controller and whose
  subject matches the filter.

### Decision 7 — Scope, stated

In scope: every triggered ability the harvester dispatches from an event (the
four harvest sites, including `FromStack`), plus evoke's sacrifice trigger.

Out of scope:

- **Trigger suppression**: Elesh Norn's third paragraph ("Permanents entering
  don't cause abilities of permanents your opponents control to trigger") and
  Torpor Orb. This is the "Ability-suppression static" row in
  `docs/engine-seams.md`. It would sit at the same point in `harvestMatchLocked`
  (a suppressor makes the count zero before the doublers are asked), so this
  design leaves room for it, but it does not build it.
- **Copying an ability item**: Strionic Resonator (the "Ability-copy" row).
- **The "or copying" half of Veyran**, until a spell-copy event exists.
- **Echoes of Eternity's "copy it" half**, which is #666 for a permanent spell.
- **Delayed and reflexive triggers** (CR 603.2d) and the manual
  `AnnounceTrigger` button (Decision 2).
- **Prompt batching for many identical optional instances.** Owner-decided
  ([question 3](#decided-2026-09-17)): each doubled optional instance gets its
  own yes/no prompt. A grouped "how many of these N?" prompt is built only if
  play shows the prompt count matters, and it would need its own design.

No bot legal-move change: extra instances produce more prompts of kinds the
enumerator and the heuristic already answer (`trigger_prompt`, `pick_target`,
`trigger_order`).

### Consequences

- **Doubled triggers are rules-correct for every catalog trigger at once**,
  including ward (a catalog `TriggeredAbility`, so Katara doubles a warded
  Ally's ward trigger and the caster pays twice), Saga chapters under a
  source-keyed doubler, and evoke's sacrifice.
- **More prompts.** Every optional or targeted instance asks separately. With
  two Panharmonicons and a "you may" watcher, five tokens entering are fifteen
  prompts. The owner accepted this for the first release (question 3); grouping
  waits for evidence from play.
- **The loop breaker trips sooner on big doubled turns.** ADR 0055 counts
  resolutions per `(source, label)` since the last decision, and threshold 25
  was "unreachable by accident". Nine creatures entering together under two
  doublers, with one mandatory "whenever a creature enters" watcher, is 27
  resolutions with no prompt answered. The notice only suspends autopass, and
  the first pass clears it. This is accepted and recorded here, and the threshold
  is not changed for it.
- **Harvest cost**: at most one extra battlefield walk per event that matched
  a trigger (lazy, memoised per event). `BenchmarkHarvestETBEvent_80perms`
  (`cards/effects/bench_test.go:84`) is re-run in sub-PR 1 and the numbers go
  in the PR description.
- **`dispatchTriggerLocked` keeps its name and meaning** for the reflexive
  path. The harvest callers move to `harvestMatchLocked`.
- **Cloud, Midgar Mercenary loses its caveat** in sub-PR 2.
- **`docs/engine-seams.md`**: the "Trigger doubling" row moves to Closed when
  sub-PR 2 lands (the seam and its first real cards reach develop together). The cards still waiting on a second seam (Elesh Norn's
  suppression, Echoes' copy, Gandalf's flash permission, Ancient Greenwarden,
  Traveling Chocobo's #765) move to that other seam's row.

### Alternatives considered

- **Count inside `dispatchTriggerLocked`** (the issue's suggestion). Since
  #636 this would double reflexive triggers, which CR 603.2d excludes, and it
  would need a "don't double" flag that every future dispatch caller would have
  to remember. Rejected.
- **Declared causes (an enum `Entering | Dying | Attacking | …`) instead of a
  predicate.** It doesn't cover Gandalf, Cloud or Roaming Throne without a
  second escape hatch, and the helpers give the one-line cards the same
  brevity. Rejected.
- **A `Multiplier` field on the item** (one item, resolved N times). This
  breaks every ruling about separate choices: a separate target, a separate
  "you may", and separate responses (a Stifle removes one instance, not all
  of them). Rejected.
- **Cloning the built item N times after `Build`.** Build runs once, so
  per-instance targets and optional answers are impossible, and an item whose
  `Build` depends on announce-time state would be shared. Rejected in favour
  of N full dispatches.
- **Bypassing `OncePerBatch` by giving extras a different `Key`.** The next
  event of the batch would no longer see the extras as in flight, the ordering
  prompt would treat identical instances as different, and `TurnTally` would
  split one ability into two keys. Rejected.
- **Doublers as `StaticAbility` in the layer system.** Nothing is applied in a
  layer, and the harvester would have to read layer output that has no reason
  to exist. Rejected.
- **An eager per-game doubler index** kept up to date on every zone change.
  It is faster only for events that match a trigger, and it is a second source
  of truth for "what is on the battlefield" that undo and snapshot would have
  to carry. The lazy per-event scan is enough. Rejected for now.

### PR split

**Sub-PR 1 — engine, no cards.**
- `game/trigger_doubling.go`: `TriggerDoubler`, `TriggerDoublingQuery`,
  `harvestPass`, `triggerDoublersLocked`, `harvestMatchLocked`,
  `dispatchTriggerInstanceLocked`.
- `CardDef.TriggerDoublers`, `CatalogTriggerDoublers`, `Spec.TriggerDoublers`,
  `buildDef`.
- The four harvest sites move to `harvestMatchLocked`. `OnEvent` builds the
  pass and captures the subject before `harvestLTB`'s delete.
- Evoke's sacrifice trigger goes through the count.
- `StackItem.DoubledBy` / `DoubledByName`, the frames, `cloneStackItem`,
  `stackItemSnapshot`.
- Engine tests 1–14 below, with doublers stubbed through `CatalogTriggerDoublers`.
- The benchmark re-run.
- **Merges only with sub-PR 2.** Sub-PR 2 is opened against this PR's branch,
  and the two reach develop together, so no develop build carries the seam
  without a real card using it (owner policy, 2026-09-17).
- Checks: `go test ./internal/game/... ./internal/ws/... ./internal/aiseat/... ./internal/legal/...`,
  then `go test ./...` and `make lint`.

**Sub-PR 2 — vocabulary and reference cards.**
- `cards/effects/trigger_doubling.go` (Decision 6).
- Panharmonicon, Teysa Karlov and Isshin, Two Heavens as One, each with the
  rulings above as card tests. Cloud's caveat removed. Panharmonicon's tests
  include an evoked Mulldrifter, so the evoke-sacrifice path ships with a real
  card.
- The Saga chapter I reading checked (Decision 2) and the linked-state grep
  (Decision 5).
- Census regenerated (`go test ./internal/cards/coverage -update`), and the
  clone-baseline gate passes.
- The engine-seams row moves to Closed.

**Sub-PR 3 — wire and client** (owner question 1, option (a)).
- `StackItemView.doubled_by` / `doubled_by_name`, and the same on the
  `trigger_prompt` / `pick_target` `PendingChoiceView`. Update
  `docs/protocol.md` and `client/src/lib/protocol.ts`.
- The stack overlay and `ChoicePromptModal` show the tag
  "· additional (Panharmonicon)" after the item's label. No log kind and no
  reveal-strip cue.

**Card PRs** — the rest of the first wave (owner question 2, option (b)):
Katara, Naban, Yarok, Annie Joins Up, Starfield Vocalist, Harmonic Prodigy,
Roaming Throne, Felix Five-Boots and Wulfgar of Icewind Dale, plus Elesh Norn,
Mother of Machines and Veyran, Voice of Duality with declared caveats (see
[Cards](#cards)). Each card follows AGENTS.md: completeness declared, caveats
weaker than printed and never stronger.

### Test plan

Engine (`internal/game`, doublers stubbed):

1. **One doubler, mandatory trigger**: two items on `PendingTriggers`, same
   `Label`, the second with `DoubledBy` set, and two `EventTrigger`.
2. **Stacking**: two doublers give three instances, three give four.
3. **The query carries the right controllers**: `SourceLKI.Controller` is the
   player who controlled the source when it triggered, including a stolen
   creature that dies (read off its LKI), and `DoublerLKI.Controller` is the
   doubler's. The helpers' "you control" check is tested on the cards
   (test 15).
4. **Optional instances**: two `trigger_prompt`s. Declining one leaves the
   other, and each "yes" builds its own item.
5. **Targeted instances**: two `pick_target`s with independent picks. With an
   empty legal set, both are dropped with no prompt (CR 603.3d).
6. **`OncePerBatch`**: three `EventAttack` in one batch under one attack
   doubler gives exactly two instances. A second batch after both resolve
   triggers again.
7. **`OncePerBatch` limitation pinned**: a non-matching first event followed
   by a matching one gives one instance (Decision 4).
8. **Doubler dying alone**: a stubbed "creature dying" doubler on a creature
   destroyed by `DestroyPermanentForEffect`, with a dies-watcher on the
   battlefield, gives two watcher instances.
9. **Doubler in a wipe**: doubler, watcher and a third creature all destroyed by
   `DestroyPermanentsForEffect` give two instances per death for the watcher,
   including the death announced after the doubler's own.
10. **Legend rule**: a second copy of the doubler enters, one goes to the
    graveyard, and a dies-watcher of the dying copy triggers three times.
11. **Doubler entering in a batch**: `putOntoBattlefieldFromZoneLocked` with
    the doubler and two creatures, and a "whenever a creature enters" watcher,
    doubles all three announcements.
12. **Ability removal**: a doubler with `AbilitiesRemoved` adds nothing, both
    live and in LKI.
13. **Exclusions**: a reflexive trigger from a doubled source's resolution, a
    delayed trigger scheduled by a doubled source, a warp exile, and a manual
    `AnnounceTrigger` each queue exactly one item. Evoke's sacrifice under an
    entry doubler queues two.
14. **Clone and snapshot**: a doubled item on `PendingTriggers`, one on
    `StackMeta`, and one inside a `pick_target` frame survive
    `Clone`/`RestoreFrom` with `DoubledBy` intact. The snapshot round trip
    keeps the two fields, and a snapshot without them decodes as not doubled.

Cards (`internal/cards/effects`):

15. **Panharmonicon**: an ETB doubled; its own entry doubles an "artifact
    enters" watcher; a land entering is not doubled; a creature entering under
    an opponent's control doubles my watcher; an opponent's Panharmonicon
    doesn't double my watcher; "enters with a counter" is not doubled.
16. **Teysa Karlov**: a dies trigger doubled; "whenever you sacrifice" not
    doubled; "leaves the battlefield" of a dying creature doubled; an artifact
    creature dying doubles "whenever an artifact is put into a graveyard from
    the battlefield"; Teysa dying herself doubles Blood Artist.
17. **Isshin**: "whenever this attacks" doubled, and "becomes tapped" from
    attacking not doubled.
18. **Cloud, Midgar Mercenary**: an attached Equipment's combat-damage trigger
    doubled while equipped, and not doubled once the Equipment is moved.
19. **Elesh Norn, Mother of Machines**: an entry doubled for her controller's
    watcher, and an opponent's entry watcher still triggers (the declared
    caveat, pinned so the suppression seam changes it deliberately).
20. **Veyran, Voice of Duality**: a magecraft trigger from casting an instant
    doubled; the "copy" half is the declared caveat.
21. **Catalog soak** (#601) with the first-wave cards added: no stalls on the
    extra prompts.
22. **Clone-baseline gate**: no new duplicate bodies (the helpers exist so
    that holds).

Client (sub-PR 3): the overlay tag and the prompt tag are checked by hand
until #689, and any pure helper (the tag text) is unit-tested with vitest.

### Cards

Oracle text checked against the Scryfall dump on 2026-09-17. "Doubler only"
means that on a read of the full oracle text, every other clause already has an
engine seam.

**Doubler only, or doubler plus clauses that already exist:**

| card | the doubling text | other clauses |
|---|---|---|
| Panharmonicon | "If an artifact or creature entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time." | — |
| Teysa Karlov | "If a creature dying causes a triggered ability of a permanent you control to trigger, …" | "Creature tokens you control have vigilance and lifelink." |
| Isshin, Two Heavens as One | "If a creature attacking causes a triggered ability of a permanent you control to trigger, …" | — |
| Katara, the Fearless | "If a triggered ability of an Ally you control triggers, …" | — |
| Naban, Dean of Iteration | "If a Wizard you control entering causes a triggered ability of a permanent you control to trigger, …" | — |
| Yarok, the Desecrated | "If a permanent entering causes a triggered ability of a permanent you control to trigger, …" | deathtouch, lifelink |
| Annie Joins Up | "If a triggered ability of a legendary creature you control triggers, …" | ETB: 5 damage to target creature or planeswalker an opponent controls |
| Starfield Vocalist | "If a permanent entering the battlefield causes a triggered ability of a permanent you control to trigger, …" | warp {1}{U} |
| Harmonic Prodigy | "If a triggered ability of a Shaman or another Wizard you control triggers, …" | prowess, written as a cast trigger (no `prowess` helper exists yet) |
| Roaming Throne | "If a triggered ability of another creature you control of the chosen type triggers, it triggers an additional time." | ward {2}, "as this creature enters, choose a creature type", "is the chosen type in addition" |
| Felix Five-Boots | "If a creature you control dealing combat damage to a player causes a triggered ability of a permanent you control to trigger, …" | menace, ward {2} |
| Wulfgar of Icewind Dale | "If a creature you control attacking causes a triggered ability of a permanent you control to trigger, …" | melee, written from `EventAttack` (the issue's borderline card) |

**Doubler plus a clause another seam owns:**

| card | the other clause | waits on |
|---|---|---|
| Elesh Norn, Mother of Machines | "Permanents entering don't cause abilities of permanents your opponents control to trigger." | ability-suppression static (no issue yet). **Ships in the first wave** with the suppression declared as a caveat (owner question 2) |
| Veyran, Voice of Duality | "cast **or copy**", both halves | a spell-copy event. **Ships in the first wave** with the copy half declared as a caveat (owner question 2) |
| Echoes of Eternity | "Whenever you cast a colorless spell, copy it." | #666 |
| Gandalf the White | "You may cast legendary spells and artifact spells as though they had flash." | the per-player flash row |
| Ancient Greenwarden | "You may play lands from your graveyard." | play from graveyard |
| Traveling Chocobo | "You may play lands and cast Bird spells from the top of your library." | #765 |
| Windcrag Siege | "As this enchantment enters, choose Mardu or Jeskai." | shipped: the as-enters choice landed in #1646 (ADR 0071's 2026-09-27 amendment), and the doubler itself is gated by `TriggerDoubler.ActiveWhen` (#1647) since a `TriggerDoubler` is not an ability-list entry the ordinary designation gate can reach |

**Noticed while checking, not in the issue's list:** Drivnod, Carnage Dominus
has Teysa's exact text. Its other ability ("{B/P}{B/P}, Exile three creature
cards from your graveyard: Put an indestructible counter on Drivnod.") is not
checked here, and the card goes to whichever batch issue owns it.

### Decided (2026-09-17)

Answered by the owner on 2026-09-17. The options are kept, the chosen one is
marked **(chosen)**, and the recommendation text is kept for the record.

1. **How does a player see that a trigger is an extra instance?** Today the
   public log doesn't show triggers at all (`protocol/log.go` projects no
   `EventTrigger`). Players see a trigger as its stack-overlay entry and, when
   it asks something, its prompt.
   - (a) **(chosen)** A tag on the stack-overlay entry and on its prompt, for example
     "Mulldrifter — draw two cards · additional (Panharmonicon)". This needs
     two additive wire fields and no log kind.
   - (b) (a) plus a public log line per extra instance.
   - (c) Nothing: the instances look identical.

   **Recommendation: (a).** A player's question is "why is this on the stack
   twice?" and they ask it while looking at the stack or the prompt, so the
   answer belongs there. A log line for extra instances only, in a log that
   shows no triggers at all, would stand out for the wrong reason. (c) makes
   Panharmonicon look like a bug.

   **Decision: (a).** No log line, under the cross-ADR log rule (the log
   records changes to the game's outcome and turn structure, not per-object
   annotations), and no reveal-strip cue. Applied in Decision 5 and sub-PR 3.

2. **Which cards ship in the first wave?**
   - (a) The twelve "doubler only" cards above, plus Cloud's caveat removal.
   - (b) **(chosen)** (a) plus Elesh Norn, Mother of Machines and Veyran, Voice of Duality,
     each with a declared caveat for the missing half. Both caveats are weaker
     than printed.
   - (c) The seam plus the three reference cards (Panharmonicon, Teysa,
     Isshin) and Cloud, leaving the rest to the batch issues.

   **Recommendation: (b).** Both omissions only weaken their controller, which
   AGENTS.md allows once declared, and Cloud already set that precedent. Elesh
   Norn's doubling half is still a strong commander on its own. The cost is
   that a player who picks Norn for the hoser half doesn't get it until the
   suppression seam exists, and the caveat has to say so plainly.

   **Decision: (b).** Elesh Norn and Veyran ship with their caveats declared
   (tests 19 and 20 pin them). Under the owner's cross-cutting policy every
   new seam path ships with at least one real card, so sub-PRs 1 and 2 reach
   develop together and helpers without a first-wave card wait for it
   (Decision 6).

3. **Many identical optional instances: one prompt each, or grouped?** Two
   Panharmonicons and a "you may" watcher turn five tokens entering into
   fifteen yes/no prompts.
   - (a) **(chosen)** One prompt per instance, as the rules describe. This ships with
     sub-PR 1 and needs no new prompt kind.
   - (b) Group untargeted identical optional instances (same source, same
     label, same event) into one "how many of these N?" prompt. This is a new
     prompt shape with enumerator and bot work.

   **Recommendation: (a) now, and (b) only if play shows the spam matters.**
   Tokens entering already give one prompt per token today, so doubling
   multiplies an existing cost without creating a new one. (b) is a real
   feature with its own ADR-sized questions (what happens when the instances
   differ in `NoLegalTarget`, and how bots answer a count).

   **Decision: (a).** One yes/no prompt per doubled optional instance.
   Grouping is considered only if play demands it.


### Implementation checkpoint (2026-09-17)

The #752 implementation combines the engine, reference-card, and wire/client
stages so the mechanic ships with Panharmonicon, Teysa Karlov, Isshin and
Cloud's previously missing ability. The remaining card wave stays on #752.

Review also found that characteristics alone did not preserve a copied
permanent's catalog identity or an Equipment's attachment through a zone
change. The leave snapshot now carries a small companion record containing
the oracle ID, active face and attachment. Harvesting reads that record before
discarding it alongside the existing LKI; clone and snapshot round trips carry
it too. This keeps copied doublers and attached Equipment triggers correct
when they leave individually, as well as in a simultaneous batch.

The Saga reading was checked against the pinned [August 2026 rules
text](https://media.wizards.com/2026/downloads/MagicCompRules%2020260819.txt),
714.2b and 714.3a, and Wizards' [Saga
explanation](https://magic.wizards.com/en/news/feature/dominaria-mechanics-2018-03-21).
Chapter abilities trigger from lore counters; the entry counter is a
replacement effect. Therefore entry doublers do not double the chapter just
because its Saga entered. A regression test uses an artifact Saga so its
permanent type would otherwise satisfy Panharmonicon's filter.

The linked-state audit searched the catalog's exile/imprint helpers and
stored card IDs. `b27ExiledWith` keeps all matching exiles across resolutions;
Angel of Serenity's linked return already consumes that whole list. Duplicant
intentionally selects the last creature because its Oracle text says so.
Delayed-return payloads carry their own slices; none of these paths overwrites
a single per-source linked-card slot. The doubled Angel regression independently
targets two creatures and requires both to return.

### Note (2026-09-22, #1184): a trigger can watch an ACTIVATION now

`triggerHarvester.OnEvent` returns immediately on `EventTrigger`, and
that early return is still right: `EventTrigger` is the "an item
reached `PendingTriggers`" breadcrumb, a trigger that fires further
triggers does so at RESOLUTION, and re-entering the harvest at
announce would spam. The cost of it was that nothing could watch an
activation at all, because `ActivateCatalogAbility` announced with
that kind and no other — and with no ability identity on the event
either.

`EventActivateAbility` (`game/events.go`) is the announcement said in
a kind the harvest walks like any other. It carries the ability's
`Label` — the same string the activation record is keyed by — and an
`Exhaust` bit, both stamped at the announce because by the time a
watcher runs the source may be gone (a `SacrificeSelf` cost) and the
ability list may have been renumbered.
`EventManaAbilityActivated` carries the same two stamps, so a watcher
that means "any activated ability" (CR 605.1a) watches both kinds:
`effects.WheneverYouActivateAnExhaustAbility` does exactly that.

Nothing else about the harvest changes. The trigger goes on the stack
ABOVE the ability that made it and resolves first (CR 603.3b), the
zone wrappers work as they do for any other event — Afterburner
Expert's is `InGraveyard` — and the public log stays silent on the new
kind for the reason it is silent on `EventTrigger`: the LogResolve of
the ability it becomes already says it.

See [ADR 0020](0020-activated-abilities.md)'s note of the same date
for the other two seams the same issue opened, and
[docs/engine-seams.md](../engine-seams.md) for the "whenever an
opponent activates an ability" row that closes on this event next.

### Note (2026-09-22, #1210): the opponent-activation watch, on that same event

The note above ends by pointing at the `docs/engine-seams.md` row that
"closes on this event next". This is that row, closed, and it needed
one constructor and no engine change at all — which was the claim
#1184 made when it built the event wider than its two cards.

`effects.WheneverAnOpponentActivates(label, of, includeMana, build)`
(`cards/effects/triggers_common.go`) is the one declaration shape:

- **Whose activation.** `ByAnOpponent` on `Event.Actor`, the same
  predicate "whenever an opponent draws a card" uses. The ACTOR and
  not the source's controller, because that is what the cards print
  ("whenever an opponent activates an ability") and because the two
  differ exactly where it matters — an ability activated from a
  permanent somebody else controls.
- **Which abilities.** `includeMana` picks the watched kinds:
  `EventActivateAbility` alone, or both it and
  `EventManaAbilityActivated`. "…if it isn't a mana ability" is
  therefore not a predicate a card writes and can forget — it is the
  absence of a kind from `Watches`, decided once, at the only place
  that knows (CR 605.1a: a mana ability never reaches the stack and
  has its own event kind precisely so a watcher can tell).
- **Of what.** `of CardPredicate` runs against the ability's SOURCE
  object, looked up live on the battlefield — "an ability of an
  artifact, creature, or land **on the battlefield**" (Harsh Mentor),
  "of a creature or land" (Runic Armasaur). A nil predicate is "any
  source", which is the plain "whenever an opponent activates an
  ability" clause.
- **"That player."** `build func(activator uuid.UUID) Effect` is
  handed `Event.Actor`, captured in the `Build` closure the way Ob
  Nixilis, the Hate-Twisted captures a drawer. It does NOT target:
  Harsh Mentor's 2 damage goes to the activator with no target
  chosen, so hexproof and "can't be the target of" do nothing about
  it, and CR 608.2b re-checks nothing at resolution.

`Optional(…)` composes as it does with every other trigger, which is
the whole of Runic Armasaur's "you may draw a card".

One engine line did change, and a back-out is what found it:
`EventManaAbilityActivated` set `Source` but not `CardID`, so a
source predicate looking the ability's object up by `CardID` matched
nothing on the mana path — a silent miss, not an error. Both emit
sites (the hand click and the auto-tapper's executor, CR 605.3) now
carry the same stamps `EventActivateAbility` always has, which is
what #1184 intended when it said a watcher can read the same fields
off either kind.

**Cards:** Harsh Mentor and Runic Armasaur, both `full`.

What stayed open on this family, and was a different row: a watch on
an ability's activation whose TARGET clause reads the triggering event
(`docs/engine-seams.md`, "Trigger/target clause reading the triggering
event's data"). A `Build` closure can capture an event field and hand
it to a non-targeted effect — that is what this note's `build`
argument is — but `TriggeredAbility.Targets` is a static `TargetSpec`
evaluated by the harvester and cannot see the event at all. **Closed
the same day by #1223** — see the amendment below, whose
`TriggeredAbility.TargetsFrom` is exactly that clause.

## Amendment 2026-09-22 — the triggering event is DATA on the item (CR 603.2, CR 603.10) · Accepted · S45

Issue [#1223](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1223).
Decision 2 above says `Build` builds and does not resolve. It is
silent on what happens to the EVENT, and the answer the catalog
settled on by default was "capture it in the `Effect` closure". That
works for exactly one reader and fails for three.

### What was actually missing

`TriggeredAbility` is handed the event twice — `AppliesTo(ev, …)` and
`Build(ev, …)` — and both hands go away immediately. `AppliesTo`
returns a bool; `Build` returns a `*StackItem`, and the only way past
it is a closure. A closure reaches resolution and reaches nothing
else:

- **The target clause.** A trigger's targets are chosen as it is put
  on the stack (CR 603.3d), by the harvester, from
  `TriggeredAbility.Targets` — STATIC catalog data declared before the
  game started. Scrap Trawler's "target artifact card in your
  graveyard with lesser mana value" could not be stated at all,
  because "lesser" is lesser than the mana value of the artifact that
  just died and no closure runs between the event and the clause.
- **The copy.** CR 707.10 copies a triggered ability together with the
  choices made when it triggered, and the Strionic Resonator rulings
  are explicit that the copy remembers the same event. A closure
  reached a copy only by accident of sharing a func pointer, and
  nothing said so.
- **The snapshot.** A func has no wire form, so a game paused on an
  unanswered CR 603.3d prompt could not be a restore point for any
  trigger whose effect needed its event.

The catalog's other way of reaching a past event — walking `g.Events`
backwards for the last one of a kind — is not a fourth answer either.
A sweep of all 29 log-walking helpers in `cards/effects` found NONE
re-deriving its own triggering event: every one of them is reading
game HISTORY (a per-turn tally, a counter total before the change, the
nearest resolution's actor) that the triggering event could never
answer. That is a good thing and this field is what keeps it true —
such a scan finds the wrong event the moment two of a kind land in one
batch (two creatures dying to one wrath), and nothing at all a
priority round later.

### Decision 8. One typed value, `StackItem.Trigger`

```go
type TriggerContext struct {
        Event  Event           // the triggering event, verbatim
        Object *ObjectSnapshot // CR 603.10 LKI about Event.CardID
}
```

`Event` whole, rather than a hand-picked subset, for two reasons: the
alternative is a new field for every clause that wants one more fact,
and `Event` is already the flat, closure-free, JSON-tagged shape the
log is built out of. `Amount` is the damage dealt or the life gained
or the counters placed; `StackItemID` is the ability that was
activated; `OldZone` / `NewZone` say where an object went.

Read by the effect as `effects.Context.Trigger()` and by the clause as
`TriggeredAbility.TargetsFrom`. The zero value — `Fired()` false — is
the honest answer for every cast spell, every activated ability and
every item restored from a snapshot written before the field existed.

It is **plain data**, which is the whole decision: it survives a copy
(copying it is copying a struct), it survives a snapshot (so a paused
trigger prompt is still a restore point, `carried` in
`snapshot_drift_test.go`), and it survives the log being walked past.

### Decision 9. `ObjectSnapshot` is a flat value, and it is NOT a `Card`

Half of what a trigger wants to know is about an object that has gone:
the creature that died, the artifact put into a graveyard. CR 603.10
judges such an ability on what the object looked like immediately
before the event, and the engine already keeps that reading in
`Game.lastKnownBattlefield` — written by every battlefield exit,
DELETED as `harvestLTB` ends. `objectSnapshotLocked` reads it while it
is still there and copies by value.

Not a `*Card`: a pointer into a zone slice is already invalid when an
LTB trigger is built. Not a `Card`: it carries `effective`,
`PrintedSelf` and the rest of the layer bookkeeping, none of which
means anything once the object is gone and all of which would have to
be snapshotted to keep the item restorable.

The two halves come from different places on purpose. Types, colours
and name are CHARACTERISTICS and come off the CR 603.10 snapshot; the
oracle id, the owner and the mana value are facts about the CARD and
come off the card, because a permanent's mana value is computed from
its printed cost (CR 202.3b) and `Characteristic` carries none.

**Power and toughness are deliberately absent.** "A creature with
power equal to the power of the creature that died" is a real printed
clause and is not answerable yet: `lastKnownBattlefield` holds a
`Characteristic`, the engine applies +1/+1 and -1/-1 counters OUTSIDE
it (`Card.PowerForComparison`), and `MoveCard` clears a permanent's
counters on the way off the battlefield — so by the time a
dies-trigger is harvested the number is gone from both places. A
`Power` field would silently read 2 for the 3/3 a counter made, which
is the one case such a clause is written about. It waits for a second
LKI store, and ADR 0072 §2 already records what a second store costs.

### Decision 10. The clause is a function of the event: `TargetsFrom`

```go
TargetsFrom func(tc TriggerContext, source *Card, g *Game) *TargetSpec
```

Non-nil wins over `Targets`, which such a card leaves nil. A FUNCTION
returning a whole spec rather than a predicate that receives the
context, because the event changes a clause's COUNT and LABEL as
readily as its predicate — "up to X target creatures", where X is the
damage dealt — and only a spec can say all three.

It is called at every point the dispatch needs the clause (the CR
603.3d fillability check, the target walk, the spec stamped on the
item for the CR 608.2b re-check) rather than cached in a frame, and
the contract is that every call must agree: the answer is a function
of the trigger context, which is carried and therefore stable, and of
the board, which is not and must not be — a clause re-read after an
optional prompt should see the board as it is, exactly as
`refreshTargetChoicesLocked` re-reads a frozen legal set for the same
reason.

### Decision 11. The stamp is the dispatch's, not the card's

`Build` is catalog code, ~2,200 declarations of it, and every one
would have had to remember. The two sites that call `Build`
(`buildOrPickTriggerLocked` and `finishPickTargetLocked`) call
`stampTriggerContext` immediately afterwards instead, so a card gets
the event on its item whether or not its author thought about it.

The context is built ONCE, in `dispatchTriggerInstanceLocked`, and
carried from there through every prompt frame (`triggerResumeFrame`,
`modePickFrame`, `pickTargetFrame` all hold `tc TriggerContext` where
they used to hold `ev Event`). That is not tidiness: the CR 603.10
snapshot is deleted from `lastKnownBattlefield` as the harvest ends,
which is before any prompt can be answered, so a context rebuilt later
would be missing exactly the half a dies-trigger needs.

**A CR 603.12 reflexive trigger is stamped with nothing.**
`QueueReflexiveTriggerForEffect` hands the dispatch a synthetic
`EventResolve` describing the resolution that created the trigger,
precisely because there was no triggering event; stamping it would
tell a card its trigger fired off a resolution and hand it a snapshot
of the parent's source, and both would be lies. The test is an ability
that WATCHES NOTHING — `len(t.Watches) == 0` — which is a fact rather
than a flag, since the harvester refuses to fire an ability that
declares no `Watches`.

### Cards

Scrap Trawler ships `full` (the clause is cut to the mana value of the
artifact that died, so a chain resolves in descending order as it does
in paper). Cloudstone Curio ships `caveats` — "shares a permanent type
with it" is built from the event, and the caveat is the Whitemane Lion
posture, that the permanent to return is picked as the trigger goes on
the stack rather than at resolution.

### Still not covered

Power and toughness on the object snapshot (above). Three helpers sit
next to this and are NOT replaced by it, each for a stated reason:
`b16EnteredFromStack` wants the ETB's ORIGIN ZONE, which
`EventETB` does not carry; `b30CastFromHand` wants a different,
earlier `EventCast`; and `random_effects.WheneverYouRollDice` wants
the whole die BATCH, of which the triggering event is one member.

## Amendment 2026-09-24 — resolution-time LKI for a permanent that has left (CR 608.2h) · Accepted · S38

Issue [#1379](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1379).
This ADR owns the engine's CR 603.10 last-known-information store
(`Game.lastKnownBattlefield`, written by `snapshotLKILocked` at every
battlefield exit and read by the harvester), so the resolution-time
sibling lands here rather than under a new number.

### What was missing

CR 603.10 looks back in time when a trigger is HARVESTED. CR 608.2h
is a different rule, applied when an ability RESOLVES: "where X is
that creature's power" is read then, and if the creature has left
the battlefield by that point, the ability uses the creature's
last-known information. The engine had nothing to answer that with.

- `lastKnownBattlefield` is deleted by `harvestLTB` as soon as the
  exit's dispatch ends, a priority round before anything resolves.
- `lastKnownCounters` (#1218) has the same lifetime and holds
  counters only.
- `LookupCardForEffect` finds the card in its new zone, with printed
  power, no counters and no continuous effects.

So every card that read a permanent at resolution chose one of two
wrong answers when the permanent had gone. Cream of the Crop fell back
to the power the creature had when the ability TRIGGERED, which misses
a pump in response. Tribute to the World Tree, Warstorm Surge and
Murderous Redcap did nothing. Claustrophobia could not find the
creature it had enchanted. Each shipped a caveat that said so.

Decision 9 above left "power and toughness on the object snapshot"
open for the same reason: the number was gone from both stores before
a trigger was even built.

### Decision 12. A second LKI store, per departed OBJECT, for the rest of the turn

```go
lastKnownPermanents map[uuid.UUID][]PermanentInfo // game.go

type PermanentInfo struct {
        Epoch          int            // Card.ObjectEpoch while on the battlefield
        Left           bool           // true on a record; false on a live read
        Controller     uuid.UUID
        Characteristic Characteristic // Effective(), as the CR 603.10 snapshot
        Power, Toughness int          // counters included, not clamped
        Counters       map[string]int
        AttachedTo     TargetRef
}
```

Written by `rememberDepartingPermanentLocked` from
`battlefieldExitLocked`, in the same beat, from the same live card and
the same layer cache as the CR 603.10 snapshot. There is no recompute,
for the reason the paid-tap freeze (#759) gives: in a board wipe the
permanents leave one at a time, and a recompute part-way through would
read a creature after its lord had already gone.

**Keyed by object, not by card.** The map key is the instance ID,
which survives a zone change. The value is a list, one entry per
battlefield object that card has been this turn, told apart by
`Card.ObjectEpoch` (CR 400.7). A creature bounced and replayed twice
in response to a trigger is two departed objects, and the trigger
names only the first.

**Power and toughness include counters.** `Characteristic` excludes
them by design, and `MoveCard` zeroes `Card.Counters` a line after the
snapshot. So the record takes `PowerForComparison()` and
`CurrentToughness()` while the counters are still there. This is the
number Decision 9 could not supply.

**Lifetime: the turn.** Cleared at the turn boundary alongside
`lastKnownStack` (#1255), for the argument that record makes. Whatever
refers to a departed permanent is a stack object or a pending trigger,
and the stack is empty before a turn can end. The size is bounded by
the turn's battlefield exits. A card leaving the GAME (CR 800.4a)
drops its records in `forgetObjectLocked`, with the other per-object
maps.

**Carried everywhere.** Clone and RestoreFrom copy it, so an undo
across the removal spell rewinds it. The persisted snapshot carries it
too (`GameSnapshot.LastKnownPermanents`, `carried` in
`snapshot_drift_test.go`). This is unlike `lastKnownStack`, whose only
readers are closures: the object ref below is plain data on the item,
so a data-driven reader can use the record after a restore.

### Decision 13. One read, live-or-last-known: `PermanentForEffect(ObjectRef)`

```go
type ObjectRef struct{ ID uuid.UUID; Epoch int }

func (g *Game) PermanentForEffect(ref ObjectRef) (PermanentInfo, bool)
func (g *Game) PermanentRefForEffect(cardID uuid.UUID) (ObjectRef, bool)
```

The read returns the LIVE permanent (after a layer recompute) while
the named object is still on the battlefield, so a Giant Growth in
response counts. Once the object has left, it returns the record. A
live card with a different epoch is a new object, and the read never
answers with it: the record is the answer.

The ref comes from the trigger context. `ObjectSnapshot` (Decision 9)
gains `Epoch`, the epoch of the object the event was about. For an
exit, that is the epoch the object had ON THE BATTLEFIELD, taken from
the record. It is not the epoch its card has now in the graveyard.
`ObjectSnapshot.Ref()` hands it to the read. On the card side,
`effects.Context.TriggeringPermanent()` is the one-line spelling, and
the proof cards use nothing else. An effect that is not a trigger takes
`PermanentRefForEffect` when it is built.

**Read, never written to.** Last-known information answers questions
about an object. It cannot receive counters, be tapped or be moved.
Tribute to the World Tree's "otherwise put two +1/+1 counters on it"
checks `Left` and does nothing for a creature that has gone. That
includes the case where the card is back on the battlefield as a new
object.

### Cards

Cream of the Crop, Tribute to the World Tree and Claustrophobia ship
`full`. Warstorm Surge and Murderous Redcap keep a narrower caveat.
A creature that has left now deals damage equal to its last-known
power. It deals that damage as a plain source, because the damage
tail reads lifelink and deathtouch off the battlefield only.

### Still not covered

- ~~**Damage-source keywords from a departed source.**~~ **Closed by
  [#1396](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1396)**
  (see the 2026-09-24 note below and
  [ADR 0056's 2026-09-24 amendment](0056-infect-wither-toxic.md)).
- **An exit that bypasses `battlefieldExitLocked`.** A player leaving
  the game takes their permanents with them (CR 800.4a), and those
  permanents get no record. The read answers false, and each card
  chooses the weaker answer.
- **`ObjectSnapshot` still has no power field.** A clause judged at
  HARVEST ("a creature with power equal to the power of the creature
  that died") can now read `PermanentForEffect(snapshot.Ref())` while
  it builds. No card needs it yet.

### Note 2026-09-24 — the record's first engine reader (#1396)

The non-combat damage tail now reads Decision 12's record when the
damage source has left the battlefield: `deathtouch` from the recorded
abilities and `lifelinkTo` from the recorded controller. The decision
belongs to [ADR 0056](0056-infect-wither-toxic.md) (its Decision 2 step
4), so the amendment lives there as Decisions 9-11. Two things it
changed here:

- **`PermanentInfo` gains `Tapped`** (ADR 0056 Decision 11). A status,
  not a characteristic, and last-known information all the same: Mana
  Vault's "if this artifact is tapped" is re-checked at resolution.
  It is written with the rest of the record, from the live card, before
  `MoveCard` clears the flag.
- **An object ref can now name a damage source.**
  `Game.DealDamageFromObjectForEffect(ObjectRef, target, amount)` and
  `effects.DealDamage{SourceObject: &ref}` read the named object's
  record, never a new object the card has become (Decision 13's rule,
  applied to the damage source). A bare instance ID reads the last
  object the card was, and only while the card has not moved since it
  left (ADR 0056 Decision 10).

## Amendment 2026-09-24 — every ability item names its source OBJECT (CR 400.7) · Accepted · S38

Issue [#1418](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1418),
found while building #1396. This ADR owns the trigger dispatch and the
two object reads Decisions 12–13 added, so the fix lands here.

### What was missing

A triggered stack item carried its source's instance ID
(`StackItem.SourceCardID`) and nothing that said which OBJECT that card
was. `StackItem.SourceEpoch` existed (#812) but only the two activation
announce paths stamped it, and `AbilitySourceGoneForEffect` asked it of
activated items only.

Decision 13 gave a trigger one way to name an object: the triggering
event's snapshot, `ctx.TriggeringPermanent()`. That covers a trigger
whose event is ABOUT its own source (an enter trigger, a dies trigger).
An upkeep, draw-step, cast or "another creature enters" trigger has no
source on its event. The only read left was
`PermanentRefForEffect(item.SourceCardID)`, which answers with the object
the card is now. A source that left and came back before its trigger
resolved was read as the new object.

- **Mana Vault** (its caveat after #1396): a Vault flickered in response
  to its draw-step trigger was judged as the new, untapped Vault and
  dealt no damage.
- **Living weapon, Hero's Blade** and any "attach this" trigger: the
  Equipment on the battlefield after a flicker is a new object (CR
  400.7), and the trigger attached it anyway. Only activated equip was
  refused.
- **Uthros Research Craft**'s "put a charge counter on this" landed on
  the new, counterless Craft.

### Decision 14. `StackItem.SourceObject ObjectRef`, stamped on every ability item where it is made

```go
type StackItem struct {
        …
        SourceEpoch  int       // unchanged: post-cost, activated only
        SourceObject ObjectRef // #1418: the object the ability came from
}
```

A new field rather than widening `SourceEpoch`, for two reasons.
`SourceEpoch`'s readers (ninjutsu, hideaway) want the reading taken
AFTER the costs are paid. And zero is a real epoch, so an int alone
cannot say "not stamped", which an item restored from an older
snapshot needs to say. The ref's ID is that bit.

Where the stamp is made. None of these is in catalog code, which is
the Decision 11 rule again: ~2,200 Builds would each have to remember.

| Path | What it names |
|---|---|
| Trigger dispatch, after `Build` (`stampTriggerSource`, beside `stampTriggerContext` at both call sites) | the dispatch's VALUE copy of the source, taken when the ability triggered and carried through every prompt frame. When the event is about the source itself (a dies trigger), the event snapshot's ref instead: the permanent that left, not the graveyard card |
| `queueHarvestedTriggerLocked`, when the item has none | the object the source card is now (`sourceObjectRefLocked`): live permanent, the permanent it was while its exit is being dispatched, the card in another zone, or a ceased token's last record |
| Catalog activation (`activated.go`) | the object read BEFORE any cost is paid, so a source sacrificed to its own ability is still "this permanent", as its record |
| Manual `ActivateAbility`, `AnnounceTrigger` | the object the card is now |
| `ScheduleDelayedTriggerForEffect` (new `DelayedTrigger.SourceObject`) | CR 603.7d: the resolving ability's own `SourceObject` when that ability is scheduling it from the same card, otherwise the object now. Copied onto the fired item |
| `QueueReflexiveTriggerForEffect` | CR 603.12: the parent's `SourceObject` |
| `createAbilityCopyLocked` | CR 707.10: the original's |

Spells carry none: a spell is its own source, and "this spell" is not a
permanent read. Clone, `RestoreFrom` and the persisted snapshot carry
the field on `StackItem` and `DelayedTrigger` (`sourceObject`, nil when
unstamped; `carried` in `snapshot_drift_test.go`).

### Decision 15. One read: `SourceObjectForEffect`, and `SourcePermanent()` on the card side

```go
func (g *Game) SourceObjectForEffect(item *StackItem) (ObjectRef, bool)

func (c *Context) SourceRef() (game.ObjectRef, bool)
func (c *Context) SourcePermanent() (game.PermanentInfo, bool) // = PermanentForEffect(SourceRef())
```

`SourcePermanent` is Decision 13's rule applied to "this": the live
permanent while the object is still there, and its record once it has
gone, including when the card is back as a new object. Like
`TriggeringPermanent`, it is something to read, never something to act
on. A clause that changes the permanent checks `Left` first.

An unstamped ability item (a pre-#1418 snapshot) falls back to
`PermanentRefForEffect`, which is what every reader had before. A spell
answers false.

### Decision 16. `AbilitySourceGoneForEffect` asks the object of every stamped item

The kind gate from [ADR 0036 decision 19](0036-attachments.md) (#812,
"Why the kind, not a sentinel") existed because only activated items had
an epoch. Every ability item has one now, so the gate reads the stamp
instead. The sentinel argument still holds: the ref's ID, not the epoch,
is what says an item was stamped. A stamped item whose source's epoch no longer matches is
gone. An unstamped one keeps the rule it was put on the stack under.

This is a deliberate behaviour change for every trigger that attaches or
counters its own source: living weapon (Batterskull, Batterbone, Kaldra
Compleat), Maul of the Skyclaves, Mithril Coat, Hero's Blade and Uthros
Research Craft. A source flickered in response is not the object the
trigger names, so those effects now do nothing. Each did it to the new
object before. That was stronger than printed (the #259 direction),
because CR 400.7 says the new object has no memory of the ability.

### Cards

Mana Vault ships `full`. Its caveat was this issue, and its upkeep untap
now checks the object too. Batterskull, Hero's Blade and Uthros Research
Craft were already `full` and now honour CR 400.7, one test each
(`cards/effects/source_object_test.go`).

### Still not covered

- **A spell's own object.** A spell item carries no `SourceObject`. "This
  spell" is never a permanent read, and CR 707.10's self-copy has its own
  slot (`Game.resolving`). Nothing asks yet.
- **Continuations frozen before a prompt.** A pause frame that captured an
  instance ID before #1418 still reads the card, not the object. The
  pick-target and trigger-prompt frames carry the source `Card` by value,
  which is why the stamp can be taken from them. Frames are not persisted
  anyway (`ChoiceResumeFrames` in the census).

### Note 2026-09-24 — the record's second engine reader: the CR 608.2b re-check (#1429)

Decision 12's record now also answers the CR 608.2b target re-check
for an ability whose source has left the battlefield.
`stackItemSourceLocked` (`server/internal/game/targets.go`) used to
look the source up in whatever zone held it, which is its graveyard
card, so protection was tested against the wrong object. It now returns
a snapshot of the record, found by `departedAbilitySourceLocked`:

- **A stamped item** (every ability item since Decision 14) names its
  object through `SourceObjectForEffect`. The object is read live while
  it is on the battlefield. Once it has left, its record is used. An
  object that was never a permanent (a graveyard trigger) has no record,
  and its card is read where it is. This is Decision 14's stamp doing
  what it was built for: a graveyard trigger from a card that died
  earlier this turn is judged as the graveyard card, and an activation
  that sacrificed its own source is judged as that source even after
  the card comes back.
- **An unstamped item**, restored from a snapshot written before
  Decision 14, keeps ADR 0056 Decision 10's instance-ID rule ("current
  epoch = recorded + 1").

Spells never consult the record. The decision belongs to ADR 0072
(protection's targeting source, §2), so the amendment lives there:
[ADR 0072 amendment 2026-09-24 (#1429)](0072-protection.md).

## Amendment 2026-09-24 — the primitives that act on "this" ask for the object (CR 400.7) · Accepted · S38

Issue [#1432](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1432),
the follow-on Decision 16 left open. Decisions 14–16 gave every ability
item its source's OBJECT and gave two readers (`AbilitySourceGoneForEffect`
for attach and station, `SourcePermanent()` for "this" as data). The
primitives that ACT on "this" did not ask.

### What was missing

About a hundred catalog files resolve an ability with its source's
instance ID as the target: `AddCounter{Target: item.SourceCardID}` (64
literals), `SacrificePermanent` (13), `UntapTarget` (12),
`BoostUntilEOT` / `GrantKeywordUntilEOT` (9), `BounceToHand`,
`ExileTarget`, `Transform`, `BecomePrepared`, crew's
`BecomeCreatureUntilEOT{}` (a zero target means the source), and a
further ~20 bodies that call a game mutator on the source directly
(`AddCounterThenForEffect`, `UntapTargetForEffect`,
`TuckToLibraryForEffect`, `SetClassLevelForEffect`, …). An instance ID
survives a zone change. A source that left and came back under the same
ID while its ability waited — bounced and replayed, died and
reanimated — is a NEW object (CR 400.7) with no memory of the ability,
and every one of those acts landed on it anyway: the counter, the
untap, the sacrifice that removed the wrong permanent. Stronger than
printed, the #259 direction. (The engine's own exile-and-return mints a
fresh instance ID, so `Flicker` itself never reached this; every route
that keeps the ID did.)

### Decision 17. The question lives in the primitive, keyed on "target == this item's source"

```go
func (g *Game) AbilitySourceIsNewObjectForEffect(item *StackItem) bool // game/ability_source.go
func (c *Context) isNewSourceObject(target uuid.UUID) bool             // effects/source_object_guard.go
func sourceIsNewObject(g *game.Game, item *game.StackItem) bool        // same, for (g, item) bodies
```

Every effects primitive that acts on a permanent asks
`ctx.isNewSourceObject(target)` before it acts and does nothing when
the answer is yes: `AddCounter` (both signs), `SacrificePermanent`,
`TapTarget`, `UntapTarget`, `BounceToHand`, `ExileTarget` (so `Flicker`
and `ExileThenIfItWas`), `DestroyTarget`, `Regenerate`, `GainControl`,
`Transform` (so `TransformThis`), `ExileAndReturnTransformed`,
`PreventNextDamage`, `PutIntoLibrary`, `PutIntoLibraryAtDepthOrBottom`,
`ExileWithPermission`, `Airbend` / `AirbendAll`, `Earthbend`,
`BecomePrepared`, `BecomeCreatureUntilEOT` / `BecomeArtifactCreature`,
the pinned-target path of `eotSnapshot` (so `BoostUntilEOT`,
`GrantKeywordUntilEOT`, `GrantAllCreatureTypesUntilEOT`,
`RestrictUntilEOT` and every card-local caller), the list primitives
through `withoutNewSourceObject` (`TurnFaceDown`, `PhaseOut`,
`PhaseOutUntilLeaves` — which also refuses an `Until` that is a new
object — `DoesntUntapNextUntapStep`, `TapAndFreeze`,
`DoesntUntapWhile`), fight (`b10Fight`, CR 701.12b: no damage at all),
and the "for as long as ~" duration builders (`DurationWhileSourceRemains`,
`DurationWhileYouControlSource`, the hold in
`TapAndHoldWhileThisRemainsTapped`: CR 611.2b, the effect never begins).
A skipped act with a `Then` tells it `false`: nothing was sacrificed,
exiled or bounced, so "if you do" does not pay.

`b09SourceStillOnBattlefield(g, item)`, the "is this still here" guard
~20 bodies put in front of an act on the source and its payoff
(Coalition Relic's mana, Caldera Pyremaw's damage), now means the same
object. The direct-mutator bodies ask `sourceIsNewObject(g, item)`
themselves, and the two "sacrifice it. When you do / If you do" bodies
that only checked the zone (`b08OverlookSacrifice`,
`b27SacrificeSelfThenTutorColorlessCreature`) ask before the payoff.
Devoted Druid moved onto `UntapTarget`. The engine's own evoke trigger
(`queueAltCostEntryTriggerLocked`) now asks `AbilitySourceGoneForEffect`,
so an evoked creature that left and came back is not sacrificed.

**Why the primitive and not a `Self()` target.** A `Self()` helper that
carries the ref is exact — it knows the card MEANT "this" — but it only
reaches the call sites someone edits: 109 literals in 103 files, plus
the direct-mutator bodies, plus every card written after. Keying on "the
target is this item's own source" reaches all of them with no card edit,
including the ones not written yet, and the regression tests below make
a new primitive decide. **Why not the game mutators**, keyed on the
resolving slot: the slot outlives its resolution until the next event
batch (resolving_item.go), so a mana ability of the new object activated
in that window would be misjudged, and engine-internal calls made during
a resolution would be caught with it.

**The cost, stated.** The ID cannot tell "this" from "each creature you
control" when the loop happens to reach the source. A card that loops
over its own "each" set with a per-permanent primitive, whose source
left and came back in response to that very ability, skips the new
object. That is weaker than printed, never stronger, and it needs both
halves at once. Mass primitives that select with a `Match` predicate are
not asked, so they are not affected.

### Decision 18. The card-following rule: the check governs a PERMANENT that came back without this effect's help

`AbilitySourceIsNewObjectForEffect` is true only when all four hold: the
item is a stamped ability (never a spell; an unstamped pre-#1418 item
keeps today's behaviour), the target is its source card, that card is
**on the battlefield** as an object other than the stamp, and **this
resolution did not put it there**. The last two are the exceptions, and
they are one rule — *text that follows the card is not asked*:

1. **Off the battlefield, the primitive follows the card.** "When this
   dies, return it to its owner's hand", "put it on the bottom of its
   owner's library", "return this card from your graveyard", suspend's
   time counters in exile. The first two name the card in the zone it
   went to (the CR 400.7 exception for a leaves-the-battlefield
   trigger); the rest name a card that was never a permanent. None of
   them is a permanent that left and came back, and the primitive keeps
   its own rule for that zone.
2. **A move this resolution made is followed.** CR 400.7 lets the rest
   of an effect find the object the effect itself moved: unearth returns
   the card and then gives IT haste; "return this card from your
   graveyard to the battlefield, then put a +1/+1 counter on it" puts
   the counter on. `beginResolvingLocked` records the source's epoch as
   the item starts to resolve (`resolvingItem.sourceEpoch`). Nothing
   else can move a card while an item resolves, so a source whose epoch
   has changed since was moved by the effect. A source that was already
   a new object when the resolution began is the case the guard exists
   for.

A game restored from a snapshot in the middle of a paused resolution has
no resolving slot, and falls back to the stamp alone.

### Decision 19. A delayed or reflexive trigger follows a source its creator put onto the battlefield

Decision 14 made a delayed trigger (CR 603.7d) and a reflexive trigger
(CR 603.12) inherit their creator's stamp. With Decision 17 that stamp
is acted on, and unearth shows why inheriting it verbatim is wrong: the
ability is activated from the graveyard, so its stamp is the graveyard
card, and the "exile it at the beginning of the next end step" it
schedules is about the creature it has just returned. Inherited as-is,
the returned creature would be a stranger to its own delayed trigger
and never be exiled.

`followedSourceObjectLocked(item)` is the stamp, unless the resolving
item's own resolution has moved its source ONTO THE BATTLEFIELD — then
the permanent that move made. Only a battlefield arrival is followed: a
source the resolution moved anywhere else keeps its stamp, so a "when
you do" after "exile this" still reads the permanent it was through
`SourcePermanent()`. A creature that leaves and comes back before the
delayed trigger fires is, correctly, a new object to it (CR 603.7c).

### Cards

Proof, one test each (`cards/effects/source_object_primitives_test.go`),
each with its no-flicker control:

- **Bartolomé del Presidio** (`AddCounter` on self, activated): no
  counter on a Bartolomé that came back.
- **Underworld Breach** (`SacrificePermanent` on self, end-step
  trigger): a Breach that came back is not sacrificed.
- **Devoted Druid** (untap self, activated): the new Druid stays tapped.
- **Mulldrifter**, evoked (the engine's sacrifice-on-entry trigger): the
  new Mulldrifter stays, and the draw still happens.
- **Krenko, Tin Street Kingpin** ships `full`: a Krenko that is not the
  attacker any more (removed, or back as a new object) takes no counter,
  and the Goblins are its last-known power through `SourcePermanent()`.
  Its caveat — the counter landing on the card in the graveyard — was
  this issue.

Exceptions: the unearth pair (`TestDregscapeZombieUnearthsWithHaste`,
`TestUnearthedCreatureIsExiledAtTheNextEndStep`) are the regression for
Decisions 18.2 and 19, and two test-only specs pin Decision 18
(`TestAnEffectThatMovedItsOwnSourceStillFindsIt`,
`TestADiesTriggerStillActsOnTheCardInTheZoneItWentTo`).

Regression: `TestEveryPrimitiveThatActsOnACardAsksAboutItsSource` parses
the package and fails for any `Apply` on a type with a `Target`, `Card`
or `Targets` uuid field that neither asks nor delegates to a primitive
that does, unless the type is in the test's exemption table with its
reason (`DealDamage`: damage FROM the source reads last-known
information, and damage TO it is dealt by loops as often as by "this";
`ReturnFromGraveyard`, `ReturnFromExile`, `GrantFlashbackToCard`,
`PlotExiled`: the card is in another zone; `CreateTokenCopy`: copiable
values, for which last-known information is the answer).
`TestNoCardActsOnItsOwnSourceThroughAGameMutatorWithoutAsking` does the
same for a card body that calls a game mutator on `item.SourceCardID` /
`ctx.Source()` directly. It is best effort: an ID copied into a variable
first is not followed, and a spell moving itself is allowed.

### Still not covered

- ~~**The "each" loop corner**~~ (Decision 17, the cost). Closed by
  [#1463](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1463) —
  Decision 20, below.
- **Durations keyed on the source outside the builders.** A card that
  builds a `game.Duration` by hand from `ctx.Source()` instead of
  `DurationWhileSourceRemains` is not asked.
- **Damage to "this".** `DealDamage` with the source as its target is
  exempt (above), so "deals N damage to itself" is not refused on a new
  object.

## Amendment 2026-09-24 — an "each" loop reaches its source's new object (CR 400.7, CR 611.2c) · Accepted · S38

Issue [#1463](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1463),
the cost Decision 17 stated.

### What was missing

Decision 17 keys the #1432 question on "the target is this item's own
source". The ID cannot say whether the card meant "this" or "each
creature you control", so every per-permanent primitive applied in a
loop over a set the card read off the board skipped one member: the
source's new object, when the source left and came back in response to
that very ability. Per CR 400.7 that object knows nothing of the
ability, but it is a creature you control like any other, and a set
the effect reads at resolution includes it (CR 611.2c). Weaker than
printed. Measured: 28 functions read a set off the board and loop a
guarded primitive over it, among them three whose source is a member
of its own set — **Steel Overseer** ("each artifact creature you
control"), **Mazirek, Kraul Death Priest** and **Kalonian Hydra**
("each creature you control"). A fourth shape hit it the same way: a
permanent CHOSEN at resolution from the ones you control
(`ReturnOneYouControl` — **Whitemane Lion**, the bounce lands —
`UntapUpToLands`, Gluntch's counters, Teferi, Akosa's taps), where the
new object is on offer and a player who picks it saw nothing happen.

### Decision 20. The group call site says so: `ctx.asGroupMember()`; "this" stays the default

```go
func (c *Context) asGroupMember() *Context                   // effects/source_object_guard.go
func (c *Context) isNewSourceObjectAsThis(target uuid.UUID) bool
```

`asGroupMember` returns a copy of the Context with one unexported flag
set. `isNewSourceObject` — the question every guarded primitive already
asks, directly or through `withoutNewSourceObject` and `eotSnapshot` —
answers false under that flag. A loop over a set read off the board,
and a continuation acting on a permanent chosen at resolution, applies
its primitive with `ctx.asGroupMember()`. The 28 loop sites and 4 pick
sites are one-token edits; the only primitive touched is crew's
zero-`Target` path (below).

**The three shapes considered.**

1. *A `Self()` marker on the "this" call sites*, so the primitive
   guards only a target that was marked. Exact, but it flips the
   default: the 109 "this" literals, the ~20 direct-mutator bodies and
   every card written after would have to remember the marker, and one
   that forgets acts on a stranger — stronger than printed, the #259
   direction, which is exactly what #1432 closed. Decision 17 already
   rejected it for reach; it is rejected again for its failure
   direction.
2. *An opt-out field on each per-permanent primitive*
   (`AddCounter{…, AnyObject: true}`). Same default as the choice
   below, but one field on ~25 structs, every one of which must honour
   it, and a new primitive that forgets to read the field silently
   ignores the opt-out. The #1465 scan would also have to learn the
   field.
3. *Chosen: the mark rides the Context.* How the text reached the
   target is a fact about the call site, and the Context is the one
   thing the call site hands every primitive. One field, one
   constructor, one check, and every primitive — current or future —
   honours it for free, because it already asks through
   `isNewSourceObject`. The default stays guarded, so a loop that
   forgets the mark fails in the stated direction: weaker, never
   stronger.

**A copy, not a mutation.** The caller's own Context keeps asking, so
an ability that does "put a counter on each creature you control, then
sacrifice this" is judged correctly on both halves. A `Then` handed
the marked Context inherits it — it continues the act on the same
member. A continuation that rebuilds `NewContext(g, item)` later (a
queued prompt's answer) starts unmarked, which is again the safe side.

**Sites that name the source by construction ignore the mark.**
`isNewSourceObjectAsThis` is the unmarked question, and it is what the
"for as long as ~" duration builders, `PhaseOutUntilLeaves.Until`,
`TapAndHoldWhileThisRemainsTapped`, fight, Eden's and Overlook's
self-sacrifice, and crew's zero-`Target` `BecomeCreatureUntilEOT{}`
ask. Each of those is "this" whatever Context it was handed, so a
duration built inside a group loop still never begins on a stranger
(CR 611.2b).

**Chosen targets keep the default.** A target is chosen as an object;
one that left and came back is already an illegal target at
resolution (CR 608.2b, the #1429 re-check), so asking again is
redundant and harmless. Spells are never judged at all (Decision 18),
so a spell's "destroy all creatures" loop behaves identically with or
without the mark; it carries the mark anyway so the rule reads the
same everywhere.

### Cards

Proof, one test each with its no-flicker control
(`cards/effects/source_group_member_test.go`):

- **Steel Overseer** (activated): the Overseer bounced and replayed in
  response still gets its counter, as does the other artifact creature.
- **Mazirek, Kraul Death Priest** (triggered): Mazirek back in response
  still counts himself.
- **Kalonian Hydra** (attack trigger, `b08DoubleCountersOnEachCreatureYouControl`):
  a Hydra back in response is no longer attacking but is still a
  creature you control, and its counters double.
- **Whitemane Lion** (a choice at resolution, `ReturnOneYouControl`):
  the new Lion is offered, and picking it returns it.

Primitive level: `TestAPrimitiveReachedAsAGroupMemberActsOnTheSourcesNewObject`
runs every #1432 primitive case four ways — live or flickered, "this"
or group-marked — and only "this" + flickered does nothing (the zero
`Target` crew case stays guarded under the mark).
`TestWithoutNewSourceObjectKeepsAGroupMember` pins the list primitives,
the copy semantics, and a duration built under the mark.

Regression: `TestEveryEachLoopReachesItsMembersAsAGroup` parses the
package and fails for any function that reads a set off the board
(`BattlefieldCardsForEffect`, a range over `Battlefield.Cards`, or a
package function returning `[]uuid.UUID` that does either) and applies
a guarded per-permanent primitive inside a loop without saying
`asGroupMember`, unless it is in the test's exemption table with its
reason (Rakdos Charm: the loop exiles a graveyard; the board read is
mode three's damage). It shares its parse with #1465's
`TestEveryPrimitiveThatActsOnACardAsksAboutItsSource` (`scanPrimitives`),
which still holds unchanged.

### Still not covered

- **Choices at resolution are not scanned.** The four pick sites were
  found by hand; a new `ChoosePermanents` / `ChooseCardsPrompt`
  continuation that forgets the mark skips a picked new object —
  weaker, never stronger.
- **A set built without a board read the scan recognises** (a zone's
  `Cards` other than the battlefield, a helper returning `[]game.Card`)
  is not followed. Same direction.

## Amendment 2026-09-24 — a batch that commutes needs no CR 603.3b order · Accepted · S48

Issue [#1511](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1511).
Follow-ups: [#1530](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1530).
Found alongside it: [#1529](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1529).

### What was wrong

Sub-PR 8 skipped the CR 603.3b prompt for one shape only: every item in
the seat's batch had the same source and the same label (two Bident
draws). Prowess (#706, ADR 0014 amendment 2026-09-24) fires on every
noncreature spell, one trigger per instance per creature. So two prowess
creatures asked their controller to order two "+1/+1 until end of turn"
triggers on every spell, and every answer gave the same board. That
prompt is rules-correct, but it cost a click per spell for a choice with
no consequence.

### Decision

**A seat's batch drains without a prompt when every item in it is a
`StackItem.Commutes` item.** `seatNeedsTriggerOrder` (`game/mutations.go`)
now skips the prompt for either of two shapes: all identical (unchanged)
or all commutative. A commutative item counts only while it carries no
targets and no modes (`commutesForOrdering`). An auto-ordered batch
keeps its queue order, which is harvest order, and the APNAP drain
places it exactly as it would place an answered prompt.

**`Commutes` is a closed, engine-owned class. Prowess is its only
member.** The prowess trigger's `Build` (`game/prowess.go`) sets it; no
card file may. Commutativity is a property of a PAIR of effects, so a
new member needs an argument against every existing member, recorded
here, before it joins. The flag is carried by `cloneStackItem` (undo)
and by the snapshot (`commutes`, additive and omitempty: a file written
before it restores with the flag false, which is the safe direction and
just prompts again).

**Why two prowess triggers commute.** Each one:

- reads only its own source, asking "is this still the object that
  triggered, on the battlefield" (the #1432 new-object check), and no
  prowess changes that for another creature;
- writes only a layer-7c +1/+1 until end of turn, pinned to its own
  source's instance ID and battlefield-entry stamp. Layer-7c
  modifications add up whatever their timestamps (CR 613.4c), and 7d
  switching runs after all of 7c whatever the order;
- emits no event: `registerScopedStaticLocked` appends to the registry
  and bumps the layer version, and nothing watches for a P/T change or
  for an ability resolving.

So every order leaves the same board, the same effects and the same
end-of-turn expiry. The only thing that differs is the sequence of
`EventResolve` breadcrumbs in the log.

**What the skip gives up, stated honestly.** Opponents get priority
between resolutions. Order therefore still decides which creature is
already pumped when someone responds halfway through the batch, for
example Shocking the second creature before its pump resolves. That is
a sliver of strategy, not a difference in the final board. The skip
takes it away. #1530 tracks an "always ask me" per-player preference
to give it back.

### Rejected

- **"All commutative items plus at most one other trigger"** (the shape
  #1511 suggested). The other trigger's position among the pumps is a
  real choice whenever it reads what they change. Caldera Pyremaw
  granted prowess is the example: "deals damage equal to its power"
  deals one more after its own pump than before it. The engine can't
  see what an `Effect` closure reads, so it can't tell a Pyremaw from
  a harmless trigger. Even Monastery Mentor's untargeted token trigger
  doesn't provably commute with a pump: the token's entry can trigger
  an ability that resolves between the pumps and reads them. **So
  Mentor and Sokka, Tenacious Tactician boards still prompt**, since
  each is prowess plus a token trigger. #1530 records the narrowest
  way to go further. *(Closed 2026-10-11 by the #3013 amendment's
  sub-PR 3 below, through declared footprints rather than this
  class: the `Commutes` skip itself is unchanged.)*
- **"Identical-effect triggers from copies of the same card with no
  targets."** A shared oracle ID and label say the closures are the
  same code, not that they commute. An untargeted trigger that reads
  its own source and writes shared state can resolve differently per
  order. Same-source, same-label batches were already skipped by
  sub-PR 8, and that shape is unchanged.
- **A client-side "always this order" memory.** The prompt is queued
  by the server, and the client's Settings are local (`settings.ts`).
  With no server-side per-seat preference seam, a client toggle could
  only answer the prompt faster, not skip it. The preference belongs
  with #1530.

### Bot

A skipped prompt is never queued, so `legal.EnumerateFor` offers the
caster ordinary priority instead of a `trigger_order` answer.
`legal/prowess_test.go` (`TestProwessTriggersAreAutoOrderedAndPassedLikeAnyOther`)
pins it: no `resolve_choice` move, and a pass is available.

### Tests

- `game/trigger_order_commutes_test.go`: the prowess `Build` marks its
  item; a table over `seatNeedsTriggerOrder` (two or three commuting →
  no prompt; commuting plus one untargeted other, plus a targeted one,
  or plus a moded one → prompt); the real drain with three prowess
  items (drains) and with three plus a targeted trigger (held, four
  IDs); and Clone + RestoreFrom keeping the flag.
- `cards/effects/prowess_order_test.go`: two and then three prowess
  creatures cast through with no prompt and pump once each. A Mentor
  board still prompted (until #3013 sub-PR 3; its tests are now in
  `trigger_independence_prowess_test.go`). Undo, restored from before the cast and from
  mid-batch, settles without a prompt and pumps exactly once.

The targeted case is pinned at the drain rather than end to end,
because of #1529. A targeted trigger picks its target before it joins
the queue, and by then the untargeted part of its batch has already
drained, so it never gets ordered at all. That gap predates this
amendment, and this amendment doesn't change it. (Closed by the #1529
amendment below, which also pins the targeted case end to end.)

## Amendment 2026-09-24 — a trigger still announcing holds its batch (CR 603.3b / 603.3d) · Accepted · S39

Issue [#1529](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1529).
Filed while building: [#1539](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1539).

### What was wrong

A triggered ability that has something to announce doesn't reach
`PendingTriggers` when it fires. The harvest opens a prompt first:
the CR 603.5 "you may" (`trigger_prompt`), the CR 603.3c mode pick
(`mode_pick`) or the CR 603.3d target pick (`pick_target`). The item
is built from the answer. Meanwhile the next priority boundary drained
the rest of the queue, so the batch-mates that had nothing to announce
went onto the stack alone. The announcing trigger then drained by
itself, on top.

The result was that an announcing trigger always resolved first and
was never offered in a CR 603.3b ordering prompt. The issue's repro
was a prowess creature plus Caldera Pyremaw, and that board has no
choice to lose. It matters when one trigger reads what another
changes: a Pyremaw that has prowess itself deals 5 damage if it is
pumped and then burns, and 4 if it burns and then is pumped. The
engine always chose 4. The same drain also broke APNAP across seats:
a non-active player's plain trigger drained first, and the active
player's targeted trigger landed above it.

### Decision

**Targets (and modes, and the "you may") first; then the whole batch
is ordered.** `drainPendingTriggersAPNAPLocked` (`game/mutations.go`)
returns "held" without placing anything while
`triggerAnnouncementOpenLocked` reports an open prompt that carries a
trigger continuation: `triggerResume`, `modePickResume` or
`pickTargetResume`. A `pick_target` for a spell copy (`copyResume`,
CR 707.10c) builds no triggered ability, so it doesn't count. When the
last announcement is answered, the item joins `PendingTriggers`, and
the ordinary drain runs `seatNeedsTriggerOrder` over the complete
batch. The announced item is then in the `trigger_order` prompt like
any other.

**The hold covers the whole queue, not only the announcing seat.**
This is the rule the ordering prompt already follows (sub-PR 8), and
for the same reason: APNAP placement has to see every seat's complete
batch. A seat-only hold would let the non-active player's triggers go
on the stack before the active player's, which is the second bug
above.

**Every blocking prompt already stops the table, so the hold stops
nothing new.** All three announcement kinds block (`choice_gate.go`),
and every pass is already refused while one is open. So the hold
changes only which items are on `StackMeta` and which are on
`PendingTriggers` during a window in which nobody can act anyway.
Each kind has an always-legal answer in the enumerator. A seat that
leaves has its prompts dropped by the departure table.

**Every way an announcement ends now runs the boundary.** An answer
that queues the item already did so (`finishPickTargetLocked`,
`ResolveTriggerPrompt` "yes", `ResolveModePick`). The paths that end
the announcement without an item now run it too, or the held batch
would wait for an unrelated action:
- `ResolveTriggerPrompt` answered "no".
- A frame-less answer to `ResolvePickTargets` or `ResolveModePick`.
- A target walk that CR 603.3d removes at a later required clause.
- `Concede` by the chooser of the last open announcement. Here the
  drain runs only when the concession closed an announcement that was
  open before it, so a concession at any other moment is unchanged.

The CR 704.3 checks inside the loop (`refreshTargetChoicesLocked`,
`sweepEliminatedChoicesLocked`, an elimination by state-based action)
already continue the loop, and the loop drains once nothing is
announcing.

**Why choosing targets before ordering is rules-correct.** CR 603.3b
has the player put each trigger on the stack in the order they choose.
CR 603.3d chooses each trigger's targets as it is put on the stack.
Between the first target choice and the last placement, no player
receives priority, no state-based action is checked and nothing
resolves. Putting an ability on the stack changes nothing a target
clause reads except the stack itself. So the legal set computed before
the order is chosen is the set the rules would compute at each
placement, and every answer the player can give under one scheme they
can also give under the other. The one exception is under *Still not
covered*.

**#1511's commuting-batch skip is unchanged.** `seatNeedsTriggerOrder`
and `commutesForOrdering` are untouched. An all-prowess batch still
drains without a prompt. A batch of prowess plus a targeted trigger
now reaches the prompt end to end from a cast, with the targeted item
in it, where #1511 could pin it only at the drain.

**Undo mid-announcement.** `Clone` used to share the
`pickTargetFrame` pointer between the live game and the undo snapshot.
`ResolvePickTargets` advances `step` and appends to `picked` in place,
so undoing an answered pick restored a prompt whose walk was already
finished, and answering it again silently dropped the trigger. This
predates #1529. It became easy to reach once the answer to a pick was
routinely followed by an ordering prompt the player might want to take
back. `clonePickTargetFrame` (`game/clone.go`) gives the snapshot its
own cursor, picks and modes. Everything else on the frame is read-only
once the walk starts and stays shared, like every other server-only
continuation.

### Rejected

- **Order first, then choose targets as each item is placed.** This is
  the literal reading of CR 603.3b/d, and the outcome is identical
  except for the case under *Still not covered*. But no `StackItem`
  exists until the announcement is answered: `Build` runs at the end
  of the target walk, and a "you may" hasn't been asked yet. The
  ordering prompt would need placeholder items and a drain that pauses
  halfway through placement across N further prompts. A CR 603.3d
  removal in the middle of placement would have to shrink an order the
  player already gave. And both the clone and the snapshot would have
  to carry a half-placed batch. That is a new continuation shape on
  the one path every trigger in the game takes, bought for no
  observable difference on any card in the catalog.
- **Hold only the announcing seat's items.** This fixes the issue's
  single-seat repro and keeps the APNAP bug. See above.
- **Build the item before the target pick and stamp targets later.**
  `Build` is catalog code (~2,200 declarations), and some of it reads
  the board when it builds. Running it before the pick would change
  what those Builds see. It still wouldn't answer the "you may" or the
  mode pick, which decide whether there is an item at all.

### Bot

The enumerator needed no change: `trigger_order` was already
enumerated as the queue order and its reverse (`canonicalOrders`).
`legal/trigger_batch_order_test.go`
(`TestBotOrdersATargetedTriggerWithItsBatch`) pins the new position.
A bot seat answers the target prompt, is then offered only the
ordering prompt, every offered order names the targeted item, and the
engine accepts each one, putting the first-to-resolve item on top.

### Wire

Unchanged. `pending_triggers` now also holds a batch while one of its
triggers is still announcing, and `docs/protocol.md` says so. The
client already renders the queue in the stack lane.

### Tests

- `game/trigger_batch_hold_test.go`:
  - `TestAnnouncingTriggerHoldsItsBatch` covers each of the three
    announcement kinds: the plain batch-mate is held, then both are
    ordered and placed as answered.
  - Declining a "you may" releases the batch.
  - Whole-queue APNAP: the active player's targeted trigger lands
    under the non-active player's plain one.
  - A target walk that CR 603.3d removes at its second clause
    releases the batch.
  - A concession releases a batch held for the conceder's trigger.
  - The pick frame is not shared with an undo snapshot.
- `cards/effects/trigger_batch_order_test.go`:
  - The issue's prowess plus Caldera Pyremaw repro: held, one prompt
    with both items, burn ordered under the pump.
  - A Pyremaw with prowess deals 5 pump-first and 4 burn-first.
  - An illegal "target opponent" pick is still refused and the batch
    stays held.
  - Undo to the ordering prompt and to the open target prompt, each
    re-answered.
- Eight existing card tests (Phlage ×2, Fury's evoke, Satyr Enchanter
  with Verduran Enchantress, Scrapshooter's gift, Betor, the
  aristocrats and blue-draw theme decks) had relied on the old stack.
  Each now answers the new prompt with `answerTriggerOrderLastQueuedFirst`,
  which rebuilds exactly the stack the old drain produced, so their
  assertions are unchanged.

### Still not covered

- **A trigger that targets an ability on the stack (CR 115.6) can't
  target a batch-mate.** Under strict placement, a batch-mate placed
  earlier would be on the stack when a later trigger chose its
  targets. Here the whole batch is still on `PendingTriggers` when
  targets are chosen. No catalogued triggered ability targets a
  triggered ability today.
- ~~**Becomes-the-target triggers join the batch that targeted.**~~
  **Closed by [#1539](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1539)**
  (see the 2026-10-02 note below). `finishPickTargetLocked` used to
  emit `EventBecomesTarget` when the item joined the queue, so Monk
  Gyatso's or a ward trigger landed in the same drain. When it shared
  a controller with the batch, it was offered in that batch's ordering
  prompt, although CR 603.3 would put it above the whole batch. This
  predated #1529, which only widened the set of orders it could reach.
- The snapshot can't carry a trigger continuation. This is unchanged:
  the census in `snapshot.go` already reports it, and such a snapshot
  is not a full-fidelity restore point. The hold keys on the
  continuation rather than on the prompt kind, so a prompt restored
  without one doesn't hold anything: the batch drains at the next
  boundary, and answering the prompt builds nothing.

### Note 2026-10-02 — becomes-the-target triggers wait for the batch (#1539)

CR 603.3d chooses a triggered ability's targets as it is put on the
stack, and CR 603.3b puts "abilities that triggered during this
process" on the stack only after the whole batch is placed and the
state-based actions are checked again. So a trigger set off by a
batch-mate's target choice is never part of that batch.

- **The event is owed, not emitted.** `finishPickTargetLocked` sets
  `StackItem.TargetsAnnouncePending` on the item and queues it. The
  drain that places the batch (`drainPendingTriggersAPNAPLocked`)
  emits each placed item's `EventBecomesTarget` once every seat's items
  are on the stack (`announcePlacedTargetsLocked`), in placement order.
  What that harvests lands on an emptied queue, and the
  `runStateChecksLocked` loop places it as the next batch, above the
  one that targeted, whichever seat controls it.
- **Both halves of the bug go.** A same-controller watcher is no longer
  offered in its controller's ordering prompt. A ward on the active
  player's creature, targeted by a non-active player's trigger, is no
  longer placed under that trigger by APNAP order.
- **The manual announce uses the same flag.** `AnnounceTrigger` used to
  drain, emit and drain again (#974). It now sets the flag, and the
  drain emits. The result is the same, except when the drain is held,
  where the old order let the ward join the held batch.
- **Carried.** A batch held behind its ordering prompt is a restore
  point, so the flag is in `Clone` and in the snapshot
  (`targetsAnnouncePending`, additive, shape v7).
- **Spells, activated abilities and copies are unchanged.** Each one
  is already on the stack when its event is emitted. What it sets off
  goes on at the next priority boundary, with any other trigger from
  the same cast, activation or resolution, which is CR 603.3b's batch
  for them. A copy's CR 707.10c re-target prompt is a mid-resolution
  prompt (#1289), so its event is part of the resolution's batch too.

## Amendment 2026-10-01 — trigger suppression (CR 603.2) · Accepted · S50

Issue [#1735](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1735)
(the Elesh Norn item). The doubling addendum's Decision 7 left room for
this at the same point in `harvestMatchLocked`, and said why: "a
suppressor makes the count zero before the doublers are asked".

### What was missing

"Creatures entering don't cause abilities to trigger" (Torpor Orb,
Hushwing Gryff, Tocatli Honor Guard), Hushbringer's "… entering or
dying …", and the third paragraph of Elesh Norn, Mother of Machines:
"Permanents entering don't cause abilities of permanents your opponents
control to trigger." `TriggerDoubler` can only add instances. Nothing
else in the engine could stop a trigger, and neither of the obvious
neighbours fits:

- It is **not a replacement effect** (CR 614). The event still happens:
  the creature enters, the token is created, the creature dies. Nothing
  is replaced, so there is no CR 616 ordering and nothing a replacement
  could watch.
- It is **not an ability removal** (CR 613.1f). The permanents keep
  their abilities, and the same ability still triggers on any event the
  static does not name. Its own entry is what stops it, not its source.
  So `AbilitiesRemoved` and `LoseAllAbilities` don't fit either.

It is a static that changes whether an event triggers an ability at all
(CR 603.2).

### Decision

**`Spec.TriggerSuppressors []game.TriggerSuppressor`, the doubler's
sibling, asked first at the harvest.** `game/trigger_suppression.go`:

```go
type TriggerSuppressor struct {
	Label      string
	Suppresses func(g *Game, q TriggerSuppressionQuery) bool
	ActiveWhen Designation // ADR 0071, as on TriggerDoubler
}
```

`TriggerSuppressionQuery` is `TriggerDoublingQuery` with two changes:
the static's card is the `Suppressor`, and `SourceIsPermanent` replaces
`FromSpell`. The suppressor cards' wording ("abilities of permanents
your opponents control") is about the object the ability belongs to.
`harvestMatchLocked` now takes a `triggerOrigin` instead of `fromSpell`:
a permanent (the battlefield walk, `harvestLTB`, the simultaneous-exit
pass, evoke's sacrifice), a spell (`FromStack`), or neither (an
emblem's ability, a card's ability that works from another zone, CR
113.6). The doubler reads one bit of it, as before.

**The order is the composition rule.** `harvestMatchLocked` asks the
suppressors first, then the once-per-batch guard, then the doublers:

- **A suppressed ability is not doubled.** Elesh Norn doubles an ability
  that a permanent entering "causes to trigger", and this one did not
  trigger. With an Elesh Norn on each side of the table, each player's
  enters triggers are stopped by the other player's, and nothing is
  left for either to double. Nothing triggers at all.
- **A suppressed match does not spend a once-per-batch slot.** A
  "whenever one or more artifacts and/or creatures enter" under Torpor
  Orb, with a creature and then an artifact entering together, triggers
  once, for the artifact. Asking after the guard would have spent the
  slot on the creature and then dropped the artifact.
- **Nothing is made**: no instance, no "you may" prompt, no target
  prompt, no `EventTrigger`, no `TurnTally.Triggered` count. CR 603.3d
  and the prompts never see it.

**One candidate list for both.** `scanTriggerDoublersLocked` became
`scanTriggerModifiersLocked`. The per-event candidate is now a
`modifierCandidate` carrying both slices and a `live` bit. The scan is
still lazy and still runs once per event that matched anything, so a
game with no suppressor pays one length check per candidate.
`hasTriggerModifierLocked` (formerly `hasTriggerDoublerLocked`) also
catches the layers up for an entering subject when a suppressor is on
the battlefield. Torpor Orb has to see a land that enters animated as
the creature it is.

### Which board is asked

CR 603.10 answers it. This is the rule the PR chose for a suppressor
that enters or leaves at the same time as the event:

- **Away from a leaves-the-battlefield event**, the board is the one
  immediately after the event. A suppressor that enters with the
  permanent applies: Torpor Orb and a creature flickered together, and
  Elesh Norn's own entry for both of her paragraphs (the doubler
  addendum already counted that half). A suppressor that has already
  left does not apply. That includes a member of an open simultaneous
  exit that has moved to its owner's graveyard. That copy is visible
  to the doubler and is not on the post-event board, which is what
  `modifierCandidate.live` records. A suppressor that is still on the
  battlefield when the entry is announced has already stopped the
  trigger, even if it leaves later in the same resolution.
- **A leaves-the-battlefield ability looks back in time** (CR 603.10a).
  The candidates are the ones the doubler already reads for that event:
  the open batch's pre-move copies and the leaving permanent's
  last-known information. A Hushbringer that dies in the same wipe as
  the creatures still stops their dies triggers, and its own death
  triggers nothing.

### Scope

Every harvest site goes through `harvestMatchLocked`, so every
event-caused triggered ability is covered: the battlefield, emblems,
declared zones, `FromStack`, `harvestLTB`, the simultaneous-exit pass,
and evoke's sacrifice. **Evoke's sacrifice is the creature's own enters
trigger** (CR 702.74a), so under Torpor Orb an evoked creature stays,
and under an opponent's Elesh Norn so does an evoked creature of yours.
That is the printed interaction, and the card tests pin it.

**Event-conditioned delayed triggers are covered too.** Torpor Orb's
text makes no exception for a delayed ability. `fireEventDelayedTriggersLocked`
now takes the harvest pass and keeps a suppressed match queued, because
the "next time" it waits for has not come. Its source is the permanent
that created it, if that permanent is still on the battlefield (CR
603.7e). The case on develop: an earthbent land that dies under
Hushbringer does not come back.

Not covered, because no suppressor's event can cause them: reflexive
triggers (caused by a resolution), state triggers (CR 603.8: a game
state, not an event), triggered mana abilities, the manual
`AnnounceTrigger` button, and `WardSuppressions` (#1560). Ward
suppression is scoped to the warded permanent, not to an event, and
stays where it is.

The cause helpers (`SuppressesEntering`, `SuppressesDying`) use the
doubler's own `isEntering` / `isDying`. So "entering" includes token
creation (CR 701.7a) and an `EventZoneMove` onto the battlefield, and a
"leaves the battlefield" ability is stopped when the creature died,
exactly as Decision 6 counts it for the doubler. Replacement effects
("enters tapped", "enters with counters") and `AsEnters` choices are
never triggers, so they are never suppressed. That is the rulings'
"replacement effects are unaffected".

**No wire, snapshot or bot change.** A suppressed ability leaves nothing
to attribute. Nothing is stored: the answer is computed per event from
the battlefield. The enumerator sees fewer prompts of kinds it already
answers.

### Rejected

- **A negative doubler count.** It would read the suppression as "minus
  one instance", which goes wrong under two suppressors or a doubler,
  and it would still have to sit in front of the once-per-batch guard.
  That is a separate slot in all but name.
- **A check inside each card's `AppliesTo`**, the way the ward
  suppression is read. Every enters- and dies-trigger in the catalog
  would have to remember to ask, and the evoke sacrifice has no
  `AppliesTo` at all.

### Cards

Torpor Orb, Hushwing Gryff, Tocatli Honor Guard, Doorkeeper Thrull
("Artifacts and creatures entering …", `SuppressesEntering(Or(Artifact(),
Creature()))`) and Hushbringer ship `full`. Elesh Norn, Mother of
Machines loses its caveat and goes `full`. That is every card the
seam row's waiting list named for this item.
The roadmap's "Stopping abilities" row narrows to the scoped ability
removal (Dress Down, Tishana's Tidebinder), and "trigger suppression"
joins the coverage table's mechanic probes beside trigger doubling.

### Still not covered

- **A suppressor that is not about entering or dying.** No catalogued
  card needs one yet. The query already carries the event and the
  subject, so one would be a new cause helper, not an engine change.
- The doubler's `OncePerBatch` first-event limitation (Decision 4) is
  unchanged. A suppressor cannot make it worse: a suppressed first
  member no longer spends the slot.

## Amendment 2026-10-05 — "always ask me to order my triggers" · Accepted · S48

Issue [#1530](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1530),
item 2. #1511's skip leaves the board identical in any order, but it
takes away one small choice: which creature is pumped first while
opponents can respond between resolutions. A seat that wants that
choice back sets a preference.

- **A per-seat flag on the server.** `Player.TriggerOrderAlwaysAsk`,
  default false, set by the `set_trigger_order_preference` action
  (`{always_ask}`, player-scoped: a seat sets only its own, the admin any
  seat). The skip decision is the server's, so the server must hold the
  value; the client's Settings store cannot reach it.
- **`seatNeedsTriggerOrder(items, alwaysAsk)`.** With the flag set, a
  batch of two or more with any item not yet `Ordered` prompts, the
  commuting and identical skips included. An answered batch is not asked
  twice. Every other seat's batch is decided as before.
- **Persistence: server-side, reconciled by the client.** The flag is
  cloned, snapshotted (additive `triggerOrderAlwaysAsk`, shape file
  updated, no schema bump: absent restores as false) and exposed in the
  seat's own `PlayerView` only. The client does not fire on connect;
  when its own seat's view disagrees with the local setting it sends the
  action once. One rule covers a toggle, a reconnect, a second device and
  a server restored from an older snapshot, without a join-time hook.
  Re-sending on join alone would have left the server copy stale after a
  restart and sent a redundant action on every join.
- **A setting, not a play.** It mints no undo entry (`MintsNoUndo`), and
  `RestoreFrom` carries the live value across an undo of an earlier
  action, as it does `Settings`. It is allowed during the opening roll.
  The legal-move enumerator never offers it, so no bot sets it.

Item 1 of #1530 (commutativity for token triggers) is unchanged and
stays open.


## Amendment 2026-10-09 — "whenever you put one or more loyalty counters on a planeswalker" (#2797)

Inspired Tethermage's trigger is one trigger per placement, not per counter (CR 603.2c). The engine already
gives it: `EventCounterPlaced` is one event per KIND per permanent per placement, carrying the total after the
change, and loyalty is one kind. A +2 loyalty ability, a Doubling Season that makes one counter two, and a
"put three loyalty counters" are each one event and so one trigger
(`TestLoyaltyCounterTriggerFiresOncePerPlacementNotPerCounter`, `…IsOneEventWhenTheCountersAreDoubled`); a
−N cost or damage is a removal and triggers nothing.

What the trigger must NOT collapse is two planeswalkers: "put a loyalty counter on each planeswalker you
control" is one placement per permanent, and a trigger on "a planeswalker" (an object) fires for each
(`TestLoyaltyCounterTriggerFiresForEachPlaneswalker`). So it is deliberately NOT `OncePerBatch`, which is
the right tool for "one or more counters on THIS permanent" (Aragorn) and the wrong one here.

"You put" reads the event's Actor first (a loyalty cost names its activating player, CR 606.4, and so does
any effect that goes through `AddCounterBy…`), and falls back to `b12CountersPlacedBy`'s attribution when
nobody is named, which credits only a planeswalker's controller. The fallback is weaker than printed where
it is wrong (a counter you put on an opponent's planeswalker by an effect that names no placer is not
credited), never stronger. A planeswalker entering with its counters counts (CR 122.6: "put" covers an
object given counters as it enters). The card side is `effects.WheneverYouPutLoyaltyCountersOnAPlaneswalker`
(`loyalty_counter_trigger.go`).


## Amendment 2026-10-09 — three trigger-order modes, and copies of one ability need no order (#1968) · Accepted · S60

Issue [#1968](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1968). Builds on #1511's skip and #1530's
"always ask", both above. Item 1 of [#1530](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1530)
(Mentor and Sokka boards) is unchanged.

### Owner decisions (2026-10-09)

1. #1530's checkbox becomes a three-way setting, `gameplay.triggerOrder`:
   - **Ask only when the order matters** (`when_it_matters`): the default, and the old checkbox off.
   - **Always ask** (`always`): the old checkbox on.
   - **Never ask: order them for me** (`never`): no CR 603.3b prompt for this seat at all.
2. When the game orders a seat's triggers, it uses **the order they triggered in**: the order the batch was
   collected in, which is what every skipped batch has always used.
3. The default skip widens to copies of the same ability with no targets and no modes, from different source
   objects too: two Soul Wardens, or a set of tokens with the same trigger.

### Decision: the mode

- **`Player.TriggerOrder` (`game.TriggerOrderMode`) replaces `Player.TriggerOrderAlwaysAsk`.** The zero value
  is `when_it_matters`, so a seat that never chose has the default. `seatNeedsTriggerOrder(items, mode)` returns
  false for `never`, keeps #1530's rule for `always` (any batch of two or more with an item not yet `Ordered`),
  and runs the skips for `when_it_matters`. A mode this binary does not know behaves as the default. A `never`
  batch is placed by the same APNAP drain as any other, in queue order, and it still waits for other seats'
  prompts (CR 603.3b is APNAP).
- **Wire.** `set_trigger_order_preference` takes `{trigger_order: "when_it_matters" | "always" | "never"}`. The
  #1530 form `{always_ask: bool}` is still read for a cached client: `true` is `always`, `false` is
  `when_it_matters`, and `trigger_order` wins when both are sent. The seat's own `PlayerView` carries
  `trigger_order` (`"always"` or `"never"`, absent for the default) and still carries
  `trigger_order_always_ask` (true for `always`) for an older client. Both are private to the seat.
- **Snapshot, additive.** The seat gains `triggerOrder` (recorded with `-update-shape`, no bump). An `always` seat
  still writes `triggerOrderAlwaysAsk: true`, so a binary from before this change restores it as always; a
  `never` seat reads there as the default. A file from before it has only the boolean, and `true` restores as
  `always`. An unknown `triggerOrder` is kept as it is, so a roll-forward gets it back.
- **Undo** carries the live mode across `RestoreFrom`, as it carried the flag. It mints no undo entry.
- **Bots** are unchanged: they never set the mode, so they have the default and answer any prompt with their
  policy as before.
- **Client.** `gameplay.triggerOrder` (synced) replaces `gameplay.alwaysAskTriggerOrder` with no settings version
  bump, because the old key says which blob it is: a stored or synced `alwaysAskTriggerOrder: true` becomes
  `always`, anything else the default, and the old key is dropped. An account copy from an older client goes
  through the same migration (ADR 0110 §4). The reconcile (`triggerOrderPref.ts`) sends the enum, and now
  forgets a send the server refused (matched by the error frame's `replyTo`), so the next frame sends it again.

### Decision: copies of one source-blind ability

**A seat's batch drains without a prompt when every item is a copy of one source-blind catalog ability**
(`copiesOfOneSourceBlindAbility`, `game/mutations.go`):

- every item is a stamped catalog trigger (`Body` `catalog/triggered`) naming the **same `AbilityRef`**: the same
  catalog key, slot, row and label. Two Soul Wardens share their oracle ID's row; tokens from one template share
  the token's key; a granted bundle's rows share the bundle's key. Two different cards with the same words are
  two rows and still ask;
- the running catalog still hands that row back under the ref (`abilityRefMatched`), and the row is
  **`TriggeredAbility.SourceBlind`**;
- no item carries anything of its own: no targets, modes, payload, X or division.

**`SourceBlind` is engine-owned.** The catalog registry computes it as it files each definition
(`effects.fileDef` → `classifySourceBlind`, `cards/effects/source_blind.go`) and overwrites whatever a card file
set. A row is source-blind only when it has no `Build`, no target or mode clause, and its `Effect` is an
`effects.Do` whose every step is on a short list: `GainLife`, `DrawCards`, `MillCards`, `GetEnergy`,
`CreateToken`, `BecomeTheMonarch`, and `Scry` / `Surveil` with no `Then`. The registry can see the steps because
`Do`'s closure answers a probe item with them (and `Do` is `noinline`, so every Do closure shares one code
pointer the registry can recognise). Anything it cannot see, a hand-written closure above all, is not
source-blind and keeps asking. The safe direction: a row wrongly left off costs a click; a row wrongly put on
takes a real choice away.

### Why such copies commute (CR 603.3b)

CR 603.3b lets the controller put their simultaneous triggers on the stack "in any order they choose". The
order is theirs; the engine picks it only where every order gives the same game, or where they asked it to.

A triggered ability on the stack "has the text of the ability that created it, and no other characteristics"
(CR 603.3), and once triggered it exists independently of its source (CR 113.7a). Two copies of one row have the
same text and the same controller (the drain buckets by controller). What can still differ between them is
their source object, their trigger context, and anything chosen or recorded for each as it was put on the
stack. The conditions above rule out the last; a source-blind effect reads neither of the first two. So each
copy applies the same function of the game state, and resolving them in either order applies it, lets whatever
it caused go on the stack above the next copy and resolve first, then applies it again. The two orders produce
the same sequence of game states.

**Exceptions found, which keep asking:**

- **An effect that reads its source.** CR 608.2h: an effect that needs information from a specific object,
  "including the source of the ability itself", uses that object's current or last known information. Two
  sources can differ (counters, power, tapped, attachments), and one copy can change what the other reads: "Double
  the +1/+1 counters on this creature, then put a +1/+1 counter on each other creature you control" from two
  sources gives different boards in the two orders. No primitive on the list reads its source this way.
- **An effect whose source performs the action**, damage above all. A source's damage carries its own
  characteristics (lifelink, CR 702.15b; deathtouch; infect; commander damage), and state-based actions run
  between the two resolutions, so the first copy's damage can remove a creature the second would have hit, and
  which source dealt it decides who gains life. Impact Tremors is the example: two of them still ask
  (`TestTwoImpactTremorsStillAsk`).
- **An effect that reads its trigger context.** Two copies in one batch can come from different events ("you
  gain life equal to that creature's toughness"). No primitive on the list reads the event.
- **A continuation closure** (`Scry.Then`), a `Build` fill-in (per-source `Params`, payload or label), a target
  or mode clause: each can differ per copy, or cannot be read.

**What is left, stated honestly.** `GainLife` passes its source on as the gain's attribution (CR 119.9: "whenever
a source causes [a player] to gain life"), and `Scry` / `Surveil` as the prompt's. So the two orders differ in
which of two objects with the same ability is named first on two otherwise identical events. And, as with
#1511, opponents get priority between the two resolutions (CR 117.3b), so they can respond with the first copy
resolved and the second waiting; because the copies are interchangeable, which one is waiting does not change
what a response can do. **Always ask** gives the choice back to anyone who wants it.

Only the existing identical skip (same source and label) still ignores the trigger context. That shape is
unchanged.

### Rejected

- **Widen by identity alone** (same oracle ID and label, no targets or modes). That is the shape #1511 rejected,
  for the reason it gave: a shared row says the closures are the same code, not that they commute. The class
  above is identity plus a check of what the effect reads.
- **Simulate both orders on a clone and compare.** Exact in principle, but token IDs, prompts, hidden zones and
  random effects make the two results different objects even when they are the same game, and it would cost a
  clone per batch.
- **Mark source-blind rows by hand in card files.** It would rot and could be wrong silently. The registry
  derives it from what the row declares, and a card file cannot set it.

### Tests

- `game/trigger_order_mode_test.go`: each mode on two different untargeted triggers (when_it_matters and always
  ask; never does not, and keeps the collected order); never on a targeted pair; APNAP across a `never` seat and
  a default seat; always on an all-prowess batch; restore from the old boolean, the snapshot round trip, undo, the
  refusals and the parser; two copies of a source-blind row (no prompt, collected order), of a row that is not
  (prompt), under always (prompt), and with a target, mode, payload, X, another row or no stamp on one copy
  (prompt).
- `cards/effects/source_blind_test.go`: the `Do` probe; the classification table (a gain, a draw and a token,
  scry with and without `Then`, a damage step, a hand-written closure, an empty `Do`, a target clause, a `Build`);
  the registry marks Soul Warden and not Impact Tremors; two Soul Wardens see a creature enter, go on the stack
  with no prompt and gain 2; two Impact Tremors still ask; under never they do not.
- `actions/trigger_order_preference_test.go`: the enum through `Dispatch`, the old boolean, both keys.
- `protocol/trigger_order_preference_view_test.go`: each seat sees only its own mode.
- Client: `settingsTriggerOrder.test.ts` (the migration, local and synced), `triggerOrderPref.test.ts` (the
  reconcile sends the enum, reads an older server's boolean, and retries a refused send).

## Amendment 2026-10-09 — independent triggers need no order (#2884) · Accepted · S60

Issue [#2884](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2884). Builds on the #1968 amendment above, whose
three modes and copies skip are unchanged.

### Owner report and decision (2026-10-09)

Vivi Ornitier ("put a +1/+1 counter on Vivi Ornitier and it deals 1 damage to each opponent") and Ugin, Eye of the
Storms ("exile up to one target permanent that's one or more colors") triggered off one spell, and the game asked
the owner to order them under **Ask only when the order matters**. The two effects cannot touch each other unless
Ugin targets Vivi. Decision: "when it matters" asks only when one trigger's result can change what another does.

### Decision

**In `when_it_matters` mode (and for a mode this binary does not know), a seat's batch drains without a prompt when
its items are pairwise independent** (`Game.triggersIndependentLocked`, `game/trigger_independence.go`), after the
existing skips. It goes on the stack in the order it was collected, like every skipped batch, and APNAP placement is
unchanged. `always` and `never` are unchanged.

**The read and write sets are declared data.** Each catalog row carries an engine-owned
`TriggeredAbility.Footprint`, a list of `FootprintStep`s that the registry derives as it files the definition
(`effects.classifyFootprint`, `cards/effects/footprint.go`), next to `SourceBlind`. The registry reads the steps of
an `effects.Do` through the probe #1968 added, and asks each step to declare itself (`footprinter`). A row gets a
footprint only when it has no `Build`, no "you may" (`OptionalPrompt`), no mode clause, no `TargetsFrom` clause,
and every step declares one. Otherwise the footprint is nil and the row's triggers keep asking. `Purpose` and
ADR 0142's declared answers are not used: they say what an ability is for, are declared by hand, and are never read
by the rules engine. A footprint is derived from the effect the engine runs.

Three primitives were added so that rows which were hand-written closures can be declared:

- `CounterOnThis{Kind, N}`: N counters on the item's source, only while it is on the battlefield.
- `DamageEachOpponent{N}`: the source deals N damage to each opponent, in one damage event (`damageToEachOpponent`).
- `ExileChosenTarget{}`: exile the first card among the item's targets that is still legal (CR 608.2b), the body
  `b27ExileChosenTarget` now calls.

Four cards moved onto them with no change in behaviour: Vivi Ornitier, Ugin, Eye of the Storms' battlefield
trigger, Guttersnipe and Firebrand Archer. Ugin's cast trigger keeps its `Build` (it fills in the caster as
controller), so it still asks: see the question below.

### Why independent items commute (CR 603.3b, 603.3d, 608.2)

CR 603.3b lets each player put their triggered abilities on the stack "in any order they choose". As in #1511 and
#1968, the engine picks the order only where every order gives the same game.

CR 603.3d says the rest of putting a triggered ability on the stack is the casting process of 601.2c–d, so each
item's targets are chosen as it goes on the stack and are fixed before any item resolves. The check runs at the
drain, after the last target has been chosen (the #1529 hold), so it reads the targets the items will resolve with.

CR 608.2 is what resolving reads. An item checks its targets' legality (608.2b), follows its instructions in order
(608.2c), and takes any information it needs from the game or from an object when the effect is applied (608.2h).
So an item's result is a function of the part of the game state it reads at its own resolution. If nothing item A
writes is read or written by item B, and the other way round, B reads the same state whether A resolved first or
not, and each writes the same things in either order. The two orders end in the same state.

Two things happen between the resolutions, and the check covers both:

- **State-based actions** are checked before a player receives priority (CR 117.3b, 704). The ones a footprint can
  cause are a player losing at 0 or less life (CR 104.3b, 704.5a) or after drawing from an empty library
  (CR 104.3c). The check works out which players the whole batch could eliminate (its total life loss against
  their life, its total draws against their library). A player who leaves the game takes what they own with them,
  and abilities they control on the stack cease to exist (CR 800.4a). So the batch asks if its controller could be
  eliminated, and a pair asks if one item could eliminate a player the other touches (a gain of life, a library,
  a target, or an object they own or control). Two losses of life on one player commute even then.
- **Abilities that trigger during a resolution** go on the stack before the next priority (the last sentences of
  CR 603.3b), above the items still waiting, so where they land depends on the order. The check asks whenever
  anything watches an event kind the batch causes, whether or not it would fire: a triggered ability on a
  permanent or an emblem, one that works from another zone, a delayed trigger or an "until" return, a replacement
  effect (they change the event itself), and speed's inherent trigger (CR 702.179d). Damage from a source with
  lifelink (CR 702.15b), infect, wither or toxic is not on the list at all.

### What counts as a read and a write

Objects are cards and permanents by instance ID; "you" is the item's controller.

| Step | Reads | Writes | Events |
| --- | --- | --- | --- |
| `GainLife` | | your life, up | life change |
| `DamageEachOpponent` | the source (whatever the damage tail reads of it: the permanent, its last-known record once it has left, nothing once it has moved on; with no lifelink, infect, wither or toxic) | each opponent's life, down | damage, life change |
| `DamageController` | the source, as for `DamageEachOpponent` | your life, down | damage, life change |
| `TargetPlayerLosesLife` | the targeted player | that player's life, down | life change |
| `CounterOnThis` | the source | the source, if on the battlefield | counter placed |
| `ExileChosenTarget` | the targets | each card target | zone move, leaves the battlefield, unattach |
| `DrawCards` | | your library and hand (and a loss on an empty library) | draw |
| `MillCards`, `Surveil` (no `Then`) | | your library and graveyard | mill or surveil, zone move (a surveil also emits mill for the cards it puts in the graveyard: #3013 sub-PR 1) |
| `Scry` (no `Then`) | | your library | scry |
| `GetEnergy` | | nothing another step reads (energy only adds up) | player counter placed |
| `CreateToken` | | nothing another step reads (a token whose template has no text) | token created, enters, zone move |
| `BecomeTheMonarch` | | the monarch | monarch changed |

Every chosen target is a read. A pair conflicts when:

- one writes an object the other reads or writes. A tie on one object always asks, whatever the two do to it: Ugin
  targeting Vivi while Vivi puts a counter on herself;
- both touch one player's library or graveyard, or both become the monarch;
- one gains a player life and the other makes them lose it, and that player could reach 0 in between;
- one could eliminate a player the other touches (above);
- one writes an object that shapes the rules the other resolves under. That is an object with a static ability, a
  replacement effect, a trigger doubler or suppressor, or a rules-bearing slot; one whose text the engine leaves to
  the table; one that is attached or has something attached; or one named by a delayed trigger, an "until" return,
  a continuing effect's duration, or another card's record (found by walking those records for its ID, so a
  record the check has never heard of still counts).

A token step is declared only for a template whose definition has no triggered ability, no static ability or
replacement effect and no rules-bearing slot.

**The board can make the whole check give up** (`triggerOrderBoardOpaqueLocked`): a permanent or emblem with a
rules-bearing slot ("can't gain life", "damage can't be prevented", a player's hexproof, "can't lose", targeting
restrictions; `cardDefFieldClass` classifies every `CardDef` slot, and a test fails on a new one), a permanent
whose text the engine leaves to the table, a state trigger (CR 603.8 reads the game state itself), a player
static, or a continuing effect that is more than a fixed change to the characteristics of the objects it locked
in (CR 611.2c): a prevention or replacement, a "can't", a doubling, or a rule read live.

**Unknown always asks:** a hand-written closure, a `Build` (since relaxed, see the note below), a count of the board, a "may", a mode clause or chosen
modes, a `TargetsFrom` clause, an item with a payload, X, a division or `Params` of its own, and any step not in the
table.

### What is still given up

- **A static ability whose condition reads what an item writes.** The check reads the abilities of the objects an
  item writes, not the conditions of every static on the board. A static on another permanent that turns on or off
  with a life total, a counter or a count of permanents, and that changes a characteristic another item reads (the
  damage source's lifelink, the color Ugin's target needs), could make the order matter. The rules-bearing slots
  and continuing effects above are covered; layer statics on uninvolved permanents are not.
- **Priority between the resolutions.** As in #1511 and #1968, opponents get priority between the items
  (CR 117.3b) and can respond with one resolved and the other waiting.
- **Which object did it, and when.** Two damage triggers on an opponent who dies after the first: the second deals
  no damage to them, so "dealt damage by this source this turn" differs between the orders. Two token makers give
  their tokens different timestamps. The board is otherwise the same.

**Always ask** gives the choice back to anyone who wants it.

### Question for the owner

Ugin, Eye of the Storms' cast trigger has a `Build` that only fills in the caster as the controller, so casting
Ugin with Vivi out still asks. The check could accept a `Build` whose item carries nothing the footprint does not
read (no `Params`, payload, X or division). Recommended: **(a) yes**, in a follow-up, since the built item is
inspected anyway. (b) Keep every `Build` as unknown.

### Tests

- `cards/effects/trigger_independence_2884_test.go`: the registry's footprints (Vivi, both Ugin rows, Guttersnipe,
  Firebrand Archer, Soul Warden, Impact Tremors, two probes) and what it refuses; Vivi and Ugin on another
  permanent (no prompt, collected order, both resolve); Ugin targeting Vivi (asks); Guttersnipe and Firebrand Archer
  (no prompt), with an Ajani's Pridemate on the board (asks), with a lifelink source (asks); Soul Warden and Essence
  Warden (no prompt) and with a Pridemate (asks); a token maker beside a creature count (asks) and beside damage
  (no prompt); damage that can eliminate the opponent whose permanent Ugin targets (asks); always ask on an
  independent pair (asks); APNAP through a JSON restore, with the active seat's independent pair held behind the
  other seat's prompt and placed below it in collected order.
- `game/trigger_independence_test.go`: every `CardDef` slot is classified, `defBearsRules`, the ID walk, and
  undescribed items still ask.

### Note (2026-10-09): a plain `Build` is accepted

The owner answered the question above with (a). A row with a `Build` fill-in now gets the footprint of its declared
`Effect` like any other row; the footprint never comes from the `Build`. At the check, an item such a row built is
accepted only when it is exactly what the footprint describes (`builtItemIsPlain`, `game/trigger_independence.go`).
It may set only the fields the engine stamps and the footprint reads: identity, kind, controller and owner, source,
label, the triggering event, the row stamp in `Params.Ability`, and targets from the row's declared target clause.
Any other field set, such as `Params`, a payload, X, a division or modes, keeps the prompt, and so does a new
`StackItem` field until it is added to that list. Ugin, Eye of the Storms' cast trigger moved onto
`Do(ExileChosenTarget{})`, so casting Ugin with Vivi out no longer asks unless Ugin targets Vivi. Tests:
`cards/effects/trigger_independence_built_test.go`.

## Amendment 2026-10-11 — the independence check, on the boards people play (#3013) · Accepted

Issue [#3013](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3013). Builds on the #1511, #1968 and #2884
amendments above. **Accepted on 2026-10-11** with the owner's answers below. The sub-PRs at the end implement
it, and #3013 stays open until they land.

### Owner decisions (2026-10-11)

1. **The direction:** approved. Widen #2884 with the changes under *Recommendation*. Option (b), engine-owned
   commuting classes, is rejected.
2. **Migration:** all of the one-sentence backlog (about 94 cards), in batches of about 25.
3. **Watchers and opaque boards:** no change for now. The rule is revisited once sub-PR 1's counter shows how
   often each reason fires.
4. **"Use this order again"** (c2): yes, as its own issue, which the owner files.
5. **The prompt says why it asked:** yes, as its own issue, which the owner files.

### The report

The owner was asked to order two triggers whose order cannot change anything. Both fire when another Elf enters
under their controller:

- **Marwyn, the Nurturer:** "Whenever another Elf you control enters, put a +1/+1 counter on Marwyn."
- **Wolverine Riders:** "Whenever another Elf you control enters, you gain life equal to its toughness."

The counter goes on Marwyn and the life is read off the entering Elf. Every order ends with the same board and
the same life total.

### Why it still asks

The issue says the skip covers only prowess (`StackItem.Commutes`). That was true after #1511. Since then
#1968 added copies of one source-blind row and #2884 added pairwise independence from declared footprints,
which is the "declared effect footprints" direction the issue asks for. It runs in **Ask only when the order
matters**, the default. It still asks here because neither row has a footprint:

- **Marwyn's row is a hand-written closure** (`AddCounter{Target: item.SourceCardID, …}.Apply`). The registry
  can read only an `effects.Do` of primitives that declare a step, so the row has no footprint. Its effect is
  exactly `Do(CounterOnThis{Kind: "+1/+1", N: 1})`, which does have one.
- **Wolverine Riders' row is a closure too** (`b20GainLifeEqualToToughnessEffect`), behind a `Build` that stamps
  `Params.Amount` (the toughness at trigger time, kept as a fallback in case the Elf leaves). No step in the
  vocabulary can say "an amount read off the trigger's subject". Also, an item with `Params` set is never plain
  (`builtItemIsPlain`).

So the theory is not missing anything. The gap is coverage: the check can describe only a few of the rows a
real board puts in a batch.

### How common these batches are

Counted on `develop` at `ab916eef6` (2026-10-11). A throwaway test walked `effects.All()`, ran each triggered
row through `footprintOfRow`, and matched the card's oracle text from the Scryfall dump. The test was not
committed, and the figures are approximate where they rely on regexes.

| Measured | Count |
|---|---:|
| Triggered rows in the catalog | 2,792 |
| — with a footprint | 502 (18%) |
| — a hand-written closure | 1,866 (67%) |
| — "you may" / modes / a `TargetsFrom` clause / a state trigger | 175 / 40 / 30 / 50 |
| — an `effects.Do` with a step outside the vocabulary | 129 |
| Rows that watch `EventETB` | 976 (35% of all rows) |
| — with a footprint | 275 (28%) |
| Cards that print "Whenever another … enters" or "one or more other … enter" | 63 |
| — their rows: footprint / closure / may / modes | 23 / 46 / 6 / 3 |
| Cards whose every trigger line is a one-sentence shape the vocabulary has or is one step from | 230 |
| — with no hand-written closure / with at least one | 136 / 94 |

The one-sentence lines on cards that are not yet footprinted include 39 that put a +1/+1 counter on the source
(none footprinted), 26 drains ("each opponent loses N life", none: no step loses life), 43 draws, 43 tokens, 14
fixed gains and 7 "gain life equal to".

What that means at a Commander table:

- **The owner's board makes this batch all the time.** Wolverine Riders makes an Elf at the beginning of *each*
  upkeep, so at a four-seat table the prompt comes four times a round before the owner casts a single Elf, and
  once more for each Elf they cast.
- **"Another creature enters" payoffs are a staple.** A deck with two of them makes a batch of two each time a
  creature or token enters. If a batch pairs two ETB rows at random from the catalog, both have footprints
  about 8% of the time (0.28²). That's a rough figure, but it explains why the #2884 skip rarely fires.
- **Not measured: the other reasons the check asks.** Even with every row described, the check asks when
  anything at the table watches an event kind the batch causes (36 catalog rows watch a life change and 21 a
  counter placement, any player's included). It also gives up on an opaque board: a permanent whose text the
  engine leaves to the table, or a rules-bearing slot. How often each reason fires in real games is unknown,
  so sub-PR 1 counts it.

### Options

#### (a) Declared effect footprints

This is #2884's mechanism, already Accepted. Its vocabulary matches the one the issue sketches:
`CounterOnThis` is "counters on its own source", `GainLife` is the controller's life, and `CreateToken` is a
token creation. Two items conflict when one writes what the other reads or writes
(`accessesConflictLocked`). The issue's sub-questions, answered for it:

- **How the declarations stay honest.** A footprint is not declared by hand. The registry derives it from the
  effect the engine runs (`classifyFootprint`), and it overwrites anything a card file sets. So a footprint
  can't drift from its card's code. It can be wrong in two ways only:
  1. A primitive's `footprint()` misdescribes its own `Apply`. Today nothing checks this. Sub-PR 1 adds a
     conformance test.
  2. The card's code misreads the printed text. That is a card bug, and every reader of the card inherits it,
     not just this one. `TestAbilitiesMatchOracleText` and the card's own tests are the existing guards.

  Coverage is a separate question from honesty. A row left without a footprint only costs a click. Sub-PR 1
  adds a shrink-only audit over the oracle fixtures, so the uncovered list can only get shorter.
- **Events the effects emit.** Each step declares the event kinds it causes (`addFootprintStepLocked`). If
  anything at the table watches one of them, the batch asks (`somethingReactsLocked`): a triggered ability on a
  permanent or emblem, one in another zone, a delayed trigger, a replacement, or speed's inherent trigger. An
  Ajani's Pridemate on any side keeps a life-gain batch asking. A "whenever one or more +1/+1 counters are put"
  watcher, or a counter replacement such as Hardened Scales, keeps a counter batch asking. #2884's tests pin
  the Pridemate case.
- **What the skip gives up when opponents respond.** Opponents get priority between the resolutions
  (CR 117.3b). The order decides what they see when they respond:
  - **Marwyn and Riders:** whether Marwyn is already bigger, or the life already gained.
  - **Prowess and Monastery Mentor's token trigger:** with the token first, a one-damage sweep cast between
    the two kills the new Monk. With the pump first, the Monk isn't there yet.

  A skipped batch goes on in collected order. **Always ask** gives the choice back.
- **#1530's "always ask me".** It shipped (#2334), and #1968 turned it into three modes. The widening proposed
  here applies only under **Ask only when the order matters**. **Always ask** and **Never ask** don't change.
  Nothing new needs to ship alongside it.

#### (b) Narrower engine-owned classes

Each class is argued pairwise, the way #1511 argued prowess. The classes:

- **P:** prowess.
- **C:** a +1/+1 counter on its own source.
- **G:** gain N life, with N fixed when the ability triggers.
- **G′:** gain life equal to a characteristic of the trigger's subject. This is what Wolverine Riders actually
  is.

| Pair | Commute? |
|---|---|
| P × P | Yes (#1511). |
| P × C | Yes. On different objects they don't touch. On the same object, the pump and the counter both add to its power and toughness, so the result is the same. |
| P × G, C × G | Yes, provided nothing watches the counter or the life change. |
| C × C | Yes on different sources, again provided nothing watches counter placement. |
| G × G | Yes. Gains add, and a gain can't eliminate anyone. "Can't gain life" stops both in either order. |
| G′ × G′ | Yes. Each reads its own subject. |
| G′ × C | **No, when the subject is C's source.** Ghostly Flicker on Marwyn and another Elf with Riders out: Riders triggers for both Elves and Marwyn for the other one. Riders' trigger for Marwyn gains 1 if it resolves before Marwyn's counter and 2 after. |
| G′ × P | Same rule as G′ × C. In practice it's rare, because a cast and an entry are seldom one batch. |

Why it is rejected:

1. **The owner's example isn't in class G.** Riders reads the Elf's toughness when its ability resolves (CR
   608.2h). Fixing it at trigger time would change the card: an opponent's −2/−2 on the Elf in response would be
   ignored.
2. **G′ needs a check per instance, not per class.** It commutes with C except when the subject is C's source.
   Telling those cases apart means comparing object IDs, and that is (a).
3. **Every pair says "provided nothing watches".** So each class would still need the reactor, opacity and
   elimination checks that #2884 already runs. The classes would rebuild (a), cover less, and need a new
   recorded argument for every member.

#### (c) Other directions found

- **c1. Reaction-aware.** Two ways a watcher could stop forcing the prompt:
  - It would not fire on the events this batch makes (an opponent's Pridemate watches *its controller's*
    gains).
  - Its own trigger has a footprint, and that footprint is independent of the rest of the batch.

  The first needs `AppliesTo` asked about events that have not happened yet, against a board the batch is
  changing. The second needs the check to recurse. Both are worth doing only if sub-PR 1 shows reactors are a
  common reason. **Deferred.**
- **c2. "Use this order again."** A per-seat standing answer, keyed by the batch's multiset of `AbilityRef`s
  and applied to the next batch with the same rows. This doesn't rely on the triggers commuting: the order is
  the player's own choice, which CR 603.3b leaves to them. It also covers batches where order *does* matter
  (Caldera Pyremaw with prowess). It's orthogonal to this amendment, and gets an issue if the owner wants it.
- **c3. Simulate both orders on a clone.** Rejected by #1968 (token IDs, prompts, hidden zones, randomness, and
  a clone per batch). Nothing has changed since.

### Recommendation

**(a): no new mechanism. Widen #2884's reach** with three vocabulary additions, one hardening fix and a
migration. The rules the check applies don't change.

1. **An amount read off the trigger's subject.** A step `GainLifeEqualToSubject{Stat}` (toughness first; the
   power variant comes when a card needs it).
   - It reads the subject through `PermanentForEffect` (Decision 13): the live permanent while it's on the
     battlefield, its last-known information once it has left. So it needs no `Build` and no `Params.Amount`.
   - Its footprint is `FootprintGainLife` plus a new `AmountFromSubject` flag. The check turns the flag into a
     read of `Trigger.Event.CardID` while that object is on the battlefield. Last-known information is frozen,
     so a subject that has left is not a read.
2. **Life loss for each opponent.** A step `EachOpponentLosesLife{N}`, the drain. It writes each opponent's life
   down, adds N to `lifeLoss` for the elimination check, and emits a life change. It doesn't name a source as
   the cause of the loss, so it has no lifelink question.
3. **A footprint for the engine's keyword triggers, starting with prowess.** The prowess trigger reads and writes
   its source (a +1/+1 until end of turn) and emits nothing.
   - The footprint comes from a table in `game` keyed by stack body (`"prowess/pump"`). `triggerAccessLocked`
     looks there for a keyed item, as it looks up the catalog row for a stamped one.
   - The `Commutes` skip stays as it is: it reads no board, so an all-prowess batch keeps skipping on an opaque
     board.
   - Monastery Mentor's and Sokka, Tenacious Tactician's token rows already have footprints. This closes
     #1530 item 1.
4. **Hardening.** `tokenTemplateInert` reads only a template's `Triggered`. A template whose keywords make
   keyword triggers that watch an event the batch causes would slip through (evolve watches entering). No
   catalogued token has evolve today. The check should refuse such a template anyway, with a test.
5. **Migration.** Move one-sentence closures onto `Do` primitives that already declare a step, starting with
   Marwyn (`Do(CounterOnThis{"+1/+1", 1})`) and the three cards on the `b20` body (Trostani, Selesnya's Voice;
   Verdant Sun's Avatar; Wolverine Riders). The 94 cards in the census are the backlog.

**The owner's pair under the existing rules, once 1 and 5 land:**

- **Marwyn** reads and writes Marwyn, and emits a counter placement.
- **Riders** reads the entering Elf, writes its controller's life upward, and emits a life change.

They conflict only when the Elf *is* Marwyn, which is the Ghostly Flicker case above, and the check correctly
asks then. Otherwise the batch goes on with no prompt, unless something at the table watches counters or life
or the board is opaque.

### What is still given up

The #2884 list is unchanged: statics whose condition reads what an item writes, priority between resolutions,
and "which object did it, and when". For this amendment's new cases:

- **Opponents respond between resolutions** (see the Mentor example above).
- **A life gain can flip a threshold static during the window.** For example, "as long as you have 30 or more
  life" can turn on between the two resolutions. The final state is the same.

**Always ask** gives the choice back.

### Snapshot, undo, wire and bot

- **Footprints are catalog data.** The registry derives them at registration, and game state never holds one.
  Prowess's footprint is an engine table. So there's no snapshot field, no `Clone` change, no shape bump and no
  new route in `closure_fields.txt`.
- **A migrated row keeps its label and its slot**, so a queued or stacked item still restores onto the same row
  (ADR 0041 P9). A Riders item that an older binary queued carries `Params.Amount`:
  - The new effect still accepts it, using it as the fallback.
  - `builtItemIsPlain` refuses it, so that batch asks. That's the safe direction.
- **Undo** restores the state from before the drain, and the drain decides again from the same inputs.
- **Wire and bot are unchanged.** A skipped prompt is never queued, as #1511 pinned for the bot. The optional
  "why it asked" line below would be an additive wire field.

### Sub-PRs

1. **Guards and measurement. No behaviour change.**
   - `cards/effects/footprint_conformance_test.go` covers every type that implements `footprinter`; a new type
     without a case fails. Each one resolves on a fixture board with an event recorder and a before-and-after
     diff of players and objects. The test fails if a step changes anything, or emits an event kind, outside
     what `addFootprintStepLocked` declares for it.
   - `cards/coverage/trigger_footprint_audit_test.go` reads `testdata/oracle/`, the Scryfall projection that CI
     already holds and `TestOracleFixtureIsCurrent` refreshes nightly against the dump.
     - It lists the catalogued cards that have a one-sentence trigger line in the vocabulary's shapes and a row
       with no footprint.
     - It compares that list with `testdata/unfootprinted_trigger_lines.txt`, which may only shrink, like
       `TestLegacyTriggerBuildsOnlyShrink`. The update flag never adds a line.
   - **Measuring the reasons.** `triggersIndependentLocked` returns why it asked: an undescribed row, a
     reactor, an opaque board, an object tie, a library, an elimination or the monarch. A closed-label counter
     records it (`cmdctrl_trigger_order_prompts_total{reason}`, ADR 0123).
2. **The owner's example.**
   - The changes: `GainLifeEqualToSubject` and `AmountFromSubject`; Marwyn moves onto `CounterOnThis`; the three
     `b20` cards move onto the new step.
   - New tests:
     - Riders' upkeep Elf and a cast Elf go on with no prompt.
     - A Pridemate at the table asks.
     - Marwyn and another Elf entering together under Riders asks.
     - An Elf that left before resolution gives its last-known toughness.
     - An older item carrying `Params.Amount` resolves as before and asks.
     - Always ask and never ask behave as before.
   - The audit list shrinks.
3. **Prowess.**
   - The changes: the keyed-body footprint table, and the `tokenTemplateInert` keyword fix.
   - New tests:
     - A Mentor board and a Sokka board cast a noncreature spell with no prompt.
     - A Mentor board with an ETB watcher asks.
     - Prowess plus Caldera Pyremaw still asks (Pyremaw is a closure, and described, it would tie on the
       source).
     - An evolve token template is refused.
   - This closes #1530, since its item 2 shipped already.
4. **Then: the drain step and the migrations.** `EachOpponentLosesLife`, then the census backlog by shape in
   batches of about 25 cards (counters on the source, drains, draws, tokens, fixed gains). Card tests stay
   unchanged and each PR shrinks the audit list. These are `tier:2` card work.

c1 and c2 are not on this list. Each is decided on sub-PR 1's data, or filed as its own issue.

### Questions for the owner

Answered on 2026-10-11: see *Owner decisions* at the top of this amendment.

1. **The direction.** Recommended: (a), widening #2884 with the four changes above, and (b) rejected.
2. **How much to migrate.** (a) The whole one-sentence backlog (94 cards). (b) Only cards in the decks the table
   plays (the curated decks and saved libraries), with the rest as they're touched. Recommended: **(a)**, since
   each batch is mechanical and the audit pins the list.
3. **Reactors and opaque boards.** An opponent's Pridemate, or one permanent the engine leaves to the table,
   keeps a batch asking. (a) Keep that until sub-PR 1's counter shows how often each reason fires, then decide
   on c1 and on narrowing the opacity rule. (b) Start c1 now. Recommended: **(a)**.
4. **"Use this order again"** (c2): file it as its own issue? Recommended: yes, as a `tier:3-design`.
5. **Say why the prompt appeared.** Should the `trigger_order` prompt show the reason, for example "Ajani's
   Pridemate reacts to life gain"? That's additive on the wire, plus a client line. Recommended: yes, after
   sub-PR 1, since it uses the same reason value.

### Note (2026-10-11): sub-PR 1, guards and measurement

- **Reasons.** `triggerOrderAskLocked` returns why a seat is asked (`game.TriggerOrderReason`): `always`,
  `undescribed`, `board_opaque`, `reactor`, `ripple`, `elimination`, `object_tie`, `rules`, `library` or
  `monarch`. The reason is the first conflict the check finds, and it never changes whether the seat is
  asked. The drain counts each prompt once, as it is queued, in `cmdctrl_trigger_order_prompts_total{reason}`
  (ADR 0123). A held drain passes the boundary again until the prompt is answered, and those passes don't
  count again.
- **Conformance.** `cards/effects/footprint_conformance_test.go` resolves every primitive that declares a step,
  on a real board. It fails on any player, object or event kind the step touches outside its declaration
  (`game.FootprintDeclarationForTest`, the check's own reading). A new primitive without a case fails. The
  test found one gap on its first run. A surveil's graveyard leg emits `EventMill` per card, as an ordinary
  mill does, and the step did not declare it. So a watcher of milling alone (Colossal Grave-Reaver) did not
  keep a surveil batch asking. The step now declares it, which only adds prompts.
- **Coverage.** `cards/coverage/trigger_footprint_audit_test.go` reads the oracle fixtures. It lists a
  catalogued card when it prints more one-sentence trigger lines of a shape the vocabulary has, or is one
  primitive away from, than it has rows with a footprint. Its shapes are wider than the census above
  (energy, the monarch, damage to each opponent, and any count written as a digit), so it lists 157 cards
  where the census found 94. The list is `cards/coverage/testdata/unfootprinted_trigger_lines.txt` and only
  shrinks. Sub-PRs 2 and 4 work through it.

### Note (2026-10-11): sub-PR 2, the owner's example

- **The step.** `GainLifeEqualToSubject{Stat: SubjectToughness}` reads the trigger's subject through
  `ctx.TriggeringPermanent` (Decision 13). That is the permanent as it is now while the object is on the
  battlefield, and as it last existed there once it has left. An item that a binary from before this change
  queued carries `Params.Amount`, which is used only when the subject has no record. Such an item is never
  plain, so its batch asks.
  - Its footprint is `FootprintGainLife` with `AmountFromSubject`. The check counts the subject as read while
    it is on the battlefield.
  - A zero or negative toughness gains nothing. The old body passed a negative toughness to the life change,
    which made it a loss.
- **The cards.**
  - Wolverine Riders, Trostani, Selesnya's Voice, Verdant Sun's Avatar and Righteous Valkyrie use the step.
    The `b20` body and Valkyrie's copy of it are gone.
  - Marwyn, the Nurturer is `Do(CounterOnThis{"+1/+1", 1})`.
  - Every row keeps its label, so a queued item restores onto the same row.
  - Five lines leave the audit list.
- **A rule the proposal did not have: a read through the layers.** An amount equal to a creature's toughness is
  the output of every static that applies to that creature. Another item in the batch can change what such a
  static reads: a counter, a creature entering, a life total for Righteous Valkyrie's "+2/+2 while you have 7
  more life". So the order can show although no object is shared. Example: Valkyrie and Trostani both gain
  for one Angel, and the first gain switches on the anthem that the second gain reads. So a subject read
  (`triggerAccess.charReads`) conflicts with any other item that changes game state when the board has a
  static that can change power or toughness: layer 7, a copy, or an ability removal (`layerStaticsLocked`).
  - The one exception: an item that only moves life totals passes unless one of those statics declares
    `DependsOnLifeTotal`. That is the layer engine's own contract for what such a static reads.
  - The reason is `characteristic`.
  - The cost: an Elf lord such as Elvish Archdruid keeps the Marwyn and Riders batch asking, although its
    output cannot change there. The check can't see what a static reads. If sub-PR 1's counter shows
    `characteristic` is common, a static that declares it reads nothing the batch writes is the next step.
- **Still not covered (unchanged).** Two triggers from one source with one label are skipped by sub-PR 8's
  identical rule, which has never read the trigger context. So two Righteous Valkyrie gains for two creatures
  entering together are auto-ordered, even across her +7 line.

### Note (2026-10-11): sub-PR 3, prowess

- **The footprint.** `keyedBodyFootprints` (`game/trigger_independence.go`) declares the footprint of an engine
  keyword trigger by its stack body. Prowess (`"prowess/pump"`) is its only entry: one `FootprintPumpSource`
  step. It reads its source, writes it while it is on the battlefield, and emits nothing.
  - `triggerAccessLocked` looks a keyed item up there, as it looks a stamped item's row up in the catalog. It
    accepts the item only when it is exactly what the prowess `Build` makes: no targets, modes, `Params`,
    payload, X or division, and no field outside `builtItemPlainFields`. `Commutes` joined that list, since it
    says how a batch may be ordered and changes nothing the item does.
  - The `Commutes` skip is unchanged and runs first, so an all-prowess batch still skips on an opaque board.
  - `game/trigger_footprint_keyed_test.go` is the table's conformance test. It resolves each body's real item
    and fails on any event, any player or object change, or any continuing effect outside the declaration. A
    body without a case fails.
- **The boards.** A Monastery Mentor board and a Sokka, Tenacious Tactician board cast a noncreature spell
  with no prompt, and so does the next spell, with the new token's prowess in the batch. #1530 item 1 is
  closed.
- **Two rules the proposal did not have.**
  1. **The printed-keyword static is not a rule the object shapes.** The registry turns `PrintedKeywords` into
     one self-only layer-6 static, so every catalog creature with a printed keyword had a `Static` slot, and
     `objectInertLocked` called it rules-bearing. Writing Monastery Mentor (printed prowess) therefore set
     `rules`. Such a row applies the object's own printed keywords to the object and reads nothing, so
     changing the object can't change what it does to anything else. The registry now marks that row
     (`StaticAbility.PrintedKeywordRow`, engine-owned: a card file's mark is cleared), and `defBearsRules`
     ignores it. This also applies to `CounterOnThis` and `ExileChosenTarget` writes on such creatures, which
     now pass where they used to ask under `rules`.
  2. **A pump reaches past its object only through what the layers compute.** A pump changes only its
     object's power and toughness, never its zone, counters or attachments. So a static on the pumped object
     can change only what the layers compute for other objects. Sokka's grants are such statics. A pump on an
     object that bears rules therefore sets `pumpRules`, not `rules`. `pumpRules` conflicts only with an item
     that reads through the layers: an amount off the subject (`charReads`), a damage source's lifelink,
     infect, wither or toxic, or a chosen target's legality (`traitReads`). The reason is `rules`. So Sokka's
     pump beside Firebrand Archer asks, and Mentor's doesn't.
     - Given up: a static on the pumped object that reads its power and *grants a triggered ability* to
       something, which would then watch the rest of the batch. This is the #2884 "static whose condition reads
       what an item writes" case, extended from other permanents' statics to the pumped object's own.
- **Hardening: a token's keyword triggers.** `tokenTemplateInert` read only a template's catalog `Triggered`.
  A token step now carries the template's keywords (`FootprintStep.TokenKeywords`), and the check reads the
  keyword triggers they make (`tokenKeywordWatches`, through `keywordTriggersFor`). If one of them watches an
  event kind the batch causes, the batch asks with the reason `reactor`. An evolve token beside another token
  maker asks. A Monk with prowess doesn't, because prowess watches a cast and no step casts. No catalogued
  token has evolve today. A keyword a static would *grant* the token once it enters is not read, which is the
  same given-up case as above.
- **Unchanged, by the owner's decisions.** Watchers still force the prompt. Ty Lee, Chi Blocker's own "when
  this enters" keeps Sokka's batch asking (`reactor`), as an ETB watcher such as Soul Warden keeps Mentor's.
  **Always ask** and **Never ask** behave as before. Caldera Pyremaw beside prowess still asks
  (`undescribed`).
- **Tests.** `cards/effects/trigger_independence_prowess_test.go` covers the keyed footprint, the Mentor and
  Sokka boards over two spells, the ETB watcher, Ty Lee, Pyremaw, Sokka's pump beside damage and Mentor's,
  always ask, the evolve and Monk tokens, and the registry's printed-keyword mark. In `game`:
  `trigger_footprint_keyed_test.go` (keyed conformance, `tokenKeywordWatches`, `defBearsRules` on the
  printed-keyword row). `prowess_order_test.go`'s "Mentor still prompts" test is gone.

### Note (2026-10-11): sub-PR 4, the drain and the first migrations

- **The step.** `EachOpponentLosesLife{N}` is "each opponent loses N life": the item's source is passed to the life
  change as its attribution (as `GainLife` does) and nothing else reads it.
  - Its footprint, `FootprintLoseLifeEachOpponent`, reads no object. It writes every opponent still in the game
    down, adds N to their `lifeLoss` for the elimination check, and emits a life change. It is not damage, so it
    has no lifelink, infect, wither or toxic question.
  - A drain is `Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})`. The two steps run in that order, as the
    hand-written bodies did. The step is not on the source-blind list, since a replacement of life changes could
    read the attribution.
  - `footprint_conformance_test.go` has its case.
- **Batch 1, the drains.** 24 cards move onto it with their labels and trigger conditions unchanged: Ayara, First of
  Locthwain; Bastion of Remembrance; Bontu's Monument; Cauldron Familiar; Cauldron of Essence; Cliffhaven Vampire;
  Consuming Sepulcher; Corpse Knight; Cruel Celebrant; Elas il-Kor, Sadistic Pilgrim; Funeral Room; Gixian
  Puppeteer; Kardur, Doomscourge; Marauding Blight-Priest; Marionette Apprentice; Mirkwood Bats; Nadier's Nightblade;
  Obyra, Dreaming Duelist; Plague Belcher; Poison-Tip Archer; Psychosis Crawler; Sanctum Seeker; Wayward Servant;
  Zulaport Cutthroat.
- **Left on the list.** Indulging Patrician and Lunar Convocation check "you gained life this turn" as they resolve
  (an intervening "if"), which no step expresses yet.
- **Batch 2, counters on the source.** 33 cards move onto `Do(CounterOnThis{Kind: "+1/+1", N: 1})` with their
  labels and conditions unchanged: Ajani's Pridemate; Animar, Soul of Elements; Brineborn Cutthroat; Champion of
  Lambholt; Champion of the Perished; Chandra's Emberling; Chasm Skulker; Danitha, Spear of Agony; Feral Ghoul;
  Frenzied Raider; Honored Dreyleader (its second row); Idol of False Gods; Inspired Tethermage; Juri, Master of
  the Revue; K'rrik, Son of Yawgmoth; Laelia, the Blade Reforged; Managorger Hydra; Massacre Girl, Most Wanted;
  Murktide Regent; Nesting Dovehawk; Orca, Siege Demon; Ravenous Squirrel; Stalwarts of Osgiliath; Toothy,
  Imaginary Friend; Unflinching Hortimancer; Vadrik, Astral Archmage; Valley Mightcaller; Vengeful Warchief;
  Vincent, Vengeful Atoner; Wildwood Scourge; Yahenni, Undying Partisan; Yoshimaru, Ever Faithful; Youthful
  Valkyrie.
  - **One difference.** Some of the old closures (Animar, Brineborn Cutthroat, Chandra's Emberling, Chasm Skulker,
    Danitha, Frenzied Raider, Inspired Tethermage, Massacre Girl, Nesting Dovehawk, Toothy, Valley Mightcaller,
    Yahenni, Yoshimaru) had no "still on the battlefield" check, so a source that had gone to the graveyard in
    response still received the counter. `CounterOnThis` puts nothing on a source that has left, as the printed
    "this creature" does (CR 400.7), so this is a fix, not a loss.
  - **Left on the list.** Edgar, Moonlit Sovereign (re-checks "you cast no spells this turn" as it resolves),
    Mausoleum Harpy (re-checks the city's blessing) and Runaway Steam-Kin (re-checks fewer than three counters)
    all have an intervening "if" that no step expresses. Rapacious Guest, Surly Badgersaur and Titanbones,
    Towering Heart have a second shape on the same card and are in a later batch.
- **Batch 3, damage to each opponent.** 23 cards move onto `Do(DamageEachOpponent{N})`, which already had a
  footprint, so no new step: Agate Instigator; Black Waltz No. 3; Brimstone Vandal; Coruscation Mage;
  Electrostatic Field; Erebor Flamesmith; Fiery Inscription; Firespitter Whelp; General Kreat, the Boltbringer;
  Glint-Horn Buccaneer; Impact Tremors; Kessig Flamebreather; Lambholt Raconteur and Lambholt Ravager;
  Molten Gatekeeper; Purphoros, God of the Forge; Reckless Fireweaver; Refraction Elemental; Syr Konrad, the Grim;
  Tinybones, Pocket Nuisance; Unstable Amulet; Warleader's Call; Weftstalker Ardent. Labels and conditions are
  unchanged.
  - **A test changed meaning.** `TestTwoImpactTremorsStillAsk` pinned that two Impact Tremors ask, because their
    row was a closure and is not source-blind. With a footprint each reads only its own source and lowers the
    opponents' life, which commutes, so the batch now goes on with no prompt
    (`TestTwoImpactTremorsSkipThePromptByTheirFootprint`). The registry test's Impact Tremors case now expects the
    damage footprint.
  - **Left on the list.** Lux Artillery: its 10 damage re-checks "30 or more counters" as it resolves
    (an intervening "if").
- **Batch 4, token makers.** 17 cards move onto `Do(CreateToken{...})`, `Do(CounterOnThis{...})` or `Do(MillCards{...})`
  with labels and conditions unchanged: Arahbo, the First Fang; Autarch Mammoth; Beetleback Chief; Bridled Bighorn;
  Cosmium Kiln; Desolation Twin; Elanor Gardner; Gilded Ghoda; Lossarnach Captain; On Wings of Gold; Open the
  Graves; Rapacious Dragon; Restless Anchorage; Sidisi, Brood Tyrant; and the counter rows of Rapacious Guest, Surly
  Badgersaur and Titanbones, Towering Heart, which finish those cards.
  - **No new footprint.** `CreateToken` already declares what a token entering does: it emits token created, enters
    and zone move, so any permanent on the battlefield that watches entering (an ETB watcher, a replacement that
    watches the event) keeps the batch asking with the reason `reactor`. A token whose template has a trigger,
    static, replacement or rules-bearing slot is refused, and an evolve-style keyword trigger on it is read too.
    `trigger_independence_tokens_test.go` pins a Zombie beside a drain skipping, and the same pair asking beside
    Impact Tremors.
  - **Copies of one maker.** Two sources with the same label (two Open the Graves) were already auto-ordered by the
    source-blind rule (#1968), which `CreateToken` is on the list for. That rule never read the board, so a migrated
    token maker joins it. It needs no order because the copies make the same token.
  - **Left on the list.** Tapped tokens (`CreateTokenAdvanced` has no footprint step): Alchemist's Talent, Diregraf
    Colossus, Don Andres, Nuka-Cola Vending Machine, Overseer of the Damned, Tormod, the Desecrator. Resolution-time
    re-checks (intervening "if"): Crested Sunmare, Field of the Dead, Keeper of the Accord, Ginger, Queen of Sweets,
    Faramir, Field Commander (wraps its draw). Counted or X amounts: Arboreal Alliance, Cadira, Caller of the Small.
- **Batch 5, draws.** 11 cards move onto `Do(DrawCards{N})`, with labels and conditions unchanged: Eternal of Harsh
  Truths; Frostbite Pyromental; Guidelight Matrix; High-Society Hunter; Ichor Wellspring; Ingenious Infiltrator;
  Llanowar Visionary; Seafloor Oracle; Vanquisher's Banner; Vedalken Archmage; Way of the Mind Sculptor.
  - **No new footprint.** `FootprintDraw` already marks the controller's library as touched (so a draw conflicts with
    a scry, a surveil, a mill or another draw of the same player, reason `library`), counts the draws for the empty
    library check, and emits a draw. A "whenever you draw" watcher on the board keeps the batch asking (`reactor`).
    `trigger_independence_draws_test.go` pins a draw beside a drain skipping, a draw beside a scry asking
    (`library`), and a draw beside Chasm Skulker asking (`reactor`).
  - **Eternal of Harsh Truths** builds its row with `WhenAttacksAndIsNotBlockedEffect`, whose callback gets the
    defending player. This payoff does not use it, so the row's `Effect` is replaced with the `Do` after the
    constructor builds it.
  - **Left on the list (intervening "if", re-checked as the ability resolves):** Agent of Treachery (three
    permanents you don't own), Beastbond Outcaster, Colossal Majesty and Garruk's Uprising (a creature with power 4
    or greater), Bugenhagen, Wise Elder (power 7), Deadeye Brawler (the city's blessing), Padeem, Consul of
    Innovation (the greatest artifact), Satoru, the Infiltrator (nothing cast with mana in the batch), The Gaffer
    (3 life gained this turn).
- **Batch 6, the rest that can be expressed.** 20 cards, and three steps' reach widened. Each change has a
  conformance case (`footprint_conformance_test.go` now keys a variant as `<Type>:<variant>`) and a skip and an
  ask test (`trigger_independence_final_test.go`).
  - **Gain life** (`Do(GainLife)`): Bishop of Wings (both rows, its Spirit too); Distinguished Conjurer; Glowing One;
    Interplanar Beacon; Lifecreed Duo; Merchant Ship (the row's `Effect` is replaced after its constructor builds it,
    as for Eternal of Harsh Truths); Nyx-Fleece Ram.
  - **Mill, scry, surveil:** Cephalid Illusionist; Hermes, Overseer of Elpis (scry and Bird); Lazav, the
    Multifarious; Restless Spire; Thassa, God of the Sea.
  - **Energy:** Decoction Module (`Do(GetEnergy)`). The old body passed no placer to the counter replacement window;
    `GetEnergy` names the controller, as CR 122.1 says the player getting energy does. Only a replacement that reads
    the placer can tell, and any such replacement on the board already makes the batch ask.
  - **Power variant.** `GainLifeEqualToSubject{Stat: SubjectPower}` is the same step as the toughness one (the
    footprint doesn't name the stat), for Conclave Mentor's death trigger. The subject is the Mentor itself, read as it
    last existed.
  - **Tapped tokens.** `CreateTokenAdvanced` declares the `FootprintCreateToken` step when its spec only sets
    `EntersTapped` (no counters, no granted keywords, the controller's own). Entering tapped changes only the new
    token's own state and emits no event, which the conformance case confirms. Alchemist's Talent, Diregraf Colossus,
    Don Andres (already a `Do`), Nuka-Cola Vending Machine, Overseer of the Damned and Tormod, the Desecrator use it.
    Boards with a permanent that watches an entering object, a zone move or a token still ask: Tormod beside On Wings
    of Gold asks (both watch zone moves), as do Overseer, Talent and Colossus (their own entry rows or replacement).

### Closing note (2026-10-11): what stays on the ratchet, and why

The audit's list holds 24 cards. Each is an exception, grouped here by the reason no footprint is declared. The
list is the testable record of this table, and `TestOneSentenceTriggersDeclareFootprints` keeps it from growing.
Every one of these keeps the CR 603.3b prompt when its trigger is in a batch, which is the safe direction.

**Intervening "if" re-checked as the ability resolves (CR 603.4).** The row's effect first reads game state and does
nothing if the condition is false, so its footprint would have to say that it reads that state. The check has no way
to say so today, and the effect can't be a bare `Do` without losing the re-check.

| Card | The condition re-checked at resolution |
| --- | --- |
| Agent of Treachery | Three or more permanents you control but don't own (end step draw) |
| Beastbond Outcaster | You control a creature with power 4 or greater |
| Colossal Majesty | You control a creature with power 4 or greater |
| Garruk's Uprising | You control a creature with power 4 or greater (its entry row; the other row is a `Do`) |
| Bugenhagen, Wise Elder | You control a creature with power 7 or greater |
| Crested Sunmare | You gained life this turn |
| The Gaffer | You gained 3 or more life this turn |
| Indulging Patrician | You gained 3 or more life this turn |
| Lunar Convocation | You gained life this turn (the loss row: and lost life) |
| Deadeye Brawler | You have the city's blessing |
| Mausoleum Harpy | You have the city's blessing |
| Edgar, Moonlit Sovereign | You cast no spells this turn |
| Field of the Dead | You control seven or more lands with different names |
| Keeper of the Accord | An opponent controls more creatures than you |
| Ginger, Queen of Sweets | You are the monarch |
| Inventors' Fair | You control three or more artifacts |
| Lux Artillery | Thirty or more counters among artifacts and creatures you control |
| Padeem, Consul of Innovation | You control the artifact with the greatest mana value |
| Runaway Steam-Kin | This creature has fewer than three +1/+1 counters |
| Satoru, the Infiltrator | Nothing entering this batch was cast with mana |
| Faramir, Field Commander | A creature died under your control this turn (the draw is wrapped in that check) |

What a step for these would need: a `Do` step that carries a condition, `If{Cond, Then...}`, where each `Cond` is a
named value that declares the state it reads (life gained this turn, a creature's power, the monarch, a cast tally,
a counter total). The footprint of the step would be the `Then` footprint plus a read of that state, and the check
would conflict it with any item that writes the state. The unit of work is one condition at a time.

**An amount that is counted or chosen.**

| Card | The amount |
| --- | --- |
| Arboreal Alliance | X, taken from the cast (a `Build` fills `Params.Amount`) |
| Cadira, Caller of the Small | One Rabbit for each token you control, counted as it resolves |

What a step would need: for X, a footprint that allows `Params.Amount` set by the engine's own cast record, which
`builtItemIsPlain` refuses today; for a count of the board, a step that declares the set it counts (tokens you
control) as a read, so a token maker beside it conflicts.

**The subject is not the event's object.**

| Card | The subject |
| --- | --- |
| Ikra Shidiqi, the Usurper | The creature that dealt combat damage. `EventDealDamage` names it in `Event.Source`, and `TriggerContext.Object` only snapshots `Event.CardID`, so `GainLifeEqualToSubject` can't find it and last-known toughness would be lost once it has left. |

What a step would need: the trigger context to snapshot the damage source as its object (a `TriggerContext` change,
so a snapshot field and a closure-fields entry), after which `GainLifeEqualToSubject{SubjectToughness}` fits.

## Amendment 2026-10-12 — an identical batch is asked again when its copies read their subjects (#3021) · Accepted

Issue [#3021](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3021). Found by #3020 (the #3013 amendment's
sub-PR 2), which recorded it as "still not covered".

### The question

Sub-PR 8 skips the CR 603.3b order prompt when every item in a seat's batch has one source card and one label. Its
argument is that such items are interchangeable. That holds when the copies differ only in their source and
read nothing that differs between them. It does not hold when each copy reads its own triggering subject.
Two Righteous Valkyrie triggers, for two Angels that enter together, are one source and one label. Each gains life
equal to a different creature's toughness. Valkyrie's anthem (+2/+2 while you have 7 more life than you started
with) is a layer-7 static that reads the life total, so the first gain can cross the line and change what the
second one reads. The order shows, and sub-PR 8 skipped it without looking.

### Decision

**Yes. The same-source, same-label skip also passes the characteristic test of the #3013 amendment when the items read
their triggering subjects.** Nothing else of the independence check is added to it.

- `Game.identicalBatchOrderShowsLocked` (`game/trigger_independence.go`) runs only for a batch the pure
  `seatNeedsTriggerOrder` has just skipped. It asks (reason `characteristic`) when all of these hold:
  - the seat's mode is "when it matters" (an unknown mode behaves as that one; "never" and "always" are untouched),
  - at least one item is unanswered and every item is identical by the sub-PR 8 test (`identicalItems`),
  - every item is one the footprint check can describe, and at least one reads a characteristic of its subject
    (`triggerAccess.charReads`, set by `GainLifeEqualToSubject`),
  - for some pair, `characteristicReadConflict` holds: one copy changes something a static on the board could read
    (any change for a layer-7, copy or ability-removal static, and only a life total when the only static
    declares `DependsOnLifeTotal`; an item that only moves life passes the first kind).
  `characteristicReadConflict` is the test #3020 added, lifted out of `accessesConflictLocked` so both callers share
  one definition.
- **The narrowing is the point.** The full check is not run on an identical batch. It would add object ties (two
  counters on one permanent), reactors and the rest to boards that have never asked, for items that really are
  interchangeable. An item the check cannot describe keeps the old skip, for the same reason. So the prompt
  appears only for a copy that reads a different creature's toughness beside a static that can move it:
  Righteous Valkyrie's own batch. Trostani, Verdant Sun's Avatar and Wolverine Riders beside no such static,
  or beside a lord that does not read life (Elvish Archdruid), stay unasked, and so do two Bident draws, two
  prowess pumps and two Annihilator triggers.
- No wire or client change. The prompt is the existing `trigger_order`, and the enumerator (`legal/choices.go`) and
  the heuristic (`aiseat/heuristic/choices.go`) already answer it for any batch. It is counted under the existing
  `characteristic` metric label.
- Where the old rule is stated above (the sub-PR 2 note's "Still not covered", the #1968 amendment's "Only the
  existing identical skip still ignores the trigger context") it was true on 2026-10-11 and is superseded here for
  the subject-reading case only.

### Tests

`cards/effects/trigger_independence_3021_test.go`: two Valkyrie gains for two Angels ask and are counted under
`characteristic`; the bot is offered both orders and the engine accepts one; "never" does not ask; an answered batch
is not asked again; Trostani alone and Riders beside Elvish Archdruid do not ask. The unchanged-skip cases are
`game/trigger_order_test.go` (identical draws) and `cards/effects/annihilator_test.go`.


## Amendment 2026-10-13 — the order prompt says why it asked (#3015) · Accepted

Issue [#3015](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3015). The owner chose it on 2026-10-11 (decision 5 of
the #3013 amendment), to build after sub-PR 1's reason counter. #3021's `characteristic` case (above) is covered too.

### The question

The CR 603.3b "Order your triggers" prompt says only that. A player who is asked for an order that looks irrelevant
(two life gains, a counter) cannot tell whether the game found a real conflict or gave up on a card it cannot read.
The engine already knows: `triggerOrderAskLocked` returns a `TriggerOrderReason`, and the drain counts it. The prompt
should say it.

### Decision

**Yes. Each prompt carries one sentence, built on the server from the reason, in the player's terms.**

- `PendingChoice.TriggerOrderWhy` (`game/pending_choice.go`) holds the sentence. It rides the snapshot
  (`pendingChoices[].triggerOrderWhy`, `omitempty`, an additive change that zero-values to "no sentence") and the
  wire as `trigger_order_why` on the `trigger_order` pending choice (`docs/protocol.md`). The client shows it under
  the heading. Server-side because the prompt's other text (`reason`, the labels) is built there, and because the
  facts a sentence needs (the card behind a static, the event a watcher watches) exist only where the check runs.
- **The check carries the facts it already found.** `triggerOrderConflictLocked` and `accessesConflictLocked` return a
  `triggerOrderWhy` (the reason plus the card, the event, the life direction or the player the sentence names)
  instead of the bare reason. `somethingReactsLocked` became `reactorLocked`, which returns the card and event kind
  it matched; `triggerOrderBoardOpaqueLocked` returns the first card that made the board opaque;
  `characteristicReadConflict` returns the card behind the static it found. The yes/no decision and the reason are
  unchanged (`triggerOrderAskLocked` is now a wrapper over `triggerOrderAskWhyLocked`), and the metric counts the
  same labels. Where the old code took the first of a map, it now takes the first in a fixed order, so the sentence
  is stable; no decision depends on which.
- **The words** (`game/trigger_order_why.go`, `triggerOrderSentenceLocked`). Every `TriggerOrderReason` has one, plus
  a generic fallback for a reason this binary has no case for. None uses the check's vocabulary (a test fails on
  "footprint", "characteristic", "batch", "engine" and the reason labels).

  | Reason | Sentence |
  |---|---|
  | `always` | You've set your triggers to ask every time. |
  | `undescribed` | *Wolverine Riders'* trigger is handled by the players, so the game can't tell whether the order matters. |
  | `board_opaque` | *Humility* could change how these triggers resolve. |
  | `reactor` | *Ajani's Pridemate* watches life totals. Named by the family of event the card declares it watches (life totals, damage, cards being drawn, permanents entering, and so on), never by a direction. A token the items create is "A token these triggers create watches …". |
  | `ripple` | Something else in play is keeping track of *Oblivion Ring*. |
  | `elimination` | You (or the player's name) could leave the game partway through these triggers, which changes what the rest do. |
  | `object_tie` | These triggers both change *Marwyn*. (One writes and one reads: "One of these triggers changes *Marwyn*, which another looks at.") |
  | `rules` | *Sokka's* trigger changes the rules the other triggers resolve under. |
  | `library` | These triggers both work on the same library or graveyard. |
  | `monarch` | These triggers both make you the monarch. |
  | `characteristic` | The life gain from one trigger can change *Righteous Valkyrie's* bonus, which another trigger counts. For an identical batch (#3021): "The first life gain can change …". |

  A legend is named by the part of its name before the comma ("Marwyn", "Sokka"); a face-down permanent is
  "a face-down permanent", because every seat reads the sentence (CR 708.2). A reason with no card to name (a player
  static, a delayed trigger) gets the same sentence with "Something in play".
- **What the sentence is not.** Display text. It is not read back, not part of the resolve payload, and not shown to
  the enumerator or a bot: `legal/choices.go` and the heuristic read `TriggerOrderIDs` only, and the #3021 test that
  offers a bot both orders runs against a prompt that now carries the sentence. A `reactor` sentence names the family of event the
  card watches and never a direction in it. A card's declared watch is an event kind (`change_life`), which does not
  say whether it cares about a gain or a loss, and the check cannot evaluate the card's own condition (which is why
  it asked). "Reacts to life gain" would be false for a card that watches life loss when the batch only gains, so the
  sentence says "watches life totals" for every watcher of the kind. If a watcher's direction is ever declared as
  data, the specific wording can return for it alone; until then the family wording is used everywhere. (The
  `characteristic` sentence does say "life gain" or "life loss": that is what the item itself does, which is known.)
- **A prompt queued before this change** has no sentence and shows the old heading alone.

### Tests

`game/trigger_order_why_test.go`: every reason in `TriggerOrderReasons()` has a case with the exact sentence, the
banned-word check, the fallbacks, the sentence stored on a queued prompt (and unchanged by the held drain), the
legend short name. `cards/effects/trigger_order_why_3015_test.go`: the real boards each reason can reach (reactor three
ways, undescribed, object tie, characteristic two ways, library, elimination, rules, always) produce their sentence.
`protocol/trigger_order_why_view_test.go`: the field rides the wire and is omitted when empty or on another kind.
`ChoicePromptModal` render test in `client/src/lib/choiceSheets.render.test.ts`.

## Amendment 2026-10-13 — villainous choice (#3140, CR 701.55) · Accepted · Backlog

Issue [#3140](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3140). `option_pick` (decision 6 above) owns "choose one of the following" addressed to a seat that is not the controller, so a villainous choice is an option pick and not a new kind. Numbering: no ADR was reserved; this amends the ADR that owns the prompt.

### The rules

From the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`):

- **CR 701.55a:** "[A player] faces a villainous choice — [option A], or [option B]" means "[A player] chooses [option A] or [option B]. Then all actions in the chosen option are performed."
- **CR 701.55b:** "While facing a villainous choice, a player may choose an option that is illegal or impossible. In that case, they perform as much of the action as is possible. This is an exception to rule 608.2d."
- **CR 701.55c:** "A replacement effect may replace an instruction to face a villainous choice with an instruction to face that choice some number of additional times. In that case, the entire process described in rule 701.55a is performed for that player the appropriate number of times one at a time."
- **CR 701.55d:** "If more than one player is instructed to face a villainous choice, the entire process described in rule 701.55a is performed for each of those players one at a time in APNAP order. This is an exception to rule 608.2e."

### Decisions

1. **A registered pair, faced through an option pick.** `effects.NewVillainousChoice(key, question, A, B)` registers one keyed option-pick continuation (`option-pick/villainous-<key>`, ledger line `option <key>`) holding the two bodies. `effects.FaceVillainousChoice{Players, Choice, Carry}` queues the question for the first player, addressed to them and not to the controller. Each option is a `VillainousOption`: the printed label, a `Body(ctx, Facing)` that captures nothing, and three hints for a bot. Nothing is a closure on the prompt, so a table waiting on the answer is a restore point (the prompt carries `pickThen` and `optionCarry`, both already on the wire since #2854).
2. **Both options are always offered (CR 701.55b).** The queue prunes nothing and the enumerator already offers every option of an option pick, so a player with no creature may pick "sacrifice a creature" and sacrifice nothing. This is the general rule 608.2d's exception, and the bodies do as much as they can. `VillainousOption.Possible` does not hide an option: it zeroes the bot hints of one that would do nothing to this player, so a bot takes the free one.
3. **One at a time, in APNAP order (CR 701.55c, d).** The prompt's carry is the players still to face the choice, `uuid.Nil`, then the card's own data. `FaceVillainousChoice` lists the players in APNAP order, each repeated once per time they face it, and each answer's continuation performs the chosen body and then asks the next entry. A player who has left the game is skipped (CR 800.4a). No count is stored: the list is the order.
4. **"An additional time" is a count read when the choice is faced.** `Spec.OpponentsFaceVillainousChoicesAgain` (The Valeyard) is read through `CatalogExtraVillainousChoice` and `CatalogAbilityKey`, as `ExtraVote` is: `Game.VillainousChoiceTimesForEffect(facing)` is one plus one for each such permanent under the control of someone other than the facing player. Two Valeyards make three questions, which is what two replacement effects give (CR 614.5 stops one effect applying to an event twice, not a second effect applying to the modified instruction). The count is fixed when the instruction is performed.
5. **The hints.** `ChoiceOption.Harm` (additive, omitted when 0; wire `pick_options[].harm`) is what an option costs the chooser besides life and mana, in cards, and `LifeCost` is the life. The engine never reads either. The heuristic (`aiseat/heuristic/villainous_choice.go`) takes the cheapest loss, prices life through `LifeCostValue`, and refuses an option that costs its last life. A prompt whose options carry no harm keeps the enumerator's first answer. The hints may not read a hidden zone: the controller's hand is never read into one.
6. **A follow-up question is not sequenced.** A body that asks one (a sacrifice, a discard) leaves it open beside the next question in the chain. The engine has no "after every prompt this body opened" seam; the effect is the same as the rules' because each answer stands alone, and the open question is a limit to lift if a card ever depends on the order.

### The cards

The Valeyard (its caveat is gone: Full), Damocles Base, Sword of Kang, The Dalek Emperor, This Is How It Ends, Sycorax Commander, Dr. Eggman, Davros, Dalek Creator and Great Intelligence's Plan, all Full. The Dalek token is a new `tokens_table.go` row. Six of the fourteen on #3140 are left for a card slice, with what each needs: Missy (a creature card put onto the battlefield from a graveyard face down and tapped with a listed Cyberman body; the library put is the only one that carries a listing today), Hunted by The Family (a no-duration "becomes a 1/1 white Human creature and loses all abilities" on up to four targets, each faced by its own controller), Genesis of the Daleks (the total power of Daleks that died this turn), The Master, Gallifrey's End (choosing the opponent with the most life, with the tie going to the controller), Midnight Crusader Shuttle (a creature put into an attack already declared, attacking that player) and Ensnared by the Mara (exile until a nonland card from an opponent's library, then a free cast from exile).

### Consequences

Additive in the wire (`harm`), the snapshot (`ChoiceOption.Harm` in `pickOptions` and `councilVote.options`, recorded in `testdata/snapshot_shape/v8.txt`; a binary before it drops the key and restores the same question) and the effect-key ledger (seven `option` lines). An older binary refuses a restore point with a villainous question open, as for any keyed option pick it lacks the key for. No client change beyond the `harm?` field on the type: the prompt is an ordinary option pick.

**Back-outs.** Each shape has a test that fails without it: `TestAVillainousChoiceIsAskedOfTheFacingPlayerNotTheController` (the addressee), `TestAnImpossibleOptionIsStillOnOfferAndCostsNothing` (both options offered), `TestTheValeyardMakesAnOpponentFaceTheChoiceAnAdditionalTime` and `TestTheAdditionalTimeIsCountedWhenTheChoiceIsFaced` (the count), `TestSeveralPlayersFaceItOneAtATimeInTurnOrder` and `TestEachPlayersAdditionalTimesAreAskedBeforeTheNextPlayer` (APNAP and one at a time), `TestAVillainousChoiceSurvivesCloneAndARestoreAndCarriesOn` (the keyed continuation), and the heuristic's `TestTheBotTakesTheCheaperOfTwoBadOptions`.

## Amendment 2026-10-13 (second) — the six remaining villainous-choice cards (#3143) · Accepted · Backlog

Issue [#3143](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3143). The six cards the first amendment left each needed one piece besides the keyword action. Numbering: no ADR was reserved; this amends the ADR that owns the prompt. All six ship Full.

### Decisions

1. **One choice per target is a chain.** Hunted by The Family has a villainous choice for each of up to four targets, each faced by its own controller, and CR 701.55d wants them in APNAP order, one at a time. The card sorts the targets by their controllers' seat order and carries them; each option body starts the next target's question once its own last question is answered. `Facing.Remaining` (new, the number of questions this `FaceVillainousChoice` has still to ask, counting only players still in the game) is the signal: a Valeyard's additional times (CR 701.55c) repeat the same target and must not start the next one more than once. Later steps re-check each target by the clause (a creature on the battlefield the caster does not control), not by the stack item's target list, because the spell has left the stack by then. The choice is built from `init`, since its bodies read it again.
2. **A tally of the power that died, by creature type.** `TurnTally.DiedPowerBySubtype` is the table-wide total power of creatures that went from the battlefield to a graveyard this turn, per creature type they had as they died (a changeling under one `*` bucket, as `EnteredSubtypes` does), read through `Game.DiedPowerWithSubtypeThisTurn`. Power is the departure record's (layers and counters, CR 608.2h), types are the event's (CR 603.10a). A tally and not a graveyard walk, because a token that died is gone by the time anyone asks (CR 704.5d). Additive within the current snapshot schema (`diedPowerBySubtype`, omitted when empty), carried by `cloneTurnTally`. Genesis of the Daleks reads it after its own sweep, so the Daleks that option destroys are among those that died.
3. **Putting a creature that is already in play into combat.** `Game.PutOnBattlefieldAttackingForEffect(id, defender)` is the second door onto `stampEntryAttackerLocked`, the CR 506.3c rule ninjutsu and attacking tokens use: the creature attacks without having been declared, so no "whenever ~ attacks" trigger fires. It runs the layer pass first, so a control change registered by the same effect is applied before the stamp and cannot take the creature out of combat afterwards (CR 506.4). Midnight Crusader Shuttle gains control until end of turn, checks the control actually changed, taps the creature and calls it.
4. **A tied "most life" opponent.** `effects.OpponentsWithTheMostLife` lists them; one is the answer and several go to `ChoosePlayer{Only: tied}` for the controller to break the tie. `ChoosePlayer.Only` is new: a pool computed from the board, in place of `Among`.
5. **A dies trigger that acts on "it".** `eventCardTriggerItem` and `deadCardInGraveyard` (`dead_card_trigger.go`) record the card and its object epoch when the trigger fires and find it again when it resolves, as Ares' own copy of the shape does; The Master (exile it, then the choice) and Missy (return it face down) use them.
6. **Nothing else was needed.** Missy's graveyard put was already `ReturnFromGraveyardFaceDownForEffect` (the issue was written before it was seen); Hunted by The Family's no-duration effect is `ScopedEffectFor` with `IndefiniteDuration` pinned to the object and the existing type, colour, ability and base-P/T mods; Ensnared by the Mara is `MillToZone` with an `Until` and the free-cast grant Emergent Ultimatum uses. "Chaos ensues" (Missy) is Planechase's planar die: it does something only with a planar deck, which a Commander game on this engine does not have, so the clause has nothing to perform.

### The cards

The Master, Gallifrey's End; Genesis of the Daleks; Hunted by The Family; Midnight Crusader Shuttle; Ensnared by the Mara; Missy. All Full.

### Consequences

One additive field in the snapshot (`turnTally.diedPowerBySubtype`), six effect-key `option` lines, no wire change.

**Back-outs.** Each shape has a test that fails without it: `TestDiedPowerTotalsTheDalekPowerThatDiedThisTurn` and `TestDiedPowerIsThePowerItHadWithCounters` (the tally and its LKI power), `TestPutIntoCombatSurvivesTheControlChangeBeingApplied` (the layer pass first), `TestHuntedByTheFamilyWithTheValeyardAsksEachControllerTwiceAndChainsOnce` (`Facing.Remaining`), `TestHuntedByTheFamilyAsksEachControllerInTurnOrderAndSkipsATargetThatLeft` (the chain and the re-check), `TestTheMasterAsksTheControllerToBreakATieAmongTheTiedOnly` (`ChoosePlayer.Only`).

## Amendment 2026-10-13 (third) — choices made as the ability resolves, not as it goes on the stack (#3172) · Accepted · Backlog

Issue [#3172](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3172), slice "cav-a, choices made at the right time". Twenty-four catalog cards carried a caveat that came down to one of three things: a choice settled when the ability went on the stack (as its target) instead of as it resolved, a choice replaced by a fixed rule, or a condition read once instead of twice. Numbering: no ADR was reserved; this amends the ADR that owns triggers on the stack. All twenty-four ship Full.

### Decisions

1. **A resolution-time "may sacrifice" is one helper.** `maySacrificeOne` (`may_sacrifice.go`) is `ChoosePermanents` with a floor of zero over the controller's own permanents, then the ordinary simultaneous sacrifice, and it runs the card's `then` only for a permanent that really left (`SacrificeAllThenForEffect`'s landed list, not the pick). Ruthless Technomancer, Disciple of Freyalise, Chitterspitter, The Goose Mother and Ziatora use it. The creature's power is its last-known information (`departedCreaturePower`).
2. **A keep-set is asked once per seat, then the rest are sacrificed together.** `keepCreaturesThenSacrificeTheRest` (`keep_creatures_sacrifice_rest.go`) runs `ChooseCardsRunThenForEffect` with one leg per seat in APNAP order, `most` as the ceiling and a set rule (`totalPowerAtMost`) as `Validate`, then sacrifices everything not kept in one batch so a Blood Artist sees every death. It replaces Archfiend of Depravity's "sacrifice one at a time until two remain" and Slaughter the Strong's lowest-power auto-pick.
3. **"Any number of counters" is a number per kind.** `chooseCounters` (`choose_counters.go`) asks a `ChooseNumber` for each kind of counter on the permanent, in sorted order, bounded by what it has and what is left of the ceiling. Glissa Sunslayer removes up to three; Resourceful Defense moves any number. No new prompt kind.
4. **"Discard two unless you discard a <kind> card" is a set rule.** `discardTwoUnlessOne` registers a draw-then continuation whose discard prompt has a floor of one and a `Validate` that accepts a single card only when it qualifies. A hand of one is discarded whole, an empty hand asks nothing (CR 701.9a).
5. **Resolution-time re-checks of an intervening "if"** (CR 603.4) for Thopter Spy Network and Emeria, the Sky Ruin: the second read sits at the top of the `Effect`.
6. **Cards that needed no engine change.** Expel the Interlopers asks its number with `ChooseNumber` (0 to 10, a bot suggestion that trades their creatures against yours); Mount Doom, Breena, Teval, Blossoming Tortoise, Extract from Darkness, Command Beacon, Erebos, Chain of Smog and Tireless Provisioner use `QueueChooseCardsForEffect`, `ChoosePermanents`, `MayChoice`, `PickOption` and the discard prompt as they stand. Blossoming Tortoise's discount is `ActivationCostsLess` over lands you control. Breena's per-turn tally was removed: the batch guard already resets with each step, so an additional combat triggers her again.
7. **Three caveats were already false.** Buried Alive asks "up to three" at any library size (a search that names a quality may find fewer, #3053), Ziatora's damage already counted an anthem (the departure record holds the post-layer power), and Resourceful Defense already triggers for a token that leaves. Each now has a test.

### Consequences

No wire change, no snapshot change, no new prompt kind, no effect-key ledger line. Prompts a closure holds (`maySacrificeOne`'s `then`, `chooseCounters`' `then`) make the table unrestorable while they are open, as every `Then`-closure prompt already does (the Volcano Hellion posture).

**Back-outs.** Each shape has a test that fails without it: `TestRuthlessTechnomancerPaysNothingForACancelledSacrifice` (the landed list, not the pick), `TestB22SlaughterTheStrongKeepsLowestPowerWithinFour` (the set rule), `TestB18ThirstForKnowledgeAcceptsOneArtifactOrTwoCards` (the unless rule), `TestCavAEmeriaRechecksThePlainsOnResolution` and `TestCavAThopterSpyNetworkRechecksTheArtifactOnResolution` (the re-checks), `TestCavAGlissaChoosesHowManyCountersOfEachKind` (the ceiling), `TestB17BreenaTriggersAgainInAnAdditionalCombat` (no per-turn tally), `TestB31BlossomingTortoiseDiscountsLandActivations` (the discount).

## Amendment 2026-10-13 — Graveyard watchers use the actual exit set (#3417) — Accepted

`Event.Batch` spans a whole resolution and cannot tell whether two battlefield
exits were simultaneous. Nether Traitor must see a creature that dies in a
later instruction after the Traitor has entered the graveyard, including when
both instructions belong to one resolving spell.

Expose `Game.InSameSimultaneousExitForEffect(a, b)` as a read-only query over
the already published simultaneous-exit copies. A graveyard watcher is
excluded only when both it and the qualifying exiting creature are members of
that exit. Checking both IDs matters: an outer batch can remain published
while its continuation starts a later instruction. Existing graveyard scope,
last-known owner/type checks, resolution-time payment and object-epoch guards
remain the authoritative checks for their respective questions.

This adds no event or persisted state and introduces no second batching
system. Tests distinguish sequential instructions inside one resolution from
an actual simultaneous wipe/sacrifice, and retain the existing recursion
regressions. Clear Nether Traitor's ordering caveat only with those tests.

## Amendment 2026-10-14 — a source that died, damage to you, and the targeted drain · Accepted

Found by replaying eight four-seat `boteval arena` games (esper-control, mono-black-aristocrats, simic-ramp,
wubrg-dragons): 81 ordering prompts in 32 seat-games, nearly all "X's trigger is handled by the players". Three
gaps in the footprint vocabulary covered most of them.

1. **A damage step reads what the damage tail will read.** `DamageEachOpponent` refused whenever its source had
   left the battlefield, which is the aristocrats case: Syr Konrad dies in the wipe whose deaths he triggers on, and
   as the deck's commander goes on to the command zone. `footprintDamageSourceLocked` now builds the tail the step
   will build (`effectDamageTailLocked`) and refuses only lifelink, infect, wither or toxic on it, so the two
   cannot disagree: the permanent while it is on the battlefield, its last-known record once it has left (CR
   608.2h), no traits once the card has moved on and no record matches. The source stays a read, because an item
   that moves the card back changes which object deals the damage.
2. **`DamageController{N}`** is "~ deals N damage to you" (City of Brass, Midnight Reaper): the same source
   questions, and the controller's life, down.
3. **`TargetPlayerLosesLife{N}`** is "target player loses N life", the first half of Blood Artist's and Falkenrath
   Noble's drain (`Do(TargetPlayerLosesLife{N: 1}, GainLife{Amount: 1})`). It reads and lowers the targeted
   player's life. A target who could reach 0 in the batch keeps the prompt through the existing elimination rule,
   since a later drain then has no legal target.
4. **Two draws commute.** A library touched by two items asked whatever they did to it, so Midnight Reaper's three
   "draw a card" for one wipe asked among themselves. Two items that only draw from the top of one library put the
   same cards in the same hand in either order. The library tie (reason `library`) now asks only when one side
   mills, scries or surveils that library. A draw that could empty it is still the elimination rule's (the draws are
   counted), and a "whenever you draw" watcher still the reactor rule's.

Tatyova, Vindictive Vampire and the Pest token move onto `Do` with steps that already existed. Moving the Vampire
showed it saw its own death, against its oracle's "another creature you control"; its condition now excludes
itself (`TestVindictiveVampireDoesNotSeeItsOwnDeath`). No wire or snapshot
change. Each new step has a case in `footprint_conformance_test.go` (`DamageEachOpponent:departed` and
`DamageController:departed` resolve with the source in the graveyard, `DamageEachOpponent:moved-on` with it exiled
from there).

**Back-outs.** `TestSyrKonradDyingInTheWipeHeTriggersOnDoesNotAsk` and
`TestSyrKonradMovedOnFromTheGraveyardDoesNotAsk` fail without item 1,
`TestADepartedLifelinkSourceStillAsks` without its lifelink refusal, `TestTwoTargetedDrainsDoNotAsk` without item 3,
`TestTwoTargetedDrainsAskWhenTheTargetCanDie` if the drain stopped being a player read, and the Syr Konrad test
again without item 4 (Reaper's three draws). The draw-beside-scry case in `trigger_independence_draws_test.go`
keeps item 4 from going further.

## Amendment 2026-10-14 (second) — a tapland in play is not a reactor · Accepted

A second arena replay (mono-green-stompy, mono-red-aggro, mono-white-weenie, boros-voltron; seeds 101–108) and the
first round's games both had prompts whose reason was a land "watching cards changing zones": Myriad Landscape,
Shipwreck Marsh. Their only replacement is "This land enters tapped" (`SelfEntersTapped` and its conditional
siblings). It watches zone moves, so the reactor rule counted it for every batch that made a token or moved a
card, as long as the land was on the battlefield.

1. **`EntryOnly` replacements on a permanent in play are set aside** (`reactorLocked`). Such a replacement changes
   only its own object's entry, which a permanent in play has already made, and no described footprint step puts
   that object onto the battlefield again. `ReplacementEffect.EntryOnly` was display-only. It is now a promise the
   check relies on, so `TestEntryOnlyReplacementsApplyOnlyToTheirOwnEntry` probes every catalog row that declares
   it with another card entering, and fails if one applies.
2. **Aurelia, the Law Above's** five-attacker drain moves onto `Do(DamageEachOpponent{N: 3}, GainLife{Amount: 3})`.
   Beside her three-attacker draw it no longer asks (`TestB16AureliaFiresOncePerDeclarationAtThreeAndFive` now pins
   that). Wearing Batterskull, her damage has lifelink and the batch still asks, as it should.

Second round, same seeds: 16 → 13 prompts. In the first round's games, the Shipwreck Marsh batches now stop on the
real reactor beside it (Smothering Tithe watches draws) instead.

**Back-outs.** `TestATaplandInPlayIsNotAReactor` fails without item 1; `TestAureliasTriggersDeclareFootprints`
without item 2.


## Amendment 2026-10-14 (third) — a paid branch for counter-unless-pays, and a per-controller tally of what died (#3591, #3592) · Accepted

Two small gaps found clearing the caveats of the incubate cards (Assimilate Essence, Brimaz, Blight of Oreskos).

1. **`CounterUnlessPaidPrompt.OnPaid` (#3591).** "Counter target spell unless its controller pays {4}. If they do, you incubate 2." The prompt underneath (`queuePayUnlessLocked`) already carried a paid branch (`payUnlessExtras.onPay`, echo and may-pay use it); `QueueCounterUnlessPaidForEffect` never passed one. It now passes `OnPaid`, so exactly one of the two branches runs for an answer: a payment that was made runs `OnPaid`, a decline or a "pay" the chooser cannot fund runs the decline. The stack guard is the one that was already there: `answerPayUnlessLocked` drops both branches when the guarded object has left the stack, which is right for the paid half too, since the payment then bought off nothing. On the card side `effects.CounterUnlessPaid.OnPaid` is a `*Context` branch bound to the resolving spell, so `Controller()` is the caster ("you"), not the payer. Nil changes nothing for the other counter-unless-pays cards.
2. **`TurnTally.DiedSubtypes` (#3592).** "If a Phyrexian died under your control this turn." The tally had a per-controller count of creatures that died and a table-wide power by type; this is the per-controller count by creature type, with the same last-known types (CR 603.10a), the same "under the controller it left from" (`LeftUnderControlOf`), the same changeling bucket (`enteredAllCreatureTypes`) and the same keys as `EnteredSubtypes`. Read through `Game.DiedWithSubtypeThisTurn(player, subtype)`. It shares `diedSubtypesLocked` with `recordDiedPowerLocked`, so the two tallies cannot read different types for one death. Additive in snapshot schema v8 (`turnTally.diedSubtypes`, omitted when empty, carried by `cloneTurnTally`, reset by `resetTurnTallyLocked`).

**Back-outs.** `TestACounterUnlessPaidRunsOnPaidOnlyForAPaymentThatWasMade` and `TestAssimilateEssenceIncubatesTwoForTheCasterWhenTheTaxIsPaid` fail without the `onPay` pass-through; `TestDiedSubtypesCountsPerControllerAndPerType` (and Brimaz's `TestBrimazBlightProliferatesAtTheEndStepAfterAPhyrexianDiedUnderYou`) fail without the record call; `TestDiedSubtypesSurvivesCloneTurnTally` without the clone line.

## Amendment 2026-10-14 (fourth) — an intervening "if" as a footprint step (#3033) · Accepted

The #3013 closing note left 24 cards on the ratchet. 21 of them re-check an intervening "if" as the ability resolves
(CR 603.4), and a bare `Do` can't say that. This amendment adds the step that closing note sketched and moves the
21 onto it. Arboreal Alliance, Cadira and Ikra Shidiqi are left for the next slice.

### Decision

1. **The step.** `effects.If{Cond, Then}` runs its `Then` steps only when `Cond` holds as the item resolves. Its
   footprint is one `FootprintCondition` step that names the state the condition reads (`FootprintStep.Cond`),
   followed by the footprint of every `Then` step, declared as if it runs. If a `Then` step can't declare a
   footprint, the row gets none, as with a plain `Do`. An `If` can sit inside another `If` (Lunar Convocation's
   "gained and lost life").
2. **The conditions.** Each `effects.Condition` is a named value whose `condition()` names one
   `game.FootprintCond`, and `addConditionReadLocked` reads each one:

   | Condition | Cards | What the check reads |
   | --- | --- | --- |
   | `CondLifeGained{N}`, `CondLifeLost{N}`, `CondCreatureDied` | The Gaffer, Crested Sunmare, Indulging Patrician, Lunar Convocation, Faramir | Nothing, once the tally already holds. The tally only rises during a turn, and a turn can't end with the batch on the stack. A tally that doesn't hold yet is refused, since another item could make it true. |
   | `CondCitysBlessing` | Deadeye Brawler, Mausoleum Harpy, Resplendent Griffin | Nothing, once held: the blessing is never lost (CR 702.131c). Refused before then. |
   | `CondNoSpellsCast` | Edgar, Moonlit Sovereign | Nothing while it holds, since no step casts a spell. Refused once a spell was cast. |
   | `CondMonarch` | Ginger, Queen of Sweets | `monarchRead`. It conflicts with an item that makes a player the monarch. |
   | `CondFewerCountersOnThis{Kind, N}` | Runaway Steam-Kin | A read of the item's source, like a counter on it. |
   | `CondPowerAtLeast{N}`, `CondBoard{Test}` | Agent of Treachery, Beastbond Outcaster, Colossal Majesty, Garruk's Uprising, Bugenhagen, Field of the Dead, Keeper of the Accord, Inventors' Fair, Lux Artillery, Padeem, Satoru | `boardReads`: the whole battlefield, read through the layers. `CondBoard`'s test may read only the battlefield, the item's controller, source and triggering event, and this turn's event log. |

3. **What a board read conflicts with** (`boardConditionSees`). Another item conflicts if it changes or pumps a
   permanent, creates a token, moves a card, puts a counter on a permanent, makes someone the monarch, or gives a
   player counters. Since a static may read the monarch or a player's counters without declaring it, those two
   conflict too. An item that only moves life totals conflicts only when some static on the board declares
   `DependsOnLifeTotal`. One that only touches a hand or library conflicts only when a static declares
   `DependsOnHandSize` or `DependsOnLibraryTop`. A pump on an object that bears rules (`pumpRules`) conflicts, as it
   does with a characteristic read. So does an item that can take a player out of the game, since that player's
   permanents leave with them (CR 800.4a).
4. **A new reason, `condition`.** A conflict found by rule 3, or by the monarch read, is counted under
   `cmdctrl_trigger_order_prompts_total{reason="condition"}`. The prompt then says: "<Card>'s trigger checks its
   "if" again as it resolves, and another of these triggers can change the answer."

### What changes for players

- **Lunar Convocation.** With life both gained and lost, the drain and the Bat no longer ask for an order.
  `TestB30LunarConvocationDrainsOnGainAndMakesABatOnGainAndLoss` now pins that.
- **Edgar, Runaway Steam-Kin and the blessed counters.** These move onto `CounterOnThis`, which already puts
  nothing on a source that has left (CR 400.7). Edgar's old body had no such check.
- **Keeper of the Accord.** Its Soldier row no longer reads a missing trigger context.
- **Unchanged.** Every row keeps its label and its trigger-time condition, so a queued item restores onto the same
  row, and there's no snapshot, wire or closure-field change.

### Tests

- `game/trigger_condition_test.go` covers each condition kind. It checks the refusal before a tally holds and the
  empty read after, the monarch and source reads, a board read against every step kind, the static-driven life
  and hand cases, and elimination.
- `footprint_conformance_test.go` resolves an `If` case for every `Condition` type
  (`TestEveryConditionHasAConformanceCase`), and a nested one.
- `trigger_independence_if_test.go` checks that each of the 21 rows declares its condition. On real boards, two
  board-condition draws skip the prompt, a board condition beside a token maker asks under `condition`, and The
  Gaffer beside Indulging Patrician skips it.
- The ratchet file goes from 24 lines to 3.

### Note (2026-10-14): the last three, counted and X amounts and Ikra

- **X from the cast.** `TriggeredAbility.XFromCast` makes the engine set the built item's `XValue` to its source's
  `Card.CastX` (CR 107.3m). The row needs no Build. The check accepts that `XValue` only on such a row, since it is
  a constant of the item fixed before the batch exists. Arboreal Alliance is `Do(CreateXByXToken{Template})`. An
  item queued before this change carries X in `Params.Amount`, which the step still reads, and is never plain.
  Arboreal Alliance's own "when this enters" row watches entering, so the check still treats the board as having a
  reactor. Its batches still ask (`reactor`) until the deferred c1 rule exists.
- **A count of the board.** `CreateTokensForEach{Template, Count}` declares a board read (`FootprintCondBoard`)
  and then the token step, so a token maker or anything else that changes the battlefield conflicts with it.
  Cadira, Caller of the Small uses it.
- **The subject is the event's source.** `TriggeredAbility.SubjectIsEventSource` snapshots `Event.Source` as
  `TriggerContext.Object`, at the moment and with the CR 603.10 reading the event's own object gets. That is
  catalog data, so there is no snapshot field and no new closure route. Ikra Shidiqi is
  `Do(GainLifeEqualToSubject{SubjectToughness})`. The creature that dealt the damage gives its toughness as it
  last existed once it has left, where the old body gave the toughness it had when it dealt the damage. An older
  queued item's `Params.Amount` is the fallback.
- **The ratchet file is empty.** `unfootprinted_trigger_lines.txt` keeps only its header, and the audit still fails
  on any new one-sentence trigger written as a closure.

## Amendment 2026-10-15 (fifth) — the attachments of a permanent that left (#3780) · Accepted

Decision 12's `PermanentInfo` records what a departed permanent was attached TO (`AttachedTo`, `AttachedObject`) but not what was attached to IT. By the time a sacrifice-cost ability resolves, the state-based action has already unattached an Equipment from the departed host (CR 704.5n) or put an Aura into a graveyard (CR 704.5m), so the attachment's own link is gone. Zack Fair's "attach an Equipment that was attached to Zack Fair to that creature" (the Cloud, Ex-SOLDIER deck, #3656) had nothing to read and shipped with a caveat.

- **`PermanentInfo.Attachments []ObjectRef`.** Every Aura, Equipment and Fortification attached to the permanent as it last existed, each as the object (CR 400.7), in battlefield order. It is written at the same moment and from the same live card as the rest of the record (`rememberDepartingPermanentLocked`), and filled on a live read too, so a reader sees one shape either way. A pinned link whose host has the same ID but a newer epoch (#3159) names an earlier object and is not counted (`attachmentObjectsLocked`).
- **`Game.LiveAttachmentsForEffect(refs)`.** The refs that are still the same object on the battlefield, in order. An Aura the state-based action binned is gone; an Equipment that was blinked is a new object and is skipped.
- **Zack Fair.** Of those, only an Equipment that could legally be attached to the target (`AttachLegalForEffect`, CR 701.3b) counts. One is attached; with several the controller picks one (`choose_cards`, min and max 1, "an Equipment"); none does nothing. A target that left in response ends the ability before this runs (CR 608.2b). Zack Fair is Full.
- **Snapshot.** `lastKnownPermanents{}[].attachments` is additive in v9 and recorded in the shape file. A file from before this carries no attachments, so the Equipment is simply not moved; a binary from before this that drops the key restores the game, with the same result. No bump.

**Back-outs.** Without the write in `rememberDepartingPermanentLocked`, `TestPermanentLKIRecordsWhatWasAttachedToIt`, `TestZackFairMovesTheEquipmentThatWasAttachedToHim` and `TestZackFairWithTwoEquipmentAsksWhichMoves` fail. Without the epoch comparison in `LiveAttachmentsForEffect` the same game test fails; without the pinned-epoch skip `TestPermanentLKIIgnoresALinkToAnEarlierObject` fails; without the copy in `cloneLastKnownPermanents` `TestPermanentLKIAttachmentsRideCloneAndSnapshot` fails.

## Amendment 2026-10-15 (sixth) — "that spell's mana value" counts X in the trigger snapshot (#3782) · Accepted

Krile Baldesion returns "target creature card with mana value equal to that spell's mana value". The clause is built as
the trigger goes on the stack (CR 603.3d) by a `TargetsFrom` that reads only its trigger context, so restore can
re-derive it and the row needs no `TargetsFromReadsBoard`. The context's `Object.ManaValue` was `Card.ManaValue()`,
which reads `{X}` as zero, so a Fireball cast for X=3 would have offered the one-drops.

`objectSnapshotLocked` now takes the value the spell has on the stack (CR 202.3e): when the object is a spell on the
stack with an announced X, the snapshot's `ManaValue` is the cost's mana value with that X. The snapshot is taken as
the event fires, while the spell is still on the stack, and is carried verbatim on the item and across a clone or
restore, so the value cannot drift once the spell resolves. Every other object (a permanent, a card in a graveyard)
still reads `{X}` as zero, which is what Scrap Trawler's "lesser mana value" wants. No wire, snapshot-schema or
closure-field change.

**Back-outs.** `TestACastTriggersSnapshotCountsTheXTheSpellWasCastWith` and `TestKrileCountsTheXAnXSpellWasCastWith`
fail without the X read; `TestKrileUsesTheTurnOnlyWhenACardIsReturned` fails without the `NoteDoneThisTurn` call.

## Amendment 2026-10-15 (seventh) — a became-target event says what kind of object targeted (#3779) · Accepted

"Whenever one or more creatures you control become the target of an activated ability" (Professor Hojo) has to tell an
activated ability from a spell or a triggered ability. `EventBecomesTarget` carried `StackItemID` but not the kind of
the thing it named, so a trigger would have had to look the item up on the stack itself, which works only while the
item is still there.

### Decision

1. **`Event.TargeterKind`** (`StackItemKind`: `spell`, `activated` or `triggered`) is stamped by
   `emitBecameTargetLocked`, once for the whole fan-out, from `StackMeta[itemID].Kind`. Every caller creates its stack
   item before it emits (the cast, the catalog and sandbox activations, a spell copy, an ability copy, a change of
   targets, and the CR 603.3b drain for a trigger), so one helper stamps all of the sites. An item that is not on the
   stack reads empty, which matches nothing. A copy of an activated ability is an activated ability (CR 707.10).
2. **The predicate** is `creatureYouControlBecameTargetOfAnActivatedAbility` in `professor_hojo.go`: a card target, a
   creature on the battlefield controlled by the source's controller, `TargeterKind == StackItemActivated`.
3. **"One or more … only once each turn" is one gate.** An activation naming two creatures is two events;
   `b11TriggeredThisTurn` declines the second because the first trigger is already counted from the moment it is put
   on the stack, the same way Welcoming Vampire collapses a batch. A trigger that was countered still counts.
4. **Snapshot.** `events[].targeter_kind` and the same field inside every embedded trigger event are additive under v9
   and omitted when empty. A binary that drops the key reads the old events as having no kind, which only means an
   older Hojo trigger does not fire for them.

## Amendment 2026-10-14 (fifth) — "use this order again" for a repeated batch (#3016) · Accepted (2026-10-15)

Issue [#3016](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3016), c2 of the #3013 amendment. The owner chose
the feature on 2026-10-11 (decision 4 there) and asked for this amendment **for review before anything is built**.
No code changes with it. It builds on [ADR 0127](0127-answering-repeated-prompts-for-you.md) (standing answers),
whose §1 lists the trigger order as "never covered"; this amendment is the exception. Numbering: an amendment to
ADR 0018, so no ADR number is taken.

### The question

A seat that orders the same rows again and again is asked again and again. The #3013 work removes the prompt where
the order cannot change the game, and #3540 (2026-10-14) took #3016's reported pair (Pitiless Plunderer and
Falkenrath Noble) off the prompt entirely. What is left is a batch the check still asks about: a reactor on the
board, an undescribed row, or a pair whose order really matters (prowess beside Caldera Pyremaw). The player there
has a policy ("Treasure first, then the drain") and gives it ten times a game. CR 603.3b makes the order theirs to
choose, and nothing stops them choosing it once, as a CR 732.1a shortcut the table can see.

The issue's five questions: what counts as the same batch; how a remembered order meets the three modes (#1968,
#2334); undo and snapshots; the bot seat; how a player sees and clears it. Two more came up: hidden information, and
when a remembered order expires.

### What exists

- **The prompt.** `drainPendingTriggersAPNAPLocked` (`game/mutations.go`) asks a seat whose batch
  `triggerOrderAskWhyLocked` says needs an order, by queueing one `PendingChoiceTriggerOrder` with the items' IDs
  (`TriggerOrderIDs`) and the #3015 sentence. `ResolveTriggerOrder` (`game/pending_choice.go`) takes a resolution
  order, marks the items `Ordered`, dequeues the prompt (a player decision, so the CR 732 loop run restarts) and
  runs the drain again.
- **Row identity.** A stamped catalog trigger carries `Params.Ability`, an `AbilityRef` `{key, slot, ref, name}`
  (ADR 0041 P9). An engine keyword trigger carries a keyed body (`"prowess/pump"`). A closure trigger with no ref,
  and an engine-built delayed or reflexive trigger, have neither.
- **Standing answers** (ADR 0127, `game/auto_answer.go`). A prompt with an `AutoAnswerKey` and a rule for it is
  answered by the room after the commit that raised it, as a commit of its own (`NextAutoAnswer`, then
  `AutoAnswer`). That commit is the chooser's and is `freeUndo`. It writes an `auto_answer` log line marked
  "(automatic)" and does not restart the loop run. Undoing it reopens the prompt marked `AskedByHand: undone`. A
  controlled chooser (ADR 0167 §4), an open loop notice and a bot seat are never answered. The rules live in
  `Player.AutoAnswers` (account-synced, `always` or `never`).
- **The client.** `ChoicePromptModal.svelte` draws the order as a click-to-number list with **Resolve in this order**
  and **Auto-order** (#3036). The dock's notice for an automatic answer has **Undo** and **Ask me next time** (ADR
  0127 §6).

### A. Which batches are "the same"

- **(A1) The set of distinct rows. Recommended.** Each item maps to a *row identity*: its `AbilityRef` as
  `key|slot|ref|name`, or, for an engine keyword trigger, its keyed body plus its source's catalog key
  (`prowess/pump@<key>`, since prowess on two different creatures pumps two different creatures). The batch key is
  the sorted set of distinct identities. Counts don't matter: three Falkenrath drains and one Solemn draw are the
  same batch as one drain and one draw. Source instances don't matter: a second Blood Artist, or a token copy, is
  the same row.
- **(A2) The multiset of rows.** A1 with counts. Three drains and a draw would differ from one and one, so the player
  answers once per batch size. More exact, and more prompts in exactly the case the issue is about.
- **(A3) Source instances and labels.** A new copy of a card, or a card that dies and comes back, starts over.
  Rejected: it fails the aristocrats board, where sources come and go.
- **(A4) Oracle ID and label.** Like A1, but it merges rows the catalog keeps apart. Rejected: `AbilityRef` is
  already the identity the snapshot trusts.

An item with no row identity leaves the whole batch unkeyed, and it is asked as today. So does an item whose source
is face down or in a hidden zone (G).

### B. What is remembered

**A ranking of the batch's rows, not a permutation of item IDs.** Applying it is a stable sort of the new batch by
rank, so copies of one row stay together in the order they triggered (owner decision 2 of #1968). An answer that
interleaves copies of one row (Plunderer, Noble, Plunderer) is not a ranking. The **Use this order again** toggle is
disabled on it, with the hint "Only an order that keeps each card's triggers together can be reused."

### C. Where it lives and when it expires

- **(C1) Per seat, per game. Recommended, and what the issue says the owner chose.** `Player.TriggerOrderMemory`, a
  map from batch key to ranking, at most **32** entries (a 33rd replaces the one applied least recently). It is game
  state: cloned, snapshotted additively as `triggerOrderMemory` (shape file updated, no schema bump; absent restores
  empty), and gone when the game ends. It needs no expiry inside the game: it applies only when the same rows
  trigger together, and a row no longer in play never does.
- **(C2) The account's synced settings, like ADR 0127's rules.** One answer for every game with that deck. Rejected
  for now: the right order depends on the board around it, and the issue chose per game.

### D. Opting in and out

- **(D1) A toggle on the prompt. Recommended.** The trigger-order prompt shows **Use this order again** under the
  list. It is off by default, and on when the batch already has a remembered order (a prompt asked by hand). **Resolve
  in this order** with it on sends `remember: true`, and the server stores the ranking in the same commit. With it
  off, `remember: false` clears any remembered order for that key. The toggle is hidden on a prompt with no key.
- **(D2) A "Same as last time" button, and nothing standing.** One click per prompt instead of none. Simpler (no
  automatic answer, no new state), but it doesn't remove the interruption the issue is about.
- **(D3) Remember every answer without asking.** Rejected: a one-off order would silently become a policy.

**Seeing and clearing it.** An applied order is an ADR 0127 automatic answer, so it reuses all of that: the public
log line "Luke ordered Pitiless Plunderer, then Falkenrath Noble (automatic)", the chooser's notice with **Undo** and
**Ask me next time**, and the `freeUndo` commit. **Ask me next time** sends a new action, `forget_trigger_order
{key}` (seat-scoped, mints no undo entry, never offered by the enumerator). The game menu gets **Forget remembered
trigger orders (N)**, which sends it with no key to forget all. Every name goes in `client/src/lib/labels.ts`, each
with a hint (the AGENTS.md labels contract).

### E. The three modes

- **When it matters** (the default): a remembered order answers only a batch the check would ask. A batch the check
  skips is never touched, since its order can't change the game.
- **Always ask.** **(E1) A remembered order still answers. Recommended.** The toggle is an explicit, per-batch choice
  made after the global one, so it is the more specific instruction. (E2) Always ask wins and the toggle is hidden.
  That keeps #1530's promise literally, but an always-ask player then can't use the feature at all.
- **Never ask:** no prompt is queued, so nothing is answered and nothing is learnt. A memory from before the switch is
  kept but unused. "Never" means "order them for me", and the game's order is the order they triggered in.

### F. Applying it, and the validity check

The prompt is queued as today, with `AutoAnswerKey = "trigger_order|" + batch key` (empty when unkeyed).
`NextAutoAnswer` reads the rule from `TriggerOrderMemory` instead of `AutoAnswers` for this kind, and `AutoAnswer`
gains a `PendingChoiceTriggerOrder` case. Before it answers, it checks again, on the live batch:

1. every item in `TriggerOrderIDs` still maps to a row identity, and their set equals the stored key;
2. no source is face down or in a hidden zone;
3. the chooser is not controlled by another player (asked by hand, `controlled`), and no loop notice is open
   (`loop`).

If a check fails, the prompt is asked by hand with that reason, as in ADR 0127. Otherwise the answer is the stable
sort, applied through `ResolveTriggerOrder`'s body without `dequeueChoiceLocked`'s player-decision mark. So a
remembered order never restarts the CR 732 loop run, and a Plunderer and Noble sacrifice loop still reaches ADR
0055's breaker. Applying the order inside the drain instead was considered and rejected: it would leave nothing to
undo, no notice and no log line.

### G. Hidden information

The key and the ranking are private to the seat: `auto_answer_key` and `trigger_order_remembered` are on the
chooser's view only. The applied order is public, because the items go on the stack in that order and the log line
names only what the stack already names. A face-down source is never keyed: its row identity is its real catalog key,
and an automatic log line built from it would name the card to the table (CR 708.2). An item whose source is in a
hidden zone is not keyed for the same reason.

### H. Undo and snapshot

- **The memory is written by the commit that answered the prompt.** Undoing that commit takes the memory back with
  the answer. This differs from `AutoAnswers` and `TriggerOrder`, which `RestoreFrom` carries across an undo as
  settings. The memory is part of the play that taught it, not a setting, so `RestoreFrom` does not carry it.
- **`forget_trigger_order` mints no undo entry.** An undo of a commit made after the forget keeps it forgotten; an
  undo of a commit made between teaching and forgetting brings the order back. That is the cost of not carrying the
  memory across `RestoreFrom`, and it is rare.
- **Undoing an automatic answer** reopens the prompt marked `undone`, as ADR 0127 §6 does. The player answers by
  hand, and the toggle decides whether the memory stays.
- **Snapshot.** `triggerOrderMemory` is additive. The prompt's `AutoAnswerKey` is already snapshotted. A prompt
  queued by an older binary has no key and is asked by hand.

### I. Bots and agent seats

- **Bots don't use it.** They answer an order prompt from the heuristic at no cost, a bot seat never has a memory,
  and `autoAnswerRuleLocked` already skips `IsBot`.
- **Agent seats (MCP).** #3016's evidence came from one: Claude-QA was asked the same pair about ten times. **(I1)**
  `legal` offers an agent seat one extra move on a keyed trigger-order prompt, "resolve in the order shown and use it
  again", which sends `remember: true`. Recommended, since an agent seat pays for every prompt in tokens. **(I2)**
  Agent seats don't get it.

### J. ADR 0193 (proposed, PR #3699)

A recorded loop replays the prompt answers it recorded; a standing order is answered by the room, like any standing
answer. They share no state. A remembered order that fires during a recording is an automatic commit, which ADR
0193 §2 would record or skip; that is its call.

### Wire

- `resolve_choice` for `trigger_order` gains `remember?: boolean`. Absent means "change nothing", so an older client
  is unaffected.
- `PendingChoiceView` for `trigger_order`: `auto_answer_key` (exists) and `trigger_order_remembered`, both on the
  chooser's view only.
- `PlayerView.trigger_order_memory_count`, own view only, for the menu entry.
- `forget_trigger_order {key?: string}`; no key forgets all.
- The `auto_answer` log line gains a trigger-order rendering (`Event.Call = "order"`, the row labels in order).
- `docs/protocol.md` documents each; `docs/bot.md` says bots ignore it.

### Tests (when built)

- `game`: batch keys (counts and source instances ignored; a closure row, a delayed trigger and a face-down source
  leave the batch unkeyed); a remembered order applied by stable sort; an interleaved answer refused for `remember`;
  each validity failure asked by hand with its reason; the loop run not restarted, and a sacrifice loop stopped by
  the breaker; "never" untouched and "always" answered (E1); a skipped batch untouched; the 33rd entry; clone,
  snapshot and shape; undoing the teaching commit removes the memory.
- `ws`: the automatic commit is the chooser's and `freeUndo`; undoing it reopens the prompt marked `undone`.
- `legal`: the agent-seat move, never offered to a bot.
- `protocol`: the fields on the chooser's view only; the log line.
- Client (vitest): the toggle's default and disabled states and hint, the payload, the notice's **Ask me next time**,
  the menu entry, `labels.test.ts`. Playwright: two batches of the same rows, the second answered automatically and
  shown in both seats' logs.

### Delivery (after review)

1. Server: keys, memory, `AutoAnswer`'s new case, `forget_trigger_order`, snapshot, view, log, `legal`, docs.
2. Client: the toggle, the notice reuse, the menu entry, labels and tests.

### Owner decisions (2026-10-15)

The owner approved the recommendations on 2026-10-15. Each one is accepted as written:

1. **The same batch.** (A1) the set of distinct rows, counts ignored. **Accepted.**
2. **Rankings only.** Only an order that keeps each row's copies together is remembered (B). **Accepted.**
3. **Opt-in.** (D1) a toggle on the prompt, off by default. **Accepted.**
4. **Always ask.** (E1) a remembered order still answers. **Accepted.**
5. **Undo.** The memory is game state, undone with the answer that taught it and not carried across `RestoreFrom`
   (H). **Accepted.**
6. **Agent seats.** (I1) MCP seats get the remember move. **Accepted.**
7. **The cap.** 32 remembered orders per seat per game, the least recently applied replaced. **Accepted.**
