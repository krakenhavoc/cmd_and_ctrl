---
title: "Create and cast a copy of a card named from a fixed list"
date: 2026-11-25
issues: [6155]
---
**Create and cast a copy of a card named from a fixed list** (#6155, CR 707.13, CR 707.12, CR 707.10a) — `Game.CopyNamedCardAndOfferCastForEffect` (`game/named_card_copy.go`) creates a copy of a card defined by name. It is built from the card's printed record, the same `conjurable_cards.go` registry conjure reads, which is the engine's Oracle card reference. The copy is created outside the game, in the caster's `Player.Sideboard` container, and marked `CardCopy`. It is offered for a cast from there at its own costs, during the resolution, through the `resolution_cast` prompt: `ResolutionCastCandidateLocked` now also finds a card copy in the chooser's own outside container. A copy nobody casts is removed by the card-copy sweep, which now scans that container too. "A name that hasn't been chosen" is the per-object `Card.Notes` memory (#6097). The catalog side is `effects.CopyNamedCardAndOfferCast`, with `CheckNamedCardCopies` as the boot guard. **Cards:** Garth One-Eye (Full).
