---
title: "Return a card transformed, attached to a named host"
date: 2026-11-12
issues: [5677]
---
**Return a card transformed, attached to a named host** (#5677, CR 303.4f / 303.4h / 303.4i, CR 712.14a, CR 712.8a, CR 400.7) — `Game.ReturnFromGraveyardTransformedForEffect` had no host, so a back face that is an Aura asked its controller CR 303.4f's question as it entered: Radiant Restraints ("attached to target opponent") could have gone on any player, its controller included.
- **The entry.** `Game.ReturnFromGraveyardTransformedAttachedForEffect(cardID, controller, host)` sets the back face on the card in the graveyard as the plain transformed return does, and seeds `host` on the entry event as the fixed host every other "attached to" entry uses (`ReplacementEvent.auraHost` / `auraHostFixed`, `game/aura_host.go`). The Aura is attached between the zone move and `EventETB`, and nobody is asked.
- **An illegal host.** The host is checked against the back face's own enchant clause for the entering controller. One the Aura can't enchant, or one that is gone, cancels the entry (CR 303.4i): the card stays in the graveyard and is put back on its front face (CR 712.8a). A back face that is not an Aura, an Equipment or a Fortification enters unattached (CR 303.4h).
- **The card side.** `effects.ReturnFromGraveyard{Transformed: true, AttachedTo: ref}` is the primitive, and `effects.ReturnThisCardTransformedAttachedToTarget` is the resolution of "return this card to the battlefield transformed under your control attached to target …": the source must still be the graveyard object it became when it left the battlefield (CR 400.7), and the host is the trigger's target.
- **Snapshot.** None. The fixed host already rides a paused entry (`auraHost`, `auraHostFixed`), and no effect key, mod kind or wire field is new.
- **Cards:** Radiant Grace // Radiant Restraints, Vengeful Strangler // Strangling Grasp (both Full).
