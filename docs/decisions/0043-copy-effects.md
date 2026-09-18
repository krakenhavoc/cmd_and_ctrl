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
and every watcher's, sees the copy (CR 707.2).

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
comes first.

**Superseded 2026-09-18 by [ADR 0067](0067-layer-dependency-ordering.md).**
This section's deferral is spent: CR 613.8 dependency ordering is
implemented for layer 4, detected by trial application rather than by
the declared read/write set this section priced and declined. Every
pair in the table above now comes out the same whichever card entered
first, the caveats are off Urborg, Song of the Dryads, Arixmethes,
Maskwood Nexus and The Warring Triad, and the AGENTS.md §7 hold list
is released. The paragraphs above are left as written, as the record
of why #414 did not build it.

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

## Amendment (2026-09-18, #920): a resolving spell keeps its stack metadata until resolution finishes

This ADR is about copying a PERMANENT as it enters; CR 707.10's copy
of a SPELL on the stack shipped separately in S30
(`server/internal/game/spell_copy.go`, #95). Both live here because
they share the copiable-value rule, and this amendment is about the
spell half.

**The bug.** `resolveTopOfStackLocked` deletes the resolving item's
`StackMeta` entry before it calls `OnResolve`, and
`resolveTopAbilityLocked` does the same. `CopySpellForEffect` finds
its source card and its announce-time choices through that map, so an
effect that says "that player may copy THIS spell" — CR 707.10, where
the copy is created by the resolving spell's own effect — found
nothing and did nothing. The whole Chain cycle prints that sentence;
Chain of Vapor is the card the #568 work stopped at.

**Decision: one slot on the game, `Game.resolving`**
(`server/internal/game/resolving_item.go`), holding the live
`*StackItem` and a VALUE copy of the card the stack held.

Two things are gone by the time the copy is actually made, and the
slot answers for both:

- the META, deleted above — the whole subject;
- the CARD, because the copy decision is a PROMPT. Chain of Vapor's
  "may sacrifice a land" pauses the resolution,
  `resolveTopOfStackLocked` carries on and routes the spell to its
  owner's graveyard, and the answer arrives afterwards. The copy is
  therefore built from last-known information. In the rules the spell
  is still on the stack at that moment — CR 608.2m puts it into the
  graveyard as the final step of its own resolution — and LKI is how
  that difference is spelled here.

**The lifetime is one event batch, and that is the terminal
outcome.** The slot is set when an item begins to resolve and cleared
by `beginEventBatchLocked`, which runs at exactly the two points where
play moves on (#829, CR 603.2c): the next resolution, and the cursor
entering a step. So it survives any number of paused continuations
belonging to the resolution that opened it — ADR 0013 §5's case — and
not one event past it. It is safe because a resolution-time prompt
BLOCKS the table (`choice_gate.go`): the cursor cannot walk past an
open copy decision, and nothing else can resolve under it.

A separate defer-clear at the end of `resolveTopOfStackLocked` was the
obvious alternative and is wrong for exactly this reason — the
function returns while the question is still open, so the clear would
land before the answer.

**It widens no other lookup.** `stackSpellLocked` falls through to the
slot only when the requested ID is the resolving item's own, so
Reverberate pointed at a spell that was countered in response still
gets `ErrCardNotFound`, which is the outcome its callers are written
for. The slot parks ABILITY items too, with no card, so that "the item
currently resolving" has one answer rather than a hole; nothing copies
an ability yet, and that remains its own seam.

**Snapshot classification.** `resolving` is `dropped` in
`snapshot_drift_test.go`, for the reason `enteringTokens` is: between
actions it is set only for a resolution paused on a prompt, and that
prompt's own resume frame is already counted in
`ContinuationCensus.ChoiceResumeFrames`. Its `*StackItem` carries an
`Effect` closure the snapshot could not represent anyway. `Clone`
SHARES the pointer rather than deep-copying it, because the paused
prompt's resume frame holds that same `*StackItem` (`may_choice.go`'s
contract) and a second copy would be an item the frame is not writing
through; the struct is replaced wholesale at each batch and never
mutated in place, so nothing can diverge.

**Cards.** Chain of Vapor (`CompletenessFull`) and Chain of Smog (one
caveat: the two cards discarded are random, which is `DiscardCards`
everywhere in the catalog and not this card's doing). Both give the
copy to the player who chose to make it, per CR 707.10b, which is what
lets the chain walk around the table.

## Amendment 2026-09-18 — an "except" clause may grant an ability and add a subtype (CR 707.9a, CR 707.9b) · Accepted · S45

The Consequences above declare "an 'except' clause that GRANTS an
ability rather than editing printed values" out of scope. That
deferral is spent. [#665](https://github.com/krakenhavoc/cmd_and_ctrl/issues/665)
lifts it; the paragraph is left as written, as the record of why the
original PR stopped where it did.

### What was actually missing

Not a grant mechanism — a place for the grant to LIVE. CR 707.9a's
second sentence is the hard half: an ability a copy effect grants is
part of the copiable values, so a later copy copies it. A closure
hanging off the replacement that made the copy cannot be copied again
and cannot be snapshotted. And every ability the engine reads is
found through `CatalogKey`, which after a copy is the COPIED card's
key, so a granted ability had nowhere to key on.

### Decision 6. The grant rides in `PrintedValues`, as a NAME

`PrintedValues.GrantedAbilities []string` holds catalog keys, not
closures, and `Card.GrantedAbilities` is the flat printed field it
materialises onto — the same pairing `Keywords` and `TypeLine`
already have. Everything else follows for free: `CopiableValuesOf`
copies it, so a Clone of a Phantasmal Image inherits the grant
(CR 707.9a); the snapshot carries it and `clone.go` deep-copies it,
because names serialise; `restorePrintedSelf` clears it on the way
out, because CR 400.7 makes the card in the graveyard the card again.

The abilities themselves are static catalog data, declared by the
card that grants them (`effects.Spec.Grants`, a list of
`AbilityGrant{Key, Triggered, Static, Activated}`) and filed in the
same `defs` map cards use, under `game.GrantKey(Key)`. That is the
shape [ADR 0064](0064-emblems.md) / #623 already uses for an emblem,
for the same reason: an object needs abilities the catalog can find,
and it is not a card. `Register` panics on an empty key, a duplicate
key, or a bundle with no abilities; the census still counts cards,
because a grant is not one.

### Decision 7. ONE lookup seam: the catalog KEY carries the grants

`CatalogKey(c)` is already the single place a `Card` becomes a
catalog answer — #940 put CR 708.2a's face-down silence there as an
empty return. A card carrying grants now returns a composite:

	no grants  →  "<oracle_id>"                       (unchanged)
	grants     →  "<oracle_id>|grant:<name>|grant:…"

and `catalogDef` answers a composite key by MERGING the card's
definition with each grant's (`game/copy_grants.go`). Nothing else in
the engine changed: the trigger harvester, the layer pass's static
gather, `ActivateCatalogAbility`, the replacement gather, the cost
modifiers and the view's auto bit all already went through those two
functions, so all of them see a granted ability without a reader of
their own. There is no second path to keep in sync, which is the
whole point of doing it at the key rather than beside it.

Three consequences worth stating:

- **Ability removal still wins.** `CatalogAbilityKey` returns `""`
  for a permanent under a CR 613.1f effect, before any of this is
  consulted, so "loses all abilities" takes the granted one too.
- **Face-down still wins.** `CatalogKey` returns `""` for a face-down
  permanent first (CR 708.2a): no text means no granted text.
- **A grant on a card with no catalog entry is the normal case.**
  Phantasmal Image copying an imported vanilla bear has exactly one
  ability, and it is the granted one, so `mergedCatalogDef` returns a
  definition even when the base lookup is nil.

Two sites read a catalog key as an IDENTITY rather than as a lookup
and now strip the suffix with `game.BaseCatalogKey`: `EmblemKey`
derivation (`game/emblem.go`) and the catalog's own `effects.Lookup`
of the Spec registry (`batch16_helpers.go`).

### Decision 8. `AddSubtype` / `RemoveSubtype` join the except-clause helpers

CR 707.9b's "it's an Illusion in addition to its other types" is an
ADD, not a SET: the copied Bear stays a Bear. The set form ("except
it's a 4/4 black Zombie") already lives in the token-copy template
(`retypedTypeLine`) and deliberately stays there.

### Cards

**Phantasmal Image** (new; caveat: the declined 0/0 survives, the
engine-wide printed-zero-toughness convention Clone also declares)
and **Sakashima the Impostor**, which drops its caveat and is now
Full.

### Still not covered

A copy effect with a DURATION (Mirage Mirror, Cytoshape) is
unchanged and still open — including the question this amendment
raises for it: when a timed copy effect ends, the grant ends with it,
because the grant is part of the copiable values the effect
installed. Granted MANA abilities have no slot (`AbilityGrant` names
three kinds), because no printed copy effect grants one and the
engine reads mana abilities off the card object as well as the
catalog, which would be a second seam. A granted ability's TEXT is
not projected onto the wire, so the client shows the copy's abilities
without it.
