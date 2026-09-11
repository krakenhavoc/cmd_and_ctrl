# ADR 0039 — Layer 4 is authoritative: the type predicates read `Effective()`

**Status:** Accepted · 2026-09-11 · follows ADR 0012

## Context

ADR 0012 shipped the CR 613 layer engine: `Game.recomputeLayersLocked`
rebuilds every battlefield card's `Characteristic` from printed values
and applies the active static abilities in layer order, and
`Card.Effective()` returns the result. Decision 8 of that ADR routed
the wire through it — `CardView.power`, `toughness`, `type_line` are
post-layer.

Nothing else was. `Card.IsCreature()`, `IsLand()`, `IsArtifact()` and
the rest read `Card.TypeLine`, the **printed** line. So a Layer-4
type-changing static was computed, projected onto the wire, and then
invisible to combat, state-based actions, targeting, the catalog's own
`Creature()` predicate, and the synthetic basic-land mana ability.

Four independent reports landed on the same wall in one week:

- **#255** (flicker) — "Layer 4 type changes are inert… this is what
  makes devotion-gated 'isn't a creature' (Thassa, all the Theros
  gods) unimplementable."
- **#258** (Hashaton) — declined to catalogue **Urborg, Tomb of
  Yawgmoth**, because its Layer-4 static "would apply cleanly and do
  *nothing*."
