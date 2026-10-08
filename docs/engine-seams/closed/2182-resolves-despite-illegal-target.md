---
title: "Abilities that resolve even if their target becomes illegal"
date: 2026-10-08
issues: [2182]
---
**Abilities that resolve even if their target becomes illegal** (#2182, [ADR 0019](decisions/0019-structured-targeting.md) amendment 2026-10-08): `TargetSpec.ResolvesIfIllegal`, set with `.StillResolves()`, exempts an item from the CR 608.2b all-targets-illegal removal in `spellAllTargetsIllegalLocked`. It is a clause-level field read off the announced clauses, so it needs no stack-item or snapshot change and works for a spell, a trigger or a modal option. The effect treats an illegal target as unaffected (`Context.IsTargetLegal`), and `ExchangeControl.ApplyAndReport` reports whether an exchange happened. **Cards:** Gilded Drake (Full).
