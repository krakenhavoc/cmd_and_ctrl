---
title: "Returning a card to the battlefield transformed"
date: 2026-10-08
issues: [1900]
pr: 2660
---

**Returning a card to the battlefield transformed** (#1900, CR 712.14a / 701.27a, [ADR 0079 amendment of 2026-10-08](decisions/0079-transforming-a-permanent.md#amendment-2026-10-08-1900-returning-a-card-from-a-graveyard-transformed)) — `Game.ReturnFromGraveyardTransformedForEffect` sets the back face on a double-faced card in a graveyard, then runs the ordinary graveyard entry, so the CR 614 pipeline finds the back face's self-replacements and the card enters tapped and transformed with any "enters with" counters. The card keeps its InstanceID and gets a fresh entry timestamp. A card with no permanent other face stays in the graveyard. The primitive is `ReturnFromGraveyard{Transformed: true}`; the three Ojer gods share `ojerDiesReturnTransformed`. **Cards:** Ojer Pakpatiq, Deepest Epoch // Temple of Cyclical Time and Ojer Kaslem, Deepest Growth // Temple of Cultivation, both Full; Ojer Axonil, Deepest Might gains its dies trigger and keeps one caveat (Temple of Power's transform-back needs a noncombat-damage-by-colour tally).
