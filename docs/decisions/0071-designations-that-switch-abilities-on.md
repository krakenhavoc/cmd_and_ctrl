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
source's current level and demands exactly `level-1` (CR 716.2e). Nothing else
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
