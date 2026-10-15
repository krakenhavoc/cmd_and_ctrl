---
title: "Worms of the Earth"
date: 2026-10-15
issues: [3576]
---
**Worms of the Earth** (#3576, [ADR 0187](decisions/0187-any-player-pays-per-object-graveyard-and-either-payments.md) slice 8) closes the "Unless any player pays" row (#1903): no card waits on it now.
- **The card.** `cards/effects/worms_of_the_earth.go` combines the land-play prohibition, the #1985 "lands can't enter" restriction and an each-upkeep `AnyPlayerPays` whose payment is `EitherPayment(SacrificePayment(2, …), DamagePayment(5))`. The first payer destroys it, and with no payer it stays.
- **Cards:** 1. Rhystic Cave and Leonin Arbiter stay out of scope (ADR 0150).
