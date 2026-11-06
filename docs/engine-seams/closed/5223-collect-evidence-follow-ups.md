---
title: "Whenever you collect evidence, and collect evidence X"
date: 2026-11-06
issues: [5223]
---
**Whenever you collect evidence, and collect evidence X** (#5223, [ADR 0020 amendment 2026-11-06](decisions/0020-activated-abilities.md), CR 701.59a, 107.3f) — `game.EventCollectEvidence` is emitted once per collection by the cast path, the activation and the effect form (`Game.NoteCollectEvidenceForEffect`), collect evidence 0 included, carrying the exiled cards (`Event.CardIDs`, additive in v9) and their total mana value; `effects.WheneverYouCollectEvidence` watches it. Collect evidence X as an effect is `effects.RegisterCollectEvidenceX`: a keyed yes/no, any number of graveyard cards, then X up to their total, gated by `Game.CollectEvidenceXMetForEffect` (`FloorMet` with the floor at X). **Cards:** Evidence Examiner, Surveillance Monitor, Incinerator of the Guilty (all Full). Urgent Necropsy, whose X is the total mana value of its targets, still waits.
