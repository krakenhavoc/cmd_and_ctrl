---
title: "Variable-count tap-others costs"
date: 2026-09-24
issues: [1421]
pr: 1470
---
**Variable-count tap-others costs** (#1421, CR 107.3 / 118.3 / 602.2b, [ADR 0073 amendment 2026-09-24](decisions/0073-optional-additional-costs-and-the-cast-gate.md) Decisions 8–10) — `TapOthersCost.Filter.CountFromX` makes "Tap X untapped … you control" an announce-time cost: the shared bounds helper turns X into the exact payment width, `AbilityCost.DemandsX` exposes the prompt even without `{X}` mana, and the validator refuses a different number of `tap_ids` before anything taps. The wire reuses `LegalTargetsView.count_from_x`; the client picker supplies both the IDs and their count as `x_value`; bots offer the first three useful nested payments. Registration refuses a second claimant on the same X and refuses the shape on a mana ability, which has no X announcement (CR 605.3b). **Secluded Starforge** and **Apothecary White** now ship `full` as the two proof cards.
