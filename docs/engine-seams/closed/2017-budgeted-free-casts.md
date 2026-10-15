---
title: "Free casts with a total mana value budget"
date: 2026-10-15
issues: [2017]
---
**Free casts with a total mana value budget** (#2017, CR 601.2, CR 608.2g, CR 202.3, [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md) amendments of 2026-10-14 and 2026-10-15) closes the "Free casts with a total mana value budget" row: no card waits on it now.
- **The running total.** `game.CastPermission.ManaValueBudget` is the remaining total one `ScopeCards` permission may still open across every card it names. `spellManaValueWithinCap`, the gate the cast path, the view's offers and the bot enumerator share, refuses a card whose mana value is over what is left, and a cast spends it through `consumePassClosedGrantLocked` (a fresh `Cards` slice and budget pointer, never written in place). Built for the copies of Uldaros Theorix by the 2026-10-14 amendment.
- **Cards an impulse exile landed.** `effects.ExileTopWithPermission.TotalManaValueAtMost` stamps the same free, flash-speed, cast-only permission over the cards that reached exile, closing on the holder's next pass with `LapseStaysInExile`. The impulse exile already granted one permission over the landed cards, so the seam's "helper" was this field; it refuses the options that mean another window. A land among the cards is stranded and an uncast card stays in exile.
- **Primeval Spawn.** `effects.SelfExiledInsteadIfNoManaSpent` is "if this creature would enter and it wasn't cast or no mana was spent to cast it, exile it instead", Freestrider Commando's read (`EntrySpentForEffect(ev).None()`) with the entry rewritten to exile, so an exiled Spawn never enters and never leaves. Its leaves-the-battlefield ability exiles the top ten and offers the free casts against a total of 10. Declared caveat: with strict mana off a cast Spawn is not known to have spent nothing.
- **Cards:** 1 (Primeval Spawn), beside Uldaros Theorix.
