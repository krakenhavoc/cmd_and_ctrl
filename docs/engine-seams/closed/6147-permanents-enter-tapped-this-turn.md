---
title: "Permanents enter tapped this turn"
date: 2026-11-25
issues: [6147]
---
**Permanents enter tapped this turn** (#6147, CR 614.1d, 611.2, 514.2) — "Permanents enter tapped this turn" (Due Respect) and "Lands you control enter tapped this turn" (Nahiri's Lithoforming) are a CR 614.1d replacement a resolving spell creates: one `ScopedEffect` (`ScopeGame`, until end of turn) with a `ModPermanentsEnterTapped` mod (`permanentsEnterTapped`), written by `Game.PermanentsEnterTappedThisTurnForEffect(source, controller, types, label)`. `Mod.Types` narrows it to permanents with one of those lower-case card types (none is every permanent) and `Mod.Player` to permanents entering under that player's control (`uuid.Nil` is anyone's). The replacement gather reads it as #6093's `creaturesEnterTappedForOpponents` is read, live, as each permanent enters, so a land played or a creature cast later that turn enters tapped and one the next turn does not. A filter naming a non-permanent type is refused at registration and restore. **Cards** (1, Full): Nahiri's Lithoforming; Due Respect is built by its OLD batch.
