---
title: "A graveyard cast at its mana cost that counts toward a once-each-turn limit only when made"
date: 2026-10-15
issues: [3785]
---
**A graveyard cast at its mana cost that counts toward a once-each-turn limit only when made** (#3785, [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md) amendment 2026-10-15) — "you may cast target instant or sorcery card from your graveyard ... Do this only once each turn". `effects.OfferGraveyardCast` grants a flash-timed, pass-closed cast at the card's mana cost under `AltCostKeyGrantedCost` with the exile replacement and a cast follow-up that writes the once-each-turn tally, so only a cast that is made spends the turn's use; a claim of the permission's own synthesised offer (`AlternativeCost.FromPermission`) now counts as the permission's cast even for a card that prints flashback. `effects.CastFromGraveyard()` is the "spells you cast from your graveyard" cost predicate, and the batch trigger was `OncePerBatch` over the two life-loss events. **Cards:** Emet-Selch of the Third Seat (one declared caveat: the spell is exiled whenever it would leave the stack).
