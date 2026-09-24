# ADR 0025 — Alternative costs to cast

**Status:** accepted (S22)
**Extends:** [ADR 0021](0021-additional-costs.md) (the *additional*
cost this is deliberately not), [ADR 0019](0019-structured-targeting.md)
(the clause a cost is now allowed to rewrite),
[ADR 0020](0020-activated-abilities.md) (the cost-as-a-struct model).

## Context

The [Aang triage](../decklists/aang-is-so-flashy.md) named this the
keystone gap, and it was right for a reason that isn't obvious from a
card count: **three cards already in the catalog were shipping with
their headline mode missing.**

- **Cyclonic Rift** was a two-mana single bounce. Nobody plays it for
  that. The reason it is a Commander staple is the {6}{U} overload —
  a one-sided board wipe at instant speed.
- **Vandalblast** was a one-mana Naturalize.
- **Slithermuse** could only be hard-cast for {2}{U}{U}, which is not
  how the deck uses it.

Plus a group of blocked cards — evoke, cleave, spree, foretell, plot,
warp — that all say the same thing in different words: *pay this
instead of the mana cost*.

[ADR 0021](0021-additional-costs.md)'s `AdditionalCost` is the other
kind: paid **alongside** the mana cost. Nothing in the engine could
express a cost that **replaces** it, and the two are not variations on
a theme — an additional cost adds to a number, an alternative cost
substitutes for it, and one of them can also rewrite the card's text.

## Decisions

### 1. A slice of structs on the Spec, keyed by a wire-stable name

```go
type AlternativeCost struct {
    Key              string       // "overload", "evoke", "cleave"
    Label            string       // "Overload {4}{R}"
    ManaCost         string       // paid INSTEAD of the printed cost
    Targets          *TargetSpec  // cleave: swap the clause
    ClearsTargets    bool         // overload: delete the clause
    SacrificeOnEntry bool         // evoke: the entry trigger
}
```

A **slice**, not a single struct: spree and the modal-cost cards offer
more than one, and widening later would churn every signature.

**Keyed by name**, because unlike every other announce-time choice this
one has to survive the round trip. `Modes` are indexes into a list the
client can see; an alternative cost has to be claimable by the client
(`cast_spell`'s `alternative_cost`), recorded on the stack
(`StackItem.AltCost`), and read back at resolution
(`ctx.PaidAltCost("overload")`). A stable string is the only thing that
survives all three without a lookup table. `Register` panics on a blank
or duplicated key, so the failure is at boot rather than at a
mysteriously-rejected cast mid-game.

### 2. The cost is not the whole clause — three rewrites ride with it

This is the decision that matters, and the one a "just swap the mana
cost" implementation gets wrong.

Overload does not merely cost {6}{U}. It says *change "target" in its
text to "each"*. If the price swaps but the `TargetSpec` doesn't, the
announce gate still demands a target the spell no longer has, and worse,
the CR 608.2b re-check fizzles the spell the moment that target leaves —
so an overloaded Rift would be *counterable by removing one creature*.
`ClearsTargets` deletes the clause, and `TargetSpecUnderAlternativeCost`
is applied at announce **and** at the re-check so both judge the spell
under the text it was actually cast with.

Cleave is the same machinery pointed the other way: it *swaps in* a
wider clause rather than deleting one. Wash Away goes from "counter
target spell that wasn't cast from its owner's hand" to "counter target
spell", and both clauses are real — the printed one is enforced, not
waved through.

Evoke's rewrite is not a targeting one: *it's sacrificed when it
enters*. Modelled as a triggered ability (CR 702.74a) queued by the
engine as the permanent lands, **not** as a sacrifice folded into
resolution. The difference is the entire reason to evoke a Slithermuse:
the creature genuinely enters, the sacrifice uses the stack so opponents
get a window, and the leaves-the-battlefield trigger then goes on the
stack and draws. Folding it into resolution produces the same hand size
and the wrong stack.

The constructors (`Overload`, `Evoke`, `Cleave`) exist so a card file
cannot get half of this. A hand-written
`game.AlternativeCost{ManaCost: "{4}{R}"}` compiles, casts for four, and
still demands a target — a strictly worse Vandalblast that looks right.

### 3. Alternative replaces the mana cost and *only* the mana cost

Additional costs survive the swap: CR 601.2f is evaluated independently
of the cost chosen at 601.2b, so a card charging both charges both.
Commander tax likewise layers on top (CR 903.8 taxes whatever cost is
being paid), which is why `effectiveCostLocked` applies the tax *after*
the substitution rather than before.

The engine-side test that matters is the arithmetic one: a pool of
exactly {4}{R} must pay an overload and end empty. Six spent would mean
the printed {R} was charged on top — a bug that passes every test which
only checks the cast succeeded.

### 4. Claiming a cost the card doesn't offer is a rejection

Not a fall-back to the printed cost. Silently charging full price for a
cast the player meant to overload is the worst available failure: the
mana is gone, the board isn't wiped, and nothing says why. Same posture
as ADR 0021's "a card with no additional cost that arrives with
`discard_ids` is rejected rather than ignored". An overload arriving
*with* targets is rejected for the same reason — the client is confused
about which cost it is paying.

### 5. The picker opens first, and here that IS a rules-driven order

ADR 0021 was explicit that `DiscardCostModal` opening before targeting
is a UI choice, since the whole cast rides one message. The alternative
cost is different: the answer decides what the targeting prompt is
allowed to offer, so asking afterwards would mean re-opening it. The
chain is alternative cost → discard → sacrifice → X → modes →
targeting.

The picker lists **"Its mana cost"** as an ordinary first option and
defaults to it. An additional cost is a demand; an alternative cost is
an offer, and declining has to be as cheap as accepting.

### 6. The client stays ignorant of what the keywords mean

`AlternativeCostView` carries the target clause **the spell has under
that cost**, already resolved server-side. So the client's rule is: an
offer with no `target_mode` casts immediately, an offer with one enters
targeting on the legal set it carries. Nothing in `client/src` knows
what "overload" does, and the next keyword needs no client change.

This also fixes a castability trap: `canCastFromHand` greyed out a
Cyclonic Rift with nothing to bounce, because the *printed* clause had
an empty legal set. It now passes if **any** cost option's clause is
satisfiable.

### 7. Paid off the ADR 0021 debt rather than adding to it

ADR 0021 flagged that `xValue` / `discardIDs` / `sacrificeIDs` had
become three parallel positional parameters threaded through `begin` /
`beginForMode` / `continueCast`, and that a fourth should trigger a
bundle. This was the fourth. They are now one `CastChoices` object, and
`applyCastChoices` is the single place that knows the wire names.

## Consequences

- Cyclonic Rift, Vandalblast and Slithermuse ship with the mode people
  actually play them for. **Wash Away** is new.
- `StackItem` gains two fields, both of which are facts that stop being
  readable a moment after announce: `AltCost` and `CastFromZone`. The
  second is the one-field addition the triage asked for — CR 601.2a
  moves the card to the stack and nothing on it remembers where it came
  from, so Wash Away's clause needs it stamped. (The sibling request,
  a source zone on `EventCast` for Appa's "whenever you cast a spell
  from exile", is still open; this is the StackItem half, not the event
  half.)