- **#344** — "layer 4 is wire-only in this engine", blocking Enduring
  Curiosity (#321): the returned permanent "would still attack, block,
  be targeted as a creature, and feed its own trigger."
- **#348** — the same wall on The Seriema's "it's an artifact creature
  at 7+".

`effects/targets.go` carried the comment `--- type predicates
(post-layer, so type-changers compose) ---` above a block of
predicates that were not post-layer. The intent had been written down
years before the wiring existed.

## Decisions

### 1. The `Is*` predicates read `Effective().Types`; `PrintedIs*` is the other half

`Card.IsCreature` / `IsLand` / `IsInstant` / `IsSorcery` / `IsArtifact`
/ `IsEnchantment` / `IsPlaneswalker` / `IsBattle` / `IsPermanent` all
delegate to `Card.HasCardType(lowerType)`, which reads the post-layer
type set. `Card.HasSubtype` is the matching accessor for subtypes.

`Card.PrintedIsCreature` and `Card.PrintedIsLand` are the deliberate
printed-value surface for callers that must not move when a Blood Moon
lands — CR 707.2 copiable values, deck-construction checks, anything
asking "what does this card actually say."

**Why an explicitly named second surface rather than a bool
parameter:** a call site's choice between printed and effective is a
rules decision. Naming it means the choice shows up in the diff that
makes it, and a reviewer can see which one a new call site picked
without reading the argument list.

**Call-site assignment.** Every existing `Is*` call site was left on
the effective accessor, and that is not laziness — it is the audit
result. The ~90 non-test call sites split three ways:

- **On-battlefield reads** (combat declarations, the SBA loop,
  activated-ability legality, the catalog's `AppliesTo` and target
  predicates, the mana-ability derivation): effective is the answer
  they always wanted. This is the whole point of the change.
- **Off-battlefield reads** (cast legality from hand, land-drop
  checks, stack-resolution routing, graveyard and exile predicates):
  `Card.effective` is nil there, so `HasCardType` falls through to the
  printed type line and the behaviour is *byte-identical* to before.
  CR 113.6 says a static ability only does anything while its source
  is on the battlefield; there is nothing for the effective view to
  say about a card in a hand.
- **Genuinely printed reads**: there were none. The one place that
  reads a printed characteristic on purpose —
  `DistinctCardTypesInAllGraveyards` — already called
  `printedCharacteristic()` explicitly, and the 704.5f placeholder
  exemption already reads the printed `Card.Toughness` field directly
  rather than through any predicate.

So `PrintedIs*` ships with test coverage and no production caller
today. It exists because the next two changes need it: multi-face
cards (ADR 0034) and Layer 1 copy (#335) both have to distinguish
"printed on this face" from "what the engine currently says."

### 2. `Effective()` stays passive; that is what forbids the recursion

The layer pass calls `AppliesTo` per candidate per effect, and
`AppliesTo` predicates call `IsLand()`. If a type predicate triggered
a recompute, the pass would re-enter itself through every predicate,
forever.

It cannot, because `Card.Effective()` is a pure read of the cached
`Card.effective` pointer and always has been. Nothing in this change
adds a recompute call below the predicate layer. The invariant is
written on `HasCardType`: **a type predicate never starts a layer
pass.** Freshness is a caller obligation, exactly as ADR 0012 decision
3 already made it for `CurrentPower` / `CurrentToughness`.

A predicate reading a *partially applied* resolution mid-pass is not a
bug: `applyLayerLocked` walks buckets in CR 613 order, so a Layer-6
predicate sees post-Layer-4 types, which is what CR 613 asks for. Two
Layer-4 effects see each other in `EnteredBattlefieldAt` order, which
is CR 613.7.

### 3. Caching: reuse `layerVersion`, add five recompute calls, not forty

ADR 0012's `layerVersion` / `lastResolvedVersion` pair already
fast-paths a clean recompute to two atomic loads. Every read path
already refreshes: `ReadSnapshot` calls
`RecomputeLayersIfStaleLocked`, and `legal.EnumerateFor` goes through
`ReadSnapshot`. The SBA loop refreshes at the head of
`stateBasedActionsLocked`, which runs after 28 mutation sites.

Five write entry points read effective types *before* any SBA pass
would have refreshed them, so they now refresh at their head:
`CastSpell` (CR 601.2c announce-time target legality), `ActivateAbility`,
`ActivateLoyalty`, `TapCard` (the CR 302.1 gate only binds creatures)
and `ActivateManaAbility` (the ability list is partly type-derived —
see decision 4).

`PlayCard` and `AnnounceTrigger` were audited and deliberately left
alone: the first reads a hand card's type, the second reads no types.

**Why not a blanket refresh on every exported mutation:** forty
identical lines would be forty places for the next person to wonder
whether the call is load-bearing. Five annotated calls each say what
they are protecting.

### 4. Intrinsic land mana comes from effective subtypes, per CR 305.6

The S15 synthetic ability (`basicLandColor`) required the **Basic
supertype** and read the **printed** type line. Both were wrong, and
the same wrongness twice:

- CR 305.6 has never mentioned the supertype. The intrinsic
  `"{T}: Add {B}"` comes from the **Swamp land type**. Requiring the
  supertype left every printed dual — Bayou, Overgrown Tomb, a Triome
  — producing no mana at all unless someone hand-wrote a catalog
  entry. `battle_lands.go` documented that shortcut in a comment as
  the reason its cycle declares a pipe ability the printed card puts
  in reminder text.
- Urborg grants the Swamp *type*, never the supertype. The old rule
  could not have worked for it even in principle.

`intrinsicLandManaAbilities` now walks the card's **effective**
subtypes and emits one ability per distinct basic land type, **in
subtype order**. Subtype order matters: a Layer-4 grant appends, so a
land with no static on it keeps the exact ability list, the exact
ability indices, and therefore the exact auto-tapper first-ability
choice it had before.

`ManaAbilitiesForCard` merges rather than choosing: instance
abilities, else catalog abilities, **plus** the intrinsic abilities
for colours the declaration cannot already make. Sunken Hollow's
declared "Add {U} or {B}" pipe covers both its land types, so it gains
nothing and its client row stays one click; a checkland under Urborg
gains the {B} its pipe cannot make.

**Known narrowing:** a Spec that deliberately declared a land's
ability as something *other* than its land types would shadow the type
half. No catalog card is in that position today.

### 5. Hand-built `Card.effective` fixtures must start from printed

Four test fixtures stamped `&Characteristic{Power, Toughness,
Abilities}` directly, leaving `Types` empty. That was harmless while
nothing read `Types`; with the predicates rerouted it declares a
permanent with no card types at all — which is precisely what the
engine would produce for a creature that had lost every type, so the
fixtures' "creatures" correctly stopped being creatures and 38 combat
tests failed.

Fixtures now go through `printedEffectiveWith`, which seeds from
`printedCharacteristic()` and layers abilities on top. The failure was
the change working, not the change breaking: a fixture that lies about
its types is a fixture that will pass while the card is dead.

## Consequences

- **Urborg, Tomb of Yawgmoth ships** — its entire printed text is one
  Layer-4 static. `urborg_test.go` proves it end-to-end: real `{B}` in
  a real mana pool through `ActivateManaAbility`, on every player's
  lands, revoked when Urborg leaves, and visible on the wire so the
  client can render the click.
- **Every printed dual land produces mana without a Spec.** Bayou,
  the shocklands, the Triomes: CR 305.6 now applies to them.
- The Theros gods' "isn't a creature", The Seriema (#348), Enduring
  Curiosity (#321) and Kormus Bell-style animations become ordinary
  catalog work. The engine tests in `layer4_authoritative_test.go` pin
  both directions — a type-add makes a land attack-legal, a type
  removal makes `DeclareAttacker` refuse.
- `effects/targets.go`'s "post-layer, so type-changers compose"
  comment is now true.
- **Layer 1 (copy) remains unshipped** and is not blocked by anything
  here. `Layer1Copy` still has no producers. Clone's copy choice is an
  as-enters replacement (CR 614.1c), not targeting, so no targeting
  work produces the prompt: the entry-replacement pipeline exposes
  `EntersTapped`, `EntersWithCounters`, a zone redirect and `Cancel`,
  and nothing that rewrites the entering card's identity. That is a
  new replacement capability plus an as-enters choice prompt plus
  client UI — a feature, not a follow-on, and deliberately left to
  #335.
