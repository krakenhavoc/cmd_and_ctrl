---
title: "Must attack an opponent with the most life"
date: 2026-10-08
issues: [2744]
---
**Must attack an opponent with the most life** (#2744, CR 508.1d, [ADR 0045 amendment 2026-10-08](decisions/0045-combat-restrictions.md)) — `AttackRequirement.MostLifeOpponentOf` is a requirement obeyed only by an attack on a player still in the game who is an opponent of that seat and has the most life among its opponents (`Game.isMostLifeOpponentLocked`); ties count each tied player. Life is read when the attack is judged, so the CR 508.1d maximisation, the verb refusals, the checkpoint, `MustAttackForEffect` and the enumerator all see the current totals. The refusal sentence reads "… must attack an opponent with the most life if able." Card side: `effects.AttacksAnOpponentWithTheMostLifeEachCombat(exempt)` and `youControlACreatureNamed`, which reads layer 1's copied name. **Card:** Galactus, Devourer of Worlds (Full). Gideon Jura's requirement on one permanent (#2567) should be a sibling field.
