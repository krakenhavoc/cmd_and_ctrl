---
title: "A permanent you controlled left the battlefield this turn"
date: 2026-10-05
issues: [2148]
---
**A permanent you controlled left the battlefield this turn** (#2148, revolt, CR 702.136) — `PlayerTurnTally.PermanentsLeft` is bumped from every `EventLTB` under the controller the permanent had as it left (CR 603.10a), whatever the route or destination: a sacrifice, a bounce, an exile, a death, a token that ceased to exist, a permanent that left and came back. An opponent's permanent leaving does not count for you, and the cell resets with the rest of the turn tally. Read through `Game.PermanentLeftThisTurn`. Card side (`cards/effects/revolt.go`): `Revolt(g, you)` for a spell's resolution, `WhenThisEntersIfRevolt` and `AtYourEndStepIfRevolt` for the intervening if (checked as the trigger fires and again as it resolves, CR 603.4) and `SelfEntersWithCountersIfRevolt` for "enters with counters if". Cards: Shortcut to Mushrooms, Fatal Push, Renegade Rallier, Hidden Stockpile, Narnam Renegade, Silkweaver Elite, Airdrop Aeronauts and Countless Gears Renegade. The table-wide nonland sibling, void (#2128), still wants its own cell.
