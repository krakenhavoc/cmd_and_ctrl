---
title: "A choice among several payments as a spell resolves"
date: 2026-10-09
issues: [2854]
---
**A choice among several payments as a spell resolves** (#2854) — an `option_pick` option can carry a mana cost (`ChoiceOption.ManaCost`, wire `pick_options[].mana_cost`), so "may pay {1} or {2}" is one choice with three outcomes, as printed (CR 118.12, 118.12a). An option the chooser can't pay is dropped when the prompt is queued (CR 118.3), and the free option always stays. The chosen option is paid through the auto-tapper, as a pay-unless is, and refused with the prompt still open if the board has moved since the question. A keyed continuation (`effects.OptionPickThen`, riding `pickThen` with `optionCarry`) keeps a table waiting on the answer a restore point. The legal-move list re-checks each costed option and prices it as `MoveCost.Mana`, the heuristic pays the cheapest one, and the client shows each option's cost. **Cards:** Winter's Chill, Lim-Dûl's Hex and Thrull Wizard (Full).
