---
title: "Enter as a copy of a card in a graveyard"
date: 2026-11-02
issues: [4797]
---
**Enter as a copy of a card in a graveyard** (#4797, [ADR 0043](decisions/0043-copy-effects.md)'s 2026-11-02 amendment, CR 707.2) — `ResolveCopyTarget` reads a copy_target pick from a graveyard when it is not on the battlefield (`copySourceCardLocked`), so a card's `Candidates` decide the zone: `effects.anyCreatureCardInAGraveyard` offers every creature card in every graveyard. The copied object is named before the except clause runs (`ReplacementEvent.CopySourceForEffect`), so an except clause can declare a "when you do" reflexive trigger with `QueueReflexiveTriggerAfterReplacedMoveForEffect`, which the entry finisher (`executeEntryToBattlefieldLocked`) now creates once the permanent has landed. **Cards:** Superior Spider-Man, Body Double, Voq, Torchbearer (Full). Activated Sleeper and Lazotep Convert need only their filter and except clause.
