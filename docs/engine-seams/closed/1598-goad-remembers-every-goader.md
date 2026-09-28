---
title: "Goad remembers every goader"
date: 2026-09-28
issues: [1598]
---
**Goad remembers every goader** (#1598, CR 701.15a / 701.15c, [ADR 0045 amendment 2026-09-28 (#1598)](decisions/0045-combat-restrictions.md) Decision 53). A creature goaded by several players carries every goad. `Card.Goads` has one entry per goader, and each entry ends as its own goader's next turn begins. The stamp is `Player.TurnsBegun + 1`, the same one `UntilYourNextTurn` uses, and `sweepExpiredGoadsLocked` runs at turn start. A re-goad refreshes the goader's entry. `attackRequirementsOfLocked` adds goad's pair per entry, so CR 508.1d's counting makes the creature attack a player who goaded it neither time when it can, and any goader when every opponent goaded it. The snapshot adds `goads` beside `goadedBy`, which now holds the latest goader, and a pre-#1598 `goadedBy`-only file restores as a one-goader set. The wire adds `CardView.goaders` beside the unchanged `goaded_by`. **Alela, Cunning Conqueror** loses her last caveat and ships `full`; this retires the "one goader per creature" caveat that #1571's entry left.
