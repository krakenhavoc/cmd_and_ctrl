# ADR 0104 — Gaining control of a spell on the stack

**Status:** Proposed · 2026-09-30 · S50 — Seams from the deck re-checks
**Issue:** [#1745](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1745). Invert Polarity is waiting on it in the ViviVoltron deck request, [#1640](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1640).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-09-30. I ran `git fetch --all --prune` and read every `docs/decisions/` file name on all remote heads. Numbers 0103 and 0105 are reserved for ADRs being written in parallel, and 0105 already exists on a branch. No branch has 0104, so this one takes **0104**.
**Builds on:** [ADR 0063](0063-durations-and-control.md) (layer-2 control from a spell or ability), [ADR 0102](0102-entering-under-another-players-control.md) (the would-be controller of an entering permanent), [ADR 0060](0060-leaving-the-game.md) (CR 800.4), [ADR 0019](0019-structured-targeting.md) and its 2026-09-22 amendment (choosing new targets, #1196), [ADR 0054](0054-dice-rolls-and-coin-flips.md) (the won/lost coin flip), [ADR 0043](0043-copy-effects.md) (copying a spell), [ADR 0040](0040-mana-pipeline.md) (mana spent and spend riders), [ADR 0066](0066-granted-cast-and-play-permissions.md) (cast permissions), [ADR 0033](0033-ai-bot-seat.md) (the bot) and [ADR 0041](0041-game-persistence.md) (restore points).

This PR is the ADR only. No engine code changes until the owner has answered the questions at the end.

---

## Context

Invert Polarity ({U}{U}{R}, instant):

> Choose target spell, then flip a coin. If you win the flip, gain control of that spell and you may choose new targets for it. If you lose the flip, counter that spell.

The engine already has the other pieces: the called coin flip (ADR 0054, `FlipCoinForEffect`), countering a spell (`CounterTarget`) and choosing new targets for a spell on the stack (#1196, `ChangeTargets`). What it lacks is a way to **change who controls a spell**.

### The cards

A search of the Scryfall dump for "gain control of … spell" and "exchange control of … spell" finds **six cards**. All six are legal in Commander and none is catalogued.

| Card | Cost | Text (abridged) | Also needs |
|---|---|---|---|
| Invert Polarity | {U}{U}{R} instant | Choose target spell, then flip a coin. Win: gain control of it and you may choose new targets. Lose: counter it. | nothing else |
| Aethersnatch | {4}{U}{U} instant | Gain control of target spell. You may choose new targets for it. | nothing else |
| Commandeer | {5}{U}{U} instant | Pitch two blue cards instead of paying. Gain control of target noncreature spell. You may choose new targets for it. | `AlternativeCost.ExileFromHand` fixed at one card (`cardComponent` returns 1); the pitch needs a count of two |
| Perplexing Chimera | {4}{U} creature | Whenever an opponent casts a spell, you may exchange control of this creature and that spell. If you do, you may choose new targets for the spell. | nothing else |
| Sudden Substitution | {2}{U}{U} instant, split second | Exchange control of target noncreature spell and target creature. Then the spell's controller may choose new targets for it. | nothing else (split second and two-clause targeting both exist) |
| Chef's Kiss | {1}{R}{R} instant | Gain control of target spell that targets only a single permanent or player. Copy it, then reselect the targets at random for the spell and the copy. The new targets can't be you or a permanent you control. | a random retarget with an exclusion (open question 2) |

Two more matched the search and are not this seam: Exert Influence and Skyfire Kirin steal a creature, not a spell.

### How the engine reads a spell's controller today

`StackItem.Controller` (`game/stack.go`) is "the player who cast / activated this item". It is set once, at cast, and nothing rewrites it. Every question about the spell's controller reads that field, live:

- **"You" at resolution.** `effects.Context.Controller()` returns `ctx.Item.Controller`. About 660 card files in `cards/effects` read it, directly or through the context.
- **Target legality.** `stackItemSourceLocked` (`game/targets.go`) builds the `TargetSource` from `item.Controller`. The announce gate, the CR 608.2b re-check at resolution and the retarget offer all use it. So "target opponent", hexproof and protection are judged against whoever controls the spell *now*.
- **A permanent spell's entry.** `resolveTopOfStackLocked` (`game/mutations.go`) builds the entry event with `Actor: item.Controller`, and `landEntryLocked` writes `Card.Controller = ev.Actor`.
- **CR 800.4a.** `cleanupStackForEliminatedLocked` removes every stack item whose `Controller` is the departed player, and exiles the card if it was a spell.
- **Other readers.** These include the exile-cause controller (`move_cause.go`, Ranar), the damage source's controller (`damage_tail.go`), the mana-spend trigger condition (`mana_spend_rider.go`) and the wire's `StackItemView.controller`.

So if the field were rewritten, most of the rules would follow on their own. The work is in the places where the field is **not** enough.

### Four things I found that this ADR has to deal with

1. **The spell card on the stack has a second, stale `Controller`.** `Card.Controller` is stamped with the owner when the deck is loaded (`game.go`). `CastSpell` re-stamps the stack card only for a face-down cast (`mutations.go`). Target predicates over the stack zone read the *card* (`YouControl`, `OpponentControls` in `cards/effects/targets.go`), not the item. Today this only matters when a player casts someone else's card (Gonti, Ragavan, Wrexial). Then Counterflux's "target spell you don't control" judges by the card's owner. A spell that changes controller would make the drift common.
2. **A `ScopedEffect` that gives control to a departed player is not ended.** ADR 0060 §4 says "there is no separate registry of control grants to walk", because every control effect was then an Aura's static ability. Since #756, `GainControlForEffect` writes a `setController` record into `Game.ScopedEffects`. Nothing in `leaveGameObjectsLocked` ends those records, and `modApply` applies them whether or not the player is still in the game. By reading the code, an Act of Treason whose caster concedes leaves the creature under the departed player, and step 4 then exiles it. CR 800.4a's own example says Runeclaw Bears goes back to Bianca. This design leans on that step (§3, §7), so the fix belongs here.
3. **The Adventure permission names the owner. The rule names the controller.** `grantAdventureCastFromExileLocked` (`game/adventure.go`) gives the later cast to `exiled.Owner` and cites CR 715.4. In the pinned rules the clause is **CR 715.3d**: "its controller exiles it. For as long as that card remains exiled, **that player** may play it." Today it only shows when an Adventure is cast off someone else's card. A stolen Adventure spell would make the difference visible.
4. **`docs/engine-seams.md` says control of a spell "is not layer 2 at all".** The rules say it is. CR 611.1: a continuous effect "modifies control of objects". CR 613.1b puts control-changing effects in layer 2. CR 400.7a carries such an effect from a permanent spell onto the permanent it becomes. What is true is that the engine's layer pass only walks the battlefield. That is an implementation choice, and it is the choice the options below are about.

### The rules

Every rule below was checked against the pinned text (`MagicCompRules 20260819.txt`).

- **CR 109.4.** Only objects on the stack or on the battlefield have a controller. **CR 109.5:** "you" on an object is its controller.
- **CR 601.2a.** The player who casts a spell becomes its controller.
- **CR 608.2c.** "The controller of the spell or ability follows its instructions." **CR 608.2b** re-checks targets at resolution. A target that became illegal because the controller changed is illegal ("Other changes to the game state may cause a target to no longer be legal").
- **CR 608.3a / 608.3c.** A permanent spell, and an Aura spell, enters "under the control of the spell's controller".
- **CR 110.2b.** "If an effect causes a player to gain control of another player's permanent spell, the first player controls the permanent that spell becomes, but the permanent's controller by default is the player who put that spell onto the stack."
- **CR 400.7a.** Effects from spells and abilities "that change the characteristics or controller of a permanent spell on the stack continue to apply to the permanent that spell becomes."
- **CR 611.1, 611.2a, 611.2c.** Control changes are continuous effects. With no stated duration, one lasts until the end of the game. The affected set is locked when the effect begins.
- **CR 613.1b, 613.7, 613.7b.** Layer 2. A later effect beats an earlier one. An effect from a resolving spell is timestamped when it is created.
- **CR 115.7d / 115.7e.** "Choose new targets": any number may stay, even illegal ones. Changed ones must be legal. Only the final set is judged.
- **CR 701.6a.** A countered spell goes to its owner's graveyard.
- **CR 701.12a / 701.12b.** If the whole exchange can't happen, none of it does. An exchange between objects the same player controls does nothing.
- **CR 705.2.** The flipper calls the coin and wins or loses.
- **CR 707.10.** A copy "is controlled by the player under whose control it was put on the stack" and is owned by that player. Decisions are copied. Control is not a decision.
- **CR 708.5.** You may look at a face-down spell *you control*.
- **CR 715.3d.** A resolving Adventure is exiled by its controller, and "that player" may cast it later.
- **CR 702.27a (buyback), 702.34a (flashback), 702.109a (dash), 702.185a (warp).** These return or exile to the *owner* or through the *owner*. **CR 702.74a (evoke):** "its controller sacrifices it". **CR 702.174c (gift):** "gives a gift" is read off the controller of the resolving spell.
- **CR 800.4a.** In order: the departed player's owned objects leave; effects giving them control end; stack objects they control that are not cards cease to exist; anything they still control is exiled. **CR 800.4b:** an object never changes to a departed player's control. **CR 800.4c:** if a control effect ends and the default controller has left, the object is exiled.

---

## Options

### A. Rewrite `StackItem.Controller` in place

The steal writes the new player into the field and remembers nothing.

- **For:** one line, and about 660 readers follow for free.
- **Against:** it forgets the three things the rules still ask about.
  - CR 110.2b's default controller. The permanent would enter with the thief as its base, so it never reverts.
  - CR 800.4a's reversion. A thief who leaves would have the spell exiled rather than handed back.
  - Two steals in a row. If the second thief leaves, the spell should go back to the first thief, not to the caster.

### B. Extend layer 2 to the stack

Make the steal a `ScopedEffect` with a `setController` mod pinned to the stack object. Teach the layer pass to walk the stack and write `StackItem.Controller` from the result. Re-pin the record to the permanent at resolution.

- **For:** it is the rule as written. One registry, one duration model and one timestamp sort cover spells and permanents alike.
- **Against:** it is the largest change by far, for six cards. The layer pass, its invalidation, the affected-set pins, the duration sweep's garbage collection and the snapshot's `AffectedObject` all assume battlefield objects. Nothing else wants a stack object in layer 2. No catalogued static ability changes a spell's control, and no printed card does either ("spells you cast" statics change characteristics, which the engine already handles differently).

### C. A control record on the stack item, handed to layer 2 at resolution (recommended)

Keep `StackItem.Controller` as the one value everyone reads. Record *why* it is what it is as data on the item, and hand that record to the existing layer-2 machinery at the moment CR 400.7a says it carries over.

- **For:** every reader stays as it is. The three things option A forgets are data that can be replayed. The permanent side is `ScopedEffect`, which already exists. Everything is plain data, so a table with a stolen spell is a restore point.
- **Against:** the stack gets a small layer-2 answer of its own, alongside the battlefield's. It is one function over a list sorted by timestamp. It exists only because nothing but a resolving spell ever changes a spell's control, and while the object is on the stack that is the only kind of effect it needs to answer.

---

## Decision (proposed): option C

### 1. How control of a stack object is represented

Two new fields on `StackItem`:

```go
// DefaultController is the player under whose control this item was
// put on the stack — the caster, or for a copy the player who made it
// (CR 110.2b, CR 707.10). Zero means "Controller is the default",
// exactly as Card.BaseController does on the battlefield; the first
// control change stamps it. No build site changes.
DefaultController uuid.UUID

// ControlChanges are the effects that have given this spell to another
// player since, oldest first (CR 611.1, CR 613.7b).
ControlChanges []SpellControlChange

type SpellControlChange struct {
	Player     uuid.UUID // who gained control
	Timestamp  int64     // CR 613.7b: when the effect was created
	Source     ObjectRef // the spell or ability that did it
	SourceName string
}
```

**One materialiser:** `stackControllerLocked(item)`. It returns the player of the latest change who is still in the game, or `DefaultController`. `StackItem.Controller` is always its answer, so no reader changes.

**One writer:** `GainControlOfSpellForEffect(sourceID, spellID, player, label) bool`. It:

1. refuses anything but a spell on the stack, a departed player (CR 800.4b) and a no-op (the player already controls it);
2. stamps `DefaultController` if it is zero;
3. appends the change and re-materialises;
4. **re-stamps `Card.Controller` on the spell card in the stack zone**, so the target predicates read the same answer (finding 1);
5. for a face-down spell, adds the new controller as a knower (CR 708.5; see open question 7);
6. emits one `EventSpellControlChanged` (§8).

`CastSpell` stamps the stack card's `Card.Controller` from the caster for every cast, not just a face-down one. That closes finding 1 for the casts that are wrong today.

**The exchange:** `ExchangeControlOfSpellAndPermanentForEffect(sourceID, spellID, permanentID, label) bool` covers Perplexing Chimera and Sudden Substitution. It checks both objects first and does nothing if either is gone, or if one player controls both (CR 701.12a/b). Then it writes the spell's change and registers the permanent's `setController` record with **one shared timestamp**. That is how `ExchangeControlForEffect` already makes two permanents one effect.

**Card side** (`cards/effects/control.go`):

- `GainControlOfSpell{Spell, Controller, ChooseNewTargets}`. `Controller` defaults to the resolving item's controller, as `GainControl` does.
- `ExchangeControlOfSpellAnd{Spell, Permanent, ChooseNewTargets}`.

Neither takes a duration. Every printed card is indefinite, and on the stack "indefinite" means "until it leaves".

**Abilities are refused.** No printed card gains control of an ability. The writer is for spells, so `DefaultController` and `ControlChanges` are meaningless on an ability item.

### 2. What "you" means at resolution

The spell's current controller, because `ctx.Controller()` reads `item.Controller` and that is now the materialised answer (CR 608.2c, 109.5). A stolen Lightning Bolt deals its damage as the thief's. A stolen Divination draws for the thief. A stolen Wrath's "you" is the thief.

Three things follow and need no code:

- **Targets are re-checked for the new controller** (CR 608.2b). A stolen Murder aimed at "target creature an opponent controls" is illegal at resolution if the creature now belongs to the thief and the thief kept the target. The spell does nothing. That is the reason the cards say "you may choose new targets".
- **"Each opponent" is the thief's opponents,** because it is computed from the controller at resolution.
- **"Its controller" in someone else's spell is the thief.** Mana Leak on a stolen spell asks the thief to pay.

### 3. Who controls the permanent a stolen permanent spell becomes

The permanent **enters under the spell's controller** (the thief). The entry event keeps `Actor: item.Controller`, so the would-be controller of ADR 0102, the self-replacements and the ETB triggers all see the thief (CR 110.2b's first half, CR 603.3a). In addition, as the permanent lands:

- **`Card.BaseController` is stamped to `DefaultController`,** not captured lazily from `Controller`. That is CR 110.2b's second half: the permanent's default controller is the player who put the spell on the stack.
- **Each `SpellControlChange` becomes an indefinite `setController` `ScopedEffect` pinned to the new permanent,** with the change's **original timestamp** (CR 400.7a, 613.7b). This uses `registerScopedEffectLocked`, which takes an explicit timestamp.

The first layer pass reseeds control from `BaseController` (the caster), applies the records, and lands on the thief. `Card.Controller` already equals that, so **no control-change event fires on entry**, and no "whenever an opponent gains control" trigger fires by accident.

The permanent then behaves like any stolen permanent. An Act of Treason later sorts against it by timestamp. If the thief leaves the game, the record ends and the permanent goes back to the caster (§7).

A stolen permanent spell whose own replacement changes the entering controller (ADR 0102's Captive Audience, stolen) enters under the player that replacement names. The carried record still applies on top in layer 2, so the thief controls it. `BaseController` is then the replacement's player, because that is who it entered under (CR 110.2). This is a corner, not a question. The implementation PR pins it with a test.

### 4. Choosing new targets

"You may choose new targets for it" is the existing retarget offer. It is `ChangeTargets{StackID, Policy: RetargetChooseNew, Optional: true}`, asked of the **new** controller, **after** the control change. The order matters and is enforced by the card-side primitive (`ChooseNewTargets: true`), not left to each card:

- The legal set is built by `stackItemSourceLocked`, which reads `item.Controller`. So it has to be the thief's by then, or "target opponent" and hexproof are judged for the wrong player.
- CR 115.7d/e (keep any, change to legal ones only, judge the final set) is already `retargetCheckLocked`'s job.

Sudden Substitution's "then the spell's controller may choose new targets" is the same call. The chooser is `item.Controller` after the exchange, which is how the card reads.

`PendingChoiceRetarget` is data, not a closure. A table paused on it is still a restore point, and its CR 800.4 row ("dropped, targets unchanged") already fits: if the thief leaves before answering, the targets stay.

### 5. Invert Polarity's coin flip

The card is one flip whose result chooses the branch:

```go
Targets: TargetSpell("target spell"),
OnResolve: func(item *game.StackItem, ctx *Context) error {
	spell, me, src := item.Targets[0].ID, ctx.Controller(), ctx.Source()
	ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper: me, Source: src, Coins: 1,
		Question: "Invert Polarity — call the flip",
		Then: func(g *game.Game, r game.CoinFlipResult) error {
			if len(r.Won) == 1 && r.Won[0] {
				return g.GainControlOfSpellThenRetarget(src, spell, me, true) // §1 + §4
			}
			return g.CounterSpellForEffect(src, spell) // honours "can't be countered"
		},
	})
	return nil
},
```

(The two `g.` calls in the sketch are placeholders for the §1 writer with §4's offer, and the existing counter door. The PR names them for real.)

- **The spell is a target.** If it has left by resolution, Invert Polarity does nothing and nobody flips (CR 608.2b). While the call is open, no other move is legal, so the spell cannot leave in between.
- **The player who flips calls the coin** (CR 705.2). That is the existing `coin_call` prompt, which the client and the bot already answer.
- **Losing the flip is an ordinary counter.** A Cavern-protected spell is neither countered nor stolen.
- **The continuation captures only IDs**, as ADR 0054 requires, so an undo resolves it against the restored game.
- **Restore points:** `CoinFlipSpec.Then` is a closure, counted as `census:ChoiceResumeFrames`. A table paused on the call is not a restore point, which is true of every won/lost flip today. Once the call is answered, the stolen spell (and any retarget prompt) is plain data again.

### 6. Cast permissions, copies, cast records and mana riders

This is the carried question: how should a change of a spell's controller carry through permissions, copies and cast records? One principle answers every row. **What happened while the spell was being cast stays with the cast. What happens as it resolves or later belongs to the controller at that time, unless the rule names the owner.**

| Record | After the steal | Why |
|---|---|---|
| `EventCast`, `CastTally`, storm's count, "spells cast this turn", `WheneverYouCast` | Unchanged. The caster cast it and the thief did not. | CR 601.2a. The cast is a past event. |
| `StackItem.Paid` (mana, X, kicker, sacrificed or exiled objects, gift's promised opponent) | Carried untouched. | These are decisions and costs of the cast. CR 707.10's "decisions made for it" says the same about copies. |
| Mana spend riders (ADR 0040) | Unchanged. They fired at the spend, and their `Applied` stamps ride `Paid.Mana`. Cavern's "can't be countered" still protects the stolen spell. Hall of the Bandit Lord's haste still reaches the permanent. A Goggles trigger was queued for the caster before the steal. | The rider is a fact about the mana, not about who holds the spell. |
| `CastProvenance` on the permanent | Carried, plus a new field, `Caster`. | "If you cast it" (The One Ring, Zacama, Tiamat) must be false for a thief. Today `b16EnteredFromStack` answers "it came from the stack", which a stolen spell also did. The readers become `Provenance.Caster == controller`. That also stops a token from a copied permanent spell (never cast) from passing. |
| Permissions that let it be cast | Spent. Any rider the permission put on the spell (`ExileOnLeavingStack`, `ExileOnResolution`) rides the item and still applies. | CR 400.7g/h. The permission's job ended at the cast. |
| Permissions **created at resolution** | Adventure: the **controller** at resolution (CR 715.3d; finding 3). Warp's later cast: the **owner** (CR 702.185a). | The rule names who. |
| Where the card goes | Buyback: the owner's hand. Flashback: exile. Dash: the owner's hand. A plain instant: the owner's graveyard. | CR 702.27a, 702.34a, 702.109a, 608.2n. None of them changes. |
| Evoke's sacrifice, gift's "gives a gift" | The controller: the thief sacrifices the evoked creature and is the player who gives the gift. | CR 702.74a, 702.174c. |
| Copies of a stolen spell | Controlled and owned by whoever makes the copy. A storm trigger that resolves after the steal still makes the caster's copies, because the trigger is the caster's. | CR 707.10. `CopySpellForEffect` already takes the copier as `controller`. |
| A stolen copy | Same record as a card. Its owner never changes. If it is countered it ceases to exist, as now. | CR 707.10, 707.10a. |

### 7. Countering, and CR 800.4a

**Countering** needs nothing new. A countered stolen spell goes to its **owner's** graveyard (CR 701.6a), because `routeStackCardToGraveyardLocked` already routes by owner. "Counter target spell you don't control" reads the stack card's `Card.Controller`, which §1 keeps in step. "Counter unless its controller pays" asks the thief.

**A player leaving the game** is the one place the record earns its keep. `cleanupStackForEliminatedLocked` gains a step before its sweep, following CR 800.4a's order.

- **The thief leaves.** Their `SpellControlChange` entries end (step 2), and the spell re-materialises to the previous thief or to the default controller. Only then does the sweep look for items the departed player still controls. A copy they still control ceases to exist (step 3). A card they still control is exiled (step 4). That happens only if the spell reverts to nobody who is still in the game. That is also the CR 800.4c case.
- **The caster leaves.** A spell they own leaves the game with them (step 1, already done by `removeObjectsOwnedByLocked`). A spell they cast but do not own (cast off another player's card) stays under the thief.
- **A stolen permanent spell that has resolved.** Its control is now a `ScopedEffect`. **PR 1 ends every `setController` record that names a departed player** in `leaveGameObjectsLocked` step 2. That is finding 2, and it fixes Act of Treason at the same time. The permanent reverts to its `BaseController`. If that player has left too, the existing `exileGhostControlledLocked` (CR 800.4c) exiles it.
- **A steal aimed at a departed player** does nothing (CR 800.4b). `GainControlOfSpellForEffect` refuses it.

### 8. The wire, the client and the bot

**Event.** A new `EventSpellControlChanged{CardID: spell, Actor: gained, Target: lost, Source}`. It is not `EventControlChanged`, on purpose. `AnOpponentGainedControlOfAPermanentYouOwn` looks the card up with `LookupCardForEffect`, which finds a stack card, and would fire on a stolen spell. That is "a permanent you own" misfiring on a spell. A separate kind makes the misfire impossible rather than needing a zone check in every predicate. The log gate needs an arm: "Bob gained control of Alice's Lightning Bolt."

**Wire.** `StackItemView.controller` stays the live controller. It gains `default_controller`, sent only when it differs from `controller`. The stack card's `CardView.controller` follows §1's re-stamp. There is no new verb and no new `PendingChoiceKind`. The coin call, the "you may" prompt and the retarget prompt already exist, so the choice-gate and departure tables are unchanged.

**Client.** `stackLane.ts` labels `item.controller` as the "caster" (`casterSeat`, `casterName`). The lane keeps the controller's colour and adds a chip, "taken from Alice", when `default_controller` is present (open question 6). Target phrases already read the stack item's controller, so "targets your Lightning Bolt" follows the steal.

**Bot.** The enumerator (`internal/legal`) needs nothing new. "Target spell" is an existing clause, the `may` prompt and the coin call are existing answers, and the retarget prompt is enumerated since #1196. The policy reads `controller` from the filtered view, which is live, so a bot does not counter a spell it now controls. Three small changes:

- The model prompt prints "cast by <controller>" (`aiseat/model/prompt.go`, `improvise.go`). It becomes "controlled by X" plus "(cast by Y)" when `default_controller` is present.
- The heuristic scores "gain control of target spell" like a counter of an opponent's spell, plus the spell's value to the thief, and never targets its own spell.
- Perplexing Chimera's "you may exchange" is answered yes when the spell's mana value is at least the Chimera's (5) or the spell is a permanent spell. This is a starting heuristic. The arena report is what it gets judged by.

### 9. Snapshots

Everything is data, and there is no new closure field:

- **`StackItem.DefaultController` and `StackItem.ControlChanges`** are additive. The shape file is updated with `-update-shape`. Since #1497, a binary refuses a stack-item field it does not know, so a rollback refuses a file carrying a steal. That is the designed rollback case, not a bump.
- **`CastProvenance.Caster`** is an additive `Card` field, recorded the same way.
- **The carried permanent control** is an ordinary `setController` `ScopedEffect`.
- **The closure ratchet** gains no route.
- **The corpus** gets one new board in the current `v<N>/`: a stolen spell on the stack with a retarget prompt open. The writer allows a new board without a bump.

The one pause that is not a restore point is Invert Polarity's open coin call (§5), and that is already true of every called flip.

### 10. Cards covered, and how it ships

**PR 1 (engine).**

- The §1 fields, writer, materialiser and exchange.
- The stack-card `Controller` stamp at cast.
- §3's hand-off at resolution.
- §7's CR 800.4a step, and ending departed `setController` records.
- The event and its log arm, the wire field, the client chip and the bot prompt text.
- The card-side `GainControlOfSpell` and `ExchangeControlOfSpellAnd`.
- `CastProvenance.Caster`, with the "if you cast it" readers moved to it.
- The Adventure permission holder (question 4).
- Tests: engine tests with fixture spells, and a test that pins Act of Treason's reversion first.

**PR 2 (cards).**

- Invert Polarity, Aethersnatch and Perplexing Chimera.
- Sudden Substitution.
- Commandeer, with `ExileFromHand` taking a count (the escape-style `Min`).
- One oracle-text fixture per card.
- The registry's `control-of-a-spell` seam flipped to implemented, and a closed-seam fragment.

**Waiting:** Chef's Kiss, unless the owner puts it in scope (question 2).

That is **five of the six cards**, with the sixth waiting on one decision.

---

## Consequences

- `StackItem.Controller` is no longer "the player who cast it". Code that needs the caster reads `DefaultController` (zero means `Controller`) or, for a permanent, `Provenance.Caster`. The field comment changes to say so.
- A spell's control has two homes: the stack record while it is a spell, and a `ScopedEffect` once it is a permanent. They meet at one function, at resolution.
- The stack card's `Card.Controller` is kept equal to the item's. That fixes "spell you don't control" for spells cast off another player's card, which is a behaviour change today.
- Act of Treason and every other resolving-spell theft now revert when the thief leaves the game, instead of exiling the permanent. This is a behaviour change, and it is the one CR 800.4a's example describes.
- The Adventure permission follows CR 715.3d's controller if question 4 is answered that way. That is a behaviour change for Adventures cast off another player's card.
- `docs/engine-seams.md` and the roadmap registry stop saying control of a spell "is not layer 2 at all". It is layer 2 in the rules, and the engine answers it in two places.

## Out of scope

- **Control of an ability** on the stack. No printed card does it.
- **Control of a player's turn** (Mindslaver). It is a different rule (CR 723) and a different seam.
- **A random retarget** ("reselect the targets at random"), unless question 2 brings Chef's Kiss in.
- **Desertion** ("put that card onto the battlefield under your control instead"). It is a counter-redirect and already expressible (ADR 0102, "Out of scope").

---

## Open questions for the owner

1. **The shape.** Should control of a spell be a record on the stack item, handed to layer 2 when the spell resolves (option C, recommended)? The alternatives are to extend the layer pass to the stack (option B: most faithful, largest), or to rewrite `StackItem.Controller` with no memory (option A: smallest, and it gets CR 110.2b and CR 800.4a wrong).
2. **Chef's Kiss.** Should PR 2 include it, which needs a random retarget that excludes "you or a permanent you control"? Or should it wait on a new seam row (recommended: the other five do not need it, and a random retarget is its own small design)?
3. **The two latent bugs.** Should PR 1 fix the stale stack-card `Controller` for spells cast off another player's card, and end `setController` records that name a departed player (the Act of Treason reversion)? Both are needed by this design, and the recommendation is to fix both in PR 1, each with its own test first. The alternative is to file them as separate issues and land them ahead of PR 1.
4. **The Adventure permission.** CR 715.3d says the spell's **controller** may cast the card from exile later. The engine grants it to the **owner** and cites the older rule number. Should PR 1 follow CR 715.3d (recommended)? That also changes Adventures cast off another player's card today.
5. **The event.** Should a stolen spell emit its own `EventSpellControlChanged` (recommended, so no "gains control of a permanent" trigger can see a spell)? Or should it reuse `EventControlChanged`, with a zone check added to every predicate that reads it?
6. **What the table sees.** Should the stack lane show a "taken from Alice" chip while a stolen spell is on the stack (recommended)? Should the permanent it becomes show one too, the way a stolen creature does not today?
7. **A stolen face-down spell** (a morph cast face down, taken by Aethersnatch). Should the thief become able to look at it while the caster keeps knowing it (recommended: knowledge cannot be taken back)? Or should the engine follow CR 708.5 literally and remove the caster's look?
8. **The bot and Perplexing Chimera.** Is "exchange when the spell's mana value is at least 5 or it is a permanent spell" the right starting rule? Or should the bot decline until the arena has measured the card?
