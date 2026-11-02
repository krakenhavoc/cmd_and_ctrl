---
title: "Mana that can only be spent to cast spells from exile"
date: 2026-11-02
issues: [4460]
---
**Mana that can only be spent to cast spells from exile** (#4460, CR 601.2a, [ADR 0040](decisions/0040-mana-pipeline.md)) — the positive twin of #2811's tag: `game.ManaRestrictCastFromExile` admits a payment only when it is a cast (`SpendPurposeCast`) whose `ManaSpendContext.CastFrom` is exile, and refuses an activation, a cast from the hand, a graveyard, the command zone or a library, and a cast whose zone the caller did not name (weaker than printed, never stronger). It rides the same spend context, so the view, the bot's planner (`protocol.ManaRestrictionsAdmitCast`) and the engine agree. **Cards:** Interdimensional Web Watch — Full (its two-colour mana ability carries the tag; its enters trigger is Reckless Impulse's exile-and-play). Lady Octopus, Inspired Inventor reaches its "artifact spell with mana value N or less" free cast through the existing `PermissionFilter.CardType` and `MaxSpellManaValue`, so no filter field was added.
