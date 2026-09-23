# ADR 0069 — Face-down objects (CR 406.3a, CR 708)

**Status:** Accepted · 2026-09-18 · S43 — Hand special actions and face-down objects
**Issues:** [#656](https://github.com/krakenhavoc/cmd_and_ctrl/issues/656) (this ADR),
[#697](https://github.com/krakenhavoc/cmd_and_ctrl/issues/697) (the two lifecycle bugs it fixes)
**Numbering:** on 2026-09-18 every remote ref was listed with
`ls-tree -r --name-only <ref> docs/decisions/` per AGENTS.md §4 — 270 refs at
first writing, 275 at the re-check immediately before the push. The highest
file present on any branch was `0064-emblems.md` (the #927 emblems PR) and
then, at the re-check, `0065-modal-and-multi-target-clauses.md`. 0061–0068 are
being drafted concurrently by the S43 agents (0061 token creation and discard,
0062 abilities and special actions from the hand, 0063 durations and control,
0064 emblems, 0065 modal clauses, and three more not yet pushed), so this ADR
takes **0069**, which was free on both passes. 0005, 0024, 0029 and 0030 stay
permanently unused per AGENTS.md §4.
**Builds on:** [ADR 0034](0034-multi-face-cards.md) (`CatalogKey`, the face
materialisation `printedCharacteristic` reads),
[ADR 0012](0012-layer-system.md) (where layer 0 sits),
[ADR 0046](0046-layer-6-authoritative.md) (`CatalogAbilityKey`, the "what does
this permanent DO" accessor this ADR generalises),
[ADR 0028](0028-admin-context-menu.md) §7 (the sandbox `move_card` action that
leaked the flag), [#646](https://github.com/krakenhavoc/cmd_and_ctrl/issues/646)
(the non-knower redaction and the card back)
**Designs, does not implement:** morph, megamorph, disguise, cloak
([#95](https://github.com/krakenhavoc/cmd_and_ctrl/issues/95)),
foretell ([#658](https://github.com/krakenhavoc/cmd_and_ctrl/issues/658)),
suspend ([#659](https://github.com/krakenhavoc/cmd_and_ctrl/issues/659)),
the `turn_face_up` special action (CR 116.2g, the verb is
[ADR 0062](0062-abilities-and-special-actions-from-the-hand.md) §4)

## Context

Face-down is one bit on develop. `Card.FaceDown` is set in exactly one place
— the face-down exile branch of the shared exit primitive
(`game/zone_route.go`) — reached by exactly one card, Necropotence. Nothing
puts a permanent face down, nothing turns one face up, and the layer engine
has never heard of the flag.

Two families are waiting on it and they want different things:

| | foretell (#658) | morph / manifest / disguise / cloak (#95) |
|---|---|---|
| zone | exile | battlefield (and the stack, for a morph or disguise cast) |
| who may look | **the owner** (CR 702.143d, permitted by CR 406.3a) | **the controller** (CR 708.5) |
| characteristics | none that anything reads (CR 406.3a) | a 2/2 creature, no name, no text, no subtypes, no mana cost, colourless (CR 708.2) |
| abilities | the card's, for the cast out of exile | **none** (CR 708.2a); ward {2} for disguise and cloak |
| reveal | at game end and when the owner leaves (CR 702.143f) | when it leaves the battlefield (CR 708.9) |

and what develop actually does with the one bit it has:

1. **The owner cannot see their own face-down exile.** The face-down branch
   calls `ClearKnown()`, which is right for Necropotence (CR 406.3: no player
   may examine it) and wrong for every other face-down object in the game.
2. **A face-down permanent would run its real card.** Every catalog hook keys
   on `CatalogKey(card)` → oracle ID (`game/effect_hooks.go`, one `CardDef`
   lookup since #627). A face-down Ixidron'd Sheoldred would still drain.
3. **Nothing projects the 2/2.** `Effective()`, targeting, `IsCreature()` and
   the wire would all read the real card's printed power, toughness and type
   line off a permanent that is a 2/2 vanilla creature by rule.
4. **The flag survives a zone change on two paths** (#697): the sandbox
   `move_card` action (`mutations.go`, live today) and the cast-from-exile push
   (`mutations.go`, latent until foretell). `MoveCard` resets battlefield-only
   state, the exile grant, counters and the DFC face, but never `FaceDown`;
   the two callers that do reset it do so themselves, after the fact, which is
   a reset per caller rather than a rule.
5. **`StackOverlay.svelte`'s hover guard never fires** (#697): it tests
   `known_by_you === false`, and the field is `omitempty`, so the server never
   sends `false`. A spell or ability source the viewer cannot read still opens
   a blank zoom panel.
6. **The bots read the flag in two places** and would read a face-down
   permanent as "unknown" rather than as the 2/2 it is.

The decisions below are one state, one viewers rule, one projection point,
one suppression predicate and one reset point. No card gets a special case.

## Decisions

### 1. One face-down state: the flag keeps a kind, and the kind is the rule

`Card.FaceDown bool` stays — it is what the wire, the client and the snapshot
already key on — and gains

```go
// FaceDownKind is WHY this object is face down. Meaningless (and
// always "") while FaceDown is false.
type FaceDownKind string

const (
    FaceDownNone       FaceDownKind = ""
    FaceDownExiled     FaceDownKind = "exiled"     // CR 406.3  — Necropotence
    FaceDownForetold   FaceDownKind = "foretold"   // CR 702.143b
    FaceDownManifested FaceDownKind = "manifested" // CR 701.40
    FaceDownMorphed    FaceDownKind = "morphed"    // CR 702.37
    FaceDownDisguised  FaceDownKind = "disguised"  // CR 702.168
    FaceDownCloaked    FaceDownKind = "cloaked"    // CR 701.58
)
```

Two fields rather than one because they answer different questions and are
read by different code: `FaceDown` is "is there a back showing" (the wire, the
client, the redaction), `FaceDownKind` is "which rule put it there" (the
engine). They are kept in step by construction: nothing sets either field
directly. `Card.SetFaceDown(kind)` and `Card.ClearFaceDown()` are the only
writers, `SetFaceDown("")` is `ClearFaceDown()`, and `FaceDown == true` with
an empty kind is unrepresentable through them.

The field is declared with the other string fields, **above** the bool block
at the end of `Card` — a `string` is 16 bytes at alignment 8 and dropping one
into the bool block strands a bool, which `card_layout_test.go`'s
`TestCardHasNoInteriorPadding` (#620 / #633) fails on.

**Why a kind and not a zone test.** The engine has no `Card.Zone`; a card
knows nothing about where it is. Every question this ADR answers — who may
look, is there a 2/2, is the catalog suppressed — is answered from the kind
alone, so none of the read points needs a zone or a `*Game`. That is what
lets the projection live on a `Card` method and the suppression live in
`CatalogKey`.

Two of the six kinds are exile states (`exiled`, `foretold`) and four are
CR 708.2 object states (`manifested`, `morphed`, `disguised`, `cloaked`).
`Card.FaceDownIsPermanent()` is that partition, and it is the **one
predicate** decisions 2 and 3 are both written against.

Suspend (#659) exiles face **up** with time counters (CR 702.62b), so it
takes no kind here; it is listed above only because #659 and #658 share the
special-action verb.

### 2. Who may look: derived from the kind, written into `KnownBy`

`KnownBy` is what the wire projection reads (S13.5, #646) and it stays the
single source of truth. The face-down rule does not add a second visibility
channel — it **decides what `KnownBy` is set to** when the face-down state is
entered:

| kind | may look | rule |
|---|---|---|
| `exiled` | nobody | CR 406.3 — "no player may examine it" |
| `foretold` | the **owner** | CR 702.143d |
| `manifested`, `morphed`, `disguised`, `cloaked` | the **controller** | CR 708.5 |

> **2026-09-23 — a row added by [ADR 0091](0091-hideaway.md) (#1331).**
> `hideaway` | the **controller of the permanent that exiled it** | CR 702.75a.
> The first row whose answer is another object's controller, so it is read
> through `Card.HiddenBy` and kept current by a state-check sweep as that
> permanent changes hands; CR 406.3 keeps anyone who has already looked.

`faceDownViewersLocked(c)` is that table and it is called from exactly the two
places that put a card into a face-down state (the exile route, decision 5, and
the battlefield entry, decision 7). The old `ClearKnown()` is the
`exiled` row of the same table, not a separate behaviour.

Three consequences worth stating:

- **Sticky knowledge is not a leak here, it is the point.** `KnownBy` is
  sticky across zone moves by design, and a morph's controller knowing their
  own card is exactly that. What the face-down write does is *replace* the
  set, not add to it: a scryed library card carried into a face-down exile
  would otherwise leave one seat able to read a card the rules say nobody can
  (the reason `ClearKnown` was there in the first place).
- **`markCardKnownInZoneLocked` must not run for a face-down landing.** Exile
  and the battlefield are public zones, so the ordinary path marks every seat.
  Both face-down writers skip it; that is the one line that separates
  "public zone" from "public object".
- **Spectators and admins keep seeing everything** (`FilterViewFor`'s existing
  contract: the empty viewer ID knows every card). A face-down object is
  hidden from *players*, and the sandbox's spectator view is a debugging
  surface, not a seat. Stated because #656 asked; unchanged from develop.

### 3. Characteristics: the CR 708.2 body is the printed characteristic

A face-down **permanent** (and, later, a morph or disguise on the stack) *is*
a 2/2 creature with no name, no text, no subtypes, no mana cost and no
colour. That is not an effect applied to the real card; CR 708.2 says the
object has those characteristics, full stop. So it goes in at **layer 0** —
`Card.printedCharacteristic()`, the baseline the whole CR 613 pass is applied
to:

```go
func (c Card) printedCharacteristic() Characteristic {
    if c.FaceDownIsPermanent() {
        return faceDownCharacteristic(c)
    }
    …
}
```

`faceDownCharacteristic` returns `Power: 2, Toughness: 2, Types: ["Creature"]`
and everything else zero. The controller baseline (`baseController()`) is
carried through unchanged — layer 2 is orthogonal to this.

**Disguise and cloak's ward {2} does NOT ride `Characteristic.Abilities`**
(CR 702.168a, CR 701.58a). #656's checklist asked where it hangs "once the
catalog key is blank", and the answer is not "as a granted keyword": this
engine has no `ward` token in `canonicalKeywords` and deliberately so —
the table is closed, a keyword joins it in the change that teaches the engine
to honour it, and *"a bare token has nowhere to put the cost"*
(`game/keywords.go:37-97`, `cards/effects/attachments_batch3.go:204-211`).
Ward is a structured CR 702.21a **triggered ability** built by
`effects.Ward(WardCost, label)` (`cards/effects/ward.go:68-138`) watching
`EventBecomesTarget`. So the face-down projection's ward is a
`game.TriggeredAbility` the trigger harvester takes from the face-down object
itself, in the same read where decision 4 has just told it the catalog has
nothing — a `faceDownTriggeredAbilities(c)` beside `faceDownCharacteristic(c)`,
returning the ward ability for `disguised` and `cloaked` and nothing for the
other four kinds. **Designed, not built:** nothing creates a disguised or
cloaked object until #95, `Ward` lives in the catalog package which `game`
cannot import, and inverting that is a boundary change that must not ride
along in this PR. It is written down here so #95 does not have to rediscover
that the obvious answer (a keyword string) is the one this tree has already
rejected.

**Why layer 0 and not a layer-1 copy override.** Because every later layer and
every reader then sees the 2/2 for free: an anthem pumps it to 3/3, a
"creatures you control get +1/+1" predicate counts it, `Effective()` returns
it, targeting sees a legal creature, and the wire ships 2/2 with no further
plumbing. A layer-1 override would have to be re-asserted by every
recompute and would still leave the off-battlefield readers below wrong.

**The three accessors that bypass `Characteristic` are patched to the same
function.** `HasCardType`, `HasSubtype`, `HasSupertype` and `EffectiveColors`
each take a fast path off the printed fields when the layer cache is cold
(`c.effective == nil`) — deliberately, and the comments say so — and
`HasKeyword` reads `Card.Keywords` directly. Each grows one guard that
delegates to `faceDownCharacteristic`, so there is still one definition of
the 2/2 and five call sites of it, rather than five definitions. They are
listed here because "the projection is in one place" would otherwise be a
half-truth: the *value* is in one place, the *reads* are where they always
were.

`PrintedIsCreature` / `PrintedIsLand` are **not** patched. They are the
copiable-value surface (CR 707.2) and answer "what does this card say",
which is still the real card.

**A face-down card in exile has no characteristics at all** (CR 406.3a) and
gets **no** projection. Nothing in the engine reads the characteristics of an
exiled card for rules purposes, and the one thing that reads them for the wire
is the owner's own view of their own foretold card, which must show the real
card because the owner may look at it (CR 702.143d). Giving an exiled
face-down card a 2/2 body would be inventing a creature in exile. This is the
whole reason decisions 2 and 3 split on `FaceDownIsPermanent()` rather than on
`FaceDown`.

### 4. Catalog suppression: `CatalogKey` returns the empty key

CR 708.2a: a face-down permanent has **no text**, and therefore no triggered,
activated, mana, static or replacement abilities, no cost modifiers, no
"as enters" hook, no printed keywords and no catalog entry of any kind.

`CatalogKey(c Card)` is the one function in the tree that turns a `Card` into
a catalog key, and `catalogDef("")` already returns nil for the empty key —
every `Catalog*` reader already treats that as "this card has no entry". So
the suppression is one predicate at the lookup:

```go
func CatalogKey(c Card) string {
    if c.FaceDownIsPermanent() {
        return ""   // CR 708.2a
    }
    …
}
```

Roughly 45 production call sites inherit it, including the four that
deliberately bypass `CatalogAbilityKey` (the layer pass's static gather, the
LTB trigger's last-known identity, `HasKeyword`, the ETB hook). Putting the
predicate one level lower, in `CatalogAbilityKey`, would have missed all four,
and putting it in `catalogDef` is impossible — that function sees a string,
not a card.

**Exile is deliberately not suppressed.** A `foretold` card in exile keeps its
catalog entry, because the cast out of exile needs `CastableZones`,
`AlternativeCosts` and `Targets` to price and validate it. Nothing off the
battlefield runs a trigger, static or replacement off that entry (CR 113.6,
and the harvester does not scan exile), so keeping it costs nothing and
suppressing it would break #658 before it is written.

**Turning face up (designed, not built).** The `turn_face_up` special action
(CR 116.2g, the verb is ADR 0062 §4) is `ClearFaceDown()` plus a
`markCardKnownInZoneLocked` plus `RecomputeLayersLocked`. The moment the kind
goes, `CatalogKey` answers again and the real def is back — no restore step, no
cached def to invalidate, because nothing was ever stored. "When this permanent
is turned face up" triggers (CR 708.8) are an `EventTurnedFaceUp` emitted by
that same action; the event kind is not added here because nothing emits it.

### 5. Lifecycle: `MoveCard` is the one reset point (CR 400.7)

A card that changes zones is a **new object** with no memory of the old one
(CR 400.7), and "face down" is a property of an object in a zone. So:

```go
// MoveCard, unconditional, for every source and every destination:
c.ClearFaceDown()
```

alongside the resets `MoveCard` already owns (tapped, counters, damage,
attachment, the exile grant, the DFC face). That is #697's whole fix: the
sandbox `move_card` action and the cast-from-exile push both call `MoveCard`,
so both stop leaking without either of them learning about face-down, and so
does every caller added later.

**The destination sets it back, after the move.** `zone_route.go`'s face-down
exile branch already ran after `MoveCard`; it now calls `SetFaceDown(kind)`
plus the decision-2 viewers instead of `FaceDown = true` plus `ClearKnown()`.
The battlefield entry (decision 7) does the same in its own entry loop. This
is what makes "a foretold card stays face down while it sits in exile" true
without a special case: a move **within** exile is not a move at all
(`moveCardByRefLocked` returns early on a same-zone move, and the route never
re-enters the same zone).

Three now-redundant clears are **deleted**: `zone_route.go`'s two `FaceDown =
false` arms (the library arm and the default arm) and `battlefield_put.go`'s
clear on the exile→battlefield return. `entry_tail.go`'s clear stays — that is
a *new instance* being minted for a returning exile, not a move, and it resets
the whole card.

**A face-down permanent that leaves the battlefield is revealed** (CR 708.9),
through the S22 reveal frame rather than through anything new:
`RevealForEffect` (`game/reveal.go:84`) already marks every seated player a
knower and emits one `EventRevealCards`, and it moves nothing — a reveal is
not a zone change, which is exactly the shape CR 708.9 wants. It is called
from the shared exit primitive after the move, on the card's *pre-move*
face-down state, so the reveal names the card in its destination zone even
when that zone is a hidden one (that is what "its owner reveals it" means).
It is in the route rather than in `MoveCard` because `MoveCard` is a pure
two-zone function with no `*Game` and no seat list.

**Foretold cards are revealed at game end and when their owner leaves**
(CR 702.143f). `revealFaceDownOwnedByLocked(playerID)` is the hook, and it
runs from `leave_game.go`'s `leaveGameObjectsLocked` **before**
`removeObjectsOwnedByLocked` sweeps the leaver's cards out of exile — after
the sweep there is nothing left to reveal. The game-end half is left for #658:
there is no game-end sweep to hang it on, and no card is foretold until #658
ships.

### 6. Wire and client: three fields, and the client keys on the permission

`CardView` grows two fields beside the `face_down` it already has:

```go
FaceDownKind string `json:"face_down_kind,omitempty"` // which CR 708 state
FaceVisible  bool   `json:"face_visible,omitempty"`   // may THIS viewer look?
```

`face_down_kind` is public — everyone at the table can see that a permanent is
a morph and that an exiled card is foretold. `face_visible` is the per-viewer
answer to decision 2, stamped by the filter beside `known_by_you` and never
trusted from the input, and is exactly `face_down && known_by_you`. It is on
the wire rather than derived client-side because it is the rules permission and
the client should not be re-deriving a visibility rule from two fields.

**A face-down permanent's 2/2 is public and survives redaction.** This is the
one place the #646 redaction rule needed extending. `redactCardForViewer`
strips `name`, `type_line`, `colors`, `power`, `toughness` and `abilities`
because they identify the card — but on a face-down permanent those fields
carry the CR 708.2 projection, which identifies nothing and which every player
can see. So the projection is re-stamped after the redaction for a face-down
permanent, from `game.FaceDownPublicCharacteristic`, rather than by widening
the allowlist: the `face_down_view_test.go` zone × viewer table stays exactly
as strict as #646 made it for every other card, and the face-down permanent
gets its own cells. Everything that names the card — `scryfall_id` (the art),
`mana_cost`, `faces`, `layout`, `oracle_id`, `auto`, `target_mode`, the ability
lists — is still stripped from a non-knower, and `CatalogKey` suppression
(decision 4) means most of it was never stamped.

**Client.** `showsCardBack` (`client/src/lib/cardBack.ts`, #646) is unchanged
— `face_down && known_by_you !== true` and `face_visible` agree by
construction. `Card.svelte` renders the real face plus a small "face-down"
badge when `face_visible` is true, so the owner of a foretold card and the
controller of a morph see what it is and that it is face down; everyone else
sees a back with the kind as its label. `StackOverlay.previewCardFor` swaps
`c.known_by_you === false` for `c.known_by_you !== true` plus
`showsCardBack(c)` — the #697 half. **Both**, and in that order: the second is
the `Card.svelte` rule, which draws a back only for a card that is *face down*
and unknown, while the symptom #697 describes is a redacted stack spell that
is not face down at all and rendered as a blank FRONT. "Not a knower" is what
makes the panel blank, so that is the guard; consulting `showsCardBack` as
well keeps the overlay and the card from ever disagreeing. The overlay has no
test file and mounting it for one guard is not worth it, so the rule is lifted
into `client/src/lib/stackPreview.ts` and tested there — the same
logic-extraction the `cardBack.ts` fix used, and the dominant pattern in
`client/src/lib`.

**Bots** read a face-down permanent as what it is. `heuristic/score.go`'s
`permanentValue` currently prices `!KnownByYou && (FaceDown || Name == "")` as
`Unknown`; a face-down permanent now arrives as a 2/2 creature with a type
line, so it is priced as the vanilla creature it is, and the `Unknown` arm is
left for cards that really are unreadable. `model/prompt.go`'s `cardName` says
"a face-down card" for anything the seat cannot look at — it keys on
`face_visible` rather than on `known_by_you`, so a bot never reads a hidden
face and never mislabels its own morph.

### 7. `ManifestForEffect`: the primitive, shipped; manifest the mechanic, not

CR 701.40a — "put the top card of your library onto the battlefield face down
as a 2/2 creature" — is the smallest real thing that makes a face-down
permanent, and it is `putOntoBattlefieldFromZoneLocked` with one option:

```go
type ZoneEntryOptions struct {
    …
    FaceDown FaceDownKind  // enter face down in this state
}
```

which (a) skips the CR 110.4 nonpermanent refusal, because manifest works on
any card, (b) sets the face-down state and the decision-2 viewers in the entry
loop instead of clearing the flag and marking every seat, and (c) therefore
makes the ETB hook and every EventETB trigger see the empty catalog key. Three
short edits in one file, and the whole CR 614 entry pipeline, the replacement
window and the undo path come with it.

It ships **because the model is untestable without it.** A test-only helper
that sets three fields proves the projection and nothing about the entry
pipeline — not that a face-down entry runs no ETB trigger, not that the
controller is the only knower on a public zone, not that the reveal fires on
the way out. Morph, megamorph, disguise and cloak stay deferred to #95: they
need the `turn_face_up` action, the face-down **cast** (CR 708.4, a spell on
the stack with no characteristics), and the morph cost. None of that is here.

### 8. Undo, clone and persistence

`FaceDownKind` is **`carried`** — per-object state an undo must restore
exactly, like `FaceDown` itself — and it is declared in
`snapshot_drift_test.go`'s `cardFields` plan, whose reflection walk fails the
build for any `Card` field with no disposition. Being a string type it rides
`clone.go`'s `out := c` value copy with no entry of its own, exactly as
`NamedTribe`, `ChosenColor` and `Layout` do. It joins `FaceDown` on
`cardSnapshot` as `faceDownKind` and in both halves of the mirror
(`snapshotCard`, `restoreCard`), and in `snapshot_test.go`'s `enrich` fixture
so the exact round-trip test actually covers it.

**No schema bump.** `SnapshotSchemaVersion` stays 1: the policy at
`snapshot.go:67-82` bumps only when an older file would decode into a game
that is *subtly wrong*, and "adding a field that zero-values correctly does
NOT need a bump".

**Old-snapshot migration.** A pre-ADR snapshot with `faceDown: true` and no
`faceDownKind` is a Necropotence exile — the only face-down object that could
exist before this ADR. `restoreCard` stamps `FaceDownExiled` for that case, so
an old save comes back with the visibility it had (nobody may look) rather
than as a face-down object with no rule attached and no viewers table row.

**The census does not move.** `TakeCensus` counts catalog specs and
`Completeness` declarations only; an engine field changes nothing in it
(`cards/coverage/census.go:113-130`).

## Consequences

- **#697 closes on the one-line reset**, and every future caller of `MoveCard`
  is covered by it rather than by a code review.
- **#658 (foretell) becomes implementation.** The special action exiles with
  `SetFaceDown(FaceDownForetold)`, the owner is the viewer by the decision-2
  table, the cast out of exile is face up because `MoveCard` says so
  (CR 406.3a), the wire and the client already draw it, and the CR 702.143f
  reveal hook exists.
- **#95 (morph family) becomes implementation** of the *cast* and the
  *turn-face-up action*. The object model, the 2/2, the suppression, the ward
  hook and the reveal are here.
- **A face-down permanent that leaves the battlefield loses its LTB trigger
  the right way** (CR 708.2a: it has none) only while it is still face down.
  The LTB harvest reads the card *after* `MoveCard` has cleared the flag, so a
  face-down permanent's real dies-trigger would fire. Recorded as a known gap
  rather than fixed here: nothing in the engine makes a face-down permanent
  with a dies-trigger, and the fix belongs with the last-known-information
  work the harvester already does for CR 603.10.
- **`CatalogKey` returning "" is silent.** A future call site that should have
  seen the real card on a face-down permanent gets "no catalog entry" rather
  than an error. That is the same failure mode the function's own doc comment
  already warns about for a forgotten `CatalogKey`, and it is the correct
  default for CR 708.2 — the risk runs toward *less* text, never more, which
  is the direction #259 says the sandbox must always err.

## Alternatives considered

**One field, `FaceDownKind`, with `FaceDown` derived.** Cleaner on paper.
Rejected because `FaceDown` is already on the wire, in the snapshot, in the
client and in two bot readers, and because the bool is the thing three of those
four actually want; a derived accessor would have meant touching all of them
to ask a question they already have the answer to.

**Suppress the catalog at `CatalogAbilityKey`.** It is the *documented*
accessor for "what does this permanent DO" and it already carries the CR 613.1f
ability-removal arm, so it reads like the right home. Rejected: four ability
readers deliberately bypass it (the layer pass's static gather, the LTB
trigger, `HasKeyword`, the ETB hook), each for a good reason, and each would
have needed its own face-down arm. `CatalogKey` is below all four.

**Materialise the 2/2 onto the card, `PrintedValues`-style (ADR 0043).**
Rejected: it would have to be undone on turn-face-up and on every zone change,
which is two more reset points for a value that is a pure function of one
field.

**A separate `FaceDownViewers []uuid.UUID` on `Card`.** Rejected: it is a
second visibility channel next to `KnownBy`, and the wire projection reads
`KnownBy`. Two sources of truth for "who can see this" is the bug class this
ADR exists to close, not one to add.
