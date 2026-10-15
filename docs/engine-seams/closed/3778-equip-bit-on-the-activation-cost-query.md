---
title: "The equip bit on an activation's cost query"
date: 2026-10-15
issues: [3778]
---
**The equip bit on an activation's cost query** (#3778, [ADR 0020 amendment 2026-10-15](decisions/0020-activated-abilities.md)) — `game.AbilityCostSubject.Equip`, copied from `ActivatedAbilityShape.Equip` by `abilityCostQueryLocked`, lets a board cost modifier price equip abilities and no others, without reading the label. Card side: `effects.AnEquipAbilityCost`, `YouActivate`, `ThatTargetThisPermanent`, and the two compositions `EquipAbilitiesYouActivateCostLess` and `EquipAbilitiesThatTargetThisCostLess` (the second sets `ReadsTargets`, so the enumerator and the view price each target on its own). **Cards (Full):** Helitrooper, Dwarven Mauler, Fervent Champion, Bureau Headmaster and Cloud, Planet's Champion. **Not written, no engine field missing:** Bladehold War-Whip, Strong Back, Nahiri, Storm of Stone, Fighter Class, Éowyn, Lady of Rohan.
