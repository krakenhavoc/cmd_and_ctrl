---
title: "Equip-only mana"
date: 2026-10-30
issues: [4600]
---
**Equip-only mana** (#4600, [ADR 0020](decisions/0020-activated-abilities.md) amendment of 2026-10-30, Decision 75, CR 106.6, CR 702.6a) — "Spend this mana only to cast Equipment spells or activate equip abilities" (Ronin, Shadow Stalker). The spend context for an activation is built from the source card alone, which cannot tell an equip ability from an Equipment's other abilities, so `ManaSpendContext` gained an `Equip` bit beside `PowerUp`, set from `ActivatedAbilityShape.Equip` by the activation, the legal-move enumerator and the `?ability=` cost preview. `game.ManaRestrictEquip` (`"purpose:equip"`) matches it, and the card joins it to `ManaRestrictCast` plus `ManaRestrictSubtype("Equipment")` with `ManaRestrictAnyOf`. The snapshot gains the spend context's `Equip` key (`pendingChoices[].assist.spend`), additive under v9. **Cards:** Ronin, Shadow Stalker (Full).
