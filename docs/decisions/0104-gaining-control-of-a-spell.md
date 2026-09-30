# ADR 0104 — Gaining control of a spell on the stack

**Status:** Accepted · 2026-09-30 · S50 — Seams from the deck re-checks
**Owner decisions:** 2026-09-30. All eight open questions are answered; see [Owner decisions](#owner-decisions-2026-09-30) at the end. The owner chose option B, extending layer 2 to the stack.
**Issue:** [#1745](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1745). Invert Polarity is waiting on it in the ViviVoltron deck request, [#1640](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1640).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-09-30. I ran `git fetch --all --prune` and read every `docs/decisions/` file name on all remote heads. Numbers 0103 and 0105 are reserved for ADRs being written in parallel, and 0105 already exists on a branch. No branch has 0104, so this one takes **0104**.
**Builds on:** [ADR 0063](0063-durations-and-control.md) (layer-2 control from a spell or ability), [ADR 0102](0102-entering-under-another-players-control.md) (the would-be controller of an entering permanent), [ADR 0060](0060-leaving-the-game.md) (CR 800.4), [ADR 0019](0019-structured-targeting.md) and its 2026-09-22 amendment (choosing new targets, #1196), [ADR 0054](0054-dice-rolls-and-coin-flips.md) (the won/lost coin flip), [ADR 0043](0043-copy-effects.md) (copying a spell), [ADR 0040](0040-mana-pipeline.md) (mana spent and spend riders), [ADR 0066](0066-granted-cast-and-play-permissions.md) (cast permissions), [ADR 0033](0033-ai-bot-seat.md) (the bot) and [ADR 0041](0041-game-persistence.md) (restore points).

This ADR was written plan-first. The engine and card changes land in separate PRs that follow it.

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

### B. Extend layer 2 to the stack (chosen)

Make the steal a `ScopedEffect` with a `setController` mod pinned to the stack object. Give the layer pass a stack step that reads those records and writes `StackItem.Controller` from the result. Re-pin the record to the permanent at resolution.

- **For:** it is the rule as written. Control of a spell is a continuous effect (CR 611.1) applied in layer 2 (CR 613.1b), and CR 400.7a carries it onto the permanent. One registry, one duration model and one timestamp sort cover spells and permanents alike. The reversions the rules ask for — a thief leaving the game, a second steal, a permanent's default controller — all fall out of machinery that already answers them for permanents.
- **Against:** it is the largest of the three. The affected-set pin, the duration pin and the layer pass all learn about a stack object. Nothing else in the catalog wants a stack object in layer 2 yet, so for now the stack step applies layer 2 only.

### C. A control record on the stack item, handed to layer 2 at resolution

Keep a list of steals on the `StackItem`, turn it into `StackItem.Controller` with a small function of its own, and hand it to layer 2 only when the spell becomes a permanent.

- **For:** smaller than B, and every reader stays as it is.
- **Against:** it is a second answer to "who controls this object", written beside the layer engine rather than in it. Every rule the layer engine already applies to control — timestamps, durations, the end of a departed player's effects — would have to be restated for the stack's copy.

---

## Decision: option B

The owner chose B on 2026-09-30: "I am always going to favor being stringent to the rules." A and C stay above as the alternatives considered.

### 1. How control of a stack object is represented

**The steal is a `ScopedEffect`.** It has one `setController` mod, an `Indefinite` duration (every printed card states none, CR 611.2a) and a CR 613.7b timestamp taken when it is created. The record is data, like every other `ScopedEffect`.

**It is pinned to the stack object.** Two pins learn the stack, each with two additive fields:

- `AffectedObject.OnStack` and `AffectedObject.Epoch` name the spell by instance ID and by the `Card.ObjectEpoch` it has on the stack. `MoveCard` bumps the epoch on every zone change, so CR 400.7 is checked, not assumed: the same card cast again is a new object the record does not match.
- `Duration.PinnedOnStack` and `Duration.PinnedEpoch` are the same pin on the duration. `durationExpiredLocked` ends a stack-pinned record once that object is no longer on the stack, so a stolen spell that is countered or resolves as an instant leaves no record behind.

**Only `setController` is accepted on a stack pin.** Registration panics on any other mod kind with a stack pin, as it does on every malformed record. A future card that changes a spell's other characteristics this way extends the stack step to its layer in the same change.

**The default controller.** `StackItem.BaseController` is the player who put the item on the stack: the caster, or for a copy the player who made it (CR 110.2b, CR 707.10). It is captured lazily, exactly as `Card.BaseController` is: zero means "the current controller is the base", and the stack step stamps it before it applies the first record.

**The stack step of the layer pass.** `recomputeLayersLocked` runs `stackControlPassLocked` after the battlefield pass. For each spell on the stack it:

1. reseeds the controller from `BaseController`;
2. applies every live `setController` record pinned to that object, in CR 613.7 timestamp order, the same sort the battlefield bucket uses;
3. materialises the answer into `StackItem.Controller` **and** into the spell card's `Card.Controller` in the stack zone;
4. collects each delta by value, to be emitted after the store as `EventSpellControlChanged` (#930's pattern).

Because the step materialises into the field every reader already reads, about 660 card files and the engine's readers do not change.

**The writers.** In the engine:

- `GainControlOfSpellForEffect(sourceID, spellID, player, label) bool`. It refuses anything but a spell on the stack, a departed player (CR 800.4b) and a no-op (the player already controls it). Then it registers the record and recomputes, so the controller has changed by the time it returns.
- `ExchangeControlOfSpellAndPermanentForEffect(sourceID, spellID, permanentID, label) bool`. It checks both objects first and does nothing if either is gone or one player controls both (CR 701.12a/b). Then it registers the spell's record and the permanent's record with **one shared timestamp**, as `ExchangeControlForEffect` does for two permanents.

On the card side (`cards/effects/control.go`): `GainControlOfSpell{Spell, Controller, ChooseNewTargets}` and `ExchangeControlOfSpellAnd{Spell, Permanent, ChooseNewTargets}`. `Controller` defaults to the resolving item's controller.

**The stack card's controller.** `CastSpell` stamps the spell card's `Card.Controller` from the caster for every cast, not just a face-down one. That closes the first finding: "target spell you don't control" is judged by the spell's controller for a spell cast off another player's card.

**Abilities are not covered.** No printed card gains control of an ability, and the writer refuses one.

### 2. What "you" means at resolution

The spell's current controller, because `ctx.Controller()` reads `item.Controller` and the stack step keeps it current (CR 608.2c, 109.5). A stolen Lightning Bolt deals its damage as the thief's. A stolen Divination draws for the thief. A stolen Wrath's "you" is the thief.

Three things follow and need no code:

- **Targets are re-checked for the new controller** (CR 608.2b). A stolen Murder aimed at "target creature an opponent controls" is illegal at resolution if the creature now belongs to the thief and the thief kept the target. The spell does nothing. That is the reason the cards say "you may choose new targets".
- **"Each opponent" is the thief's opponents,** because it is computed from the controller at resolution.
- **"Its controller" in someone else's spell is the thief.** Mana Leak on a stolen spell asks the thief to pay.

### 3. Who controls the permanent a stolen permanent spell becomes

The permanent **enters under the spell's controller** (the thief). The entry event keeps `Actor: item.Controller`, so the would-be controller of ADR 0102, the self-replacements and the ETB triggers all see the thief (CR 110.2b's first half, CR 603.3a).

As the permanent lands (`landEntryLocked`, before any event goes out):

- **`Card.BaseController` is stamped with the stack item's default controller.** That is CR 110.2b's second half: the permanent's controller by default is the player who put the spell on the stack. When an ADR 0102 entry-controller replacement changed whom the permanent enters under, the default is that player instead (CR 110.2).
- **Each record pinned to the stack object is re-pinned to the permanent** (CR 400.7a). The record is replaced, never edited: the same mods, timestamp and sequence number, with the pin moved to `PinObject(permanent, entry stamp)` and the duration moved to `PinnedTo(permanent)`. A copy of a permanent spell becomes a token with a new instance ID. The re-pin follows whatever ID actually entered.

The next pass reseeds from `BaseController`, applies the record and lands on the thief. `Card.Controller` already equals that, so **no control-change event fires on entry**, and no "whenever an opponent gains control" trigger fires by accident. From then on the permanent is an ordinary stolen permanent: a later Act of Treason sorts against the record by timestamp, and a thief leaving the game hands it back (§7).

### 4. Choosing new targets

"You may choose new targets for it" is the existing retarget offer: `ChangeTargets{StackID, Policy: RetargetChooseNew, Optional: true}`. It is asked of the **new** controller, **after** the control change. The card-side primitive enforces the order (`ChooseNewTargets: true`), so no card has to get it right:

- The legal set is built by `stackItemSourceLocked`, which reads `item.Controller`. It has to be the thief's by then, or "target opponent" and hexproof are judged for the wrong player. The writer's recompute is what makes it so.
- CR 115.7d/e (keep any, change to legal ones only, judge the final set) is already `retargetCheckLocked`'s job.

Sudden Substitution's "then the spell's controller may choose new targets" is the same call. The chooser is the spell's controller after the exchange.

`PendingChoiceRetarget` is data, not a closure. A table paused on it is still a restore point. Its CR 800.4 row ("dropped, targets unchanged") already fits: if the thief leaves before answering, the targets stay.

### 5. Invert Polarity's coin flip

The card is one called flip whose result chooses the branch. `FlipCoinForEffect` queues the `coin_call` prompt. Its continuation, which captures only IDs, takes one of two branches:

- **Won:** `GainControlOfSpell{ChooseNewTargets: true}`.
- **Lost:** the ordinary counter.

Rules and consequences:

- **The spell is a target.** If it has left by resolution, Invert Polarity does nothing and nobody flips (CR 608.2b). While the call is open no other move is legal, so the spell cannot leave in between.
- **The player who flips calls the coin** (CR 705.2). The client and the bot already answer `coin_call`.
- **Losing the flip is an ordinary counter.** A Cavern-protected spell is neither countered nor stolen.
- **Restore points.** `CoinFlipSpec.Then` is a closure (`census:ChoiceResumeFrames`), so a table paused on the call is not a restore point. That is true of every called flip today. Once the call is answered, the stolen spell and any retarget prompt are plain data again.

### 6. Cast permissions, copies, cast records and mana riders

One principle answers every row. **What happened while the spell was cast stays with the cast. What happens as it resolves, or later, belongs to the controller at that time, unless the rule names the owner.**

| Record | After the steal | Why |
|---|---|---|
| `EventCast`, `CastTally`, storm's count, "spells cast this turn", `WheneverYouCast` | Unchanged. The caster cast it, the thief did not. | CR 601.2a. The cast is a past event. |
| `StackItem.Paid` (mana, X, kicker, sacrificed or exiled objects, gift's promised opponent) | Carried untouched. | These are decisions and costs of the cast. CR 707.10's "decisions made for it" says the same about copies. |
| Mana spend riders (ADR 0040) | Unchanged. They fired at the spend, and their `Applied` stamps ride `Paid.Mana`. Cavern's "can't be countered" still protects the stolen spell. Hall of the Bandit Lord's haste still reaches the permanent. A Goggles trigger was already queued for the caster. | The rider is a fact about the mana, not about who holds the spell. |
| `CastProvenance` on the permanent | Carried, plus a new field, `Caster`: the player who cast the spell, empty for a copy. | "If you cast it" must be false for a thief. The readers become `Provenance.Caster == controller`. |
| Permissions that let it be cast | Spent. A rider the permission put on the spell (`ExileOnLeavingStack`, `ExileOnResolution`) rides the item and still applies. | CR 400.7g/h. |
| Permissions **created at resolution** | Adventure: the **controller** at resolution (CR 715.3d). Warp's later cast: the **owner** (CR 702.185a). | The rule names who. |
| Where the card goes | Buyback: the owner's hand. Flashback: exile. Dash: the owner's hand. A plain instant: the owner's graveyard. | CR 702.27a, 702.34a, 702.109a, 608.2n. |
| Evoke's sacrifice, gift's "gives a gift" | The controller: the thief sacrifices the evoked creature and is the player who gives the gift. | CR 702.74a, 702.174c. |
| Copies of a stolen spell | Controlled and owned by whoever makes the copy. A storm trigger that resolves after the steal still makes the caster's copies, because the trigger is the caster's. | CR 707.10. |
| A stolen copy | The same record as a card. Its owner never changes. Countered, it ceases to exist, as now. | CR 707.10, 707.10a. |
| A stolen face-down spell | Only the new controller may look. The caster loses the look. | CR 708.5, applied literally (owner decision 7). |

### 7. Countering, and CR 800.4a

**Countering** needs nothing new. A countered stolen spell goes to its **owner's** graveyard (CR 701.6a), because `routeStackCardToGraveyardLocked` already routes by owner. "Counter target spell you don't control" reads the spell card's `Card.Controller`, which the stack step keeps in step. "Counter unless its controller pays" asks the thief.

**A player leaving the game** follows CR 800.4a's order. `cleanupStackForEliminatedLocked` gains a first step, `endControlEffectsForLocked`. It drops every `ScopedEffect` record that gives the departed player control, whether the record is pinned to a spell or a permanent, and recomputes. Then the existing sweep runs.

- **The thief leaves.** Their records end (step 2), and the spell falls back to the previous thief or to its default controller. Only then does the sweep look for items the departed player still controls. A copy they still control ceases to exist (step 3), and a card they still control is exiled (step 4).
- **The spell falls back to a player who has already left** (CR 800.4c). The sweep treats a spell controlled by *any* departed player as unowned by anyone still in the game, and exiles it.
- **The caster leaves.** A spell they own leaves the game with them (step 1, already done by `removeObjectsOwnedByLocked`). A spell they cast but do not own stays under the thief.
- **A stolen permanent spell that has resolved.** Its record is an ordinary battlefield `ScopedEffect` now, and the same `endControlEffectsForLocked` ends it. It goes back to its `BaseController`, or is exiled by the existing `exileGhostControlledLocked` when that player has left too. This is the second finding: it fixes Act of Treason's reversion at the same time.
- **A steal aimed at a departed player** does nothing (CR 800.4b). The writer refuses it.

### 8. The wire, the client and the bot

**Event.** A new `EventSpellControlChanged{CardID: spell, Actor: gained, Target: lost, Source}`. It is deliberately not `EventControlChanged`. `AnOpponentGainedControlOfAPermanentYouOwn` looks the card up with `LookupCardForEffect`, which finds a stack card, and would fire on a stolen spell. The log gate gets an arm: "Bob gained control of Alice's Lightning Bolt". A face-down spell's name is redacted for anyone who may not look.

**Wire.** `StackItemView.controller` stays the live controller, and gains `default_controller`, sent only when it differs. The spell card's `CardView.controller` follows the stack step. There is no new verb and no new `PendingChoiceKind`: the coin call, the "you may" prompt and the retarget prompt already exist.

**Client.** A "taken from X" chip in two places:

- **The stack lane**, while a stolen spell is on the stack. X is the default controller.
- **Any permanent controlled by a player who does not own it.** X is the owner. This covers a permanent from a stolen spell and every other stolen permanent alike, and it is derived from the `controller` and `owner` the wire already carries.

**Bot.** The enumerator needs nothing new: "target spell" is an existing clause, and the `may` prompt, the coin call and the retarget prompt are existing answers. The policy reads `controller` from the filtered view, which is live, so a bot does not counter a spell it now controls. The rest:

- **Model prompt.** It prints "cast by <controller>" today. It becomes "controlled by X" plus "(cast by Y)" when `default_controller` is present.
- **Perplexing Chimera.** The bot exchanges when the spell's mana value is 5 or more, or it is a permanent spell (owner decision 8).

### 9. Snapshots

Everything is data, and no new closure route is added. The new fields are all additive:

- `AffectedObject.OnStack` and `AffectedObject.Epoch`;
- `Duration.PinnedOnStack` and `Duration.PinnedEpoch`;
- `StackItem.BaseController`;
- `CastProvenance.Caster`.

The shape file is updated with `-update-shape` under the current schema version. Older binaries refuse an unknown affected-set field (ADR 0041 P4) and an unknown stack-item field (#1497), so a rollback refuses a file carrying a steal. That is the designed rollback case, not a bump.

The one pause that is not a restore point is Invert Polarity's open coin call (§5). That is already true of every called flip.

### 10. Cards covered, and how it ships

**Engine PR:**

- the stack pins, the stack step and the two writers;
- the stack card's `Controller` stamp;
- the re-pin at resolution;
- `endControlEffectsForLocked` and the CR 800.4c sweep;
- the event, its log arm, the wire field, both chips and the bot's prompt text;
- `CastProvenance.Caster` and the "if you cast it" readers;
- the Adventure permission (CR 715.3d);
- the literal CR 708.5 knowledge change;
- the card-side primitives.

Each rule gets a test first, including the two latent bugs.

**Cards PR:** Invert Polarity, Aethersnatch, Perplexing Chimera and Sudden Substitution. Commandeer joins them if giving `ExileFromHand` a count is small and clean; otherwise it waits on its own seam row.

**Waiting:** Chef's Kiss, on a new "random retarget" seam row.

---

## Consequences

- `StackItem.Controller` is no longer "the player who cast it". It is layer 2's answer for a spell. Code that needs the caster reads `BaseController` (zero means `Controller`) or, for a permanent, `Provenance.Caster`.
- The layer pass writes to the stack as well as the battlefield. The stack step is layer 2 only, and it says so.
- The stack card's `Card.Controller` is kept equal to the item's. That fixes "spell you don't control" for spells cast off another player's card, a behaviour change today.
- Act of Treason and every other resolving-spell theft now revert when the thief leaves the game, instead of exiling the permanent. This is the behaviour CR 800.4a's example describes.
- The Adventure permission follows CR 715.3d's controller. That is a behaviour change for Adventures cast off another player's card.
- A face-down spell stops being visible to its caster once it is stolen (CR 708.5).
- `docs/engine-seams.md` and the roadmap registry stop saying control of a spell "is not layer 2 at all".

## Out of scope

- **Control of an ability** on the stack. No printed card does it.
- **Control of a player's turn** (Mindslaver). It is a different rule (CR 723) and a different seam.
- **A random retarget** ("reselect the targets at random"). Chef's Kiss waits on it.
- **Desertion** ("put that card onto the battlefield under your control instead"). It is a counter-redirect and already expressible (ADR 0102, "Out of scope").
- **Other layers on the stack.** No catalogued effect changes a spell's characteristics through a `ScopedEffect` yet.

---

## Open questions for the owner

These are the questions as asked. The owner's answers follow.

1. **The shape.** Should control of a spell be a record on the stack item, handed to layer 2 when the spell resolves (option C, recommended)? The alternatives are to extend the layer pass to the stack (option B: most faithful, largest), or to rewrite `StackItem.Controller` with no memory (option A: smallest, and it gets CR 110.2b and CR 800.4a wrong).
2. **Chef's Kiss.** Should the cards PR include it, which needs a random retarget that excludes "you or a permanent you control"? Or should it wait on a new seam row (recommended)?
3. **The two latent bugs.** Should the engine PR fix the stale stack-card `Controller` for spells cast off another player's card, and end `setController` records that name a departed player (the Act of Treason reversion), each with its own test first (recommended)? Or should they be separate issues that land first?
4. **The Adventure permission.** CR 715.3d says the spell's **controller** may cast the card from exile later. The engine grants it to the **owner** and cites the older rule number. Should the engine PR follow CR 715.3d (recommended)?
5. **The event.** Should a stolen spell emit its own `EventSpellControlChanged` (recommended)? Or should it reuse `EventControlChanged`, with a zone check added to every predicate that reads it?
6. **What the table sees.** Should the stack lane show a "taken from Alice" chip while a stolen spell is on the stack (recommended)? Should the permanent it becomes show one too?
7. **A stolen face-down spell.** Should the thief become able to look at it while the caster keeps knowing it (recommended)? Or should the engine follow CR 708.5 literally and remove the caster's look?
8. **The bot and Perplexing Chimera.** Is "exchange when the spell's mana value is at least 5 or it is a permanent spell" the right starting rule?

## Owner decisions (2026-09-30)

1. **Option B: extend the layer pass to the stack.** "I am always going to favor being stringent to the rules." C, which the draft recommended, is kept above with A as the alternatives considered.
2. **Chef's Kiss waits** on its own seam row, "random retarget", with a reason on the registry's `Waiting` list.
3. **Both latent bugs are fixed in the engine PR, each test-first.** These are the stale stack-card `Controller` for spells cast off another player's card, and ending `setController` records that name a departed player (Act of Treason's reversion, CR 800.4a).
4. **CR 715.3d.** The Adventure's later cast goes to the spell's controller.
5. **A new `EventSpellControlChanged`.**
6. **The chip appears on the stack lane and on the permanent.** The permanent's chip is general: any permanent controlled by a player who does not own it shows it.
7. **CR 708.5 literally.** A stolen face-down spell can be looked at by its new controller only. The caster loses the look.
8. **The bot exchanges with Perplexing Chimera** when the spell's mana value is 5 or more, or it is a permanent spell.
