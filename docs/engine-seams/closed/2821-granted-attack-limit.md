---
title: "An attack limit granted to other permanents"
date: 2026-10-09
issues: [2821]
---
**An attack limit granted to other permanents** (#2821, [ADR 0093](decisions/0093-abilities-granted-to-other-permanents.md) amendment of 2026-10-09) — `AbilityGrant` gains an `AttackLimits` slot, and `mergeCatalogParts` merges it into the recipient's composite definition. The attack check already reads limits through the recipient's composite key, so a granted "no more than N creatures can attack this permanent" counts attacks on the recipient and ends with the grant. An attack limit is not a layer effect, so ADR 0093 Decision 10's refusal of granted statics doesn't apply. **Card:** Tomik, Orzhov Lawmage (Full; its caveat is gone).
