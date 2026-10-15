---
title: "Casting a copy of the card an Aura enchants"
date: 2026-10-14
issues: [3157]
---
**Casting a copy of the card an Aura enchants** (#3157; CR 707.12, CR 608.2h, CR 701.3b, CR 704.5m; [ADR 0154](decisions/0154-auras-that-enchant-a-card-in-a-graveyard.md)'s 2026-10-14 Spellweaver Volute amendment).
- **The enchanted card, read at resolution.** `game.EnchantedCardObjectForEffect` returns the live Aura's host. Once the Aura has left, it returns the host recorded in `PermanentInfo.AttachedObject`, which is written only while the link still named that object. A card exiled before the Aura left therefore reads as nothing.
- **The copy.** `effects.CopyAndCast` gains `LastKnown`, which copies a card the effect saw in a graveyard after it has left. It also gains `FollowUpObject`, carried on `CastPermission.FollowUpObject` to the follow-up as `CastFollowUp.Object`, so an "if you do" knows which card the copy was made from.
- **The move.** `game.GraveyardCardsAuraCanEnchantForEffect` lists every graveyard card the Aura could be moved onto: its clause plus the CR 701.3b test, with no targeting. The follow-up asks for one first (one candidate is forced), attaches to it, then exiles the original. With no candidate the original is still exiled and the Aura falls off.
- **Snapshot.** `castPermissions[].followUpObject` and `lastKnownPermanents{}[].attachedObject`, additive in v9.
- **Cards:** Spellweaver Volute (Full).
