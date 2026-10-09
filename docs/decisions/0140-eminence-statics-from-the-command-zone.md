# ADR 0140 — Eminence: a static that works from the command zone

**Status:** Accepted · 2026-10-09 · Reality Fracture set (tracker [#2795](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2795); outside docs/sprints.md at the owner's request)
**Issue:** [#2797](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2797), piece 5 of the five planeswalker pieces. The other four are dated amendments to the ADRs that own them (below), not new decisions.
**Numbering:** the AGENTS.md §4 sweep on 2026-10-09 (`git log --all --remotes -- docs/decisions`) found 0139 (`0139-empower-jace.md`, the #2796 builder's) as the highest number on any branch. This ADR takes **0140**, reserved for it in the #2797 claim comment.
**Builds on:** [ADR 0048](0048-cost-modification.md) and its addendum (the CR 601.2f cost pass and its one gatherer, `activeCostModifiersLocked`), [ADR 0093](0093-abilities-granted-to-other-permanents.md) (`CatalogAbilityKey`, the removal-aware key every static is read through), and [ADR 0115](0115-commanders-die.md) (the command zone as a place a card is, not just a place it is sent).

---

## Context

The Ur-Sphinx prints an **eminence** ability:

> Eminence — As long as The Ur-Sphinx is in the command zone or on the battlefield, other Sphinx spells you cast cost {1} less to cast.

Eminence is not a rules term. It is an ability word, so it has no rules meaning of its own; the rest of the sentence is what matters, and it is a **static ability that functions in the command zone**. CR 113.6 says an object's abilities function only on the battlefield unless the ability says otherwise, and an ability that names the command zone is such a statement. Every static in this engine is read from the battlefield (or, for a spell's own cost, from the stack): `activeCostModifiersLocked` walks `g.Battlefield.Cards`, and nothing walks a command zone. A grep of the tree for `eminence` finds nothing, so no card has this and no seam names it.

About twenty Commander-legal cards print an eminence ability (Edgar Markov, Inalla, Arahbo, The Ur-Sphinx, Derevi's relatives and the Slivers' commanders among them). They fall into three shapes: a **cost reduction** (this card), a **triggered ability** that watches casts or ETBs from the command zone ("whenever you cast another Vampire spell, if Edgar Markov is in the command zone or on the battlefield, …"), and a **replacement or damage static** (Arahbo's anthem, Inalla's copy). Only the first is needed now.

## Decision

### 1. A cost modifier says it works from the command zone

`game.CostModifier` gains one field, `Eminence bool`. A modifier with it set is gathered from a card in its owner's command zone as well as from a permanent on the battlefield. `activeCostModifiersLocked` walks the battlefield exactly as before and then calls `appendEminenceCostModifiersLocked`, which visits every seat's `Command` zone and adds each card's modifiers whose `Eminence` is set (and which price this kind of announcement: `CostModifier.pricesAnnouncement`, the cast / activation / special-action partition that was inlined in the battlefield loop and is now one method both loops call).

The card side is `effects.Eminence(modifier)`, which sets the flag, so a card file reads as the card does:

```go
CostModifiers: []game.CostModifier{
    Eminence(CostsLess(1, "Eminence — other Sphinx spells you cast cost {1} less to cast.",
        YourSpell(), OtherSpellOfCreatureType("Sphinx"))),
},
```

**It never applies from anywhere else.** A card in a hand, library, graveyard or exile has no cost modifiers gathered at all. A card in the command zone contributes **only** the modifiers marked `Eminence`: a Sphere of Resistance in a command zone taxes nobody (CR 113.6), because an ordinary modifier does not say it works there. `TestEminenceWorksOnlyInTheCommandZoneAndOnTheBattlefield` runs the six zones, and `TestACommandZoneCardsOrdinaryCostModifierDoesNothing` pins the second sentence.

### 2. "You" is the player whose command zone it is

A card in the command zone carries no controller of its own that a rule can rely on (the field is whatever the importer stamped). The modifier is bound with `Source.Controller` set to the seat that owns the zone, so `YourSpell()`, which compares the caster with `Source.Controller`, reads the printed "you": The Ur-Sphinx in my command zone discounts the Sphinxes **I** cast and no one else's (`TestEminenceStacksAcrossZonesAndBelongsToItsOwner`).

"Other" is `OtherSpellOfCreatureType`, which refuses the modifier's own source by `InstanceID`. The command-zone card and the spell it becomes on the stack are one object (a commander is cast from the command zone), so The Ur-Sphinx costs its full {6}{W}{U}{B} from the command zone, as printed.

### 3. The same accessor as a battlefield static, so ability removal is the same

The command-zone walk asks `costModifiersOf`, the accessor the battlefield walk uses, which reads `catalogAbilityKeyOf`. An object that is not on the battlefield has no layered characteristic in this engine, so nothing removes the abilities of a command-zone card; that matches the card's printing, since no effect in the set can reach a card there. A battlefield Ur-Sphinx that loses all its abilities stops discounting, exactly as any other static (`TestEveryCardDefSlotIsClassifiedForAbilityRows` and the ability-removal tests already cover the key). Two Ur-Sphinxes stack, one in each zone.

### 4. Pricing readers all see it

`CastPriceReadsTargetsForEffect` (the enumerator's "does the price depend on the targets" question) also visits the command zones for an eminence modifier that reads targets. Every other reader of the price (`CastSpell`, the legal-move enumerator, the auto-tap preview, the wire's `castable_here`) goes through `applyCostModifiersLocked` and so through the one gatherer; `TestEnumeratorPricesEminenceFromTheCommandZone` dispatches every enumerated cast against a clone to show the bot is never offered a price the engine refuses.

## Consequences

- The first command-zone static. Eminence **triggers** (Edgar Markov) and eminence **replacements/statics** (Arahbo) are not built; each would add one more gatherer that visits the command zones and honours a flag on its own declaration (`TriggeredAbility.Zones` already names the zones a trigger watches from, and is the likely home). Nothing here forecloses them, and nothing here pretends to do them.
- A card whose eminence has no declared shape ships without that line, and the card file says so (`Caveats`). The Ur-Sphinx is whole: its cost line is an `Eminence` modifier and its attack trigger needs nothing else.
- The wire is unchanged. A command-zone tile already lists the card's static rows (`ability_rows.go`: off the battlefield "the rows are what the card prints"), and the discount shows in the price the hand tile and the cast button already read.
- No snapshot change: the modifier is catalog data, read live and never stored.

## Amendments made alongside this ADR

The other four planeswalker pieces of #2797 are not new decisions, so each is a dated amendment to the ADR that owns its machinery:

- **[ADR 0093](0093-abilities-granted-to-other-permanents.md)** — a class grant of a loyalty ability ("Planeswalkers you control have '[−8]: …'", Kiora of Salt and Sand), with ADR 0109 §2 already lifting Decision 10's exclusion.
- **[ADR 0032](0032-planeswalkers.md)** — the CR 704.5i exemption (Sanctum Lurker).
- **[ADR 0018](0018-triggers-on-the-stack.md)** — "whenever you put one or more loyalty counters on a planeswalker" (Inspired Tethermage).
- **[ADR 0066](0066-granted-cast-and-play-permissions.md) (the activation twin, #1208)** — the stored half of an activation-timing statement, "until end of turn, you may activate loyalty abilities of Jace planeswalkers you control … any time you could cast an instant" (Jace's Machinations).
