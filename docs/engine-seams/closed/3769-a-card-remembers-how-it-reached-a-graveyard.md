---
title: "A card remembers how it reached a graveyard"
date: 2026-10-15
issues: [3769]
---
**A card remembers how it reached a graveyard** (#3769, [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md) 2026-10-15 amendment) — `Card.PutInGraveyardFrom` and `Card.PutInGraveyardThisTurn` record the zone a card was put into a graveyard from and that the arrival was this turn. `MoveCard` stamps them on every move and clears them on every move that does not end in a graveyard (CR 400.7), `onTurnBeganLocked` ends the "this turn" half, and `PermissionFilter.PutInGraveyardThisTurn` reads "put there from anywhere other than the battlefield this turn". Both card fields and the filter flag are additive snapshot fields. **Cards:** Banon, the Returners' Leader (Full).
