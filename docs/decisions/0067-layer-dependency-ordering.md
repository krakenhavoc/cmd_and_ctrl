# ADR 0067 — Layer dependency ordering (CR 613.8) and effects that outlive their source (CR 613.6)

**Status:** Accepted · 2026-09-18 · S38 ("Layers: dependency ordering, ability grants, ability removal") · tracked on [#668](https://github.com/krakenhavoc/cmd_and_ctrl/issues/668), [#669](https://github.com/krakenhavoc/cmd_and_ctrl/issues/669) and [#670](https://github.com/krakenhavoc/cmd_and_ctrl/issues/670).
**Numbering:** checked on 2026-09-18 against every remote branch after
`git fetch origin` (the AGENTS.md §4 loop over `git branch -r` +
`git ls-tree`). `0061`-`0064`, `0066` and `0069` exist; `0065`, `0067`
and `0068` do not. 0067 is taken here because 0065 and 0066 were
claimed by sprint-mates on the same day.
**Supersedes:** [ADR 0043](0043-copy-effects.md) §5's deferral of
dependency detection, and with it [ADR 0012](0012-layer-system.md)'s.
**Amends:** [ADR 0046](0046-layer-6-authoritative.md) §4 (the
ability-removal fixed point is gone, and the gather no longer silences
anything) and §5.
**Builds on:** [ADR 0039](0039-layer-4-authoritative.md) (layer 4 is
authoritative), [ADR 0046](0046-layer-6-authoritative.md) (layer 6 is),
[ADR 0036](0036-attachments.md) and
[ADR 0063](0063-durations-and-control.md).

## Context

CR 613 applies continuous effects in seven layers, and within a layer
CR 613.7 orders them by timestamp. `server/internal/game/layers.go`
has done exactly that since S16, and two things it did around that
were wrong.

**CR 613.8 was not implemented at all.** ADR 0012 and ADR 0043 §5
deferred dependency detection on the explicit grounds that "the
catalog contains no pair that can produce one". That premise died on
2026-09-13/14, when Maskwood Nexus (#404), Arixmethes, Slumbering Isle
(#480) and Song of the Dryads (#563) merged within a day of each
other. Five layer-4 pairs came out differently depending on which card
entered first:

| pair | timestamp-only result | rules result |
|---|---|---|
| Urborg, Tomb of Yawgmoth (older) + Song of the Dryads on a Sol Ring | `[Forest]`, {G} only | a Forest Swamp, {G} and {B} |
| Urborg (older) + a slumbering Arixmethes | no land subtypes, {G}{U} only | a Swamp too |
| Maskwood Nexus (older) + a Vehicle crewed later | not every creature type: no lord pump, no haste | every creature type |
| Maskwood Nexus (older) + The Warring Triad under eight graveyard cards | not a creature, yet still every creature type | not every creature type |
| Maskwood Nexus (older) + a slumbering Arixmethes | a land that is every creature type | not every creature type |

All five are the same shape: a type-ADD whose "applies to" reads a
card type or subtype, sharing layer 4 with a type-SET that writes one.
Crew is the ordinary case rather than the exotic one — the crew
effect is stamped when the ability resolves, so a Vehicle crewed with
a Nexus already out was wrong for all six catalogued Vehicles.

**CR 613.6 was implemented backwards.** ADR 0046 §4's recompute
iterated to a fixed point and, on the second pass, held a silenced
permanent out of the GATHER — so a source that lost its abilities in
layer 6 contributed nothing to layers 1-5 or 7 either. CR 613.6 says
the opposite: an effect that has started to apply "will continue to be
applied to the same set of objects in each other applicable layer …
even if the ability generating the effect is removed during this
process". Magus of the Moon's ruling (2021-03-19) is the sentence in
one card: "If Magus of the Moon loses its abilities, it continues to
turn nonbasic lands into Mountains."

**Song of the Dryads removed more than CR 305.7 removes.** Its loss
was modelled as a full layer-6 `LoseAllAbilities`, so a Sol Ring that
Boros Charm had made indestructible lost indestructible when the Song
landed on it, and a wrath destroyed a Forest the rules leave alive.

**"Is every creature type" was stored in the wrong place.** It rode
the `changeling` keyword in `Characteristic.Abilities`, so a layer-6
ability removal deleted a layer-4 type fact (against Maskwood Nexus'
2021-02-05 ruling) and a layer-4 subtype set failed to overwrite it
(against the same card's other ruling).

Everything below is one pass over `layers.go`, because the four are
one mechanism seen from four sides.

## Decision 1 — CR 613.8 detection is a trial application, judged on the instruction

`server/internal/game/layer_dependency.go` owns the whole of it.

**Detection.** Effect A depends on effect B, in the same bucket, when
applying B changes either

1. **what A applies to** — `A.AppliesTo(T)` differs between the world
   before B and the world after it; or
2. **what A does to the things it applies to** — A's `Apply`, run
   against **the same starting characteristic** in both worlds,
   produces a different result.

The fixed base in (2) is the load-bearing part and the reason this is
not an output diff. Urborg on a Command Tower that a Song has already
turned into a Forest produces a different characteristic than it would
have on the untouched Tower — and Urborg does not depend on the Song
there, because the Tower was a land either way and "append Swamp" is
the same instruction. CR 613.8a asks what the effect DOES, not what it
was done to; running A against an identical base in both worlds is how
the engine asks that question. `TestUrborgAndSongOnAnExistingLandStayIndependent`
is the assertion, and it is the test an output-diff detector fails.

**Ordering.** CR 613.8b and 613.8c, literally: take the first effect
in the timestamp-ordered list that depends on nothing else still
waiting, apply it, and recompute the relation. When every remaining
effect depends on another — a loop — CR 613.8c says to abandon
dependency for that bucket, and index 0 (the earliest timestamp) is
that. `TestDependencyLoopFallsBackToTimestampOrder` pins the
termination.

**Why not a declaration on the card.** ADR 0043 §5 priced the
alternative: a required "what I read / what I write" field on every
`StaticAbility` — ~1700 catalog entries — whose wrong answers are
invisible until two specific cards meet. A trial application asks the
rules' own question of the effects themselves and needs nothing from a
card file. The one thing a card file still declares is Decision 2's
`ContinuesAfterRemoval`, which is a different question.

**Scope: layer 4, and the reason is measured.** The machinery is
layer-agnostic; `dependencyOrderedLayers` is the one-line list of
buckets that pay for it. Layer 4 is in it because it is the only
bucket the catalog can build a dependency in — no catalogued layer-6
effect reads a keyword, and no 7c effect reads effective P/T.

On the committed 80-permanent benchmark board
(`cards/effects/bench_test.go`), turning it on for every layer costs
**more than ten times**: widening `dependencyOrderedLayers` to every
bucket takes `BenchmarkLayerRecompute_80perms` to ~570-745 µs, and
from 334 to 6794 allocations, because that board
carries ten Glorious Anthems in layer 7c and every one of the 90
ordered pairs is trial-applied on every recompute to discover what
the rules already guarantee — an anthem cannot change what another
anthem applies to. Adding a bucket to the list needs a catalogued pair
that wants it and a benchmark that says what it costs.

**Cost, layer 4.** O(N²) trials for a bucket of N effects, and each
trial is bounded by the number of objects the candidate DEPENDENCY
applies to — not by the size of the battlefield — because an effect
can only change what another sees on the objects it applies to. The
relation is computed once per bucket; if it is empty, the bucket falls
straight through to timestamp order and no further trials run.

Measured on the same 80-permanent board plus Urborg, Maskwood Nexus
and two attached Songs of the Dryads
(`BenchmarkLayerRecomputeLayer4Dependencies_80perms`, `-count=8
-benchtime=3000x`):

| board | before (develop `046ee617`) | after |
|---|---|---|
| `LayerRecomputeLayer4Dependencies_80perms` | 120.7-124.4 µs, 54 656 B, 758 allocs | 106.5-112.9 µs, 89 464 B, 666 allocs |
| `LayerRecompute_80perms` (no layer-4 effect) | 42.5-45.9 µs, 23 904 B, 333 allocs | 48.3-50.3 µs, 23 928 B, 334 allocs |

Two numbers, in opposite directions, and both are the decision.

On a board that HAS a layer-4 dependency the recompute got about **12%
faster**, because Decision 2 deleted the ability-removal fixed point
and the second full pass it used to run costs more than the trials do.
Bytes per op go up: a trial clones a characteristic per probe.

On a board that has none it got about **13% slower** — roughly 6 µs on
eighty permanents, which is the deliberately saturated case and about
ten times a normal Commander board. `applyOneEffectLocked` takes a
fast path when no effect in play can remove an ability (`st.track`
false), so the inner loop over (effect, object) is exactly the S16
one; the residue is the restructured pass itself and the one extra
allocation for `layerPassState`. Profiling attributes it inside
`printedCharacteristic` / `ParseTypeLine`, which this change does not
touch, so some of it is code layout rather than work. It was measured
against unchanged control benchmarks in the same binaries
(`HarvestTapEvent`, `HarvestETBEvent`, `CounterWithReplacementGather`,
all flat), so it is not global noise either.

That trade is accepted rather than hidden: the recompute runs on a
`layerVersion` bump, 50 µs on the worst board the repo benchmarks is
not a budget anybody is spending, and the alternative is four rules
bugs. If it ever matters, the residue is in the pass structure and not
in CR 613.8 — the dependency machinery does not run at all on that
board.

**Declared narrowings**, both stated rather than hidden:

- The probe objects are the ones the candidate dependency applies to.
  An effect whose `Apply` reads a board-wide count rather than its
  target ("+1/+1 for each artifact you control") can be changed by an
  effect that never touches the same object, and would be seen only if
  the two also share an object. No catalogued layer-4 effect is in
  that class, and the ones that count a board are in layer 7c, which
  is not dependency-ordered.
- The relation is recomputed after each application only while the
  bucket still has an edge. A bucket whose relation starts empty is
  applied in timestamp order without re-asking. If X does not change
  what Y does, applying X leaves Y's behaviour alone by definition, so
  an empty relation stays empty; the narrowing is that this is assumed
  rather than re-verified.

## Decision 2 — CR 613.6: a removal applies in its layer and reaches forwards only

The old rule was "a silenced permanent contributes nothing to any
layer". The new rule, in one sentence: **an ability removal is applied
in the layer it happens in, it cannot reach back into the layers that
have already run, and an effect that has already started applying
carries on into the later ones.**

Concretely, for an effect from a source the pass has silenced:

| bucket | what happens | why |
|---|---|---|
| before the removal's bucket | applies normally | the removal has not happened yet |
| the removal's own bucket | timestamp order decides, via `Card.HasLostAllAbilities` | two Song of the Dryads enchanting each other settle on "the earlier one wins" |
| after it | stops, **unless** the effect declares `ContinuesAfterRemoval` and the same source had already applied to that same object | CR 613.6 |

`ContinuesAfterRemoval` is an opt-in `bool` on `StaticAbility`. The
engine models one printed sentence as one static per layer, so it
cannot see on its own that "is a 0/1 Insect with indestructible and
loses all other abilities" is one continuous effect spanning layers 4,
6 and 7b. The flag says so, and it is set by the four attachment
helpers that build those sentences (`SetAttachedColors`,
`LoseAllAbilities`, `SetAttachedBasePT`). The default — `false` — is
right for a standalone ability, which simply stops existing: an
animated Glorious Anthem under a Humility gives nothing, because its
only applicable layer is after the removal.

The pairing with "the same source had already applied to that same
object" is CR 613.6's "the same set of objects". It is tracked in
`layerPassState`, whose maps are nil on every board with no
ability-removing effect on it.

**The fixed point is gone.** ADR 0046 §4's two-to-four passes existed
only so the gather could learn what to exclude; nothing excludes
anything at gather time any more, and a removal that reaches forwards
only is fully resolved by one walk of `layerOrder`. `maxLayerPasses`,
`sameCardSet` and `silencedSetLocked` are deleted.

**The enabling rule, CR 704.5p.** ADR 0046 §4 justified the old
gather-time silence with "a Song of the Dryads on a Mind Control gives
the creature back: layer 2 runs long before layer 6 could have told it
to". The outcome is right and the reason was not. The Song'd Control
Magic stops being an Aura, so it becomes **unattached** as a
state-based action (CR 704.5p; Song's 2014-11-07 ruling), and
"enchanted creature" then names nothing. That rule was not implemented
([#675](https://github.com/krakenhavoc/cmd_and_ctrl/issues/675)), and
without it a CR 613.6-correct layer 2 would have kept the creature
stolen — so `attachmentLegalLocked` learns it here: an attached
nonbattle, noncreature permanent that is neither an Aura, an Equipment
nor a Fortification is illegally attached, and `attachmentSBALocked`'s
existing non-Aura branch unattaches it and leaves it on the
battlefield. The two repros #675 names are pinned.

## Decision 3 — CR 305.7 is a layer-4 effect, and it removes the land's own text

"An effect that sets a land's subtype to one or more of the basic land
types" is one continuous effect with three clauses, all in layer 4:
the land's old land types go and the named ones replace them; it loses
the abilities generated from its rules text; and it gains the basic
land type's intrinsic mana ability.

`effects.SetsBasicLandType(applies, types, subtypes)` is that effect,
declaring `Layer4Type` **and** `RemovesAbilities`. Layer 4 is the
whole decision: a removal in layer 4 cannot touch a grant that lands
in layer 6, so an ability another effect gave the permanent survives
whenever it was given — which is CR 305.7's last sentence and Song of
the Dryads' 2014-11-07 ruling word for word. The intrinsic mana
ability is not declared at all: `ManaAbilitiesForCard` derives it from
the EFFECTIVE subtypes (CR 305.6), so it arrives with the type and
survives the same effect's removal, which is what the rule says.

`RemovesAbilities` is therefore no longer a layer-6-only declaration.
`LoseAllAbilities` stays exactly what it was — CR 613.1f's "loses all
abilities", in layer 6, for Darksteel Mutation and Kenrith's
Transformation, which print those words.

Song of the Dryads and Magus of the Moon are both
`SetsBasicLandType`; so is any future Blood Moon, Yavimaya or
Prismatic Omen.

## Decision 4 — "Every creature type" is a layer-4 fact, not a keyword

`Characteristic.AllCreatureTypes bool`, next to `Subtypes`.

The keyword was only ever storage — ~345 entries in `Subtypes` would
make the wire type line unreadable and every subtype loop quadratic —
but storing a layer-4 type in the layer-6 ability list leaked into the
rules answer in both directions. A removal emptied `Abilities` and
took the types with it; a later type SET left the keyword behind and
the Elk stayed a Goblin.

One storage, and both rulings fall out of where it lives:

- **Printed changeling** seeds the flag in `printedCharacteristic`,
  which is the layer-0 baseline the pass starts from. CR 702.73a is a
  characteristic-defining ability and CR 613.2 applies CDAs before
  every other effect in their layer, so the baseline IS "first in
  layer 4". One projection, one place, and the keyword stays in
  `Abilities` as the printed source of the flag.
- **A grant** (`AllCreatureTypesGrant`, `GrantAllCreatureTypesUntilEOT`)
  sets the same flag in layer 4.
- **A layer-4 subtype SET clears it**, through
  `Characteristic.SetSubtypes`, which is now the only way an effect
  replaces the subtype list. An ADD (`append`) deliberately does not
  go through it, because adding a type takes nothing away.
- `HasAllCreatureTypes` reads the flag on the battlefield and falls
  back to the printed keyword elsewhere, which is CR 113.6 for the
  grant and CR 702.73a's every-zone reach for the printed one.

**The wire.** The client's `changeling` badge is now a PROJECTION,
not storage: `viewOfAbilityBadges` appends the token when the
permanent is every creature type and does not already carry it, so a
Maskwood-granted creature still shows the badge and the engine still
stores the property once.

## Decision 5 — Nothing new is snapshotted

`Characteristic` is the layer engine's per-pass cache.
`snapshot_drift_test.go` classifies `Card.effective` as `rebuilt`
("restore forces a recompute"), and every field this ADR adds lives
there: `AllCreatureTypes` on the characteristic, and the removal
stamps and CR 613.6 bookkeeping in `layerPassState`, which does not
outlive a single pass. Undo, clone, snapshot and restore all get the
right answer by recomputing, and no schema version moves.
`TestCloneRecomputesTheSameLayerAnswer` says so.

## Consequences

- **Cards released from the AGENTS.md §7 hold list.** Magus of the
  Moon ships in this change and is the CR 613.6 and CR 305.7 proof.
  Arcane Adaptation, Leyline of Transformation, Encroaching Mycosynth,
  Yavimaya, Cradle of Growth and Prismatic Omen are unblocked: each is
  a layer-4 type-add that reads a type other layer-4 effects write,
  which is now the resolved case rather than the broken one.
- **Caveats cleared** on Urborg, Tomb of Yawgmoth, Song of the Dryads,
  Arixmethes, Slumbering Isle, Maskwood Nexus and The Warring Triad.
  Urborg, Song of the Dryads and Arixmethes are `CompletenessFull`.
- The recompute is one pass again, which it has not been since S24.
- A new card whose static spans layers has one thing to get right that
  it did not before: `ContinuesAfterRemoval` on the later-layer
  halves. It only matters when the source is silenced, and the four
  attachment helpers already set it.

## Alternatives considered

- **A declared read/write set on every `StaticAbility`** — ADR 0043
  §5's design. Rejected on the cost it named: a required field on
  ~1700 entries, wrong invisibly.
- **A topological sort computed once per bucket.** Cheaper, and not
  what CR 613.8c says: the rule re-examines after each application,
  and the loop escape is defined in terms of that iteration.
- **Diffing the effect's output.** Simpler to write and wrong on the
  first negative case in the catalog (Urborg + Song on a Command
  Tower), which is why the trial holds the base fixed.
- **Keeping the ability-removal fixed point and limiting the gather
  exclusion to layers 6 and 7.** Two rules for one question, and it
  would still have broken CR 613.6's continuation. One rule, applied
  per layer, is the decision.
- **A sentinel subtype for "every creature type"** instead of a
  `bool`. It would make "a set clears it" free, and it would put a
  string nothing prints into the wire type line and into every
  subtype loop in the engine. The flag plus one funnel
  (`SetSubtypes`) buys the same property with no leak.
