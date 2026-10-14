---
title: "A revealed-hand exile that is linked to its source"
date: 2026-10-14
issues: [3516]
---
**A revealed-hand exile that is linked to its source** (#3516, [ADR 0116](decisions/0116-a-card-filter-on-the-revealed-hand-pick.md) amendment 2026-10-14, CR 607.2a) — "You choose a nonland card from it. Exile that card. … you may cast the exiled card" needs the exile to be "exiled with this", and the revealed-hand exile never used the link ADR 0155 §8 gave a resolving ability's exile. `ChooseFromRevealedHand{Exile: true, Linked: true}` now captures the resolving item's source OBJECT as the pick is raised (`RevealedHandDiscard.ExiledWith`, `PendingChoice.PickExiledWith`) and routes the chosen card through `ExileLinkedThenForEffect`, so `Card.ExiledWith` is stamped only if the card really reaches exile and a source that left or came back (CR 400.7) has no claim on it. The permission half was already there: a gated standing cast permission over exile (`CastPermissionGate`, `GraveyardAtLeast(7)`) filtered on `PermissionFilter.ExiledWithSource`, with `AnyType` mana. One additive field on `pendingChoices[]` under schema v8, with one fixture. **Cards** (1): Null Summoner (Full).
