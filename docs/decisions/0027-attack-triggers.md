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

## Amendment (2026-09-28, #1675): the leaving permanent's card types ride the same event

**Status:** Accepted · S37 — Combat correctness · [#1675](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1675)

### Context

The #1661 amendment above gave a third party's trigger condition the
combat state a leaving permanent had. Its builder noticed the same hole
one field over: `diedCreature(ev, g)` — the reader behind Blood Artist,
Zulaport Cutthroat, Grave Pact, Midnight Reaper and about thirty other
"whenever a creature dies" cards — tested `IsCreature()` on the card
**as it sits in the graveyard**. In the graveyard no continuous effect
applies to the card any more, so a permanent that was a creature only
because of an effect — a crewed Vehicle, an animated manland, a Gideon
on his own turn, an artifact Tezzeret animated — is an artifact, a land
or a planeswalker again, and none of those triggers fired. The reverse
was wrong too: a printed creature that an effect had made a noncreature
still read as a creature once it died. `diedCreature`'s own comment
called this a sandbox gap, because the CR 603.10 snapshot
(`lastKnownBattlefield`) is keyed to the dying card's own triggers and
is deleted when the harvest for the event ends.

CR 603.10a is the same rule as before: a leaves-the-battlefield ability
looks back in time, so "whenever a creature dies" asks what the
permanent was as it last existed on the battlefield.

### Decision 1 — `Event.LastKnownTypes`, stamped with the combat state

```go
LastKnownTypes []string // post-layer Characteristic.Types as it left — EventLTB only
```

`battlefieldExitLocked`'s look-back is now `exitLKI` (was `combatLKI`):
the combat state plus a copy of `c.Effective().Types`, read with the
card still on the battlefield — the post-layer value
`snapshotLKILocked` writes for the dying card's own triggers, not the
printed type line. The three `EventLTB` emit sites already call
`stamp`; it now writes the types too, so every route (destroy,
sacrifice, the zone route's bounce / exile / sandbox move, and the
sandbox drag to the stack) carries them with no new call site.

Card **types** only. Subtypes and supertypes are the same question
("whenever a Zombie you control dies" missing a changeling-granted
Zombie) and would ride the same field shape, but no issue asks for them
yet and each is a reader change as well as a stamp; they are left open
below rather than stamped for nobody.

Same reasoning as Decision 1 above for the event rather than a map: the
event log is already shared by `Clone`, serialised by the snapshot and
rewound by an undo, so this adds no game state (the v7 shape file
records the additive key). `cloneTriggerContext` copies the slice as it
does `Event.Colors`.

### Decision 2 — `Event.WasType` and `leftAsType`

`Event.WasType(cardType) (was, known bool)` answers the question
case-insensitively and says whether the event carries the types at all.
`known` is false for every kind but `EventLTB`, and for an `EventLTB`
that predates the field (a restored snapshot's log) or was built by
hand in a test.

Card side, `leftAsType(ev, c, cardType)` in `effects/helpers.go` reads
`WasType` and falls back to the card as it sits only when the event
does not know — the pre-#1675 reading, never a false "no". Every
dies / leaves-the-battlefield condition that tests a card type now asks
it:

| Reader | File | Cards |
|---|---|---|
| `diedCreature` | helpers.go | every "whenever a creature dies" |
| `creatureYouControlLeftWithoutDying` | batch13_helpers.go | Dour Port-Mage, Aang |
| `b21ArtifactOrCreatureYouControlDied` | batch21_helpers.go | Agent of the Iron Throne |
| `b23ArtifactPutIntoGraveyardFromBattlefield` | batch23_helpers.go | Disciple of the Vault, Viridian Revel |
| `b10LandYouControlDied` | batch10_helpers.go | Titania, Protector of Argoth |
| `scrapTrawlerArtifactHitTheYard` | scrap_trawler.go | Scrap Trawler |
| `aCreatureYouControlLeft` | outpost_siege.go | Outpost Siege |
| `anotherArtifactOrCreaturePutIntoGraveyardFromBattlefield` | slice296c_helpers.go | Tarrian's Soulcleaver |
| inline | the_ozolith.go, cruel_celebrant.go | The Ozolith, Cruel Celebrant |

The per-turn tally's `CreaturesDied` already read `lastKnownBattlefield`
at the event, so Mahadi, Emporium Master was right all along; its caveat
("only a creature because of another effect isn't counted") was stale
and is cleared, with a test pinning the behaviour.

### Still open

- **Subtypes and supertypes as they last existed.** "Whenever a Zombie
  / Vampire / Faerie you control dies" (Undead Augur, Headless Rider,
  Crossway Troublemakers, Tegwyll) still reads the graveyard card's
  subtypes, so a creature that was one of those only through a grant
  (Maskwood Nexus, a changeling-granting effect) is missed — weaker than
  printed, and declared in those cards' comments. *Subtypes closed by
  the #1679 amendment below; supertypes remain open.*
- **Controller as it last existed.** Every "you control" dies clause
  reads `Card.Controller` off the graveyard card. MoveCard does not reset
  it, so that is right for every board that does not change control in
  the same event that moves the permanent.

## Amendment (2026-09-28, #1679): the leaving permanent's subtypes ride the same event

**Status:** Accepted · S37 — Combat correctness · [#1679](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1679)

### Context

The #1675 amendment stamped a leaving permanent's card **types** and
left its subtypes open. The tribal dies-triggers — "whenever another
Zombie you control dies" and its kin — resolved the dead card with
`diedCreature` and then asked `dead.HasSubtype("Zombie")` of the card
**as it sits in the graveyard**. No grant applies there: a Bear that
was a Zombie under Maskwood Nexus, under a changeling grant or under a
lord's type grant is a Bear again, and Diregraf Captain did not drain.
The reverse was wrong as well: a Zombie that Kenrith's Transformation
had made an Elk read as a Zombie once it died. The cards' comments
declared the first half as "weaker, never stronger".

CR 603.10a again: the ability looks back in time, so the question is
what subtypes the permanent had as it last existed on the battlefield.

### Decision 1 — `Event.LastKnownSubtypes` and `Event.LastKnownAllCreatureTypes`

```go
LastKnownSubtypes         []string // post-layer Characteristic.Subtypes as it left — EventLTB only
LastKnownAllCreatureTypes bool     // it was every creature type (changeling, Maskwood Nexus)
```

`exitLKILocked` copies `c.Effective().Subtypes` beside the types, and
records "every creature type" as **one flag** rather than writing the
~345 entries of `AllCreatureTypes` into the list — the same reason
`Characteristic.AllCreatureTypes` is a flag (see `HasAllCreatureTypes`).
The flag is `HasAllCreatureTypes(c)` for any permanent that is not
face down; a face-down permanent is never every creature type
(CR 708.2 — the changeling underneath is text it does not have), which
is the branch `Card.HasSubtype` takes. `stamp` writes both on every
`EventLTB` emit site, so every exit route carries them with no new call
site; `cloneTriggerContext` copies the slice. No new game state: the
fields ride the event log (the v7 shape file records the additive keys).
No wire change: `protocol/log.go` does not project them.

### Decision 2 — `Event.WasSubtype` and `leftAsSubtype`

`Event.WasSubtype(subtype) (was, known bool)` answers with
`Card.HasSubtype`'s semantics: case-insensitive, and a permanent that
was every creature type has every **creature** type (CR 205.3m) — not
"Forest" or "Equipment". `known` is `WasType`'s: the stamp is present
exactly when `LastKnownTypes` is, so a stamped creature with no subtypes
is a known "no", never an unknown.

Card side, `leftAsSubtype(ev, c, subtype)` in `effects/helpers.go`
falls back to `c.HasSubtype` only for an unstamped event. Every tribal
dies condition asks it:

| Reader | File | Cards |
|---|---|---|
| `b17SelfOrZombieYouControlDied` | batch17_helpers.go | Undead Augur |
| `b27SelfOrNontokenZombieYouControlDied` | batch27_helpers.go | Headless Rider |
| `anotherZombieYouControlDied` | batch35_helpers.go | Diregraf Captain, Plague Belcher |
| `b25AnotherGoblinYouControlDied` | batch25_helpers.go | Pashalik Mons |
| `b34VampireYouControlDied` | batch34_helpers.go | Crossway Troublemakers |
| `b36AngelYouControlDied` | batch36_helpers.go | Bishop of Wings |
| `b36AnotherFaerieYouControlDied` | batch36_helpers.go | Tegwyll, Duke of Splendor |
| `anEggYouControlDied` | atla_palani_nest_tender.go | Atla Palani, Nest Tender |
| inline | omnath_locus_of_rage.go | Omnath, Locus of Rage |

Every one of those cards was already `CompletenessFull` with the gap
declared only in its prose; the prose is corrected, and no caveat
changes.

### Still open

- **Supertypes as they last existed.** `b33LegendaryCreatureYouControlDied`
  (Rakdos Joins Up) reads the graveyard card's supertype, so a creature
  that was legendary only through an effect is missed. Same shape, one
  more field, when a card needs it. *Closed by the #1682 amendment
  below.*
- **Controller as it last existed** — unchanged from the #1675 note.
  *Closed by the #1682 amendment below.*
- **Maskwood Nexus off the battlefield.** Its second sentence ("creature
  cards you own that aren't on the battlefield") is still a declared
  caveat on the Nexus; once it lands, the graveyard fallback would agree
  with the stamp for the Nexus case, and the stamp remains what answers
  for every grant that stops at the battlefield.

## Amendment (2026-09-28, #1682): the leaving permanent's supertypes and controller ride the same event

**Status:** Accepted · S37 — Combat correctness · [#1682](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1682)

### Context

The #1675 and #1679 amendments stamped a leaving permanent's card types
and subtypes and left two facts open, both still read off the card as it
sits in its new zone:

- **Supertypes.** Rakdos Joins Up's "whenever a legendary creature you
  control dies" asked `isLegendary` of the graveyard card. A Clone that
  entered as a copy of a legend (CR 707.2 — the copy effect is a
  battlefield effect) was legendary as it died and is a plain Clone in
  the graveyard, so the trigger did not fire.
- **Controller.** Every "a creature you control dies" / "a creature an
  opponent controls dies" clause compared `dead.Controller` with the
  source's controller. "You control" is a question about the
  **permanent**, and a card in a graveyard has no controller at all
  (CR 108.4). The field happens to still name the player who controlled
  the permanent — `MoveCard` does not reset it, so a stolen creature
  sacrificed by its thief reads as the thief's, which is right — but
  that is a side effect of what the move leaves behind, not the rule.
  It is wrong for any reader that looks after the card has moved on (a
  reanimation writes its new controller onto the card before the entry
  pipeline runs), and for any future move that honours CR 108.4.

CR 603.10a again: the ability looks back in time.

### Decision 1 — `Event.LastKnownSupertypes` and `Event.LastKnownController`

```go
LastKnownSupertypes []string  // post-layer Characteristic.Supertypes as it left — EventLTB only
LastKnownController uuid.UUID // Card.Controller as it left — EventLTB only
```

`exitLKILocked` copies `c.Effective().Supertypes` beside the types and
subtypes, and reads `c.Controller` — the field layer 2 materialises its
answer onto (`materialiseControlLocked`), which is what every "you
control" reader on the battlefield asks. `stamp` writes both on every
`EventLTB` emit site, so every exit route (destroy, sacrifice, bounce,
the zone route, the sandbox move and the sandbox drag to the stack)
carries them with no new call site; `cloneTriggerContext` copies the
slice. No new game state: the fields ride the event log (the v7 shape
file records the additive keys, which zero-value to "unknown" for a
file written before them). No wire change: `protocol/log.go` does not
project them.

### Decision 2 — `Event.WasSupertype`, `Event.LeftUnderControlOf`, and their card-side readers

`Event.WasSupertype(supertype) (was, known bool)` answers
case-insensitively, with `WasType`'s `known`: the stamp is present
exactly when `LastKnownTypes` is, so a stamped permanent with no
supertypes is a known "no". `Event.LeftUnderControlOf() (controller,
known)` is known for a stamped `EventLTB` only; a permanent always has a
controller, so `uuid.Nil` only ever means "not stamped".

Card side, in `effects/helpers.go`, `leftAsSupertype(ev, c, supertype)`
and `leftUnderControlOf(ev, c)` fall back to the card as it sits only
for an unstamped event — the pre-#1682 reading, never a false answer.
Rakdos Joins Up asks `leftAsSupertype(ev, dead, "Legendary")`. Every
dies / leaves-the-battlefield condition that asks "you control" or "an
opponent controls" asks `leftUnderControlOf`:

| Reader | File | Cards |
|---|---|---|
| `ACreatureYouControlDied` | triggers_common.go | every `WheneverACreatureYouControlDies` — Zulaport Cutthroat, Grave Pact, Butcher of Malakir, Dictate of Erebos, Bastion of Remembrance, Dark Prophecy, Moldervine Reclamation, Vindictive Vampire |
| `anEggYouControlDied` | atla_palani_nest_tender.go | Atla Palani, Nest Tender |
| `b16SelfOrAnotherCreatureYouControlDied`, `b16AnotherCreatureYouControlEnteredOrDied` | batch16_helpers.go | Vengeful Bloodwitch, Daxos |
| `b17SelfOrZombieYouControlDied` | batch17_helpers.go | Undead Augur |
| `b18OpponentsCreatureDied` | batch18_helpers.go | Sangromancer, Mari, Spiteful Banditry |
| `b19AnotherNontokenCreatureYouControlDied` | batch19_helpers.go | Liesa, Forgotten Archangel; Yedora, Grave Gardener |
| `b22CreatureYouControlDied`, `b22SlimedCreatureYouDontControlDied` | batch22_helpers.go | Cauldron of Essence, Toxrill |
| `b25AnotherGoblinYouControlDied` | batch25_helpers.go | Pashalik Mons |
| `b27SelfOrNontokenZombieYouControlDied` | batch27_helpers.go | Headless Rider |
| `b33LegendaryCreatureYouControlDied`, `b33OpponentsCreatureDied`, `b33OpponentsNontokenCreatureDied` | batch33_helpers.go | Rakdos Joins Up, Patron of the Vein, Overseer of the Damned |
| `b34VampireYouControlDied` | batch34_helpers.go | Crossway Troublemakers |
| `anotherZombieYouControlDied`, `anotherCreatureYouControlDied` | batch35_helpers.go | Diregraf Captain, Plague Belcher; Pitiless Plunderer, Garna, Elas il-Kor |
| `b36AngelYouControlDied`, `b36AnotherFaerieYouControlDied` | batch36_helpers.go | Bishop of Wings, Tegwyll |
| `b39AnotherCreatureYouControlDied` | batch39_helpers.go | Erebos, Bleak-Hearted |
| `edeaCreatureYouControlButDontOwnDied` | edea_possessed_sorceress.go | Edea — the controller half only; "don't own" stays `dead.Owner` |
| `deathTyrantCombatDeath` | death_tyrant.go | Death Tyrant (both halves) |
| `b10LandYouControlDied` | batch10_helpers.go | Titania, Protector of Argoth |
| `creatureYouControlLeftWithoutDying` | batch13_helpers.go | Aang, Dour Port-Mage |
| `b14PermanentYouControlLeft` | batch14_helpers.go | Resourceful Defense |
| `b21ArtifactOrCreatureYouControlDied` | batch21_helpers.go | Agent of the Iron Throne |
| `b31MunitionsYouControlLeft` | batch31_helpers.go | Weapons Manufacturing |
| `aCreatureYouControlLeft` | outpost_siege.go | Outpost Siege |
| `scrapTrawlerArtifactHitTheYard` | scrap_trawler.go | Scrap Trawler |
| inline | ares_god_of_war.go, cruel_celebrant.go, massacre_wurm.go (condition **and** "that player"), midnight_reaper.go, nadiers_nightblade.go, omnath_locus_of_rage.go, open_the_graves.go, pawn_of_ulamog.go, teysa_orzhov_scion.go, the_ozolith.go, vraan_executioner_thane.go, yahenni_undying_partisan.go | as named |

Engine side, the per-turn tally's `CreaturesDied` (Barrensteppe Siege's
Mardu "if a creature died under your control this turn") credits the
stamped controller too.

**Not switched, on purpose:** every clause that means the OWNER — "a
creature card you own", "its owner's graveyard", "under its owner's
control" — reads `Owner`, which a move never changes and which is not
this question. The replacement effects that watch a creature about to
die (Liesa, Stone of Erech, Kismet-style entry checks) run before the
move, with the permanent still on the battlefield, and read it there.

### Still open

- **Colour as it last existed.** Teysa, Orzhov Scion's "whenever
  another black creature you control dies" reads `dead.HasColor("B")`
  off the graveyard card, so a creature that was black only through an
  effect is missed (and one an effect made non-black counts). Same
  shape, one more field, when a card needs it.
