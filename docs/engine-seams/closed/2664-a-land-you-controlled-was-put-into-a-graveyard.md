---
title: "A land you controlled was put into a graveyard this turn"
date: 2026-10-08
issues: [2664]
---
**A land you controlled was put into a graveyard this turn** (#2664, [ADR 0049](decisions/0049-card-engine-seam-review.md) 2026-10-08 amendment) — `PlayerTurnTally.LandsToGraveyard` is bumped from the `EventLTB` branch that already feeds `CreaturesDied`: the exit's destination is a graveyard and the permanent was a land as it last existed (CR 608.2h), under the controller the exit stamped (CR 603.10a). A bounce or an exile does not count, nor does a nonland permanent or an opponent's land; the cell resets with the rest of the turn tally and is an additive field in schema 7. Read through `Game.LandToGraveyardThisTurn`. **Cards:** The Lady of Otaria (Full), whose end-step trigger is an intervening "if" checked as the step begins and again as it resolves (CR 603.4).
