---
title: "`ReturnFromGraveyard` has no \"enters tapped\" option"
date: 2026-09-23
issues: [1284]
pr: 1305
legacy_order: 22
---
**`ReturnFromGraveyard` has no "enters tapped" option** (#1284) — the last piece of "Ability activatable from a non-battlefield zone" that #1221's activated half left on the row for Reassembling Skeleton and Drownyard Temple, both of which print "Return this card from your graveyard to the battlefield tapped" behind the closed graveyard activation. `ReturnFromGraveyard.Tapped` is `SearchLibrary.TappedOnEntry`'s shape one primitive over: it threads through a new `Game.ReturnFromGraveyardTappedForEffect` to `returnFromGraveyardLocked`'s existing (previously unexposed) `tapped` parameter, which already stamped `EntersTapped` on the CR 614 entry event — the same field `ReturnToBattlefieldForEffect`'s exile-return path has used since #1178, just never reachable from the graveyard-under-control path. A new function rather than a fourth parameter on `ReturnFromGraveyardUnderControlForEffect`, which has call sites across the catalog and the engine's own tests all meaning "untapped": a bare trailing bool is the kind of change a diff reviews past. Both cards are `full`; the shared body (`returnThisFromGraveyardTapped`) lives in `reassembling_skeleton.go` and is called from `drownyard_temple.go`. **Still open on this row**: the per-instance exile GRANT (Greater Gargadon while suspended) named below has no shape yet.
