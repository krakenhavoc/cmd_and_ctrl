---
title: "A creature that returns to the battlefield as an Aura"
date: 2026-10-23
issues: [2809]
---
**A creature that returns to the battlefield as an Aura** (#2809, CR 303.4f / 303.4g, CR 205.1a, CR 613.1d, CR 613.1f, CR 613.6, CR 704.5m, [ADR 0154](decisions/0154-auras-that-enchant-a-card-in-a-graveyard.md)'s 2026-10-23 amendment, decision 21).
- **The clause.** `Spec.ReturnedAura` (`CardDef.ReturnedAura`) is the enchant clause of the Aura form, read through `ReturnedAuraSpecFor`. It is not `Targets`, which a creature spell would read as its own cast-time target. `auraEnchantSpecOf` reads it first for a permanent whose `Card.ReturnedAura` is set, so `attachmentLegalLocked` holds the Aura to "Forest you control" (CR 704.5m) and an Aura of this kind attached to nothing is illegal.
- **The entry.** `effects.ReturnFromGraveyard{AsAura}` (`Game.ReturnFromGraveyardAsAuraForEffect`, `game/returned_aura.go`) sets `ReplacementEvent.entersAsAura`, so `enteringAuraLocked` judges the entry an Aura entry with that clause: CR 303.4f's host choice is asked of the controller as it enters (one candidate is forced), and with none the card stays in the graveyard (CR 303.4g). The landing sets `Card.ReturnedAura` before the attach and `EventETB` (`seedReturnedAuraEntryLocked`), so an enters trigger finds an Aura and a "whenever a creature enters" watcher sees none.
- **The form.** `effects.ReturnedAuraForm(grant)` is three statics that apply only while the flag is set: the layer-6 grant to the enchanted permanent, the layer-4 type set (Enchantment, Aura; supertypes stay) and the layer-6 removal of the card's own abilities. The grant is declared first so the removal has not yet silenced its source (CR 613.6). `effects.WhenThisDiesReturnItAsAnAura` is the trigger, with "if it was a creature" read from last-known information (CR 603.4).
- **Snapshot.** `cards[].returnedAura`, additive in v9 and cleared at both CR 400.7 sites.
- **Cards:** Harold and Bob, First Numens (Full). Bronzehide Lion needs only its own clause and grant.
