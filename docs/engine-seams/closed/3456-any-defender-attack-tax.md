---
title: "Attack taxes on every attack, with X from the activation"
date: 2026-10-14
issues: [3456]
---
**Attack taxes on every attack, with X from the activation** (#3456, CR 508.1a, CR 611.2; follows [#3398](engine-seams/closed/3398-timed-attack-tax.md)) — War Tax: "{X}{U}: This turn, creatures can't attack unless their controller pays {X} for each attacking creature they control."
- **Storage.** The `ModAttackTax` record's `Player` may now be zero, the any-defender form: it taxes an attack on any player or planeswalker, where a named `Player` still defends that player alone. No snapshot key is added, so v8 is untouched; a zero-player record restores as written, and an older binary already refuses the kind.
- **Mana.** The record's `Mod.Text` is the cost already resolved from the activation's X (`{2}` for X=2), so every activation writes its own record and several stack as separate lines in the one price. X=0 writes nothing.
- **Reading.** Unchanged: `priceAttackDeclarationLocked` asks `durationAttackTaxLines`, so the declaration verb, the legal-move enumerator and the bot all charge it.
- **Card side.** `Game.AnyAttackTaxForEffect(source, mana, duration, label)`; a zero duration is end of turn.
- **Cards:** War Tax (Full).
