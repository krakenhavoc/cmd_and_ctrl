---
title: "Mana that can't pay generic costs"
date: 2026-10-05
issues: [2170]
---
**Mana that can't pay generic costs** (#2170, [ADR 0040](decisions/0040-mana-pipeline.md)) — `game.ManaRestrictNoGeneric` is a symbol-level spend tag, read where the generic part of a cost is paid rather than where the object being paid for is matched: the pool solver (`ManaPool.attemptSpend`), the missing-mana breakdown, the distinct-colours pick and the auto-tapper's pool credit (`poolShortfalls`), so the view's castability, the bot's affordability probe and the payment agree. {N}, {X}, cost increases and commander tax are refused it; coloured, hybrid and Phyrexian symbols, the widened symbols of "spend mana as though it were mana of any color" and symbols a cast permission folded into the generic demand take it. Kruphix's conversion keeps the tag. **Cards:** Jegantha, the Wellspring (caveat: companion). **Not built:** The Rebellious Intelligence (not legal in Commander; its upkeep ability needs a random card from outside the game).