- The stack overlay shows the cost paid, because it changes what the
  spell does — a responder needs to know whether the Rift on the stack
  is a wipe or a bounce.
- Airbend's piece 1 is now built. Pieces 2 and 3 (a permission with no
  expiry, exile-a-target-permanent-with-permission) are unchanged.

## Alternatives considered

**A boolean `IsAlternative` on `AdditionalCost`.** Rejected: the two
share a name and nothing else. An additional cost is mandatory,
unnamed, and never touches the card's text; an alternative cost is
optional, has to be named on the wire, and rewrites the clause. One
struct doing both would be a discriminated union pretending to be a
record.

**A parsed cost string ("overload:{4}{R}:each").** Same rejection ADR
0020 and ADR 0021 made: the shapes are few, and a cost mini-language
has to be maintained against a handful of cards.

**Overload as a second mode.** Tempting — `ModeSpec` already carries
per-option target clauses, and "choose one: target, or each" reads
close. Rejected because modes don't change the price. A player could
choose the overload mode and pay {R}, which is not a simplification but
a different and much stronger card. The Vandalblast comment in the
pre-S22 catalog said exactly this, and it was right.

**Evoke as a sacrifice inside `OnResolve`.** Rejected for ADR 0021's
reason one zone over: it produces the right hand size and the wrong
stack. No response window, and the leave-trigger's ordering relative to
the sacrifice inverts.

## Known limitations

- **"Cast this turn" is not checked** on Wash Away. A spell on the
  stack was cast this turn in every situation the engine can currently
  produce, so the clause is satisfied by construction rather than by a
  timestamp. A card that parks a spell on the stack across turns would
  need a real cast-turn stamp.
- **Free alternative costs work but are untested by a real card.**
  `ManaCost: ""` parses to the zero cost, which is exactly right for
  "without paying its mana cost" — but nothing in the catalog exercises
  it yet.
- **Alternative costs that are not mana** (Force of Will's "exile a
  blue card and pay 1 life", Daze's "return an Island") have no shape
  here. `AlternativeCost` is a struct so they can arrive as fields, the
  same way `AdditionalCost` grew its `Sacrifice` clause.
- **Spree and the modal-cost cards** need a cost *per chosen mode*,
  which is a `ModeSpec` × `AlternativeCost` cross product this does not
  attempt. The slice shape is the half of it that exists.
