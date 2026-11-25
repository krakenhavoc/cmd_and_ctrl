---
title: "Lose all abilities except mana abilities"
date: 2026-11-25
issues: [6246]
---
**Lose all abilities except mana abilities** (#6246, CR 613.1f, 605.1a) — Blood Sun's "All lands lose all abilities except mana abilities". `StaticAbility.KeepsManaAbilities` narrows a `RemovesAbilities` static: the removal is applied as any other (`AbilitiesRemoved`, so triggered, non-mana activated, static and replacement abilities and keywords are gone), and `Characteristic.ManaAbilitiesKept` lets the object's own declared mana rows answer (`untaxedManaAbilityRows`). A layer-6 grant that adds only mana abilities survives it (Chromatic Lantern), and the intrinsic mana ability of a basic land type stays as under any removal. A full removal in the same pass, earlier or later, still takes everything (`game/lose_abilities_except_mana.go`). Card helper: `effects.LandsLoseAllAbilitiesExceptMana()`. **Cards:** Blood Sun (b216) is built by its OLD batch.
