---
title: "\"Can't attack you\" from a spell, held by a counter"
date: 2026-10-13
issues: [3400]
---
**"Can't attack you" from a spell, held by a counter** (#3400, CR 508.1c, CR 611.2b) — Promise of Loyalty, from the Tidus, Yuna's Guardian deck request (#3246). `Characteristic.AttackTargetRestrictions` was written only by layer-6 statics (ADR 0106 §2); a resolved effect had no way to put one on a creature.
- **The mod.** A new `ScopedEffect` kind, `cantAttackPlayer` (`game.CantAttackPlayerMod`, layer 6), reads `Mod.Player`: the protected player, fixed when the spell resolves (CR 611.2c). It appends `AttackTargetRestriction{NotPlayer}`, attributed to the record's source. Registration and restore refuse a record naming nobody.
- **The clause.** `AttackTargetRestriction.NotPlayer` refuses that player and any planeswalker they control, and not a battle they protect (the text does not say it). It is read by `canAttackTargetWithLocked` with every other restriction, so both declaration verbs, the CR 508.1d requirement search, the enumerator's per-attacker list and the digest follow. The refusal sentence names the player ("Grizzly Bears can't attack Alice or planeswalkers Alice controls (Promise of Loyalty)."), and the card shows the CAN'T ATTACK chip.
- **The duration.** The existing counter-held duration (ADR 0109 §2): `effects.CounterThenWhileItHasIt` puts the vow counter on and pins the record to the creature as the same object, and it ends for good when the last vow counter leaves.
- **Snapshot.** One additive key within schema v8 (`NotPlayer` on the restriction, recorded in `snapshot_shape/v8.txt`); the mod kind rides the existing record. No closure route added.
- **Promise of Loyalty.** Each player, from the active player, chooses one creature they control (a choice, not a target; Summon: Valefor's chained `ChoosePermanents`). Once the last has answered, the vow counters go on and every other creature is sacrificed together.
- **Cards:** Promise of Loyalty (Full). Elrond of the White Council (#2926) will want the sibling "can't attack its owner" mod.
