---
title: "A cast spell's mana value in the trigger snapshot counts X"
date: 2026-10-15
issues: [3782]
---
**A cast spell's mana value in the trigger snapshot counts X** (#3782, [ADR 0018](decisions/0018-triggers-on-the-stack.md) amendment 2026-10-15) — "that spell's mana value" for a spell cast with X is the X on the stack (CR 202.3e). `ObjectSnapshot.ManaValue` read `{X}` as zero; `objectSnapshotLocked` now takes the announced X for a spell on the stack, so a `TargetsFrom` that reads only its trigger context can cut a clause to it and stay restorable (no `TargetsFromReadsBoard`). Permanents and cards in other zones are unchanged. **Cards:** Krile Baldesion (Full): the clause is "creature card with that mana value in your graveyard", the "may" is asked at resolution, and "do this only once each turn" is recorded only when a card is returned (`NoteDoneThisTurn`).
