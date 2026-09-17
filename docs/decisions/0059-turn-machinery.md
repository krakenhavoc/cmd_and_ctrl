# ADR 0059 — Turn machinery: extra turns, extra phases and steps, and one turn identity

**Status:** Accepted · 2026-09-17 · unscheduled (card-coverage audit, wave 2) · tracked on [#753](https://github.com/krakenhavoc/cmd_and_ctrl/issues/753). The owner's answers to the open questions are recorded in [Decided (2026-09-17)](#decided-2026-09-17).
**Numbering:** 0052 is reserved for the emblems ADR
([#623](https://github.com/krakenhavoc/cmd_and_ctrl/issues/623)), and
0056-0058 are being drafted in parallel for
[#748](https://github.com/krakenhavoc/cmd_and_ctrl/issues/748),
[#749](https://github.com/krakenhavoc/cmd_and_ctrl/issues/749) and
[#751](https://github.com/krakenhavoc/cmd_and_ctrl/issues/751). On
2026-09-17 all 208 remote branches (after `git fetch origin --prune`)
were checked with the AGENTS.md §4 loop. None has a
`docs/decisions/0056-*` file or anything higher. The highest is 0055.
**Amends:** [ADR 0026](0026-delayed-triggers.md) ("No extra steps"),
[ADR 0022](0022-impulse-exile.md) §2 (the turn number a grant expires
on), [ADR 0035](0035-until-end-of-turn-effects.md) §3 (the "counts
ROUNDS" caveat), and the [ADR 0054 addendum](0054-dice-rolls-and-coin-flips.md#turn-scoping-uses-a-per-turn-index-turnnumbermaxplayers--activeseat)
(the RNG turn index).
**Builds on:** [ADR 0006](0006-priority-foundation.md) (the cursor and
priority), [ADR 0041](0041-game-persistence.md) and
[ADR 0044](0044-surviving-a-deploy.md) (snapshots, restore points),
[ADR 0033](0033-ai-bot-seat.md) (enumerator, bots, public log).
**Fixes along the way:** [#766](https://github.com/krakenhavoc/cmd_and_ctrl/issues/766)
(a departed active player's successor inherits the old turn's state).
Sub-PR 1 builds the single rotation seam that both issues ask for.
**Coordinates with:**
- [#717](https://github.com/krakenhavoc/cmd_and_ctrl/issues/717) (the
  real second combat damage step). This ADR designs the cursor once,
  and #717 extends it ([Decision 12](#decision-12--the-second-combat-damage-step-717-is-a-plan-entry)).
- [#755](https://github.com/krakenhavoc/cmd_and_ctrl/issues/755)
  ("until your next turn"). #755 builds its durations on the per-seat
  turn count and the turn-began hook defined here ([Decision 1](#decision-1--turn-identity-seq-round-and-turnsbegun)).
  #755 is expected to fold in the other "this turn" registries as well:
  `TurnScopedStatics`, [ADR 0057](0057-win-and-lose-by-effect.md)'s
  `TurnScopedGameEndGates` and [ADR 0045](0045-combat-restrictions.md)'s
  addendum `TurnScopedBlockRules`. Until it does, all three are swept by
  `sweepTurnEndLocked` ([Decision 6](#decision-6--one-rotation-seam)).
- [ADR 0057](0057-win-and-lose-by-effect.md) (#749, effect wins and
  losses). An effect loss by the active player during a resolution
  (Final Fortune, Last Chance, the Pact cycle) must not run this ADR's
  rotation inside the resolving callback. ADR 0057 Decision 3 defers it
  to the next SBA pass, and [Decision 6](#decision-6--one-rotation-seam)
  here states the same constraint.
- [#751](https://github.com/krakenhavoc/cmd_and_ctrl/issues/751)
  ("next untap step" marker). Nothing in this ADR is needed for it, but
  the marker must be consumed by an untap step, not by a turn change
  ([Decision 7](#decision-7--per-turn-state-across-an-extra-turn)).

Line references are to `origin/develop` at `684f2786`.

## Context

### The cursor walks a fixed list

- `turnSequence` is a fixed list of 12 steps (`server/internal/game/turn.go:41-54`).
  `Turn.advance` (`turn.go:169-197`) moves to the next index. After
  cleanup it always rotates to `(ActiveSeat + 1) % numSeats` (`:185`),
  and it bumps `Number` only when that seat is 0 (`:186-189`).
- `advanceCursorLocked` (`game.go:699-712`) is documented as "the single
  seam every step transition goes through". It calls `advance`, skips
  eliminated seats, and calls `onTurnAdvanceLocked` (`game.go:720-755`).
- Nothing can add a phase, a step or a turn. ADR 0026 records this as a
  known limit (`0026-delayed-triggers.md:206-209`), and Y'shtola Rhul
  ships without its extra end step because of it
  (`cards/effects/yshtola_rhul.go:15-20`).

### "Is this a new turn?" compares seats

`Turn.IsNewTurn` is `t.ActiveSeat != next.ActiveSeat` (`turn.go:158-160`),
and `onTurnAdvanceLocked` returns early unless it is true (`game.go:721-723`).
If a seat takes two turns in a row, the second turn gets no
`layerVersion` bump, and `LoyaltyActivatedThisTurn`,
`SpellsCastThisTurn`, `LandsPlayedThisTurn`, `DrawnThisTurn` and
`TurnTally` (with its `Resolved` / `Triggered` once-per-turn gates and
`FirstEvent`, `turn_tally.go:60-93`) are never cleared.

### `Turn.Number` counts rounds, and its readers disagree about what it means

The doc comment on `Turn` says it is "the turn number since the game
started" (`turn.go:93-96`), and ADR 0022 §2 says an expiry turn number is
"correct under extra turns". But `Number` goes up only when the cursor
wraps to seat 0, so it counts **rounds**. The non-test readers
(38 lines, plus 113 lines in tests) fall into four groups:

| group | readers | what they want |
|---|---|---|
| **identity of this turn** | `ExilePlayPermission.Active` / `UntilTurn` / `NotBeforeTurn` (`exile_play.go:162-173`, `:201`, `:224`, `:302`; `mutations.go:514`, `:549`, `:1384`, `:1399`; `cascade.go:181`, `:215`; `batch17_helpers.go:442`; `batch32_helpers.go:442`), warp's floor (`alternative_cost.go:627`), `ScopedStatic.ExpiresAfterTurn` (`turn_scoped_statics.go:141`, `:170`), `DelayedTrigger.CreatedTurn` (`delayed.go:137`), the RNG index (`rng.go:110`), `EventStepBegan.Amount` (`game.go:937`), which feeds the log's `turn` and the reveal window (`protocol/reveal_frame.go:185-213`), the heuristic's per-turn memory (`aiseat/heuristic/heuristic.go:312`, `threat.go:87`, `:123`, `concede.go:47`) | a value that changes on **every** turn |
| **"your Nth turn" / "your next turn"** | Starting Town (`starting_town.go:23-33`), the "until the end of your next turn" grant stamped `Number + 2` and re-stamped by a delayed trigger (`batch19_helpers.go:103-121`, `:161`; `batch20_helpers.go:222`), Teferi, Time Raveler's unwired +1 (`teferi_time_raveler.go:51`) | a **per-seat** turn count |
| **the first turn of the game** | the CR 103.8a draw skip (`game.go:1021`) | "is this turn 1" |
| **display** | `protocol/view.go:1373` (`TurnView.number`), `PhaseDisplay.svelte:125` (`T{turn.number}`), `replay.ts:58`, `bugReport.ts:164`, `ws.ts:549`, `aiseat/model/prompt.go:146`, `cmd/gamecli/main.go:208`, `ws/persist.go:260`, the log's "Turn %d" text (`protocol/log.go:694`) | a number people read |

Several files already work around the round count in comments
(`turn_scoped_statics.go:93-99`, `batch25_helpers.go:78-82`,
`batch33_helpers.go:93`). Two readers are wrong today, before any extra
turn exists (see [Bugs found while drafting](#bugs-found-while-drafting)):
Starting Town ignores whose turn it is, and the reveal window keeps a
reveal for the rest of the **round**, not the turn its doc promises
(`docs/protocol.md:433`).

### Two hand-written rotations (#766)

`eliminatePlayerLocked` (`mutations.go:2479`) calls
`advancePastEliminatedLocked` (`mutations.go:4954-4997`), which writes the
next seat's `Untap` into `g.Turn` by hand and runs only
`runStepEntryHooksLocked` (`:4993`). It never reaches
`onTurnAdvanceLocked`, never runs the cleanup body (`game.go:1044-1090`)
and never calls `clearCombatLocked` (`mutations.go:4909`). #766's probe
shows the departed player's attacker dealing damage again in the next
player's combat. `PassTurn` (`mutations.go:5005`) reaches
`onTurnAdvanceLocked` but also skips the cleanup body.

### What the rules ask (CR, Aug 7 2026)

- **500.7** Extra turns are added directly after the specified turn,
  one at a time; the most recently created is taken first.
- **500.8** Extra phases are added directly after the specified phase;
  the most recently created occurs first. **500.9** does the same for
  steps. **500.10a**: "you get" an additional step or phase on a turn
  other than your own adds nothing.
- **505.1a** Only the first main phase of a turn is precombat. Every
  other main phase is postcombat. **505.1b** "First main phase", "second
  main phase" count main phases of this turn.
- **506.1** "There are two combat damage steps" with first or double
  strike (#717).
- **302.6** Summoning sickness is about control "continuously since
  their most recent turn began", not about the untap step.
- **614.10a** Anything scheduled for a skipped turn won't happen.
- **800.4j** A player who leaves during their turn: the turn continues
  without an active player. **800.4k** A turn of a player who has left
  doesn't begin. **800.4m** An effect lasting until that player's next
  turn lasts until that turn *would have begun*.

Scryfall rulings that constrain the design (checked 2026-09-17):

- Relentless Assault (2025-09-19): "creates an additional combat and
  main phase only if it resolves during a main phase". Cast in an
  opponent's main phase, the opponent attacks again.
- Full Throttle (2025-02-07): no main phase between the two combats; off
  a main phase there are no additional combat phases.
- Karlach, Fury of Avernus (2022-06-10) and Aurelia, the Warleader
  (2024-11-08): no main phase is added; end of combat goes straight to
  the next beginning of combat.
- World at War (2010-06-15): if the anchoring main phase has ended, no
  phases are created. A second copy inserts "after the original
  postcombat main phase and before the newest combat phase".
- Final Fortune (2004-10-04): "If you end up skipping the extra turn
  that is gained, you do not lose the game."
- Time Stretch, Time Warp, Magistrate's Scepter: the most recently
  created extra turn is taken first.
- Sphinx of the Second Sun (2020-11-10): "until your next turn" effects
  don't expire during the additional beginning phase.
- Medomai the Ageless (2013-09-15): an extra turn is any turn created by
  a spell or ability, whoever controls Medomai.

### The cards

From the audit on #753: extra combats are the only core blocker for 28
cards and one of the core blockers for 38. Extra turns: 14 and 25. The
cards named "ready once this lands" are the first-wave candidates in
[Card first wave](#card-first-wave). Cards also blocked elsewhere
(Combat Celebrant: exert; Port Razer and Bloodthirster: a per-target
attack restriction; Walk the Aeons: buyback, #664; Nexus of Fate: a
shuffle replacement from any zone; Temporal Trespass: delve) stay with
their other seams.

## Decision 1 — Turn identity: `Seq`, `Round` and `TurnsBegun`

```go
type Turn struct {
    Seq            int   // 1 on the game's first turn; +1 every time a turn begins, extra turns included
    Round          int   `json:"Number"` // was Number; same meaning, see below
    ActiveSeat     int
    PriorityHolder int
    Phase          Phase
    Step           Step
    Extra          bool  // this turn was created by an effect (CR 500.7)
    ExtraRef       int   // the ExtraTurn.Ref this turn came from; 0 for a normal turn
    OrderSeat      int   // the seat whose NORMAL turn this is, or last was; rotation continues from here
    PhaseID        int   // identity of the phase in progress, unique within the turn (Decision 3)
    PhaseOrdinal   int   // nth phase of this family this turn (Decision 8)
    StepOrdinal    int   // nth occurrence of this step this turn (Decision 8)
}

type Player struct {
    // ...
    TurnsBegun int // turns this seat has begun, or would have begun after leaving (CR 800.4m)
}
```

- **`Seq` is the identity of a turn.** It is 0 only before the first
  turn (the opening shuffle), 1 on the first turn, and goes up by exactly
  one each time a turn begins. `IsNewTurn` becomes
  `t.Seq != next.Seq`. Every reader in the Context's "identity" group
  moves to it.
- **`Number` is renamed `Round`, and keeps its meaning.** It still
  counts rounds: it goes up when normal rotation moves to a seat index
  lower than or equal to the previous normal turn's seat, which is
  today's "wrapped to seat 0" (including when seat 0 has left). Extra
  turns never change it. The struct tag keeps the snapshot key
  `"Number"` (`Turn` is embedded by value in `GameSnapshot`,
  `snapshot.go:139`, and has no tags today, so the key is the Go field
  name).
- **Why rename rather than redefine `Number`.** Redefining the field in
  place compiles cleanly and silently changes every reader. Renaming it
  makes the compiler list every non-test reader (38 lines, comments
  included), and each one has to be
  moved to `Seq`, `Round` or `TurnsBegun` on purpose
  ([Decision 2](#decision-2--every-reader-is-moved-on-purpose)). The six
  existing workarounds show that a field named "Number" that means
  "round" gets misread. The 113 test lines are a mechanical rename: they
  compare against rounds on purpose.
- **`TurnsBegun` is the per-seat count.** It goes up for the active seat
  when its turn begins, normal or extra. It also goes up when a turn of a
  seat that has left would have begun and is skipped instead (CR 800.4k,
  both a normal rotation and a queued extra turn). So "until your next
  turn" ends at the right moment for a departed player (CR 800.4m). A
  turn skipped by an effect ("skip your next turn", CR 614.10a) does not
  count; no catalog card does that today.
- **One hook for #755.** `onTurnBeganLocked(seat, skipped bool)` runs
  when `TurnsBegun` changes. The per-turn resets (Decision 7) run only
  when `skipped` is false. #755's "until your next turn" sweep registers
  here and runs in both cases. This ADR adds the hook; #755 adds the
  sweep.

## Decision 2 — Every reader is moved on purpose

| reader | becomes |
|---|---|
| `ExilePlayPermission.UntilTurn`, `NotBeforeTurn`, `Active(player, turn)`, the cleanup sweep, `cascade.go`, `batch17/32_helpers.go` | renamed `UntilSeq`, `NotBeforeSeq`; `Active(player, seq)`; stamped from `Turn.Seq` |
| warp's floor (`alternative_cost.go:627`) | `Seq + 1`. CR 702.185a says "after the current turn has ended", which is the next turn of any seat, not the next round |
| `ScopedStatic.ExpiresAfterTurn` | renamed `ExpiresAfterSeq`, stamped and swept on `Seq`. #755 replaces it with a duration |
| [ADR 0057](0057-win-and-lose-by-effect.md)'s `ScopedGameEndGate.Seq`, if ADR 0057's sub-PR 3 lands first | already named `Seq`, stamped from `Turn.Number` until this sub-PR; re-stamped from `Turn.Seq` here. #755 replaces it with a duration |
| `DelayedTrigger.CreatedTurn` | renamed `CreatedSeq` (wire `created_seq`) |
| RNG index (`rng.go:110`) | `Turn.Seq` ([Decision 9](#decision-9--rng-turn-scoping-uses-seq)) |
| `EventStepBegan.Amount` | `Seq`. A new `Event.Round int` (omitempty) carries the round for display |
| reveal window (`reveal_frame.go:213`) | compares `Seq`, so "this turn" means this turn |
| heuristic per-turn memory | `TurnView.seq` |
| CR 103.8a draw skip (`game.go:1021`) | `Seq == 1` (the first turn of the game; `StartingSeat` stays in the condition) |
| Starting Town | "it's your turn and your `TurnsBegun` ≤ 3". This also fixes the bug below. The caveat is removed |
| "until the end of your next turn" grant (`batch19_helpers.go:121`, `batch20_helpers.go:222`) | keeps its re-stamp trigger; the backstop becomes `Seq + 2*MaxPlayers`. If other players' extra turns push the controller's next turn past it, the grant lapses early: weaker, never stronger. #755's duration replaces both halves |
| comments (`turn_scoped_statics.go:93-99`, `batch25_helpers.go:78`, `batch33_helpers.go:93`, `teferi_time_raveler.go:51`, the `Turn` doc) | rewritten |
| display (`TurnView.number`, `PhaseDisplay`, replay, bug report, model prompt, gamecli, persist log) | `Round`, unchanged ([decided](#decided-2026-09-17), question 1): "Turn N" is the round, and an extra turn shows the same number with an "Extra turn" mark from `Turn.Extra` |

`game.TurnsBegunFor(player uuid.UUID) int` and
`game.IsExtraTurn() bool` are the card-side accessors.

## Decision 3 — The turn plan: the rest of the turn is data

The fixed index walk is replaced by a list of the steps still to come in
this turn:

```go
type PlannedStep struct {
    Step    Step
    PhaseID int // which phase instance this step belongs to
}

// on Game, carried by clone and snapshot
TurnPlan      []PlannedStep // steps after the current one, in order; empty at cleanup
ExtraTurns    []ExtraTurn   // Decision 5; the LAST element is taken next
NextPhaseID   int           // per turn; the template's five phases are 1..5
NextExtraRef  int           // per game; mints ExtraTurn.Ref
```

- **When a turn begins**, `TurnPlan` is filled from `turnSequence` after
  `Untap`, with phase ids 1-5 for beginning, precombat main, combat,
  postcombat main and ending. `TurnSequence()` stays as the template and
  its tests keep passing.
- **Advancing** pops the head of `TurnPlan` into `Turn.Step`,
  `Turn.Phase` (from `PhaseOf`) and `Turn.PhaseID`. When `TurnPlan` is
  empty (after cleanup), the next turn begins through the rotation seam
  ([Decision 6](#decision-6--one-rotation-seam)).
- **Why on `Game` rather than inside `Turn`.** `Turn` is copied by value
  everywhere (`prev := g.Turn`, the wire view, event payloads), and a
  slice inside it would alias between copies. Keeping `Turn` all scalars
  keeps those copies safe.
- **Why a materialised list rather than "fixed sequence plus inserts".**
  With inserts beside a fixed index, "most recent first after the same
  anchor", phases added after an inserted phase, and #717's extra damage
  step all become special cases. With a list, each is one splice. The
  list is at most a few dozen entries.
- **Skipped steps (CR 500.11) are unchanged.** The step-transition
  replacement window still cancels a step on entry, and the cursor pops
  the next entry. Removing the declare blockers and combat damage steps
  when nobody attacks (CR 508.8) could be a plan edit too, but it is not
  part of this ADR.

## Decision 4 — Extra phases and steps (CR 500.8-500.10a, 505.1a)

```go
type PhaseKind string // "beginning", "main", "combat", "ending"

type PhaseAnchor struct {
    Kind AnchorKind // AnchorThisPhase | AnchorThisMainPhase | AnchorNthMainPhase
    N    int        // for AnchorNthMainPhase (World at War: 2)
}

// AddPhasesForEffect adds phases, in the printed order, directly after
// the anchor phase of the CURRENT turn. It returns the new phases' ids
// in the order they will occur, or nil when nothing was added: the
// anchor is not the current phase kind (Relentless Assault resolving
// outside a main phase), or the anchor phase has already ended
// (World at War after the second main phase). Emits EventPhasesAdded.
func (g *Game) AddPhasesForEffect(source uuid.UUID, anchor PhaseAnchor, kinds ...PhaseKind) []int

// AddStepAfterCurrentForEffect adds one step directly after the step in
// progress, in the same phase (CR 500.9; Y'shtola Rhul's end step).
func (g *Game) AddStepAfterCurrentForEffect(source uuid.UUID, step Step) bool
```

- **Where the splice goes.** The insertion point is the first
  `TurnPlan` entry whose `PhaseID` differs from the anchor phase's id.
  New phases go there, ahead of anything already inserted after the same
  anchor. That gives CR 500.8's "most recently created occurs first"
  with no extra rule. It also gives World at War's ruling (a second copy
  goes after the original postcombat main and before the newest combat)
  and Sphinx of the Second Sun's example (a combat added later in the
  same main phase happens before the beginning phase added earlier).
- **What a kind expands to.** `combat` gives the five combat steps;
  `main` gives **`StepPostcombatMain`** (CR 505.1a: an added main phase
  is never the first); `beginning` gives untap, upkeep, draw; `ending`
  gives end, cleanup. Each added phase gets a fresh `PhaseID`.
- **Anchors.** `AnchorThisPhase` is "after this phase" (Aurelia,
  Karlach, Hellkite Charger). `AnchorThisMainPhase` is "after this main
  phase" and adds nothing unless the current phase is a main phase
  (Relentless Assault, Seize the Day, Aggravated Assault and Full
  Throttle rulings). `AnchorNthMainPhase` counts main phases per CR
  505.1b across the phases already begun and those still planned. It
  adds nothing if that phase has ended.
- **Whose turn.** The phases go on the current turn, whoever controls
  the source, as the Relentless Assault and Seize the Day rulings say.
  CR 500.10a ("you get" on someone else's turn adds nothing) is checked
  by the card, which knows its own wording. No first-wave card says "you
  get".
- **Consequences that fall out of the existing step hooks:**
  - The Saga lore action is on `StepPrecombatMain` (`game.go:951-959`),
    so it does not run again in an added main phase (CR 714.3c / 505.1a).
    `EventBeginPrecombatMain` doesn't fire again either.
  - `AtYourPostcombatMain` triggers fire in every postcombat main phase,
    which is what Sphinx of the Second Sun's ruling says.
  - Sorcery timing (`legal/legal.go:371`, `mutations.go:1448`),
    attack declaration (`mutations.go:4136`, `legal/combat.go:38`) and
    `clearCombatLocked` at end of combat all key on the step name, so a
    second combat works with no changes there.
  - Delayed triggers "at the beginning of the next end step" fire at the
    next `StepEnd` entry, including an added end step. That is correct:
    it is the next end step.

## Decision 5 — Extra turns (CR 500.7)

```go
type ExtraTurn struct {
    Ref    int       // minted from NextExtraRef; stable identity for "that turn"
    Seat   int
    Source uuid.UUID // for the log
}

// TakeExtraTurnsForEffect gives `player` n extra turns directly after
// the current turn. Returns the refs, the one taken first first.
// Emits one EventExtraTurnAdded per turn.
func (g *Game) TakeExtraTurnsForEffect(player, source uuid.UUID, n int) []int
```

- **The queue is a stack.** Adding pushes onto the end of `ExtraTurns`,
  and the rotation seam pops from the end. "After this one" puts the new
  turn directly after the current turn, ahead of turns queued earlier,
  and CR 500.7's "most recently created first" is the stack order. A
  Time Warp cast during the first of Time Stretch's two turns is taken
  before the second.
- **Every printed extra turn is "after this one".** The API has no
  anchor. If a card ever names another turn, the anchor is added then.
- **Normal rotation resumes from `OrderSeat`.** An extra turn leaves
  `OrderSeat` unchanged. When the stack is empty, the next normal seat is
  the next seat after `OrderSeat` still in the game. For example, seat 0
  casts Time Warp on seat 2 in a four-player game. Seat 2's extra turn
  follows seat 0's turn, and then seat 1 takes its normal turn.
- **A departed player's queued turn is dropped** as it would begin
  (CR 800.4k). It still counts toward `TurnsBegun` (CR 800.4m;
  Decision 1), and delayed triggers bound to it are dropped
  ([Decision 8](#decision-8--phase-and-combat-identity-bound-delayed-triggers-attack-history)).
- **In APNAP order** (CR 500.7, several players at once): the effect calls
  the API once per player in APNAP order. No first-wave card does this.
- **Loops.** Every extra turn in the first wave costs a card or three
  charge counters, so an extra-turn loop always has a player decision in
  it. The loop breaker ([ADR 0055](0055-loop-breaker.md)) counts within
  a turn and is not extended here.

## Decision 6 — One rotation seam

```go
// sweepTurnEndLocked is the non-interactive part of cleanup (CR 514.2),
// moved out of the StepCleanup case unchanged: damage and deathtouch
// marks, ClearTurnScopedReplacementsLocked,
// ClearExpiredTurnScopedStaticsLocked, clearExpiredExilePlayLocked.
// Every later "this turn" registry sweeps here too:
// ClearExpiredTurnScopedGameEndGatesLocked (ADR 0057 Decision 4) and
// the TurnScopedBlockRules sweep (ADR 0045 addendum), whichever of them
// has landed.
func (g *Game) sweepTurnEndLocked()

// beginNextTurnLocked picks the next turn, stamps the cursor and runs
// the turn-began hook. It does not run step entry hooks; callers do,
// as they do today.
func (g *Game) beginNextTurnLocked()
```

`beginNextTurnLocked` does, in order:

1. Pop `ExtraTurns` from the end. Drop an entry whose seat has left, and
   run `onTurnBeganLocked(seat, skipped=true)` for it. The first entry
   whose seat is still in the game is the next turn (`Extra = true`,
   `ExtraRef = Ref`).
2. If the stack is empty, the next seat after `OrderSeat` still in the game
   takes a normal turn. Each departed seat passed over gets
   `onTurnBeganLocked(seat, true)`. `Round` goes up if the rotation
   wrapped. `OrderSeat` becomes the new seat.
3. `Seq++`. The cursor is set to that seat's `Untap`, with
   `PhaseID = 1` and ordinals reset. `TurnPlan` is refilled from the
   template, which also drops any added phases the old turn didn't
   reach. `NextPhaseID` is reset.
4. `onTurnBeganLocked(seat, false)`: `TurnsBegun++`, the per-turn
   resets (Decision 7), then `EventTurnBegan` (`Actor`, `Amount = Seq`,
   `Label = "extra"` for an extra turn).

The callers:

- **`advanceCursorLocked`** pops `TurnPlan`, or calls
  `beginNextTurnLocked` when it is empty. The eliminated-seat loop
  (`game.go:704-710`) moves into step 2.
- **`advancePastEliminatedLocked`** keeps its priority-holder branch.
  When the active seat has left, it runs `clearCombatLocked`, then
  `sweepTurnEndLocked`, then `beginNextTurnLocked`, then the step entry
  hooks. This fixes #766: the resets run, damage and "until end of turn"
  effects end, and attackers leave combat. After an SBA loss it runs once
  repeated SBA passes have settled, before the waiting triggers are
  drained. A lord dying in one pass can make another creature lethally
  damaged on the next; the sweep must not clear that damage between
  checks (ADR 0057 Decision 2).
- **`PassTurn`** does the same `clearCombatLocked` and
  `sweepTurnEndLocked` before `beginNextTurnLocked`. The discard to
  hand size is skipped, as today. This is a sandbox verb, and the skip is
  stated in its doc comment.
- **An effect loss by the active player during a resolution**
  ([ADR 0057](0057-win-and-lose-by-effect.md) Decision 3: Final Fortune
  and Last Chance's end-step trigger, the Pact cycle's decline). This is
  the fourth way to reach the seam, and it must **not** call
  `advancePastEliminatedLocked` itself. Beginning the next turn runs the
  untap step, the per-turn resets, step entry hooks and upkeep triggers,
  and none of that may run inside a callback that is still resolving.
  `effects.LoseTheGame` leaves the game at once (stack and prompt
  cleanup), returns `game.ErrStopResolution` when the loser controls the
  resolving object, and sets `Game.ActiveSeatLeftPending`. The next SBA
  loss pass (the resolution bookend's sweep, or a later prompt answer's)
  consumes the flag: it runs ADR 0057's `checkGameOverLocked()` and, if
  the game goes on, `advancePastEliminatedLocked` as above.

**The constraint, for every caller.** `beginNextTurnLocked` and the step
entry hooks never run inside a resolving callback, a replacement's
`Apply`, or a prompt continuation that is still part of a resolution.
They run from the cursor advance, from an SBA pass, or from a player
action (`Concede`, `PassTurn`). A new caller that can be reached
mid-resolution defers the same way ADR 0057 does.

**CR 800.4j** (the departed player's turn continues without an active
player): **decided** ([question 4](#decided-2026-09-17), option (a)).
The turn ends at once through the seam, as a declared simplification
of CR 800.4j. The rest of that turn does not happen, so "at the
beginning of the end step" triggers and delayed triggers scheduled for
it wait for the next player's turn. When the departure is an effect loss
during a resolution, "at once" means at the next SBA pass (above). The
departed player's objects are #769's problem, not this ADR's.

## Decision 7 — Per-turn state across an extra turn

An extra turn is a turn. Everything "this turn" or "once each turn"
resets when it begins, and nothing resets when an added phase or step
begins.

- **`onTurnBeganLocked(seat, false)`** runs today's
  `onTurnAdvanceLocked` body unchanged: the `layerVersion` bump, the four
  per-turn maps, and `resetTurnTallyLocked`, which covers `Resolved`,
  `Triggered`, `EnteredSubtypes`, `LoopRun`, `FirstEvent` and
  `LoopNotice`. Because it keys on `Seq`, a second consecutive turn now
  resets. Planeswalker loyalty (CR 606.3), land drops, "first spell each
  turn" and every `b11TriggeredThisTurn` gate open again on an extra
  turn, as printed.
- **Two things move from the untap step to the turn-began hook:**
  - Clearing `SummonedThisTurn` for the active seat's permanents
    (`untap.go:326-330`). CR 302.6 is about when the turn began. With the
    clear in the untap step, Sphinx of the Second Sun's extra untap step
    would cure summoning sickness mid-turn (stronger than printed), and a
    skipped untap step leaves creatures sick for a whole turn. That
    second case is a live bug under Stasis today
    ([Bugs found](#bugs-found-while-drafting)).
  - The `UndosRemaining` refresh (`game.go:996`), for the same reason:
    the budget is "per turn" (`ws/room.go:92`).

  For an ordinary turn nothing happens between the turn beginning and
  its untap step, so moving these changes nothing a player can see.
- **Nothing resets on an added combat.** CR 506-511 have no "this combat"
  tallies. A card that says "each combat" reads the combat identity from
  Decision 8.
- **#751's "next untap step" marker** must be consumed when an untap
  step of that player actually happens, not when a turn begins. Sphinx
  of the Second Sun's ruling says permanents untap in the added untap
  step, and a skipped untap step is not an untap step (CR 614.10a).

## Decision 8 — Phase and combat identity, bound delayed triggers, attack history

**Ordinals.** When a step begins (after the CR 500.11 skip window, so a
skipped step is not counted), `StepOrdinal` is the count of that step
begun this turn. When the first step of a phase begins, `PhaseOrdinal`
is the count of phases of that family begun this turn. Precombat and
postcombat main are one family, "main", per CR 505.1b. Card helpers:

- `IsFirstCombatPhase(g)`: `Phase == PhaseCombat && PhaseOrdinal == 1`
  (Karlach, Genji Glove, Raiyuu).
- `IsFirstStepOfItsKind(g)`: `StepOrdinal == 1` (Y'shtola Rhul's "first
  end step of the turn").

**Binding a delayed trigger to a turn or a phase.**

```go
type DelayedTrigger struct {
    // ... existing fields
    OnExtraTurn int // fire only in the extra turn with this Ref; 0 = any turn
    OnPhaseID   int // fire only in this phase of turn OnSeq; 0 = any phase
    OnSeq       int
}
```

- Final Fortune / Last Chance: `At: StepEnd, OnExtraTurn: ref`. The
  trigger is dropped when that extra turn is dropped or ends without
  reaching its end step, which is the ruling ("If you end up skipping the
  extra turn … you do not lose the game"). When it resolves, it calls
  ADR 0057's `effects.LoseTheGame` and returns its error. The loser is
  the trigger's controller, so that is `game.ErrStopResolution`, and the
  rotation waits for the resolution bookend's SBA pass (Decision 6).
- Moraug and World at War ("at the beginning of that combat"):
  `At: StepBeginCombat, OnSeq: g.Turn.Seq, OnPhaseID: id` using the id
  returned by `AddPhasesForEffect`. The trigger is dropped at the next
  turn change. The `OnPhaseID` binding ships with the first of those two
  cards (see [Card first wave](#card-first-wave)).
- `fireDelayedTriggersLocked` (`delayed.go:170`) checks the binding
  beside `ControllerTurnOnly`. `beginNextTurnLocked` sweeps bindings
  that can no longer match.

**Attack history.** `TurnTally` gains
`Attacks []AttackRecord{Attacker uuid.UUID, EnteredAt int64, Defender uuid.UUID, PhaseID int}`,
appended by `turnTallyListener` on `EventAttack` and reset with the
tally. `EnteredAt` is the attacker's `EnteredBattlefieldAt`, so a
flickered creature is a new object (CR 400.7). Accessors:
`AttackedThisTurn(card)`, `TimesAttackedThisTurn(card)`,
`AttackedPlayersThisTurn(card)`. This is in scope because "untap all
creatures that attacked this turn" is in four first-wave candidates
(Relentless Assault, Full Throttle, World at War, Waves of Aggression)
and Moraug's anthem counts attacks. Today those would each walk the log
(`b18AttackedThisTurn`, `batch18_helpers.go:75`). The **restriction**
"can't attack a player it has already attacked this turn" (Port Razer,
Bloodthirster) is an [ADR 0045](0045-combat-restrictions.md) per-target
restriction. It is not built here.

## Decision 9 — RNG turn scoping uses `Seq`

`rngTurnIndexLocked` (`rng.go:109-111`) returns `g.Turn.Seq`. The ADR 0054
addendum used `Number*MaxPlayers + ActiveSeat` only because `Number`
counted rounds. That formula gives an extra turn the same index as the
turn before it (same round, same seat), so what an undo taught a player
would survive into their extra turn. `Seq` changes on every turn, which
is the property the addendum wanted. The opening shuffle still runs at
index 0.

- **Seeded test outcomes change once more.** Draws after the first turn
  begins are keyed on 1, 2, 3… instead of 4, 5, 8, 9 (two players) or 4,
  5, 6, 7, 8 (four). Sub-PR 1 re-pins and lists them.
- **Restored games keep their counters** ([Decision 10](#decision-10--clone-snapshot-and-games-saved-before-this-change)):
  a pre-change snapshot restores with `Seq` equal to its old index, so
  `rngTurn` still matches.
- The ADR 0054 addendum gets a one-line pointer to this decision.

## Decision 10 — Clone, snapshot, and games saved before this change

- **Clone** deep-copies `TurnPlan` and `ExtraTurns`, and copies
  `NextPhaseID`, `NextExtraRef` and `Player.TurnsBegun`. `Turn` stays a
  value copy. The drift test (`snapshot_drift_test.go:66-`) marks every
  new field `carried`.
- **Snapshot.** All of it is pure data. `GameSnapshot` gains
  `turnPlan`, `extraTurns`, `nextPhaseId`, `nextExtraRef`, and
  `playerSnapshot` gains `turnsBegun`. `delayedTriggerSnapshot` gains the
  three binding fields.
- **Stamped values change unit, so their keys change too.** `UntilSeq`,
  `NotBeforeSeq`, `CreatedSeq` and `ExpiresAfterSeq` are new JSON keys.
  After a rollback, an older binary reads a newer snapshot and sees
  zeros. Grants then lapse at the next cleanup (weaker) and warp's floor
  disappears (a warped card can be recast one turn early) until the next
  deploy. Keeping the old keys with new units would be worse: an old
  binary would compare turn counts against round numbers and keep grants
  alive for many rounds.
- **Restoring a snapshot written before this change** (no `Seq` in
  `turn`), with no `SnapshotSchemaVersion` bump (ADR 0041 Decision 5):
  - `Seq = Round*MaxPlayers + ActiveSeat`, which is exactly the RNG index
    the game was already using. `Seq` stays strictly increasing from
    there, so every comparison below holds. It is not a dense count, which
    matters only for display, and display doesn't use `Seq`.
  - `OrderSeat = ActiveSeat`, `Extra = false`, `ExtraTurns = nil`.
  - `TurnPlan` is rebuilt from the template after the current step (no
    old game can have added phases). Phase ids are 1-5 and ordinals are 1.
  - `TurnsBegun` for each seat: `Round` for seats up to and including
    the active seat, `Round - 1` for later seats.
  - Old `UntilTurn = r`: if `r <= Round`, the grant ends at this turn's
    cleanup, so `UntilSeq = Seq`, which is exactly what the old sweep did.
    If `r > Round`, `UntilSeq = r*MaxPlayers + MaxPlayers - 1` (the end
    of round `r`). Old `NotBeforeTurn = r`: `0` if `r <= Round`, otherwise
    `r*MaxPlayers` (the start of round `r`). `CreatedTurn = r` becomes
    `r*MaxPlayers`.
  - [ADR 0057](0057-win-and-lose-by-effect.md)'s turn-scoped gates, if
    that ADR's sub-PR 3 lands first. An entry's `Seq` was stamped from
    `Turn.Number`. A stamp equal to the current `Round` is from this
    turn (earlier turns' entries were swept at their cleanup), so it
    becomes this turn's `Seq` and still ends at this turn's cleanup. An
    older stamp is dropped.
  - `TurnScopedStatics` hold closures and are already in the census
    (`snapshot.go:487-489`), so no restorable snapshot has any.
- **Test:** the committed fixture `testdata/snapshot_pre683.json` still
  restores and plays to the next turn. A pre-change snapshot holding an
  impulse grant and a warp floor restores with the converted bounds.

## Decision 11 — Wire and client

`TurnView` (`protocol/view.go:1283`) gains:

| field | meaning |
|---|---|
| `seq` | `Turn.Seq` |
| `extra` (omitempty) | this is an extra turn |
| `phase_id`, `phase_ordinal` (omitempty) | Decision 8 |
| `upcoming` (omitempty) | `[{step, phase_id}]`, the rest of `TurnPlan`. Public: the turn structure is public information |
| `extra_turns` (omitempty) | seat indices of queued extra turns, next first |

`number` keeps its current meaning, the round (decided, question 1).
`CardView.exile_play.not_before_turn` becomes `not_before_seq`, and `DelayedTriggerView.created_turn` becomes
`created_seq`. A tab left open across the deploy sees no floor, offers a
cast the server refuses, and is fixed by reloading.
`LogEvent.turn` carries `Seq`, and `LogStep` entries gain `round`.
`docs/protocol.md` and `client/src/lib/protocol.ts` follow.

Client changes that are not presentation choices:

- `zoneBrowser.logic.ts` `impulseGrantFor` compares `turn.seq` with
  `not_before_seq`.
- `combatBeats.ts` `blocksOn` groups block entries by the combat step
  they happened in (the `LogStep` seq, as ADR 0053's "Beats across frames"
  already identifies a step) rather than by `turn`, so blocks from one
  combat never pair with damage from another.
- Per-step stops (`settings.ts`) stay keyed by step name. An added
  combat's declare attackers stops where the first one did.
- The autopass safety belt (`Game.svelte:233-241`) keys on entering the
  viewer's own `precombat_main`, which every turn, extra or not, has once.

**How the board shows it** ([decided](#decided-2026-09-17), questions 1
and 2):

- **"Turn N" is the round**, as today. An extra turn shows the same
  number with an **"Extra turn"** mark on the turn bar, from
  `TurnView.extra`.
- **The phase strip** labels a repeated phase "Combat 2" / "Main 3" from
  `phase_ordinal` and renders added phases from `upcoming`. Queued extra
  turns show as "Next: Alice (extra)" from `extra_turns`.
- **Log lines** for each added turn and each added phase
  (`EventExtraTurnAdded`, `EventPhasesAdded`). These change the turn's
  structure, which the owner's log rule covers. Per-object annotations
  are not logged.
- **No reveal-strip cue** when an extra turn or phase is created.

## Decision 12 — The second combat damage step (#717) is a plan entry

This ADR does not build #717, but the plan is designed so #717 is a
splice and not a second cursor design:

- #717 adds `Sub string` to `PlannedStep` and `Turn` (`"first_strike"`,
  `"regular"`, the tags ADR 0053 already puts on damage events). On
  entry to a `StepCombatDamage` with first or double strike present, the
  entry hook splices a second `StepCombatDamage` with `Sub: "regular"`
  directly after the current step (CR 510.4). That uses the same
  insertion `AddStepAfterCurrentForEffect` uses.
- Both damage steps count toward `StepOrdinal` for `combat_damage`. No
  card reads that ordinal, and `PhaseOrdinal` / `PhaseID` are unaffected.
- The client's step list and stops treat the second step as another
  `combat_damage`, which #717's own scope already lists.

If #717 lands first, it builds the plan as described here, and this
ADR's sub-PR 1 builds on it.

## Decision 13 — Enumerator, bots and the catalog soak

- **Enumerator** (`internal/legal`): nothing counts combats or turns.
  Sub-PR 3 adds agreement tests for a second declare attackers step and
  for an extra turn's per-turn gates.
- **Heuristic**: per-turn memory moves to `seq`
  (`heuristic.go:312`, `threat.go:87,123`, `concede.go:47`). The
  heuristic already attacks whenever it is in declare attackers, so an
  added combat is used without changes. Activating Aggravated Assault
  again is left to its normal scoring.
- **Model prompt** (`aiseat/model/prompt.go:146`): the turn line keeps
  the round as "Turn N" (question 1) and says "(extra turn)" and
  "combat 2" when they apply, so a model tier does not misread an added
  combat as a repeat of the first.
- **Catalog soak** (`aiseat/catalog_soak_test.go:336`): the turn budget
  stays on `Round`, plus a second cap of
  `Seq <= turnBudget * seats * 2`. A runaway extra-turn or extra-combat
  line is then reported as a stall (per the soak's rules), not left to
  hang until the wall clock. The first-wave cards join the soak pool.

## Decision 14 — Out of scope, stated

- **CR 800.4j in full.** The owner chose the declared simplification
  (Decision 6, question 4): the departed active player's turn ends at
  once. The departed player's objects are #769.
- **Controlling another player's turn** (CR 723: Mindslaver, Emrakul,
  the Promised End).
- **"Skip your next turn"** and skipping steps of an extra turn (Savor
  the Moment). The step-transition replacement can express the latter,
  but binding it to "that turn" is a card-PR question.
- **The CR 514.3a extra cleanup step.** It could be a plan splice, but
  [ADR 0035](0035-until-end-of-turn-effects.md)'s simplification stands.
- **Skipping declare blockers and combat damage when nobody attacks**
  (CR 508.8).
- **Per-target attack restrictions** (Port Razer, Bloodthirster).
- **"Until your next turn" durations** (#755) and **"next untap step"**
  (#751). Both use this ADR's identity, but neither is built here.

## Consequences

- **Two consecutive turns by one seat reset per-turn state.** This is
  a correctness change that only extra turns can reach. It is covered by
  tests.
- **#766 is fixed** by sub-PR 1: a departed active player's successor
  starts clean.
- **Starting Town and the reveal window change behaviour today.**
  Starting Town enters tapped on an opponent's turn, as printed. The
  public reveal window empties at each turn change instead of each
  round.
- **Stasis stops leaving creatures summoning sick,** and a skipped
  untap step no longer skips the undo refresh.
- **A warped card becomes castable on the next turn of any seat,** not
  the next round, as CR 702.185a says. It matters only with a flash
  enabler.
- **Seeded tests are re-pinned once** (Decision 9), and the 113 test
  lines using `Turn.Number` are renamed.
- **The rollback window degrades exile grants** (Decision 10), in the
  weaker direction except warp's floor.
- **`Turn` gains seven scalar fields and `Game` four.** The plan adds up
  to about 30 small structs per clone. That cost is negligible next to
  the zones.
- **Card caveats need re-checking once extra combats exist.** Aurelia,
  the Law Above (`CompletenessFull`, but its comment says it is weaker
  under an extra combat, `aurelia_the_law_above.go:31-33`), Chivalric
  Alliance, Firemane Commando, Breena, the Demagogue and
  `b16PlayerAttackedWithAtLeast` (`batch16_helpers.go:195-197`) each
  either gain a caveat or move their dedup to a per-combat key
  (`PhaseID`) in the card follow-up.
- **An active player who leaves ends their turn at once** (declared
  simplification of CR 800.4j). At a table of three or more, other
  players' "at the beginning of each end step" triggers skip that one
  turn. An effect loss during a resolution moves the turn on at the
  resolution bookend, not inside the resolution.
- **Nothing a player sees about turn numbers changes** in a game without
  extra turns. An extra turn carries the round's number and an "Extra
  turn" mark.
- **`docs/engine-seams.md`**: the "Extra turns primitive" and "Extra
  combat and main phases" rows (`:97`, `:112`) move to Closed when
  sub-PR 2 lands.

## Alternatives considered

- **Redefine `Number` as the turn count and keep its name.** It
  compiles, so nothing forces a reader to be re-examined, and the display
  and every test assertion change meaning silently. Rejected for the
  rename (Decision 1).
- **Keep `Number` and derive identity as `Number*MaxPlayers + ActiveSeat`**
  (the RNG addendum's index). An extra turn by the same seat in the same
  round gets the same value. Rejected.
- **Compare seats plus a "consecutive" flag in `IsNewTurn`.** It fixes
  the reset but leaves every stamped value ambiguous. Rejected.
- **Insert phases beside a fixed index** (`turnSequence` plus an
  "inserted after phase X" list). "Most recent first", phases added after
  an added phase, and #717's split step each need their own rule.
  Rejected for the plan (Decision 3).
- **Store the plan inside `Turn`.** `Turn` is copied by value in dozens
  of places, and a slice would alias. Rejected.
- **A queue of extra turns keyed by "after turn Seq N".** No printed card
  anchors anywhere but "after this one", and a stack gives CR 500.7's
  order directly. Rejected as unneeded.
- **Model CR 800.4j fully in sub-PR 1.** Priority and step ends without
  an active player touch every active-player-keyed rule. It is left to
  the owner (question 4).
- **Bump `SnapshotSchemaVersion` instead of converting old stamps.** A
  rollback would abandon every live game (ADR 0041 Decision 5), and the
  conversion is exact for every value an old game can hold. Rejected.
- **Keep the RNG index formula.** Explained in Decision 9. Rejected.
- **Walk the event log for attack history** (the current
  `b18AttackedThisTurn` pattern). It is O(events) per query, and Moraug's
  anthem runs on every layer recompute. Rejected for the tally.

## PR split

**Sub-PR 1 — turn identity and the rotation seam (fixes #766). No new
cards, no added phases.**
- `Turn.Seq`, the `Number` → `Round` rename, `OrderSeat`, `Extra`,
  `ExtraRef`; `Player.TurnsBegun`; `IsNewTurn` on `Seq`.
- Every reader in Decision 2, and the unit renames of the stamped fields.
- `sweepTurnEndLocked` (naming ADR 0057's and ADR 0045's turn-scoped
  sweeps if they have landed), `beginNextTurnLocked`, `onTurnBeganLocked`;
  `advancePastEliminatedLocked` and `PassTurn` through the seam;
  `EventTurnBegan`. If ADR 0057's sub-PR 2 has landed, its
  `ActiveSeatLeftPending` consumer goes through the same seam, and a
  test pins that no turn begins inside a resolution.
- Summoning sickness and the undo refresh move to the turn-began hook.
- The RNG index on `Seq`, and the re-pinned seeded tests listed in the PR.
- Snapshot fields, restoring pre-change snapshots, drift test rows.
- `TurnView.seq`, `not_before_seq`, `created_seq`, log `turn`/`round`,
  reveal window, heuristic memory, client `impulseGrantFor`.
- Starting Town's fix, and its caveat removed.
- Docs: `protocol.md`, the ADR 0022 / 0026 / 0035 / 0054 pointers.

**Sub-PR 2 — the plan, extra phases and steps, extra turns.**
- `TurnPlan`, `PlannedStep`, `NextPhaseID`, ordinals; `advanceCursorLocked`
  pops the plan.
- `AddPhasesForEffect` (anchors `AnchorThisPhase` and
  `AnchorThisMainPhase`), `AddStepAfterCurrentForEffect`,
  `TakeExtraTurnsForEffect`, `ExtraTurns`; `EventPhasesAdded`,
  `EventExtraTurnAdded`, and their log lines.
- Delayed triggers bound to an extra turn (`OnExtraTurn`);
  `TurnTally.Attacks` and its accessors.
- `TurnView.extra`, `phase_id`, `phase_ordinal`, `upcoming`,
  `extra_turns`; `combatBeats.ts` grouping.
- Card-side helpers: `ExtraCombatAfterThisPhase`, `ExtraCombatAndMainAfterThisMain`,
  `TakeExtraTurn`, `IsFirstCombatPhase`, `CreaturesThatAttackedThisTurn`.

**Sub-PR 3 — bots and soak.** Heuristic, model prompt, enumerator
agreement tests, soak caps, and the first-wave cards added to the soak pool.

**Sub-PR 4 — client presentation** (questions 1 and 2, Decision 11):
the "Extra turn" mark beside the round, "Combat 2" / "Main 3" labels,
added phases from `upcoming`, "Next: Alice (extra)". No reveal-strip
cue.

**Card PRs** — the first wave, option (b) (question 3; the list is in
[Card first wave](#card-first-wave)), the caveat re-checks above, and
removing Time Stretch from the batch 28 skip list
(`batch28_test.go:170`). The engine parts that only World at War or
Moraug use (`AnchorNthMainPhase` and the `OnPhaseID` binding) ship in
the first of those card PRs, not in sub-PR 2, so every seam path lands
with a real card. Each card follows AGENTS.md: completeness declared,
caveats weaker than printed and never stronger.

Sub-PR 1 can merge on its own and is useful without the rest (#766, the
three live bugs below). Sub-PRs 2-4 depend on it in order.

## Test plan

Engine (`internal/game`):

1. **Consecutive turns reset.** A seat takes an extra turn: every map in
   `onTurnBeganLocked`, `TurnTally` (including `Resolved` / `Triggered`)
   and `LoyaltyActivatedThisTurn` is empty at the extra turn's upkeep,
   and `layerVersion` went up.
2. **`Seq`, `Round`, `TurnsBegun`** over a four-seat game with an extra
   turn for seat 2 after seat 0, and seat 3 conceded: order
   0, 2 (extra), 1, 2, 0…; `Round` goes up only on the wrap; seat 3's
   `TurnsBegun` still goes up each round.
3. **CR 500.7 order.** Time Stretch (two turns) and, during the first of
   them, Time Warp: the Warp turn comes before the second Stretch turn,
   and then normal rotation resumes after the original seat.
4. **CR 800.4k.** A queued extra turn for a player who concedes is
   dropped. A delayed trigger bound to it is dropped (Final Fortune's
   loss never happens).

   **Effect loss in the active player's own extra turn.** Three seats,
   Final Fortune's end-step trigger resolves: `LoseTheGame` returns
   `ErrStopResolution`, the player is eliminated at once, no step entry
   hook or upkeep trigger runs before the resolution returns, and the
   bookend's SBA pass rotates once to the next normal seat after
   `OrderSeat`. The same holds for a Pact declined in its controller's
   upkeep.
5. **#766 (all four of its tests).** Four-seat concede in combat damage:
   caches empty at the next seat's upkeep, no `DamageMarked`, no stale
   `AttackingTarget` / `BlockingTarget`, the next player takes no damage
   in its own combat. A turn-scoped prevention shield and a turn-scoped
   static registered on the departed player's turn are gone. An SBA
   elimination takes the same path.
6. **`PassTurn` sweeps**: marked damage and until-end-of-turn effects
   end.
7. **CR 500.8 order.** Two "after this phase" combats from the same
   phase run newest first. Relentless Assault in precombat main gives
   main, combat, main (postcombat), combat, main (postcombat), ending.
   Resolving outside a main phase adds nothing.
8. **World at War anchors** (in World at War's card PR). Two copies
   cast in the precombat main both
   insert after the second main phase, and the pair created by the second
   copy comes first (the ruling's "before the newest combat phase"). A
   copy cast in the third main phase adds nothing.
9. **CR 505.1a.** The Saga lore action and `EventBeginPrecombatMain`
   happen once. `AtYourPostcombatMain` fires in every added main phase.
10. **Ordinals.** `IsFirstCombatPhase` is true only in the first combat.
    Karlach's "first combat" trigger does not re-add a third combat.
11. **Bound delayed trigger.** Final Fortune's end-step loss fires in the
    extra turn's end step, not the current one. With World at War's card
    PR: "at the beginning of that combat" fires in the added combat and
    not in a later one.
12. **Summoning sickness.** Under Stasis, a creature that entered on its
    controller's previous turn can attack. With an added beginning
    phase, a creature cast this turn stays sick.
13. **Attack history.** A creature attacking in two combats has
    `TimesAttackedThisTurn == 2`. A flickered creature starts at 0.
14. **RNG.** The same stream gives different values on a turn and on the
    extra turn that follows it. Undo within an extra turn rewinds.
15. **Snapshot.** Round-trip with a non-empty plan, an added phase in
    progress, queued extra turns and a bound delayed trigger. Restore a
    pre-change snapshot: `Seq` equals the old RNG index, the plan is the
    template tail, grants and the warp floor convert as Decision 10
    says, and the game plays into the next turn. `snapshot_pre683.json`
    still restores.
16. **Starting Town** enters tapped on an opponent's turn in round 1 and
    untapped on its controller's third turn.
17. **Reveal window** is empty at the next seat's turn.

Room (`internal/ws`):

18. Undo across `TakeExtraTurnsForEffect` and `AddPhasesForEffect`
    restores the queue and the plan.

Enumerator and bots:

19. `legal`: declare-attackers moves are offered in an added combat, and
    `dispatchAll` accepts them. A planeswalker can be activated again in
    an extra turn.
20. `aiseat`: a bot game with Relentless Assault and Time Warp in the
    pool finishes without stalls (catalog soak, `AISEAT_GAME_TESTS=1`).

Client (vitest):

21. `impulseGrantFor` with `seq` / `not_before_seq`. `blocksOn` groups by
    combat step: two combats in one turn don't pair a block from the
    first with damage from the second.

## Card first wave

Decided: option (b) of [question 3](#decided-2026-09-17). Oracle text
checked against the Scryfall dump on 2026-09-17. Each card PR still
checks for gaps outside this seam, and a card found to be blocked
elsewhere drops out rather than shipping stronger than printed.

**Extra turns**
- **Time Warp** {3}{U}{U}: "Target player takes an extra turn after this one."
- **Time Stretch** {8}{U}{U}: "Target player takes two extra turns after this one." (remove from `batch28_test.go:170`)
- **Temporal Manipulation** {3}{U}{U} and **Capture of Jingzhou** {3}{U}{U}: "Take an extra turn after this one."
- **Magistrate's Scepter** {3}: "{4}, {T}: Put a charge counter on this artifact. / {T}, Remove three charge counters from this artifact: Take an extra turn after this one."
- **Final Fortune** {R}{R} (instant) and **Last Chance** {R}{R} (sorcery): "Take an extra turn after this one. At the beginning of that turn's end step, you lose the game." The loss uses [ADR 0057](0057-win-and-lose-by-effect.md)'s `effects.LoseTheGame`, which returns `game.ErrStopResolution` (the loser controls the trigger) and defers the rotation to the next SBA pass. These two cards wait for ADR 0057's sub-PR 2; before it, `LoseTheGameForEffect` (`effect_api.go:818`) defers the whole loss to the SBA flag, which ADR 0057 shows is the wrong timing.

**Extra combats**
- **Relentless Assault** {2}{R}{R}: "Untap all creatures that attacked this turn. After this main phase, there is an additional combat phase followed by an additional main phase."
- **Seize the Day** {3}{R}: "Untap target creature. After this main phase, there is an additional combat phase followed by an additional main phase. / Flashback {2}{R}"
- **Aggravated Assault** {2}{R}: "{3}{R}{R}: Untap all creatures you control. After this main phase, there is an additional combat phase followed by an additional main phase. Activate only as a sorcery."
- **Full Throttle** {4}{R}{R}: "After this main phase, there are two additional combat phases. / At the beginning of each combat this turn, untap all creatures that attacked this turn."
- **Karlach, Fury of Avernus** {4}{R}: "Whenever you attack, if it's the first combat phase of the turn, untap all attacking creatures. They gain first strike until end of turn. After this phase, there is an additional combat phase."
- **Hellkite Charger** {4}{R}{R}: "Flying, haste / Whenever this creature attacks, you may pay {5}{R}{R}. If you do, untap all attacking creatures and after this phase, there is an additional combat phase."
- **Aurelia, the Warleader** {2}{R}{R}{W}{W}: "Flying, vigilance, haste / Whenever Aurelia attacks for the first time each turn, untap all creatures you control. After this phase, there is an additional combat phase."

**Caveat re-checks:** Aurelia, the Law Above; Chivalric Alliance;
Firemane Commando; Breena, the Demagogue; `b16PlayerAttackedWithAtLeast`;
Starting Town (sub-PR 1).

**Option (b) adds (chosen):**
- **Y'shtola Rhul** (caveat removal): "… Then if it's the first end step of the turn, there is an additional end step after this step."
- **Sphinx of the Second Sun** {6}{U}{U}: "Flying / At the beginning of each of your postcombat main phases, there is an additional beginning phase after this phase."
- **World at War** {3}{R}{R}: "After the second main phase this turn, there's an additional combat phase followed by an additional main phase. At the beginning of that combat, untap all creatures that attacked this turn. / Rebound". **Drops out if rebound isn't supported** (owner). On `develop` at the time of this decision it isn't: the engine has no rebound keyword or cast-from-exile trigger, only a `"rebound"` counter in a zone test (`zone_test.go:178`). World at War ships when rebound does, and it brings `AnchorNthMainPhase` and the `OnPhaseID` binding with it (or Moraug brings the binding first).

Not in any option: Medomai the Ageless ("can't attack during extra
turns" is easy with `Turn.Extra`, but it is not an audit-ready card),
Moraug (per-object anthem plus bound trigger; a good second-wave card),
Port Razer and Bloodthirster (per-target restriction), Combat Celebrant
(exert), Walk the Aeons (#664), Nexus of Fate, Savor the Moment,
Mindslaver.

## Bugs found while drafting

Each can ship in sub-PR 1 or on its own. None is filed yet.

1. **Stasis leaves creatures summoning sick.** `performUntapStepLocked`
   clears `SummonedThisTurn` (`untap.go:326-330`), and a canceled untap
   step never reaches it (`game.go:904-911`). Probe on `684f2786`: Stasis
   on the battlefield, a creature with `SummonedThisTurn` on seat 0's
   turn 1; at seat 0's round-2 upkeep, `HasSummoningSickness` is still
   true. Stasis is `CompletenessFull` (`stasis.go:34`).
2. **Starting Town enters untapped on an opponent's turn** in rounds
   1-3. Its predicate is `g.Turn.Number <= 3` and ignores the controller
   (`starting_town.go:32-34`), but the card says "your first, second, or
   third turn". Reachable with any instant-speed "put a land onto the
   battlefield". The declared caveat understates this.
3. **The public reveal window lasts a round, not a turn.** It compares
   the `EventStepBegan.Amount` round with `g.Turn.Number`
   (`reveal_frame.go:185-213`), but `docs/protocol.md:433` says "the
   reveals that happened this turn". In a four-player game, seat 0's
   reveal stays in the frame through seats 1-3's turns.
4. **Warp's floor is a round, not a turn** (`alternative_cost.go:627`,
   `Number + 1`). CR 702.185a: castable "after the current turn has
   ended". With a flash enabler, a warped card exiled on seat 1's turn
   can't be cast until seat 0's next turn. Weaker than printed.

## Decided (2026-09-17)

The owner answered questions 1, 3 and 4 on 2026-09-17. Question 2
follows from those answers and the owner's cross-cutting display policy
for these seams, given the same day. The options are kept as they were
proposed. The chosen one is marked.

1. **What number does the table see as "Turn N"?**
   - (a) The round, as today (`T3` for every seat's third turn). An extra
     turn shows the same number with an "extra turn" mark. **Chosen.**
   - (b) The game's turn count (every turn, extra turns included; a
     four-player game reaches `T12` in round 3).
   - (c) Both: `R3 · T11`.

   **Decision: (a)**, as recommended. "Turn N" shows the round, with an
   "Extra turn" mark on an extra turn. Commander tables count "turn 3" as
   everyone's third turn, and nothing a player sees changes in games
   without extra turns. `seq` is still on the wire for tools and the bug
   report. Applied in Decisions 2, 11 and 13.

2. **How does the board show extra turns and extra phases?**
   - (a) Minimal: an "Extra turn" badge on the turn bar, the phase strip
     labels a repeated phase "Combat 2" / "Main 3" and renders added
     phases from `upcoming`, and queued extra turns show as "Next: Alice
     (extra)". Log lines for each added turn or phase. **Chosen.**
   - (b) (a) plus a cue on the reveal strip (`RevealBanner`) when an
     extra turn or phase is created.
   - (c) Log lines only, with the turn bar unchanged.

   **Decision: (a).** This follows from the owner's answers rather than a
   separate one: question 1's "Extra turn" mark rules out (c), the policy
   "no reveal-strip cues for these events" rules out (b), and the log
   rule ("log changes to the game's outcome and turn structure, not
   per-object annotations") keeps (a)'s log lines for added turns and
   phases. Players need to know an extra combat is coming before they
   pass priority out of main phase, so the turn bar has to show the plan.
   Applied in Decision 11 and sub-PR 4.

3. **Which cards ship in the first wave?**
   - (a) The seven extra-turn cards and seven extra-combat cards above,
     plus the caveat re-checks.
   - (b) (a) plus Y'shtola Rhul, Sphinx of the Second Sun and World at
     War, which exercise an added step, an added beginning phase and an
     "Nth main phase" anchor with a bound trigger. **Chosen.**
   - (c) The seam only, with cards left to the batch issues.

   **Decision: (b)**, as recommended, and World at War drops out if
   rebound isn't supported. Without these cards, the step insertion, the
   beginning-phase expansion and the phase-bound delayed trigger ship
   with no real card using them, like the `AllowStop` argument in ADR
   0054. The owner's cross-cutting policy makes that a rule: every new
   seam path ships with at least one real card, even with a declared
   weaker caveat. Rebound isn't supported on `develop` today, so
   `AnchorNthMainPhase` and the `OnPhaseID` binding move out of sub-PR 2
   and ship with World at War (or Moraug) instead. See
   [Card first wave](#card-first-wave) and the PR split.

4. **When the active player leaves mid-turn (CR 800.4j), what happens
   to the rest of their turn?**
   - (a) It ends at once: combat is cleared, the cleanup sweep runs, and
     the next player's turn begins. Declared as a simplification. #766's
     bug is fixed either way. **Chosen.**
   - (b) It plays out with no active player. Priority goes to the next
     player in turn order, and the remaining steps and the end step
     (with its triggers) still happen.

   **Decision: (a)**, as recommended: the rest of the turn ends at once
   through the rotation seam, a declared simplification of CR 800.4j.
   The departures this covers are a concede, an SBA loss, and an effect
   loss during a resolution (Final Fortune and Last Chance's end-step
   trigger, the Pact cycle's decline, via
   [ADR 0057](0057-win-and-lose-by-effect.md)'s `effects.LoseTheGame`).
   For the last, "at once" is the next SBA pass, because the rotation
   never begins a turn inside a resolving callback (Decision 6, ADR 0057
   Decision 3). (b) touches every rule keyed on the active
   player (priority start, "your turn" checks, the end-step triggers of
   other players' cards). It only matters at tables of three or more,
   and only in the rest of one turn. The visible difference is that other
   players' "at the beginning of each end step" triggers skip that one
   turn.
