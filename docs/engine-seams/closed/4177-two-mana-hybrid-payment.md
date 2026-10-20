---
title: "Paying a two-mana hybrid symbol with two generic"
date: 2026-10-19
issues: [4177]
---
**Paying a two-mana hybrid symbol with two generic** (#4177, [ADR 0040](decisions/0040-mana-pipeline.md) amendment of 2026-10-19) — `{2/B}` is one {B} or two mana of any kind (CR 107.4e, 601.2b), and the parser, the mana value and the cost splice read it from the start, but the pool solver and the auto-tapper booked the colour half only. The pool solver now walks the single-colour symbols first, then the hybrids (`requirementRuns`, `widenedLast`), and `foldUnpaidTwobrid` turns a hybrid no mana of its colour paid into its generic (`allocateRest`, `MissingFor`); `solveColored` tries the coloured slots and, last, defers the hybrid to the generic recruit. The colour is taken whenever it can be had. **Cards:** Beseech the Queen, now Full; Karn, Gilded Guardian, Spectral Procession and the other `{2/C}` cards are payable without their colours.
