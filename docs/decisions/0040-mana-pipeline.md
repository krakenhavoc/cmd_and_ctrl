# ADR 0040 — The mana pipeline: restricted, derived, scaled and gated mana (S32)

**Status:** Accepted · 2026-09-11 · Sprint S32 · Issue #352

## Context

The card-coverage roadmap ranks the mana pipeline **#2 of 24 missing
mechanics** across the next 2000 Commander cards — 98 cards it is the
sole blocker for, 183 it appears in, **115 it is the dominant blocker
for** — and until #352 it was the only entry in the top ten with no
tracking issue. Exotic Orchard, the highest-ranked card the catalog
did not have in the entire format (rank 9), sits in it.

Everything the engine knew about mana was declarative. A
`ManaAbility` had a cost (tap, sacrifice, sacrifice-another, and
since #267 life), a fixed `Produced` string, and an optional rider.
`ManaToken` had a `Restrictions []string` field that had been sitting
there since S15 carrying the comment *"a list of opaque tags the cost
validator can honour"* — and in the four sprints since, nothing had
ever written one and nothing had ever read one.

Five card families were unwritable as a result, and they are not one
thing:

| Sub-gap | Cards |
|---|---|
| a mana component in a *mana* ability's cost | the 10 Signets, the 10 Odyssey filter lands, Cabal Coffers |
| spend restrictions | Ancient Ziggurat, Eldrazi Temple, Shrine of the Forsaken Gods, Delighted Halfling, Cavern of Souls |
| colours derived from the board | Exotic Orchard, Reflecting Pool, Fellwar Stone, Mox Amber |
| scaled amounts | Cabal Coffers, Gaea's Cradle, Nykthos, Serra's Sanctum |
| activation gates | Temple of the False God, Mox Opal, the Tainted and Verge cycles |

A sixth — mana produced by a resolving *spell* — landed with batch 01
(#341): `AddManaForEffect` and the `AddMana` primitive, for Dark
Ritual and Mana Drain's refund. It covers a fixed produced string and
is untouched here.

## Decisions

### 1. Restricted mana ships whole or not at all

This is the load-bearing decision and it is a rule, not a mechanism.

A spend restriction has two halves: stamping the tag on production,
and refusing the payment at spend time. The production half is ten
lines. Shipping it alone would turn "add {C}{C}, spend only on
colorless Eldrazi" into plain `{C}{C}` on a land with no drawback —
Eldrazi Temple, Shrine of the Forsaken Gods and Delighted Halfling
would each ship **strictly better than printed**, which is the #259
rule and the one direction this project refuses.

So the enforcement came first, and the cards came after it.

### 2. Restrictions are tags, not closures

`ManaToken.Restrictions` stays `[]string`, with a small closed
vocabulary: `purpose:cast`, `purpose:activate`, `colorless`,
`type:X`, `subtype:X`, `supertype:X`. All must hold (AND).

**Why not a predicate func:** a restricted token is *game state*. It
sits in the pool across an undo boundary, `clone.go` deep-copies it,
and the client may one day want to render it. A func clones as a
pointer, compares to nothing, and logs as an address. Tags read in an
event log and in a test failure.

**An unknown tag denies.** A card file that invents a tag the matcher
does not know makes its mana unspendable rather than unrestricted.
Failing closed is the only safe default when the failure mode in the
other direction is "the card is better than printed".

### 3. `ManaSpendContext` — one struct threaded through every payment

`ManaPool.attemptSpend` needed to know *what* the mana is being paid
for. `ManaSpendContext` carries the purpose (cast / activate / unknown)
and the object's types, subtypes, supertypes and colours, built by
`ManaSpendForCast(Card)` / `ManaSpendForAbility(Card)`.

`CanPayFor` / `SpendManaFor` / `MissingFor` take it; the existing
`CanPay` / `SpendMana` / `Missing` survive as shorthands passing the
**zero** context, which matches no restriction at all. That is the
conservative default: a caller that has not said what it is paying
for gets no restricted mana. It also means every spend path that was
not updated fails safe rather than permissively.

Updated call sites, all of them: `applyCastCostLocked` and
`applyAutoTapLocked` (cast), `payAbilityManaCostLocked` (CR 602
abilities), `ActivateManaAbility` (the new mana component),
`legal.canPay` (move enumeration), and the `/auto-tap` preview
endpoint — so the "missing {R}{R}" breakdown a player sees matches
the payment the engine would attempt.

`payCostLocked` (pay-unless / may-pay) deliberately keeps the zero
context: such a cost is neither a cast nor an activation, and no
restriction the catalog writes today would admit it.

### 4. The solver spends restricted mana FIRST

`spendOrder` filters the pool to tokens the context permits and sorts
them most-restricted-first, pool order breaking ties. Restricted mana
is use-it-or-lose-it; spending the Ziggurat {G} before the Forest {G}
is the play a human makes, and the only one that does not waste it.

This finally implements the "restriction-first" heuristic the S15
comment on `CanPay` has claimed since the file was written — every
token compared equal because nothing ever set `Restrictions`.

### 5. `ProducedFunc` covers derived AND scaled with one slot

Both families are "the produced string is not known until activation":
Exotic Orchard's colours come from the opposing board, Cabal Coffers'
count from your Swamps. One `func(*Game, controller, source) string`
serves both, evaluated after the cost is paid (CR 605.3a — a mana
ability resolves the instant it is activated, all of it).

Returning `""` is a first-class answer meaning "produced no mana".
That is the printed behaviour for Exotic Orchard facing no opposing
lands and for Gaea's Cradle with an empty board: the source still
taps, nothing arrives.

**The recursion guard.** Deriving "what could that land produce"
SKIPS any ability that itself has a `ProducedFunc`. Two Exotic
Orchards, or an Orchard and a Reflecting Pool, would otherwise recurse
until the stack ran out. CR 106.6b answers the circular case with "no
mana" and so does this; where the real rules would resolve a one-way
chain it is one colour short, which is the weaker-than-printed
direction and is declared on all four card files.

**Producible mana is read from two sources** — the card's mana
abilities (catalog, intrinsic or the synthetic basic-land shape)
unioned with `Card.ProducedMana`, Scryfall's `produced_mana` stamped
at deck import. Neither alone is enough: the catalog knows 300-odd
cards, and a decklist's other 60 lands only have the Scryfall field.

### 6. A `Condition` predicate for activation gates

`func(*Game, controller, source) bool`, checked before any cost is
validated (CR 602.5a), returning `ErrConditionNotMet`. A failed gate
taps nothing and spends nothing.

Both this and `ProducedFunc` are **read-only and run under `g.mu`** —
held for write by `ActivateManaAbility`, for read by the auto-tapper.
A public locking mutator inside one deadlocks; the doc comments say
so on both the `game` and `effects` copies.

### 7. The auto-tapper refuses two new source kinds

`autoTapAbilityFor` already skipped sacrifice costs, life costs and
riders on one principle: *no further player decisions, no hidden
costs*. Two more now join them.

- **A mana cost is recursive.** Funding a Signet's `{1}` means solving
  a second cost to pay the first. The activation path deliberately
  refuses to auto-tap into a mana ability anyway — a mana ability
  resolves with no priority window, and "tap three lands to filter
  one" is a decision with consequences the planner cannot weigh. The
  player floats the mana and clicks, which is how a Signet is played
  on paper.
- **Restricted output is a decision, not a resource.** Spending a
  Ziggurat {G} on the cast in front of you may be right, or may waste
  the only mana that could have cast the creature you were saving it
  for.

Both are simplifications in the **weaker**-than-printed direction:
the cards stay fully hand-activatable from the ability menu, and mana
already floated from them pays for a cast exactly like any other —
that is what the spend context is for.

`materializePlanLocked` re-evaluates the gate and the derived output
rather than carrying them over from planning, so planner and executor
can never disagree, and it stamps restrictions defensively even
though the filter above means it should never see one. *"The
auto-tapper is the one path that mints unrestricted copies of
restricted mana"* is precisely the bug #259 warns about, and one line
is cheaper than trusting a filter two files away.

### 8. Restrictions ride the `PendingChoice`, not just the ability

A pipe slot (`{W|U|B|R|G}`) does not mint its token at activation —
it queues a `PendingChoiceMana` and the token appears later, in
`ResolveManaChoice`, by which time the ability shape is out of scope
and the source may have been sacrificed. `PendingChoice.ManaRestrictions`
carries them across.

This is **new game state** and is deep-copied in `clone.go` beside
`ColorOptions`, for the reason every slice there is: an undo that
shared the backing array would let the restored game mutate the live
one — and the thing being shared decides what the mana may legally
pay for.

## Consequences

- **21 cards.** The ten Signets, Cabal Coffers, Gaea's Cradle, Temple
  of the False God, Mox Opal, Exotic Orchard, Reflecting Pool, Mox
  Amber, Delighted Halfling, Shrine of the Forsaken Gods and Eldrazi
  Temple are new; **Fellwar Stone was fixed**, replacing a declared
  over-permissive simplification (full five-colour pipe regardless of
  the board) that was the only batch-01 simplification pointing the
  #259 direction.

- **The bulk of the remaining group is now data.** Of the 96 cards
  still naming the mana pipeline as their dominant blocker, at least
  29 need no further engine work: the ten Odyssey filter lands, the
  seven Duskmourn Verges, the four Tainted lands, Nimbus Maze,
  Prismatic Lens, Castle Garenbrig, Spire of Industry, Great Hall of
  the Citadel, Cabal Stronghold, Circle of Dreams Druid and Elvish
  Archdruid.

- **Still blocked, and on other mechanics.** "As this enters, choose a
  creature type" (Cavern of Souls, Secluded Courtyard, Unclaimed
  Territory) is roadmap rank 12, not this. Chrome Mox needs imprint;
  Chromatic Lantern and The World Tree need a layer-6 static that
  grants a mana ability to other permanents; Nykthos needs devotion;
  Gilded Lotus needs "three mana of any ONE colour", which is one
  colour pick producing N tokens rather than N independent picks.

- **`ManaAbilityCost` is complete but for counters.** S15 left a note
  that "mana / life / counter sub-costs land with later sprints when a
  catalog card demands them". #267 took life; this takes mana. No
  catalog card has yet demanded counters.

## Addendum (#742): one colour pick, N tokens

"Three mana of any ONE colour" (Gilded Lotus, Lotus Field) is now
expressible, without a new ability field. The produced-mana grammar
takes a count after a colour inside a pipe brace: `"{W3|U3|B3|R3|G3}"`
is one `mana_pick` whose answer mints three tokens of the picked
colour, and the counts may differ per colour (`"{G4|U1}"` is Nyx
Lotus's devotion). The counts ride on `ProducedManaEntry.Amounts` and
`PendingChoice.ManaAmounts`, reach the wire as
`PendingChoiceView.color_amounts`, and are carried by clone and the
snapshot. Three rules keep the rest of the pipeline unchanged:

- A single-option count (`"{G3}"`) is expanded by the parser into three
  ordinary `{G}` slots, so nothing downstream sees an amount for the
  common case, and a zero count drops its option.
- `ResolveManaChoice` and `AddManaForEffect` mint the picked colour's
  amount; an ordinary pick has no entry and mints one.
- ~~`AddManaForEffect` still narrows a pick to the commander's colour
  identity by default.~~ *Superseded 2026-09-17, see the addendum
  below:* `AddManaForEffect` offers the printed colours with the
  commander's identity listed first, and
  `AddManaOptions{NarrowToCommanderIdentity: true}` (and the `AddMana`
  primitive's field of the same name) is the effect-side twin of the
  mana ability's narrowing flag, for printed "in your commander's color
  identity" text. No effect in the catalog sets it.
- **The auto-tapper plans around a one-colour-N-mana source.** Its model
  is one slot, one mana, one colour choice per slot, and a Gilded Lotus
  planned as three any-colour slots could be booked for `{W}`, `{U}` and
  `{B}` at once, a plan the one-colour activation cannot honour. It is
  the restricted-output exclusion's shape: the player taps the source by
  hand (one prompt) and the cast spends the floated mana. Planning such
  a source inline, choosing the colour that pays the most of the
  remaining requirements, is possible later if it turns out to matter.

## Addendum — 2026-09-17: "any color" offers all five, identity first

**Owner decision.** A mana source whose printed text says "any color" or
"any one color" offers **all five colours**, with the controller's
commander colour identity listed first. Nothing is narrowed away. Before
this, the engine narrowed every multi-option ("pipe") slot to the
commander's identity by default, and a card had to opt out with
`IgnoreCommanderIdentity`. Birds of Paradise, Treasure, Lotus Petal and
the other "any color" sources that had not opted out showed a mono-green
deck a single `{G}` button, which is narrower than the printed card.

- **Default: order, don't narrow.** `game.manaPickOptionsFor` is the one
  list that `ActivateManaAbility`, `AddManaForEffect` and the auto-tapper
  (both the planner and the executor) read. It keeps the printed option
  set and moves the identity's colours to the front, keeping printed
  order within each group. Under a Golgari commander, Birds offers
  `B, G, W, U, R`, and a Scrubland offers `B, W`. With no identity (no
  commander, or a placeholder with no colours), the printed order is
  left alone.
- **`NarrowToCommanderIdentity` replaces `IgnoreCommanderIdentity`** as
  its inverse, on `ManaAbilityShape`, `effects.ManaAbility`,
  `AddManaOptions` and `effects.AddMana`. Only Command Tower, Arcane
  Signet, Commander's Sphere and Path of Ancestry set it, because their
  text says "any color in your commander's color identity". The
  intersection keeps its no-overlap fallback to the raw set.
  `TestOnlyCommanderIdentityCardsNarrow` pins the list, and the
  dump-gated `TestNarrowToCommanderIdentityMatchesOracleText` checks that
  a catalog card narrows exactly when its oracle text has the clause.
- **Why the order does the work.** The client renders `color_options` in
  the order sent, so the identity colours are the first buttons.
  `legal.EnumerateFor` lists the answers in that order, and every bot
  policy breaks ties on the lowest index. The heuristic still prefers the
  colour its hand needs, then falls back to the first (identity) colour.
  The auto-tapper's `pickColorForSlot` falls back to `options[0]` for a
  slot that only pays generic mana, so a generic cost is paid in an
  identity colour. A coloured requirement outside the identity, such as
  a stolen card's `{W}`, is now payable from Birds, as printed.
- **No wire change.** `color_options` keeps its shape. Only its order,
  and the width it has for the formerly narrowed cards, change.
