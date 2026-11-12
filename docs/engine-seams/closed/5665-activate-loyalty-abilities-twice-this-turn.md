---
title: "Activate loyalty abilities twice this turn"
date: 2026-11-12
issues: [5665]
---
**Activate loyalty abilities twice this turn** (#5665, [ADR 0145](decisions/0145-meld.md) amendment 2026-11-12) — "You may activate the loyalty abilities of planeswalkers you control twice this turn rather than only once" (Urza Assembles the Titans' chapter III) is `Game.AllowLoyaltyTwiceThisTurnForEffect(player)`: Oath of Teferi's static as an effect that lasts the turn. `Game.LoyaltyTwiceThisTurn` is a set of players, read in `loyaltyLimitSpentLocked` beside the two printed statics, so CR 606.3's limit is two for every planeswalker that player controls, one that arrives later included, and an activation made before the effect counts as the first. It is a set, not a count: a second resolution leaves the limit at two, and neither static adds to it; The Chain Veil's grant is still one activation on top. `LoyaltySpentLocked` stays the one predicate the gate, the sandbox verb, the enumerator and the wire's `loyalty_activated` ask, so there is no wire change. Snapshot-additive at v9, carried by `Clone` and `RestoreFrom`, flushed at the turn boundary. **Cards:** Urza Assembles the Titans (Full).
