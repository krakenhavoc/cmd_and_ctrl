# ADR 0071 — Designations that switch abilities on: Class levels, solved Cases, station thresholds, unlocked doors

**Status:** Accepted · 2026-09-18 · S46 — Permanents that change what they are
**Issues:** [#757](https://github.com/krakenhavoc/cmd_and_ctrl/issues/757)
(Classes CR 716, Cases CR 719, Rooms CR 709.5),
[#759](https://github.com/krakenhavoc/cmd_and_ctrl/issues/759)
(Station CR 702.184 / CR 721), tracker
[#889](https://github.com/krakenhavoc/cmd_and_ctrl/issues/889)
**Numbering:** on 2026-09-18 every remote head was listed with
`git ls-remote --heads origin` (286 refs) and every `docs/decisions/` filename
that exists on any ref was collected with
`git log --all --name-only --format= -- docs/decisions/`. The highest numbers
present were `0070-untap-step-choices.md` (the #826/#567 untap-choices branch)
and `0070-the-mana-spent-on-a-spell.md` (the #789/#761 branch, since renumbered
to 0068 on `develop`), so 0070 is claimed and this ADR takes **0071**. The scan
was repeated immediately before the push. 0005, 0024, 0029 and 0030 stay
permanently unused per AGENTS.md §4.
**Builds on:** [ADR 0046](0046-layer-6-authoritative.md) (`CatalogAbilityKey` —
the "what does this permanent DO" accessor this ADR extends from a key to an
ability list), [ADR 0069](0069-face-down-objects.md) (`CatalogKey` going silent
for a face-down permanent — the other object-level suppression at the same
seam), [ADR 0010](0010-card-effect-catalog.md) / #622 (the one `CardDef`
lookup), [ADR 0039](0039-layer-4-authoritative.md) (layer-4 type changes are
authoritative, which is what lets a Spacecraft really become a creature),
[ADR 0012](0012-layer-system.md) (layers), [ADR 0020](0020-activated-abilities.md)
(activated abilities; the level-up ability is an ordinary one),
[ADR 0048](0048-cost-modification.md) (cost modifiers), S27 Sagas
(`game/sagas.go` — the "a counter gates abilities" template)
**Designs, does not implement:** Rooms and door unlocks (CR 709.5,
[#757](https://github.com/krakenhavoc/cmd_and_ctrl/issues/757) §Room, tracked
for #886/#889), the station ability's own cost — "tap another untapped creature
you control" ([#758](https://github.com/krakenhavoc/cmd_and_ctrl/issues/758)),
CR 702.184c station-characteristic modifiers (Tapestry Warden), CR 721.2c
"a station card has no P/T outside the battlefield"
**Amends:** [ADR 0034](0034-multi-face-cards.md) §Rooms (out of scope because
the dump had no `room` layout) — this ADR writes down what a Room needs without
building it; [ADR 0037](0037-unimplemented-card-signal.md) §6, whose
"blocked on" rows for Fortune Teller's Talent (#333) and The Seriema (#337) are
superseded by the dated note added with this ADR

---

## Context

Four printed mechanics say the same sentence in four vocabularies:

| Printed | CR | The sentence |
|---|---|---|
| Class level | 716.2 | "as long as this Class is level N or greater, it has [abilities]" |
| Case solved | 719.3 | "Solved — [ability]" — it has the ability while it is solved |
| Station `{N+}` | 721.2 | "as long as this permanent has N or more charge counters, it has [abilities]" |
| Room door | 709.5 | a locked half has no rules text at all |

Each is a **designation** (CR 700.x): a marker a permanent has on the
battlefield that is not a counter, not a characteristic and not copiable, and
whose only job is to switch some of the permanent's own printed abilities on.

The engine has none of them. Checked on `develop` (`efd1b20`):

- `game.Card` has no level, no solved flag and no unlocked flag; nothing in
  `server/` matches `ClassLevel`, `LevelUp`, `Solved` or a door unlock.
- Nothing in the catalog can say "this static / trigger / activated ability
  exists only while …". Every ability slot on `effects.Spec` is
  unconditionally present for as long as the permanent is.
- Station appears nowhere (`docs/decisions/0037-unimplemented-card-signal.md`
  line 183), and `AbilityCost` cannot tap anything but its own source
  (`game/activated.go:35`).

The temptation is four mechanisms. The Saga precedent argues against it:
`game/sagas.go` is a subtype-keyed lifecycle with its own counter, its own
event and its own field on `TriggeredAbility` (`Chapter`), and it is the right
shape *for Sagas* precisely because a Saga's chapters are a sequence, not a
gate. Levels, solved, thresholds and doors are all the same gate, and four
copies of it would be four places to forget the layer invalidation, four places
for the trigger harvester and the activation path to disagree, and four
answers to "why is this ability offered when the card says it isn't yet".

There is also a structural reason to want exactly one. #622 collapsed
twenty-two per-slot catalog lookups into one `CardDef`; ADR 0046 made
`CatalogAbilityKey` the single question "what abilities does this object have
right now" so a seventh caller could not reintroduce the CR 613.1f bug for
free; ADR 0069 put face-down suppression at `CatalogKey`, the one place a
`Card` becomes a catalog key, rather than in each of the readers. A designation
gate is the same kind of question, asked at the same place, and it belongs
there for the same reason.

---

## Decision 1 — One gate on printed abilities, evaluated where an object becomes abilities

Every printed-ability entry the catalog declares may carry one field:

```go
ActiveWhen Designation
```

on `game.StaticAbility`, `game.TriggeredAbility`, `game.ActivatedAbilityShape`
and `game.CostModifier`. Its zero value means "no gate", which is every
ability in the catalog today, so the field costs existing cards nothing.

```go
type DesignationKind uint8

const (
	DesignationAlways         DesignationKind = iota // zero value — no gate
	DesignationClassLevel                            // CR 716.2: level N or greater
	DesignationCaseSolved                            // CR 719.3
	DesignationChargeCounters                        // CR 721.2: N or more charge counters
	DesignationDoorUnlocked                          // CR 709.5 — designed, not built
)

type Designation struct {
	Kind DesignationKind
	N    int      // ClassLevel and ChargeCounters thresholds
	Door DoorSide // DoorUnlocked only
}

func (d Designation) Active(c Card) bool
```

`Active` is the **one predicate**. It reads the object and nothing else — no
`*Game`, no lock, no zone — which is what lets it be called from inside the
layer pass (whose `AppliesTo` predicates must stay passive, see `card.go`'s
note on `Effective()`) and from the trigger harvester on the hot path.

**Where it is evaluated.** At the one point an *object* is turned into the
abilities it currently has. On `develop` that point is already named: ADR
0046's `CatalogAbilityKey(c Card) string`, plus the three call-shaped readers
that sit directly on top of it. This ADR turns those readers into four
object-level accessors in `game/designations.go`, and they become the only way
the engine asks a permanent what it does:

```go
func StaticAbilitiesForCard(c Card) []StaticAbility
func TriggersForCard(c Card) []TriggeredAbility
func ActivatedAbilitiesForCard(c Card) []ActivatedAbilityShape   // pre-existing, gains the gate
func CostModifiersForCard(c Card) []CostModifier
```

Each is one line over one shared filter:

```go
func activeOnly[T any](c Card, all []T, gate func(T) Designation) []T
```

so there is one gate function and one predicate, and adding a fifth slot is a
field and a one-line accessor. Adding a new *caller* — a walk over a zone the
harvest did not use to visit — is nothing at all, as long as it asks the
accessor: #925's declared-zone harvest (`game/trigger_zones.go`) reads a card
as it sits in a graveyard or in exile, and goes through `TriggersForCard` for
exactly this reason. A second, parallel `CatalogTriggers` read there would fire
a Case's "Solved — …" ability off a Case in a graveyard, where CR 400.7 has
already taken the designation away. An inactive ability is simply **not in the list
the object hands back**, so the layer pass, the trigger harvester, the
activation path, the legal-move enumerator, the cost pricer and the wire view
all agree with no second filter anywhere — the same property ADR 0046 bought
for ability removal and ADR 0069 bought for face-down.

Three deliberate details:

1. **Which key accessor each uses is unchanged.** `StaticAbilitiesForCard`
   keeps `CatalogKey`, because the layer pass gathers before it silences and
   `layers.go:317` explains why; the other three keep `CatalogAbilityKey`.
   The gate is orthogonal to the key, and moving it would be a separate,
   wrong change.
2. **The per-slot `Catalog*` function variables stay.** Two dozen test files
   stub them to inject catalog behaviour without importing `effects`. The
   accessors read *through* them, so a stub still works and the gate still
   applies to what the stub returns.
3. **Mana abilities and replacement effects do not get the field yet.** Not a
   principle — no printed Class, Case or Spacecraft in the dump gates one, and
   the field plus its accessor is the same two lines on the day one does. The
   place to add it is `ManaAbilitiesForCard` / `gatherActiveReplacementsLocked`,
   which already have the object in hand.

### Layer invalidation

A designation change must bump `layerVersion` or a gated static goes stale.
Charge counters already do (`layer_listener.go`, `EventCounterPlaced`, which is
emitted for removals too). The two new events join the same switch:
`EventClassLevel` and `EventCaseSolved` bump unconditionally. That is the whole
of the "Layer invalidation" checkbox on #757 and #759.

---

## Decision 2 — The designations themselves

### Class level — `Card.ClassLevel int` (CR 716.2)

A **designation, not a counter**. CR 716.2b: a Class permanent with no level
designation is level 1, so the field's zero value reads as level 1 and
`ClassLevelOf(c)` is `max(1, c.ClassLevel)`. Because it is not a counter it
cannot be proliferated, doubled by Vorinclex, or removed by a counter-removal
cost — which falls out of not being in `Card.Counters` rather than needing a
rule anywhere.

The "Level N" ability is an **ordinary activated ability** (CR 716.2a), built
by one constructor rather than hand-written per card:

```go
func LevelUp(level int, cost game.AbilityCost) ActivatedAbility
```

It produces `SorcerySpeed: true` (CR 716.2d) and a `Condition` that reads the
source's current level and demands exactly `level-1` (CR 716.2a). Nothing else
is special: it is announced, paid and put on the stack by the same
`ActivateCatalogAbility` path every other activation uses, its cost is an
ordinary `AbilityCost` (mana, since #958 also counters), the legal-move
enumerator offers it through `ActivatedAbilitiesForCard`, and the client
already renders `condition_unmet` for the levels that are out of reach.

Its effect calls `g.SetClassLevelForEffect(cardID, n)`, which sets the field
and emits `EventClassLevel` with `Amount = n`. "When this Class becomes level
N" is then an ordinary trigger watching that event — and one gated
`ClassLevel(N)`, so it does not exist until the level that prints it is
reached.

CR 716.2c: **level is not a copiable value.** A copy of a level-3 Class is
level 1. This falls out of where the field lives: `CopiableValuesOf` projects
printed characteristics, `Card.ClassLevel` is battlefield state next to
`Tapped` and `Counters`, and nothing in the copy path reads it.

CR 400.7: cleared on battlefield exit, in the two places `NamedTribe` and
`ChosenColor` are cleared (`zone.go`, `entry_tail.go`).

### Case solved — `Card.Solved bool` (CR 719.3)

"To solve — [condition]" is a triggered ability with an intervening if
(CR 719.3a): *at the beginning of your end step, if [condition], this Case
becomes solved*. One constructor:

```go
func ToSolve(label string, condition func(g *game.Game, controller, source uuid.UUID) bool) game.TriggeredAbility
```

watching `EventBeginEndStep`, gated on `ev.Actor == source.Controller`, skipped
when the Case is already solved, and **re-checking the condition at resolution**
(CR 603.4). Its effect calls `g.SolveCaseForEffect(cardID)`, which sets the flag
and emits `EventCaseSolved`.

"Solved — [ability]" is a `CaseSolved` gate on any of the four slots; Case of
the Shattered Pact gates a trigger, and a Case that gates a static or an
activated ability needs nothing new.

A solved Case **stays solved** while it is on the battlefield (CR 719.3b), it
is not copiable, and it is cleared on battlefield exit — the same three
sentences as the level, for the same reasons and in the same places.

### Station thresholds — plain charge counters (CR 721.2)

No new state. A station threshold is `Counters[CounterCharge]`, the counter
type that has existed since the S13.x registry (`counter_types.go`), and the
`{N+}` gate reads the count **live** on every evaluation, so a permanent that
loses charge counters loses the abilities again — which is CR 721.2a read
literally and is the one behavioural difference from a Saga's ratchet.

CR 721.2b, "with a P/T box it is also a creature with base power and toughness
[P/T] in addition to its other types", is two gated statics from one
constructor:

```go
func SpacecraftAt(n, power, toughness int) []game.StaticAbility
```

— a `Layer4Type` self-static that appends `Creature`, and a
`Layer7PT`/`SubLayer7B_Set` self-static that sets base P/T, both
`ActiveWhen: ChargeCounters(n)`. Because ADR 0039 made layer 4 authoritative,
that is a real creature to combat, targeting and the SBAs, with no wire-only
gap. Keywords printed on a threshold line ("7+ | Flying") are
`ThresholdKeywords(n, "flying")`, a gated `Layer6Ability` self-grant.

**Summoning sickness** needs no rule of its own. CR 302.6 asks how long the
permanent has been under its controller's control, not how long it has been a
creature, and `Card.SummonedThisTurn` already records exactly that (it is
stamped on entry and cleared at the controller's untap step, and #537 made the
check a creature rule). A Spacecraft that entered last turn and stations to its
threshold this turn can attack; one that entered this turn cannot.

### The station ability itself — designed here, built on #758

CR 702.184a: *"Tap another untapped creature you control: Put a number of
charge counters on this permanent equal to the tapped creature's power.
Activate only as a sorcery."*

The cost is #758's, and this ADR does not build it. It **names the field #758
should add**, so the constructor below can be written the day it lands:

```go
// AbilityCost gains exactly one field:
TapOthers *TapOthersCost

type TapOthersCost struct {
	Count         int          // 1 for station; 3 for Heritage Druid
	Filter        *TargetSpec  // "creature you control"
	ExcludeSource bool         // "another" — true for station
	Label         string       // the picker's prompt
}
```

and then

```go
func Station() ActivatedAbility   // Cost: TapAnother(Creature()), SorcerySpeed: true
```

whose effect puts `power` charge counters on the source, where `power` is the
tapped creature's **effective power at the moment the cost is paid** — the
reading crew already takes (CR 702.122a, `validateCrewCostLocked`), and the one
that makes the tapped creature's subsequent death irrelevant. The stack item
therefore carries the number, not the creature's ID.

**Out of scope, stated:** CR 702.184c modifiers that make station read a
different characteristic (Tapestry Warden's toughness) — they are a static
ability over a cost, which is a different seam; and CR 721.2c, "a station card
has no P/T outside the battlefield", which is a view and importer rule. The
Seriema ships showing its printed 5/5 in hand.

---

## Decision 3 — Rooms (CR 709.5): designed, not built

Written down here so #886/#889 does not re-derive it, and so ADR 0034's "Rooms
are out of scope" has a successor. **Nothing in this ADR's implementation
builds any of it**, and `DesignationDoorUnlocked` is reserved rather than
usable: `Active` returns false for it, no `effects` constructor produces one,
and a guard test fails the build if a registered spec declares one.

1. **A Room is a split-layout permanent whose two halves are doors.** Each
   half has its own name, mana cost and rules text; a locked half has none of
   them (CR 709.5a, CR 116.2m). Which half each characteristic belongs to **is**
   copiable (CR 709.5b) — unlike every other designation here — so it belongs
   with the face model of ADR 0034, not with `Card.ClassLevel`.
2. **The state** is two bools, `Card.DoorsUnlocked [2]bool` or a small bitfield
   in the existing bool block, cleared on battlefield exit like the rest.
   The half that was cast enters unlocked; a Room put onto the battlefield
   without being cast enters with neither half unlocked (CR 709.5d).
3. **The gate** is `DoorUnlocked(left)` / `DoorUnlocked(right)` on the abilities
   printed on that half — the same field, the same filter, the same accessors.
4. **Casting either half** is CR 709.3 and is *not* Room-specific:
   `CastableFaces` casts a split card front-only (`game/face.go:202`) and
   `deck/validate.go:302` says so. Right-half casting is a prerequisite and has
   no issue of its own yet.
5. **Unlocking** is a special action at sorcery timing that pays the locked
   half's mana cost (CR 709.5e) — a new action type with a legal-move
   enumerator case, not an activated ability. "When you unlock this door" and
   "fully unlock" (CR 709.5h–i) are ordinary triggers on the unlock event.

---

## Decision 4 — Wire and client

Two fields on `CardView`, both omitempty, both public (a level and a solved
badge are visible to every player in paper):

```go
ClassLevel int  `json:"class_level,omitempty"`
Solved     bool `json:"solved,omitempty"`
```

`class_level` is projected only for a permanent that actually has a level
designation — i.e. a battlefield permanent whose level is 2 or more, or any
Class on the battlefield, which is the "an uncatalogued Class should still show
its level" ask on #757 answered the way an uncatalogued Saga shows its lore.

**Which printed abilities are active needs no new wire surface.** An inactive
activated ability is absent from `ActivatedAbilitiesForCard`, so it is already
absent from `activated_abilities` in the view and from the right-click menu;
an inactive static or trigger has no per-ability wire representation to grey
out in the first place. Charge counters ride the existing generic `counters`
map, exactly as lore counters do — the Saga chapter rendering the client has is
the counter pip, and it needs nothing added.

The client renders the two fields as badges on the card, next to the counter
pips.

---

## Decision 5 — Bots

Level-up is an ordinary activation and needs no enumerator change: `internal/legal`
reads `game.ActivatedAbilitiesForCard`, which is the accessor the gate lives
in, so a bot is offered exactly the abilities the engine will accept — the
#544 invariant. Solving is not a move (it is a trigger), and station is not
offered until #758 gives its cost a payment shape.

---

## Decision 6 — Undo, clone and persistence

Per ADR 0041's classification, `Card.ClassLevel` and `Card.Solved` are both
**carried**.

- `clone.go` needs no change: `cloneCard` starts `out := c` and both are value
  fields.
- `snapshot.go` gains the two fields on `CardSnapshot` and one line in each of
  the two projection sites. A restore that dropped them would resurrect a
  level-3 Wizard Class as a level-1 one with two of its three abilities gone,
  and a solved Case unsolved — silently, because both are legal states.
- `snapshot_backfill.go` needs nothing: the zero values are the correct
  reading of an old snapshot (level 1, unsolved).
- Undo across a level-up restores the level, because undo restores the card.

---

## Consequences

**Good.** One gate, one predicate, one place. Class levels, solved Cases and
station thresholds are the same three lines of catalog declaration, and
unlocked doors will be a fourth with no engine change beyond the state. The
layer pass, the harvester, the activation path, the cost pricer, the
enumerator and the wire cannot disagree about whether an ability exists,
because they all ask the same accessor. Nothing in the choice-kind gate moved
and no new `PendingChoiceKind` was added.

**Costs, stated.** The gate runs per ability per object per gather. It is a
struct-field compare and, for charge counters, one map read — measurably
cheaper than the catalog lookup it follows, and it short-circuits on the zero
value for every ability in the catalog that has no gate. `Card` grows 8 bytes
for `ClassLevel` (`Solved` goes in the existing bool block, which
`TestCardHasNoInteriorPadding` enforces).

**What this does not fix.** A Class with no catalog entry shows its level and
nothing else, exactly as an uncatalogued Saga shows its lore — the levels are
visible and a player resolves the abilities by hand. And the gate answers "does
this ability exist"; it does not answer "did this ability exist when the event
happened", so an LTB trigger printed on a level-3 line is judged on the
last-known-information snapshot like every other LTB trigger (`harvestLTB`),
which reads the level off the snapshot because the level is on the card.

## Cards this unblocks in this PR

| Card | What it proves |
|---|---|
| Wizard Class (#298) | a Class end to end: level 1 always on, a "becomes level N" trigger, a level-3 trigger that does not fire at level 2 |
| Fortune Teller's Talent (#333, [#757](https://github.com/krakenhavoc/cmd_and_ctrl/issues/757)) | a gated **cost modifier**; levels 1 and 2 are caveats on [#765](https://github.com/krakenhavoc/cmd_and_ctrl/issues/765) |
| Case of the Shattered Pact | "To solve" at the end step with an intervening if, and a gated **trigger** |
| The Seriema (#337, [#759](https://github.com/krakenhavoc/cmd_and_ctrl/issues/759)) | gated **statics** across layers 4, 6 and 7b — the Spacecraft that becomes a creature at 7+; the station ability itself is a caveat on [#758](https://github.com/krakenhavoc/cmd_and_ctrl/issues/758) |

---

## Addendum (2026-09-22): the zone is a second dimension on the same accessors, not a second designation (#1221)

**Status:** Accepted · 2026-09-22 · tracked on
[#1221](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1221). This
status covers this section only; every decision above stays accepted
and unchanged.

### Context

Decision 1 above put ONE gate on printed abilities and evaluated it at
the one point an object is turned into the abilities it currently has
— the four object-level accessors in `game/designations.go`. #1221
opens a second question at exactly the same point: not "does this
permanent have this ability" but "does this ability function **where
this card is**" (CR 113.6). An unearth ability is on the card in its
owner's graveyard and is not on the same card on the battlefield.

Two questions, one place they get asked. The temptation is to make
that one mechanism.

### Decision: the zone stays out of `Designation`, and out of the accessors' return value

`Designation.Active(c Card) bool` reads the object and nothing else —
no `*Game`, no lock, **no zone** — and that is the property that lets
it run inside the layer pass and on the trigger harvester's hot path.
A `DesignationInZone` kind would have had to take one, because a card
does not know which pile holds it.

So the zone is a sibling predicate and not a `Designation`:

```go
func AbilityFunctionsFromZone(ab ActivatedAbilityShape, zone ZoneKind) bool
```

one function in `game/ability_zone.go`, read by the three consumers
that have the zone in hand — `ActivateCatalogAbility`, the legal
enumerator, and the view's stamp. It is the same "one predicate,
every consumer" shape `activeOnly` has; it simply cannot be the same
FUNCTION, because its extra argument is one the object does not carry.

**And the accessors do not filter by it.** `ActivatedAbilitiesForCard`
returns the card's full list, gated on designation only, and each
consumer skips what does not function where it is looking. That is
deliberate, and the reason is an index: `ActivateAbilityParams.Index`
is the ability's position in the card's FULL list, which is what the
engine validates against and what the wire and the enumerator publish.
An accessor that returned a zone-filtered slice would renumber it, and
a card with a battlefield ability at 0 and a graveyard ability at 1
would have its graveyard row published as index 0 and activated as
the wrong ability. A filtered VIEW keeps the index; a filtered SLICE
loses it.

The consequence is one line of discipline rather than a guarantee: a
fourth consumer that reads `ActivatedAbilitiesForCard` and forgets the
zone predicate would offer a cycling row on a battlefield permanent.
The accessor's own doc names the predicate for that reason, and the
enumerator, the view and the engine each have a test that a
zone-mismatched ability is absent — which is #544's invariant asked of
the zone instead of the designation.

### Layer invalidation, and the half this addendum does not build

Decision 1's "Layer invalidation" note says a designation change must
bump `layerVersion`. The zone dimension owes the same debt on the
STATICS side — a card entering or leaving a graveyard changes which
statics the layer pass should gather — and #1117 already paid it: the
graveyard half of the zone-keyed bump condition is not gated on a
flag, so any event whose `OldZone` / `NewZone` crosses a graveyard
boundary invalidates.

`StaticAbility.Zones`, and the layer pass's walk over a controller's
own graveyard (CR 113.6c — Anger, Wonder, Brawn), are the other half
of #1221 and are NOT in this addendum. They land next, against this
record, and the rule they will follow is the one above: the zone is a
second dimension read at the same point, not a second designation.

---

## Addendum (2026-09-22): statics that function from a graveyard (#1221)

**Status:** Accepted · 2026-09-22 · tracked on
[#1221](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1221). The
other half of the same issue; the addendum above is the activated
half's record and is unchanged. This status covers this section only.

### Context

The addendum above said the zone is a second dimension read at the
point Decision 1 put the designation gate, and that the STATIC half
would land next against this record. It has.

`Game.activeStaticAbilitiesLocked` gathered from three sources —
floating scoped effects, emblems, and a walk of `g.Battlefield` — so a
static printed on a card in a GRAVEYARD was never gathered, never
sorted into a bucket and never applied. CR 113.6c is the clause four
printed cards quote:

> **Anger** — "As long as this card is in your graveyard and you
> control a Mountain, creatures you control have haste."

with Wonder, Brawn and Valor the same sentence in the other colours.
The engine-seams row "Statics read from a non-battlefield zone" is
those three plus Valor, and the neighbouring "Layer invalidation" row
had already named Wonder as unblocked on ITS axis and still blocked on
this one.

### Decision: `StaticAbility.Zones`, gathered by one narrow walk behind a boot-time index

The field is `ActivatedAbilityShape.Zones`' and `TriggeredAbility.Zones`'
sibling, with the same shape, the same nil default and the same
per-DECLARATION scope:

```go
Zones []ZoneKind   // nil means the battlefield

func StaticZones(s StaticAbility) []ZoneKind
func StaticFunctionsFromZone(s StaticAbility, zone ZoneKind) bool
func StaticZoneUnsupported(zone ZoneKind) string
```

`game/static_zones.go`, and it is `trigger_zones.go` (#925) one
consumer over — deliberately, down to the index:

- **`supportedStaticZones` is the graveyard and nothing else.** A zone
  the gather does not walk would be a declaration the engine silently
  ignored, so `effects.Register` refuses the rest at boot with the
  reason. The hand is the obvious next one and is one entry plus one
  line in `zonesOfKindLocked`'s existing switch on the day a card
  needs it.
- **The index is built once, at `effects.Register`.** A blanket "walk
  every graveyard on every recompute" would be a real cost on a
  four-player table with sixty cards in the yards, paid by every table
  whether or not anything in any deck functions from there.
  `staticZones.zones()` answers "which zones does ANY registered card
  declare a static from" — nil for almost every game, and the whole
  walk is then one slice read — and `declares(oracleKey)` lets a
  walked zone skip the catalog lookup for every card that does not.
- **The battlefield walk gains one comparison per ability.** That is
  what keeps a declared zone from ALSO applying on the battlefield: a
  battlefield Anger is a 2/2 with its own printed haste and no anthem,
  which is what the card says and why it prints the keyword as a
  separate line.
- **The gather goes through `StaticAbilitiesForCard`**, not the raw
  `CatalogStaticAbilities` hook. This is the second place an object
  becomes statics, and Decision 1's whole point is that there is one
  such place per slot: a parallel read here would apply a Case's
  "Solved —" static off a Case in a graveyard, where CR 400.7 has
  already taken the designation away.

### Three details of the bound source, each a rule rather than a choice

1. **`Controller = Owner` on the bound copy.** CR 108.4: a card
   outside the battlefield and the stack has no controller, so
   "creatures YOU control" is its OWNER's. An Anger milled out of an
   opponent's library helps that opponent. Word for word what
   `harvestFromDeclaredZone` does for triggers and
   `ActivateCatalogAbility` does for activations.

2. **`live` is false.** The flag exists for one thing — CR 613.1f
   ability-removal silencing — and nothing on the board can silence a
   card in a graveyard, because Darksteel Mutation applies to a
   permanent. Saying so at the bind is what keeps `applyBucketLocked`
   from asking.

3. **The timestamp is the source's last battlefield entry**, which is
   when it died for the card this exists for and zero for one milled
   straight out of a library. CR 613.7 wants the time the object
   entered the zone it is in and no field records that. It is
   unobservable for this whole family — every incarnation is a
   layer-6 keyword GRANT, and grants commute — so the choice was
   between an approximation and a new `Card` field that the snapshot,
   the clone and the drift guard would all carry for a number nothing
   can read. **Stated rather than hidden:** a graveyard static that
   SET a characteristic would need the real thing, and there is none
   in print that the catalog wants.

### Layer invalidation: already paid, and now observable

Decision 1's "Layer invalidation" note says a designation change must
bump `layerVersion` or a gated static goes stale. The zone dimension
owes the same debt and #1117 had already paid it for the graveyard:
the zone-keyed condition in `layerVersionBump.OnEvent` is deliberately
NOT gated on a flag, so **any** event whose `OldZone` / `NewZone`
crosses a graveyard boundary bumps — arrivals and departures alike —
and a battlefield crossing bumps on the other arm. A creature dying,
a card milled, a card discarded and a graveyard exiled are therefore
all covered, with no new bump and no new flag.

What #1221 adds is a test. Before this addendum the widening was
justified by a family of ten battlefield statics that READ a graveyard
and could not have declared a flag they predate; now there is a static
whose SOURCE is in one, and "the cached resolution survived the
graveyard arrival" is a one-line assertion instead of an argument.

### Cards

Anger, Wonder, Brawn and Valor, all four, all `full`. One constructor
(`effects.incarnationAnthem(keyword, land)`) because the four differ
in two strings and nothing else, and four card files because
one-file-per-card is what keeps `git blame` honest.

The clause's other half — "and you control a Mountain" — reads a basic
land TYPE and not the basic supertype, so a Sacred Foundry turns Anger
on. It walks `g.Battlefield.Cards` directly rather than through
`BattlefieldCardsForEffect`, which copies the pile: this runs inside
`AppliesTo`, once per candidate creature per recompute, and a copy per
call would make the pass quadratic in the board.

### Still out of scope

- **Yixlid Jailer** ("cards in graveyards lose all abilities") is NOT
  this row and never was: it is a battlefield static ABOUT graveyards,
  and what it wants is the ability-suppression seam.
- **"As long as this card is in your hand" statics.** No catalog card
  asks yet, and `supportedStaticZones` refuses the zone at boot until
  one does.
- **A graveyard static that SETS a characteristic**, for the timestamp
  reason above.
- **Mana abilities and replacement effects still have no zone field**,
  unchanged from Decision 1 note 3 — and the mana half now has a
  card asking (#1228).

---

## Addendum (2026-09-23): the station ability, and which power it reads (#759)

**Status:** Accepted · 2026-09-23 · tracked on
[#759](https://github.com/krakenhavoc/cmd_and_ctrl/issues/759), trackers
[#889](https://github.com/krakenhavoc/cmd_and_ctrl/issues/889) (permanents that
change what they are) and [#887](https://github.com/krakenhavoc/cmd_and_ctrl/issues/887)
(mana and costs). This status covers this section only. It **amends** two
things above: Decision 2's "The station ability itself", in which power the
ability reads, and Decision 5's "station is not offered until #758 gives its
cost a payment shape". Everything else stays accepted and unchanged.

### Context

Decision 2 designed `Station()` on a cost that did not exist yet, and named the
field #758 should add. #758 has since added it: `AbilityCost.TapOthers
*TapOthersCost` (#1092, wired into both activation paths by #1102), with the
four fields named here. It deliberately shipped no consumer. Two things were
still owed:

1. **A record of what the cost tapped.** `payTapOthersCostLocked` returned the
   tapped cards and `ActivateCatalogAbility` threw them away. It runs AFTER the
   stack item is built (ADR 0020 §4: a tap-watching trigger must sit above the
   ability), so nothing on the item could say which creature paid.
2. **The wire.** No view options, no enumerator payment and no client picker
   existed for the component on an activated ability. An ability with this cost
   was reachable from Go and from nothing else, which is ADR 0037 §5's "a card
   nobody can activate".

Decision 2 also got one thing wrong. It said the ability reads the tapped
creature's power **at the moment the cost is paid**, by analogy with crew. The
analogy does not hold. Crew's power is part of the COST (CR 702.122a taps
"any number of untapped creatures you control with total power N or greater").
Station's power is in the EFFECT: "put a number of charge counters on this
permanent equal to the tapped creature's power" (CR 702.184a, pinned August 7,
2026 edition). CR 608.2h says an effect that needs information from a specific
object reads it when the effect is applied. It uses the object if the object is
still in the zone it was expected to be in, and the object's last-known
information if not. The Edge of Eternities release notes say the same of station
directly: *"Use the tapped creature's power as the station ability resolves …
If that creature isn't on the battlefield at that time, use its power as it
last existed on the battlefield,"* and *"If the tapped creature has negative
power, no charge counters are put onto or removed from the permanent with
station."*

The difference is observable. Pump the tapped creature in response and it
stations for more. Shrink it in response and it stations for less.

### Decision 1: the payment record carries the tapped OBJECT, with a last-known fallback

`PaidCost` gains one field. It sits next to `Sacrificed` (#1213) and
`ReturnedAttacking` (#1227) and was added for the same reason: it is a fact
about the announcement that nothing at resolution could recompute.

```go
TappedOthers []PaidTap   // one per permanent the TapOthers component tapped

type PaidTap struct {
	ID    uuid.UUID // the instance tapped
	Epoch int       // Card.ObjectEpoch as it was tapped: WHICH object (CR 400.7)
	Power int       // PowerForComparison as tapped; rewritten once as it leaves
	Left  bool      // Power is now last-known and final
}
```

`ActivateCatalogAbility` stamps it from `payTapOthersCostLocked`'s snapshot,
AFTER the item exists. That is the one line Luke's handoff on #759 named. Clone
and snapshot carry it like the rest of the record. A CR 707.10 ability copy
carries it too, in a slice of its own, because the choice of what to tap was
made when the original was put on the stack.

**The reader** is `Game.PaidTapPowerForEffect(PaidTap)`. A card reaches it as
`Context.TappedPower()`.

- If the creature is on the battlefield as the SAME object (instance ID and
  epoch both match), it returns the creature's power now, read after a layer
  recompute.
- If the creature has left, it returns the `Power` frozen as it left.

**The freeze** is `freezePaidTapsOnExitLocked`, called from
`battlefieldExitLocked`. That is the one place every battlefield exit passes
through, and it runs with the card still on the battlefield, next to the
CR 603.10 snapshot that leave-the-battlefield triggers are judged on. It
rewrites `Power` on every stack item whose record names the leaving object and
marks it `Left`. The engine's own `lastKnownBattlefield` cannot do this job: it
lives for one mutation and is gone by the time the ability resolves.

Rejected:

- **Banking the number at payment** (Decision 2's reading). Simpler, and wrong
  in both directions against the release notes.
- **Storing only the ID and reading `LookupCardForEffect` at resolution.** A
  creature that is bounced and replayed keeps its instance ID, and the new
  object's power is not the tapped creature's (CR 400.7). The epoch tells the
  two apart, and the freeze keeps the answer once the object is gone.

The fallback `Power` is the power at payment. It is used only if an object
leaves by a route that never reaches the exit choke point, such as a player
leaving the game.

### Decision 2: `Station()` is a constructor over existing shapes

```go
func Station() ActivatedAbility {
	return ActivatedAbility{
		Label:        StationLabel,
		Cost:         TapAnotherUntapped("another untapped creature you control", Creature()),
		SorcerySpeed: true,
		Effect:       stationEffect,
	}
}
```

- `TapAnotherUntapped` is a `TapOthersCost` with a count of one and
  `ExcludeSource` set, which is the printed word "another". It is not the `{T}`
  symbol, so a creature that arrived this turn may station (CR 302.6). It does
  not target, so a hexproof creature may too.
- There is no `ActiveWhen`. CR 721.4: every station card has its station
  ability at every counter count.
- The effect puts `TappedPower()` charge counters on the source through the
  CR 614 counter window, so Doubling Season doubles them. When the power is zero
  or negative it puts nothing on. A negative `AddCounter` would REMOVE counters,
  which is the one outcome the ruling forbids.
- A source that has stopped being this permanent gets nothing: destroyed in
  response, or replayed as a new object. That check is
  `AbilitySourceGoneForEffect`. Its doc named attaching as the only effect that
  cannot happen without its source. "Put counters on THIS permanent" is the
  second, because `AddCounterForEffect` would otherwise stamp the counters onto
  a card in a graveyard.
- `effects.Plus` now carries `TapOthers`. It used to drop the field silently,
  so a composed "{T}, tap another untapped creature" would have become the bare
  `{T}` ability, which is the #259 direction.

### Decision 3: the wire, the enumerator and the bot

- **View:** `activated_abilities[i]` and `mana_abilities[i]` carry the same
  `tap_others_label` / `tap_others_options`, built from
  `TapOthersOptionsForEffect`, the same walk the validator uses (#544). The
  source is left off when the ability also prints `{T}` (CR 118.3). The names
  are kept apart from the cast-side `tap_cost` (convoke, waterbend), which is a
  different component.
- **Payload:** `tap_ids` rides the matching `activate_ability` or
  `activate_mana_ability` action (#1102 / #758).
- **Enumerator:** one move per candidate creature for a one-permanent clause, in
  battlefield order. This is deliberately NOT the fuel order the sacrifice and
  return payments use: tapping spends nothing, and which creature is best to tap
  is the POLICY's call. The named creatures are excluded from the mana
  affordability probe (#1242's rule).
- **Heuristic:** a tapped creature costs a blocker (the same price as crew). It
  also costs the creature's attack when the move comes before combat on the
  bot's own turn and the creature could swing. The payoff is a small amount per
  point of power, which is the one number on the move that measures what
  station buys. So a bot stations after combat, or with a summoning-sick
  creature, and not with its attackers.
- **Client:** the sacrifice picker with the verb "Tap", asked after the return
  pick and skipped when the board offers exactly one permanent. Both ability
  menus grey the row when nothing can pay.

This supersedes Decision 5's "station is not offered".

### Cards

| Card | What it proves |
|---|---|
| The Seriema (#337) | the station ability end to end, to 7+; goes **Full** |
| Galvanizing Sawship | the smallest complete station card: one threshold, a P/T box, haste |
| Uthros Research Craft | a gated **trigger** (3+) that adds a charge counter itself, and 12+ with a layer-7c modifier over the 7b base |
| Adagia, Windswept Bastion | a **Planet**, which is a land with station: enters tapped, taps for mana, stations from another creature, and has a gated activated ability at 12+ that makes a legendary token copy |

### Still out of scope

- **CR 702.184c** modifiers (Tapestry Warden's "toughness instead of power").
  They are a static over this ability, and no catalog card has one.
- **CR 721.2c**, no P/T outside the battlefield. Unchanged from Decision 2.
- **Variable-count tap-others costs** (#1421), such as "Tap X untapped Foods you
  control". The fixed `TapOthersCost.Count` cannot represent an announced X;
  that needs the announce path used by variable sacrifice costs.
- **Planets whose threshold line is a MANA ability** (Evendo, Waking Haven;
  Uthros, Titanic Godcore). `ActiveWhen` is not on mana abilities (Decision 1,
  note 3).

---

## Addendum (2026-09-23): a fifth designation, Harnessed (CR 701.64, #1321)

**Status:** Accepted · 2026-09-23 · tracked on
[#1321](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1321),
tracker [#889](https://github.com/krakenhavoc/cmd_and_ctrl/issues/889).

### Context

CR 701.64 prints a fifth designation, same shape as the four Decision
1 named:

> **701.64a** "Harness [this permanent]" means "If this permanent
> isn't harnessed, it becomes harnessed."
> **701.64b** Harnessed is a designation… Once a permanent becomes
> harnessed, it stays harnessed until it leaves the battlefield.
> **702.186b** "∞ — [Ability]" means "As long as this permanent is
> harnessed, it has [ability]."

The Mind Stone is the first (and, in the dump checked for this ADR,
only) printed card that uses it: `{5}{W}, {T}: Harness The Mind
Stone.` gates `∞ — At the beginning of your end step, exile up to one
other target nonland permanent you control, then return that card to
the battlefield under its owner's control.` Nothing about it needed a
new mechanism — CR 701.64b is CR 719.3b's sentence with a different
noun — so this addendum is Decision 2's Case-solved section, copied
for the fifth kind, not a new design.

### Decision: `DesignationHarnessed`, `Card.Harnessed`, `HarnessForEffect` — Solved's three pieces, verbatim

- **`DesignationHarnessed`** joins the `DesignationKind` enum
  (`game/designations.go`), `Active` gains `case
  DesignationHarnessed: return c.Harnessed`, and `Harnessed()`
  constructs the gate — the same three lines `DesignationCaseSolved` /
  `CaseSolved()` are.
- **`Card.Harnessed bool`** sits in the bool block next to `Solved`,
  for `TestCardHasNoInteriorPadding`'s reason (Decision 6's "8 bytes
  for `ClassLevel`, `Solved` costs nothing" arithmetic — a bool costs
  nothing here either). Not copiable (CR 701.64 says nothing about a
  copy inheriting it, and it is battlefield state next to `Tapped`,
  not a printed characteristic — the same non-argument that makes
  `ClassLevel` / `Solved` non-copiable), cleared on battlefield exit
  in the same two places (`zone.go`, `entry_tail.go`), carried by the
  snapshot (`cardSnapshot.Harnessed`, both projection sites).
- **`HarnessForEffect` / `IsHarnessed`** are `SolveCaseForEffect` /
  `IsSolved` with the noun changed, including the idempotency:
  CR 701.64a is worded "if this permanent ISN'T harnessed, it
  becomes" — an "if", not a legality restriction — so nothing in the
  engine refuses a second activation of "Harness [this permanent]",
  and `HarnessForEffect` simply does nothing the second time, exactly
  as `SolveCaseForEffect` does for a Case solved twice.
- **`EventHarnessed`** joins `EventClassLevel` / `EventCaseSolved` in
  the layer listener's designation-bump case. Silent on the public
  log (`silentEventKinds["EventHarnessed"] =
  silentBoardStateIsVisible`), for `EventCaseSolved`'s reason: the
  state is a wire badge, not a narrated line.

**Wire:** `CardView.harnessed` (omitempty), unconditional once a
permanent is on the battlefield — unlike `class_level` / `solved` it
carries no subtype probe, because CR 701.64 names no card type the
way CR 716 names Class and CR 719 names Case; any permanent can print
a Harness ability. Cleared on the non-knower redaction with
`class_level`, `solved` and `prepared`. See
[docs/protocol.md](../protocol.md).

**Catalog side** (`cards/effects/designations.go`): `Harnessed()`
wraps `game.Harnessed()` exactly as `Solved()` wraps `game.CaseSolved()`,
and `Harness(label, cost)` is the "Harness [this permanent]" activated
ability — no `ActiveWhen`, no `Condition`, for the idempotency argument
above; a `Condition` that refused the second activation would make the
ability unactivatable rather than a no-op, which CR 701.64a does not
say. The "∞ — [ability]" line itself is an ordinary `TriggeredAbility`
with `ActiveWhen: Harnessed()` — no new catalog machinery, which is
the whole point of Decision 1's one-gate design paying off a fifth
time.

**Proof card:** The Mind Stone
([the_mind_stone.go](../../server/internal/cards/effects/the_mind_stone.go)),
`CompletenessFull` — indestructible, `{T}: Add {W}`, the Harness
ability and the ∞ end-step flicker (Thassa, Deep-Dwelling's blink,
CR 400.7 new-object semantics and all, with `Flicker.Controller` left
at its zero value for "under its OWNER's control" rather than
Thassa's "under YOUR control").

### Consequences

Good: a fifth designation is three fields and a constructor, and every
consumer that already reads a gate (the layer pass, the harvester, the
enumerator, the wire) needed no change to honour it — Decision 1's
whole argument, paid off again. `DesignationDoorUnlocked` stays
reserved and unbuilt; this addendum does not touch Decision 3.

Nothing here builds a general "Harness" cost component analogous to
`AbilityCost.Loyalty` — The Mind Stone's cost is an ordinary
`Plus(ManaCost, TapCost)`, because CR 701.64a is a printed EFFECT
("harness [this permanent]"), not a cost syntax, and no card prints
"Harness" as part of a larger cost the way a Class's level-up is
printed as part of an activation instruction.

---

## Amendment (2026-09-27): a chosen option is a gate too — the anchor-word Sieges (CR 614.12, #1572)

**Status:** Accepted · 2026-09-27 · tracked on
[#1572](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1572),
tracker [#889](https://github.com/krakenhavoc/cmd_and_ctrl/issues/889);
the proof card came from the Edea deck tracker
[#1565](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1565).

### Context

"As this enchantment enters, choose Khans or Dragons. • Khans — [ability]
• Dragons — [ability]" is the Siege cycles' shape (Fate Reforged's five,
Tarkir: Dragonstorm's five). The words are ANCHOR WORDS: each points at the
ability printed after it, and a Siege has exactly one of the two abilities —
the one after the word chosen as it entered (the Fate Reforged rulings, and
the same answer for a copy: it makes its own choice).

Two things were missing. The answer had nowhere to live: the as-enters family
had four per-permanent fields (`NamedTribe`, `ChosenColor`, `ChosenPlayer`,
`ChosenName`), each with its own validated vocabulary and its own writer, and
a word the CARD supplies fits none of them. And nothing could say "this
ability exists only when the word is Temur" — which is, word for word, what
this ADR's gate is for.

### Decision

**A sixth gate kind, not a card-side check.** `DesignationChosenOption`
joins the enum, `Designation` grows an `Option string` (the anchor word),
and `Active` answers `d.Option != "" && c.ChosenOption == d.Option`.
`game.ChosenOptionIs(word)` builds it. Everything Decision 1 promised follows
unchanged: the four accessors drop the unchosen line, so the harvester never
matches a Dragons trigger on a Khans Siege, the layer pass never sees a Temur
anthem on a Jeskai one, and a targeted trigger on the wrong line never asks
anybody to pick a target. The issue suggested a `ChosenOptionIs(src, opt)`
reader for card `AppliesTo`s instead; a reader would work for triggers but
leaves the unchosen static in the list the layer pass walks and puts the
same `if` in every card file, which is exactly the "if the Class is level 3
inside an Apply" this ADR ruled out.

**An empty answer matches no word.** Between entering and answering the
Siege has NEITHER ability. An open `PendingChoice` stops priority, so the
window is unobservable, and the direction is the safe one (weaker than
printed, never stronger).

**The answer is `Card.ChosenOption string`, the as-enters family's fifth
member**, with the family's lifecycle verbatim: per instance, cleared at both
CR 400.7 sites (`zone.go`'s battlefield exit and `resetAsNewObjectLocked`),
carried by clone and the snapshot (`chosenOption`, `carried` in
`snapshot_drift_test.go`; additive, so the v7 shape file was updated in place
and no schema bump), not a copiable value (CR 707.2 — `CopiableValuesOf`
never reads it). The one difference from its siblings is the vocabulary: the
options are the card's own words, so there is nothing to validate beyond "one
of the offered options".

**The prompt is the existing `option_pick`.** `Game.QueueChooseOptionAsEntersForEffect(chooser,
source, question, options)` queues one option per word, in printed order,
from the permanent's `AsEnters` hook (S26's declared simplification, fifth
use: queued as the permanent enters rather than by pausing the CR 614
pipeline). No option names a seat or a card, so nothing prunes the list while
it is open and the answered INDEX is stable — `Then`, not `ThenSeat`. A
prompt dropped unanswered runs with `NoChoiceIndex` and stores nothing. The
gate, `internal/legal`, the wire and the client modal answer it with no
change; the bot takes the first word (the enumerator's always-legal option).

**Announced and invalidating.** `setChosenOptionLocked` emits
`EventOptionChosen` (actor = the chooser, `Label` = the word). The layer
listener's designation arm bumps on it — the answer switches a gated static
on — and the public log narrates it as `choose_option` ("P1 chose Temur for
Frostcliff Siege"), `choice` redacted with the card's name like its
siblings'. `CardView.chosen_option` carries the word, public, cleared on the
non-knower redaction.

**Card side** (`cards/effects/choose_option.go`):
`ChooseOptionAsEnters(label, options...)` for `Spec.AsEnters`, `ChosenIs(word)`
for the gate, `WhenChosen(word, trigger)` / `StaticWhenChosen(word, static)`
to stamp it, and `ChosenOptionOf` for an effect that has to say the word.
`TestEveryAnchorWordGateIsOffered` runs every registered anchor-word card's
`AsEnters` and fails when a gate names a word the prompt never offers (that
line would be switched off forever) or an offered word gates nothing.

### Cards

Frostcliff Siege (the proof card: a gated combat-damage trigger and three
gated layer statics), Palace Siege, Citadel Siege, Barrensteppe Siege,
Outpost Siege and Frontier Siege — all `CompletenessFull`.

### Still not covered

- **Windcrag Siege.** Its Mardu line is a CR 603.2d trigger doubler, and
  `game.TriggerDoubler` has no `ActiveWhen`. The doubling query already hands
  over the doubler as a `Card`, so the gate field plus one `Active` check in
  the doubling pass is the missing piece — Decision 1 point 3's "same two
  lines on the day one does", for a fifth slot. Writing the word check inside
  the card's `Applies` instead would work and is exactly the per-card `if`
  this amendment exists to avoid.
- Glacierwood, Hollowmurk and Monastery Siege were not attempted here (the
  PR kept to six cards); their lines look expressible on existing machinery
  (`GatedCastPermissions`, a once-per-turn trigger plus "whenever you attack",
  and a target-reading `CostModifier`), unverified.
- Mana abilities and replacement effects still take no gate (Decision 1
  point 3); no Siege needs one.

---

## Amendment (2026-09-28): a sixth designation, Monstrous (CR 701.37, #1700)

**Status:** Accepted · 2026-09-28 · tracked on
[#1700](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1700),
tracker [#889](https://github.com/krakenhavoc/cmd_and_ctrl/issues/889).
Relates to [#1657](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1657)
(whose builder skipped Polukranos for want of this) and
[#1563](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1563) (divided
damage).

### Context

CR 701.37 is a keyword action that sets a designation:

> **701.37a** "Monstrosity N" means "If this permanent isn't monstrous, put
> N +1/+1 counters on it and it becomes monstrous."
> **701.37b** Monstrous is a designation that has no rules meaning other
> than to act as a marker that the monstrosity action and other spells and
> abilities can identify. Only permanents can be or become monstrous. Once
> a permanent has become monstrous, it stays monstrous until it leaves the
> battlefield. Monstrous is neither an ability nor part of the permanent's
> copiable values.
> **701.37c** If a permanent's ability instructs it to become monstrous,
> but that permanent is already monstrous, the ability does nothing. …
> Abilities that trigger when a permanent "becomes monstrous" … "X" refers
> to the number chosen for X in the monstrosity ability.

The engine had neither the marker nor an event, so two printed shapes could
not be written: "When this creature becomes monstrous, …" (Polukranos,
Stormbreath Dragon, Hydra Broodmaster, Ember Swallower, Arbor Colossus) and
"As long as this creature is monstrous, it has …" (Fleecemane Lion,
Domesticated Hydra, Hundred-Handed One). The second is exactly what this
ADR's gate is for; the first needs an event that carries X.

### Decision: Harnessed's three pieces, plus the counters and the X

- **`Card.Monstrous bool`** (`game/card.go`), in the bool block beside
  `Harnessed`. Every lifecycle rule is Harnessed's: not copiable
  (`CopiableValuesOf` never reads it — CR 701.37b says so in terms),
  cleared at both CR 400.7 sites (`zone.go`'s battlefield exit and
  `entry_tail.go`'s `resetAsNewObjectLocked`), carried by clone (`out := c`)
  and the snapshot (`cardSnapshot.Monstrous`, both projection sites,
  `carried` in `snapshot_drift_test.go`). Additive with a correct zero value
  ("not monstrous"), so the v7 shape file was updated in place with no
  schema bump, as #1572's `chosenOption` was.
- **`DesignationMonstrous`** is appended to the `DesignationKind` enum (no
  existing value moves), `Active` answers `c.Monstrous`, and
  `game.Monstrous()` builds the gate. Nothing else changed for the four
  accessors to honour it — Decision 1's argument, paid off a sixth time.
- **`Game.MonstrosityForEffect(cardID, n)`** is the keyword action and the
  one writer of the flag outside the snapshot restore. Three rules, each one
  line:
  1. **"Isn't monstrous" is checked on resolution.** A permanent that is
     already monstrous returns immediately — no counters, no event. There is
     no activation `Condition`: CR 701.37a is an *if* inside the effect, so
     a second activation is legal, and one activated in response to the
     first finds the creature monstrous by the time it resolves. A
     `Condition` would get the in-response case wrong.
  2. **The counters ride the CR 614 pipeline**
     (`AddCounterByThenForEffect`, placed by the permanent's controller), so
     Doubling Season and Hardened Scales apply, and a window holding both
     pauses on the CR 616 ordering prompt. The designation and the event are
     the placement's **continuation**, so nothing — no trigger, no gated
     static — sees the creature monstrous before its counters are on it. The
     continuation re-finds the card and re-checks the flag, because the
     pointer is not stable across the pause. The creature becomes monstrous
     even when the counters are replaced away entirely; the two halves of
     the instruction are not conditional on each other.
  3. **The event carries the announced N**, not the counters that landed:
     `EventBecameMonstrous` (Source / CardID / Target = the permanent,
     Actor = its controller, **Amount = N**). CR 701.37c and the Theros
     rulings make "X" in a becomes-monstrous trigger the X of the
     instruction, so a Doubling Season that doubles Polukranos's counters
     does not double its damage.

  A permanent not on the battlefield is a quiet no-op rather than an error
  (its ability's source left in response; CR 701.37b's "only permanents").
  A negative N reads as zero.
- **`EventBecameMonstrous` bumps the layer version** with the other
  designation events in `layer_listener.go`. This is load-bearing at
  X = 0: no counter is placed, so no `EventCounterPlaced` invalidates the
  pass, and without the bump a Domesticated Hydra made monstrous for
  {G}{G}{G} would not have trample
  (`TestDomesticatedHydraAtXZeroStillGainsTrample`). Silent on the public
  log (`silentBoardStateIsVisible`), for `EventHarnessed`'s reason: the
  counters are narrated, the designation is a badge.

**Wire:** `CardView.monstrous` (omitempty), read straight off a battlefield
permanent like `harnessed` — no card type owns monstrosity, so there is no
subtype probe. Public, and cleared on the non-knower (face-down) redaction
with the other designations. No client change in this amendment: the client
renders none of `harnessed`, `solved` or `monstrous` as a badge yet, and the
+1/+1 counters it does render are the visible half.

**Catalog side** (`cards/effects/monstrosity.go`):

| Printed | Constructor |
|---|---|
| "{cost}: Monstrosity N." | `Monstrosity(cost, n)` |
| "{cost}: Monstrosity X." | `MonstrosityX(cost)` — N is `item.XValue` |
| "When this creature becomes monstrous, …" | `WhenBecomesMonstrous(label, effect)` |
| "… X …" in that trigger | `MonstrosityXOf(item)` — the event's Amount |
| "As long as this creature is monstrous, it has …" | `MonstrousKeywords(kws...)` |

The activated ability's label is built from the cost ("{5}{R}{R}:
Monstrosity 3."), which is the printed line without reminder text, so
`TestAbilitiesMatchOracleText` holds by construction. Not sorcery speed.
Both activation constructors run `sourceIsNewObject` first (#1432), so the
ability of a creature that left and came back does not make the new object
monstrous.

**Polukranos's divided X.** #1563's `DivideSpec` refuses `FromX` on a
trigger, because a trigger announces no X. It does not need to: Polukranos's
clause comes from `TriggeredAbility.TargetsFrom`, which reads X off the
trigger context (`tc.Amount()`) and returns
`TargetCreature(…, OpponentControls()).WithCount(0, X).Dividing(Divide(X))`
— a fixed `Total` sized per trigger instance, which the CR 603.3d walk
settles exactly as it settles Inferno Titan's 3. "Any number" caps at X
because every target must be assigned at least one (CR 601.2d); X = 0
returns a nil clause, so the trigger targets nothing. `TargetsFrom` reads
only its trigger context, so the clause is re-derivable on restore and
`TargetsFromReadsBoard` stays false. This does not depend on #1699's
`AmountKey`; that PR's shape would also express it, and neither conflicts.

**Bots.** Monstrosity is an ordinary activation; `internal/legal` needs no
change. Polukranos and Hydra Broodmaster declare `XMatters` (X = 0 spends
the monstrosity for nothing), so the enumerator's X floor is 1 for them;
Domesticated Hydra takes an `xMattersAllowlist` entry, because X = 0 still
buys it trample. The enumerator will still offer a second activation on a
creature that is already monstrous — legal and useless, the same posture
Harness has.

**Roadmap:** a `monstrosity` mechanic entry (`internal/roadmap/registry.go`)
probed by the keyword action after a cost or the `Monstrous` gate.

### Cards

All eight are new. Seven ship `full`; one ships with a caveat.

| Card | What it proves |
|---|---|
| Polukranos, World Eater | Monstrosity X, the trigger's X, the X-sized divided clause, the fight-back from each still-legal target |
| Stormbreath Dragon | Monstrosity N and a becomes-monstrous trigger reading hand sizes on resolution; flying, haste, protection from white |
| Hydra Broodmaster | the trigger's X is the announced X under Doubling Season (X X/X tokens, not 2X) |
| Ember Swallower | a symmetrical edict, each seat asked three times in APNAP order |
| Arbor Colossus | a targeted becomes-monstrous trigger; no legal target, no prompt (CR 603.3d) |
| Fleecemane Lion | gated hexproof and indestructible |
| Domesticated Hydra | Monstrosity X and a gated keyword; X = 0 still switches it on |
| Hundred-Handed One | gated reach — **caveat:** "can block an additional ninety-nine creatures" has no shape (a blocker maps to one attacker), Brave the Sands' gap. *Lifted by #1706 ([ADR 0045](0045-combat-restrictions.md) Decisions 60–63): a gated `CanBlockAdditional(99)`, and the card is `full`.* |

### Still not covered

- **"Enters or becomes monstrous"** (Alpha Deathclaw, Protector of the
  Wastes) is one ability watching two events — expressible today
  (`Watches: {EventETB, EventBecameMonstrous}`); not attempted here to keep
  to eight cards. Protector's "controlled by different players" is a
  cross-target constraint that needs checking separately.
- **Cost reductions on the monstrosity ability** (Grim Giganotosaurus,
  Nemesis of Mortals) need an activated-ability cost modifier reading the
  board — unverified.
- **Monstrosity X where X is not announced** (Clay Golem's die roll,
  Maester Seymour's counter count) needs `MonstrosityX` to take its N from a
  rule rather than `item.XValue`; the engine half already takes any N.
- **Shipbreaker Kraken**'s "don't untap for as long as you control this
  creature" rides ADR 0058's duration, and **Sealock Monster**'s land-type
  change rides layer 4 — both unverified.
- The rest of the monstrosity cards are keywords plus the ability and are
  one file each.


## Amendment (2026-10-08): a seventh designation, Saddled — the one that lasts a turn (CR 702.171, #2695)

**Status:** Accepted · 2026-10-08 · tracked on
[#2695](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2695), under the
deck-request tracker [#2077](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2077).
Follow-ups: [#2704](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2704)
(cards that act on "creatures that saddled it") and
[#2705](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2705) (the Pilot
cards' "as though its power were 2 greater").
**No new ADR.** Saddle is a designation that switches abilities on, which is
this ADR's subject, and its cost is ADR 0020's crew cost over other creatures;
neither needed a decision of its own, so this amends the one that owns the gate.
No ADR number was taken.

**Rules text.** The CR text file was not available to this change, so the
sub-rule letters below are the ones #2695 quotes (702.171a the ability, 702.171b
the designation, 702.171c "creatures that saddled it this turn"), not a reading
of the pinned edition. Nothing here depends on a letter; re-check them against
the September 25, 2026 text when it is to hand.

### Context

Saddle N is Outlaws of Thunder Junction's Mount keyword: "Tap any number of
other untapped creatures you control with total power N or more: This permanent
becomes saddled until end of turn. Saddle only as a sorcery." Cards then read
"whenever this creature attacks while saddled" and "as long as it's saddled".
About thirty Commander-legal cards print it, and `ability_rows.go` already named
`saddle` as a keyword activated ability while nothing implemented it. Guidelight
Matrix shipped without its Mount half, with a caveat saying so.

Crew already pays "tap creatures with total power N or more" and is the template
for the cost. Saddled is the new part, and it differs from every designation
this ADR has so far:

| Designation | Lasts |
|---|---|
| class level, solved, harnessed, monstrous, Ring-bearer | until the permanent leaves the battlefield |
| station charge counters | while the counters are there |
| **saddled** | **until end of turn**, or until it leaves the battlefield |

### Decision: Monstrous's pieces, swept at cleanup

- **`Card.Saddled bool`** and **`Card.SaddledBy []ObjectRef`** (`game/card.go`).
  It is a Card field and not a scoped effect for the reason every designation is:
  `Designation.Active` reads a `Card` and nothing else, so "as long as it's
  saddled" can only be answered off the object. Not copiable, cleared at both
  CR 400.7 sites (`zone.go`, `entry_tail.go`), carried by clone and the snapshot
  (`cardSnapshot.Saddled` / `SaddledBy`, additive within v7 — the shape file was
  updated in place, no bump; `carried` in `snapshot_drift_test.go`).
- **`DesignationSaddled`** is appended to the enum, `Active` answers
  `c.Saddled`, and `game.SaddledGate()` builds the gate (`effects.Saddled()`).
  Nothing else changed for the four accessors to honour it, which is this ADR's
  Decision 1 paying off a seventh time.
- **`Game.SaddleForEffect(cardID, saddlers)`** is the one writer. It does
  nothing for a permanent that is not a Mount (`MountSubtype`), so Alacrian
  Armory's "becomes saddled if it's a Mount" is the primitive itself. It emits
  **`EventBecameSaddled`** only when the Mount was not already saddled: "for
  the first time each turn" (Stubborn Burrowfiend) is therefore the event's own
  shape and needs no per-turn counter. A second saddle in a turn still records
  its saddlers. The event bumps the layer version with the other designation
  events (`layer_listener.go`) and is silent on the public log
  (`silentBoardStateIsVisible`).
- **The sweep.** `sweepTurnEndLocked` (the CR 514.2 cleanup sweep) calls
  `clearSaddledLocked`, which clears the flag and the saddlers on the
  battlefield and the phased-out zone and bumps the layer version when anything
  was cleared, so a gated static switches off with the designation. #2695 also
  says the designation ends when the Mount phases out, so `phaseOutLocked`
  clears it too, unlike the other designations, which ride through a phase-out
  (CR 702.26d). The sweep of the phased-out zone is only a backstop. If the
  pinned rules turn out to let a saddled Mount stay saddled through a phase-out,
  that one clear is the line to remove; it errs weaker, never stronger.
- **`SaddledBy` and `Game.SaddlersOf`** answer "creatures that saddled it this
  turn". The record is the creatures tapped to pay for the saddle ability that
  resolved, as objects (instance ID and `ObjectEpoch`); `SaddlersOf` returns
  the ones still on the battlefield as the same objects, so a creature that left
  or came back is dropped (CR 400.7). It is replaced, never appended in place,
  because clone shares the backing array.

### Decision: the cost is crew's, over other creatures

`AbilityCost.Saddle int` sits beside `Crew` rather than being a flag on it,
because the two differ in the one thing the validator checks (the source is not
a legal payment) and in everything a probe or a card wants to read. Both go
through `validateCrewCostLocked`, which now takes the source: a Saddle cost
refuses the source anywhere it is named (`ErrInvalidParam`), then runs crew's
walk unchanged. Controller, untapped, a creature, no repeats, **no
summoning-sickness check** (tapping to saddle is not a {T} cost), power read at
payment from the post-layer value, the printed number a floor. The creatures
ride the existing `crew_ids`. Their tap is paid in the crew loop, and the
creatures are written onto the item's payment record
(`PaidCost.TappedOthers`, station's field, which already survives clone, undo,
a CR 707.10 copy and a restore) so the saddle effect can read who paid at
resolution. `effects.Saddle(n)` is the whole ability: label `Saddle N`,
sorcery timing (`SorcerySpeed`, CR 602.5d), and `saddleEffect`, which does
nothing for a Mount that left and came back (#1432). `Register` refuses a
saddle cost on an any-player ability, as it does crew.

### Decision: wire, enumerator, bots

- **Wire.** `CardView.saddled` (public, omitempty, cleared on the face-down
  redaction). A saddle ability rides crew's two fields, `crew_cost` (the saddle
  number) and `crew_options` (untapped creatures the controller controls with
  the Mount left out), plus `activated_abilities[i].saddle: true`, so the
  client's one picker, the `crew_ids` payload and the legal-move lookup need no
  second shape. Documented in `docs/protocol.md`.
- **Enumerator.** `crewPayment` takes the creature to leave out and the Mount is
  passed for a saddle cost, so a Mount with nothing else to tap is offered no
  activation (#544) and an offered one names a set the engine accepts. Sorcery
  timing is the ability's, which `legal` already judges. Bots need nothing
  else: saddling is an ordinary activation, and a Mount's attack payoff is the
  heuristic's to weigh like any trigger.
- **Client.** The designation badge slot gains SADDLED (`Card.svelte`), and
  `CrewCostModal` words the picker "Saddle" when the ability says so. The
  context menu's disabled reason says "no other untapped creatures to saddle
  with".

### Decision: read "isn't saddled" when the trigger is built

Caustic Bronco pays differently if it "isn't saddled", judged as the trigger
resolves. A Mount removed in response is judged as it last existed (CR 608.2h),
and no last-known-information record carries the designation. Because saddle is
sorcery speed the designation cannot change while the trigger waits, so
reading it when the trigger is built is the same answer, and `item.Params`
carries it. A card whose condition can change in the window should not copy
this.

### Cards

Thirteen ship, twelve `full` and one with a caveat. They prove the cost and the gate
(every Mount), the designation set from outside (Guidelight Matrix), the first
saddle of a turn (Stubborn Burrowfiend), and "creatures that saddled it"
(Rambling Possum).

| Card | What it proves |
|---|---|
| Guidelight Matrix | `BecomeSaddled` from an activated ability, sorcery speed, a Mount-only target; its caveat is cleared |
| Gilded Ghoda | the base shape: `Saddle(1)` and an attack trigger that reads the designation |
| Drover Grizzly | a group keyword grant fixed at resolution |
| Seraphic Steed | a token on a saddled attack |
| Gloryheath Lynx | a library search from the trigger |
| Bounding Felidar | counters on each other creature, life for each once they settle |
| District Mascot | an entry counter, a counter-removal cost, a trigger on the Mount itself |
| Bulwark Ox | a targeted saddled-attack trigger and a group grant behind a predicate |
| Ornery Tumblewagg | a beginning-of-combat trigger beside a saddled one; doubling counters |
| Caustic Bronco | the designation carried on the item (above) |
| Stubborn Burrowfiend | `WhenBecomesSaddled`, once a turn |
| Rambling Possum | `SaddlersOf` through a card-set prompt with a floor of zero |
| Guardian Sunmare | ward and a nonland search; **caveat:** an Aura fetched this way enters unattached (Zur the Enchanter's gap) |

### Still not covered

- **Cards that act on a specific saddler** — a target restricted to the
  saddlers (Giant Beaver), a sacrifice that reads the saddler's power (The
  Gitrog), a tapped-and-attacking copy (Calamity), a flicker with a chosen
  saddler at end of combat (Fortune) — are #2704. The data is built.
- **"Saddles Mounts and crews Vehicles as though its power were 2 greater"**
  (the Pilot token and Pilot creatures) and Interface Ace's toughness rule are
  #2705. They need a per-creature contribution that the payment, the picker's
  running total and the enumerator all read.
- **An event for "whenever this creature saddles a Mount or crews a Vehicle"**
  (Canyon Vaulter, Reckless Velocitaur) is part of #2705.
- **The remaining Mounts** (Gila Courser, Dracosaur Auxiliary, Lagorin,
  Congregation Gryff, Brightfield Mustang, Unswerving Sloth, Bridled Bighorn,
  Autarch Mammoth, Venomsac Lagac, Trained Arynx, Brightfield Glider, Quilled
  Charger, Alacrian Jaguar, Archmage's Newt) and the two support cards
  (Kolodin, Alacrian Armory) need nothing the seam lacks; they are one file
  each.
## Amendment (2026-10-08): a designation that gives abilities, Suspected (CR 701.60, #2698)

**Status:** Accepted · 2026-10-08 · tracked on
[#2698](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2698), S58 deck
requests (Barbed Servitor, Night - Sauron The Slayer,
[#2062](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2062)).
No new ADR number: this extends the designation family this ADR owns. The
rule lettering below (701.60a, c, d) is the issue's; the CR text file is not
in the repository, so it was not checked against the September 2026
edition.

### Context

> **701.60a** Some spells and abilities instruct a player to suspect a
> creature. That creature becomes suspected until it leaves the battlefield or
> a spell or ability causes it to no longer be suspected.
> **701.60c** A suspected permanent has menace and "This creature can't
> block" for as long as it's suspected.
> **701.60d** A suspected permanent can't become suspected again.

Suspected is a designation, but the six before it all do the same thing: they
switch a permanent's OWN printed abilities on, and a Spec declares the gate.
Suspected does the opposite. It gives ANY creature two things its card never
prints. About nineteen Commander-legal cards suspect creatures or read
"suspected", and none of them can say what suspected does, because it is not
on any Spec.

### Decision: Monstrous's lifecycle, a continuous effect of its own

- **`Card.Suspected bool`** (in the bool block) and **`Card.SuspectedAt
  int64`** (beside the other timestamps; a lone `int64` after the bool run
  strands padding, `TestCardHasNoInteriorPadding`). Monstrous's lifecycle in
  every respect except one: not copiable (`CopiableValuesOf` never reads it),
  cleared at both CR 400.7 sites, carried by clone and the snapshot (additive
  to the v7 shape, `carried` in `snapshot_drift_test.go`). The one difference:
  it is **kept through a control change**. Caught Red-Handed steals a
  creature and suspects it, and the creature goes home still suspected.
- **No `DesignationKind`.** The gate exists to hang on a printed ability, and a
  suspected creature's menace is on none. The effect is built where the
  keyword counters are (`keyword_counters.go`): `suspectContinuousEffectsLocked`
  adds ONE source-less layer-6 `ContinuousEffect` per suspected permanent to
  `activeStaticAbilitiesLocked`'s gather. Source-less means CR 613.6's
  silencing never reaches it: the designation is not an ability of the
  permanent. Its `Apply` appends `menace` (`AppendKeywordAbility`, so a
  creature that already has it keeps one) and ORs in `CantBlock`.
- **The timestamp is the moment it became suspected** (`SuspectedAt`, stamped
  by `SuspectForEffect`; a restore from before the field falls back to the
  permanent's own timestamp). A "loses all abilities" that is older leaves the
  menace; one that is newer takes it away. Both are tested, and both are
  CR 613.7.
- **`CantBlock` is a restriction bit, and that is stricter than a literal
  reading.** Restrictions have no layer and nothing clears them
  (`restrictions.go`), so a suspected creature that later loses all abilities
  still cannot block, where the rule's wording makes "can't block" an ability
  that would go with the rest. Chosen on purpose: it errs toward the
  restriction, never toward a creature blocking when the table expected it not
  to, and it lets the block gate, the enumerator and the view stay as they
  were. Recorded in `game/suspect.go`.
- **`Game.SuspectForEffect(id)`** reports whether it suspected. It refuses a
  permanent that is not on the battlefield, one that is not a creature, and
  one already suspected (CR 701.60d; this is also what keeps `SuspectedAt`
  from being rewritten), and bumps the layer version itself rather than
  emitting an event: no shipped card triggers on becoming suspected, and an
  event kind is a log-gate entry, a wire doc line and a replay surface for
  nothing yet. The first "whenever a creature becomes suspected" adds
  `EventSuspected` and moves the bump onto it, as `EventBecameMonstrous` did.
  **`UnsuspectForEffect`** and **`IsSuspected`** complete the set.
- **`PermanentInfo.Suspected`** keeps the designation in last-known
  information. Agency Coroner's "if the sacrificed creature was suspected" is
  read after the cost has put the creature in a graveyard, where the flag is
  gone; `Context.SacrificedPermanent()` carries it.

**Wire:** `CardView.suspected` (omitempty), read straight off a battlefield
permanent. Public, and **kept on a face-down permanent**, unlike Monstrous:
Monstrous names an ability the hidden card has, and Suspected names something
that was done to the object in front of the table, and is the reason the
creature cannot block. The client renders it as `SUSPECTED` in the existing
designation badge slot, at the head of its priority chain, because it is the
one designation that changes what the creature may do right now. The bot's
board text says `suspected` beside the `menace` already in the ability list.

**Catalog side** (`cards/effects/suspect.go`):

| Printed | Constructor |
|---|---|
| "suspect it" / "suspect this creature" | `Suspect{Target: ctx.Source()}` |
| "suspect up to one target creature" | `Targeting(…, UpToOneTargetCreature(…))` with `SuspectEachLegalTarget` |
| "it's no longer suspected" | `Unsuspect{Target: id}` |
| "all suspected creatures are no longer suspected" | `UnsuspectAll{}` (`Match` narrows it) |
| "suspected creatures" in a target clause | `Suspected()` / `NotSuspected()` |
| "Sacrifice a suspected creature" | `SacrificeASuspectedCreature()` |
| "suspect enchanted creature" | `suspectEnchantedCreature` |

`Suspect` and `Unsuspect` run `isNewSourceObject` (#1432) like every primitive
that names a source. A card whose ability source IS the object that came back
(Presumed Dead's granted dies trigger) calls the game mutator directly.

**Bots.** Nothing new to enumerate: a suspected blocker is not offered a block
and a suspected attacker needs two blockers, both through `BlockOptionsLocked`
(`legal/suspect_test.go`). The board text gains the word.

### Cards

Nineteen. Fifteen ship `full` (Person of Interest, Rune-Brand Juggler,
J. Jonah Jameson, Rubblebelt Braggart, Repeat Offender, Clandestine Meddler,
Absolving Lammasu, Agrus Kos, Eliminate the Impossible, Caught Red-Handed,
Convenient Target, Case of the Stashed Skeleton, Reasonable Doubt, Agency
Coroner, Deadly Complication). Barbed Servitor carries Brash Taunter's
"damage from two sources is reflected separately" caveat; It Doesn't Add Up and
Presumed Dead carry "an entry that asks a question comes back unsuspected";
Incriminating Impetus carries Shiny Impetus's goad caveat.

### Still not covered

- **Frantic Scapegoat** chooses one of the creatures that entered together.
- **Nelly Borca** needs a batched "one or more creatures an opponent controls
  deal combat damage to one or more of your opponents".
- **Hot Pursuit** binds a goad to the one creature it suspected and gates a
  take-control trigger on two lost players.
- **Airtight Alibi** says "can't become suspected", which is a restriction on
  the suspect action that `SuspectForEffect` does not yet read.
