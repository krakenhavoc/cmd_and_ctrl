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

Six write entry points read effective types *before* any SBA pass
would have refreshed them, so they now refresh at their head:
`CastSpell` (CR 601.2c announce-time target legality),
`ActivateAbility`, `ActivateLoyalty`, `TapCard` (the CR 302.6 gate
only binds creatures), `ActivateManaAbility` (the ability list is
partly type-derived — see decision 4) and `PlayCard` (the card
played is a hand card, but its enters-tapped replacement asks about
the *board*: "unless you control a Swamp").

`AnnounceTrigger` was audited and deliberately left alone: it reads
no types.

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

### 5. `Card.HasSubtype` is the one subtype question

Card types were not the only thing being asked of the printed line.
Five card files had grown their own subtype matchers — two exact-match
loops over `Effective().Subtypes`, one `EqualFold` loop, and two
substring scans of `Card.TypeLine` — and the substring ones were
wrong in the same way `IsCreature` was.

All five now go through `game.Card.HasSubtype`, which reads effective
subtypes with the same nil-cache fallback as `HasCardType`. That
makes three real fixes, and each one had a comment explaining why it
couldn't be done:

- **`youControlLandTyped`** (via `IsLandWithSubtype`) — a checkland's
  "unless you control an Island or a Swamp". Under Urborg every land
  is a Swamp, so Drowned Catacomb enters untapped off a Forest. Its
  library-search callers (Nature's Lore, the fetchlands) are
  unaffected: a card in a library has no layer cache, so `HasSubtype`
  hands them the printed line verbatim.
- **`controlsLandSubtype`** — the same condition for the other
  cycle. Its note said `Effective()` was unreachable "because the
  checkland is the entering card and the layer cache is maintained
  only for battlefield permanents." The entering card is `src`,
  which the loop skips; every card it actually asks about is a
  battlefield permanent with a live cache. The note was reasoning
  about the wrong object.
- **`isPirate`** — its note said reading `Effective()` from inside a
  Layer 6/7 `AppliesTo` "recurses into the computation being
  performed." It does not; see decision 2. What it sees is a
  *partially applied* view, which is CR 613 working as intended —
  and `lordOfAtlantisOtherMerfolk` had been reading
  `Effective().Subtypes` from exactly that position since S16.

Two comments asserting an impossibility that the file next door was
already doing is what a shared helper prevents.

### 6. Hand-built `Card.effective` fixtures must start from printed

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
  lands, revoked when Urborg leaves, visible on the wire so the
  client can render the click, and — the part nobody wrote for
  Urborg — turning a checkland on.
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

## Amendment (2026-10-15, #2041): losing a card type loses its subtypes (CR 205.1a)

CR 205.1a: "If an object's card type is removed, the subtypes correlated
with that card type will remain if they are also the subtypes of a card
type the object currently has; otherwise, they are also removed for the
entire time the object's card type is removed." Until now each effect
that removed a type had to remember the second half, and
`RemoveTypesMod("Creature")` did not: an "isn't a creature" Weeping Angel
was still an Alien Angel to every Angel count and tribal lord.

**The pass applies it, not the effects.** Every layer-4 application
(`applyOneEffectLocked`, and the CR 613.8 probe's `applyRaw`) runs through
`applyTypeChange` (`game/card_type_loss.go`): it notes which subtype lists
the object's card types allow before the effect and after it, and drops
the subtypes of every list the effect closed. Creature types are allowed
by Creature or Kindred (Tribal), land types by Land, artifact types by
Artifact, enchantment types by Enchantment, planeswalker types by
Planeswalker (CR 205.3j's list, newly carried), spell types by Instant or
Sorcery, battle types by Battle. "Is every creature type" goes with the
creature list. A subtype on none of the lists is kept.

**In the removing effect's slot.** The drop is part of the effect that
removed the type, so CR 613.7 order decides: a later "becomes an artifact
creature" makes the object a creature again without the creature types the
earlier removal took (it is still removing the type, and CR 205.1b's
"retains its prior subtypes" retains none), and a removal after an
animation takes the animation's creature type too. When the removal ends,
the next pass rebuilds from printed values and the subtypes are back. A
filter at the end of layer 4 would have given the first case's subtypes
back, which is why it is per application.

**Not here.** The gaining half, CR 205.3d's "an object can't gain a
subtype that doesn't correspond to one of that object's types", is not
enforced by the pass: a scoped record is one continuous effect per mod,
and a record listing its subtype before its card type would lose its own
subtype. `SetTypesMod`'s registration check refuses the mismatch it can
see.

## Amendment (2026-10-15, #3862): layers 4 and 5 for objects off the battlefield

Decision 1 of this ADR left an object off the battlefield at its printed values. Its reason was that
"a static ability only does anything while its source is on the battlefield, so there is nothing
for the effective view to say there". That reason confuses where the SOURCE is with where the
OBJECT is. A battlefield static can be about objects anywhere (CR 611.3a), and three cards in the
catalog say so:

- Mycosynth Lattice: "All cards that aren't on the battlefield, spells, and permanents are
  colorless." Its ruling: it "makes everything, in every zone of the game, colorless."
- Maskwood Nexus: "The same is true for creature spells you control and creature cards you own that
  aren't on the battlefield."
- Secret Arcade: "Nonland permanents you control and permanent spells you control are enchantments
  in addition to their other types."

Each shipped its permanent half with a caveat. The visible failures were a Cavern of Souls named for
Elf refusing to pay for a Bear in hand under a Nexus, and Dusty Parlor not seeing the creature spell
Secret Arcade makes an enchantment.

**1. The declaration.** `StaticAbility.AffectsZones` lists the zones whose objects a layer-4 or
layer-5 static's `AppliesTo` is asked of. `ZoneStack` means spells: abilities on the stack are not
spells, and a spell's `Controller` is settled by layer 2 first. `ZoneHand`, `ZoneLibrary`,
`ZoneGraveyard`, `ZoneExile` and `ZoneCommand` mean cards, whose "you" is their owner (CR 108.4).
A card whose sentence names both kinds declares two statics, so each predicate reads the right
"you". `effects.OffTheBattlefield`, `CardsOffTheBattlefield` and `SpellsOnTheStack` are the lists.

`effects.Register` refuses any of these:

- a layer other than 4 or 5;
- a missing `AppliesTo` or `Apply`;
- a removal, grant, borrow or `ContinuesAfterRemoval`;
- a source zone, a spell clause or a zone grant on the same static;
- the battlefield.

`GrantedStaticProblem` and the emblem check refuse it too. The battlefield gather skips such a
static, so a predicate written for a hand card cannot reach a permanent.

**2. The step.** `zoneLayerPassLocked` (`game/zone_layers.go`) runs in `recomputeLayersLocked` after
layer 2 and after the zone grant step. It gathers every such static whose source still has its
abilities after the battlefield pass, in layer order and then timestamp order (CR 613.7). It then
walks every spell on the stack and every card in each seat's hand, library, graveyard and command
zone and in exile. For each object, it applies the statics that accept it to a copy of the printed
baseline, through `applyTypeChange` for layer 4 (CR 205.1a's subtype loss), and stamps the result on
`Card.zoneLayered`. While a static is applied, the object's stamp is the characteristic in progress,
so a later predicate reads what the earlier effects made it, as on the battlefield. With no such
static on the battlefield, the step does nothing. `Game.zoneLayersStamped` makes the first such pass
clear the old stamps once. A catalog index (`IndexZoneLayers`) keeps the battlefield walk to one map
read per permanent.

**3. The readers.** There are six:

- `hasCardType`
- `hasSubtype`
- `HasSupertype`
- `EffectiveColors`
- `HasAllCreatureTypes`
- `effectiveOf` (and so `Effective()`)

Off the battlefield, each one now reads `zoneLayered` before it falls through to the printed card.
Every other reader goes through these six: the cast's spend context, a tutor's filter, a "colorless
spell" check, protection, the wire's type line, and `cdaAnywhereLocked`. A commander's colour
identity is not its colour (CR 903.4), so `printedIdentityOf` now reads the printed colours instead
of `Effective()`. Under a Lattice, a commander in the command zone is colourless, and its identity
is unchanged.

**4. Freshness.** The stamp is a derived cache like `effective`, with the same rules:

- `MoveCard` clears it, because the card in its new zone is a new object (CR 400.7).
- The battlefield pass clears it on a permanent.
- It is never snapshotted. The drift test classifies it `rebuilt`, a clone shares the pointer, and a
  restore in place forces a recompute that re-derives it.

While such a static is live, `layerVersionBump` bumps the layer version for a cast and for a card
moving between two zones off the battlefield. The trigger harvester recomputes before it reads
`EventCast` or such a move, as it already did for `EventETB`. So "whenever you cast an enchantment
spell" sees the creature spell Secret Arcade made an enchantment, and a card drawn under a Nexus is
every creature type from the moment it is in hand.

**5. Not covered.** Dependency ordering (CR 613.8) among these effects: off the battlefield no
catalogued effect makes a card a creature, so timestamp order is the whole answer. Cards outside the
game (a sideboard) are in no zone of the game.

**Cards.** Mycosynth Lattice, Maskwood Nexus and Secret Arcade // Dusty Parlor are now Full. Pinned by
`TestMycosynthLatticeMakesCardsAndSpellsInEveryZoneColorless`,
`TestMaskwoodNexusMakesYourCreatureCardsAndSpellsEveryType`,
`TestMaskwoodNexusReachesACardDrawnWhileItIsOut` and
`TestSecretArcadeMakesYourPermanentSpellsEnchantments` (`internal/cards/effects`).
