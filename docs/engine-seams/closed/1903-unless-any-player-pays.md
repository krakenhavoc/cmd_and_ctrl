---
title: "Unless any player pays"
date: 2026-10-13
issues: [1903]
---
**Unless any player pays** (#1903, CR 101.4, CR 118.12, CR 119.4, CR 119.8, [ADR 0150](decisions/0150-other-players-pay-any-player-offers-and-assist.md) PRs 1–2), in `game/any_player_pays.go` and `game/pay_unless_action.go`.
- **The offer.** `QueueAnyPlayerPaysForEffect` asks each player still in the game for one ordinary pay-unless at a time, in APNAP order from the active player. The first payer ends the offer, and the effect happens if everyone declines. Stopping at the first payer is a recorded simplification of CR 101.4; on every card in scope a second payment buys nothing. Each link holds the step, can guard a spell, and takes no standing answer. `PendingChoice.AnyPlayerOffer` drives the asked player's header and a read-only dock line for everyone else. `HelpsController` covers offers whose payment keeps the controller's own permanent (Icy Prison).
- **Payments.** Mana (#3203), plus life, damage and nothing (#3236). The existing sacrifice and discard payments also go through the chain, and `PerPayer` prices each player separately (Temporal Extortion).
- **Bot.** A bot never pays against its own effect. Against an opponent's effect it pays like any tax, without waiting for later opponents to pay.
- **Cards:** 21. The 7 mana offers came first (Rhystic Circle, Rhystic Tutor and others), then 14 more (Browbeat, Phantasmagorian and others). Still waiting: Cleansing (#3205) and Carrion Rats, Carrion Wurm and Worms of the Earth (#3206). Rhystic Cave and Leonin Arbiter are out of scope.
