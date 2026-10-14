# ADR 0173 — Shield counters

Date: 2026-10-14
Status: Accepted
Sprint: S58 — Deck requests, October batch
Issues: #2000, #3246

## Context

CR 122.1c (September 25, 2026 edition): "One or more shield counters on a
permanent create a single replacement effect and a single prevention effect
that protect the permanent. These effects are 'If this permanent would be
destroyed as the result of an effect, instead remove a shield counter from
it' and 'If damage would be dealt to this permanent, prevent that damage and
remove a shield counter from it.'"

The printed reminder text ("If it would be dealt damage or destroyed, remove
a shield counter from it instead") is the older wording and is looser than
the rule in two ways that matter here:

- The replacement covers a destruction **as the result of an effect** only.
  CR 704.5g and 704.5h destroy by a state-based action, not by an effect.
- The damage half is a **prevention** effect, so it is subject to CR 615.12:
  under "can't be prevented" it still applies, prevents nothing, and its
  additional effect (removing a counter) still happens.

`game.CounterShield` existed only as a name. Shield Broker, Protection Magic
and Summon: Magus Sisters waited on it.

## Decision

1. **Two engine-owned built-ins** (`game/shield_counters.go`), registered in
   `NewGame` after the existing built-ins (appended, since a CR 616 prompt
   names a built-in by index). One registration each is what "a single
   replacement effect and a single prevention effect" means: two counters
   never ask an ordering question of each other. The counter is not an
   ability, so a permanent that loses all abilities keeps the protection.
2. **Destruction.** The replacement applies to a `RepEventMove` with
   `Destruction` set and the permanent holding a shield counter; it cancels
   the move and removes one counter. The state-based-action sweep destroys
   through a new `sbaDestroyRoute`, which sets
   `ReplacementEvent.StateBasedDestruction`, and the shield does not apply
   to that. Regeneration still replaces both, as before.
3. **Damage.** The prevention is a `Prevention: true` replacement whose
   `Replace` only prevents, with "remove a shield counter" as its ADR 0108
   §8 additional effect (`Then`, `ThenPerRecipient`). So damage dealt at the
   same time by several sources is all prevented for one counter, and
   unpreventable damage is dealt and still costs one counter (CR 615.12).
   A built-in has no object of its own, so the unexported
   `ReplacementEffect.thenOnDamaged` owes the follow-up on the damaged
   permanent, keyed on a slot (`builtinFollowUpSlot`, -1) that never merges
   with the permanent's own catalog prevention static.
4. **Ordering.** Neither built-in is preemptive. With a regeneration shield,
   a Phantom's own prevention or any other applicable effect, the affected
   permanent's controller chooses (CR 616.1).
5. **Card side** is only `AddCounter{Kind: game.CounterShield}`. "For as long
   as it has a shield counter on it" (Shield Broker) is the existing
   counter-held duration (ADR 0109 §2).

## Consequences

- Lethal unpreventable damage to a creature with two shield counters
  removes one counter and the creature still dies (the second counter does
  not stop a state-based destruction). That follows the current rule text,
  not the reminder text.
- A follow-up with slot -1 can appear in a restore point's owed list; it is
  plain data and needs no schema change.
- Indestructible is checked before the replacement window opens, so an
  indestructible permanent keeps its shield counter when an effect tries to
  destroy it.
