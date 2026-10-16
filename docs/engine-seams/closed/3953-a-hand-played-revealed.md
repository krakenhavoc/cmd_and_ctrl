---
title: "A hand played revealed"
date: 2026-10-16
issues: [3953]
---
**A hand played revealed** (#3953, CR 400.2, 701.20a, 611.2b, 611.3a, [ADR 0199](decisions/0199-a-hand-played-revealed.md)) — a player who plays with their hand revealed is a per-player state, never stamped on `Card.KnownBy`. `Game.HandRevealedLocked` (`game/hand_revealed.go`) derives it from the battlefield (`Spec.HandsRevealed`: `HandRevealOpponents` or `HandRevealEachPlayer`, read off the ability key) and from a granted `PlayerStatic.HandRevealed` with a CR 611.2 duration (`GrantHandRevealedForEffect`), tested live. The view's `stampHandRevealed` makes every card in such a hand known to every seat before the per-viewer redaction, so a card drawn later is shown and the hand is hidden again the moment the effect ends. `PlayerView.hand_revealed` names the sources, shown as a HAND OPEN seat badge. Snapshot: additive. **Cards:** Stromgald Spy, Telepathy and Revelation (all Full).
