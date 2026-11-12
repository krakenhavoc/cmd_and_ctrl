---
title: "Becomes the target of a backup ability"
date: 2026-11-12
issues: [5651]
---
**Becomes the target of a backup ability** (#5651, [ADR 0018 amendment 2026-11-12](decisions/0018-triggers-on-the-stack.md)) — a became-target event said the targeter was a triggered ability but not which one, and a backup ability (CR 702.165a) is an ordinary enters trigger on the stack. `Event.TargeterKeyword` is stamped by `emitBecameTargetLocked` beside `TargeterKind`: the `TriggeredAbility.Keyword` of the catalog row the targeting stack item names, read through its ref, so nothing new rides the item and a restored item or a copy (CR 707.10) answers the same. `effects.Backup` stamps `game.KeywordBackup`; backup has no keyword chip, so the row keeps its place on the tile. The copy is the existing `copyTheTriggeringAbility` and keeps the backup creature as its source (CR 707.10b), so it grants that creature's abilities; "only once each turn" is `TriggersOnlyOnceEachTurn`, which also declines the copy's own targeting. The event field is additive under snapshot v9. **Card:** Mirror-Shield Hoplite (Full).
