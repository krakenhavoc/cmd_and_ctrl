# ADR 0082 — Casting face down, and turning face up (CR 708.4, CR 116.2g)

**Status:** Accepted · 2026-09-22 · S43 — Hand special actions and face-down objects
**Issues:** [#1194](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1194) (this ADR),
[#95](https://github.com/krakenhavoc/cmd_and_ctrl/issues/95) (the morph family),
tracker [#886](https://github.com/krakenhavoc/cmd_and_ctrl/issues/886)
**Numbering:** on 2026-09-22 every remote head was fetched (406 branches) and
its `docs/decisions/` listed with `ls-tree` per AGENTS.md §4. The highest number
present on any branch was `0081` (earthbend and object-keyed delayed triggers),
so this ADR takes **0082**. 0005, 0024, 0029 and 0030 stay permanently unused.
**Builds on:** [ADR 0069](0069-face-down-objects.md) (the face-down OBJECT
model this ADR turns into MECHANICS — the kind, the viewers table, the CR 708.2
body at layer 0, `CatalogKey` suppression, `MoveCard`'s reset, the CR 708.9
reveal, `ZoneEntryOptions.FaceDown`),
[ADR 0062](0062-abilities-and-special-actions-from-the-hand.md) §4 (the
`special_action` verb and the per-kind timing table, which reserved
`turn_face_up`),
[ADR 0066](0066-granted-cast-and-play-permissions.md) (granted casts),
[ADR 0073](0073-optional-additional-costs-and-the-cast-gate.md) (the announce
order and the one cast gate),
[ADR 0034](0034-multi-face-cards.md) (`CatalogKey`, the face materialisation)

## Context

ADR 0069 shipped the face-down **object**: what one *is*, who may look at it,
what its characteristics are, and what happens when it moves. It shipped one
thing that makes one — `ManifestForEffect` — and said so in that function's own
doc comment: "MANIFEST THE MECHANIC IS NOT SHIPPED."

What is missing is every way a *player* reaches that object, and the way back
out of it:

| | rule | missing |
|---|---|---|
| cast it face down for {3} | CR 708.4, CR 702.37b | a cast that produces an object with no text |
| put it there from a library | CR 701.34a | the effect-side primitive exists; no card calls it |
| turn it face up | CR 116.2g, CR 708.6 | the special action, and what it costs |
| megamorph's counter | CR 702.109b | — |
| disguise / cloak's ward {2} | CR 702.168a, CR 701.58a | `FaceDownKind.HasWard` has no consumer |

and one question none of the five answers on its own, which is the question
this ADR exists for: **a face-down permanent has no text, so where does the
price of turning it face up come from?**

36 cards wait on the row (`docs/engine-seams.md`, "Face-down objects
(CR 406.3a / 708)").

## Decisions

### 1. Casting face down is an ALTERNATIVE COST that stamps the object

CR 702.37b is two permissions on one keyword: "you may cast this card as a 2/2
face-down creature spell for {3}", and "any time you have priority, you may
turn this permanent face up by paying its morph cost." The first half is,
exactly and without stretching, an alternative cost (CR 118.9): a price paid
**instead of** the mana cost, chosen at CR 601.2b, that changes what the spell
is.

So it rides `game.AlternativeCost`, which already carries five other clauses a
keyword staples to a price (overload's cleared targets, cleave's swapped
clause, evoke's sacrifice trigger, warp's exile, escape's counter). One new
field:

```go
// FaceDown is CR 708.4: this cost casts the card FACE DOWN.
FaceDown *FaceDownCast
```

```go
// FaceDownCast is the CR 708 half of morph, megamorph and disguise:
// WHICH face-down state the cast produces, and what the permanent
// costs to turn face up again.
type FaceDownCast struct {
    Kind          FaceDownKind // morphed (CR 702.37b) or disguised (CR 702.168a)
    FaceUpCost    string       // the morph / disguise cost (CR 702.37b)
    FaceUpCounter bool         // megamorph's +1/+1 counter (CR 702.109b)
}
```

**One struct rather than three fields**, because `alt.FaceDown != nil` is then
the single predicate the cast path branches on, and because the three facts are
one keyword's worth of data that is meaningless apart.

**`FaceUpCost` lives on the CAST offer even though nothing pays it at cast
time.** That is decision 4, and it is the whole reason the struct is shaped
this way: a face-down permanent has no catalog entry, so the price of turning
it up cannot be read off the object. It has to be read off the **card**, and
the card's declaration of "I may be cast face down" is the only place in the
catalog that already knows which face-down state this card produces.

**The alternative-cost price is the KEYWORD's, not the card's.** Morph,
megamorph and disguise all cast for {3} (CR 702.37b, CR 702.109a, CR 702.168a),
so `effects.Morph("{1}{U}")` takes only the *face-up* cost and fills the `{3}`
in itself — the same division `effects.Foretell` already draws between
foretell's fixed {2} and the card's own foretell cost.

### 2. The object is stamped at announce, and every gate below reads the CR 708.2 object

`CastSpell` takes a **value copy** of the card out of the source zone and reads
it a dozen more times before anything moves (ADR 0034's "single
highest-leverage line"). So casting face down is one line on that copy, at
exactly one point in the announce order:

```go
// after resolveAlternativeCostLocked and validateCastPathLocked,
// before every gate that reads the card's text:
if alt != nil && alt.FaceDown != nil {
    card.SetFaceDown(alt.FaceDown.Kind)
}
```

Everything below that line then gets CR 708.2a for free, because `CatalogKey`
already answers `""` for a face-down permanent (ADR 0069 decision 4):

- no target mode, no `TargetSpec`, no clauses — a face-down Willbender takes no
  targets, which is the rule rather than a special case;
- no modes, no additional cost, no optional additional costs, no tap cost;
- no cost modifiers of its own;
- `card.IsCreature()` is true and `card.IsInstant()` is false, so the
  sorcery-speed gate demands sorcery timing — CR 708.4's "as a 2/2 creature
  spell";
- the per-turn cast tally counts it as a creature spell;
- `CastGateLocked` judges it as the creature spell it is, so "you can't cast
  creature spells" stops it and "you can't cast blue spells" does not
  (CR 708.2: a face-down spell is colourless).

**Why after the path validation and not before.** `resolveAlternativeCostLocked`
and `validateCastPathLocked` are the two gates that must read the REAL card: one
asks whether the card offers this key at all, the other whether the key may be
claimed from this zone. Stamping before them would erase the offer being
claimed. So the stamp goes between "the claim is legitimate" and "what is the
spell", which is also where CR 601.2b sits relative to CR 601.2c–f.

**`HasNoManaCost` is checked above the stamp and is unaffected**: a face-down
cast pays an alternative cost, so `castPaysPrintedCost` is already false.

### 3. The spell on the stack, and the permanent it resolves into

**On the stack.** The stack push already branches on "is this public"
(`markCardKnownInZoneLocked`). A face-down cast takes the other arm — the
CR 708.5 viewers, which for a morph is its controller and nobody else — through
the same `applyFaceDownLandingLocked` the exile route and the manifest entry
use. One writer, three callers.

**Onto the battlefield.** The fact rides the stack item and then the entry
event:

```go
StackItem.FaceDown        FaceDownKind // carried through a snapshot and a clone
ReplacementEvent.FaceDown FaceDownKind // seeded by the caller, read by the finisher
```

and `executeEntryToBattlefieldLocked` — the ONE finisher for every settled
battlefield entry (#653, #478) — takes the same two-armed branch the stack push
does. That is the whole of "the permanent enters face down": the CR 614 window,
the pause-and-resume, the counters, the provenance stamp, the zone-move event
and the undo path are the ones every other permanent gets.

**A face-down entry runs no ETB trigger and no "as enters" choice**, and
nothing says so: the state is set before `fireETBHookLocked` reads
`CatalogKey(moved)`, which answers `""`. CR 708.2a falling out rather than
being special-cased, exactly as ADR 0069 designed it for manifest.

**`putOntoBattlefieldFromZoneLocked` moves its `opts.FaceDown` onto the event**
rather than keeping it in a local: the fact is then on the event for both entry
paths, and a replacement effect that inspects the event sees the same truth
whichever door the permanent came through.

### 4. The price of turning face up is read off the CARD, not off the OBJECT

This is the decision the rest of the ADR is arranged around.

CR 708.6 lets the controller turn a face-down permanent face up "by paying its
morph cost", and CR 708.2a says the permanent has **no text**. Those are not in
tension in paper, because the player is holding a card they are allowed to look
at (CR 708.5) and the rule points at *the card*, not at the object. They are in
tension here, because `CatalogKey` is the engine's one door from a `Card` to its
text and ADR 0069 deliberately shut it.

So the door opens once, narrowly, with a name that says what it is:

```go
// faceUpCatalogKey is the catalog key of the CARD UNDERNEATH a
// face-down object — CatalogKey with the CR 708.2a suppression lifted.
//
// EXACTLY ONE RULE may ask, and it is CR 708.6: "you may turn this
// permanent face up by paying its morph cost". That cost is printed on
// the CARD, and the player reading it is the one CR 708.5 allows to
// look. Every other reader wants the OBJECT and must keep using
// CatalogKey, which answers "" — a face-down permanent has no text.
func faceUpCatalogKey(c Card) string
```

Unexported, in `face_down.go`, with one caller (`TurnFaceUpOffer`). A second
caller is a rules bug, and the doc comment says so.

**What it does NOT do is restore the card's abilities.** Nothing is cached and
nothing is materialised: the read is a pure function of the card, taken at the
moment the offer is priced. The instant the kind goes, `CatalogKey` answers on
its own again and the real def is back — ADR 0069 decision 4's "no restore step,
no cached def to invalidate, because nothing was ever stored" holds unchanged.

### 5. The offer is DERIVED per kind, and it is one function

There is no catalog declaration of "this permanent may be turned face up".
There cannot be: a manifested Mountain has no such declaration and must not be
turnable, and a manifested Grizzly Bears has none and must be. The rule is a
function of the KIND and the CARD:

| kind | may be turned up | cost | rule |
|---|---|---|---|
| `morphed` | if the card declares a morph cast | its morph cost | CR 702.37b |
| `disguised` | if the card declares a disguise cast | its disguise cost | CR 702.168b |
| `manifested` | if the card is a **creature card** | its mana cost | CR 701.34d |
| `cloaked` | if the card is a **creature card** | its mana cost | CR 701.58b |

```go
// TurnFaceUpOffer is CR 708.6's answer for ONE face-down permanent:
// may this be turned face up, and at what price. nil means "no", which
// is a manifested Mountain, an Ixidron'd Sheoldred (CR 708.7) and every
// face-up permanent in the game.
func TurnFaceUpOffer(c Card) *SpecialAction
```

It returns a `SpecialAction` with `Kind: turn_face_up`, so the engine, the
legal-move enumerator and the wire projection all read the offer through
`SpecialActionOffered`, which they already do for foretell and suspend. Three
readers, one rule, and the rule is in the engine.

**"Creature card" is `PrintedIsCreature`, deliberately.** ADR 0069 left that
accessor unpatched because it is the copiable-value surface (CR 707.2) and
answers "what does this card say" — precisely the question CR 701.34d asks. The
face-down projection would answer "yes, it's a 2/2 creature" for a manifested
Island, and that is the wrong question.

**CR 701.34e is out of scope and caveated.** A manifested card that also has
morph may be turned face up for *either* its mana cost or its morph cost. This
ADR's offer is one row per permanent; the second row would need a per-offer key
on the wire and a picker in the client, for a case no card in the catalog
reaches. Recorded here rather than discovered later.

### 6. Turning face up is a special action on a BATTLEFIELD card

ADR 0062 §4's verb is "one entry, one timing table, no stack" and it is right
about all three. What it assumed, because both of its kinds were hand keywords,
is that the card is in the actor's hand — `PerformSpecialAction` scans `p.Hand`
and the enumerator walks it.

That assumption becomes a **per-kind zone**, which is the same shape the file
already uses for the timing and the performer:

```go
// specialActionZoneLocked is WHERE a kind's card lives.
//   foretell, suspend   the actor's HAND   (CR 702.143a, CR 702.62a)
//   turn_face_up        the BATTLEFIELD    (CR 708.6)
```

Three per-kind tables (zone, timing, performer) and nothing else forks — which
is what ADR 0062 §4 promised, and is still true with a third kind that is not a
hand keyword at all.

**Ownership vs control.** The hand kinds get "is this yours" for free, because a
hand holds only its owner's cards (CR 108.4). The battlefield does not, so the
battlefield arm checks `Controller == playerID` — CR 708.6 says *controller*,
and a stolen morph is turned up by the thief.

**Timing: any time the actor has priority, INCLUDING under split second.**
CR 702.37c says "any time you have priority"; CR 702.61b stops players casting
spells and activating abilities that are not mana abilities and says nothing
about special actions. So the row is `return true`, and it is the second row of
the table (after foretell) whose whole content is the split-second asymmetry
ADR 0062 §4 wrote the table down for.

**Cost.** `PerformSpecialAction` already pays `sa.Cost` through
`payAbilityManaCostLocked` with the **zero** `ManaSpendContext` — mana
restricted to casting spells or activating abilities cannot pay for a special
action, because it is neither. Already conservative in the #259 direction, and
it needs no change for a morph cost.

### 7. Turning face up is not a new object, and the trigger fires off an event

CR 708.8: "turning a permanent face up doesn't cause it to become a new
object." So `turnFaceUpLocked` touches nothing that identifies the object —
`InstanceID`, `ObjectEpoch`, counters, damage, attachments, combat state,
`EnteredBattlefieldAt` and `SummonedThisTurn` all ride through. A morph that is
attacking stays attacking; a morph that has been out since last turn can attack
the moment it is turned up. It is four lines:

```go
c.ClearFaceDown()                                  // CR 708.6
g.markCardKnownInZoneLocked(g.Battlefield, cardID) // the battlefield is public again
g.InvalidateLayersLocked()                         // the real characteristics are back
g.EmitEvent(Event{Kind: EventTurnedFaceUp, ...})   // CR 708.8's trigger
```

plus megamorph's `+1/+1` counter (CR 702.109b) between the clear and the event,
so a "when turned face up" trigger already sees it.

**`EventTurnedFaceUp` is a new event kind, not a reuse.** The candidates were
`EventTransform` (ADR 0079) and `EventETB`, and both are wrong in the same
direction: they mean "a different object is here now" and "an object arrived",
and CR 708.8 is the one transition in the game that explicitly means neither. A
card reading "whenever this transforms" would fire on a morph under the reuse,
which is a rules bug in the loudest possible place.

**The trigger needs no new shape.** `effects.WhenTurnedFaceUp` is an ordinary
`game.TriggeredAbility` watching the new kind, on the battlefield, with the
default zone list — because the harvester reads the source's abilities through
`CatalogKey`, and by the time the event is emitted the permanent is face up and
its catalog entry answers again. The ORDER is the rule: clear first, emit
second. Emitting first would harvest against an object with no text and drop
the very trigger CR 708.8 exists for.

### 8. Ward while face down is a hook, not a keyword

ADR 0069 decision 3 settled this, and it is restated because this ADR is where
it is built. Disguise and cloak give the face-down object **ward {2}**
(CR 702.168a, CR 701.58a), and it is a CR 702.21a *triggered ability*, not a
token in `canonicalKeywords` — that table is closed, and a bare token has
nowhere to put the cost.

`effects.Ward` lives in the catalog package, which `game` cannot import, so it
arrives the way every other catalog fact does: a function variable the effects
package sets at init, read from `TriggersForCard` and gated on
`FaceDownKind.HasWard()` — the predicate ADR 0069 shipped with no consumer.

### 9. The view: the 2/2 is public, the button is the controller's

The redaction is already right (ADR 0069 decision 6): `redactCardForViewer`
strips everything that names the card and `stampFaceDownPublicBody` puts the
CR 708.2 projection back, on the battlefield as in exile. Nothing changes there.

What changes is one projection. `special_actions` is stamped only onto cards in
the viewer's own **hand** today. It is now stamped onto a face-down
**battlefield** permanent as well, from the same `viewOfSpecialActions` and with
the same per-kind `available` answer, so the client greys a row rather than
re-deriving a timing rule.

It leaks nothing, for the reason the hand rows do not: the redaction clears
`special_actions` for every non-knower, and the only knower of a face-down
permanent is its controller (CR 708.5).

**The client renders the back already** — `showsCardBack` is
`face_down && known_by_you !== true`, which is zone-independent and has been
since #646. The one client change is that the context menu's special-action
section, built today only for a hand card, is built for a battlefield card too.

### 10. What is NOT here

- **Hideaway** (Mosswort Bridge) is a face-down EXILE with a cast permission,
  not a CR 708.2 object. It is the foretell shape (ADR 0066 + ADR 0069), and a
  neighbour of #1194 rather than part of it.
- **Ixidron** turns other creatures face down without a cast. That is a
  primitive — "turn target permanent face down" — and CR 708.7 says such a
  permanent can never be turned face up, which `TurnFaceUpOffer` already answers
  `nil` for. The primitive is not built here.
- **CR 701.34e** (a manifested card with morph, turnable two ways) — decision 5.
- **A face-down permanent's LTB trigger** still fires off the real card, because
  the harvester reads the card after `MoveCard` cleared the flag. ADR 0069
  recorded it as a known gap and it stays one; it belongs with the CR 603.10
  last-known-information work.

## Consequences

- **`AlternativeCost` grows a sixth staple clause** and the cast path grows one
  line. Every later keyword that casts a card as something it is not has a
  precedent to copy.
- **`SpecialAction` covers a permanent for the first time**, and ADR 0062 §4's
  "one verb" claim survives it: three per-kind tables, no fork.
- **A face-down permanent is priced by a read that is deliberately narrow.**
  `faceUpCatalogKey` is the one place in the tree that looks through the
  CR 708.2a suppression, and a second caller of it is a bug.
- **`EventTurnedFaceUp` is emitted by exactly one function.** If a "turn face
  down" primitive lands later (Ixidron), it emits nothing new — CR 708.7 gives
  those permanents no way back up.

## Alternatives considered

**A `turn_face_up` ABILITY on the permanent rather than a special action.**
Rejected because CR 116.2g is explicit that it does not use the stack and cannot
be responded to, and because an ability on a face-down permanent is exactly what
CR 708.2a says does not exist. The engine would have had to special-case "this
activated ability survives the no-text rule", which is the suppression's whole
point.

**A catalog declaration of the face-up cost, separate from the cast offer** — a
`Spec.Morph *MorphSpec` beside `Spec.AlternativeCosts`. Rejected: a card would
then declare the same keyword twice and could declare half of it — a morph cost
with no face-down cast, or a face-down cast nothing can undo — which is the
half-a-card failure ADR 0037 §5 forbids and `effects.Register` would have had to
grow a cross-check for. One declaration cannot be half-written.

**Materialising the real characteristics back onto the card on turn-face-up.**
Rejected for the reason ADR 0069 rejected materialising the 2/2: `ClearFaceDown`
is already the whole of it, because the projection was a pure function of the
field the whole time.

**Letting the face-down cast go through a bespoke `CastFaceDown` entry point.**
Rejected: it would have to reproduce the cast gate, the permission read, the
commander tax, the provenance stamp, the cast tally and `EventCast`, and the one
thing #259 cannot afford is two answers to "was this spell cast".
