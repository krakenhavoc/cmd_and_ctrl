# ADR 0068 — A stack item records the mana spent to cast it: converge, sunburst, adamant, and "if no mana was spent"

**Status:** Accepted · 2026-09-18 · S44 — Mana and cost components
**Issue:** [#761](https://github.com/krakenhavoc/cmd_and_ctrl/issues/761)
**Numbering:** re-checked immediately before pushing on 2026-09-18, after a
fresh fetch: every remote branch was enumerated (`git ls-remote --heads
origin`) and every ADR file reachable from any ref listed
(`git log --all --name-only -- docs/decisions/`). 0061–0067, 0069 and 0070
are all taken by this sprint's other agents — 0070 twice over, by
`0070-untap-step-choices.md` on another branch. **0068 is the one gap**,
absent from every ref and from a repo-wide filename search, so this ADR
takes it. 0005, 0024, 0029 and 0030 stay permanently unused per
AGENTS.md §4.
**Related:** [ADR 0040](0040-mana-pipeline.md) (the pool, the spend context
and the solver this extends), [ADR 0048](0048-cost-modification.md)
(`CastCost{Printed, Paid}` — what a spell was *charged*, as against what was
*spent*), [ADR 0020](0020-activated-abilities.md) (the #789 addendum, which
puts the same record on an activation), [ADR 0041](0041-game-persistence.md)
(the `carried` / `rebuilt` / `dropped` classification the new field answers
to)

## Context

The engine forgot which mana paid for a spell the instant the cost was paid.

`ManaToken{Color, Source, Restrictions}` (`game/mana.go:29`) knows perfectly
well what it is and where it came from. `SpendManaFor` returned a bool.
`attemptSpend` returned the pool that was *left*, not the tokens it had
used. `EventManaSpent` carried an Actor and a Source and no colours and no
amount. `StackItem` carried `XValue`, `CastFromZone`, `AltCost` and
`IsCopy`, and nothing about mana at all.

So five printed mechanics had nothing to read:

| Mechanic | Rule | Reads |
|---|---|---|
| Converge | CR 207.2c ability word | the number of COLOURS spent |
| Sunburst | CR 702.44a | the same number, at entry |
| Adamant | CR 207.2c ability word | how many of ONE colour were spent |
| "if no mana was spent to cast it" | — | whether the record is empty |
| "when that mana is spent" riders | — | which SOURCE made the mana |

The audit of batch issues #294–#467 found 20 cards where this was the only
core blocker and 28 where it was any blocker, plus five already-shipped
cards carrying caveats about it (Scaled Nurturer, Path of Ancestry, Lux
Artillery, Satoru, and the inert "can't be countered" halves of Cavern of
Souls and Delighted Halfling).

There is a second, quieter problem, and it is the one that decides the
shape of the answer. `applyCastCostLocked` spends from the pool **only
under Strict** (`mutations.go:1129`). A permissive cast or a `ForceCast`
emits `EventCostWarning` and leaves the pool alone — and the human default
is `strictMana: false` (`client/src/lib/settings.ts:267`). So at most tables
the engine has never seen the mana that paid for anything. "No record" and
"nothing was spent" are the same absence, and a reader that treated them
alike would counter half the spells cast at a permissive table with Vexing
Bauble.

## Decisions

### 1. One record, and it is DATA on the item

`StackItem.Paid` is a `game.PaidCost` (`game/paid_cost.go`):

```go
type PaidCost struct {
    Mana            []ManaToken
    OnPaper         bool
    CountersRemoved int
    CountersAdded   int
    LifePaid        int
}
```

It is stamped at announce, next to `XValue`, and for exactly the same
reason: by the time the item resolves, the mana is gone from the pool, the
Treasure that made it may be in a graveyard, and the counters a cost removed
are off the board. Nothing downstream could recompute any of it.

It is shared with [ADR 0020's #789 addendum](0020-activated-abilities.md),
which lands in the same PR and needs the counter half of the same record for
"Add {C} for each storage counter removed this way". **One record, designed
once**, rather than a mana record and a counter record that would have to
agree about what an announcement is. A mana ability has no stack item (CR
605.3b), so there the record lives for the length of the activation and is
handed to the one callback that needs it.

Rejected: a `ManaSpent []ManaToken` field of its own. It would have been
the second field on the item meaning "what this announcement cost", and the
third would have arrived within a sprint.

### 2. `attemptSpend` returns what it spent

`ManaPool.attemptSpend` now returns `(remaining, spent, ok)`;
`SpendManaFor` returns `([]ManaToken, bool)`. Every payment path records
what it got back: the cast gate, an activated ability's mana component, a
mana ability's mana component, and `payCostLocked`.

`CanPayFor` is unchanged — it discards both slices and keeps its bool, so
the dry-run half of the two-step pattern reads exactly as it did.

### 3. `OnPaper` is how "unknown" is said out loud

A permissive or `ForceCast` cast records `PaidCost{OnPaper: true}`. That is
a positive statement — *the engine waived this charge* — rather than an
absence, and it is what makes every reader's answer a decision rather than
an accident.

**The rule, and it is one line: unknown is never the stronger answer.**

| Reader | Known, empty | Unknown (`OnPaper`) |
|---|---|---|
| `NoManaSpent()` | true | **false** — the punisher does not fire |
| `ColorsSpent()` | empty | **empty** — converge draws nothing |
| `ManaSpentOfColor()` | 0 | **0** — adamant does not turn on |

Both directions land on "weaker than printed", which is the #259 rule, and
they land there for different reasons — a Vexing Bauble that countered
everything at a permissive table would be actively wrong, while a converge
spell that draws nothing is merely disappointing. Stating it as one
invariant rather than per card is what stops the next reader guessing.

`Known()` exists for a card that wants to say "unknown" out loud instead.
Nothing in the catalog uses it yet; the protocol view does.

### 4. A copy of a spell records a real zero

CR 707.10 and the Dawnglow Infusion ruling: mana is not an object, so
nothing was spent to cast a copy. `CopySpellForEffect` leaves `Paid` at its
zero value, which is `NoManaSpent() == true` — and that is right, not a
gap: a copied Vexing Bauble trigger really should counter the copy.

### 5. One solver, one strategy switch

The generic half of a payment has always been paid colourless-first, then
W U B R G (`mana.go:169`), which is the correct instinct — keep the coloured
mana for the next spell — and exactly backwards for converge and sunburst.

`attemptSpend` gains a `ManaSpendStrategy`:

- `SpendPreserveColors` (the default, unchanged behaviour), and
- `SpendDistinctColors`, which pays the generic half with a colour it has
  not spent yet whenever it can, before falling back to the default order.

The switch is thrown by the SPELL, not by the player: `Spec.WantsDistinctColors`
declares a converge or sunburst reader, reaches the engine through
`CatalogWantsDistinctColors`, and `applyCastCostLocked` picks the strategy.

**Only the generic half.** The coloured-requirement pass is untouched, so
the strategy can never change whether a cost is payable — it only chooses
between tokens the solver was already free to take. That is a deliberate
bound: a solver whose colour preferences could make a payable cost
unpayable is a solver nobody can reason about, and no printed converge or
sunburst card has a coloured pip the choice would reach.

**Declared limit.** The AUTO-TAPPER does not spread colours. It plans
sources against the cost, and the pool it produces is usually spent whole,
so converge works in the ordinary case; a player who floats more mana than
the spell needs and then auto-taps may converge for less than the board
allowed. Spreading the PLAN is a planner change with its own search-space
cost, and it is out of scope here.

### 6. `EventManaSpent` carries the colours and the amount

`Event` gains `Colors []string`, and the spend events fill it along with
`Amount`. The log line stops being "mana was spent" and becomes "{U}{B}{R}
spent", which is what a player watching a converge spell resolve actually
wants to see, and what a "when that mana is spent" rider will read when it
lands.

### 7. The read surfaces are on `effects.Context`

Next to `PaidAltCost`, and in its idiom — the choice was made at announce,
the resolution reads it back:

```go
ctx.ColorsSpent()            // converge's X, sunburst's counters
ctx.ColorsSpentCount()
ctx.ManaSpentOfColor("R")    // adamant
ctx.ManaSpent()              // the tokens, for a source-aware rider
ctx.NoManaSpent()            // Vexing Bauble, Satoru
ctx.ManaSpentKnown()
ctx.Paid()                   // the record itself
```

A trigger that has to reach a spell's record from an event (Vexing Bauble
watches `EventCast`) goes through `game.StackItemPaidForEffect(id)`: for a
spell the stack item's ID *is* the card's instance ID, so `EventCast.CardID`
is already the key.

### 8. Sunburst enters the way Hangarback Walker's X does

Sunburst is "enters with a +1/+1 counter for each colour of mana spent"
(CR 702.44a) — an entry replacement. The entry pipeline cannot see the
resolving spell's record without exporting the stack item to the catalog,
and there is already a precedent for the alternative: Hangarback Walker and
Goldvein Hydra put their X counters on **as the spell resolves**, a beat
before the card moves to the battlefield, and declare the difference as a
caveat.

Sunburst takes the same road, with the same caveat sentence, through one
shared `SunburstCounters()` helper rather than a per-card closure. The one
observable difference is that a "whenever you put counters on a permanent"
payoff does not see them — weaker, never stronger.

Reconsidering this means giving the entry pipeline a narrow read of the
record (`ReplacementEvent` already carries the resolving `stackItem`
unexported, for auras and evoke); that is a follow-up, not this ADR.

### 9. Undo, clone and persistence

`Paid` is **carried**. `cloneStackItem` deep-copies it — the token slice and
each token's `Restrictions` — because an undo snapshot that aliased the live
backing array would let a restore mutate the game it came from.
`stackItemSnapshot` mirrors it, and `snapshot_drift_test.go` classifies it,
with the reason: a restore that lost the record would resolve Painful Truths
for zero cards.

## Consequences

- Every payment path now knows what it spent. The five mechanics above have
  something to read, and so will the spend riders in the follow-up PR.
- "No mana was spent" is unambiguous at a permissive table, in the
  weaker-than-printed direction, by construction rather than by luck.
- `SpendManaFor`'s signature changed. Four call sites in the engine; no
  card file calls it.
- The solver has two orders instead of one, chosen by a declaration on the
  spell rather than by a prompt. CR 601.2h's "the player chooses which mana
  to spend" is therefore still not modelled as a player decision — that is
  the honest limit, and it is the same one the auto-tapper has always had.

## Out of scope

- **Spend riders.** "When that mana is spent" (Pyromancer's Goggles, Scaled
  Nurturer), entry riders (Biophagus, Opal Palace), haste grants (Hall of
  the Bandit Lord) and per-spell "can't be countered" (Cavern of Souls,
  Delighted Halfling) all need a closed tag set on the token and a fire
  point after payment. That is PR 2 in the issue's own split; the caveats
  on those cards stay.
- **A source-type snapshot on the token.** A sacrificed Treasure is gone by
  resolution (Kalain, Rain of Riches). `ManaToken.Source` is recorded, which
  is enough for every reader shipped here; snapshotting the source's types
  belongs with the riders.
- **A payment prompt.** CR 601.2h lets the player choose the tokens. The
  strategy switch is the engine choosing well instead, and a prompt would
  also have to reach the bot move list.
- **Sunburst granted by another permanent** (Lux Artillery). The keyword
  helper is per-card; granting it is a static ability over a cast, which is
  its own seam.
- **The auto-tapper spreading colours.** §5.

## Alternatives considered

1. **Record the COST rather than the payment** — reuse ADR 0048's
   `CastCost{Printed, Paid}`. Rejected: that structure records what the
   spell was *charged* after modification, in symbols. Converge counts
   tokens, and a token's colour is not derivable from a cost symbol once
   hybrid and Phyrexian are in play.
2. **Leave the pool alone and reconstruct the payment from the event log.**
   Rejected for the reason #596 gives about sacrificed Food: the log names
   sources, the sources move, and the reconstruction quietly stops being
   true for the commonest case the card was printed for.
3. **Treat an unrecorded payment as five colours** (the generous reading).
   Rejected outright — stronger than printed, at the default setting, on
   every table.
4. **A per-card decision on the unknown direction**, as the issue's
   checklist offered. Rejected in favour of §3's single invariant: five
   cards choosing individually is five chances to choose wrong, and the two
   directions happen to agree.

## Amendment (2026-09-22, [#1212](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1212)): the source snapshot, and the entry-side read

This ADR's **Out of scope** said:

> **A source-type snapshot on the token.** A sacrificed Treasure is gone by
> resolution (Kalain, Rain of Riches). `ManaToken.Source` is recorded, which is
> enough for every reader shipped here; snapshotting the source's types belongs
> with the riders.

It does not belong with the riders. A rider needs a FIRE POINT after a payment,
which is still open; a card that asks "*if mana from a Treasure was spent to cast
it*" needs only the fact, and there are more of those than of every mechanic in
this ADR's table put together. A scan of the Scryfall dump finds **107 cards**
whose oracle text contains "was spent"; **23** of them read a SOURCE rather than
a colour or a count, eighteen of those spell it *Treasure*.

Two other things in the original are corrected below.

### A1. The kinds are snapshotted at MINT, not looked up at read

`ManaToken` gains `SourceKinds ManaSourceKinds` — a six-bit set (snow, Treasure,
creature, land, artifact, enchantment) read off the producing permanent's
EFFECTIVE characteristics (`mana_source.go`).

It has to be taken at mint, and that is not a preference. A Treasure's mana
ability **sacrifices the Treasure as part of its own cost**: in
`ActivateManaAbility` the sacrifice is paid, `card` is set to nil, and only then
is the mana minted. By the time any reader asks, the permanent is in a graveyard
— and a Treasure TOKEN has ceased to exist entirely (CR 111.7), so even the
graveyard has nothing. The one source that eighteen printed cards ask about is
the one source a lookup through `ManaToken.Source` can never answer.

The snapshot therefore rides **`PendingChoice.ManaSourceKinds`** as well,
because a Treasure's five-colour slot queues a pick that is answered after the
Treasure is gone. That is the same argument `ManaRestrictions` already won on
this struct, one card family sharper.

ONE decider, four call sites, mirroring `restrictionsFor`: `manaSourceKindsOf`
is the only thing in the engine that classifies a mana source, and
`manaSourceKindsLocked` is it for a caller holding an ID.

Rejected: a copy of the source `Card` on the token. It is a second object for
the clone, the snapshot and the undo stack to carry, and it goes stale in a
different way — the question is never "what is that permanent now", it is "what
was it when it made this mana". A bitset answers exactly that in two bytes.

### A2. `ManaSpent` is one vocabulary with two homes

§7's read surfaces were six accessors on `PaidCost`, which was right while every
reader was a spell reading its own stack item. The largest un-shipped family here
is an **enters trigger**, and by the time one resolves the item is gone:

	"When this creature enters, if mana from a Treasure was spent to
	 cast it, you draw a card and you lose 1 life."   Hired Hexblade
	"…if {R} was spent to cast it, it gains haste…"   Gruul Scrapper
	"…sacrifice it unless {U} was spent to cast it."  Azorius Herald

Two homes for one fact is the moment a vocabulary either gets written down or
gets duplicated. So the questions move onto a **view**, `game.ManaSpent`
(`mana_spent.go`): `Known`, `Total`, `Count(symbol)`, `Colors`, `ColorCount`,
`None`, `CountFrom(kinds)`, `From(kinds)`, `FromTreasure`, `Snow`, `Tokens`.
Unexported fields, no JSON, built on demand. `PaidCost.Spent()` and
`CastProvenance.Spent()` both hand it out, and §3's rule — unknown is never the
stronger answer — is enforced once, inside it, instead of six times.

`PaidCost`'s existing accessors are unchanged in name and behaviour and now
delegate to the view. The one rename: `PaidCost.ManaSpent()` (the token slice) is
`PaidCost.ManaTokens()`, because `ManaSpent` is now the name of the answer.

### A3. `Card.Provenance` carries the mana (CR 400.7d)

`CastProvenance` gains `Mana []ManaToken` and `ManaOnPaper bool`, stamped in the
same `stampCastProvenanceLocked`, at the same one entry finisher, three lines
from the optional costs #719 folded in for the identical reason. `cast_provenance.go`'s
header said the mana was NOT carried "because nothing reads it after entry"; that
paragraph is struck, and the field it predicted ("a sibling of AltCost on this
struct rather than a second record") is what shipped.

It is the same tokens rather than a summary, so the view above works unchanged on
either home. Cleared on CR 400.7 with the rest of the record, deep-copied by
`Clone`, classified `carried`.

`ManaOnPaper` comes too. Without it a waived payment and a genuinely free cast
are the same empty slice on the permanent, and §3's distinction — the whole
reason `OnPaper` exists — would survive the stack and die at the battlefield.

### A4. §8 was overtaken: sunburst rides the entry pipeline now

§8 shipped sunburst as an OnResolve a beat before the permanent lands, and said
reconsidering it "means giving the entry pipeline a narrow read of the record;
that is a follow-up, not this ADR". [#1002/#1011](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1011)
is that follow-up and it landed: `EntryCountersFromCast` declares the clause, and
`castCountsFor` reads `item.Paid.ColorsSpentCount()` off the resolving item
inside the CR 614 window. The caveat sentence §8 asked for is gone from those
cards.

`CastCounts` keeps reading the ONE record and never a count of its own —
`castCountsFor` is the single reader, `entry_counters.go`. Worth restating
because the temptation when the source kinds arrived was to give `CastCounts` a
`FromTreasure` bool; it does not need one, because the only card that would use
it (Marut's "*a Treasure token for each mana from a Treasure spent to cast it*")
creates tokens rather than entering with counters.

### A5. The auto-tapper's wish is an ordering HINT, and §5's declared limit stands

§5 declared: "**the AUTO-TAPPER does not spread colours**… spreading the PLAN is
a planner change with its own search-space cost, and it is out of scope here."
The source wish is the smallest thing that is not that.

`Spec.WantsManaFrom` declares the kinds a card reads back. It reaches exactly two
comparators — the candidate sort in `autoTapPreferringLocked` and
`orderUnusedByGenericPreference` — as a tiebreak applied **after Frozen and
before every existing criterion**. It never touches `gatherTapSources`'s
candidate set. So:

- the SOURCES are identical with and without a wish, and a cost that was payable
  stays payable. That is what lets `internal/legal`'s enumerator keep calling the
  wishless `autoTapLocked` for its bool and never offer a cast the gate refuses;
- a "spend a Treasure" card is never made uncastable by having no Treasure out.

The honest caveat is the one the Frozen ordering already carries: `solveColored`
shares an `AutoTapBudget`, so reordering can change which plan a pathological
board finds first. It cannot change whether one exists over the same set.

The `/autotap` preview endpoint takes the wish too
(`AutoTapForCostPreferringExcluding`), because the preview's plan is what the
client actually taps: without it a card previewed through the UI and the same
card cast with `AutoTap` set would spend different mana.

**DECLARED LIMIT, and it is a large one.** The auto-tapper refuses every
SACRIFICE-cost mana ability outright — `autoTapAbilityFor`'s first line is
`if !a.TapCost || a.SacrificeCost { continue }`, an S15 decision this amendment
does not revisit. **A Treasure is therefore not an auto-tap source at all**, and
no ordering hint can reach one. Treasure mana is floated by hand, the record
reads the same either way, and the hint bites today only on sources without a
sacrifice cost (a mana creature for Inga and Esika, a snow land). Filed
separately.

### A6. There is no "if snow mana was spent" card

The row this work came from asked for one. A scan of the dump for `snow mana` in
oracle text returns **zero** cards: snow reaches a payment as the `{S}` COST
symbol (CR 107.4g), and `ParsedCost.HasSnow` is still informational
(`mana_cost.go:39`). `ManaSpent.Snow()` ships anyway — it is one bit of the same
snapshot and the `{S}` work has nowhere else to read from — and is documented as
having no catalog reader today rather than left as a silent hook.

### A7. §4 restated, because #1212's readers are the first to make it bite

A copy of a spell records a real zero for the mana AND inherits the kicker, and
the two halves of CR 707.10 point different ways on purpose. CR 707.10b copies
"the choices made when casting", which is the modes, X, and whether an optional
additional cost was paid; WHICH MANA PAID is not one of them — mana is not an
object, nothing was spent to cast the copy, and the ruling under CR 707.10 on
Dawnglow Infusion says so. It is not a characteristic either, so CR 707.2 does
not reach it.

So: **a copy inherits the kicker and not the mana.** A copied Hired Hexblade
draws no card. Pinned by
`TestACopyOfASpellInheritsTheKickerButNotTheManaSpent`.

### Still out of scope

- **Spend riders.** Unchanged from the original: "when that mana is spent"
  (Pyromancer's Goggles, Scaled Nurturer), entry riders (Biophagus, Opal Palace),
  haste grants (Hall of the Bandit Lord) and per-spell "can't be countered"
  (Cavern of Souls, Delighted Halfling) need a fire point after a payment as well
  as this snapshot. The seams row stays open for them.
- **Source kinds read off an ACTIVATION's record.** Forsworn Paladin and Jetmir's
  Fixer print "if mana from a Treasure was spent to activate this ability". The
  tokens carry the kinds there already; nothing reads them, and the catalog
  declaration is per card rather than per ability.
- **The auto-tapper planning a sacrifice-cost source.** A5.
- **Enforcing `{S}`.** A6.
