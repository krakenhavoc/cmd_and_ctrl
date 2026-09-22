# ADR 0018 — Triggered abilities use the stack

**Status:** Implemented · 2026-09-02 · Branch `feat/s19-triggers-on-stack`
**Addendum:** 2026-09-17 · **Accepted** · [Trigger doubling (CR 603.2d)](#addendum-2026-09-17-trigger-doubling-cr-6032d--accepted) · tracked on [#752](https://github.com/krakenhavoc/cmd_and_ctrl/issues/752)

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
has ceased to exist either way (CR 608.2m). `nil` keeps the S13.1
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
  actions after which the actor keeps priority (CR 116.3c), and
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
the CR 726 shortcut prompt, **blocks**. It is the one entry worth arguing
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
- a **CR 726 loop notice** ([ADR 0055](0055-loop-breaker.md)), checked
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
than by a code review.

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

## Out of scope (explicit deferrals)

- **Treasure's sac-for-mana** is inert until S21 ships sacrifice
  as a cost component; Smothering Tithe makes countable artifacts.
- **"You may draw"** on Rhystic Study is treated as "draw".

- **Modal triggers** ("draw a card or gain 3 life") still need a
  `ModePrompt` slot; nothing in the catalog needs one yet.
- ~~**Trigger ordering UI** for one player's simultaneous triggers
  (CR 603.3b)~~ — shipped in sub-PR 8: `PendingChoiceTriggerOrder`
  holds the APNAP drain until each seat with ≥2 *differing*
  triggers has ordered them (identical ones auto-order; the
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
| Windcrag Siege | "As this enchantment enters, choose Mardu or Jeskai." | an as-enters mode choice (the issue calls it misattributed) |

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

**Cards:** Harsh Mentor and Runic Armasaur, both `full`.

What stays open on this family, and is a different row: a watch on an
ability's activation whose TARGET clause reads the triggering event
(`docs/engine-seams.md`, "Trigger/target clause reading the triggering
event's data"). A `Build` closure can capture an event field and hand
it to a non-targeted effect — that is what this note's `build`
argument is — but `TriggeredAbility.Targets` is a static `TargetSpec`
evaluated by the harvester and cannot see the event at all.
