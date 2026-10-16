---
title: "Creatures dealt damage by one creature this turn"
date: 2026-10-15
issues: [1946]
---
**Creatures dealt damage by one creature this turn** (#1946, CR 701.19c, 614.1a, 611.2c, 400.7, [ADR 0108](decisions/0108-turn-scoped-effects-object-history-and-damage-shields.md) amendment of 2026-10-15) closes the row: no card waits on it now.
- **The record.** `TurnTally.DamagedObjects` lists every damage event dealt to a permanent this turn as a (source, victim) pair of `ObjectRef`s, combat or not, written by the tally listener from `EventDealDamage`, so damage prevented to 0 is not there. A flickered creature is a new object on either side. It resets with the turn, and clone, undo and the snapshot carry it (additive). Readers: `ObjectDealtDamageToObjectThisTurn` and `PermanentDealtDamageToPermanentThisTurn`; Zurgo Helmsmasher and Bushi Tenderfoot now read it instead of scanning the event log.
- **Runesword's riders.** A new scope, `ScopeDealtDamageBy`, whose `cantBeRegenerated` or `exileIfWouldDie` mod names the dealer in `Objects[0]` (`Game.CantBeRegeneratedIfDealtDamageByForEffect`, `Game.ExileIfDealtDamageByWouldDieThisTurnForEffect`). `Mod.DamageSince` counts only the damage dealt after the effect began. The record lasts until cleanup even if the dealer leaves.
- **Bone Shaman's ability.** `CardDef.DealtDamageCantBeRegenerated`, read live off the battlefield and grantable through `AbilityGrant` (the `AttackLimits` shape), so it ends when the creature leaves or loses the ability. Both reach `regenerationRefusedLocked`, so a refused shield is not spent and the NO REGEN chip shows.
- **The exile static.** `effects.ExileIfDealtDamageByThisWouldDie` and `ExileIfDealtDamageByEnchantedWouldDie` are catalog replacements over the record.
- **Cards** (all Full): Bone Shaman, Runesword, Kumano, Master Yamabushi, Frostwielder, Incendiary Oracle, Kumano's Pupils, Kumano's Blessing.
