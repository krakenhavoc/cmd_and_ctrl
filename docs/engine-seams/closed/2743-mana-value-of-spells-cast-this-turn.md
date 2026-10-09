---
title: "The total mana value of spells cast this turn"
date: 2026-10-08
issues: [2743]
---
**The total mana value of spells cast this turn** (#2743, CR 202.3, CR 601.2i) — `CastTally.ManaValue` is a running sum of the mana values of the spells a player has cast this turn. It is written where the rest of the tally is, in `CastSpell`, as the spell becomes cast, and read off the spell on the stack by `Game.castManaValueLocked` (`CastManaValueForEffect` for the catalog). An X spell counts its X (CR 202.3e), and a spell cast face down counts 0 (CR 708.4). It is an additive snapshot field (`manaValue`, omitted when zero). `effects.manaValueOfOtherSpellsYouCastThisTurn` reads "other spells": the tally less the resolving spell's own value, unless the spell is a copy, which was never cast. **Card:** Call Forth the Tempest (Full). Spells cast off its two cascade triggers resolve first and count; a test pins it.
