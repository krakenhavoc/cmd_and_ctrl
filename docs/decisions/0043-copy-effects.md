# ADR 0043 — Copy effects: layer 1 is the copiable-value baseline

**Status:** Accepted · 2026-09-11 · S16.5
**Supersedes:** the Layer 1 deferral in [ADR 0012](0012-layer-system.md) §5.

## Context

ADR 0012 shipped the CR 613 layer engine with layers 1, 3 and 5 as
empty buckets in `layerOrder` and no card exercising them. Layer 1 —
copy effects — was deferred to S16.5 with a named trigger: *land it
when a Clone-class card enters the catalog*.

That trigger fired twice. Clone was added to a playtest deck and
[#335](../../issues/335) reported it: "does not allow the selection of
any usable target and enters the battlefield as a 0/0." Both halves of
that sentence are one missing feature. The CR 614 replacement pipeline
had no way to express "as this enters, choose something", and the
layer engine had no way to express "this permanent's printed values
are somebody else's".

Two things about the engine changed in the meantime and both bear on
the design:

- **[ADR 0039](0039-layer-4-authoritative.md)** made the layer engine
  authoritative: `Card.IsCreature()` and friends read `Effective()`,
  so anything that rewrites a permanent's types is now visible to
  combat, SBAs and targeting rather than only to the wire.
- **[ADR 0034](0034-multi-face-cards.md)** introduced `SetFace`, which
  materialises one printed face's characteristics onto the flat
  printed fields of a `Card`. That is structurally the same operation
  a copy effect needs, with a different source.

## Decisions

### 1. A copy effect rewrites the copiable-value BASELINE, not a `Characteristic`

The obvious reading of "layer 1" is a `ContinuousEffect` at
`Layer1Copy` whose `Apply` overwrites the target's `Characteristic`.
It was rejected, because `Characteristic` carries Name, types,
colours, P/T and abilities — and a copy has to bring more than that:

- **the oracle ID.** Every catalog hook (ETB triggers, static
  abilities, replacements, activated and mana abilities, printed
  keywords) resolves through `CatalogKey`, which reads
  `Card.OracleID`. A copy that does not carry it is a Clone of Lord of
  Atlantis that grants nothing and a Clone of Blood Artist with no
  trigger — indistinguishable on the wire from a working one, which is
  the worst possible failure mode.
- **the printed P/T as PRINTED data.** `recomputeLayersLocked`
  deliberately does not run inside the SBA loop (ADR 0012 §1), so
  between a Clone entering and the next recompute the CR 704.5f
  state-based action would read the printed 0/0 and destroy it.
- **mana cost, colour identity, keywords, starting loyalty, layout
  and faces**, none of which `Characteristic` has a field for.

So `Card.applyCopy` writes the copied values into the flat printed
fields, exactly as `SetFace` writes a face's — and everything
downstream (the `Is*()` predicates, `printedCharacteristic`,
`CatalogKey`, `ManaAbilitiesForCard`, the view projection) is correct
without knowing a copy happened. CR 613.1a is satisfied by
construction: layer 1 is applied, and layers 2–7 then run on its
result. CR 707.2's "copiable values are the printed values as modified
by other copy effects" falls out too — `CopiableValuesOf` reads the
current printed fields, so a Clone of a Clone copies what that Clone
became.

`Layer1Copy` stays in `layerOrder` for the class this does not cover:
a copy effect with a DURATION and a timestamp of its own (Mirage
Mirror, Cytoshape). Those must be re-applied on every recompute and
ordered against other layer-1 effects. An entry copy never does — it
is settled once, as the permanent enters, and never changes again
while the permanent is on the battlefield.

### 2. `Card.PrintedSelf` holds the revert, cleared on battlefield leave

CR 400.7: a permanent that changes zones becomes a new object, and the
copy effect applied to the PERMANENT. A Clone that dies is a card
named Clone in its owner's graveyard.

`Card.PrintedSelf *PrintedValues` is the card's own printed values,
stashed on the first `applyCopy` and restored by the battlefield-leave
branch of the layer listener — the same place that already clears the
effective-characteristic cache, so the two halves of "this permanent
stopped existing" live together. A permanent that is already a copy
and becomes a copy of something else keeps its ORIGINAL `PrintedSelf`,
so it still reverts to the right card.

`PrintedValues` carries no closures, which is what lets it be
`carried` in the snapshot rather than `dropped`. The two
closure-bearing slices (`ManaAbilities`, `ActivatedAbilities`) are
copied onto the card separately — they exist only on tokens, and for
everything else they are re-derived from the copied oracle ID.

### 3. The copy choice is a replacement with a prompt inside it

`ReplacementEffect.CopySelector` is the third member of the family
`EntryLifeCost` started (the shockland's pay-2-life, ADR-less, in
`entry_choice.go`): the CR 614 apply-loop hits it, queues a
`PendingChoiceCopyTarget`, and bails with `errReplacementPending`. The
answer stamps `ev.EntersAsCopyOf` and the entry path materialises it
**before** `EventETB` — so the permanent never exists on the
battlefield as its own printed self, and every ETB trigger, its own
and every watcher's, sees the copy (CR 706.2).

The card's "except" clause is an edit to the values on their way in
(`Except`), not a knob on the engine: Sakashima's `SetName`,
Metamorph's `AddCardType("Artifact")`, Spark Double's
`RemoveSupertype("Legendary")`. All three are methods on
`PrintedValues`.

Declining is always legal. Every printed card in this class says "you
may", and an empty `card_ids` list is the decline on the wire.

### 4. Stack resolution became a resumable entry site

A shockland is *played*; a Clone is *cast*. The land-play branch was
the only `entryResumable` site, because the generic resume could not
reproduce two things stack resolution does: attaching a resolved Aura
to what it targeted (CR 303.4a) and queueing evoke's sacrifice trigger
(CR 702.74b). `ReplacementEvent.stackItem` carries the `StackItem`
across the pause and `executeEntryToBattlefieldLocked` does both.

That closed a latent bug older than this ADR and unrelated to copying:
**any** permanent spell whose entry drew two applicable replacements
would queue the CR 616 ordering prompt, bail with its `StackMeta`
entry already deleted, and never be pushed. The permanent was simply
lost. Nothing had hit it because no catalog pair had yet lined up on
one entry. `TestPausedPermanentEntryStillReachesTheBattlefield` pins
it.

### 5. Dependency detection (CR 613.8) stays deferred

The other half of [#159](../../issues/159) is **not** shipped, and
this is the decision rather than an omission.

CR 613.8's rule — apply A before B when A's existence changes what B
does — needs every static ability to declare its dependency inputs
(what it reads, what it writes) and a topological sort with cycle
detection inside every layer bucket, on the hottest path in the
engine. The declaration is the expensive part: it is a new required
field on `StaticAbility` that every existing card and every future one
has to fill in correctly, and a wrong answer is invisible until two
specific cards meet.

Nothing has asked for it. The trigger #159 named for this half — a
playtester hitting an Opalescence + Humility-class pathology — has not
fired, and the catalog contains no pair that can produce one:
dependency only becomes observable when one static changes whether
another static APPLIES, and every static in the catalog today is an
anthem, a keyword grant, a type-add or a CDA, all of which commute
under timestamp order.

Shipping a speculative topological sort would mean adding a required
declaration to every card in the catalog to fix a bug no game can
currently produce. Pure timestamp ordering stays; the trigger stays
armed.

**Amended 2026-09-16 (the [#159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/159)
closeout): the "no pair exists" premise is false, and has been since
the day this merged.** The paragraphs above are left as written, as
the record of why #414 did not build dependency ordering. Their
premise is not a reason to keep deferring it.

- **Type-adds do not commute with type-sets.** A type-add whose
  `AppliesTo` reads a type (Urborg, Tomb of Yawgmoth's "each land";
  Maskwood Nexus's "creatures you control") depends, under CR 613.8a,
  on any same-layer effect that writes that type. The catalog has
  type-*sets*: `SetAttachedTypes` (Song of the Dryads, Kenrith's
  Transformation, Darksteel Mutation), Arixmethes, Slumbering Isle's
  slumber, The Warring Triad's self type-strip, and crew's "becomes an
  artifact creature". Maskwood Nexus (#404) and Arixmethes (#480)
  merged on 2026-09-13, a few hours after #414. Song of the Dryads
  (#563) merged on 2026-09-14.
- **Reproduced on develop `7c1ae9b`** with throwaway probe tests
  (re-run 2026-09-16, not committed), all in layer 4. The rules result
  is the same whichever card entered first; the engine's result isn't:

  | pair | engine, in the wrong entry order | rules |
  |---|---|---|
  | Urborg (older) + Song of the Dryads (newer) on a Sol Ring | `[Forest]`, taps for {G} only | a Forest Swamp |
  | Urborg (older) + Arixmethes, slumbering (newer) | no subtypes, {G}{U} only | a Swamp too |
  | Maskwood Nexus (older) + a Vehicle crewed later | not every creature type: no lord pump, no haste from Goblin Chieftain | every creature type |
  | Maskwood Nexus (older) + The Warring Triad under eight graveyard cards | not a creature, yet still every creature type | not every creature type |
  | Maskwood Nexus (older) + Arixmethes, slumbering | a land that is every creature type | not every creature type |

  Crew is stamped when it resolves, so the Maskwood + Vehicle order is
  the usual one, for all six catalog Vehicles.
- **The engine already resolves one CR 613.8 case**: ability removal,
  by iterating to a fixed point ([ADR 0046](0046-layer-6-authoritative.md) §4).
  Part of that is itself wrong (see the note on ADR 0046 §4).
- **The cost argument is not settled either.** A required read/write
  declaration on every `StaticAbility` is one design, not the only one.

Tracked in [#668](https://github.com/krakenhavoc/cmd_and_ctrl/issues/668)
(an ADR, then CR 613.8 dependency ordering in layer 4) and
[#669](https://github.com/krakenhavoc/cmd_and_ctrl/issues/669) (ability
removal across layers, and Song of the Dryads' CR 305.7 loss), which
comes first. Until they land, [#644](https://github.com/krakenhavoc/cmd_and_ctrl/pull/644)
declares these pairs as caveats on the cards and pins them as skipped
tests, and AGENTS.md §7 "When NOT to add a catalog entry" holds back
the cards that would add more pairs.

## Consequences

- Clone, Phyrexian Metamorph, Spark Double and Sakashima the Impostor
  work, and #335 closes.
- A copy inherits the copied card's catalog behaviour for free, which
  means every card added to the catalog from now on is also a card
  Clone can usefully copy.
- A permanent spell's entry can now pause for any prompt without
  losing the permanent.
- One new carried field on `Card` (`PrintedSelf`), one on
  `PendingChoice` (`CopyOptions`), one new `PendingChoiceKind`, one
  new `EventKind`.
- Not covered, declared: a copy effect with a duration (Mirage Mirror,
  Cytoshape); an "except" clause that GRANTS an ability rather than
  editing printed values (Sakashima's own return-to-hand ability); a
  declined Clone surviving as a 0/0, which is the engine's
  pre-existing printed-0-toughness SBA convention and not specific to
  copying.
