# ADR 0216 — Planechase as a table option (CR 901)

**Status:** Accepted · 2026-11-03 · S79 — Card coverage: every Standard set, newest first (an owner request outside S79's card scope, see [Scope](#scope-an-owner-decision-that-widens-the-charter)).
**Issue:** epic [#4870](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4870). This ADR designs the epic; it does not close it.
**Numbering:** checked on 2026-11-03. `develop` ends at 0215, and `git ls-tree` of `docs/decisions/` on every one of the 23 remote heads (after `git fetch --all --prune`, the loop in [README.md](README.md#picking-a-number), run again just before the push) shows no file above 0215. This ADR takes **0216**.
**Amends:** [ADR 0054](0054-dice-rolls-and-coin-flips.md) Decision 10 (the planar die was out of scope), [ADR 0060](0060-leaving-the-game.md) (CR 800.4p was unimplemented and unfiled), [ADR 0198](0198-outside-the-game-is-the-card-database-without-a-sideboard.md) §2 (planes and phenomena are excluded from the pool). Each carries a dated note pointing here; none of their decisions is rewritten.
**Builds on:** [ADR 0075](0075-table-settings-and-host-controls.md) (`TableSettings`, one struct owned by the engine), [ADR 0207](0207-attractions.md) (a second deck in the command zone, outside the card index), [ADR 0203](0203-the-initiative-and-dungeons.md) (an object in the command zone with triggered abilities, and an inherent trigger with a keyed body), [ADR 0064](0064-emblems.md) and [ADR 0140](0140-eminence-statics-from-the-command-zone.md) (abilities that function from the command zone), [ADR 0062](0062-abilities-and-special-actions-from-the-hand.md) (one door for every CR 116 special action), [ADR 0054](0054-dice-rolls-and-coin-flips.md) (keyed random streams, undo rewinds randomness), [ADR 0063](0063-durations-and-control.md) (durations), [ADR 0041](0041-game-persistence.md) (additive snapshots, restore points, stack items name their catalog row, the closure ratchet), [ADR 0083](0083-token-abilities.md) (catalog keys as on-disk identities).

**Decided by the owner (Ian, 2026-11-03):**

1. **Planechase is an opt-in table option**, chosen when a table is created. A table without it is unchanged.
2. **One communal planar deck** (CR 901.15), built automatically by C2 from every plane card in the card pool and shuffled per game. No deckbuilding. Per-player planar decks (CR 901.3) are a later option, out of scope here, and the design must not block them.
3. **The first slice is planes only:** the planar die, planeswalking, chaos and planes. Phenomena (CR 312) come right after, in a second slice. Design for them now; ship them later.

---

## A note on the rule text

The project pins the Comprehensive Rules effective September 25, 2026 (`MagicCompRules 20260925.txt`, AGENTS.md §6). The session that wrote this ADR could not download it: `media.wizards.com` and every rules mirror were refused by the session's network policy, and no file in the tree quotes any rule in 901, 311 or 312. So:

- **The rule numbers here are the ones the owner's brief on #4870 names** (901.1–901.15, 901.5, 901.6, 901.7/901.7a, 901.8, 901.9a–d, 901.10/901.10a–b, 901.11a–c, 901.13, 901.15a–c, 311, 311.7, 312, 706.7, 116, 800.4p). None is cited from memory, and no rule text is quoted.
- **Where a decision depends on the exact wording**, this ADR says what it depends on and leaves the reading to the slice that implements it. That slice quotes the rule verbatim in the file's header comment, the way `dungeon.go` and `initiative.go` quote CR 309 and CR 726, and corrects this ADR in a dated amendment if the text disagrees.
- **Before this ADR is merged**, someone with the TXT checks each number above against it. *Done 2026-11-04 by the owner: every cited number exists in the pinned text, and the paraphrased rules match it ([Owner answers](#owner-answers-ian-2026-11-04)). The slices still quote each rule verbatim before writing code.*

---

## Scope: an owner decision that widens the charter

[PLAN.md](../../PLAN.md) §3 scopes C2 to four-player Commander and lists "other formats" as out of scope. Planechase is not another format: it is a casual variant played *on top of* a Commander game (CR 901.1), and every table without the option stays a plain Commander game. It is still a widening of what C2 does, and it is the owner's call, made on 2026-11-03 (above). Nothing here is sold, and plane cards come from Scryfall like every other card, within the Fan Content Policy (PLAN.md §3).

Out of scope here, and staying out until someone asks: Two-Headed Giant Planechase (CR 901.12), Grand Melee's limited range of influence (CR 901.14, and so the range-of-influence part of 901.13, which C2's tables never limit), per-player planar decks (CR 901.3, owner decision 2), planar deckbuilding, and any plane card outside the card database.

---

## Context

### What exists today (develop at `956d36fad`)

- **Table settings.** `game.TableSettings` (`server/internal/game/settings.go`) is one struct on `Game.Settings`, cloned, snapshotted and public on the wire (`settings`). `SettingsPatch` changes it through `Game.UpdateSettings`, from `PATCH /games/{id}/settings` in the lobby or `set_table_settings` mid-game. `FreeMulligans` is the precedent for a setting that is fixed once the game starts. `POST /games` (`lobby.createGameRequest`) takes the table's other creation-time options (`voice_room`, `lfg`, `results`).
- **A deck in the command zone.** ADR 0207 put `Player.AttractionDeck` and `Player.Junkyard` in the command zone as `*Zone` of kind `ZoneCommand`, outside the card index and unreachable by any `ZoneRef`, so nothing but the mechanic itself takes a card out. Clone, the snapshot (`restoreCommandPile`) and `leave_game` handle them; the closure census classifies each as `census:IntrinsicAbilityCards`.
- **Command-zone sources.** Emblems' statics join the layer pass (`emblemContinuousEffectsLocked`), their triggers join the harvester (`harvestFromEmblemsLocked`), and the rule gates walk them (`forEachEmblemLocked`). A dungeon is a command-zone object whose room abilities the harvester walks (`harvestFromDungeonsLocked`), and its completion is an SBA that waits for the room ability to leave the stack (`dungeonCompletionSBALocked`, `roomAbilityOutstandingLocked`). ADR 0140 lets a cost modifier work from the command zone.
- **Special actions.** `PerformSpecialAction` (`special_action.go`) is the one door: a `SpecialActionKind`, a timing table (`SpecialActionTimingOKLocked`), a cost paid through `payAbilityManaCostLocked` with the zero spend context, `EventSpecialAction`, and `noteActionTakenLocked`. `ignore_effect` is the precedent for a kind that is not a card in the actor's zone: it branches before the card lookup.
- **Dice.** Every card's roll goes through `RollDiceThenForEffect(DiceRoll)` (`dice_roll.go`): the `RepEventRollDice` replacement window (Squid Fire Knight's sticker), the order and ignore prompts as keyed option picks, then one `EventRollDie` per kept die in one `BatchSeq`, then a registered tail. Draws come from keyed streams (`randForLocked(rngStream{...})`, ADR 0054 Decisions 1–2), and undo rewinds them (Decision 4).
- **Inherent triggers.** The initiative's three inherent abilities queue a `StackItem` whose behaviour is a registered body (`DelayedBody("initiative/venture", …)`), so a table with one waiting is a restore point.
- **Durations.** `DurationKind` (`duration.go`) is persisted as an int, with `durationKindEnd` kept last; a kind a newer binary adds is refused by an older one (ADR 0041 P4).
- **The card pool.** `deck.Collection` (`server/internal/deck/collection.go`) is the card database as a collection, installed as `game.OutsideCollectionSource`. Its `notInACollection` list keeps planes, phenomena, schemes, conspiracies, vanguards, dungeons and emblems out of it (ADR 0198 §2), so a wish never hands a player a plane.
- **Snapshots.** `SnapshotSchemaVersion` is 9. Changes within it are additive only, and the shape guard (`testdata/snapshot_shape/v9.txt`) records every field. Fixtures are never edited (#522).

### What the rules ask for (numbers from the brief; text to be quoted by the slices)

- **CR 901.1–901.4**: Planechase is a casual variant played with a planar deck of plane (and phenomenon) cards and a planar die.
- **CR 901.5**: setup, including the first plane put face up at the start of the game.
- **CR 901.6**: the planar controller, normally the active player, and who it becomes when the planar controller would leave the game.
- **CR 901.7, 901.7a**: the abilities of face-up plane and phenomenon cards function from the command zone; the planar controller controls them.
- **CR 901.8**: every plane has the planeswalking ability, an inherent triggered ability that planeswalks when the planar die shows the planeswalk symbol.
- **CR 901.9a–d**: rolling the planar die is a special action of the active player, during their main phase, with priority and an empty stack. It costs {N} generic, N being the number of times they have already taken it this turn. Its faces are blank, chaos and planeswalk, and the roll triggers "whenever you roll a die" abilities, though it has no number (also **CR 706.7**).
- **CR 901.10, 901.10a–b, 800.4p**: what happens to the planar cards, and to their abilities, when a player leaves the game.
- **CR 901.11a–c**: planeswalking. The face-up cards go to the bottom of the planar deck and the top card is turned face up; "until you planeswalk away" ends; planeswalking away from *every* face-up plane at once.
- **CR 901.15a–c**: the single communal planar deck: its minimum size, that the planar controller is treated as the owner of every card in it, and how that interacts with leaving the game.
- **CR 311** (planes), **CR 311.7** (the chaos ability), **CR 312** (phenomena), **CR 116** (special actions).

---

## Decision

### 1. The table option

```go
// game/settings.go
type TableSettings struct {
    // …
    // Planechase plays the game with a communal planar deck and the
    // planar die (CR 901, ADR 0216). Fixed once the game is active.
    Planechase bool `json:"planechase,omitempty"`
}

type SettingsPatch struct {
    // …
    Planechase *bool `json:"planechase,omitempty"` // refused after Start: ErrPlanechaseLocked
}

const SettingPlanechase = "planechase"
```

- **Where it lives.** `Game.Settings.Planechase`, the struct ADR 0075 §2.2 made the one home for a table's configuration. It is cloned, carried by the snapshot (so a restored game keeps it), and public on the wire as `settings.planechase`. `omitempty` keeps every existing restore point and fixture byte-identical.
- **How it is chosen.** `POST /games` takes `"planechase": true` (`createGameRequest.Planechase`), applied through `UpdateSettings` exactly as `PATCH /games/{id}/settings` would apply it. The host can still change it with `PATCH /games/{id}/settings` while the table is in the lobby. A saved setup (ADR 0110 §5) does not carry it in this ADR's slices.
- **When it locks.** `UpdateSettings` refuses a `Planechase` patch once `State != StateLobby` with `ErrPlanechaseLocked`, the `FreeMulligans` rule. Turning Planechase on or off mid-game would be a different, larger feature.
- **How it reaches the engine.** `Game.Start` reads `g.Settings.Planechase` and, when it is set, runs `setUpPlanechaseLocked` (§3) after the libraries are shuffled and before the opening roll. Every Planechase code path is behind `g.Planechase != nil` (§3), so a table without the option runs no new code.
- **Stats.** The `games` table does not record the option in these slices. Counting Planechase games is a later `gamestats` question, not a reason to touch the privacy inventory.

### 2. Building the communal planar deck

- **The source is the card database, not a sideboard.** A new hook in the shape of `OutsideCollectionSource`:

  ```go
  // game/planechase.go
  // PlanarCardSource is the card database's planar cards (ADR 0216 §2).
  // main.go installs it from deck.NewCollection(cardIdx). Nil means no
  // index: a Planechase table then refuses to start (ErrNoPlanarCards).
  type PlanarCardSource interface {
      // PlanarCards returns one game card per planar oracle ID, built by
      // deck.toGameCard, in oracle-ID order.
      PlanarCards(kinds PlanarKinds) []Card
  }
  ```

  `deck.Collection.PlanarCards` walks the index and keeps a printing when it is `Playable()`, its Scryfall `layout` is `"planar"`, and its type line starts `Plane —` (or `Plane `). One card per `oracle_id`, the printing chosen by `printingRank` as for any card a deck names no printing of. `PlanarKinds` is `PlanarPlanes` in slice 1; slice 5 adds `PlanarPhenomena` (type line `Phenomenon`) and the setup asks for both. The order is oracle-ID order so that minting, and so ADR 0054's per-game object ordinals, are the same on every run of a seeded game.
- **ADR 0198 is amended, not reversed.** `notInACollection` keeps planes and phenomena out of the *outside-the-game* pool: a wish still never hands a player a plane, because a planar card is not a card a player owns in a Commander game. `PlanarCards` is a second, separate reader of the same index that selects exactly the planar rows that `collectible` excludes. The two never share a result.
- **Size check (CR 901.15a).** `setUpPlanechaseLocked` compares the deck with the minimum CR 901.15a sets for a single planar deck (slice 1 quotes the number and any other condition from the pinned text). The database holds far more planes than any such minimum, so in practice this only fires on a test index or a broken dump. A deck below the minimum refuses `Start` with `ErrPlanarDeckTooSmall` and the lobby says so. The game never starts with an illegal planar deck, and the server never chooses which cards to drop. One card per oracle ID keeps the deck free of duplicates, whatever 901.15a says about them.
- **Minting.** Each card is minted fresh (`mintFromCollectionLocked`'s shape: a new instance ID, a class-1 object ordinal from the creation counter) and pushed into the planar deck. The deck is then shuffled once on its own stream, `rngStream{kind: rngStreamPlanarShuffle, player: uuid.Nil}`, so a seeded game's planar deck is reproducible and no player's library stream moves.
- **The planar controller owns all of it (CR 901.15b).** Every card in the planar deck and every face-up planar card has `Owner` and `Controller` stamped with the planar controller. `setPlanarControllerLocked(id)` is the one writer of `PlanechaseState.Controller`, and it re-stamps every planar card in the same step, so "you" on a plane, ownership tests and the CR 800.4a sweep all read ordinary card fields.

### 3. Game state

```go
// game/planechase.go
// PlanechaseState is a Planechase game's planar cards and planar die
// (CR 901, ADR 0216). Nil on Game when the option is off.
type PlanechaseState struct {
    // Mode is PlanarDeckCommunal ("communal", CR 901.15). The only value
    // these slices write. CR 901.3's per-player decks would add
    // "per_player" and a Player.PlanarDeck; every reader goes through
    // planarDeckForLocked(owner), so that is one accessor's change.
    Mode PlanarDeckMode
    // Controller is the planar controller (CR 901.6). Written only by
    // setPlanarControllerLocked.
    Controller uuid.UUID
    // Deck is the communal planar deck: kind ZoneCommand, face down,
    // ordered, top last. Outside the card index; no ZoneRef reaches it.
    Deck *Zone
    // FaceUp is the face-up plane (and, from slice 5, phenomenon) cards:
    // kind ZoneCommand, outside the card index, in the order turned up.
    // Usually one card; more after a "planeswalk to two planes" effect.
    FaceUp *Zone
    // DieTurnSeq and DieRolls count planar die special actions this turn
    // (CR 901.9's cost). A count whose DieTurnSeq is not Turn.Seq is 0,
    // so no turn-start hook resets it.
    DieTurnSeq int
    DieRolls   int
}

// on Game:
Planechase *PlanechaseState
```

- **Why not `Player.Command`, emblems or a new `ZoneKind`.** For ADR 0207's reasons: every reader of `Command.Cards` means "commander card", and rules-wise the planar cards *are* in the command zone, so events and logs should say `command`. A container outside the card index means nothing except planeswalking (and slice 5's encounter) can move a planar card, and that is enforced by the container, not by a check someone has to remember.
- **Why game-level and not on a seat.** The communal deck belongs to whoever is planar controller this turn. Putting it on a `Player` would mean moving it every turn, and a departure sweep (CR 800.4a) would take it. Per-player decks, if they come, go on `Player` beside it.
- **Setup (CR 901.5).** `setUpPlanechaseLocked` sets the planar controller to the starting player once the opening roll settles (`setPlanarControllerLocked`), and puts the top card face up as CR 901.5 says. Slice 1 reads two things from the pinned text and implements them as written: whether that first card counts as planeswalking to it (that is, whether "whenever you planeswalk to ~" triggers), and what happens when the top card is a phenomenon (relevant from slice 5).
- **The planar controller each turn (CR 901.6).** `setPlanarControllerLocked(activePlayer)` runs as each turn begins, in the turn-machinery step that sets `Turn.ActiveSeat`. CR 901.6's departure rule is §8.
- **Snapshot, additive within v9.** `GameSnapshot` gains `Planechase *planechaseSnapshot \`json:"planechase,omitempty"\`` with `mode`, `controller`, `deck` and `faceUp` (`*zoneSnapshot`, written by `snapshotCompanionZone` and read by `restoreCommandPile`, the attraction deck's pair), `dieTurnSeq` and `dieRolls`. All `omitempty`. `TableSettings.Planechase` is one more `omitempty` field on the settings the snapshot already carries. The shape file `v9.txt` records the new fields in the same PR. No fixture is edited or regenerated, and none needs to be: no existing restore point has the key, and absent means "no Planechase". No schema bump.
- **Clone and undo.** `Clone` deep-copies `PlanechaseState` (both zones). Undo rewinds it with the rest of the game, and the die stream rewinds as ADR 0054 Decision 4 says: undoing a planar roll and rolling again in the same turn gives the same face. That is the decided cost of rewinding randomness, and the {N} cost is rewound with it.
- **Restore points.** Every stack item this ADR adds names a catalog row or a registered body (ADR 0041 P9, §5 and §6), and every pause is a keyed option pick, so a table waiting on any of them is a restore point, and a restore never drops one. The planar zones are plain cards, restored like the attraction deck.
- **The closure ratchet.** `Game.Planechase` adds two routes to cards, `PlanechaseState.Deck` and `PlanechaseState.FaceUp`. They are classified `census:IntrinsicAbilityCards`, as `Player.AttractionDeck` and `Player.Junkyard` are. That ceiling has been held as an owner-approved number (74, "the raise Ian approved for #4361 only"), so slice 1 needs the owner's approval to take it to 76, with the reason written beside it in `closure_fields_test.go`. See [Questions for the owner](#questions-for-the-owner).
- **The drift test** marks every new field `carried`.

### 4. The planar die: a special action

```go
// special_action.go
// SpecialActionPlanarDie is CR 901.9 (a CR 116 special action): the
// active player rolls the planar die, paying {N} generic where N is the
// number of times they have taken this action this turn. ADR 0216.
SpecialActionPlanarDie SpecialActionKind = "planar_die"
```

- **No card.** Like `ignore_effect`, the kind branches in `PerformSpecialAction` before the card lookup: `if kind == SpecialActionPlanarDie { return g.performPlanarDieLocked(p, params) }`. The wire's `card_id` is absent for this kind, and one that is sent is refused (`ErrInvalidParam`). The offer exists only while `g.Planechase != nil`.
- **The window.** `SpecialActionTimingOKLocked` gets a `SpecialActionPlanarDie` row: the actor is the active player, and `SorcerySpeedOpenLocked`, the predicate plot, unlock and companion use (main phase, priority, empty stack). It is legal under split second for the usual reason, though the empty-stack test means that never matters.
- **The cost.** `PlanarDieCostLocked(player)` is `"{N}"` with N = `DieRolls` when `DieTurnSeq == Turn.Seq`, else `"{0}"`. It is priced through `SpecialActionManaCostForEffect` (the CR 601.2f pass ADR 0062 added in its 2026-09-23 amendment), so a future cost modifier on the planar die needs no new door, and paid by `payAbilityManaCostLocked` with the zero spend context (`specialActionSpendContext`): mana that can only be spent to cast spells cannot pay it. The count goes up when the action is taken, before the roll. Slice 2 confirms against CR 901.9 whether a planar die rolled by an *effect* (Fractured Powerstone) also counts toward N, and writes that into `RollPlanarDieForEffect` below.
- **The roll goes through the dice door.** `DiceRoll` gains `Planar bool` (data, so the paused roll still rides `PendingChoice.dice` as a restore point). `performPlanarDieLocked` calls `RollDiceThenForEffect(DiceRoll{Player: p.ID, Sides: 6, N: 1, Planar: true, Then: planarRollTail})`. A planar die is still a die, so the `RepEventRollDice` window, the order prompt and Squid Fire Knight's "roll one more and ignore one" apply to it unchanged, with the ignore prompt showing faces instead of numbers. Step 3 draws from `rngStream{kind: rngStreamPlanarDie, player: roller}`, a stream of its own so a planar roll never moves a card's d20. A natural 1 is chaos, a 6 is planeswalk, and 2–5 are blank: one face each for chaos and planeswalk, and four blanks.
- **No number (CR 901.9d, CR 706.7).** Each kept die emits `EventRollDie` with `Sides: 6`, `Amount: 0` and a new `PlanarFace string` (`"blank"`, `"chaos"` or `"planeswalk"`, `omitempty`). So "whenever you roll one or more dice" (`WheneverYouRollDice`, Vexing Puzzlebox) triggers, as CR 901.9d says it must. A watcher that matches a natural result ("whenever you roll a 20") never matches, because 0 is no die's result. Slice 2 audits every existing `EventRollDie` reader for the ones that *sum* or *compare* a result, and makes each one skip a planar roll explicitly (`ev.PlanarFace != ""`) rather than relying on the 0. A card that cannot read "no result" correctly gets a caveat, not a guess.
- **What the faces do.** The tail does nothing itself. The face is an event, and the abilities that care are triggers: the planeswalking ability (§5) on planeswalk, a plane's chaos ability (CR 311.7) on chaos. Both go on the stack under the planar controller and resolve normally. A blank does nothing beyond the dice triggers.
- **The effect API.** `RollPlanarDieForEffect(player, source uuid.UUID) error` is the door for a card that says "roll the planar die" (Fractured Powerstone, later). It is the same roll with `Source` set and no cost, and it is refused when the table is not playing Planechase.
- **Log and wire.** The roll is one `LogRoll` entry with a new `planar_face`: "Alice rolled the planar die: chaos". `EventSpecialAction` carries the label `Roll the planar die {N}`.
- **ADR 0054 fit.** Nothing in ADR 0054's model changes: one door, keyed streams, undo rewinds, `BatchSeq`, the log line, the table-wide animation (ADR 0121). The planar die is one more stream kind and one more field on the event.

### 5. Planeswalking

- **The planeswalking ability (CR 901.8)** is not printed in each card file. It is inherent, so it is one registered body, `planeswalkBody = DelayedBody("planechase/planeswalk", …)`, and one `Listener` registered from `NewGame` beside `initiativeTriggers`. On an `EventRollDie` with `PlanarFace == "planeswalk"` rolled by the planar controller, it queues one `StackItem` per face-up plane, with `Source` the plane, `Controller` the planar controller and `Body` the key. It is a real triggered ability: it uses the stack, can be responded to, and a table waiting with it on the stack is a restore point. When it resolves it planeswalks (CR 901.11). Slice 2 checks CR 901.8 for whether, with two face-up planes, both trigger, and whether a second resolution finds the first already done; the body is written so that a planeswalk whose source is no longer face up does nothing.
- **The one verb.** `g.planeswalkLocked(cause PlaneswalkCause)` does CR 901.11 in order:
  1. **Away from every face-up card (CR 901.11c).** For each card in `FaceUp`, emit `EventPlaneswalkFrom` (`Actor` the planar controller, `CardID` the plane), then move it to the bottom of `Deck`. When more than one card goes, the planar controller chooses their order through a keyed option pick (`planechase/bottom-order`), so the pause is a restore point. With one card there is nothing to ask.
  2. **End "until you planeswalk away" (CR 901.11b).** Every duration of the new kind `UntilPlaneswalkAway` (below) whose `Source` is a card that just left `FaceUp` ends here, in the same step as the move.
  3. **To the top card.** Move the top of `Deck` to `FaceUp` and emit `EventPlaneswalkTo` (`Actor`, `CardID`). From slice 5, a phenomenon turned up is encountered instead (§9).
  4. Emit `LogPlaneswalk` ("Alice planeswalked to Akoum, away from Academy at Tolaria West").
- **Cards that say "planeswalk to" a named plane** use the same verb with a target card (`PlaneswalkToForEffect(player, source, card)`), which slices 4 and 5 add when a card needs them (Interplanar Tunnel, Spatial Merging). Nothing in the verb assumes the top card.
- **"Until you planeswalk away" (CR 901.11b).** A new `DurationKind`, `UntilPlaneswalkAway`, appended before `durationKindEnd`, with `Duration.Source` the face-up card whose effect made it. It is a persisted int like every kind, so an older binary refuses a restore point that holds one (ADR 0041 P4), the normal rollback posture. `effects.UntilYouPlaneswalkAway()` is the card-side constructor.
- **"Whenever you planeswalk to / from ~"** are ordinary catalog triggers on `EventPlaneswalkTo` and `EventPlaneswalkFrom`, written with `effects.WhenYouPlaneswalkToThis(label, effect)` and `effects.WhenYouPlaneswalkAwayFromThis(label, effect)`. "Away from" triggers on a card that has just left `FaceUp`, so the harvester looks back at it the way `harvestLTB` looks back at a permanent that left: `harvestFromPlanesLocked` checks the event's `CardID` against the deck's bottom for that event as well as the face-up set.

### 6. How a plane's abilities register

- **One file per card**, `server/internal/cards/effects/<snake_name>.go`, one `init()`, keyed by the plane's Scryfall `oracle_id`, exactly as for any card (docs/adding-cards.md). There is no plane registry beside the catalog: a plane is a card.
- **Statics from the command zone (CR 901.7).** A plane's static abilities are read only while it is in `FaceUp`. `planarContinuousEffectsLocked` joins the layer pass beside `emblemContinuousEffectsLocked`, through the same `staticAbilitiesOf` accessor, so ability removal behaves the same. The rule gates' walk widens from `forEachEmblemLocked` to `forEachCommandSourceLocked`, which walks emblems and then the face-up planar cards, so a cast, land-play or untap gate on a plane needs nothing new. A plane's cost modifiers are gathered by `appendPlanarCostModifiersLocked`, a sibling of ADR 0140's `appendEminenceCostModifiersLocked` called from `activeCostModifiersLocked`. It takes every modifier of a face-up planar card, not only those marked `Eminence`, because all of a face-up plane's abilities function from the command zone, and it binds `Source.Controller` to the planar controller.
- **Triggers from the command zone.** `harvestFromPlanesLocked` walks `FaceUp` as declared command-zone objects, the dungeon's shape. The controller is the planar controller (CR 901.7a). A plane's triggered ability, its chaos ability included, goes on the stack as `Body: "catalog/triggered"` with `Params.Ability` naming its catalog row (ADR 0041 P9), so a restore never drops it.
- **The chaos ability (CR 311.7)** is a triggered ability like any other, written `effects.WheneverYouRollChaos(label, effect)`: it watches `EventRollDie` with `PlanarFace == "chaos"` whose roller is the planar controller.
- **No new card-side keys.** A plane's catalog key is its oracle ID, its abilities' refs are slot refs, and `planechase/planeswalk` and `planechase/bottom-order` are new on-disk identities (never renamed, never reused, as ADR 0083 and #1497 require of every key).
- **Completeness is unchanged.** A plane declares `Completeness` truthfully, `Unreviewed` by default, with player-facing `Caveats`. "Stronger than printed is the wrong direction" holds: a plane whose clause the engine cannot express is left out of the catalog.
- **A plane with no catalog row still plays (PLAN.md §5.1).** Owner decision 2 puts every plane in the deck, whether or not it has a row. An uncatalogued plane turns face up, shows its text, and is planeswalked away from by the inherent ability like any other. Its chaos and static abilities do nothing automatically and show as `manual`, and the table settles them with the sandbox, as with any card the engine does not know.
- **The clone gate.** Plane files are ordinary catalog files, so `TestNoNewExactClonesInTheCatalog` applies unchanged. A chaos or planeswalk shape that two planes share goes in a shared, append-only helper (`effects_planechase.go`).
- **Coverage tools.** The census (`cards/coverage`) counts a plane's catalog key like any other key. A plane is never in a deck list, so `deckcoverage` and the deck checker never see one (it already routes nothing to them), and the Commander-card roadmap's denominators, which count cards a deck can hold, must not count planes. Slice 4's first batch adds a separate census line, "planes and phenomena: N of M in the card database", so planar coverage is visible without inflating the Commander numbers. Oracle fixtures are generated per card with `-update-oracle -oracle-ids=…`, as for any card PR.

### 7. Phenomena: designed now, shipped in slice 5

- **Same deck, same container.** A phenomenon is a planar card in `Deck`, minted by `PlanarCards(PlanarPlanes|PlanarPhenomena)`, owned by the planar controller.
- **Encounter (CR 312).** When `planeswalkLocked` turns up a phenomenon, it emits `EventEncounter` and the phenomenon's "when you encounter ~" trigger (a catalog trigger, `effects.WhenYouEncounterThis`) goes on the stack from `FaceUp`. Once the phenomenon is face up and no ability it is the source of is waiting or on the stack, the planar controller planeswalks away from it. That is the dungeon-completion shape (`dungeonCompletionSBALocked` waiting on `roomAbilityOutstandingLocked`), and slice 5 quotes CR 312 for the exact trigger and timing before writing it.
- **The setup.** CR 901.5's rule for a phenomenon on top at the start of the game is implemented in slice 5, in the function slice 1 writes.
- **Nothing in slices 1–4 assumes `FaceUp` holds only planes.** The wire carries `type_line`, the client draws whatever is face up, and the die is offered whatever is face up.

### 8. Leaving the game (CR 901.10, 901.10a–b, 800.4p; amends ADR 0060)

- **The communal deck is never swept.** The order is what makes this work. When the planar controller would leave the game, `leaveGameLocked` first hands the designation on as CR 901.6 says (`setPlanarControllerLocked(next)`, which re-stamps ownership, §2), and only then runs `removeObjectsOwnedByLocked` (CR 800.4a). By then no planar card is owned by the departing player, so nothing in `Deck` or `FaceUp` leaves the game.
- **Abilities and effects.** A planar ability on the stack, or an "until you planeswalk away" effect, whose controller is the departing player is handled as CR 901.10a–b and 800.4p say. Slice 2 quotes them and implements them as written. The design commitment here is only that the behaviour lives in `leave_game.go`'s departure table (ADR 0060's drop-action column), one row per case, not in the planechase code.
- **The game ending.** ADR 0060 Decision 5 holds: a departure that ends the game strips nothing.
- **Per-player decks (later).** CR 901.10's general rule, a departing player's own planar cards leaving with them, is what `removeObjectsOwnedByLocked` already does to any container it is taught about. A per-player `Player.PlanarDeck` joins that sweep the way `AttractionDeck` did.

### 9. Protocol (all fields additive)

- **`GameView.planechase`** (`omitempty`, absent when the option is off):

  ```json
  "planechase": {
    "planar_controller": "<player uuid>",
    "deck_count": 74,
    "face_up": [ CardView, … ],
    "die_rolls": 1,
    "die_cost": "{1}"
  }
  ```

  `face_up` are ordinary `CardView`s (public to every viewer: the command zone is public), with the existing `layout` (`"planar"`) and `type_line`, so the client can tell a plane from a phenomenon and lay it out landscape. `deck_count` is public like a library's size, and the deck's cards are on nobody's wire. `die_cost` is the active player's current cost, already priced by the CR 601.2f pass.
- **`settings.planechase`** (bool) on the existing settings view, and `planechase` on `POST /games` and `PATCH /games/{id}/settings` ([docs/lobby.md](../lobby.md)).
- **The action** is the existing `special_action` with `"kind": "planar_die"` and no `card_id`, plus the usual `strict` / `auto_tap`. It is in `legal_moves` when it is offered, with its cost.
- **Events and log.** `EventRollDie.planar_face`, `EventPlaneswalkTo`, `EventPlaneswalkFrom` (and `EventEncounter` in slice 5). The log gains `LogEvent.planar_face` on `roll` and a new kind `planeswalk`. `docs/protocol.md` and `client/src/lib/protocol.ts` change in the slice that adds each field.

### 10. The client (DESIGN.md)

- **The lobby toggle.** A "Planechase" switch in the table's settings (`TableSettingsPanel.svelte`, on the lobby card), with a one-line description ("A shared deck of planes and the planar die"). It is locked once the game starts, as free mulligans are, and shown read-only to non-hosts. `POST /games` sends it when the creator sets it from the new-table flow.
- **The face-up plane.** A new `board/PlanarZone.svelte`: the face-up card drawn landscape (a plane is printed sideways) at the size of a command-zone card, its name, the planar controller's seat colour as a hairline, and the deck count as a mono `.tag` ("74 planes"). Hover, focus or a long-press opens the existing hover zoom (top-right) with the card image and its oracle text, so the text is always one gesture away and is read from Scryfall, not written by C2. Where it sits follows DESIGN.md §5: not over the dock (`--dock-w`, `--dock-h`), the stack pile (left), the hover zoom (top-right) or any avatar. The proposal is the centre of the table, between the opponents' row and your board, and slice 3 settles it with screenshots at 820px and at a desktop width.
- **The die.** An action-dock button, "Roll planar die", with its current cost shown as a mana pip (`{0}`, `{1}`…), offered to the active player in their main phase. Like every dock button it is **disabled, not hidden** (ADR 0111 §1) when the window is shut, with the reason in its title. The press sends the special action, and payment uses the existing auto-tap and strict flow.
- **The result.** `DiceLayer` animates the roll for the whole table (ADR 0121) with a planar die whose faces are blank, the chaos symbol and the planeswalker symbol. The two symbols come from Scryfall's symbology like mana symbols do, never drawn by C2 (Fan Content Policy). The log line names the face. A planeswalk animates the old plane out and the new one in, respecting reduced motion.
- **Labels.** Every name a test or hint uses is a contract label in `client/src/lib/labels.ts`, rendered as `L.<key>`: `L.planarZone` ("planes", aria, the group), `L.planarDeck` ("planar deck", aria, the count), `L.rollPlanarDie` ("Roll planar die", text, the dock button) and `L.planechaseSetting` ("Planechase", text, the switch). `docs/labels.md` is regenerated with them.
- **First use.** A `PlanarZone.hint.ts` tip ("The plane is face up. Roll the die on your turn for chaos, or to planeswalk.", within ADR 0125 §3.4's limits) the first time a player sits at a Planechase table.

### 11. The bot seat

The heuristic seat rolls the planar die only when it is free: once, in its first main phase, with an empty stack, before it casts anything, and never at a cost above {0}. It answers the bottom-order pick with the order given. That is all in these slices. Pricing a chaos ability, paying {1}+ to roll again, and planeswalking away from a plane that hurts it are later, with an arena measurement (docs/bot-testing.md). The legal enumerator offers the action with its cost, so a model seat and the MCP seat see it like any special action.

### 12. Slices

Each slice is one mergeable PR with its own tests, in this order.

1. **State, option and communal deck** (server). `TableSettings.Planechase`, `SettingsPatch`, `ErrPlanechaseLocked`, `createGameRequest.Planechase`. `PlanarCardSource`, `deck.Collection.PlanarCards`, `main.go` wiring. `PlanechaseState`, `Game.Planechase`, `setUpPlanechaseLocked` (CR 901.5, quoted), `setPlanarControllerLocked` (CR 901.6, quoted) and ownership stamping (CR 901.15b), the size check (CR 901.15a, quoted), the planar shuffle stream. Clone, snapshot (`v9.txt` shape), drift test, the two closure-census lines and the ceiling (owner approval). `GameView.planechase` with no action yet. ADR 0198's amendment is carried out here. *Tests:* a seeded game's planar deck is the same on two runs; the deck holds one card per planar oracle ID and no other card; the size check refuses `Start`; the patch is refused after `Start`; a table without the option has `Planechase == nil` and an unchanged snapshot; the controller and ownership follow the active player; snapshot round trip and clone independence; the restore of an existing v9 fixture is unchanged.
2. **The die and planeswalking** (server). `SpecialActionPlanarDie` and its timing and cost, `DiceRoll.Planar`, `rngStreamPlanarDie`, `EventRollDie.PlanarFace`, the audit of every existing roll watcher, `RollPlanarDieForEffect`, the planeswalking ability (`planechase/planeswalk`), `planeswalkLocked` with the bottom-order pick, `EventPlaneswalkTo` and `EventPlaneswalkFrom`, `UntilPlaneswalkAway`, the command-zone walks (layers, harvester, gates, cost modifiers), the leave-game rows (CR 901.10, 901.10a–b, 800.4p, quoted), the log lines, the legal enumerator, `docs/protocol.md`. With it, three planes that exercise the seam: one static, one chaos trigger and one planeswalk-to trigger, chosen from the newest set that prints planes. *Tests:* the cost goes {0}, {1}, {2} within a turn and back to {0} next turn; only the active player, in a main phase, with an empty stack; a planar roll triggers Vexing Puzzlebox and matches no numeric watcher; Squid Fire Knight's sticker applies to it; undo rewinds the face and the count; a planeswalk trigger on the stack is a restore point; planeswalking with two face-up planes; a duration ends on planeswalking away; the planar controller leaving mid-turn hands on the deck and sweeps none of it.
3. **The client.** The lobby switch, `PlanarZone.svelte`, the dock button with its cost, the die faces in `DiceLayer`, the log lines, the labels and the hint, an e2e spec that creates a Planechase table, rolls, and sees a planeswalk. Screenshots at 820px in the PR.
4. **Plane cards, in batches.** One PR per batch of about fifteen, newest set first (S79's order): the newest Commander product that prints planes, then Planechase Anthology, Planechase 2012 and Planechase (2009). Each batch is ordinary catalog work with per-card oracle fixtures and truthful `Completeness`. The first batch adds the planar census line.
5. **Phenomena.** `PlanarPhenomena` in the deck, `EventEncounter`, `effects.WhenYouEncounterThis`, the planeswalk-after-encounter check (CR 312, quoted), the CR 901.5 phenomenon case, and the phenomenon cards themselves in one or two batches.
6. **The bot.** The free-roll policy above, with a bot-arena run showing it never stalls a table, and a regression test in `aiseat`.

---

## Consequences

- **A Planechase table is a Commander game with one more object in the command zone.** Nothing outside `g.Planechase != nil` changes, so the risk to ordinary tables is the fields that are always present (`omitempty`) and the widened command-zone walks, which are empty loops without the option.
- **Planar cards play before they are catalogued.** Every plane is in the deck from slice 1, and the inherent planeswalking ability works for every one of them. Until slice 4, most chaos and static abilities are `manual`. That is the charter's fallback working as designed, and it is also why the prod toggle's timing is an owner question.
- **One more stream, one more special action kind, one more duration kind.** Each follows its seam's existing rule, so the tests that hold those seams (`TestNoDirectRandomSource`, the special-action timing table, `DurationKind.Known`) cover them.
- **ADR 0054, 0060 and 0198 each carry a dated note** pointing here.

## Risks

- **The rule text is not yet checked** (see the note at the top). The design depends on it at named points; each slice quotes it before writing code.
- **The closure-census ceiling.** Slice 1 cannot merge without the owner's approval for +2.
- **Dice watchers that read a result.** A planar roll has no number. The slice 2 audit is the guard; a watcher it misses would, at worst, add 0.
- **Squid Fire Knight and the planar die.** Routing the planar die through the dice door makes the sticker apply, which is right by the rules but is the one place where a planar roll can pause. The pause is a keyed pick, so it is a restore point, and the cost is counted once per action, not per die.
- **Table space.** A landscape card at a four-seat table at 820px is the client's hardest part. Slice 3 has to show it fits without covering the dock or the stack.
- **Rewinding randomness.** As with every roll, undoing a planar roll shows the same face again this turn (ADR 0054 Decision 4). That is consistent and decided, not new.

## Alternatives considered

- **Per-player planar decks first (CR 901.3).** Rejected by the owner: it needs deckbuilding, a deck-list section and seating rules, and most groups play the communal deck. The design leaves room (`PlanechaseState.Mode`, `planarDeckForLocked`, the departure sweep).
- **The die as an activated ability of a "game" object.** Rejected: CR 901.9 makes it a special action, which doesn't use the stack and can't be responded to. Modelling it as an ability would give opponents a response window the rules don't, and would put it in the activated-ability enumerator, cost pass and stack, all with special cases. `PerformSpecialAction` already has the window, the payment and the event.
- **Planes as emblems.** Rejected: an emblem is owned by one player forever and is not a card (CR 114). A plane is a real card with an oracle ID, a printing and art, turned face up and put back, and its owner changes every turn. The emblem walks are reused (§6); the emblem container is not.
- **Planes in `Player.Command`.** Rejected for ADR 0207's reason: every reader of `Command.Cards` means "commander card".
- **A new `ZoneKind` for the planar deck.** Rejected for ADR 0207's reason: rules-wise it is the command zone, and the log should say so.
- **A curated, catalogued-only planar deck.** Rejected by owner decision 2 ("every plane card in the card pool"). It would also make the deck's contents depend on catalog progress.
- **The planeswalking ability written into each plane's card file.** Rejected: it is inherent, and a hundred copies of it would be a hundred chances to get it wrong. One keyed body, one listener.

## Owner answers (Ian, 2026-11-04)

1. **The pinned CR text.** Checked before merge against `MagicCompRules 20260925.txt` on the owner's machine. All 33 rule numbers this ADR cites exist in the pinned text, and the ones it paraphrases (901.9, 901.11, 901.15a, 311.7, 800.4p, 116) match the rules' wording.
2. **The closure-census ceiling.** Approved: `census:IntrinsicAbilityCards` may go from 74 to 76 for `PlanechaseState.Deck` and `PlanechaseState.FaceUp` (slice 1).
3. **When prod shows the toggle.** Prod waits until Planechase is implemented: the toggle ships to cmd-dev with slice 3, and prod offers it only once slice 4's plane batches have landed, so a prod table never meets a mostly manual planar deck.

## Amendment 2026-11-04 (#4870, slice 1): the planar containers live on `Game`, and what CR 901 settles

Slice 1 quoted CR 901.4, 901.5, 901.6, 901.11 and 901.15a–c verbatim in the header of `server/internal/game/planechase.go`. Three readings and one change of shape follow.

1. **`PlanechaseState` holds no cards; the two containers are `Game.PlanarDeck` and `Game.PlanarFaceUp`.** §3 put `Deck` and `FaceUp` on `PlanechaseState` and counted two closure routes. With that shape the closure census sees three: `Game.Planechase` is an edge into a struct that reaches cards, as `Game.Seats` is beside `Player.AttractionDeck`, so the ceiling would have gone to 77, past the approved 76. Putting the two `*Zone` fields on `Game` and keeping `PlanechaseState` plain data (`Mode`, `Controller`, `DieTurnSeq`, `DieRolls`) makes it exactly the two approved routes. `census:IntrinsicAbilityCards` is 76. Nothing else changes: the snapshot still writes one `planechase` object with `mode`, `controller`, `deck`, `faceUp`, `dieTurnSeq` and `dieRolls`; every reader still goes through `planarDeckForLocked(owner)` (and `planarFaceUpLocked()`); and `g.Planechase != nil` is still the switch. Slices 2 and on read `g.PlanarDeck` and `g.PlanarFaceUp` where this ADR says `PlanechaseState.Deck` and `PlanechaseState.FaceUp`.
2. **The starting plane is not a planeswalk.** CR 901.11 begins "After the game has started", and CR 901.5 says no ability of a card turned face up this way triggers. `turnUpStartingPlaneLocked` emits no event, so slice 2's "whenever you planeswalk to ~" will not see the starting plane.
3. **A phenomenon on top at setup** goes to the bottom and the next card is turned up, until a plane is (CR 901.5). Slice 1 writes that loop now, so slice 5 does not touch the setup.
4. **CR 901.15a's numbers**: at least forty cards or ten times the number of players, whichever is smaller (`MinCommunalPlanarDeck`); no more phenomena than twice the number of players; every English name different. A deck that breaks any of them refuses `Start` (`ErrPlanarDeckTooSmall`, `ErrPlanarDeckIllegal`), and with no card database `ErrNoPlanarCards`; the lobby answers 409.

**Timing of setup.** CR 901.5 turns the starting plane up once every hand is kept and the opening-hand actions are used. The opening-hand actions (ADR 0133) are prompts the mulligan close queues, so the plane is turned up as they are queued, not after they are answered. Nothing in slice 1 can tell the difference (no plane ability functions yet); slice 2, which makes plane statics function, revisits it if a plane's static can change how a Leyline enters.

**Restart (CR 727).** A restarted Planechase game gets a fresh communal planar deck: the planar cards are in none of the zones CR 727.2 walks. If the database can no longer supply one, the new game continues without the option and the failure is logged.

The wire in slice 1 is `GameView.planechase` with `planar_controller`, `deck_count` and `face_up`; `die_rolls` and `die_cost` come with the die in slice 2.

## Amendment 2026-11-04 (#4870, slice 2): the die and planeswalking, and what CR 901's text settled

Slice 2 quotes CR 901.3a, 901.7, 901.7a, 901.8, 901.9–901.9d, 901.11–901.11c and 901.13 verbatim in `planechase_die.go`, `planeswalk.go` and `planechase_sources.go`. Where the text disagrees with §§4–5 above, the text wins:

1. **One planeswalking ability per roll, not one per face-up plane.** CR 901.8: the ability "has no source and is controlled by the player whose planar die roll caused it to trigger." Its stack item names no card (`SourceCardID` nil), its controller is the roller (anyone an effect makes roll, not only the planar controller), and its body is `planechase/planeswalk`. §5's "one `StackItem` per face-up plane, with `Source` the plane" does not apply.
2. **Only the special action counts toward the die's cost.** CR 901.9's note says the cost "won't be equal to the number of times the player has rolled the planar die that turn if an effect has caused the player to roll". So `RollPlanarDieForEffect` costs nothing and does not move `DieRolls`.
3. **No number at all.** CR 901.9d: the planar roll's `EventRollDie` has `Amount` 0 and `PlanarFace`. The audit of roll readers found one card-side sum, `effects.WheneverYouRollDice` (Vexing Puzzlebox), which now skips a planar roll explicitly. The log line has `planar_face` and no `results`.
4. **A plane turned face down is a new object** (CR 901.7a): its knowers are cleared and its `ObjectEpoch` goes up. A face-up plane's statics are timestamped as it turns face up, as an emblem's are.
5. **The command-zone walks** are the emblem walks over `Game.PlanarFaceUp`. The rule gates now call `forEachCommandSourceLocked` (emblems, then face-up planes). Activated abilities of planes are not walked: no plane the slices plan prints one.
6. **Leaving the game.** CR 901.6's hand-on runs at the top of `leaveGameObjectsLocked`, before the CR 800.4a sweep (`handOnPlanarControllerLocked`), so the communal deck is never swept and CR 901.10's "turn the top card face up" never arises with a single planar deck. CR 901.10a–b (a plane leaving while the ability waits, and phenomena's abilities) cannot happen with a communal deck before slice 5. CR 800.4p was not available to this session and is not quoted; nothing in this slice depends on it.
7. **Rule 701.31 (planeswalk)** was not available either. The order implemented is CR 901.11b–c's: away from every face-up card (the planar controller orders them onto the bottom with `planechase/bottom-order` when there is more than one), then to the top card. Slice 4, whose cards say "planeswalk", checks it against 701.31.

**Deferred from slice 2:** the three sample planes. A plane is a catalog card, and a catalog card needs its oracle fixture, generated from the Scryfall dump. This session's network policy refuses Scryfall. The card-side shapes they need are in `effects_planechase.go` (`WheneverYouRollChaos`, `WhenYouPlaneswalkToThis`, `WhenYouPlaneswalkAwayFromThis`, `UntilYouPlaneswalkAway`), and the engine tests drive the same walks with test triggers. The first slice 4 batch adds them.

## Amendment 2026-11-09 (#4870, slice 3): the client, and the dev-only switch

Slice 3 ships §10. Where it differs from §10, or settles what §10 left open:

1. **The option is a dev feature until slice 4 (owner answer 3).** `appenv.Features.Planechase` (`CMDCTRL_DEV_PLANECHASE`, on by default in a dev deployment, never in production) is installed once at boot as `ws.RoomManager.SetPlanechaseOffered`. That one switch is read by `Lobby.UpdateSettings` (so `PATCH /games/{id}/settings` and a scheduled table's doors-open), by `POST /games` and `POST`/`PATCH /scheduled` before anything is made, and by the hub's `set_table_settings`. Turning the option **on** where it is not offered is refused (`game.ErrPlanechaseNotOffered`, 400 / `bad_request`); turning it off is always accepted. `GET /config` lists `features.planechase`, which is only what the client draws: the lobby's Planechase box and the settings row are shown where it is offered, or on a table that already plays it. When slice 4 lands, the feature moves out of `appenv.Features` and prod offers it.
2. **Where the plane sits.** At the table's left edge, on the line between the opponents' row and your board (35% down in the quadrant layout, whose rows split 0.7 / 1.3), at z 35. Not the centre §10 proposed: the centre covers the most battlefield, and the stack pile, which opens left of centre, would land on it. At the edge it covers a corner of the two boards beside it and nothing else: the stack pile opens right of it, and the dock and the hover zoom are on the right. Checked with a four-seat table at 820px and 1440px, with the stack open and with a die in the air (screenshots on the PR).
3. **The card** is the ordinary `Card` turned a quarter inside a landscape box, so hover, focus, long-press and the art credit are the same as any card's. A plane's Scryfall image is a portrait scan of a turned card, like a split card's, so `isSidewaysArt` now includes `layout: "planar"` and the hover zoom shows it landscape. The direction of the turn is the split card's; this session had no plane images to check it against, so the first look on cmd-dev confirms it.
4. **The die's faces are C2's own glyphs, not the printed symbols.** §10 said the chaos and planeswalker symbols "come from Scryfall's symbology like mana symbols do". Mana symbols are drawn by C2 (`ManaSymbol.svelte`: there is no symbology route, and hot-linking `svgs.scryfall.io` would be the table's only third-party request), and drawing the planeswalker symbol would be drawing a Wizards mark. So the planar die is the d6 `DiceLayer` already draws, with `lib/icons.ts`'s spark on the chaos face, its door on the planeswalk face, and four blank faces; once it lands, its caption names the face in words. The server reads 1 as chaos and 6 as planeswalk, and the client lands on the same numbers (`dice.ts`, `planarFaceNumber`), so the tumble never shows the result early.
5. **The planeswalk animation** slides the new plane in (keyed on its instance), which reduced motion turns off. The old plane is not animated out: it is in no zone the view shows once it is gone, and the log line names it.
6. **Labels.** `L.planarZone`, `L.rollPlanarDie` and `L.planechaseSetting` as §10 says. `L.planarDeck` is dynamic ("planar deck, 74 cards") so the count is part of the name a screen reader hears. The dock button's cost is a pip with a description ("costs {1}"), so its name stays the label.
7. **The e2e spec** (`tests-e2e/tests/planechase-4870.spec.ts`) skips on a server whose `/config` does not offer Planechase, which is every production-env run, the nightly included. It runs against a dev server with the card database loaded, and will run in the nightly once the option is offered in production.

## Amendment 2026-11-09 (#4870, slice 4, batch 1): the plane registration pattern, and what CR 311.7 and 701.31 settled

The first batch of plane cards (the fourteen Doctor Who planes, A–M) settles the pattern every later batch copies; [docs/adding-cards.md](../adding-cards.md#plane-cards-adr-0216-slice-4-4870) has it. Where it differs from §§5–6 or adds to them:

1. **A chaos ability triggers on anyone's chaos roll.** §6 said `WheneverYouRollChaos` watches a chaos roll "whose roller is the planar controller". CR 311.7: "This ability triggers if the chaos symbol is rolled on the planar die (see rule 901.9b), if a resolving spell or ability says that chaos ensues, or if a resolving spell or ability states that chaos ensues for a particular object. … A chaos ability is controlled by the current planar controller." So the planes use `effects.WheneverChaosEnsues`, which watches the chaos face whoever rolled (an effect can make another player roll, CR 901.9's note), controlled by the face-up card's controller. `WheneverYouRollChaos` stays for any card that does print "you roll". No card of the catalog says "chaos ensues" as an effect yet; the batch that adds one widens the shared shape, not the plane files.
2. **"When you planeswalk to ~ and at the beginning of your upkeep"** is one ability with two trigger conditions (`WhenYouPlaneswalkToThisAndAtYourUpkeep`), one row and one label.
3. **CR 701.31b** ("put each face-up plane card and phenomenon card on the bottom of its owner's planar deck face down, then move the top card of your planar deck off that planar deck and turn it face up") matches slice 2's order. Bad Wolf Bay's "Then planeswalk" is `PlaneswalkForEffect`.
4. **Four additive seams**, each behind a face-up plane:
   - `game.SplitSecondGrant.AllSpells` and a command-zone walk in `spellGrantedSplitSecondLocked`: Lake Silencio's "All spells have split second", read live, so a planeswalk away (which now refreshes the split-second cache) takes it off every spell, as its ruling says.
   - `game.ModCardsCantEnterFromExile` (an on-disk mod kind, `cardsCantEnterFromExile`), registered by `CardsCantEnterFromExileThisTurnForEffect` and read by the one entry gate (`scopedEntryRefusalLocked`): Bad Wolf Bay's "cards can't enter from exile this turn" outlives the plane, which planeswalks away in the same resolution.
   - `game.PutCountersOnFaceUpPlanarCardForEffect`: counters on the face-up card (Aplan Mortarium's exposure counters). The planar zones are outside the card index, so the ordinary counter door never reaches them; a planeswalk away clears them (CR 901.7a).
   - `effects.TimeTravel` (CR 701.56a, `option-pick/time-travel`): one keyed option pick per permanent with a time counter and suspended card, add, remove or leave.
5. **Coverage.** The census gains the planar line, "planes and phenomena in the card database with a catalog row: N of M", whose M is `coverage/planar_cards.txt` (151 rows: 139 planes, 12 phenomena), checked against the dump by `TestPlanarCardListMatchesTheDump` when the dump is present.
