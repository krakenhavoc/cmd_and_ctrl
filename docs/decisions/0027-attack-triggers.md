# ADR 0027 — Attack triggers ride an event, not combat state

**Status:** Implemented · 2026-09-10 · Branch `feat/attack-triggers`

## Context

The [Aang decklist triage](../decklists/aang-is-so-flashy.md) groups
four blocked cards under "attack triggers" and calls it "the cheapest
item on this list by a wide margin". It also draws the distinction
that makes it cheap: combat **state** already exists. `DeclareAttacker`
stamps `Card.AttackingTarget`, `ClearCombat` wipes it, and Aetherize
shipped in batch 2 by reading that state at resolution without any
engine work.

What did not exist was a **moment**. `events.go` had no `EventAttack`,
so nothing could say "whenever ~ attacks". A trigger cannot be written
against state: by the time a trigger would look at the battlefield,
the declaration has already happened and there is nothing to
distinguish "this creature is attacking" from "this creature attacked
and I already fired". Sixteen event kinds covered casting, drawing,
zone moves, damage, sacrifice and upkeep; combat contributed only
`EventDealDamage` with `Combat: true` (S19 sub-PR 7), which is the
*end* of combat's story, not the beginning.

## Decisions

### 1. `EventAttack`, one event per attacking creature

```go
EventAttack EventKind = "attack"
```

Payload:

| Field | Meaning |
|---|---|
| `CardID` | the attacking creature |
| `Actor` | that creature's controller |
| `Target` | the player being attacked |

`CardID` rather than `Source` for the creature is deliberate and it
buys something concrete: `EventETB` already puts the entering
permanent in `CardID`, so a card printed "whenever this creature
enters **or attacks**" — Sun Titan, Korvold, Uro — is one
`TriggeredAbility` with `Watches: []EventKind{EventETB, EventAttack}`
and the bare predicate `ev.CardID == source.InstanceID`. Splitting the
creature across two fields would have forced a kind switch into every
such card.

**Fires per creature**, mirroring `EventDrawCard`'s "fires per card".
Three attackers produce three events, so Hellrider stacks three
separate pings with a response window between each — paper behaviour.
The known gap is the same one `EventETB` has: cards worded "whenever
one or more creatures you control attack" (Grand Warlord Radha,
Adeline) would over-fire, because CR 603.1 batching has no
representation anywhere in the engine. No catalog card has that
wording today.

**Fires only on a creature's first declaration.** The sandbox
deliberately lets a player re-point an already-attacking creature at a
different defender — paper does not — and re-pointing is a correction,
not a second attack. `DeclareAttacker` captures
`firstDeclaration := card.AttackingTarget == uuid.Nil` before stamping
and emits only then. Without that guard, a mis-click that gets fixed
would hand the table a second Hellrider ping.

**`Target` is always a player.** `DeclareAttacker` takes a player ID
and validates it against the seats; there is no attack-a-planeswalker
path in the engine at all. So Hellrider's "the player or planeswalker
it's attacking" collapses to the player — not a clause dropped, a case
that cannot yet arise. When planeswalker defenders land, `Target`
widens to "player or permanent" exactly as `EventDealDamage`'s already
is, and consumers that treat `Target` as an opaque ID keep working.

### 2. The declaration drains, so the trigger beats blockers

`DeclareAttacker` now calls `runStateChecksLocked()` after emitting.
Attackers are declared as a turn-based action and the active player
then receives priority (CR 508.2 / 117.5) — the boundary at which
SBAs run and `PendingTriggers` drains onto the stack. This is the same
fix ADR 0018 §3 applied to `MoveCardByID`, the land branch of
`CastSpell`, and `ResolveTriggerPrompt`.

It is not cosmetic. Without the drain the trigger sits in
`PendingTriggers` until the next priority pass, and if that pass is a
step advance with an empty stack, the trigger lands **after blockers
are declared**. "Whenever ~ attacks, destroy target artifact defending
player controls" that resolves after blocks is a different card.

The drain runs only on a first declaration — the same branch that
emits — so the blast radius is exactly the set of calls that already
changed observable state.

### 3. No new game state, so no `clone.go` change

