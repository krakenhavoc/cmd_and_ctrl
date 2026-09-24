---
title: "Every replaceable event kind is switched on"
date: 2026-09-19
issues: [982]
pr: 1003
legacy_order: 58
---
**Every replaceable event kind is switched on** (#982): a `ReplacementEventKind` has obligations at five switches and no compiler to enforce any of them — `eventKindMatches` (the watch key), `affectedPlayerForEvent` (the CR 616.1 chooser), `applyResolvedReplacementEventLocked` (the resume), `finishSettledReplacementLocked` (the cancelled-event outcome) and `abandonZoneRouteLocked` (the abandoned-prompt outcome). A missing arm is silently WRONG rather than a build error: the chooser switch had no case for `RepEventDiscard` or `RepEventCreateTokens`, so both fell through to the first gathered effect's controller. `TestEveryReplacementEventKindIsSwitchedOn` reads the kinds and the switches out of the source, the way `TestEveryChoiceKindIsClassifiedAndEnumerated` does for `PendingChoiceKind`, and found a second gap on its first run — an abandoned `RepEventKeywordAction` dropped its continuation. Not a seam a card was waiting on; listed here because the next kind is.
