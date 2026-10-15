---
title: "A Saga creature that exiles from each graveyard and lets you cast those cards with mana of any type"
date: 2026-10-15
issues: [3773]
---
**A Saga creature that exiles from each graveyard and lets you cast those cards with mana of any type** (#3773, [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md) and its 2026-09-24 any-type amendment) — no new engine seam: Summon: Esper Valigarmanda is an assembly of `ChapterTrigger`, a chosen-card prompt per graveyard, `ExileLinkedThenForEffect` (the cards carry `Card.ExiledWith` naming the Saga as an object, CR 607.2a), `AddMana` over `ManaN` read from the Saga's lore counters through `SourcePermanent` (last-known when it has gone), and a one-cast `CastPermission` with `AnyType` in the pass-closed window cascade and Fevered Suspicion use. **Cards:** Summon: Esper Valigarmanda (Full). The `ExiledWith` stamp on an ability's exile that #2651 proposes to generalise is not needed here: `ExileLinkedThenForEffect` already stamps it.
