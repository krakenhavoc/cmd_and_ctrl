# ADR 0097 — Modes "that hasn't been chosen"

**Status:** Accepted · 2026-09-30 · Post-S30 — Rolling deck-driven catalog growth
**Issue:** [#1749](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1749). It relates to the deck re-checks on #1107, #1112 and #1117.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-09-30. I ran `git fetch --all --prune`, then read every `docs/decisions/` file name on every remote branch (35 heads). The highest number anywhere is **0096** (`0096-the-monarch-from-a-card-effect.md`).
**Builds on:** [ADR 0065](0065-modal-and-multi-target-clauses.md), which covers modes and the trigger's `mode_pick`, and lists "that hasn't been chosen" under Out of scope. It also builds on [ADR 0041](0041-game-persistence.md) for the snapshot shape rule, and on #936, which keys the per-object turn tally.
**Owner decisions:** 2026-09-30. There are four: scope, timing, UI and delivery. They are recorded under Decisions below.

---

## Context

Thirty-two cards print a modal clause restricted to modes "that hasn't been chosen". They fall into two families.

- **"…this turn"**. Examples: Monument to Endurance, Gala Greeters, Teval's Judgment, Breeches, Eager Pillager, Parapet Thrasher, Galadriel, Light of Valinor, The Vision, Lita, Kargan Intimidator, Genku and The Fantastic Four. Each mode is available once per turn.
- **With no duration**. Examples: Silent Hallcreeper, Demonic Pact, Captive Audience, Gandalf the Grey, Henrika Domnathi, Rejoin the Fight, Survivor's Med Kit and Kimoyo Beads. Each mode is available once, ever, for that object.

A few more cards say "hasn't been chosen" about something that is not a mode. Garth One-Eye names a card, Interrogation Robot names a question word, and Fortunate Few chooses cards. They are out of scope.

The engine has no memory of past mode choices, so the catalogued cards work around it:

- Gala Greeters and Silent Hallcreeper offer every mode every time. That is stronger than printed, and both carry a caveat saying so.
- Teval's Judgment counts its own resolutions and takes the modes in printed order. The player does not choose, and two triggers on the stack at once can collide. Counting resolutions rather than choices is the wrong event.
- Breeches, Eager Pillager can't be built. Monument to Endurance is on hold.

### What the rules and rulings say

There is no Comprehensive Rules entry for this restriction; the card text governs. The surrounding rules are these:

- **CR 700.2a and 700.2b.** A modal activated ability's modes are chosen as part of activating it. A modal triggered ability's modes are chosen as part of putting it on the stack. A mode that would be illegal can't be chosen, and **if no mode is chosen, the triggered ability is removed from the stack**. CR 603.3c says the same for triggers.
- **CR 700.2g.** A copy of a modal ability copies the modes chosen for it, and the copy's controller can't choose a different mode.
- **CR 400.7.** An object that changes zones is a new object with no memory of its previous existence.

The official rulings fill in the rest:

- **Per object.** Silent Hallcreeper (2024-09-20): the phrase "refers only to that specific Silent Hallcreeper. If [it] leaves the battlefield and then returns … it will be a new object with no memory of the modes that were chosen." Demonic Pact (2015-06-22): a second Demonic Pact may choose any mode. Breeches (2023-11-10): with two of them, "track which modes have been chosen each turn for each one's ability separately."
- **Counted when chosen.** Demonic Pact: "If the ability doesn't resolve … the mode chosen for that instance of the ability still counts as being chosen."
- **Not per controller.** Demonic Pact: "It doesn't matter who has chosen any particular mode." An opponent who gains control can choose only the modes that remain.
- **Simultaneous triggers.** Gala Greeters (2022-04-29): "you must still choose different modes for each instance … If more than three creatures enter … simultaneously, that choice is made only for the first three."
- **Exhaustion.** Breeches: "If you can't legally choose a mode because all three have been chosen that turn, that instance of the ability is removed from the stack with no effect." Demonic Pact: "if the fourth mode is the only one remaining, you must choose it."

---

## Decisions

### 1. It is declared on the `ModeSpec` (owner decision: both families)

`game.ModeSpec` gains one field:

```go
// NotChosen restricts the options to those this OBJECT's ability has
// not chosen before. The zero value is no restriction.
NotChosen ModeMemory // ModeMemoryNone | ModeMemoryThisTurn | ModeMemoryEver
```

The catalog writes it with two constructors that read like the card: `ChooseOneNotChosenThisTurn(opts...)` and `ChooseOneNotChosen(opts...)`. `effects.Register` refuses two combinations:

- `NotChosen` together with `Repeatable`. CR 700.2d's "you may choose the same mode more than once" contradicts the restriction.
- `NotChosen` on a **spell's** `Spec.Modes`. No printed spell says it, and a spell has no permanent object to remember with.

Triggered and activated abilities share the one `ModeSpec`, so Kargan Intimidator's activated "{1}: choose one that hasn't been chosen this turn" needs nothing more.

### 2. The memory belongs to the object and its ability

The memory is keyed by the ability's identity on the object: `ObjectTallyKey(source, Card.ObjectEpoch, label)`, where the label is a triggered row's `Key` or an activated row's `Label`. This is the same key the #936 "once each turn" gates use. It records the set of option indexes chosen.

- **This turn:** `TurnTally.ModesChosen map[string][]int`. It is flushed with the rest of the tally when the turn begins.
- **Ever:** `Card.ModesChosen map[string][]int`, keyed by label. `MoveCard` clears it with the rest of CR 400.7's forgetting, so an object that leaves and returns remembers nothing. It is not a copiable value (CR 707.2), so a copy starts empty.

Both are keyed by the object, never by its controller, so a change of control keeps the memory (Demonic Pact ruling). Both new fields are additive snapshot fields, so a restore point and an undo carry them. The shape record is updated in the same change and there is no schema bump. `Game.Clone` deep-copies both.

### 3. One filter decides what is offered

`choosableModeOptionsLocked` already answers "which options may be chosen now". Every mode path calls it: the harvest check, the trigger's `mode_pick`, the bot enumerator and the view. It gains the ability's identity and drops the options the memory has recorded, **before** the existing "no legal target" filter.

- A caller with no identity (a spell) excludes nothing.
- The harvest-time `EnoughChoosableModes` check therefore sees exhausted modes. A trigger with too few options left is never put on the stack (CR 700.2b, and the Breeches ruling).
- `validateModes` at activation (`activated.go`) is checked against the same filtered set, so an activation naming a used mode is refused with nothing paid.
- The bot enumerator reads the same filter, so a bot is never offered a used mode (#544).

When only one mode remains, it is the only answer, and a mandatory trigger must take it (the Demonic Pact ruling). The existing `Min` logic already produces that.

### 4. A mode is recorded when it is chosen (owner decision)

The record is written at the moment of choice. It is not written at resolution:

- **Trigger:** when the `mode_pick` answer is accepted, or when the engine takes the only possible answer without asking.
- **Activated ability:** when the activation succeeds and the item is on the stack. A failed activation records nothing.

So a countered or fizzled trigger still used its mode. A copy (CR 700.2g, `CopyAbilityForEffect`, `CopySpell`) copies the modes and records nothing. Each extra instance a trigger doubler adds chooses for itself and records its own choice.

**Simultaneous triggers.** Several instances of one ability can each queue a `mode_pick` before the first is answered. Recording a choice therefore re-narrows every still-pending `mode_pick` of the same ability of the same object. A prompt left with fewer than `Min` options is withdrawn through `dropChoiceLocked` and its trigger removed. This is Gala Greeters' "different modes for each instance … only for the first three". The answer gate also re-checks the live filter, so an answer can never name a mode another instance took first.

### 5. The table sees used modes greyed out (owner decision)

The `mode_pick` prompt and `ModeSpecView` carry every option, with each used one marked `used: true`. The picker renders a used option disabled, labelled "already chosen this turn" or "already chosen". Nothing new appears on the board. The view and the enumerator read the same filter, so what is greyed and what is refused cannot disagree.

### 6. Delivery (owner decision)

The work lands in two PRs.

- **PR 1** contains this ADR (flipped to Accepted), the engine change, and fixes to the catalogued cards that fake it. Gala Greeters and Teval's Judgment become the real shape. Silent Hallcreeper loses its "not enforced" caveat, and its "becomes a copy" caveat is re-checked against #1593's duration copy. Monument to Endurance is added. The PR also adds three roadmap registry rows found in the same re-check: this mechanic (closed by the PR, with a `docs/engine-seams/closed/` fragment), discover, and keyword counters read by the engine.
- **PR 2** is a build agent's card batch. It adds the newly buildable cards whose other clauses are expressible: Breeches, Parapet Thrasher, Galadriel, Demonic Pact, Captive Audience, Gandalf the Grey, Kargan Intimidator and the rest of the list above. It records any that are still blocked on the roadmap.

---

## Consequences

- Cards stop being stronger than printed. Gala Greeters and Silent Hallcreeper used to allow repeats. Around twenty cards become buildable.
- A trigger can now be dropped by its own history. That is what the rulings require, but it is new. The drop goes through the paths the "no legal target" case already uses (the harvest check and `dropChoiceLocked`), so nothing wedges.
- The memory is two new maps. They grow only with modal abilities that declare `NotChosen`, and both are flushed by existing events (turn start, zone change).
- Nothing changes for any modal card without the field. A caller with no ability identity excludes nothing.

## Out of scope

- Non-mode "hasn't been chosen" choices: Garth One-Eye's card names, Interrogation Robot's question words, and Fortunate Few's cards. Those are choice prompts of another kind.
- Showing the memory on the permanent itself (a chip). The owner chose the picker only. The data is on the wire if that changes.

## Amendment (2026-09-30, #1749): what PR 1 built, and where it differs

PR 1 built Decisions 1 to 5 as written. These are the places where the code says more than the ADR, or says it differently.

- **The identity is a value, `game.ModeAbility`.** It holds `{Source, Epoch, Label}` and is built with `ModeAbilityOf(card, label)`. `choosableModeOptionsLocked` and `ChoosableModeOptionsForEffect` take it as a third argument, and the zero value (a spell) excludes nothing. The enumerator's `legalModeSets` and the view's `viewOfModeSpec` take it too, so there is still one filter.
- **The `mode_pick` prompt carries the used modes beside the offer, not inside it.** Decision 5 says the prompt carries every option and marks the used ones. The prompt instead keeps `mode_options` / `mode_indexes` as the offer, meaning what may be answered, and adds `mode_used_options` / `mode_used_indexes` and `mode_not_chosen` (`"this_turn"` / `"ever"`). This keeps `ModePickSelections`, `validateModePick` and every client that reads `mode_indexes` unchanged. The picker merges the two lists by index and shows the used ones disabled. `ModeSpecView` does what Decision 5 says: each option carries `used: true`, and the spec carries `not_chosen`.
- **There is no "only possible answer taken without asking" path.** A trigger with one mode left still asks, so the one recording site for a trigger is `ResolveModePick`. It records after the answered prompt has left the queue, so the re-narrowing never touches that prompt.
- **An exile return that mints a new instance also forgets.** `Card.ModesChosen` is cleared by `MoveCard`, as written, and also by the new-object reset of an exile return (`resetAsNewObjectLocked`). A phased-out permanent is still the same object (CR 702.26d), so the "ever" lookup also finds it in `PhasedOut`.
- **`effects.Register` refuses a third combination.** It also refuses `NotChosen` on an ability with no label, because the label is half of the memory's key.
- **Silent Hallcreeper's copy bullet is built.** It uses `BecomeCopy` with `CopyIndefinite`. "Another target creature you control" excludes the Hallcreeper by name (`b03NotNamed`), as Departed Deckhand does, because a mode's static target clause is never handed its source. The card is `CompletenessFull`.
- **Registry rows.** The `discover` row already existed, from #1751. PR 1 added `modes-not-chosen` (implemented) and `keyword-counters` (missing, #1753).
- **Kargan Intimidator waits for PR 2.** A registered fixture with Kargan's shape covers the activated path: the refusal with nothing paid, and the enumerator. A protocol test covers the view's `used` mark.

