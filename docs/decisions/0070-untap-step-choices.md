# ADR 0070 — Untap-step choices (CR 502.3)

**Status:** Accepted · 2026-09-18 · S41 — Turn machinery and per-turn accounting (tracker [#884](https://github.com/krakenhavoc/cmd_and_ctrl/issues/884))
**Issue:** [#826](https://github.com/krakenhavoc/cmd_and_ctrl/issues/826)
**Numbering:** re-checked immediately before the push on 2026-09-18, after a
full `+refs/heads/*` fetch (280 remote heads). Every `docs/decisions/` path ever
touched on any remote ref was listed from the log, and every branch TIP with
`ls-tree`. `develop` carries up to 0065 plus 0069; 0062, 0066, 0067 and 0068 are
on parallel branches not yet merged. **0070** is free at every branch tip and is
taken here.

One thing the history scan turns up and the tip scan does not:
`docs/decisions/0070-the-mana-spent-on-a-spell.md` exists in the *history* of
`feat/789-761-counter-costs-and-mana-spent`
([PR #958](https://github.com/krakenhavoc/cmd_and_ctrl/pull/958)), which
renumbered that ADR **down to 0068** before this check. So 0070 was vacated
rather than claimed, no head holds it, and
`TestADRNumbersAreUniqueAndMatchTheirHeading` is green. AGENTS.md §4's "numbers
are not reused" rule is about an ADR that was published and then renumbered; its
four permanently-unused numbers are 0005, 0024, 0029 and 0030, which stay
unused.
**Builds on:** [ADR 0058](0058-doesnt-untap.md) (untap-step restrictions, the
next-untap-step marker, stun counters — its Decision 8 left this shape out on
purpose, and question 3 decided Winter Orb waits for this ADR),
[#74 / `untap.go`](../../server/internal/game/untap.go) (one untap primitive and
`UntapStepPermission`), [ADR 0013](0013-replacement-effects.md) + #710 (the
step-entry split that a paused step resumes through),
[#661 / `cleanup.go`](../../server/internal/game/cleanup.go) (the other step
whose turn-based actions have one exit function),
[ADR 0018 §6 + #730 / #794](0018-triggers-on-the-stack.md) (the choice gate),
[#544 / `chained_choice.go`](../../server/internal/game/chained_choice.go)
(`ChooseCardsPrompt`, its `Validate` hook and the enumerator contract),
[ADR 0041](0041-game-persistence.md) (`ContinuationCensus`),
[ADR 0033](0033-ai-bot-seat.md) (bot tiers and the enumerator).
**Related:** [ADR 0059](0059-turn-machinery.md) (turn machinery; a paused untap
step is inside a turn, not across one), [ADR 0055](0055-loop-breaker.md)
(answering a prompt is a player decision), [ADR 0063](0063-durations-and-control.md)
(the `Duration` model this ADR's Decision 6 would extend), [#567](https://github.com/krakenhavoc/cmd_and_ctrl/issues/567)
(cumulative upkeep — same PR, different mechanism).

## Context

### CR 502.3 has always been a determination; the engine never let it be one

> **CR 502.3.** Third, the active player determines which permanents they
> control will untap. Then they untap them all simultaneously. This turn-based
> action doesn't use the stack. Normally, all of a player's permanents untap,
> but effects can keep one or more of a player's permanents from untapping.

Two sentences, and the engine implements the second one. `untapStepSetLocked`
(`server/internal/game/untap.go`) answers "which permanents untap" as a pure
function of the board: the active player's tapped permanents, minus anything an
`UntapStepRestriction` or a next-untap marker holds back (ADR 0058), plus
anything an `UntapStepPermission` adds (#74). There is no player in that
sentence. The determination is computed, not made.

For every card in the catalog today that is the right answer, because every
effect in it is a flat "doesn't untap". Two printed families make CR 502.3's
first sentence a real decision, and both were filed out of scope by ADR 0058
Decision 8:

| shape | cards | examples |
|---|---:|---|
| "players can't untap more than N ⟨kind⟩ during their untap steps" | 9 | Winter Orb, Static Orb, Winter Moon, Damping Field, Smoke, Stoic Angel, Mungha Wurm, Dovin Baan's emblem |
| "you may choose not to untap this during your untap step" | 45 | Rust Tick, Amber Prison, Sleeper's Robe, and 42 more |

The owner decided on 2026-09-17 (ADR 0058 question 3, option (a)) that Winter
Orb, Static Orb and Winter Moon **wait for this ADR** rather than ship with an
engine-chosen pick, because under a Winter Orb lock *which land you untap* is
the decision the card is about, every turn, for everyone. A visible automatic
pick in that spot is worse than a card that honestly is not there yet.

### Why this is hard: the untap step has no priority and no prompt

The untap step grants nobody priority (CR 502.4). The engine reflects that
directly: `finishStepEntryLocked`'s `StepUntap` case runs
`performUntapStepLocked` and then **advances the cursor and recurses into the
next step's entry hook**, all inside one write lock
(`server/internal/game/game.go`). A single client-visible `advance_step` from
the end step walks End → Cleanup → the next seat's Untap → that seat's Upkeep
without ever returning to the caller. Nothing in that walk can wait for a human.

The engine has learned this shape twice already and the second time it named it:

- **#710** split `runStepEntryHooksLocked` into itself plus
  `finishStepEntryLocked` so the CR 616 replacement-ordering prompt over a step
  transition could pause the entry and have the answer finish it. One function
  is the "rest of the step entry", and both the unpaused path and the resume
  call exactly it.
- **#661** gave the cleanup step `exitCleanupStepLocked`: one function that is
  the cleanup step's exit, called from the step-entry hook and from
  `DiscardSelection`'s resume, because those two sites used to decide separately
  and the discard path inherited a bug from the divergence.

The untap step needs the same treatment and gets it here.

### What "the same question, two directions" already bought

ADR 0058 Decision 1 put the restriction and the permission in one function for
one reason: CR 502.3 asks one question and they are its two directions, so
reading them at the same moment in the same place means they cannot drift. A cap
("no more than one land") and an opt-out ("you may choose not to untap") are two
more directions of the same question. They go in the same place.

## Decision 1 — The untap step's turn-based action can pause, and it has one exit

`performUntapStepLocked` gains a return value: **did the step pause?**

```go
// performUntapStepLocked … reports whether the step PAUSED for the
// active player's CR 502.3 determination. A paused step has not
// untapped anything and has not moved the cursor; the prompt's
// continuation finishes it.
func (g *Game) performUntapStepLocked(seat int) (paused bool)
```

and the step-entry hook stops advancing the cursor itself:

```go
case StepUntap:
    if g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) {
        g.Seats[g.Turn.ActiveSeat].UndosRemaining = g.UndoLimit
        if g.performUntapStepLocked(g.Turn.ActiveSeat) {
            // CR 502.3 paused for the active player's choice. The
            // prompt's continuation calls exitUntapStepLocked.
            return
        }
    }
    g.exitUntapStepLocked()
```

`exitUntapStepLocked` (`server/internal/game/untap_choice.go`) is the untap
step's one exit, the shape `exitCleanupStepLocked` established:

```go
// exitUntapStepLocked ends the untap step and moves the cursor on.
// Untap grants no priority (CR 502.4), so "exit" means the next step's
// entry hook, immediately, in this same write lock.
//
// Two callers, which are the two ways the untap step's turn-based
// actions can finish: the StepUntap case of the step-entry hook, and
// the untap-choice prompt's continuation.
func (g *Game) exitUntapStepLocked()
```

Three consequences are decided with it:

**(a) Nothing untaps until the determination is complete.** CR 502.3 is
*determine, then untap them all simultaneously*. A paused step has untapped
nothing; the whole set — the permanents that were never in question and the ones
the player picked — untaps in one loop from the continuation, through the one
`untapPermanentLocked` primitive, exactly as the unpaused path does. The
alternative considered and rejected is below.

**(b) The prompt blocks the table.** The untap step is the first thing that
happens in a turn and CR 502.3 happens before anything else in it. A table that
could walk past the prompt would be deciding it by doing. `untap_choice` is
classified `true` in `choiceGateDecisions` — which is also the deny-by-default
answer, so this is a confirmation rather than a new latitude.

**(c) The batch boundary is the step entry, unchanged.** `advanceCursorLocked`
opens the event batch when the cursor moves (#829); a paused untap step has not
moved the cursor, so it is still inside the batch its own step entry opened, and
the answer's untaps land in that batch. Nothing new opens a batch — unlike
`repeatCleanupStepLocked`, which begins a step without moving the cursor and had
to.

**Summoning sickness and the marker sweep stay eager.** Clearing
`SummonedThisTurn` is a *different* turn-based action (CR 302.6 is about whose
turn it is, #74) and does not depend on the answer. Consuming next-untap markers
(`consumeUntapSkipsLocked`) stays where ADR 0058 put it — before any untap —
because the markers are used up by the step that happened, not by the
permanents that untapped, and a pause that re-ran it would consume twice.

**A stale answer is tolerated, not trusted.** The continuation re-checks
`g.State == StateActive && g.Turn.Step == StepUntap` before it exits the step,
the tolerance #701 and #725 each had to add to one resume path. If an admin
walked the cursor past the prompt, the picked permanents still untap (the
player's answer is honoured) and the cursor is left where it is rather than
advanced a second time.

## Decision 2 — One prompt shape: `untap_choice`, a card-set pick over the permanents in question

The prompt is a card-set pick: the candidates are battlefield permanents, the
answer is `{card_ids: [...]}`, and the client renders it with the card picker it
already has. It is `PendingChoiceChooseCards`'s shape — `ChooseCards`,
`ChooseMin`, `ChooseMax`, `chooseCardsResume` with its `Validate` hook — reached
through `QueueUntapChoiceLocked`, which builds a `ChooseCardsPrompt` and stamps a
different kind on it.

**It is a new kind, not a reuse of `choose_cards`.** The question is the same
shape; three things about it are not, and the third is the deciding one.

1. **The sign is inverted for a bot.** `choose_cards`'s heuristic branch reads
   its candidates as cards being *given up* — `valueKeptInHand` scores an answer
   by what it does **not** name (#798), because every `choose_cards` prompt whose
   candidates are the bot's own hand is a discard or a Sylvan Library put-back.
   Naming a permanent here means *untapping* it. A Winter Orb answer scored
   through the `choose_cards` branch falls to the flat `0.5` "no opinion" case
   and the bot takes the first set the enumerator offers, which is battlefield
   order — precisely the automatic pick the owner rejected for humans.
2. **The candidates are public.** `filterPendingChoices` strips `choose_cards`
   options **and its bounds** from every seat but the chooser, because those
   candidates are usually a hand and even "2 of 3" leaks the size of a hidden
   set. Untap candidates are tapped permanents on the battlefield, which every
   seat can already see; hiding them would make the prompt unreadable for
   spectators and teach the wrong lesson about the redaction rule.
3. **The player's words are different.** "Choose cards" and "choose which of
   these untap" are not the same sentence, and a kind is the only thing on the
   wire that distinguishes them.

Everything else is shared rather than copied: one validation function
(`checkChooseCardsPicksLocked`, which now accepts either kind), one resolver
body, one enumerator branch, one wire projection, one client picker. That is the
"one prompt shape" rule — one shape, two kinds, no second copy of anything.

**The prompt, precisely.**

- **Chooser:** the active player. CR 502.3 gives them the determination, and
  every card in both families is scoped to their own untap step (Decision 4).
- **Candidates:** the permanents the active player controls that are in the
  CR 502.3 set *and* in question — a member of at least one **binding** cap
  (Decision 3), or opt-out-able (Decision 5). A cap that does not bind (Winter
  Orb with one tapped land) puts nothing in the prompt.
- **Not candidates, and untapped without asking:** every other member of the
  set, including permanents another player untaps through an
  `UntapStepPermission`, and including the active player's permanents no cap
  counts and no clause lets them keep tapped.
- **Not in the set at all, and so never offered:** anything an
  `UntapStepRestriction` or a next-untap marker holds back (ADR 0058), and
  anything already untapped. A permanent that *can't* untap is not one of the N
  choices, which is the sentence ADR 0058 Decision 8 wrote for this ADR.
- **Stun counters (CR 122.1d) still apply to a chosen permanent.** They replace
  the untap inside `untapPermanentLocked`, which is downstream of every
  determination this ADR makes; a player may spend their one Winter Orb land on
  a stunned one and lose a stun counter instead. That is what the rules say, and
  the board shows the pips.
- **No prompt when there is nothing to decide.** A pause costs a round trip and
  a restore point; the step takes it only when a cap actually binds or an
  opt-out permanent is actually tapped.

**`Min` and `Max` are bounds, and `Validate` is the rule.** `Max` is a genuine
upper bound — `min(len(candidates), capMax + |candidates outside that cap|)`
over every live cap — so a legal answer is never refused by the count guard.
`Min` is `1` whenever any candidate is *not* opt-out-able, and `0` otherwise.
`Min` is doing one specific job: `ChooseCardsPrompt.Validate` is deliberately
**not** called for an empty pick (an empty answer is the one the engine can never
refuse, which is what the enumerator marks `AlwaysLegal`), so the empty set has
to be excluded by the floor or not at all. Between the floor and the ceiling,
Decision 3's solver decides.

## Decision 3 — Caps compose: one solver, "at most N per cap, and no wasted untap"

A cap is declared on the catalog `Spec`, read at the step the way ADR 0058 reads
restrictions:

```go
// untap.go
// UntapCap is one "players can't untap more than N ⟨kind⟩ during
// their untap steps" clause (CR 502.3) — Winter Orb, Static Orb,
// Winter Moon, Damping Field, Smoke.
type UntapCap struct {
    // Applies reports whether the cap is live for activePlayer's
    // untap step. Winter Orb and Static Orb say "as long as this
    // artifact is untapped"; Winter Moon says nothing and is always
    // live. Same argument order as UntapStepPermission.AppliesTo.
    Applies func(g *Game, source *Card, activePlayer uuid.UUID) bool
    // Counts reports whether target is one of the permanents the cap
    // counts. Winter Orb: a land. Winter Moon: a nonbasic land.
    // Static Orb: every permanent.
    Counts func(g *Game, source *Card, target *Card) bool
    // Max is N. A cap of zero or less is not a cap, it is a flat
    // restriction, and belongs on Spec.UntapStepRestrictions.
    Max int
    Label string
}

var CatalogUntapCaps func(oracleID string) []UntapCap
```

declared as `Spec.UntapCaps []game.UntapCap` (the four edits AGENTS.md §7
"Adding a `Spec` slot" lists), read through `CatalogAbilityKey` so an orb that
has lost its abilities caps nothing, after `RecomputeLayersIfStaleLocked` so
"nonbasic" and "land" are the effective types.

**They compose as a set system, not as a priority order.** Winter Orb, Static
Orb and Winter Moon on one board are three caps over three different, overlapping
sets. A chosen set `S` is legal for the caps iff

> for every live cap `i`: `|S ∩ counted(i)| ≤ Max(i)`

and that is the whole rule. One predicate, evaluated over the picked set, with
no ordering between caps and no per-card arbitration — which is what makes a
tenth card in the family free.

**And one more clause, because untapping is not optional.** The existence of
"you may choose not to untap this" as printed text is the proof that untapping
is otherwise mandatory: CR 502.3's "normally, all of a player's permanents
untap" is the default, and a cap only removes the excess. So a legal answer must
also be **maximal** over the permanents the player has no printed permission to
hold back:

> for every candidate `c ∉ S` that is not opt-out-able: `S ∪ {c}` violates some
> cap.

Together, those two lines are `untapChoicePlan.legal(picks)`, the ONE copy of
"is this answer legal", captured as plain data (candidate IDs, per-candidate cap
membership, per-cap ceilings, per-candidate opt-out flag) when the prompt is
queued. It is `ChooseCardsPrompt.Validate`, so it runs on the submit path under
the write lock **and** inside `legal.EnumerateFor` under the read lock through
`ChooseCardsPickLegalLocked` — the #544 rule that every answer the enumerator
offers is an answer the resolver accepts, with no superset enumerated past a
constraint the enumerator cannot see. It takes no `*Game` and captures none, for
the reason `ChooseCardsPrompt.Validate` documents.

Worked example, the one the three cards make: Winter Orb and Static Orb both
untapped, the active player with four tapped lands (one nonbasic), two tapped
creatures and a tapped Winter Moon of their own. Caps: lands ≤ 1, nonbasic lands
≤ 1, permanents ≤ 2. Candidates: all seven permanents (Static Orb's cap counts
everything and binds at 2 < 7). Legal answers are the 2-element sets with at most
one land, so "one land plus one creature" and "two creatures" — and *not* "two
lands", and not "one creature" (the second slot is not blocked by any cap, so
leaving it empty wastes a mandatory untap).

**Bots pick greedily by value.** The heuristic's `untap_choice` branch scores an
answer as the sum of `permanentValue` over the permanents it names — the same
valuation the sacrifice branch uses with the opposite sign — so the enumerated
legal sets are ranked by what untapping them is worth and the best is taken. That
is a greedy pick over legal sets rather than a plan (a bot does not know which
colours its hand will need three spells from now), and it is documented in
`docs/bot.md`.

## Decision 4 — Caps and opt-outs are about the active player's own determination

Every printed card in both families is scoped to a player's *own* untap step:
"during **their** untap steps", "during **your** untap step". So a cap is
consulted only for permanents the active player controls, and only for the
active player's own CR 502.3 determination.

This gets the known interaction right for free, and it is the same argument ADR
0058 Decision 1 makes for restrictions. Under Winter Orb plus a Seedborn Muse,
during Alice's untap step Bob untaps his permanents — but that is not *Bob's*
untap step, so Winter Orb's "during their untap steps" does not reach it, and
Bob's lands are not capped and not in Alice's prompt. Alice's own lands are
capped, and she chooses. The engine enforces the scope structurally rather than
asking each cap to check which step it is in, so a cap predicate never has to
know.

## Decision 5 — "You may choose not to untap" is a per-object opt-out in the same prompt

```go
// untap.go
// UntapOptOut declares "you may choose not to untap this during your
// untap step" (CR 502.3) — Rust Tick, Amber Prison, and 43 more. Same
// signature and argument order as UntapStepRestriction, so selfOnly
// and AttachedToSource plug in unchanged; the printed family is
// self-only today.
type UntapOptOut struct {
    Optional func(target *Card, g *Game, source *Card) bool
    Label    string
}
```

declared as `Spec.UntapOptOuts`, read beside the caps.

It is not a second prompt. An opt-out permanent is a candidate in the same
`untap_choice` prompt, and the only difference is in the solver: it is exempt
from the "no wasted untap" clause, so leaving it out is always legal.
`untap_choice` therefore covers both families with one question, one payload and
one picker — which is the point of asking "choose which of these untap" rather
than "choose N lands" and "confirm each opt-out".

**The client starts the picker pre-filled when, and only when, choosing
everything is legal by the bounds** — that is, when `Max == len(candidates)`,
which is exactly the pure opt-out board. "You may choose *not* to untap" is
default-untap, so the player's click should be the deselection, not the
selection. On a board where a cap binds there is no such default, and the picker
starts empty rather than pre-filled with an illegal set the player has to undo.

## Decision 6 — "For as long as this remains tapped" is designed here and not built here

Rust Tick and Amber Prison print two clauses. The first is Decision 5's opt-out
and ships. The second is an activated ability — "{1}, {T}: Tap target artifact.
It doesn't untap during its controller's untap step **for as long as this
creature remains tapped**" — and it is ADR 0058 Decision 8's *other* bullet: an
`UntapStepRestriction` on the source whose predicate reads a **remembered
target**, which the engine has no per-source linked-object field for.

**The design.** `Card.NextUntapSkips []UntapSkip` (ADR 0058 Decision 2) is
already a per-card list of plain data describing why this permanent does not
untap, and it is already snapshotted, cloned, and discarded on a zone change. The
linked form is two more fields on that struct, not a third mechanism:

```go
type UntapSkip struct {
    Player uuid.UUID     // as today: nil = "its controller's next untap step"
    Source uuid.UUID     // NEW: while THIS permanent remains tapped
    SourceEnteredAt int64 // NEW: CR 400.7 — the same object, not a replacement
}
```

`Source == uuid.Nil` is today's one-shot marker, consumed by the step it names.
A non-nil `Source` is a CR 611.2b "for as long as" condition in the `Duration`
model's vocabulary (ADR 0063) — a `WhileSourceTapped` `DurationCondition` — read
rather than consumed: `hasNextUntapSkipFor` gains a branch that looks the source
up and asks whether it is still on the battlefield, still the same object, and
still tapped, and `consumeUntapSkipsLocked` skips conditional entries. The effect
ends when the source untaps, leaves, or is replaced by a new object, which is
CR 611.2b, CR 400.7 and the printed text agreeing.

**Why it is not built in this PR.** It is a second mechanism with its own
snapshot migration, its own wire projection (`CardView.no_untap` grows a third
reason), its own tests, and it is not what #826 is about — the issue's checklist
is entirely about choosing which permanents untap. ADR 0058 filed it as a
separate bullet and it stays one. Rust Tick and Amber Prison ship with the
opt-out implemented and their activated abilities declared in `Caveats`, the
precedent ADR 0058's own first wave set with Mana Vault and Claustrophobia.

**Built 2026-09-23 (#1313).** ADR 0058's amendment of that date builds this
design, with one change: the entry carries a `*Duration` (ADR 0063) instead of
the two bare fields above. `WhileSourceRemainsTapped` is the condition this
section names. Rust Tick and Amber Prison now ship their activated abilities.

## Decision 7 — Persistence, undo and clone need nothing new

The prompt rides `PendingChoice` and reuses its existing fields: `ChooseCards`,
`ChooseMin`, `ChooseMax` (all carried by `pendingChoiceSnapshot` and by `Clone`
today) plus the `chooseCardsResume` frame. The frame is already named in
`captureChoice`'s continuation-slot map, so `ContinuationCensus.ChoiceResumeFrames`
counts a paused untap step with no new counter and `Restorable()` already refuses
a restore point while one is open — which is the right answer, because a step
that has determined nothing and untapped nothing is mid-action.

The only new thing on the wire is the kind string. Undo works because the
continuation receives the live `*Game` and captures only scalars (the plan is
plain data, the IDs are IDs), the `StackItem.Effect` contract every frame in the
queue already follows.

A paused untap step is not a new class of stuck table: the prompt is queued to
the active seat, the enumerator offers that seat its answers and nothing else,
and a seat that has left the game cannot be the active seat.

## Decision 8 — Docs that change with the code

- **AGENTS.md §7** — the untap section gains the cap / opt-out slots and the
  rule that a paused turn-based action has one exit function.
- **`docs/bot.md`** — the `untap_choice` prompt and how the heuristic scores it.
- **`docs/protocol.md`** — the new `pending_choice.kind`.
- **`docs/engine-seams.md`** — the untap rows move to Shipped.
- **ADR 0058** — a dated amendment retiring Decision 8's first bullet and
  pointing here.

## Decision 9 — Out of scope, stated

- **The linked-object restriction** (Decision 6). Designed, not built.
- **Exert** (CR 701.43) and **Telekinesis'** "next two untap steps", still where
  ADR 0058 Decision 8 left them.
- **A cap on another player's untap during your step** — not a limitation but a
  reading (Decision 4). If a future card really says "during each untap step",
  it declares a cap whose `Applies` ignores the active player and the solver is
  unchanged; what would have to move is the prompt's chooser.
- **Smoke and Stoic Angel** ("creatures don't untap … unless their controller
  sacrifices a creature"; "nonartifact creatures don't untap") are not caps, they
  are restrictions with a cost. They wait on a restriction that can charge.
- **Dovin Baan's emblem** waits on nothing in this ADR and on the emblem
  catalogue path (ADR 0064).

## Consequences

- The untap step is the third turn-based action that can stop for a player
  (after the CR 616 replacement window, #710, and the cleanup discard), and the
  first one that stops for a *choice the rules give*. The pattern is now
  explicit enough to copy: one exit function, a `paused` return, a continuation
  that calls the exit.
- Nine "can't untap more than N" cards and 45 "may choose not to untap" cards
  become writable, less whatever else each of them needs.
- A player under a Winter Orb lock is asked, every turn, which land untaps. That
  is a new modal in the commonest position it could possibly be in, and it is
  the whole reason the card was held back.
- Every board with no cap and no opt-out pays one extra battlefield walk per
  untap step, guarded by two nil catalog hooks.
- Nothing about a normal untap step changes: same set, same primitive, same
  events, same recursion into the upkeep, in one lock.

## Alternatives considered

- **Untap the uncontested permanents immediately and the chosen ones on the
  answer.** Rejected: CR 502.3 untaps them *simultaneously*, and splitting the
  batch buys nothing — the `EventUntapCard` listeners queue triggers onto
  `PendingTriggers` either way and they reach the stack at the same upkeep. What
  it costs is a board state the rules never produce, observable by anything that
  runs between the two batches.
- **Reuse `choose_cards` and give the heuristic a way to tell the prompts
  apart.** Rejected in Decision 2: the only thing it could tell them apart by is
  the source card, and that is a card-name special case in the bot, which is the
  one thing the design constraints forbid. The redaction rule would still be
  wrong.
- **Two prompts: "choose N" then "confirm each opt-out".** Rejected: two pauses,
  two continuations and two bot surfaces for one sentence of CR 502.3. A board
  with Winter Orb and a Rust Tick asks one question.
- **A cap as a layer-6 `Restriction` bit.** Rejected for ADR 0058 Decision 1's
  three reasons, all of which still hold, plus a fourth: a bit cannot carry N.
- **Resolve the caps for the player and show the result.** That is option (b) of
  ADR 0058's question 3, which the owner rejected on 2026-09-17. It is what the
  bots do, because a bot must always have an answer, and it is not what a human
  is offered.
- **A `Min` equal to the true minimum legal size.** It would let the client's
  submit button be exactly right instead of nearly right. Rejected: computing it
  is the same set-packing problem as the true maximum, `Validate` already refuses
  the short answer with a message the client shows (#624), and `Min ≥ 1` is all
  that is actually load-bearing (it is what stops the enumerator's `AlwaysLegal`
  empty answer from walking past a mandatory untap).

## Test plan

- The step pauses only when it must: a board with no cap and no opt-out untaps
  in one call and advances to the upkeep; one tapped land under Winter Orb does
  not prompt; two do.
- One cap: Winter Orb over four lands offers all four, `Min == Max == 1`, and
  the chosen land untaps while the other three stay tapped — and the creatures,
  which no cap counts, untap without being offered.
- Composing caps: Winter Orb + Static Orb + Winter Moon over the worked
  example's board; every legal set accepted, "two lands", "two nonbasic lands"
  and the short answer refused with `ErrChoiceSetRejected`.
- Opt-out: a tapped Rust Tick is offered with `Min == 0`; choosing nothing
  leaves it tapped, choosing it untaps it.
- Mixed: Winter Orb plus a tapped Rust Tick — `Min == 1`, and the legal sets are
  one land with or without the Tick.
- ADR 0058 interaction: a restricted permanent and a marked permanent are not
  offered and do not untap; a stunned permanent that IS chosen loses a stun
  counter and stays tapped.
- The pause is a real pause: `advance_step` is refused with `ErrChoicePending`
  while the prompt is open, the cursor is still on `untap`, and the upkeep
  trigger has not fired. Answering advances to the upkeep in the same call.
- Undo across the prompt, and `CaptureSnapshot().Restorable()` false while it is
  open.
- The enumerator offers the active seat only the prompt's answers and every one
  of them is accepted by `ResolveUntapChoice`; a bot table plays a turn under
  Winter Orb without wedging.
- `TestEveryChoiceKindIsClassifiedAndEnumerated` passes with the new kind.
- Each new assertion fails with the fix backed out.
