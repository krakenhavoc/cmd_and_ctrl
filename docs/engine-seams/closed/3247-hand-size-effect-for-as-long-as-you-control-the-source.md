---
title: "A hand-size effect that a resolving ability creates for as long as you control its source"
date: 2026-10-13
issues: [3247]
---
**A hand-size effect that a resolving ability creates for as long as you control its source** (#3247, [ADR 0113](decisions/0113-small-seams-for-the-s58-deck-requests.md) amendment of 2026-10-13) — the engine already stored "no maximum hand size" or a number as a `PlayerStatic` grant with its own `Duration` (#2108), and `Duration` already had the source-bound `ForAsLongAs` kinds (#3211). The gap was the card: The Great Synthesis declared the clause as a static of the Saga, so a countered or stifled chapter I still lifted the maximum. Chapter I now creates the grant when it resolves, with `DurationWhileYouControlSource`, so a countered chapter leaves nothing, a Saga that left first starts nothing, and the effect ends when the Saga leaves, changes controller, or is flickered by chapter III. No new record, reader or wire field. **Cards:** The Great Synthesis (caveat removed, Full), Praetor's Counsel and Wisdom of Ages (Full).
