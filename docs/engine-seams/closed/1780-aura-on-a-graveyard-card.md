---
title: "Auras that enchant a card in a graveyard"
date: 2026-10-13
issues: [1780]
---
**Auras that enchant a card in a graveyard** (#1780, CR 303.4, CR 303.4c, CR 400.7, CR 701.3b, CR 603.7, [ADR 0154](decisions/0154-auras-that-enchant-a-card-in-a-graveyard.md)) — `game/attach_graveyard.go` is the rules side, `effects/graveyard_auras.go` the card side.
- **The clause.** `effects.EnchantCardInGraveyard(noun, preds...)` is the existing `card_in_graveyard` target clause, so the cast, the CR 608.2b re-check, the view's legal set and the bot enumerator treat it as Reanimate's.
- **The attach.** `AttachForEffect` accepts a card host off the battlefield only when the attachment is an Aura whose catalogued enchant clause lists the host's zone and accepts it; anything else is still a quiet `EventAttachSkipped`. `Card.AttachedEpoch` records the host's CR 400.7 epoch for an off-battlefield host only, and `attachmentLegalLocked` reads it there, so a card that left and came back is a new object the Aura falls off. Battlefield hosts are unchanged (#3159).
- **The gained clause.** `Card.EnchantSwapped` and `Card.EnchantReturned` ("enchant creature put onto the battlefield with this Aura") are read first by `attachmentLegalLocked`, cleared at both CR 400.7 sites, and enforced by `AttachForEffect` as CR 701.3b (a swapped Aura refuses to move). The general CR 701.3b check is #3160.
- **The trigger.** `effects.AnimateEnchantedCard{Tapped}` re-checks the Aura is on the battlefield (CR 603.4), swaps the clause, returns the card under its controller's control through `ReturnFromGraveyardToBattlefieldThenForEffect` (a continuation, so a paused entry attaches once it lands), names and attaches to the new permanent, and schedules "that creature's controller sacrifices it" as an event-keyed delayed trigger with `IndefiniteDuration()`, which outlives every cleanup step and fires once.
- **Snapshot.** `attachedEpoch`, `enchantSwapped`, `enchantReturned` on the card, additive in v8.
- **Client.** The Aura's ENCHANTING badge names the graveyard card and its owner (`graveyardHostLabels`), so the table sees what is coming back while the trigger waits. No wire change.
- **Cards:** Dance of the Dead (Full; its upkeep untap is asked of the enchanted creature's controller) and Animate Dead (Caveats: put onto the battlefield without being cast it enchants nothing, CR 303.4f, #3158).
- **Not built:** Necromancy (#3156), Spellweaver Volute (#3157), the CR 303.4f host choice for every Aura (#3158), and reanimation-family bot pricing (#3161).
