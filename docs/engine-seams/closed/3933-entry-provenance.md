---
title: "Which ability put a permanent onto the battlefield"
date: 2026-10-16
issues: [3933]
---
**Which ability put a permanent onto the battlefield** (#3933, CR 603.4, [ADR 0013](decisions/0013-replacement-effects.md) §5ak) — Kodama of the East Tree's "Whenever another permanent you control enters, if it wasn't put onto the battlefield with this ability, …" needs the entry to name the ability that made it, and no entry event did, so one Kodama would chain off its own puts. Every entry's `EventETB` is now stamped, at the one place entries announce themselves, from the item resolving while it happens: `CauseSource` (the item's source object, CR 400.7) and `CauseAbility` (its catalog row). Only while the resolution is open, and never on a spell resolving into a permanent, so a land played, a spell resolving and a sandbox move name no ability.
- **The condition.** `TriggeredAbility.NotPutByThisAbility`: an entry whose cause is this trigger's own source object and row does not trigger it. A second Kodama's put triggers the first; a Kodama that left and came back is a new ability. The cause is fixed with the event, so the CR 603.4 re-check as it resolves agrees. `effects.Register` refuses the flag off an `EventETB` row with a declared Effect.
- **Cards** (1, Full): Kodama of the East Tree.
