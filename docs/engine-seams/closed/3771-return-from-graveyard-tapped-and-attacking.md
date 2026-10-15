---
title: "Returning a card from a graveyard tapped and attacking"
date: 2026-10-15
issues: [3771]
---
**Returning a card from a graveyard tapped and attacking** (#3771, [ADR 0137](decisions/0137-craft.md) amendment of 2026-10-15) — `putAttackingYourChoice` is no longer exile-only and private to the Grim Captain: it takes the zone the card must still be in, asks what it attacks (any player, planeswalker or battle its controller could attack, CR 506.3c) and moves it tapped and attacking, doing nothing when the card has since left that zone (CR 400.7). Interceptor, Shadow's Hound uses it behind `MayPay` from a graveyard-zone attack trigger. **Cards:** Interceptor, Shadow's Hound (Full).
