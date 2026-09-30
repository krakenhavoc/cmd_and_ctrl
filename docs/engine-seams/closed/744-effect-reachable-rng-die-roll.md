---
title: "Effect-reachable RNG / die roll"
date: 2026-09-24
issues: [744]
pr: 1402
legacy_order: 4
---
**Effect-reachable RNG / die roll** (#744, [ADR 0054](decisions/0054-dice-rolls-and-coin-flips.md)) — closed on the 2026-09-24 re-triage (#1386): the row was fully stale. Its text, as it stood:

  > **What was missing:** **Core delivered for #744** ([ADR 0054](decisions/0054-dice-rolls-and-coin-flips.md)): keyed, rewindable `RollDiceForEffect`, `FlipCoinsForEffect`, `ChooseAtRandomForEffect`, and resumable called flips with heads/tails/stop; per-instruction public logs, client cues, legal moves and bot policy. Stable source ordinals survive clone and snapshot. Eight representative cards are registered; Urza's Bauble now uses the keyed stream. Remaining work is the additional card wave and its card-specific dependencies, not the basic random-effect APIs.
  >
  > **Cards waiting:** Remaining: Wyll's Reversal (spell retargeting); Clown Car (ETB X provenance); Ancient Silver Dragon (permanent hand-size grant); Goblin Archaeologist; Fiery Gambit; Game of Chaos (continuation and life-change ordering); Yusri (temporary free-cast permission)
