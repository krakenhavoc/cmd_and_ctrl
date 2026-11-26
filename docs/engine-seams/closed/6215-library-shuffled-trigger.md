---
title: "A trigger on a library being shuffled"
date: 2026-11-25
issues: [6215]
---
**A trigger on a library being shuffled** (#6215, CR 701.24a, 701.24e, 701.24f) — every shuffle of a library a spell or ability causes now goes through `Game.shuffleLibraryLocked` and emits `EventLibraryShuffled` (`library_shuffled`, Actor the library's owner): a search's shuffle, "shuffle your library" (`ShuffleLibraryForEffect`), and the shuffle a "can't search libraries" effect still leaves owed. Its `Label` is `game.ShuffleByEffect`; the sandbox's manual shuffle emits it with `game.ShuffleByAction`, and the opening-hand and mulligan shuffles emit nothing. A library of zero or one card is still shuffled and announced, and two libraries shuffled at once are two events. The table's log already narrates the shuffle through the search line, so the event is silent there. The trigger matchers are `effects.ASpellOrAbilityShufflesALibrary` (Psychogenic Probe: "that player" is `ev.Actor`) and `effects.AnOpponentShufflesTheirLibrary` (Psychic Surgery, Cosi's Trickster). **Cards**: Psychogenic Probe (OLD b606), Psychic Surgery (b463) and Cosi's Trickster (b493) are built by their OLD batches.
