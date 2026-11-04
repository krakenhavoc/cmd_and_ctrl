---
title: "Card types among spells you've cast this turn"
date: 2026-11-03
issues: [4677]
---
**Card types among spells you've cast this turn** (#4677, CR 205.2a, CR 601.2i, CR 707.10) — "draw a card for each card type among spells you've cast this turn" (April O'Neil, Hacktivist). `CastTally.CardTypes` is the set of CR 205.2a card types of every spell a player cast this turn, written beside `CastTally.Colors` as each spell becomes cast and read off the spell on the stack, so a spell that has since left every zone still counts and an MDFC cast as its back face, an Adventure or one half of a split card adds only the types it had as it was cast. Kindred is a card type of its own, so a kindred instant adds two; a supertype adds none. A copy is not cast and a land played is not a spell (CR 305.1), so neither adds anything. `CastTally.CardTypesCast()` is the reader. The snapshot gains `spellsCastThisTurn{}.cardTypes`, additive under v9. **Cards:** April O'Neil, Hacktivist (Full).
