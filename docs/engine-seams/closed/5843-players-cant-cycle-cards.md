---
title: "Players can't cycle cards"
date: 2026-11-19
issues: [5843]
---
**Players can't cycle cards** (#5843, CR 702.29a / 702.29f / 602.5 / 101.2) — `game.ActivationAbility` gains a `Cycling` bit, derived from `ActivatedAbilityShape.Cycling` by `ActivationAbilityOf`, so a board-wide activation restriction can name the action. `effects.PlayersCantCycle` is that restriction: it refuses every cycling ability, typecycling included (CR 702.29f), for every player while its source is on the battlefield, and nothing is stored, so it lifts when the source leaves. It goes through `ActivationGateLocked`, so the activation, the legal-move enumerator and the view's `cant_activate` agree; the view now builds its query with `ActivationAbilityOf` too, rather than a literal that carried only the label and the power-up bit. No snapshot change. **Cards:** Stabilizer (Full).
