# ADR 0027 — Attack triggers ride an event, not combat state

**Status:** Implemented · 2026-09-10 · Branch `feat/attack-triggers`
**Amendment:** 2026-09-28 · **Accepted** · [Combat state rides the leaves-the-battlefield event (CR 603.10a)](#amendment-2026-09-28-1661-combat-state-rides-the-leaves-the-battlefield-event) · [#1661](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1661)

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
Adeline) would over-fire, because printed "one or more" batching has no
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
  callback. **Closed 2026-09-24 (#1432, [ADR 0018](0018-triggers-on-the-stack.md)
  Decision 17):** a Krenko that is not the attacker any more — removed,
  or back as a new object — takes no counter, and the Goblins are its
  last-known power through `Context.SourcePermanent()` (#1379, #1418).
  Krenko ships `full`.

## Out of scope (explicit deferrals)

- **Batched attacks** — "whenever one or more creatures you
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

## Amendment (2026-09-19, #373): all three deferred attack-trigger cards are unblocked

The *Out of scope* list closed with three Aang-deck cards, each parked
on a second blocker that had nothing to do with attack triggers. All
three of those blockers have since been built, and the list has been
a stale note for long enough that a player filed the card as a bug:
[#373](https://github.com/krakenhavoc/cmd_and_ctrl/issues/373) reported
The Mighty Thor, Jane Foster "not triggering", which is what an
uncatalogued card does.

| Card | Recorded blocker | Where it landed |
|---|---|---|
| *The Mighty Thor, Jane Foster* | flicker | [flicker.go](../../server/internal/cards/effects/flicker.go) and `ReturnFromExile{Tapped: true}` ([primitives.go](../../server/internal/cards/effects/primitives.go)) — the exact shape the card's second half asks for |
| *Phelia, Exuberant Shepherd* | a delayed trigger plus a return-from-exile primitive | `game.DelayedTrigger` ([delayed.go](../../server/internal/game/delayed.go)) and the same `ReturnFromExile` |
| *Katara, Waterbending Master* | "there is no player-scoped counter store" | `Player.Counters` with `CounterExperience` ([counter_types.go](../../server/internal/game/counter_types.go)) |

None of the three is written; all three are now ordinary card work
rather than blocked card work. This is the failure mode AGENTS.md §7
warns about one level down, where a *caveat* goes stale the day
someone else implements the mechanic: a deferral that was true the day
it was written and is load-bearing and false a month later. The
correction is recorded here rather than made by deleting the sentence,
because the sentence is why nobody picked the cards up.

Nothing about the decision itself changes: `EventAttack`,
`attackDeclared` / `attackDeclaredByYou` and the CR 603.3d timing are
as shipped, and the other deferrals in that section (batched attacks,
attacking a planeswalker) still stand.

## Amendment (2026-09-28, #1661): combat state rides the leaves-the-battlefield event

**Status:** Accepted · S37 — Combat correctness · [#1661](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1661)

### Context

This ADR's title is the decision it keeps making: a trigger is written
against an **event**, not against combat state, because by the time a
trigger looks at the board the state has moved on. The one combat
trigger family it did not reach was the one where that is literally
true — "whenever an **attacking** creature dies" (Kardur,
Doomscourge), "whenever a **blocking** creature an opponent controls
dies" (Death Tyrant), Garna's "draw a card if it **was attacking**".

A permanent leaving the battlefield is removed from combat (CR 506.4).
The engine does that in the exit itself: `MoveCard` clears
`Card.AttackingTarget` and `Card.BlockingTarget`, and
`forgetPerObjectTurnStateLocked` drops the blocked record, all before
`EventLTB` is emitted. The CR 603.10 snapshot the exit takes
(`lastKnownBattlefield`) is a `Characteristic` — no combat state — and
it is keyed to the dying card's **own** leaves-the-battlefield
triggers. So a third party's trigger condition had nothing to read, and
Kardur shipped with his drain caveated (#1660). Garna had been working
around it with an event-log walk (an `EventAttack` naming the creature
earlier in the same combat), which was wrong in both directions: a
creature put onto the battlefield attacking was never declared and read
as not attacking, and one removed from combat by a control change still
read as attacking.

CR 603.10a says leaves-the-battlefield abilities "look back in time":
every such ability, whoever controls it, judges the event against the
game as it was immediately before.

### Decision 1 — Three fields on `Event`, stamped on `EventLTB` only

```go
AttackingTarget uuid.UUID // what it was attacking — Card.AttackingTarget's domain
BlockingTarget  uuid.UUID // the attacker it was blocking
Blocked         bool      // it was a blocked attacker (CR 509.1h)
```

The same names as the `Card` fields they preserve, so a reader who
knows one knows the other. All three are zero for a permanent that was
not in combat, or had already been removed from it.

**Why the event rather than a second snapshot.** Every watcher already
receives the event; a snapshot map keyed by card would need a reader
API, a lifetime (the harvest deletes `lastKnownBattlefield` the moment
its dispatch ends), and an entry in Clone and the persisted snapshot.
The event log is already all three: `Clone` shares it, the snapshot
serialises it, and an undo rewinds it. So this adds **no game state**,
in the same sense §3 above meant it. It is also the answer that cannot
drift from the moment it describes, which is what "look back in time"
asks for.

**Why not `PermanentInfo` (ADR 0018's CR 608.2h record).** That record
answers a *resolving* ability's question about an object that has left.
No catalog card asks a combat question at resolution today; a trigger
condition, which is what every card on this list is, runs at the event.
If one does, adding the same three facts to `PermanentInfo` from the
same `combatLKI` value is a two-line change.

### Decision 2 — One reader, in the one exit step

`battlefieldExitLocked` already runs for every battlefield exit (the
destroy / sacrifice route, the general effect route and the sandbox
move). It now reads the combat state **first** — before the snapshot,
before the blocked record is forgotten, before `MoveCard` — and returns
it as a `combatLKI` value. Each of the three `EventLTB` emit sites
stamps it. One reader means combat damage, a removal spell in the
declare blockers step, a sacrifice mid-combat, a bounce and a sandbox
drag all report the same facts; the per-route test table
(`combat_lki_test.go`) pins every route.

A creature removed from combat before it left — a control change, the
engine's CR 506.4 route (`removeFromCombatLocked`) — reports nothing,
because that function has already cleared all three. That is the rule,
not a gap: it is no longer an attacking creature.

`Blocked` is read only for a creature that is attacking, and
`removeFromCombatLocked` drops the row with the attack, so the record
cannot outlive the attack it describes.

### Decision 3 — Card-side readers

`diedWhileAttacking(ev, g)` and `diedWhileBlocking(ev, g)` in
`effects/helpers.go`, beside `diedCreature`. Each returns the dead card
as it now sits in the graveyard (for "you control" / "an opponent
controls" and for "return that card") and `ok` when the event says it
was attacking / blocking. Neither checks `IsCreature` on the graveyard
card: only a creature attacks or blocks, and a crewed Vehicle that died
attacking is an artifact again by the time it is in the graveyard. The
event-picker table in AGENTS.md §7 has the row.

### Cards

- **Kardur, Doomscourge** — the drain ships; `full`.
- **Garna, Bloodfist of Keld** — moved onto `diedWhileAttacking`; the
  log walk (`b35WasAttackingWhenItLeft`) is deleted.
- **Death Tyrant** (new) — both clauses, crossed correctly (your
  blocker and their attacker are not it).
- **Ares, God of War** (new) — "return that card" guarded by the
  graveyard object epoch, Edea's shape.

### Still open

- "Attacking or blocking **alone**" (Thijarian Witness) needs the
  combat's other participants at the moment of death, which a per-card
  fact cannot carry.
- "A **goaded** attacking or blocking creature" (Baeloth Barrityl)
  needs goad as a static, continuous effect.
- Zurgo Stormrender's leaves-the-battlefield half is now expressible;
  the card waits on mobilize.
