# ADR 0170 — Choosing other permanents as a permanent enters

**Status:** Accepted · 2026-10-13 · S58 — Deck requests (October batch)
**Issue:** [#1893](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1893) (Phylactery Lich; Sin, Unending Cataclysm from deck request #3246).
**Numbering:** the AGENTS.md §4 sweep on 2026-10-13 (every remote branch's history under `docs/decisions/`, plus the files of every open PR: #3412, #3409, #3356, #3329) found **0169** the highest, so this is **0170**.
**Builds on:** [ADR 0098](0098-discard-as-a-permanent-would-enter.md) (`EntryCardChoice` with an `Action`, the paused entry and its resume), [ADR 0013](0013-replacement-effects.md) §5z (the reveal-land prompt it grew from), #2525 (proliferate's `ChooseSuggested`) and [ADR 0107](0107-state-triggers-rebound-disturb-and-damage-prevention.md) §1 (state triggers).

---

## Context

Two cards make a choice of OTHER permanents as they enter, and act on them:

- **Phylactery Lich:** "As this creature enters, put a phylactery counter on an artifact you control. When you control no permanents with phylactery counters on them, sacrifice this creature."
- **Sin, Unending Cataclysm:** "As Sin enters, remove all counters from any number of artifacts, creatures, and enchantments. Sin enters with X +1/+1 counters on it, where X is twice the number of counters removed this way."

**CR 614.1c** makes "as [this] enters" a replacement effect. **CR 614.12** applies it before the permanent is on the battlefield, so the choice and what it does happen before anything can see the permanent, and **CR 614.13a** forbids choosing the permanent itself or anything entering with it.

`Spec.AsEnters` chooses a colour, a type, a name, a player or an option, but it runs after the permanent has arrived (the engine-wide as-enters simplification), and it chooses no permanent. ADR 0098's `EntryCardChoice` is the pre-push form, with a card choice inside the paused entry, but its three actions (reveal, discard, sacrifice) each DO something to the named cards. Its `Then` callback never saw the entering event, so it could not give the permanent counters.

## Decisions

### 1. A fourth action, `EntryCardChoose`

`EntryCardChoice.Action` gains `EntryCardChoose`. It names permanents on the battlefield and the engine does nothing to them; the declaration's `Then` is the whole effect. It is asked through a new prompt kind, **`entry_choose_permanents`**, which joins the other three everywhere they are listed: `isEntryCardChoiceKind` (so the stale-candidate prune and chained resolver leave it alone), `isCardSetPickKind` (the bounds, candidate and live-zone checks), the table-blocking gate, the leave-game table (its frame rides `replacementResume`, as its siblings' do), the action router and the enumerator.

The new kind, rather than reusing `entry_sacrifice`, is for the same reason ADR 0098 gave each action its own kind: the sign differs for a bot (a named permanent is not spent), and the client's sentence differs.

There is no "if you don't" branch. The declaration's `Replace` is nil. With fewer candidates than the floor, or an entry that cannot be resumed (the sandbox `move_card` verb, ADR 0098 owner decision 6), nothing is asked and the permanent simply enters. Phylactery Lich with no artifact enters with no phylactery anywhere, and its state trigger sacrifices it.

### 2. Who may be chosen

By default the candidates are the permanents the chooser controls ("an artifact you control"). `AnyController` widens them to every permanent on the battlefield ("any number of artifacts, creatures, and enchantments"). `Matches` filters, as before, over the permanent's effective characteristics. `AnyNumber` gives a floor of `Min` (normally 0) and a ceiling of every candidate, as devour's does.

The entering permanent and every member of its simultaneous entry are excluded by the existing CR 614.13a rule in `entryChoiceCandidatesLocked`.

### 3. `Then` is handed the entering event

`EntryCardChoice.Then` becomes `func(g *Game, ev *ReplacementEvent, picked []uuid.UUID) error`, for every action. With the event, `Then` can change how the permanent enters: `ev.AddCounterAtETB` is "enters with X counters", so Doubling Season and Hardened Scales apply to them (CR 614.1c). This also unblocks Indominus Rex, Alpha's follow-on on the ADR 0098 row. No card used `Then` before, so nothing else changes.

### 4. `Then` settles inline

`Then` runs while the entry is paused, inside the prompt's resolver. What it does to OTHER objects must settle there: counter placements and removals use the `MustSettleNow` forms (`AddCounterByMustSettleNowForEffect`), which apply a CR 616 order in the gathered order rather than asking. A pausing primitive would let the entry finish before its own nested prompt was answered. This is the posture every other `mustSettleNow` path takes: weaker than printed only when two counter replacements would otherwise need ordering, never stronger.

### 5. A suggested answer

`EntryCardChoice.SuggestKey` (optional) names an `EntryChoiceSuggestion` filed at init with `game.RegisterEntryChoiceSuggestion`, which returns the declaration's recommended pick as the prompt is queued; the pick is filtered to the candidates. It is a key and not a func for the reason state triggers are (ADR 0107 §1): the declaration gains no closure for the ADR 0041 ratchet. It rides the prompt as `ChooseSuggested`, proliferate's existing field (#2525), which is already snapshotted, cloned and shipped on the view. The client pre-selects it, the enumerator always offers it first (`withSuggestedFirst`, so an "any number" walk that hits the expansion cap still reaches it), and the heuristic bot scores overlap with it (`proliferateValue`). It is a default, never a rule.

- Phylactery Lich suggests the artifact least exposed to removal: a noncreature before an artifact creature, a nontoken before a token.
- Sin suggests every candidate an opponent controls, plus its controller's own permanents whose counters are all harmful (`game.HarmfulCardCounter`, proliferate's table).

### 6. Restore points

The open prompt holds a resume frame, so like its ADR 0098 siblings it is not a restore point. The undo clone carries it. No snapshot field is added.

## Cards

- **Phylactery Lich** (Full). The counter is placed by the artifact's controller. The sacrifice is a CR 603.8 state trigger (`WhenState`) over permanents its controller controls with a phylactery counter.
- **Sin, Unending Cataclysm** (Full). Only permanents with a counter are offered: naming one with none is legal and does nothing. Every kind comes off, and X is read back from the board, so a counter that can't be removed (#1824) is not counted. The dies trigger puts Sin's last-known counters (CR 603.10) on the target, then tucks Sin from its owner's graveyard into the library and shuffles. If Sin has left the graveyard, the ability does not find it (CR 400.7).

## Consequences

- One new prompt kind on the wire (`entry_choose_permanents`), documented in `docs/protocol.md` and rendered by the client's card-set grid and board pick.
- A future "as this enters, choose a permanent" card is a declaration: `AsEntersChoosePermanents` plus a `Then`.
