# ADR 0228 — What a spell would do, read by resolving it on a copy of the game

**Status:** Accepted · 2026-11-25 · S82 — Card coverage: sets outside Standard. Owner decision on [#6225](https://github.com/krakenhavoc/cmd_and_ctrl/issues/6225) (2026-11-24): build it, real printed cards must work, ship Full.
**Issue:** [#6225](https://github.com/krakenhavoc/cmd_and_ctrl/issues/6225) (Equinox).
**Numbering:** checked on 2026-11-25 after `git fetch origin`, with the loop in [README.md](README.md#picking-a-number) over `docs/decisions/` on all 636 remote branches. The highest number on any branch is 0227, so this is **0228**.
**Builds on:** the undo stack's `Game.Clone` (its contract: a clone's mutation cannot affect the receiver), [ADR 0126](0126-bots-that-play-their-decks.md) §6 (Purpose, which this does not use), ADR 0041 P3 (the closure ratchet) and #522 (snapshots additive within a version).

---

## Context

Equinox: "Enchanted land has '{T}: Counter target spell if it would destroy a land you control.'" The engine has to answer "would this spell destroy a land you control" for any spell in the catalog.

A spell's resolution is catalog Go code (`OnResolve`). The engine can't look inside it. The issue offered two ways:

1. Read the spell's declared Purpose (ADR 0126). Purpose is a bot hint, written by hand, undeclared on most cards, and has no "destroys lands" amount. ADR 0126 says the engine never reads it.
2. Add a "destroys lands" declaration to every spell.

Option 2 is wrong more often than it looks. A land is not only "target land" and "all lands": Murder destroys a Dryad Arbor or an animated Mishra's Factory, Shatter destroys an artifact land, Disenchant an Urza's Saga, Wrath of God an animated land. So the declaration would be on every destroy spell, about 700 instants and sorceries, each with a hand-written filter that has to match its `OnResolve` forever. A missing or stale declaration makes Equinox silently wrong.

The CR answer is simpler: "would destroy" is what the spell does if it resolves now (CR 701.8a: "To destroy a permanent, move it from the battlefield to its owner's graveyard").

## Decision

**The engine resolves the spell on a throwaway copy of the game and reads what that copy destroyed.** `Game.SpellWouldDestroyALandOfForEffect(spell, player)` (`game/land_destruction.go`):

1. Clones the game with `cloneLocked`, the undo stack's deep copy. Its contract is that nothing done to the clone reaches the live game, randomness included (the clone has its own copy of the RNG counters).
2. Gives the clone no listeners, so no subscriber, broadcast or bot hears it, and marks it `simulating`, so an effect error on it is not counted in the server's metrics or log.
3. Moves the spell's card to the top of the clone's stack and stamps its item after every ability, then calls `resolveTopOfStackLocked` once, under a `recover`.
4. Reads the clone's events from that point: every card that left the battlefield with `LeaveCause` "destroy" was destroyed by the spell. The live game's board then says which of them are lands `player` controls.

This gets the rules right with nothing declared on any card:

- A target that is no longer legal is not destroyed (CR 608.2b), so a Stone Rain whose land has gone is not countered.
- An indestructible land (CR 702.12b) and a regenerated one (CR 701.19a) are not destroyed on the clone either.
- A sacrifice (CR 701.21a), an exile and a state-based death after the spell (CR 704.5g) are not "destroy", so they don't count. The clone stops when the resolution function returns, before the state-based actions.
- A text-changed spell (ADR 0160) is resolved with its changed text.

**What it does not play out.** A resolution that pauses for a player's choice stops on the clone at the first question. What the spell destroys before that question counts, and what it would destroy after it does not. No catalog land-destruction spell asks before destroying. A card whose destruction depends on a choice in its resolution and that needs this question answered gets its own decision.

**The field.** `Game.simulating` is an unexported bool, set only on the copy. It is never cloned or persisted (`snapshot_drift_test.go` classifies it as dropped), and it holds no func, so the closure census does not move.

**Other questions.** A later "if it would <verb>" condition reads the same clone through `spellWouldDestroyLocked` or a sibling beside it, rather than adding a declaration per card.

## Consequences

- Equinox (OLD b869) ships Full on this seam. Its batch builds the card: an Aura granting the enchanted land `{T}` with `effects.TargetSpell("target spell")` and `effects.CounterTargetSpellIfItWouldDestroyALandYouControl`.
- A question costs one clone of the game per resolution of the ability. That is the same copy every undo point takes, so it is cheap at our table sizes.
- Engine tests (`cards/effects/tests/u/would_destroy_land_test.go`): Stone Rain at the enchanted player's land is countered, Shock is not, Stone Rain at another player's land is not, Armageddon and Murder at a land creature are, and asking changes nothing in the live game (stack, board, log).

## Rejected

- **A declaration on every spell.** It would have to cover every destroy spell, not only land destruction, and it drifts from the code.
- **Purpose.** A bot hint, often undeclared, and ADR 0126 keeps the engine from reading it.