The whole feature is one `EventKind` constant plus an emit site.
`Game.Events` is already a value slice copied wholesale by `Clone`,
and `TriggeredAbility` needed no new field: `Watches` / `AppliesTo` /
`Build` / `Targets` / `OptionalPrompt` cover attack triggers as they
stand. That is the load-bearing reason this was the cheap item — the
S19 dispatcher was general over event kinds from the start, and S20
gave it targeting.

## Cards

Three, chosen to cover the three shapes of the `AppliesTo` matrix
rather than for deck coverage:

- **Sun Titan** (`sun_titan.go`) — "whenever this creature enters or
  attacks", one ability watching two kinds, targeted + optional,
  reanimating with `ReturnFromGraveyard{Dest: ZoneBattlefield}`. From
  the Aang list. The triage listed its other half as blocked on
  "graveyard recursion"; the primitive turned out to already exist,
  and `TargetCardInGraveyard(YouOwn(), Permanent(), ManaValueLE(3))`
  expresses the clause exactly.
- **Hellrider** (`hellrider.go`) — "whenever a creature you control
  attacks", the card that reads the payload. Each ping follows *its
  own* attacker's defender (`ev.Target`, captured in `Build`), which
  in a four-player game is not the same as "an opponent".
- **Krenko, Tin Street Kingpin** (`krenko_tin_street_kingpin.go`) —
  "whenever THIS creature attacks", and the one where the printed
  "then" is load-bearing: the +1/+1 counter goes on before the token
  count is read, so a 1/2 Krenko makes two Goblins, not one.

### Simplifications declared

- **Hellrider** — "the player or planeswalker" is always the player
  (see §1); no planeswalker case exists to lose.
- **Krenko** — if Krenko has left the battlefield when the trigger
  resolves, it makes no tokens. Paper uses last-known information
  (CR 608.2h) and would still make Goblins equal to its last power.
  The engine's LKI snapshot (`lastKnownBattlefield`) is scoped to the
  dying card's own LTB triggers and is not reachable from a resolution
  callback.

## Out of scope (explicit deferrals)

- **Batched attacks** (CR 603.1) — "whenever one or more creatures you
  control attack" is one trigger in paper and N here. Shared with the
  existing `EventETB` gap.
- **Attacking a planeswalker.** Needs `DeclareAttacker` to accept a
  permanent as the defender, which is a combat-engine change, not an
  event change.
- **"Attacks a player" vs "attacks"** (Kaalia) is expressible today —
  `Target` names the seat — but no catalog card needs it yet.
- **The other three Aang attack-trigger cards** each carry a second,
  unrelated blocker:
  - *Katara, Waterbending Master* — experience counters; there is no
    player-scoped counter store.
  - *The Mighty Thor, Jane Foster* — flicker.
  - *Phelia, Exuberant Shepherd* — exile a permanent and return it at
    the beginning of the next end step: a delayed trigger plus a
    return-from-exile primitive, both landing on the flicker branch.
    Phelia's **trigger** half needs nothing further from this ADR —
    it is `Watches: []EventKind{EventAttack}` with
    `attackDeclared(ev, source)` and a
    `TargetPermanent("up to one other target nonland permanent",
    Not(Land())).WithCount(0, 1)` clause, all of which the S19/S20
    machinery already supports. Only its effect is blocked.

## Consequences

- `helpers.go` grows `attackDeclared` (this creature) and
  `attackDeclaredByYou` (a creature you control) so the two common
  predicates are written once and the payload contract is documented
  in one place.
- Engine coverage (`attack_event_test.go`): payload fields; one event
  per attacker and none for a creature that stayed home; the event
  fires in the declare-attackers step and combat produces no more of
  them; a rejected out-of-step declaration emits nothing; re-pointing
  retargets without re-firing; a stub catalog trigger reaches
  `StackMeta` during the declare-attackers step rather than being
  stranded in `PendingTriggers`.
- Catalog coverage (`attack_triggers_test.go`): Hellrider's
  per-attacker count, its wait-for-resolution timing, its
  per-defender routing across two seats, and an opponent's Hellrider
  staying quiet; Krenko's counter-then-count ordering and its
  self-only gate; Sun Titan's attack half, its enters half, its
  mana-value filter, and the CR 603.3d no-prompt-on-empty-graveyard
  case.
- No client change. `EventKind` is not projected onto the wire yet,
  and the trigger surfaces through the existing stack overlay and
  prompt modals.
