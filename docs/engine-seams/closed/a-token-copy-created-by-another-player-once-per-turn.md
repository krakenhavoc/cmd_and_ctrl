---
title: "A token copy created by another player, once per turn"
date: 2026-11-25
issues: [6174]
---
**A token copy created by another player, once per turn** (#6174, CR 111.2, CR 707.2) — already on the engine. A token creation names its creator (`TokenCreation.Controller`), so any player can create the copy and own it. A trigger's target can be built from its event (`TriggeredAbility.TargetsFrom`, `effects.WithTargetsFrom`), which is how "target player other than its controller" is written. The once-each-turn limit on a "you may" is the per-object `NoteDoneThisTurn` record Terrasymbiosis uses. The copy reads the entering token's last-known copiable values (`TokenCopyTemplateForObjectForEffect`), and "then … if an opponent created a token this way" is the creation's keyed tail. No engine change. **Cards:** Lucy MacLean, Positively Armed (Full).
