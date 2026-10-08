---
title: "Triggers on entering from a graveyard"
date: 2026-10-08
issues: [2135]
---
**Triggers on entering from a graveyard** (#2135, [ADR 0113](decisions/0113-small-seams-for-the-s58-deck-requests.md) amendment 2026-10-08) — `EventETB` now carries `EnteredFrom` (the zone the permanent came from) and `EnteredFromOwner` (that zone's owner), stamped where the entry pipeline announces a landed permanent and on the sandbox move; a token has neither. They are fields of their own rather than `OldZone`, so the layer and tally listeners that key on an event's zones do not count an ETB as a second zone move. `effects.EnteredFromAGraveyard` and `EnteredFromYourGraveyard` are the `When` conditions. **Cards:** Treacherous Pit-Dweller, Flayer of the Hatebound, River Kelpie (Full).
