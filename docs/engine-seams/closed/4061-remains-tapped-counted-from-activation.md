---
title: "\"For as long as this remains tapped\" counted from activation"
date: 2026-10-19
issues: [4061]
---
**"For as long as this remains tapped" counted from activation** (#4061, [ADR 0109](decisions/0109-rule-gates-land-types-mana-and-cost-components.md) amendment 2026-10-19, CR 611.2b) — Preacher's ruling reads the duration from the moment the ability is activated: a source untapped while the ability waited and tapped again by resolution has broken it. `untapPermanentLocked`, the one untap path, marks every ability of that permanent on the stack (`StackItem.SourceUntapped`, additive in v9), and `DurationWhileThisRemainsTapped`, `WhileYouControlThisAndItRemainsTapped` and Old Man of the Sea refuse on it, so the whole family (Vedalken Shackles, Seasinger, Rubinia Soulsinger, Willow Satyr, Hivis of the Scale, Helm of Possession, the Couriers, Endoskeleton, and the tap-and-hold cards) is held to the stricter reading. **Cards:** Preacher (Full).
