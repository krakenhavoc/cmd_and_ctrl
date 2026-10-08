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
| megamorph's counter | CR 702.37b | — |
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
    FaceUpCounter bool         // megamorph's +1/+1 counter (CR 702.37b)
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
megamorph and disguise all cast for {3} (CR 702.37b, CR 702.37b, CR 702.168a),
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
use. One writer, three callers (four since #1209's turn-face-down).

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
| `manifested` | if the card is a **creature card** | its mana cost | CR 701.40b |
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
answers "what does this card say" — precisely the question CR 701.40b asks. The
face-down projection would answer "yes, it's a 2/2 creature" for a manifested
Island, and that is the wrong question.

**CR 701.40c is out of scope and caveated.** A manifested card that also has
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

plus megamorph's `+1/+1` counter (CR 702.37b) between the clear and the event,
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
  *Built by the 2026-09-23 amendment below (#1209), which also corrects the
  second half of this bullet: `nil` is right for an Ixidron'd Sheoldred and
  wrong for a Backslid morph (CR 702.37e).*
- **CR 701.40c** (a manifested card with morph, turnable two ways) — decision 5.
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
  *Superseded by the 2026-09-23 amendment (#1209): the primitive emits an event
  of its own, `EventTurnedFaceDown`, and CR 702.37e does give a permanent whose
  CARD prints morph a way back up whatever turned it over.*

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

## Amendment (2026-09-23, #1209): turning a permanent FACE DOWN (CR 708.2a) · Accepted · S46

**Status:** Accepted · 2026-09-23 · S46 · [#1209](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1209),
tracker [#886](https://github.com/krakenhavoc/cmd_and_ctrl/issues/886)
**Proof cards:** Ixidron, Backslide, Cyber Conversion, Master of the Veil

Decision 10 of this ADR said the primitive was not built here, and named
the rule it would have to obey. This amendment builds it. It is the last
thing on the face-down row that no keyword needs.

### A0. The rule numbers, read again

The August 2026 CR renumbered this section, and the issue that asked for
the work quotes the previous edition. What the tree cites now:

| what | rule | note |
|---|---|---|
| a face-up permanent turned face down becomes a nameless 2/2 | **708.2a** | not 708.3, which is about entering face down |
| a face-down permanent **can't** be turned face down | **708.2b** | "nothing happens" |
| a double-faced or melded permanent **can't** be turned face down | **712.16** | "nothing happens" |
| whether the rules that hid it also un-hide it | **708.7** | |
| who may look | **708.5** | |
| turn face up for a morph cost — **a card with a morph ability** | **702.37e** | disguise's twin is 702.168d |

**There is no token rule.** #1209's body says CR 708.3 makes a token that
is turned face down cease to exist. No such rule exists in CR 708 or in
CR 111, in this edition or the previous one. A token turned face down is
a nameless 2/2 that goes on sitting on the battlefield, and CR 704.5d —
which fires only off a token in a zone OTHER than the battlefield — is
unaffected. The belief probably comes from **Ixidron's own text**, which
since its Oracle update reads "turn all other **nontoken** creatures face
down": that is one card's restriction, not the rules'. Cyber Conversion
has no such word and hits a token happily.

The trap the non-existent rule was pointing at is real, though, and the
test for it is in `face_down_turn_test.go`: `Card.IsToken` reads the
PRINTED type line and not the CR 708.2 body, which has no Token supertype
in it. Had it read the projection, a face-down token would have stopped
being a token, survived CR 704.5d and handed its controller a real card
in hand — #596's bug, conjured a second way.

### A1. A seventh kind, `FaceDownTurned`

The alternative was to stamp `morphed` and let `TurnFaceUpOffer` answer
`nil` for a card that prints no morph, which decision 5 already does. It
is very nearly right and it is wrong twice, both times because
`FaceDownKind` records WHY:

- **It would price the wrong thing.** The morph arm of `TurnFaceUpOffer`
  demands that the card's declared face-down cast MATCH the state's kind.
  That is right for a cast — you cast it face down *using* that keyword,
  so the two cannot disagree — and wrong for a turn: CR 702.37e and
  CR 702.168d both key the turn-face-up permission on the CARD having the
  ability, whatever put the permanent face down. A disguise creature
  Backslid into a `morphed` state would be refused its own way back up.
- **`disguised` and `cloaked` would hand it WARD {2}.** That body is the
  one those keywords list (CR 702.168a, CR 701.58a). CR 708.2a's is "no
  text", full stop. `HasWard()` is false for the new kind, and that is not
  a detail: a Backslid disguise creature with ward {2} would be a
  protection the rules do not give it.

So the kind joins `IsPermanentState()` and gets nothing else: the viewers
rule, the 2/2 body, the catalog silence, the CR 708.9 reveal, the
snapshot and the clone all key on the partition rather than on the member,
and every one of them was correct for the seventh kind the moment it was
added. The only new arm anywhere is `TurnFaceUpOffer`'s.

### A2. `TurnFaceDownForEffect(source, ids...)` — variadic, because Ixidron

One function, one body, one event kind. The signature is
`phaseOutLocked`'s (ADR 0084) down to the shape of the batch, for the same
reason: "turn all other nontoken creatures face down" is ONE event in the
game, so the whole batch is turned over before the first `EventTurnedFaceDown`
goes out. A trigger that fired halfway through would read a board that
never existed.

It returns the IDs that actually turned, which Ixidron does not need and
a future card will: a permanent that is not on the battlefield is skipped
(CR 608.2b's per-slot re-check for free) and so is one the two refusals
name. Neither refusal is an `error` and neither emits an event, because
both rules say "nothing happens" in those words — and because the catalog
soak fails a game on any effect error, so a resolution that legally does
nothing must not throw.

**CR 712.16's predicate is `transform.go`'s.** `isDoubleFacedPermanent`
was extracted out of `CanTransform`, where its comment already called the
layout allowlist "the load-bearing line" — an `adventure` card also has
two `Faces` and is not a double-faced card. Two rules ask the same
question from opposite ends of the tree (CR 712.9 and CR 712.16) and two
copies of that allowlist would be two places to forget `modal_dfc`.

**What rides through**: `InstanceID`, `ObjectEpoch`, counters, marked
damage, tap state, attachments, the combat declarations,
`EnteredBattlefieldAt` and `SummonedThisTurn`. CR 613.7f ("a permanent
receives a new timestamp each time it turns face up or face down") is a
statement about ONE permanent, and CR 708.8 says the same for the other
direction. An Ixidron'd attacker goes on attacking as a 2/2. **The Aura
stays attached and may then die**: CR 704.5m is re-asked every SBA pass,
so an "enchant creature with flying" Aura on what is now a vanilla 2/2 is
put into its owner's graveyard by the same sweep that handles a creature
losing flying any other way — no line here.

*Not done, and filed so it is not rediscovered:* CR 613.7f's re-stamp
itself (**#1271**). The engine has no per-permanent face-change timestamp
and `turnFaceUpLocked` does not re-stamp either, so the two directions are
consistent with each other and both are wrong about layer ordering in the
same narrow way. It belongs with the CR 613.7 work, not here.

### A3. Who may look: the engine FORGETS, and a paper table does not

`applyFaceDownLandingLocked` — the one writer the exile route and the
face-down entry already share — REPLACES the knowledge set with CR 708.5's
answer, so the controller becomes the sole knower. This is the third
caller and the only one whose card was already sitting in its zone, face
up and public, a moment earlier.

It is a genuine narrowing and it is worth being honest about. Every player
SAW that creature; Ixidron's own ruling leans on CR 708.6's
differentiation rule to say that "all players must be able to figure out
what each of the creatures Ixidron turned face down is". Keeping the old
knower set would have modelled that exactly.

**Rejected**, on three grounds:

1. `KnownBy` has modelled "may look at" since ADR 0069 decision 2, not
   "saw once", and CR 708.5 is explicit: "you can't look at […]
   face-down spells or permanents controlled by another player."
2. CR 708.6 is a rule about telling two face-down objects APART at a
   physical table. A client that draws them as distinct objects in stable
   positions satisfies it without letting anyone read one.
3. It is what every digital client does, and the conservative direction
   for hidden information. A narrowing that turns out to be wrong shows
   up as a player asking; a widening that turns out to be wrong is a leak
   nobody reports.

### A4. Mutate, then announce — and the mirror argument does NOT apply

Decision 7 clears the face-down state BEFORE emitting `EventTurnedFaceUp`,
so that the trigger harvester — which reads a source's abilities through
`CatalogKey` — finds the card's text again. Read as "the readable state
comes first", that argument would put the emit before the mutation here.
It is not the argument.

The rule is that **an event is emitted after the change it reports**.
`layerVersionBump` must not be told about a change that has not happened,
and a listener earlier in the slice must not read face-up characteristics
off a card it has just been told is a 2/2 — which is the window
`turnFaceUpLocked` nils `Card.effective` at the mutation site to close,
and which this does too. Emitting first would bump the layer version
against the old state and let any listener that read `Effective()` during
dispatch cache the face-UP answer at the NEW version, where nothing would
invalidate it again.

What the order costs is the permanent's OWN "when this permanent is turned
face down" trigger: CR 708.2a has silenced it by the time the event goes
out, so the harvester cannot find it. That asymmetry belongs to the rules
rather than to this function — one direction restores text and the other
removes it — and no printed card has such an ability. CR 701.27b and
CR 701.28b contemplate one, which is exactly why `EventTurnedFaceDown` is
a KIND of its own rather than a direction flag on `EventTurnedFaceUp`: a
card that watched one direction must not fire on the other.

### A5. The log says what happened and names nobody

`LogTurnFaceDown` is narrated rather than silent, for `LogPhaseOut`'s
reason and not `LogTransform`'s: turning face down is not a zone change
and not a transform (CR 701.27b), so no other line says it, and the board
simply stops showing a card the table could read a second ago.

It names the SOURCE (public) and not the permanent — and that is not this
projection's decision. A CR 708.2 object has no name for ANY viewer, its
controller included: `CardView.Name` is the effective characteristic's
(ADR 0069 decision 6), and the controller gets the art through
`scryfall_id` and `face_visible` instead. So `resolveLogNames` finds no
name to put on the entry and the house fallback "a card" is the honest
rendering. `card_id` still rides the entry, so a client can point at the
permanent on the board — the half of the identity that IS public.

Carrying the real name in `Label` was considered and rejected: the
redaction only re-renders an entry whose name was DROPPED, so a Label on
an entry that never had a name would survive for every seat. That is a
second visibility model, which `resolveLogNames`' own doc comment
forbids.

### A6. What this does NOT build

- **CR 708.2's LISTED characteristics** (**#1270**). Cyber Conversion's
  "It's a 2/2 Cyberman artifact creature" and Yedora's "It's a Forest
  land" list an object's characteristics, which CR 708.2 allows and
  `FaceDownBody` cannot express — it is one shape per kind. Cyber
  Conversion ships with that declared; widening it is a change to ADR
  0069's object model, and Yedora's land shows it REPLACES the 708.2a
  body rather than decorating it.
- **CR 613.7f's timestamp** (**#1271**) — A2.
- **A "when this is turned face down" trigger on the permanent itself** —
  A4. No printed card has one.
- **CR 701.40c**, still, and hideaway, still — decision 10 is unchanged
  about both.

## Amendment (2026-09-23, #1270 and #1271): LISTED face-down characteristics (CR 708.2), and the face-change timestamp (CR 613.7f) · Accepted · S46

**Status:** Accepted · 2026-09-23 · S46 · [#1270](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1270),
[#1271](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1271), trackers
[#886](https://github.com/krakenhavoc/cmd_and_ctrl/issues/886) (face-down) and
[#881](https://github.com/krakenhavoc/cmd_and_ctrl/issues/881) (layers)
**Proof cards:** Cyber Conversion (caveat dropped), Yedora, Grave Gardener,
Cybership
**Changes:** ADR 0069's object model gains ONE per-object field beside the
per-kind one. The first amendment's A6 (#1270) and A2's "not done" (#1271) are
closed by this one.

### B1. Where the listed body lives: a sibling field, not a kind

CR 708.2 is written in two halves: a face-down object has "no characteristics
other than those listed by the ability or rules that allowed [it] to be face
down", and CR 708.2a is only the DEFAULT for an effect that lists nothing.
`FaceDownBody(kind)` built the default and nothing else.

#1270 set out three shapes. The one taken is the issue's (1) with (3)'s call
sites:

- **`Card.FaceDownListed *FaceDownListing`**, beside `FaceDownKind` — types,
  subtypes, power, toughness. Nothing else, because no printed card lists a
  name, a cost, a colour, a supertype or an ability, and every one of those is
  "none" on a CR 708.2 object.
- **The kind still answers every question about WHY** the object is face down:
  who may look (CR 708.5), whether it can come back up (CR 708.7), whether it
  has ward. The listing answers only "what is it", the one question a kind could
  not.
- **It is read in exactly one place**, `faceDownCharacteristic`, the layer-0
  baseline, which prefers the listing to `FaceDownBody(kind)`. So every layer
  above it, every predicate, targeting, combat and the wire got it with no
  change. The three cold-cache fast paths that bypass layer 0 (`HasCardType`,
  `HasSubtype`, and the CR 305.6 intrinsic mana read) got the same guard.

**Rejected: a kind per listed body** (#1270's option 2). Two bodies exist
today and a kind cannot carry a payload, so the third card that lists one would
need a third kind. Worse, the kind already means something (WHY), and a
"cyberman" kind would have to re-answer CR 708.7 and ward for itself.

**Rejected: the body as data on the kind at the call site only** (#1270's
option 3, alone). Listed characteristics are per OBJECT. Two Cybermen and a
Forest can sit on one battlefield, all `turned`, and the snapshot, the clone
and the wire all need to know which is which after the call returns.

**It REPLACES the default, and Yedora is the proof.** "It's a Forest land" is
not a 2/2 with a land type added. A face-down Forest is not a creature and has
no P/T. It taps for {G} because CR 305.6 gives the ability to anything with the
Forest subtype, and the engine already derives that ability from the effective
subtypes. The listing does not have to say it.

**Lifecycle.** `SetFaceDownListed(kind, listed)` is the one writer of the
triple (`FaceDown`, `FaceDownKind`, `FaceDownListed`). It keeps a listing only
on a CR 708.2 permanent state, because a face-down card in exile has no
characteristics at all (CR 406.3a). It also copies the value, so no caller can
alias a card's. `ClearFaceDown` drops it, so `MoveCard`'s CR 400.7 reset and
turning face up both take the listing away with the state it belonged to. The
field is never mutated through the pointer. Clone and snapshot deep-copy it
anyway, following `PrintedSelf`, and the drift table marks it carried.

**The doors.** `TurnFaceDownListedForEffect(source, listed, ids...)` is the
turn. `TurnFaceDownForEffect` is now that function with `nil`. The entry side
rides `ZoneEntryOptions.FaceDownListed` and `ReplacementEvent.FaceDownListed`
beside `FaceDown`, for `FaceDown`'s reason: the one entry finisher reads it,
whichever door the permanent came through, and a paused entry's resume still
has it. `ReturnFromGraveyardFaceDownForEffect(card, controller, tapped, listed)`
is the reanimation door with the face-down state on the event. It is Yedora's
door, and Missy's too once that card is written.

**CR 708.2b holds for the listing.** A listed turn aimed at a permanent that is
already face down does nothing, and "doesn't change any of its characteristics"
includes the listing. A Cyber Conversion aimed at a morph does not make it a
Cyberman.

### B2. The kind for an effect's face-down ENTRY is `turned`

Yedora and Cybership put a card onto the battlefield face down, and neither is
a keyword. Manifest (CR 701.40) is a keyword action, and CR 701.40b, which lets
a creature card be turned up for its mana cost, applies to manifest alone. A
manifested Yedora Forest would have walked straight into that arm. That is the
`PrintedIsCreature` read #1270 flagged.

`FaceDownTurned` already gives every answer an effect's face-down needs. Only
the card's own morph or disguise brings it back up (CR 702.37e, CR 708.7). It
has no ward. Its controller is the only viewer. So its meaning widens from "was
turned face down" to "**an effect, not a keyword, did this**", and where the
object was a moment earlier is no longer part of it. The wire string is
unchanged (`"turned"`), and so is the client's badge.

**Rejected: an eighth kind** (say, `put`). It would have the same arm in
`TurnFaceUpOffer`, the same ward answer and the same viewers row, so it would
be a synonym that every table has to learn twice.

### B3. Copiable values: CR 708.2's second sentence

"Any listed characteristics are the copiable values of that object's
characteristics." `CopiableValuesOf` read the card underneath for a face-down
permanent, so a Clone of a morph became a Willbender. That is the card the
Clone's controller may not look at (CR 708.5), and it made the Clone stronger
than printed. It now returns the body: the listed one, or CR 708.2a's nameless
2/2, with no name, no oracle ID (so no catalog entry), no cost and no colour.
`applyCopy`'s stash of a card's OWN values reads the raw printed fields through
`printedValuesOf`, because a card's own values are the card's whatever state it
is in.

### B4. CR 613.7f: a third timestamp field (#1271)

`EnteredBattlefieldAt` has two jobs, and CR 613.7f touches only one of them. It
is the CR 613.7d sort key for the permanent's own statics. It is also the
object's **identity pin**: every until-end-of-turn, control and earthbend
effect is keyed on `(InstanceID, EnteredBattlefieldAt)`, and summoning sickness
reads "entered this turn" beside it. Re-stamping it would have dropped every
pinned effect off a morph the moment it turned up, which contradicts CR 708.8's
"not a new object".

So the timestamp gets its own field, **`Card.FaceTurnedAt`**, stamped by both
directions: `turnFaceUpLocked` and `TurnFaceDownListedForEffect`, with one stamp
for Ixidron's whole batch. The first amendment's A2 said both or neither, and
both it is. The layer gather now sorts a permanent's statics by
`layerTimestamp()`, the **latest** of `EnteredBattlefieldAt`, `AttachedAt`
(CR 613.7e) and `FaceTurnedAt` (CR 613.7f). All three are the same rule, "the
permanent receives a new timestamp", and a later one supersedes an earlier one
of any kind. That replaces the old "AttachedAt when non-zero", which only had
two inputs. The field is zeroed on battlefield exit alongside the entry stamp,
and carried by clone and snapshot.

**What it takes to observe it:** two permanents with order-dependent effects.
The test is two layer-7b "set base P/T" lords. The morph entered first, so the
other lord's set wins (4/4). The morph is turned down and back up, and the bear
is now 1/1. Before this change it stayed 4/4.

### B5. Two holes the listing work found and closed

- **A face-down TOKEN kept its card-carried mana ability.** `CatalogKey`
  silences a face-down card, but a token carries `ManaAbilities` on the
  instance, so a Treasure hit by Cyber Conversion still tapped for mana.
  `ManaAbilitiesForCard` now declares nothing for a CR 708.2 object. The
  CR 305.6 intrinsic half is unaffected, which is how the listed Forest keeps
  its {G}.
- **`HasSubtype` answered `false` for every face-down permanent**, including
  one that a later layer-4 effect had given a subtype. It now reads the
  effective subtypes when warm and the body's when cold. Changeling on the card
  underneath still counts for nothing (CR 708.2a).

### B6. What this does NOT build

- **Missy and The Cyber-Controller.** Both are in #1270's table and neither is
  in the ranked roadmap. Missy's primitive is here
  (`ReturnFromGraveyardFaceDownForEffect` with a controller and `tapped`), but
  its second ability is a villainous choice. The Cyber-Controller needs a mill
  that reports what it milled. They are ordinary card work now.
- **A copy of a face-down SPELL** is CR 708.2's copiable values like any other
  copy. What the copy then does on resolution is not revisited here.
- **CR 701.40c** and hideaway: decision 10 still holds for both.

## Amendment (2026-10-07, #2570): manifest dread (CR 701.62a) · Accepted · S43

**Status:** Accepted · 2026-10-07 · S43 · [#2570](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2570),
tracker [#886](https://github.com/krakenhavoc/cmd_and_ctrl/issues/886),
home tracker [#2555](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2555)
**Proof cards:** the four Room doors that shipped empty in PR #2565 (Weight
Room, Ticket Booth, Slimy Aquarium, Experimental Lab) and twenty-two cards
printing manifest dread.
**Numbering:** an amendment to this ADR; no new number is taken.

Manifest (decisions 1 to 10 and `ManifestForEffect`) takes the TOP card of a
library. Manifest dread (CR 701.62a) does not: "look at the top two cards of
your library, put one of them onto the battlefield face down as a 2/2 creature
and the other into your graveyard." What was missing was the choice, and the
cards that read what the choice produced.

### C1. Composed from doors that already exist

`Game.ManifestDreadThenForEffect(player, source, then)` is a new file
(`game/manifest_dread.go`) and owns only the order the existing doors run in:

1. `LookAtTopOfLibraryForEffect(player, 2)`. "Look at", not reveal: only the
   looker becomes a knower, the rule scry, surveil and hideaway already follow.
2. A `choose_cards` prompt over the pair (floor 1, ceiling 1, zone library),
   chooser the looker. It is the prompt hideaway queues, so the wire already
   hides its options from every other seat and the enumerator and both bot
   policies already answer it. No new prompt kind, no new action, no new wire
   shape.
3. The chosen card enters through the CR 614 face-down entry manifest uses
   (`PutCardsFromLibraryOntoBattlefieldThenForEffect` with
   `FaceDownManifested`), so a question the entry asks pauses the action and
   the rest waits behind it.
4. The other card goes to its owner's graveyard through the shared exit route
   (`routeAllThenLocked`), so a commander is offered the command zone
   (CR 903.9). It is a plain zone move: it is not a mill (CR 701.17a counts
   cards off the top) and not a discard, so no mill or discard payoff sees it.

A library of one card asks nothing and manifests it. An empty library does
nothing: no event is emitted, and the continuation still runs with an empty
result so a "then put counters on that creature" can find it has none.

### C2. What the continuation is told

`ManifestDreadResult{Player, Manifested, Graveyarded}`: the permanent that
entered, and the card put into the graveyard "this way". Both are `uuid.Nil`
when they did not happen. Every card that says "that creature" or "a card you
put into your graveyard this way" reads these, and nothing reads the library
top. The continuation runs once, after the last leg settles, and captures only
scalars: it is held in a `chooseCardsFrame`, the same census category hideaway's
is, so an open manifest-dread prompt is as much a restore point as an open
hideaway prompt and no more.

The manifested card enters FIRST and the other card goes to the graveyard from
the entry's continuation. The rule text puts them in one sentence and no card
reads the graveyard between the two.

### C3. The event and the log

`EventManifestDread` (Actor, Source, `CardID` = the permanent that entered,
`Target` = the card put into the graveyard) is emitted once, after both moves.
It is its own kind rather than a flag on a manifest, for the reason
`EventSurveil` is not a flag on `EventScry`: a card that cares about manifest
dread must not fire on a Cloak or a plain manifest. The public log gets one
line, `manifest_dread` ("Alice manifested dread"), that carries no card: the
manifested card is hidden (CR 708.5) and the graveyard card already has its
zone line. The client's `LogKind` union gains the kind and nothing else.

### C4. The readers

- `effects.ManifestDread{Player, Then}` and `effects.ManifestDreadTimes{N, Then}`
  ("twice", "X times"; each repetition is its own look, its own prompt and its
  own entry window, and `Then` is told every result).
- `PutCountersOnManifested(...)`, `AttachSourceToManifested(ctx)` and the
  `WheneverYouManifestDread` predicate. The counter and attach continuations
  skip a creature that left before they ran, and attach is skipped quietly when
  the Equipment is no longer the permanent that asked (CR 400.7, CR 701.3b).
- `ManifestDreadWhenItDiesThisTurn`, for Turn Inside Out's "when it dies this
  turn": an event-delayed trigger pinned to the object, the shape Whippoorwill's
  exile uses. Its body is keyed `dies/manifest-dread` in `delayed_bodies.go` and
  the ledger, so a table with one waiting is still a restore point.
- Curator Beastie's "colorless creatures you control enter with two additional
  counters" reads the entering permanent AS IT WILL ENTER (CR 614.12): a card
  put onto the battlefield face down is a colorless creature whatever it is in
  the library, so the replacement reads `ReplacementEvent.FaceDown`, and the
  manifest Beastie itself makes arrives as a 4/4.

### C5. Bot and enumerator coverage

No new prompt, so the enumerator needed nothing. `internal/aiseat/manifest_dread_test.go`
drives both the heuristic and the random policy through the real engine and
fails on the two wedges (#544): a seat owed a question that is offered nothing,
and an enumerated answer the engine refuses. The heuristic scores a library
`choose_cards` by what it takes (`valueTakenFromLibrary`), so it manifests the
card it values more and the other one goes to the graveyard. That is a
reasonable preference rather than an optimal one: it does not yet weigh that
only a creature card can be turned face up, and a policy that did could do
better. It is declared here, not hidden.

### C6. Cards, and what stayed out

Shipped complete: Manifest Dread, Unsettling Twins, Innocuous Rat, Bashful
Beastie, Unnerving Grasp, Under the Skin, Paranormal Analyst, They Came from the
Pipes, Growing Dread, Threats Around Every Corner, Killer's Mask, Cursed
Windbreaker, Conductive Machete, Dissection Tools, Break Down the Door, Twist
Reality, Valgavoth's Onslaught, Turn Inside Out, Glitch Interpreter, Curator
Beastie, Disturbing Mirth, Stay Hidden, Stay Silent. Four Rooms lose the caveat
the manifest dread door carried: Moldering Gym // Weight Room, Ticket Booth //
Tunnel of Hate and Underwater Tunnel // Slimy Aquarium are complete;
Experimental Lab // Staff Room keeps only the Staff Room caveat.

Left on the roadmap row, each with its blocker:

- **Hauntwoods Shrieker, Zimone, Mystery Unraveler** (and Staff Room's other
  option): "turn it face up" as an EFFECT, with no cost. Face up is a special
  action with a cost (`turnFaceUpLocked` runs inside `PerformSpecialAction`);
  there is no `TurnFaceUpForEffect`. Zimone also needs "the first time this
  ability has resolved this turn", a per-object per-turn tally.
- **Defiant Survivor**: survival triggers at the beginning of the second main
  phase, and there is no event for it (`EventBeginPrecombatMain` is the only
  main-phase event).
- **Abhorrent Oculus**: "as an additional cost to cast this spell, exile six
  cards from your graveyard". `AdditionalCost` has no exile-from-graveyard
  component.

Nothing ships stronger than printed (#259).

## Amendment (2026-10-07, #2590): turning a permanent face up as an effect (CR 708.8, CR 701.40b) · Accepted · S43

**Status:** Accepted · 2026-10-07 · S43 · [#2590](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2590),
tracker [#886](https://github.com/krakenhavoc/cmd_and_ctrl/issues/886),
home tracker [#2555](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2555)
**Proof cards:** Hauntwoods Shrieker, Zimone, Mystery Unraveler, and the Staff
Room door of Experimental Lab // Staff Room (shipped with a caveat in PR #2585).
**Numbering:** an amendment to this ADR; no new number is taken. This follows
the manifest dread amendment above, which left these three cards on the
roadmap row as its first named gap.

Decision 6 made turning face up a special action with a price. Three printed
cards turn a permanent face up as an INSTRUCTION, paying nothing: "you may turn
it face up", "turn a permanent you control face up", "turn that creature face
up". The special action cannot carry them, because its price and its timing are
the action's and not the effect's. What was missing was a door for the second.

### D1. `TurnFaceUpForEffect(source, actor, id)` and one shared tail

```go
func (g *Game) TurnFaceUpForEffect(source, actor, id uuid.UUID) bool
func CanTurnFaceUpForEffect(c Card) bool
```

`turnFaceUpLocked` is split in two. The part that validates the SPECIAL ACTION
(the card is face down, the actor controls it) stays in it. The part that
turns the permanent over moved, unchanged, into `finishTurnFaceUpLocked`, and
both doors call it: the same `ClearFaceDown`, the same Room-door
materialisation, the same CR 613.7f timestamp, the same cache nil-out, the same
every-seat-is-a-knower marking, the same `EventTurnedFaceUp` emitted AFTER the
state is cleared (decision 7's order). That sharing is the contract. A "when
this is turned face up" trigger and Growing Dread's "whenever you turn a
permanent face up" cannot tell the two doors apart, which is what the rules say:
CR 708.8 does not care how it was turned.

It returns whether the permanent turned. A permanent that is gone, face up, or
refused is left exactly as it was and nothing is emitted.

### D2. What an effect still cannot do

An effect does not pay a morph cost or a mana cost, so `TurnFaceUpOffer` (decision 5) is
not the question: it prices the ACTION. Asking it would refuse a permanent an
Ixidron left with no way back up (CR 708.7 withholds the special action, not
the effect) and would demand a cost the effect never pays. The question is
`CanTurnFaceUpForEffect`:

| kind | an effect may turn it face up |
|---|---|
| `morphed`, `disguised`, `turned` | yes |
| `manifested`, `cloaked` | only if the card is a creature card (CR 701.40b, CR 701.58b) |

"Creature card" is `PrintedIsCreature`, for decision 5's reason: the question is
what the CARD says, and the face-down projection calls a manifested Island a
2/2 creature. A manifested noncreature card stays face down and the effect does
nothing to it, with no event, so abilities that trigger on a turn face up do not
see it.

### D3. No cost means no megamorph counter

CR 702.37b owes the +1/+1 counter only when the megamorph COST was paid to turn
the permanent up. The counter therefore rides the cost, and `finishTurnFaceUpLocked`
takes it as a parameter: the special action passes `sa.FaceUpCounter`, the effect
door passes false. A megamorph creature turned up by Hauntwoods Shrieker is a
creature without its counter. This is stated rather than discovered later
because "the same path" is easy to read as "the same counter".

### D4. Actor, source, and the log

`EventTurnedFaceUp.Actor` is the player whose effect it is, NOT the permanent's
controller. The Shrieker reaches an opponent's creature card; "whenever you turn
a permanent face up" is asked of the player who turned it. `Source` is the
object doing it. For the special action it remains the permanent itself, which
is the case the existing `Self` predicate and the harvester already read, and
which the event's doc comment promised.

The log's silence row for this kind said that if a second emitter ever appeared
the kind would need an arm. It has. `LogTurnFaceUp` ("turn_face_up") is written
only when `Source` is not the permanent, so the special action keeps its
`LogSpecialAction` line and does not gain a second one. The entry names the
permanent, which is public again by the time it is written, and carries the doer
in `target`, the mirror of `turn_face_down`. The client's `LogKind` union and
tone map gain the kind and nothing else.

### D5. Hauntwoods Shrieker: a reveal that leaves the permanent face down

"Reveal target face-down permanent" goes through `RevealForEffect`, the same
door CR 708.9 uses: every seat becomes a knower of the card and one reveal frame
goes out. The permanent stays face down unless the controller of the Shrieker
says yes to the following "you may turn it face up", which is asked only if the
card underneath is a creature card the engine will let an effect turn over. The
knowledge stays: a table that was shown a card does not forget it, and decision
A3 of the 2026-09-23 amendment forgot only when the permanent was turned face
down, where the controller was the only person who had ever looked. The target
clause is a new `effects.FaceDown()` predicate beside `WithMorphAbility`, which
reads the object's state and tells the targeting player nothing about the card.

### D6. Zimone: the tally already existed

"If this is the first time this ability has resolved this turn" is
`Game.ResolvedThisTurn(source, label)`, the per-object count Sephiroth reads. The
count includes the resolution in progress, so the first resolution reads 1. It
is per OBJECT (CR 400.7): a Zimone that left and came back starts again. No new
tally was built: the manifest dread amendment's note that Zimone "also needs"
one is satisfied by a reading the engine already had. The "otherwise" branch is a
`ChoosePermanents` over the controller's own face-down permanents that
`CanTurnFaceUpForEffect` accepts, floor zero and ceiling one. With nothing to
turn over it asks nothing.

### D7. Staff Room: a choice asked only when there is one

"Turn that creature face up or put a +1/+1 counter on it" is a choice of the
resolving controller. A face-up creature has only the counter, so it gets it
without a prompt, exactly as before. A face-down creature the engine will let an
effect turn over gets a `MayChoice` with both options spelled out. A manifested
noncreature card is not offered the turn and takes the counter.

### D8. Bot and enumerator coverage

No new prompt kind. `MayChoice` is a confirm, `ChoosePermanents` over one's own
board is `own_permanents`, and manifest dread's pick is `choose_cards`; the
enumerator and both bot policies already answer all three. The bounded AI-seat
run covers them with the Shrieker, Zimone and Staff Room cards in the catalog.

### D9. What this does NOT build

- **CR 701.40c** (a manifested card that also has morph may be turned up for
  either cost) remains out of scope, for decision 5's reason.
- **The special action is unchanged.** Its price, timing and offer are exactly
  what they were; only its tail moved, with no change in behaviour.
- **Morph is not made turnable "for the Shrieker's cost".** The Shrieker's
  turn is free, as printed.

Nothing ships stronger than printed (#259): the one place an effect could have
exceeded the rules, a manifested noncreature card, is refused in D2, and the
megamorph counter an effect does not owe is not given in D3.
