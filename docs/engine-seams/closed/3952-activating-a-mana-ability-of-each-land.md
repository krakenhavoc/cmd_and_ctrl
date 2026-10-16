---
title: "A player activates a mana ability of each land"
date: 2026-10-16
issues: [3952]
---
**A player activates a mana ability of each land** (#3952, CR 605.3a, 106.4, [ADR 0200](decisions/0200-activating-a-mana-ability-of-each-land.md)) — `effects.ActivateAManaAbilityOfEachLand` has a player activate one mana ability of each land they control, through the ordinary activation path (`game.ActivateLandManaForEffect`). `game.LandManaOptionsForEffect` lists each way they can do it now, with a picking slot's colours named up front, so a land with several ways asks one `option_pick`, a land with one activates it, and a tapped land adds nothing. A land with an ability whose mana cost the pool can't pay yet (a filter land) is asked after the others. `game.LoseAllUnspentManaForEffect` then empties the whole pool and returns what was lost. **Cards:** Drain Power (its caster adds the same types) and Pygmy Hippo ({C} for the amount at the next main phase this turn, `game.NextMainPhaseThisTurnForEffect`), both Full.
