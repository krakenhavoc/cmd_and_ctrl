# ADR 0183 — Compleated: entry loyalty from Phyrexian life payments

Date: 2026-10-14
Status: Accepted
Sprint: S58 — Deck requests, October batch
Issues: #2946 (seam: compleated entry loyalty from Phyrexian-symbol life payments), #2910 (All Will Be Black).
Builds on: [ADR 0032](0032-planeswalkers.md) (planeswalkers and starting loyalty), [ADR 0013](0013-replacement-effects.md) (the CR 614 pipeline), the cast provenance record (`game/cast_provenance.go`, CR 400.7d).
Numbering: every remote branch was listed with `git ls-tree --name-only <ref> docs/decisions/` after `git fetch --all --prune`. The highest number anywhere was 0182 (`develop`, your own model key for AI seats), so this is 0183. It was drafted as 0182 before that merged.

## Context

CR 702.150a: compleated means "If this permanent would enter with one or more loyalty counters on it and the player who cast it chose to pay life for any part of its cost represented by Phyrexian mana symbols, it instead enters the battlefield with that many loyalty counters minus two for each of those mana symbols."

Vraska, Betrayal's Sting ({4}{B}{B/P}, loyalty 6) is the first catalog card with it. The Wizards rulings (2023-02-04) add two facts:

- "Other replacement effects that would change the number of loyalty counters Vraska enters with will apply as normal." Compleated is one replacement among several on the same placement, and CR 616 orders them.
- "The compleated ability looks only at whether a player chose to pay 2 life for a Phyrexian mana symbol as they were casting the spell. If a player paid life for some other reason while casting the spell, that will not reduce the number of loyalty counters."

The engine has two gaps:

1. **No provenance.** `PaidCost.LifePaid` adds up every life the cast paid: Phyrexian symbols, `{B}` paid with life under K'rrik's grant (ADR 0131), and fixed life costs. Nothing records how many printed Phyrexian symbols were paid with life.
2. **No entry hook.** Starting loyalty is placed by `stampStartingLoyaltyLocked` as the planeswalker enters, through `AddCounterForEffect`, so the CR 614 counter window (Doubling Season, CR 122.6) already sees it. Nothing marks that placement as the starting loyalty, so a replacement cannot tell it from a later +1.

Subtracting counters after entry would be wrong. It would run after every other replacement, which is an order the controller did not choose, and it would let the planeswalker exist briefly with the wrong loyalty.

## Decisions

### 1. Count the printed Phyrexian symbols paid with life

`PaidCost.PhyrexianLife` is the number of printed Phyrexian symbols, hybrid Phyrexian included (CR 107.4f), that the announcement struck out to be paid with life. `strikePhyrexianLifeLocked` sets it on the one path every cast and activation takes, as the count of `ColorRequirement.Phyrexian` slots in the cost before the strike minus the count after. A `{B}` that K'rrik lets be paid with life is `LifeGranted`, not `Phyrexian`, and is not counted. Neither is a fixed life cost. `LifePaid` is unchanged.

### 2. The permanent remembers it

`CastProvenance.PhyrexianLife` copies `item.Paid.PhyrexianLife` at the entry finisher, beside X and the mana (CR 400.7d). It is zero for anything not cast: a reanimated, flickered or searched planeswalker, a token, and a CR 707.10 copy of a spell, which paid nothing and was not cast. `Card.PhyrexianLifePaid()` reads it.

### 3. The starting-loyalty placement is marked

`ReplacementEvent.StartingLoyalty` is true only on the counter event `stampStartingLoyaltyLocked` opens: the CR 306.5b placement as the permanent enters. Doublers do not read it, so they see the same placement they always did.

### 4. Compleated is the card's own replacement

`effects.Compleated()` is a `game.ReplacementEffect` the card declares like any static ability (CR 614.12: an effect that modifies how its own permanent enters). It applies to a `RepEventCounter` that is the `StartingLoyalty` placement of loyalty counters on this object, when `PhyrexianLifePaid() > 0`. It lowers the delta by two per symbol. At zero or below it cancels the placement: the planeswalker enters with no loyalty counters and dies to CR 704.5i, as printed.

Because it is an ordinary replacement in the window, a Doubling Season beside it is a CR 616 question for the planeswalker's controller. Neither is a self-replacement, a control or copy choice, or a back-face entry, so by 616.1e the controller chooses either one first. That is the ruling's "apply as normal". A permanent that enters with its abilities removed contributes no catalog replacement (`catalogAbilityKeyOf`), so compleated does not apply to it.

### 5. Snapshot compatibility

Both new fields are additive and `omitempty` in the v8 shape (`paid.PhyrexianLife` on the four `PaidCost` paths, `provenance.phyrexianLife`). A file written before them reads zero, which is "no life paid": the planeswalker already on the battlefield is unaffected, because compleated only acts as it enters. No frozen fixture changes.

### 6. Vraska, Betrayal's Sting

- **0:** you draw a card and lose 1 life, then proliferate (the optional selection already used by the catalog's proliferate cards).
- **−2:** target creature becomes a Treasure artifact with "{T}, Sacrifice this artifact: Add one mana of any color" and loses all other card types and abilities. The Irencrag's shape: one indefinite scoped effect pinned to the object (CR 611.2a). It sets the types to Artifact — Treasure in layer 4 (supertypes kept, per the ruling), removes all abilities in layer 6, and then grants the mana ability bundle, so the grant survives the removal.
- **−9:** if the target player has fewer than nine poison counters, they get the difference, computed at resolution and then placed through the player-counter window, so a replacement such as Melira changes the result (ruling).

## Consequences

- A second compleated planeswalker (Nissa, Ajani, Tamiyo, Jace, Lukka, Tezzeret) is `Replacements: []game.ReplacementEffect{effects.Compleated()}` and nothing else.
- Under K'rrik, the engine picks which symbols a life claim strikes (ADR 0131): symbols the pool cannot pay go first, then printed order. A caster who claims one symbol on Vraska under K'rrik may therefore have a `{B}` struck rather than the `{B/P}`, and Vraska then enters with full loyalty. That is the weaker direction for the caster, never stronger, and stays until a cast can name which symbols it pays with life.
- Restore points taken before this change restore with zero, which is right for every planeswalker already in play.
