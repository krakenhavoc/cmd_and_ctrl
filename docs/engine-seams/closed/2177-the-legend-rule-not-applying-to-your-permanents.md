---
title: "The legend rule not applying to your permanents"
date: 2026-10-05
issues: [2177]
---
**The legend rule not applying to your permanents** (#2177, CR 704.5j) — `Spec.LegendRuleExemptions` declares a printed "the legend rule doesn't apply" static, player-scoped (`LegendRuleDoesntApplyToYours()`) or global (`LegendRuleDoesntApply()`). `legendRuleChoicesLocked` reads the exemptions off the battlefield on every state-based action check, keyed by `CatalogAbilityKey` and never stored, so one leaving cannot revoke another's and the next check after the last one leaves applies the rule with the controller choosing. A copy effect's "except it has" clause carries it as `AbilityGrant.LegendRuleExempt`. **Cards:** Mirror Box (loses its caveat), Mirror Gallery (Full) and Sakashima of a Thousand Faces (Caveats: Partner is deck construction and does nothing).
