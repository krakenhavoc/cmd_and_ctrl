# ADR 0046 — Layer 6 is authoritative: the catalog ability lookups read the layered result

**Status:** Accepted · 2026-09-14 · follows [ADR 0012](0012-layer-system.md) and [ADR 0039](0039-layer-4-authoritative.md)
**Amended:** 2026-09-16 (the [#159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/159) closeout). Dated notes at the end of §4 and §5: the application pass silences more than CR 613.6 allows, and Song of the Dryads removes more than CR 305.7 does. [#669](https://github.com/krakenhavoc/cmd_and_ctrl/issues/669) tracks the fix, which needs [#675](https://github.com/krakenhavoc/cmd_and_ctrl/issues/675) (CR 704.5p) first.

## Context

[ADR 0012](0012-layer-system.md) shipped the CR 613 layer engine, and
layer 6 — ability-changing effects — worked from day one for the thing
it was built for: `Characteristic.Abilities` is a list of keywords, a
lord appends to it, and `Card.HasKeyword` reads it back.

Ability *removal* is the other half of layer 6 and it did not work at
all. `Characteristic.Abilities` could be emptied, and emptying it
removed the keyword badges and nothing else, because **the abilities
that matter are not in that slice**. Every activated, triggered, mana,
static and replacement ability in the engine is reached through a
`Catalog*` hook keyed by oracle ID, consulted at USE time from
wherever the use happens:

| Hook | Read from | Answers |
|---|---|---|
| `CatalogTriggers` | `harvestFromZone`, `harvestCastFromStack`, `harvestLTB`, `harvestSimultaneousExitLocked`, `SagaFinalChapter` | triggered abilities |
| `CatalogActivatedAbilities` | `ActivatedAbilitiesForCard` | activated abilities |
| `CatalogManaAbilities` | `ManaAbilitiesForCard` | mana abilities |
| `CatalogStaticAbilities` | `activeStaticAbilitiesLocked` | continuous effects |
| `CatalogReplacements` | `gatherActiveReplacementsLocked`, `ReplacementLabelFor` | replacement effects |
| `CatalogCostModifiers` | `activeCostModifiersLocked` | cost taxes and discounts |
| `CatalogUntapStepPermissions` | `activeUntapStepPermissionsLocked` | "untap during each other player's untap step" |
| `CatalogNoMaxHandSize` | `EffectiveMaxHandSizeLocked` | "you have no maximum hand size" |

Those hooks answer *what the printed card says*, which is the right
question almost everywhere and the wrong question under a layer-6
removal. A Darksteel Mutation'd Birds of Paradise still tapped for
mana; a Song of the Dryads'd Mind Control still stole the creature; a
Kenrith's Transformation'd commander still had its dies-trigger.

Issue [#76](https://github.com/krakenhavoc/cmd_and_ctrl/issues/76)
carried this as a named machinery gap blocking three Commander
staples. It is the same shape ADR 0039 fixed for layer 4: the layer
engine computed an answer and nothing in the engine read it.

## Decisions

### 1. `CatalogAbilityKey` is the seam, and there is exactly one of it

`game.CatalogAbilityKey(c Card) string` is the key to use when asking
what abilities an object has right now. It is `CatalogKey` until a
layer-6 removal applies to the card, and `""` after — and every hook
already treats an empty key as "no catalog entry", so converting a
reader is a one-token change with no other edit.

`CatalogKey` stays as the printed-value surface, exactly as ADR 0039
kept `PrintedIsCreature` alongside `IsCreature`. Naming the two
surfaces means a call site's choice between them is a rules decision
visible in the diff that makes it.

**Why one accessor and not a check at each reader.** The engine
declares **nineteen** `Catalog*` hooks. **Eight** of them answer an
ability question — the eight in the table above — and they are read
from **thirteen** call sites. Patching thirteen call sites would have
fixed thirteen call sites, and a fourteenth added next sprint would
have reintroduced the bug for free. That is the reasoning
[#539](https://github.com/krakenhavoc/cmd_and_ctrl/pull/539) used to
replace six hand-rolled zone-change movers with one exit primitive and
[#548](https://github.com/krakenhavoc/cmd_and_ctrl/pull/548) used for
untapping, and it applies here for the same reason: the defect is that
the question was asked in nine places, not that nine places got the
answer wrong.

Two of the thirteen answer from somewhere else, and the exception is
documented where it sits: `harvestLTB` reads
`Characteristic.AbilitiesRemoved` off the CR 603.10 last-known-
information snapshot, and `harvestSimultaneousExitLocked` off the copy
its batch captured, because by the time either runs the permanent has
left the battlefield and its layer cache has been cleared. A creature
that died under a Kenrith's Transformation has no dies-trigger, and
that has to stay true for the beat between the death and the Aura
falling off. A third, `activeStaticAbilitiesLocked`, is inside the
recompute and takes the silenced set as a parameter — there is no
layered result to read yet when it runs.

The other **eleven** hooks — `CatalogTargetSpec`, `CatalogModeSpec`,
`CatalogTargetMode`, `CatalogAdditionalCost`,
`CatalogAlternativeCosts`, `CatalogCastableZones`,
`CatalogCantBeCountered`, `CatalogTapPermanentsCost`,
`CatalogStartingLoyalty`, `CatalogBattleDefense`,
`CatalogPrintedKeywords` — are deliberately left on `CatalogKey`. None
of them asks about an ability of a battlefield permanent: they are
cast-time properties of a spell (which has no layered characteristic),
printed characteristics, or, in `CatalogPrintedKeywords`' case, the
layer-0 baseline that the removal is applied *to*.

### 2. `Characteristic.AbilitiesRemoved` is a bool, not a timestamp

CR 613.6 says an ability granted by an effect with a later timestamp
than the removal still applies. The two halves of that are handled in
different places because they are different things:

- **Granted keywords** live in `Characteristic.Abilities`. The layer-6
  bucket is already sorted by timestamp; a removal empties the slice
  in its own slot and a grant sorted after it appends to the emptied
  slice. Rancor attached after a Darksteel Mutation gives the Insect
  trample; attached before, it does not. **No new code** — this is the
  existing sort doing the work, and it is pinned in both directions.
- **Printed abilities** — everything the catalog answers for — are
  part of the object rather than granted to it, and a removal that
  applies to the object removes them whichever arrived first. There is
  no ordering question left, so there is nothing for a timestamp to
  answer.

An effect that removes and grants in the same breath (Darksteel
Mutation's "has indestructible, and it loses all *other* abilities")
is **one** effect: `LoseAllAbilities("indestructible")`, not a removal
plus a separate `GrantToAttached`. Two same-timestamp effects would
resolve by gather order, which is not a thing a card file should have
to reason about.

### 3. The removal is a declaration on the static, not something `Apply` does

`game.StaticAbility.RemovesAbilities bool`, and a matching method on
the `ContinuousEffect` interface. The engine empties
`Characteristic.Abilities` and sets `AbilitiesRemoved` before calling
`Apply`.

Declaring it rather than letting `Apply` clear the slice is
load-bearing twice over. A card file *cannot* ship a removal the
engine does not know about, which was the exact failure mode being
fixed. And the recompute needs the set of silenced permanents to drop
their contributions from **every other layer**, which it cannot learn
by watching an opaque closure mutate a struct.

### 4. The recompute iterates to a fixed point, capped

An ability-removing effect changes which continuous effects exist,
which is CR 613.8 dependency territory — and ADR 0012 declined
dependency *analysis*, a position this ADR does not reopen. It
resolves this one case by **iteration** instead:

1. **Discovery pass** — a full layer pass with nothing silenced,
   which learns which permanents a layer-6 removal applied to. On
   every board that has never seen one of these Auras this set is
   empty, the loop ends, and the cost is one extra battlefield walk.
2. **Application pass** — the same pass with the silenced set held out
   of the *gather*, so a silenced permanent stops contributing to
   layers 1-5 and 7 as well as to 6. This is what makes a Song of the
   Dryads on a Mind Control give the creature back: layer 2 runs long
   before layer 6 could have told it to.

Capped at four passes (`maxLayerPasses`), keeping the last pass's
answer rather than spinning.

Within layer 6 the ordering between two removers is resolved properly
rather than by iteration: `applyLayerLocked` skips an effect whose
source has already been silenced earlier in the same timestamp-sorted
bucket. Two Song of the Dryads each enchanting the other settle on
"the earlier one wins", which is what the rules say and not what a
naive fixed point would produce.

**Note, 2026-09-16: the application pass silences too much, and its
Mind Control reasoning is wrong.** CR 613.6: "If an effect starts to
apply in one layer and/or sublayer, it will continue to be applied to
the same set of objects in each other applicable layer and/or
sublayer, even if the ability generating the effect is removed during
this process." Ability removal happens in layer 6. It can't reach back
into layers 1-5, and it doesn't stop an effect that already started
applying from carrying on into layer 7. Magus of the Moon's ruling
(2021-03-19): "If Magus of the Moon loses its abilities, it continues
to turn nonbasic lands into Mountains." Holding a silenced source out
of the *gather* drops it from every layer, which is the opposite.

No catalog card shows it yet. Every silenced source today is under an
Aura that also overwrites its types, and every layer 1-5 static on a
creature applies only to itself. Magus of the Moon under Kenrith's
Transformation or Darksteel Mutation would show it, which is why
AGENTS.md §7 holds Magus back until
[#669](https://github.com/krakenhavoc/cmd_and_ctrl/issues/669) lands.

The Song of the Dryads on a Mind Control example gets the right answer
for a different reason. Song makes the Mind Control a Forest land, so
it's no longer an Aura. A noncreature, nonbattle permanent that isn't an
Aura, Equipment or Fortification becomes unattached as a state-based
action (CR 704.5p; Song's 2014-11-07 ruling: "If the permanent was an
Aura or an Equipment, it becomes unattached from whatever it was
attached to."). With nothing enchanted, "you control enchanted
creature" applies to nothing. Layer order has nothing to do with it.
The engine doesn't implement CR 704.5p yet
([#675](https://github.com/krakenhavoc/cmd_and_ctrl/issues/675)): the
Forest stays attached, and today only the gather-time silence hands
the creature back. So #675 has to land before or with #669's fix.

The fixed point also isn't the only CR 613.8 dependency the catalog
can build. Layer-4 pairs (Urborg, Tomb of Yawgmoth + Song of the
Dryads, Maskwood Nexus + a crewed Vehicle, and others) are listed in
[ADR 0043](0043-copy-effects.md) §5's amendment and tracked in
[#668](https://github.com/krakenhavoc/cmd_and_ctrl/issues/668).

### 5. What survives ability removal

CR 613.1f removes abilities. Stated once, because the boundary is the
part that is easy to get wrong in either direction:

**Removed** — activated abilities (including an Equipment's equip and
a token's own intrinsic ability), mana abilities, triggered abilities,
static abilities, replacement effects, cost modifiers, untap-step
permissions, "no maximum hand size", and keyword abilities.

**Kept** — the object. It is still a permanent, still on the
battlefield, still has its name (layer 3), its owner, its controller,
its counters, its power and toughness (layer 7, applied after 6), the
types layers 1-4 gave it, and every attachment in both directions.
Ability removal is not a zone change, not an unattach, and not a
counter removal — a planeswalker turned into a Forest by Song of the
Dryads keeps its loyalty counters and gets them back when the Song
leaves.

**Also kept: abilities an object has because of what it *is*.**
CR 305.6's intrinsic mana ability of a land with a basic land type is
granted by the land type, not printed in the rules text, so
`ManaAbilitiesForCard` gates the *declared* half and keeps the
*intrinsic* half. That is the whole of CR 305.7 for Song of the
Dryads: the Sol Ring's "{T}: Add {C}{C}" goes and the Forest's
"{T}: Add {G}" arrives, from the same effect, in the same recompute.

**Note, 2026-09-16: it isn't the whole of CR 305.7.** The rule removes
"all abilities generated from its rules text" and adds: "Note that this
doesn't remove any abilities that were granted to the land by other
effects." Song of the Dryads' ruling (2014-11-07) says the same: the
permanent "will still have any abilities it gained from other effects."
Song is built as a full layer-6 `LoseAllAbilities()`, so it also
removes abilities that other effects granted before it attached.
Reproduced on develop `7c1ae9b`: a Sol Ring given indestructible by
Boros Charm loses it when Song of the Dryads resolves on it, so a
wrath destroys a Forest the rules say survives. The loss is part of
the type-changing effect, so it belongs in layer 4 next to the type
change and should remove only the permanent's own rules text.
[#669](https://github.com/krakenhavoc/cmd_and_ctrl/issues/669) tracks
the fix, and [#644](https://github.com/krakenhavoc/cmd_and_ctrl/pull/644)
declares the gap on the card until then.

### 6. `internal/legal` needed no change, and that is the point

`ActivatedAbilitiesForCard` and `ManaAbilitiesForCard` are the
accessors the legal-move enumerator, the wire projection
(`protocol.CardView`), the lobby's ability lookup, the auto-tapper and
`ActivateCatalogAbility` all already read. Putting the check inside
them means the enumerator and the engine cannot disagree.

They must not disagree.
[#544](https://github.com/krakenhavoc/cmd_and_ctrl/issues/544): a seat
that owes a decision is enumerated that decision's answers and nothing
else, so a move list the engine refuses is not a wasted turn — it is a
**hung table**, with a deterministic policy retrying the same rejected
move forever and holding every other seat with it. The invariant is
stated as a test in `legal/ability_removal_test.go` rather than left
as a property of the call graph.

## Consequences

- Three Commander staples ship: **Darksteel Mutation**, **Kenrith's
  Transformation**, **Song of the Dryads**.
- Kenrith's Transformation is the catalog's **first layer-5 effect**.
  Layer 5 has had a bucket in the recompute since S16 with nothing to
  put in it.
- CR 704.5m and CR 704.5n now fire in real play rather than only in
  fixtures. #511 fixed 704.5m's "attached to nothing" disjunct for
  Auras an effect put onto the battlefield without a host; Song of the
  Dryads reaches the *first* disjunct from an ordinary cast — the host
  stops being a creature, so its Equipment unattaches (704.5n) and its
  "enchant creature" Auras go to the graveyard (704.5m), in the same
  settling.
- The coverage guard learns the mechanic: a caveat naming "loses all
  abilities" is now checked against an exact probe
  (`StaticAbility.RemovesAbilities` on the same spec), so the next
  card file that declares the gap while implementing it fails the
  build.
- **Not done.** A *duration-scoped* ability removal ("loses all
  abilities until end of turn", "for as long as you control this
  creature") is not here. The turn-scoped static registry from S32
  already carries `StaticAbility` values, so the layer-6 shape drops
  straight in; the "for as long as" duration does not exist at all.
  Kitesail Larcenist needs the second one.
