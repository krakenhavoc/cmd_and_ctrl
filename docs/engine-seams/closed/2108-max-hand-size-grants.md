---
title: "Maximum hand size changed by a spell, or for a while"
date: 2026-10-08
issues: [2108]
---
**Maximum hand size changed by a spell, or for a while** (#2108, [ADR 0113 amendment 2026-10-08](decisions/0113-small-seams-for-the-s58-deck-requests.md), CR 402.2 / 613.11 / 611.2) — `Game.GrantHandSizeForEffect` appends a `PlayerStatic` whose `HandSize` payload (`HandSizeGrant{Active, Kind, N, At}`) holds a no-maximum, set or modify entry with its CR 613.7b timestamp and an ADR 0063 `Duration`. `EffectiveMaxHandSizeLocked` tests the duration itself and folds every live grant with the player's single rest-of-the-game grant and the battlefield statics in timestamp order, so a second Inspired Idea reduces by three again. Additive snapshot data (`seats[].statics[].handSize`). Inspired Idea (cleave skips the reduction) and Enter the Infinite (no maximum until your next turn) ship `full`.
