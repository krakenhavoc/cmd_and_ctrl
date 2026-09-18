# ADR 0070 — A stack item records the mana spent to cast it: converge, sunburst, adamant, and "if no mana was spent"

**Status:** Accepted · 2026-09-18 · S44 — Mana and cost components
**Issue:** [#761](https://github.com/krakenhavoc/cmd_and_ctrl/issues/761)
**Numbering:** on 2026-09-18, after `git fetch origin`, every remote branch
was enumerated (`git ls-remote --heads origin`) and every ADR file reachable
from any ref listed (`git log --all --name-only -- docs/decisions/`). The
highest number present anywhere is 0069 (`0069-face-down-objects.md`);
0061–0064, 0066 and 0069 are today's other agents. 0065, 0067 and 0068 are
unpushed branches at the time of writing and are left alone. 0005, 0024,
0029 and 0030 stay permanently unused per AGENTS.md §4.
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
