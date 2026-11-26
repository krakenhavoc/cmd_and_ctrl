---
title: "You can't spend mana to cast this spell"
date: 2026-11-25
issues: [5519]
---
**You can't spend mana to cast this spell** (#5519, CR 601.2f, CR 601.2h) — `Spec.CantSpendManaToCast` (`game.CardDef.CantSpendManaToCast`, read through `CatalogCantSpendManaToCast`) refuses the cast with `ErrCantSpendMana` while any mana is left in the total cost once the cost modifiers, convoke, delve and an assist payment have taken their share (X settled at its announced value). The one check sits in `effectiveCostLocked`, which the charge reads, and in `priceCastLocked`, which the view, the preview and the bot enumerator read, so a cast that cannot be paid without mana is never offered. The empty cost that remains is charged as nothing, so the pool is never touched. It is an additive card-def field and a package-level hook, nothing reachable from `Game`. **Cards:** Hogaak, Arisen Necropolis (Full).
