---
title: "A computed mana value cap on a granted alternative cost"
date: 2026-10-14
issues: [3503]
---
**A computed mana value cap on a granted alternative cost** (#3503) — `GrantedAlternativeCost.MaxManaValueFunc` (with `MaxManaValueText` for the ability row) is a cap that is a function of the game and the caster, beside the fixed `MaxManaValue`. `covers` now takes the game and the caster and is the one reader of both caps, called from `grantedAlternativeCostsLocked`, which the cast gate, the legal-move enumerator and the view's picker all derive their offers from, so the count is read each time and none of them can hold a stale one. `effects.CastFromHandWithoutPayingManaCostUpToCreatureCount` (key `granted-free-up-to-creature-count`, hand only) counts the creatures the caster controls off effective types. Amends ADR 0118 §3. **Cards:** Omnipresence (Full).
