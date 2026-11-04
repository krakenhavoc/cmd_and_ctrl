---
title: "Pawprint modes"
date: 2026-11-02
issues: [4786]
---
**Pawprint modes** (#4786, [ADR 0065](decisions/0065-modal-and-multi-target-clauses.md)'s 2026-11-02 amendment, CR 700.2i) — "Choose up to five {P} worth of modes" is `ModeSpec.Pawprints` (the budget) over `ModeOption.Pawprints` (each bullet's printed {P}), built with `effects.ChooseUpToPawprints` and `effects.Pawprint`. `validateModes` refuses a selection whose {P}, a repeat counted again (CR 700.2d), exceeds the budget; `legal.modeSelections` prunes its walk by the same sum; `CardView.modes.pawprints` and `modes.options[i].pawprints` carry both numbers and the cast-time picker counts in {P}. Spells only; nothing is persisted. **Cards:** Hela's Arcanum, Season of Loss, Season of the Burrow (Full). Season of Gathering, Season of the Bold and Season of Weaving need nothing more from it.
