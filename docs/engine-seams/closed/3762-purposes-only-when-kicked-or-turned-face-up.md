---
title: "Purposes that only apply when kicked or turned face up"
date: 2026-10-15
issues: [3762]
---
**Purposes that only apply when kicked or turned face up** (#3762, found landing #3565; [ADR 0126](decisions/0126-bots-that-play-their-decks.md)'s 2026-10-15 amendment, [ADR 0188](decisions/0188-skipped-turns-and-turn-bound-effects.md) Decision 10) — `game.Purpose` gains `Kicked` and `FaceUp`, each a `*Purpose` on the card slot. The heuristic adds `purpose.kicked` to a cast whose `optional_costs` names the card's single kicker and to nothing else, and charges `purpose.turned_face_up` to the `turn_face_up` special action; a face-down permanent's controller is sent that field and no other (CR 708.2a). `effects.Register` refuses `Kicked` off a one-kicker card and `FaceUp` off a card with no face-down cast offer, and `Purpose.DeclaresSweep` lets `TestRealDumpPurposeAudit` accept a wipe declared this way. **Cards:** Waterspout Elemental moves its partial creature bounce and gains its skipped turn under `Kicked`, so an unkicked cast is no longer priced as a wipe; Unstable Hulk declares its skipped turn under `FaceUp`. **Not priced:** the Hulk's +6/+6 and trample until end of turn.
