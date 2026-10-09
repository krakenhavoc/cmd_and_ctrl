---
title: "Bestow"
date: 2026-10-09
issues: [2862]
---
**Bestow** (#2862, CR 702.103, CR 608.3b, [ADR 0141](decisions/0141-bestow.md)) — `game/bestow.go` is the rules side, `effects/bestow.go` the card side.
- **The cast.** `effects.Bestow("{2}{B}{B}")` is an alternative cost (key `bestow`) whose `Targets` is the enchant creature clause and whose `Bestow` bit makes the spell an Aura. `CastSpell` sets `Card.Bestowed` right after the claim, so every later gate reads an Aura spell (CR 702.103d), and stamps it on the card on the stack. `AlternativeCost.CastFaceOf` sets it too, so the view's timing stamp and the bot enumerator judge the same object.
- **Not a creature (CR 702.103b).** Off the battlefield the type readers' cold path applies it to the printed baseline (`bestowBaseline`): Enchantment — Aura, no creature type or creature subtypes, supertypes kept. On the battlefield it is a layer-4 effect from the board (`bestowContinuousEffectsLocked`), at the permanent's entry, so a later type-changing effect still applies on top. Not a copiable value.
- **Resolution.** A bestowed spell whose target is illegal is not countered (CR 702.103e, 608.3b): `resolveTopOfStackLocked` ends the bestow on the stack card, drops the targets and resolves it as a creature spell. Otherwise the entry seeds `Bestowed` from the stack card before the Aura attach and `EventETB`. A copy of a bestowed spell makes a bestowed token (CR 702.103c).
- **Unattached (CR 702.103f).** `attachmentLegalLocked` reads a bestowed Aura's enchant creature, and `attachmentSBALocked` unattaches an illegal or unattached bestowed Aura and clears the flag instead of burying it (the exception to CR 704.5m), so it stays as an enchantment creature and a 0/0 dies to CR 704.5f. `UnattachForEffect` clears it too.
- **Snapshot.** `Card.Bestowed`, additive in v7 (`bestowed`).
- **Client and bot.** The offer reaches the cast picker as any other alternative cost, with the enchant creature clause as its target. The heuristic prices a bestowed cast onto its own creature above the creature cast (`Config.BestowShare`) and onto an opponent's below passing.
- **Cards (6, all Full):** Nighthowler (its caveat cleared), Eidolon of Countless Battles, Boon Satyr, Celestial Archon, Hopeful Eidolon and Nyxborn Rollicker. `effects.PumpSelfCreatureOrAttachedPer` is "this creature and enchanted creature each get …".
- **Not built:** the other Commander-legal bestow cards are card work (several need their own seams, such as Detective's Phoenix's collect evidence and graveyard cast, or Hypnotic Siren's control change on the host).
