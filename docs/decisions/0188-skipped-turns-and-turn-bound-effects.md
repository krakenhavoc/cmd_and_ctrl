# ADR 0188 — Turn machinery, the rest: skipped turns, effects bound to one turn or one combat, and showing the turn

Date: 2026-10-14
Status: Proposed
Sprint: none (unscheduled card-coverage seam, #753; outside docs/sprints.md)
Issues: #753 (turn machinery). Registry rows: `extra-turns` (Ugin's Nexus, Savor the Moment, Magosi, the Waterveil, Lethal Vapors, Alchemist's Gambit) and `extra-combats` (World at War; Moraug, Fury of Akoum).
Builds on: [ADR 0059](0059-turn-machinery.md) (turn identity, the rotation seam, the extra-turn queue, the turn plan, and its amendments of 2026-09-30, 2026-10-05 and 2026-10-07), [ADR 0041](0041-game-persistence.md) phase 3 (ScopedEffect records, unknown-key refusal), [ADR 0063](0063-durations-and-control.md) (durations), [ADR 0178](0178-life-instructions-after-a-replaced-move.md) (instructions after a replaced move), [ADR 0126](0126-bots-that-play-their-decks.md) §6 (Purpose).
Amends: ADR 0059 Decision 14 ("skip your next turn" and skipping a step of an extra turn were out of scope) and its Decisions 11 and 13 (the presentation and bot work not yet shipped).
Numbering: 0188 was reserved for this ADR. Every remote branch was listed after `git fetch --all --prune` (`git log --remotes --name-only -- docs/decisions/`): 0185, 0186, 0187 and 0189 are taken on other branches (0189 is #3528's face-down ADR), and nothing uses 0188.

## Context

### What already shipped

ADR 0059 built most of #753. On `develop` today:

| Piece | Where | Status |
|---|---|---|
| `Turn.Seq`, `Turn.Round`, `Player.TurnsBegun`, `Turn.OrderSeat`, `Turn.Extra` / `ExtraRef` | `game/turn.go`, `game/rotation.go` | shipped (sub-PR 1) |
| One rotation seam: `beginNextTurnLocked`, `startTurnLocked`, `onTurnBeganLocked`, `sweepTurnEndLocked` | `game/rotation.go` | shipped |
| Extra-turn stack `Game.ExtraTurns`, `TakeExtraTurnsForEffect`, `queueAnchoredExtraTurnLocked` (Emrakul) | `game/extra_turns.go` | shipped (2a) |
| `DelayedTrigger.OnExtraTurn` and `sweepUnreachableBoundTriggersLocked` | `game/delayed.go` | shipped (2a) |
| Skipping a queued EXTRA turn: `RepEventExtraTurn`, `extraTurnSkippedLocked`, `effects.SkipOpponentsExtraTurns` (Trouble in Pairs) | `game/extra_turns.go`, `effects/skip_extra_turns.go` | shipped (#2529) |
| The turn plan: `Game.TurnPlan`, `AddPhasesForEffect` (`AnchorThisPhase`, `AnchorThisMainPhase`), `AddStepAfterCurrentForEffect`, ordinals, `TurnTally.Attacks` | `game/turn_plan.go`, `game/turn_tally.go` | shipped (2b) |
| `UntilEndOfCombat` keyed on `Seq` + `PhaseID`; "each combat this turn" delayed triggers | `game/duration.go`, `game/delayed.go` | shipped |
| Ending the turn (CR 724.1), CR 508.8 | `game/end_turn.go`, `stepExistsLocked` | shipped |
| Wire: `TurnView.seq`, `extra`, `extra_turns`, `phase_id`, `phase_ordinal`, `upcoming`; log `extra_turn`, `extra_turn_skipped`, `extra_phase`, `turn_ended` | `protocol/view.go`, `protocol/log.go` | shipped |
| Board: "Extra turn" mark, "Next: X (extra)" | `client/src/lib/components/board/PhaseDisplay.svelte` | shipped |
| Model prompt and MCP seat: "(extra turn)" | `aiseat/boardtext/boardtext.go` | shipped |

Full in the catalog: Time Warp, Time Stretch, Temporal Manipulation, Capture of Jingzhou, Temporal Mastery, Magistrate's Scepter, Final Fortune, Last Chance, Warrior's Oath, Expropriate, Nexus of Fate, Teferi, Master of Time, Avatar Kuruk, Trouble in Pairs, Relentless Assault, Seize the Day, Aggravated Assault, Full Throttle, Hellkite Charger, Aurelia, the Warleader, Sphinx of the Second Sun, Éomer, Marshal of Rohan and Y'shtola Rhul. Karlach carries an unrelated caveat.

### What is left

Five engine gaps, and the presentation and bot work ADR 0059 planned but did not ship:

1. **Skipping a player's next turn** (CR 614.10). Nothing can skip a NORMAL turn, and nothing can skip "the next turn you would take" when that is an extra turn (the Magosi ruling).
2. **Skipping every player's extra turns, and "instead exile it and take an extra turn"** (Ugin's Nexus). The replacement is Trouble in Pairs' without the "opponent" test. The second ability adds an extra turn to a replaced zone change, and a Replace callback must not emit events (ADR 0178).
3. **Things bound to one extra turn**: "skip the untap step of that turn" (Savor the Moment) and "during that turn, damage can't be prevented" (Alchemist's Gambit). No duration starts in a future turn.
4. **"After the second main phase this turn"** (World at War): ADR 0059 Decision 4's `AnchorNthMainPhase`, not built.
5. **"At the beginning of that combat"** (Moraug, World at War): ADR 0059 Decision 8's `OnPhaseID` binding, not built. Rebound has shipped since ADR 0059 (`KeywordRebound`, ADR 0107), so World at War waits only on 4 and 5.
6. **Presentation and bots.** The phase strip does not label "Combat 2" / "Main 3" or show added phases. The model prompt does not say "combat 2". No `legal` agreement test covers a second declare attackers step. The heuristic's `nextTurnSeat` (`aiseat/heuristic/heuristic.go:1108`) rotates from `ActiveSeat`, so during seat 2's extra turn after seat 0's it names seat 3, not seat 1. `Purpose` has `ExtraCombat` but nothing for an extra or a skipped turn.

### The waiting cards (oracle text from the Scryfall dump)

Registry `extra-turns`:

- **Ugin's Nexus** {5}, Legendary Artifact: "If a player would begin an extra turn, that player skips that turn instead. / If Ugin's Nexus would be put into a graveyard from the battlefield, instead exile it and take an extra turn after this one."
- **Savor the Moment** {1}{U}{U}, Sorcery: "Take an extra turn after this one. Skip the untap step of that turn."
- **Magosi, the Waterveil**, Land: "This land enters tapped. / {T}: Add {U}. / {U}, {T}: Put an eon counter on this land. Skip your next turn. / {T}, Remove an eon counter from this land and return it to its owner's hand: Take an extra turn after this one."
- **Lethal Vapors** {2}{B}{B}, Enchantment: "Whenever a creature enters, destroy it. / {0}: Destroy this enchantment. You skip your next turn. Any player may activate this ability."
- **Alchemist's Gambit** {1}{R}{R}, Sorcery: "Cleave {4}{U}{U}{R} … / Take an extra turn after this one. During that turn, damage can't be prevented. [At the beginning of that turn's end step, you lose the game.] / Exile Alchemist's Gambit."

Registry `extra-combats`:

- **World at War** {3}{R}{R}, Sorcery: "After the second main phase this turn, there's an additional combat phase followed by an additional main phase. At the beginning of that combat, untap all creatures that attacked this turn. / Rebound"
- **Moraug, Fury of Akoum** {4}{R}{R}, Legendary Creature — Minotaur Warrior 6/6: "Each creature you control gets +1/+0 for each time it has attacked this turn. / Landfall — Whenever a land you control enters, if it's your main phase, there's an additional combat phase after this phase. At the beginning of that combat, untap all creatures you control."

Not on any registry row, Commander-legal, and expressible once gap 1 or 2 is closed (the `extra-turns` row's `Printed` pattern does not match "skip your next turn"): Meditate ("Draw four cards. You skip your next turn."), Chronatog, Chronatog Totem, Eater of Days ("you skip your next two turns"), Wormfang Manta, Waterspout Elemental, Chronosavant, Unstable Hulk, Ral Zarek, Guest Lecturer ("Target opponent skips their next X turns"), and Gerrard's Hourglass Pendant ("If a player would begin an extra turn, that player skips that turn instead."). Stranglehold also skips extra turns but waits on "your opponents can't search libraries". Kang the Conqueror's "during that turn" waits on power-up. Medomai the Ageless ("can't attack during extra turns") needs a conditional can't-attack static, which is not this seam.

### Rules (Comprehensive Rules effective September 25, 2026)

- **CR 500.7**: extra turns are added directly after the specified turn; "The most recently created turn will be taken first."
- **CR 500.8**: added phases go directly after the specified phase; the most recently created occurs first.
- **CR 500.11**: "To skip a step, phase, or turn is to proceed past it as though it didn't exist."
- **CR 505.1a / 505.1b**: only the first main phase is precombat; "Phrases such as 'first main phase,' 'second main phase,' and so on count the number of main phases that have occurred only in the current turn."
- **CR 614.1b / 614.10**: "skip" effects are replacement effects; "Once a step, phase, or turn has started, it can no longer be skipped—any skip effects will wait until the next occurrence."
- **CR 614.10a**: "Anything scheduled for a skipped step, phase, or turn won't happen. Anything scheduled for the 'next' occurrence of something waits for the first occurrence that isn't skipped. If two effects each cause a player to skip their next occurrence, that player must skip the next two."
- **CR 502.1–502.3**: phasing, then the day/night check, then untapping, all as turn-based actions of the untap step. **CR 731.2**: the day/night check "checks the previous turn".
- **CR 603.7b**: a delayed trigger fires once unless it has a stated duration.
- **CR 611.2a**: a resolved effect lasts as long as stated. **CR 613.7a**: it gets its timestamp when created.
- **CR 616.1**: the affected player chooses the order of replacement effects on one event.
- **CR 800.4k**: a departed player's turn doesn't begin. **CR 903.10a**: 21 combat damage from the same commander "over the course of the game".

Rulings that decide behaviour:

- Magosi (2009-10-01): "If you somehow activate Magosi's third and fourth abilities during the same turn, the effects will cancel each other out. The turn you skip is the next one you would take, which is the turn created by Magosi's fourth ability. It doesn't matter in which order the abilities resolved." Also: "you'll skip your next turn even if you don't put an eon counter on Magosi."
- Ugin's Nexus (2014-09-20): extra turns "are not skipped until they would begin, so if Ugin's Nexus leaves the battlefield before that happens, the extra turns will be unaffected"; with two Nexuses destroyed together, "all but one will be exiled and an extra turn created for each".
- Alchemist's Gambit (2021-11-19): "If you somehow skip the extra turn Alchemist's Gambit gives you or skip that turn's end step, the delayed triggered ability never triggers."
- World at War (2010-06-15): "As long as World at War resolves before the first postcombat main phase of a turn ends, it will have its full effect … if one World at War creates a second combat phase and a third main phase in a turn, then a second World at War is cast during that third main phase, no additional phases are created." Several copies insert "after the original postcombat main phase and before the newest combat phase". The untap happens "when the new combat phase created by the spell begins", for every creature that attacked this turn, in any combat.
- Moraug (2020-09-25): "If the landfall ability resolves during your precombat main phase, the additional combat phase will happen before your regular combat phase. You'll untap creatures you control at the beginning of the additional combat but not at the beginning of your regular combat." Twice in the postcombat main phase gives "two consecutive additional combat phases … followed by your ending phase". "A creature that enters the battlefield attacking hasn't attacked."

Savor the Moment and Lethal Vapors have no rulings.

## Decisions

### 1. "Skip your next turn" is a charged rules record that the rotation seam spends

A skip debt is a `ScopedEffect` record with one new mod kind:

```go
// game/scoped_effects.go
ModSkipTurns ModKind = "skipTurns" // readerRule; Mod.Player skips their next Mod.Amount turns
```

```go
// game/turn_skip.go (new)

// SkipNextTurnsForEffect records that player skips their next n turns
// (CR 614.10, 614.10a): one ScopedEffect, ScopeGame, IndefiniteDuration(),
// mod {Kind: ModSkipTurns, Player: player, Amount: n}. Emits
// EventTurnSkipAdded (Actor player, Source source, Amount n).
// Returns false for n < 1 and for a player who is not seated or has left.
// Safe inside a resolution: it only writes a record.
func (g *Game) SkipNextTurnsForEffect(player, source uuid.UUID, n int, label string) bool

// consumeTurnSkipLocked spends one charge of the OLDEST (lowest Seq)
// skipTurns record for seat, through setShieldChargeLocked (copy on
// write; a record at 0 is removed). ok is false when the seat owes none.
func (g *Game) consumeTurnSkipLocked(seat int) (source uuid.UUID, ok bool)

// TurnSkipsFor is the total charge owed by player (the view and the bots).
func (g *Game) TurnSkipsFor(player uuid.UUID) int
```

- **Where it is spent.** In the rotation seam, as a candidate turn is chosen and before it begins:
  - In `popExtraTurnLocked` (`game/extra_turns.go`), after the departed-seat check and **before** `extraTurnSkippedLocked` opens the `RepEventExtraTurn` window. A spent charge drops the entry like a skipped one (`continue`).
  - In `beginNextTurnLocked`'s normal rotation (`game/rotation.go`), in the same loop that passes departed seats: a seat that owes a skip is walked past exactly as a departed seat is (`skipped := next; skipped.Step = StepCleanup; next = skipped.advance(n, g.StartingSeat)`), except that `noteTurnBegunLocked` is **not** called.
  - Each spent charge emits `EventTurnSkipped` (Actor the player, Source the record's source, Amount the extra turn's Ref or 0, Label `"extra"` for an extra turn).
- **"Next turn" means the next turn that player would begin, extra or normal.** That falls out of spending at the seam, and it is the Magosi ruling: the extra turn from the fourth ability is the one skipped.
- **A skip created during the player's own turn waits for the next one** (CR 614.10). The current turn has already begun and never passes through the seam again.
- **Two effects skip two turns** (CR 614.10a). Each record carries its own charge; Eater of Days writes Amount 2; Ral Zarek, Guest Lecturer writes Amount X.
- **Order against an extra-turn skip.** When a player owes a skip and an extra-turn replacement (Trouble in Pairs, Ugin's Nexus) also applies to the same turn, both are replacements on one event and CR 616.1 lets the affected player choose. Spending the player's own debt first is never worse for them: the turn is skipped either way, and the debt is paid. So the seam spends the debt first and does not open the window for that turn. No prompt is needed, which matters because the window cannot pause (ADR 0059 amendment 2026-10-07 §3).
- **Termination.** Every pass through the loop spends a charge, passes a departed seat, or begins a turn. Charges are finite, so the loop ends. A loop guard of `len(g.Seats) + total charges + len(g.ExtraTurns) + 1` turns a bug into a logged fault, not a hang.
- **A departed player's records** are dropped by the seam as it passes the seat (they can never be spent; CR 800.4k).

**What a skipped turn does: nothing** (CR 500.11, 614.10a). It is not stamped on `g.Turn`, so:

| Counter / record | On a skipped turn |
|---|---|
| `Turn.Seq` | unchanged (no turn began) |
| `Turn.Round` | goes up if the walk wraps past `StartingSeat`, as it does for a departed seat: Round counts table rotations |
| `Turn.OrderSeat` | a skipped NORMAL turn moves it (rotation continues after that seat); a skipped extra turn does not |
| `Player.TurnsBegun` | unchanged, as ADR 0059 Decision 1 already says. So "until your next turn" (ADR 0063) waits for the first turn that is not skipped (CR 614.10a), and so do `EndStepTurn` windows, goad and Emrakul's turn control (`beginTurnControlsLocked` runs only in `startTurnLocked`) |
| `recordPrevTurnSpellsLocked` (day/night), `recordLastTurnAttacksLocked` | run once at the top of `beginNextTurnLocked`, before any candidate is chosen, so they record the last turn TAKEN. A skipped turn is never "the previous turn" (CR 731.2) |
| Delayed triggers | an "at the beginning of your next upkeep" trigger (rebound) waits; one bound to a skipped extra turn is swept by `sweepUnreachableBoundTriggersLocked` (Final Fortune and Alchemist's Gambit: no loss, per the ruling) |

**Effects side.** `effects.SkipYourNextTurns{N int}` (`ctx.Controller` skips: "you" is the activator on an any-player row, CR 109.5) and `effects.TargetPlayerSkipsNextTurns{N int}` (a target slot) both call `SkipNextTurnsForEffect`.

### 2. Skipping every extra turn, and an extra turn after a replaced move

- `effects.SkipExtraTurns()` is `SkipOpponentsExtraTurns()` without the `ev.Actor != src.Controller` test: "If a player would begin an extra turn, that player skips that turn instead" (Ugin's Nexus, Gerrard's Hourglass Pendant). Both build from one unexported constructor in `effects/skip_extra_turns.go`. Pure cancel, as ADR 0059's 2026-10-07 amendment requires.
- **"Instead exile it and take an extra turn after this one."** A Replace callback must not emit an event, and `TakeExtraTurnsForEffect` emits `EventExtraTurnAdded`. ADR 0178 solved the same problem for life. This ADR adds a sibling:

```go
// game/move_extra_turn.go (new), beside move_life.go
// TakeExtraTurnAfterReplacedMoveForEffect declares, from a Replace
// callback, that the modified zone change also gives player an extra
// turn after the card has moved. ErrInvalidParam for a non-move event.
func (g *Game) TakeExtraTurnAfterReplacedMoveForEffect(ev *ReplacementEvent, source, player uuid.UUID) error
```

  The instruction rides `ReplacementEvent.extraTurnsAfterMove []extraTurnInstruction{Source, Player}` (fresh backing array on every append, as `lifeAfterMove` does). `runMoveTailLocked` runs them after the card lands and **before** the life instructions, in printed order, by calling `TakeExtraTurnsForEffect(player, source, 1)`. Queuing cannot pause, so it needs none of `moveLifeSequence`'s continuation machinery. `cloneReplacementResume` (`game/clone.go`) and the zone route carry it beside `lifeAfterMove`. A move abandoned before it happens discards it (CR 614.7), as ADR 0178 does for life.
- Ugin's Nexus' second ability is a self-replacement on `RepEventMove` from the battlefield to a graveyard (the Grim Repriser shape): it redirects to exile and declares the extra turn for the Nexus' controller. Once exiled, the Nexus' first ability no longer applies, so the turn it gives is taken unless another skip applies. A CR 616 order with another graveyard replacement (Rest in Peace) is the existing prompt.

### 3. Things bound to one extra turn

#### 3a. Skipping a step of that turn (Savor the Moment)

A second rules mod kind, and two additive `Mod` fields:

```go
ModSkipExtraTurnStep ModKind = "skipExtraTurnStep" // readerRule

type Mod struct {
    // ...
    Step      Step `json:"step,omitempty"`      // the step to skip (ModSkipExtraTurnStep)
    ExtraTurn int  `json:"extraTurn,omitempty"` // the ExtraTurn.Ref it is bound to
}

// SkipStepOfExtraTurnForEffect: "skip the <step> of that turn". One
// ScopedEffect, ScopeGame, IndefiniteDuration(). Returns false, writing
// nothing, when ref names no queued turn.
func (g *Game) SkipStepOfExtraTurnForEffect(ref int, step Step, source uuid.UUID, label string) bool
```

- **Spent at step entry.** In `runStepEntryHooksLocked` (`game/game.go`), after the `stepExistsLocked` walk-through and **before** the `RepEventStepTransition` window: if a record's `ExtraTurn == g.Turn.ExtraRef` and `Step == g.Turn.Step`, remove the record (copy on write) and take the cancelled-step path (`advanceCursorLocked`, then recurse), which is what Stasis' cancel does. Nothing in the untap step happens: no phasing (CR 502.1), no day/night check (CR 502.2), no untap (CR 502.3).
- **Once.** The record is spent by the first matching step of that turn. An untap step in a beginning phase added later in the same turn (Sphinx of the Second Sun) still happens. See owner question 2.
- **Swept** with the bound delayed triggers: `sweepUnreachableBoundTriggersLocked` also removes `skipExtraTurnStep` records whose ref is neither the turn now beginning nor still queued, so a record of a skipped or finished extra turn does not linger.
- **Summoning sickness still clears.** It is cleared in `onTurnBeganLocked`, not in the untap step (ADR 0059 Decision 7), so Savor's turn lets last turn's creatures attack, as it should.
- **"Next untap step" markers wait** (CR 614.10a): `SkipNextUntapStepsForEffect`'s markers are spent only by `performUntapStepLocked`, which does not run.
- **Day/night.** The check is skipped in Savor's turn. In the turn after, CR 731.2 reads Savor's turn as "the previous turn", which `recordPrevTurnSpellsLocked` already does.

#### 3b. An effect that lasts during that turn (Alchemist's Gambit)

"During that turn, damage can't be prevented" is a CR 611.2a effect created now whose duration starts later. A duration kind that is dormant until a future turn was rejected (alternative D). Instead the effect waits on the queued turn itself:

```go
type ExtraTurn struct {
    Ref    int
    Seat   int
    Source uuid.UUID
    // Effects are rules effects (no affected objects) created by the
    // effect that queued this turn, registered as the turn begins with
    // UntilEndOfTurnDuration() and their original Timestamp.
    Effects []ScopedEffect `json:"effects,omitempty"`
}

// DuringExtraTurnRuleEffectForEffect attaches a rules effect to queued
// turn ref. scope must not be ScopeNone; the record is validated now by
// the same checks appendScopedEffectLocked runs (split out as
// buildScopedEffectLocked) and stamped with the creation timestamp
// (CR 613.7a). False when ref is not queued.
func (g *Game) DuringExtraTurnRuleEffectForEffect(ref int, sourceID uuid.UUID, scope AffectedScope, controller uuid.UUID, mods []Mod, label string) bool
```

- `startTurnLocked` registers the popped entry's `Effects` **after** `onTurnBeganLocked` (so the turn-began sweep cannot remove them), each with `Duration = g.UntilEndOfTurnDuration()` (the extra turn's player, this turn) and its original `Timestamp`, minting a `Seq` for a named mod as `appendScopedEffectLocked` does.
- **A skipped or dropped turn takes its effects with it**, because they live on the queue entry. That is CR 614.10a with no sweep.
- **Rules effects only.** An effect that modifies objects would fix its affected set when it begins (CR 611.2c), which a deferred record cannot honour, so the API takes a scope, not affected objects.
- `cloneExtraTurns` deep-copies `Effects` (with `cloneMods`), and every write to `ExtraTurns` stays copy-on-write.

### 4. "After the second main phase this turn" (World at War)

```go
const AnchorNthMainPhase AnchorKind = 3 // after AnchorThisMainPhase
type PhaseAnchor struct {
    Kind AnchorKind
    N    int // AnchorNthMainPhase: which main phase of the turn (CR 505.1b)
}
```

In `AddPhasesForEffect` (`game/turn_plan.go`):

1. `begun := g.TurnTally.PhasesBegun[PhaseKindMain]`, the main phases begun this turn, including the current one if it has begun.
2. If the current phase is a main phase and `g.Turn.PhaseOrdinal == N`, the anchor is the current phase (`g.Turn.PhaseID`), and the splice point is the first plan entry with a different phase id, as today.
3. Otherwise, if `begun >= N`, the Nth main phase has ended: add nothing, return nil (the ruling's third World at War).
4. Otherwise walk `TurnPlan`, counting distinct main-phase ids in order. The `(N - begun)`th is the anchor, and the splice point is one past its last entry. No such phase: return nil.

The splice goes ahead of anything already added after the anchor, so a second copy's combat and main come "after the original postcombat main phase and before the newest combat phase", as the ruling says. `effects.ExtraCombatAndMainAfterNthMain(n int) AddPhases` is the card-side helper. CR 505.1b counts main phases of every kind, added ones included, which is why the walk reads the plan rather than the template.

### 5. A delayed trigger bound to one phase (Moraug, World at War)

```go
type DelayedTrigger struct {
    // ...
    OnSeq     int // with OnPhaseID: the turn the phase belongs to
    OnPhaseID int // fire only in this phase instance; 0 = any
}
```

- `fireDelayedTriggersLocked` adds `(dt.OnPhaseID == 0 || (dt.OnSeq == g.Turn.Seq && dt.OnPhaseID == g.Turn.PhaseID))` beside the `OnExtraTurn` test.
- `ScheduleDelayedTriggerForEffect` refuses (returns `uuid.Nil`) a binding whose phase id is neither the current phase nor in `TurnPlan`. `sweepUnreachableBoundTriggersLocked` drops a trigger with `OnSeq != 0 && OnSeq != g.Turn.Seq` as a turn begins, so a combat that never happened (the turn was ended, CR 724.1) leaves nothing behind.
- Effects side: `ScheduleDelayedTrigger.OnPhase int` (the id `AddPhasesForEffect` returned); the effects layer stamps `OnSeq` from `g.Turn.Seq`.
- **Body keys.** World at War gets its own key, `extra-combat/untap-attackers-at-that-combat` (the same function as Full Throttle's `untapCreaturesThatAttackedThisTurn`), and Moraug gets `extra-combat/untap-all-you-control`. A new key makes an older binary refuse the file (ADR 0041 phase 3). An unbound World at War untap would otherwise fire at the regular combat, where nothing has attacked yet, and miss the added one.

### 6. Timing across extra phases and turns: what holds today, pinned by tests

No code changes. These follow from the shipped design, and the slices below pin each one with a test:

- **"Until end of turn"** ends in the cleanup step of the turn (CR 514.2), after every added phase. An added combat does not end it.
- **"Until end of combat" / "this combat"** ends with the combat phase it was made in (`UntilEndOfCombat` is keyed on `Seq` + `PhaseID`), so a pump from the first combat is gone in the second.
- **"At the beginning of the next combat"** (an unbound `At: StepBeginCombat` trigger) fires at the next combat to begin, which may be an added one: it is the next combat.
- **"At the beginning of your next upkeep"** fires in an added beginning phase's upkeep in the same turn (Sphinx of the Second Sun), and waits through a skipped turn.
- **"At the beginning of each combat this turn"** (Full Throttle) fires in every combat of the turn, added ones included, and not in an extra turn that follows.

### 7. Turn counters for "your next turn"

Unchanged from ADR 0059, now that skips exist:

- `Player.TurnsBegun` counts turns that began, and a seat passed because it left (CR 800.4m). A skipped turn does not count (Decision 1).
- `Turn.Seq` counts turns that began. `Turn.Round` counts table rotations from `StartingSeat`, skipped and departed seats included, and never moves on an extra turn.
- "Turn N" on the board stays the round (ADR 0059 owner decision 1).

### 8. Commander damage and day/night

- **Commander damage** (CR 903.10a) is cumulative for the game, and the engine keeps it per player per commander (`Player.CommanderDamage`), never per turn or per combat. Two combats in one turn add up; a skipped turn changes nothing. Moraug's slice pins this: a commander that attacks in two combats under Moraug deals 7 then 8 combat damage, and the tally is 15.
- **Day/night** (CR 502.2, 731.2): the check is part of the untap step, and it reads the last turn taken. A skipped turn is not taken (Decision 1). An untap step skipped by Savor the Moment or Stasis performs no check (Decision 3a). An extra turn is a turn, so its spells count for the next check. A beginning phase added by Sphinx of the Second Sun has an untap step, so the check runs again there, reading the previous turn: the designation does not flip twice for one turn's spells, because the first check already applied that turn's count.

### 9. Wire, log and client

**View** (`protocol/view.go`, `docs/protocol.md`, `client/src/lib/protocol.ts`):

| Field | Meaning |
|---|---|
| `PlayerView.turn_skips` (omitempty) | `TurnSkipsFor(player)`: turns this player will skip |
| `TurnView.next_seat` | The seat whose turn begins next, as far as the records can say: the extra-turn stack from the top, passing departed seats and spending skip charges in a simulation, then normal rotation from `OrderSeat`. It cannot see replacement skips (Trouble in Pairs, Ugin's Nexus), which are decided as the turn would begin. `Game.NextTurnSeatForEffect()` is read-only and safe under the read lock |

**Log** (`protocol/log.go`): `turn_skip` from `EventTurnSkipAdded` ("Bob will skip their next turn (Meditate)", "… next 2 turns"); `turn_skipped` from `EventTurnSkipped` ("Bob skips their turn (Magosi, the Waterveil)"; "Bob skips their extra turn (…)" when Label is `extra`). Both change the turn's structure, which the owner's log rule covers (ADR 0059 owner decision 2). A skipped step logs nothing, as Stasis' does not.

**Client**, completing ADR 0059 Decision 11 as the owner decided it there (option (a)):

- The phase track's current group caption reads "Combat 2" / "Main 3" when `phase_ordinal > 1` (CR 505.1b counting).
- Added phases still to come (entries of `upcoming` whose `phase_id > 5`) show as small "+ Combat", "+ Main", "+ Begin" markers after the current group. They are not new step buttons: per-step stops stay keyed by step name, and an added combat's declare attackers stops where the first did (ADR 0059 Decision 11).
- Turn skips: see owner question 1.
- Every label is an `L.<key>` in `client/src/lib/labels.ts` and listed in `docs/labels.md`.

### 10. Enumerator, heuristic bot and MCP seat

- **Enumerator** (`internal/legal`): no change in behaviour. A skip is spent at the seam, not offered. Add the agreement test ADR 0059 Decision 13 left: declare-attackers moves are offered in an added combat, and `dispatchAll` accepts them.
- **Heuristic**:
  - `nextTurnSeat` reads `TurnView.next_seat`, falling back to today's walk when it is absent. This fixes the extra-turn rotation bug above.
  - `game.Purpose` gains `ExtraTurns int` and `SkipsTurns int` (the caster's own turns). Declared on the shipped extra-turn cards and on every card this ADR adds. The heuristic prices them with one new weight, `Weights.Turn` (a whole turn), positive for `ExtraTurns`, negative for `SkipsTurns`, tuned with an arena measurement. An any-player row such as Lethal Vapors' declares no Purpose, so a bot never activates another player's copy (ADR 0106 decision 2).
- **Model prompt and MCP seat** both render through `aiseat/boardtext`:
  - The turn line adds "(combat 2)" / "(main 3)" when `phase_ordinal > 1`.
  - A seat line adds "skips their next turn" / "next 2 turns" from `turn_skips`.
  - The turn line adds "next turn: <seat>" from `next_seat`.
- **Catalog soak**: the pool is the whole dealable catalog, so new cards join it automatically. The `Seq` cap from ADR 0059 Decision 13 already catches a runaway. A table where every remaining player owes skips still advances, because each seam pass spends a charge.

### 11. Snapshot, restore, undo and the closure ratchet

All additive within `SnapshotSchemaVersion` 8, recorded in `game/testdata/snapshot_shape/v8.txt` with `-update-shape`. No fixture changes, and no version bump.

| New on disk | Path | An older binary reading a newer file |
|---|---|---|
| `skipTurns` mod | `scopedEffects[].mods[].kind` | refuses the file (`ErrUnknownEffectKey`, unknown mod kind) and keeps it |
| `skipExtraTurnStep` mod, `step`, `extraTurn` | `scopedEffects[].mods[]` | refuses the file, as above |
| deferred effects | `extraTurns[].effects` | drops them: damage can be prevented in that turn (weaker), until the next deploy |
| phase binding | `delayedTriggers[].onSeq`, `.onPhaseId` | refuses the file: both binding cards use body keys an older binary lacks |
| extra turn after a move | `ReplacementEvent.extraTurnsAfterMove` | not persisted. It lives only between a Replace callback and the move's finisher, the same window as ADR 0178's `lifeAfterMove` |

- **Why skips are records, not a counter.** A rollback that dropped a skip would hand the player a turn their card made them pay away. That is stronger than printed, which the project never ships (AGENTS.md §7). A new mod kind makes the older binary refuse the file instead.
- **Clone / undo.** `ScopedEffects` and `DelayedTriggers` already clone, so spending a charge copy-on-write (`setShieldChargeLocked`) lets an undo rewind it. `cloneExtraTurns` must deep-copy `Effects`. The drift test (`snapshot_drift_test.go`) needs no new top-level rows, because every new field sits inside an already-`carried` one. Its shape coverage must show the nested paths.
- **Closure ratchet.** No new route from `Game` to a func or an interface: the records and the instruction slice are plain data. `testdata/closure_fields.txt` and the `census:` / `transient` counts are unchanged (`TestClosureFieldsReachableFromGame`).

## Alternatives considered

- **A. A replaceable "a turn would begin" event (`RepEventTurnBegin`), with "skip your next turn" as a replacement gathered into it.** This is closest to CR 614.10's wording. Rejected: the seam's windows cannot pause, so CR 616's order among a player's own skip and Trouble in Pairs' cancel would be fixed by gather order anyway. A charge is not a pure cancel, so gather order would decide which effect is spent. Decision 1 fixes that order to the affected player's dominant choice with no window at all.
- **B. A plain counter, `Player.TurnsToSkip int`.** Rejected: no source for the log line, and an older binary would silently drop it, which is stronger than printed (Decision 11).
- **C. `ExtraTurn.SkipSteps []Step` for Savor the Moment.** This is simpler, but rejected for the same rollback reason: an older binary drops it and the caster untaps.
- **D. A duration kind that starts in a future turn (`DuringExtraTurn`).** Rejected: 27 call sites read `!durationExpiredLocked` as "in force" (`grep durationExpiredLocked game/*.go`). A dormant-then-active duration would need each one to learn a third state, and each sweep would have to keep a record it now drops. Deferring the registration to the turn's start has one reader.
- **E. World at War anchored by template index ("after phase id 4").** Rejected: CR 505.1b counts main phases that occurred, and an added main phase (Relentless Assault earlier in the turn) is the second main phase. Only the plan walk is right.
- **F. Moraug's untap as Full Throttle's "each combat this turn" trigger, filtered to added combats.** Rejected: two Moraug landfalls add two combats and each untaps only at its own. The ruling's precombat case also needs "this added combat, not the regular one".
- **G. Asking the player to order their own skip against Trouble in Pairs.** Rejected: it is a prompt with one sensible answer, in a window that cannot pause.

## Delivery plan

Each slice is one PR into `develop`, squash-merged, with its tests green under `scripts/go-docker.sh test ./...` and `lint`. Cards follow AGENTS.md §7: one file per card, `Completeness` truthful, only the new cards' oracle fixtures (`-update-oracle -oracle-ids=…`). A card blocked by anything outside its slice drops out with a registry note; it never ships stronger than printed. The slice that empties a registry row's `Waiting` list sets it `StatusImplemented`, adds a closed-seam fragment under `docs/engine-seams/closed/`, and regenerates both tables (`go test ./internal/roadmap/ -update` and `-update-closed`).

### Slice 1 — `feat(server): skip a player's next turns (CR 614.10)` · tier:2-standard

Scope:
- `game/scoped_effects.go`: `ModSkipTurns` in the mod-kind table (`readerRule`).
- `game/turn_skip.go` (new): `SkipNextTurnsForEffect`, `consumeTurnSkipLocked`, `TurnSkipsFor`.
- `game/rotation.go`, `game/extra_turns.go`: spend in the normal-rotation loop and in `popExtraTurnLocked` before `extraTurnSkippedLocked`; drop a departed seat's records; the loop guard.
- `game/events.go`: `EventTurnSkipAdded`, `EventTurnSkipped`.
- `protocol/log.go`: `turn_skip`, `turn_skipped`.
- `protocol/view.go`: `PlayerView.turn_skips`.
- `docs/protocol.md`; `client/src/lib/protocol.ts` (type and log kind union only).
- `effects/skip_turns.go` (new): `SkipYourNextTurns`, `TargetPlayerSkipsNextTurns`.
- Cards: `magosi_the_waterveil.go`, `lethal_vapors.go` (`AnyPlayer: true`, no Purpose).
- Registry `extra-turns`: drop both from `Waiting`; widen `Printed` with `\bskips? (your|their) next (\w+ )?turns?\b`; rewrite `Missing` and `EngineNotes`.
- `docs/adding-cards.md` § Extra turns; one-line pointer from ADR 0059 Decision 14 to this ADR; snapshot shape file.

Acceptance (tests):
- Four seats, seat 1 owes one skip: the order is 0, 2, 3, 0, 1. Seat 1's `TurnsBegun` does not move on the skip, `Seq` goes up only on turns that begin, and `Round` goes up on the wrap.
- Amount 2, or two records of 1: the next two turns are skipped (CR 614.10a). The oldest record is spent first.
- Magosi ruling: a skip and a queued extra turn of the same player; the extra turn is the one skipped, and normal rotation then continues.
- A skip created during the player's own turn does not end that turn.
- A player owing a skip, whose extra turn an opponent's Trouble in Pairs would also skip: the charge is spent, and no `EventExtraTurnSkipped` is emitted.
- A skipped turn: no `EventTurnBegan`, no untap. `recordPrevTurnSpellsLocked` and `LastTurnAttacks` still describe the last turn taken. A rebound upkeep trigger waits to the next real upkeep. An `UntilYourNextTurn` effect survives the skip. A skipped extra turn's Final Fortune trigger is swept, and the game is not lost.
- The departed-seat path is unchanged, and a departed player's records are dropped.
- Snapshot round trip, undo across the rotation restores the charge, and a file naming `skipTurns` is refused by a mod table without it.
- Log projection for both lines.
- `TestEveryModKindHasATestCase` has a case for `skipTurns`.
- Magosi: both abilities in one turn cancel out (ruling); the skip applies even if Magosi has left before resolution (ruling).
- Lethal Vapors: an opponent activates it, the enchantment is destroyed, and the opponent skips their next turn.

Depends on: nothing.

### Slice 2 — `feat(server): skip every extra turn, and ugin's nexus` · tier:2-standard

Scope:
- `effects/skip_extra_turns.go`: `SkipExtraTurns()` and the shared constructor.
- `game/move_extra_turn.go` (new): `TakeExtraTurnAfterReplacedMoveForEffect` and `extraTurnsAfterMove` on `ReplacementEvent`, run in `runMoveTailLocked` before the life instructions; carried in `cloneReplacementResume` and wherever `zone_route.go` carries `lifeAfterMove`.
- Cards: `ugins_nexus.go`, and `gerrards_hourglass_pendant.go` if its third ability is expressible (Flash, `putIntoGraveyardFromBattlefieldThisTurn` from `effects/exile_if_dies_g4.go`, a mass return tapped). Otherwise the Pendant stays out with a registry note.
- Registry: drop Ugin's Nexus from `Waiting`. `docs/adding-cards.md`.

Acceptance (tests):
- Ugin's Nexus skips its own controller's extra turns as well as opponents'.
- An extra turn queued while the Nexus is on the battlefield is taken if the Nexus has left before it would begin (ruling).
- Destroying the Nexus exiles it and queues one extra turn for its controller, which is taken. The `extra_turn` log line follows the exile line.
- Two Nexuses destroyed together: one is exiled with an extra turn, the other stays, and that turn is skipped only if the remaining Nexus is still there (ruling).
- With Rest in Peace, the CR 616 order prompt is asked of the Nexus' controller, and choosing Rest in Peace first gives no extra turn.
- With a player who owes a skip (slice 1), the debt is spent before the Nexus' window.

Depends on: slice 1, for the last test only.

### Slice 3 — `feat(server): a step skipped and an effect during one extra turn` · tier:2-standard

Scope:
- `game/scoped_effects.go`: `ModSkipExtraTurnStep`, plus `Mod.Step` and `Mod.ExtraTurn`.
- `game/turn_skip.go`: `SkipStepOfExtraTurnForEffect`, spent in `runStepEntryHooksLocked`.
- `game/delayed.go`: `sweepUnreachableBoundTriggersLocked` also sweeps those records.
- `game/extra_turns.go`: `ExtraTurn.Effects`, `DuringExtraTurnRuleEffectForEffect`, a deep `cloneExtraTurns`; registration in `startTurnLocked`.
- `game/scoped_effects.go`: split `buildScopedEffectLocked` out of `appendScopedEffectLocked`.
- `effects/extra_turns.go`: `TakeExtraTurn` gains `SkipSteps []game.Step` and `DuringThatTurn []game.Mod`.
- Cards: `savor_the_moment.go`, `alchemists_gambit.go` (cleave drops the bracketed loss: the cleave cast schedules no bound trigger).
- Registry: drop both. `extra-turns` `Waiting` is then empty (with slices 1 and 2 merged): close the row.
- `docs/adding-cards.md`.

Acceptance (tests):
- Savor: in the extra turn nothing untaps, phases in, or flips day/night. Summoning sickness from the previous turn is gone. A "doesn't untap during its controller's next untap step" marker survives to the next real untap step. The next turn's day/night check reads Savor's turn.
- Savor under Stasis behaves the same.
- Savor, then Sphinx of the Second Sun in that turn: the added beginning phase's untap step untaps (owner question 2, recommended answer).
- Savor's extra turn skipped by Trouble in Pairs leaves no record behind.
- Alchemist's Gambit: damage can be prevented for the rest of the current turn and cannot during the extra turn. A Fog cast in that turn prevents nothing, and the effect ends at that turn's cleanup.
- Alchemist's Gambit cast for cleave: no loss. The extra turn skipped: no effect and no loss (ruling).
- Snapshot and undo round trip with a queued `effects` entry and a `skipExtraTurnStep` record.

Depends on: nothing (the row closes only after slices 1 and 2).

### Slice 4 — `feat(server): after the second main phase, and a trigger bound to that combat` · tier:2-standard

Scope:
- `game/turn_plan.go`: `AnchorNthMainPhase`, `PhaseAnchor.N`.
- `game/delayed.go`: `OnSeq`, `OnPhaseID`, the fire test, the refusal and the sweep. Snapshot `onSeq` / `onPhaseId`; clone.
- `effects`: `ExtraCombatAndMainAfterNthMain`, `ScheduleDelayedTrigger.OnPhase`, bodies `extra-combat/untap-attackers-at-that-combat` and `extra-combat/untap-all-you-control` in `delayed_bodies.go`.
- Cards: `world_at_war.go` (rebound is the existing keyword), `moraug_fury_of_akoum.go`. Moraug's anthem is a layer-7c static whose `Apply` adds `g.TimesAttackedThisTurn(target.ID)` to power, with `DependsOnAttackingStatus: true`, so it re-reads at each declaration.
- Registry `extra-combats`: drop both; `Waiting` is then empty: close the row.
- `docs/adding-cards.md` (the "Not built yet" paragraph goes).

Acceptance (tests):
- World at War in the precombat main phase: beginning, main, combat, main, [combat, main], ending.
- Two copies in the precombat main phase: the second copy's pair comes first (ruling).
- A copy cast in the third main phase adds nothing (ruling).
- A rebound cast in the upkeep anchors to that turn's second main phase.
- The untap fires at the added combat only, and untaps creatures that attacked in any earlier combat.
- Moraug's landfall in the precombat main phase: the added combat comes before the regular one, and only it untaps (ruling).
- Two landfalls in the postcombat main phase: two consecutive combats, each untapping, with no main phase between them (ruling).
- The anthem reads 0, 1, then 2 across two combats. A creature put onto the battlefield attacking is not counted (ruling).
- Commander damage from a commander attacking in both combats adds up toward 21 (CR 903.10a).
- A bound trigger whose combat never begins (the turn is ended, CR 724.1) is swept at the next turn.
- `ScheduleDelayedTriggerForEffect` refuses an unreachable phase id.
- `until end of combat` from the first combat has ended in the second; `until end of turn` lasts through both (Decision 6).

Depends on: nothing.

### Slice 5 — `feat(server): more cards that skip your next turn` · tier:1-mechanical

Scope: one file each for Meditate, Chronatog, Chronatog Totem, Eater of Days, Wormfang Manta, Waterspout Elemental, Chronosavant, Unstable Hulk and Ral Zarek, Guest Lecturer, using `SkipYourNextTurns` / `TargetPlayerSkipsNextTurns` and existing shapes (kicker, morph, coin flips, once-each-turn). Each card is `Full` or drops out with a registry note. `Purpose.SkipsTurns` is declared once slice 6 has added it, or added to these files in slice 6.

Acceptance: one test per card that the right player skips the right number of turns, plus each card's other clause.

Depends on: slice 1 (owner question 3 decides whether this slice exists).

### Slice 6 — `feat(bot): see the next turn and price extra and skipped turns` · tier:2-standard

Scope:
- `game`: `NextTurnSeatForEffect`; `protocol`: `TurnView.next_seat`.
- `aiseat/heuristic`: `nextTurnSeat` reads `next_seat`.
- `game/purpose.go`: `ExtraTurns`, `SkipsTurns`, plus `effects/purpose.go` slot checks; declared on the shipped extra-turn cards and the new skip cards.
- Heuristic: `Weights.Turn`.
- `aiseat/boardtext`: "(combat 2)", "skips their next turn", "next turn: <seat>".
- `internal/legal`: the agreement test for a second declare attackers step.
- `docs/bot.md`, `docs/protocol.md`.

Acceptance (tests):
- `next_seat` is correct across: a seat 2 extra turn after seat 0 (four seats, then seat 1), a skip charge, a departed seat, and stacked extra turns.
- The heuristic's `beforeMyUntap` is true in seat 0's end step when seat 0's extra turn is next.
- Boardtext golden lines for an added combat, a skip, and the next turn.
- `legal` offers declare-attackers moves in an added combat and `dispatchAll` accepts them.
- Arena measurement in the PR body for `Weights.Turn`, with a bot holding Time Warp casting it and one holding Meditate (if slice 5 shipped) not casting it at an even board.

Depends on: slice 1.

### Slice 7 — `feat(client): label repeated phases and show turn skips` · tier:2-standard

Scope:
- `client/src/lib/components/board/PhaseDisplay.svelte`: the "Combat 2" / "Main 3" caption, "+ Combat" markers from `upcoming`, the "Next:" line from `next_seat`.
- The player panel's turn-skip chip (owner question 1).
- `client/src/lib/labels.ts`, `docs/labels.md`; vitest for the derivations (pure functions in `client/src/lib/turn.ts`).

Acceptance (tests):
- Vitest: the caption for `phase_ordinal` 1, 2 and 3.
- Vitest: markers for a plan with two added combats and a main.
- Vitest: no markers without added phases.
- Vitest: the "Next:" text for an extra turn, a skip, and the plain case (no line).
- Vitest: the chip text for 1 and 2 skips.
- The render test that every step sits in one group still passes.

Depends on: slices 1 and 6.

Order: 1 → 2 → 3, and 4, in any order after; 5 and 6 after 1; 7 last.

## Open owner questions

1. **How does the table show a pending turn skip?**
   - (a) A "Skips next turn" / "Skips next 2 turns" chip on that player's panel, and the turn bar's "Next:" line names the real next player whenever it is not the plain next seat (a skip, or an extra turn as today). **Recommended**: the information that changes who acts next sits where players already look for it, and a game without skips looks exactly as it does now.
   - (b) The log lines only.
   - (c) (a) plus a reveal-strip cue when a skip is created. This goes against the owner's "no reveal-strip cues for these events" policy (ADR 0059 owner decision 2).
2. **Savor the Moment plus an added beginning phase in that extra turn: which untap steps are skipped?**
   - (a) Only the first untap step of that turn (the turn's own). **Recommended**: "the untap step of that turn" names one step, CR 614.10a has a skip satisfied by the first occurrence, and no ruling says more. The case needs Sphinx of the Second Sun in the same turn.
   - (b) Every untap step of that turn.
3. **Which cards ship?**
   - (a) The seven registry cards in slices 1–4, plus Gerrard's Hourglass Pendant in slice 2, plus slice 5's nine off-registry skip cards. **Recommended**: slice 5 is mechanical once slice 1 exists, and without it the new `Printed` pattern lists nine cards as waiting on a closed seam.
   - (b) The seven registry cards only, and slice 5 filed as a separate batch issue.
4. **The order of a player's own skip against an extra-turn skip.** Is it acceptable that the engine spends a player's "skip your next turn" debt before Trouble in Pairs, Ugin's Nexus or Gerrard's Hourglass Pendant can skip the same extra turn, without asking (Decision 1)?
   - (a) Yes. **Recommended**: it is the choice the affected player would always make under CR 616.1, and the window cannot pause to ask.
   - (b) No: build a pausable window first, so the player is asked. That would be a new resume point in the rotation seam, which ADR 0059 avoided on purpose.

## Consequences

- The `extra-turns` and `extra-combats` rows close. #753 stays open as the tracker until slice 7 lands.
- Two rules mod kinds, one `ExtraTurn` field, two `DelayedTrigger` fields, one anchor kind, two events, two log lines and two view fields. Everything on disk is additive in v8, and every addition that could make a card stronger after a rollback makes an older binary refuse the file instead.
- **The thing most likely to be got wrong later.** Spending a skip anywhere but the rotation seam: in `onTurnBeganLocked`, say, or in the untap step. By then the turn has begun, and CR 614.10 says it can no longer be skipped. Its `Seq`, resets and triggers would already have happened. A skip is spent while the next turn is being chosen, and never after.
