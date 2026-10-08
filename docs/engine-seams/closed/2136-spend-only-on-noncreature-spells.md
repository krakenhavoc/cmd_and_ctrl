---
title: "Mana spendable only on noncreature spells"
date: 2026-10-08
issues: [2136]
---
**Mana spendable only on noncreature spells** (#2136, [ADR 0113 amendment 2026-10-08 (fourth)](decisions/0113-small-seams-for-the-s58-deck-requests.md), CR 106.6 / 601.2h) — `game.ManaRestrictNotType(types...)` is a negated card-type tag in the mana-restriction vocabulary. It admits a cast whose spell has none of the named types and refuses an activation, an unlock and an unknown purpose, so it pays nothing a creature spell or an ability costs. No solver changed: the spend context already reaches them. Nardole, Resourceful Cyborg ships `full` (Doctor's companion is a deck-construction rule, as Partner is).
