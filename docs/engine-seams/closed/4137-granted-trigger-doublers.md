---
title: "A trigger doubler one permanent grants another"
date: 2026-10-19
issues: [4137]
---
**A trigger doubler one permanent grants another** (#4137, CR 603.2d, CR 113.10, [ADR 0018 amendment 2026-10-19](../../decisions/0018-triggers-on-the-stack.md)) — `effects.AbilityGrant.TriggerDoublers` carries a doubler on a granted bundle, `mergeCatalogParts` merges it into the recipient's composite definition, and the existing doubler scan finds it through `CatalogAbilityKey`, so the recipient's controller is "you", each recipient is one more doubler, and an ability removal or the grantor leaving ends it. `hasTriggerModifierLocked` also asks the grantor, so an entering permanent is read with fresh layers.
- **Cards (2, both Full).** Dungeon Delver ("Commander creatures you own have 'Room abilities of dungeons you own trigger an additional time.'") and The Masamune (the equipped creature's creature-dying doubler for its own and your emblems' triggers, with first strike and "must be blocked if able" while attacking).
