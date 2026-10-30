---
title: "A colour filter on next-spell permissions"
date: 2026-10-30
issues: [4615]
---
**A colour filter on next-spell permissions** (#4615, CR 105.2b, CR 118.9, [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md)) — "The next red or green creature spell you cast this turn can be cast without paying its mana cost" (World War Hulk).
- **`PermissionFilter.ColorsAny`** is an additive any-of colour clause: a string of colour letters (`"RG"`), matched in `PermissionFilter.Matches` against the card's effective colours, so a promise, a granted cost offer, the view and the bot's move list all see it through the one predicate. It ANDs with the other clauses (`CreatureOnly` plus `"RG"`), and empty constrains nothing.
- **Snapshot:** one recorded shape line per path that reaches a `PermissionFilter` (`...filter.colorsAny`, omitempty) in `v9.txt`.
- **Cards (1, Full):** World War Hulk.
