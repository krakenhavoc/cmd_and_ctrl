# ADR 0011 — Mana pool, cost model, and auto-tapper (S15)

**Status:** Accepted · 2026-04-22 · Sprint S15

## Context

S14's catalog gave us 30 cards that resolve their effects without
manual intervention, but **paying for those cards stayed
sandbox**: the player typed mana in their head and clicked
`cast_spell` regardless of what was in their pool. The next step
in the rules graft is the mana economy itself — a typed mana
pool, parsed mana costs, mana-ability activation that drops mana
into the pool, and an auto-tapper that turns "cast Cyclonic Rift
overload" into one click on a 6-island Bant manabase.

S15 ships that economy as a **permissive default with an opt-in
strict mode**. Existing manual-paper-tracking flows keep working;
players who want the engine to enforce costs flip the
`gameplay.strictMana` setting and get the full pipeline:

1. Activate a mana ability → tokens land in `Player.ManaPool`.
2. Cast a spell → server parses `ManaCost`, checks the pool,
   spends the tokens or rejects with a structured
   `insufficient_mana` error frame the client renders into a
   "Cast anyway" + "Auto-tap & cast" toast.
3. Auto-tap-and-cast → server plans the tap pass with a
   backtracking solver, taps the plan, drops produced mana
   into the pool, and proceeds to the cost check — all atomic
   under one write lock.

The sprint deliberately does **not** ship snow mana, hybrid
preferences, filter-land sub-payment, Cavern of Souls tribe
locking, phyrexian self-pay, X-mid-cast sliders, or commander
identity from the layer system. Those are S17 and beyond.

## Decisions

### 1. Permissive default, opt-in `gameplay.strictMana`

The strict-mode cost gate is **off by default**. Players who
want the engine to enforce costs flip a per-client setting
(`gameplay.strictMana`) which the client auto-stamps on every
outbound `cast_spell` action as `strict: true`.

**Why:** The pre-S15 sandbox flow ("type your mana on paper, click
cast, the engine doesn't care") is the moderate posture for
rules-light tables, draft simulators, and judge-mode debugging.
Forcing strict mode on every player would break those flows.
Per-client opt-in lets a 4-player game run with two strict and
two permissive seats — the engine handles each cast on its own
terms.

**Cross-reference:** S11.5's settings system already shipped the
client-side schema + migration plumbing; S15 just adds one bool
to the gameplay tab.

### 2. `Player.ManaPool` as an ordered slice of `ManaToken`,
       not a `map[color]int` multiset

Each token is `{Color, Source, Restrictions}`; the pool is
`[]ManaToken` in tap order. Spending walks the slice and
deducts tokens that match the cost, preferring restricted
tokens first to preserve fungible mana for the next cast.

**Why:** Restrictions on real cards are per-token, not per-color
("this mana can only be spent on creatures", "this mana can only
be spent on artifact, enchantment, or land spells"). A multiset
loses that information. The slice also preserves source
attribution so the event log can render "spent {R} from
Mountain" — useful for debugging unexpected pool drains and for
future replay analysis.

**Trade:** O(n) spend instead of O(1). At realistic pool sizes
(≤20 tokens) this is irrelevant.

### 3. Empty all pools at every step boundary (CR 106.4)

`Game.advanceStepLocked` calls `emptyAllManaPoolsLocked` before
firing the new step's entry hooks. Mana that survived the
boundary is "mana burn"-style lost.

**Why:** CR 106.4 is unambiguous — pools empty between steps.
Sandbox could have skipped this, but the rule is small and
universal; keeping it lets the auto-tapper assume "what's in the
pool right now is what was tapped this step", which simplifies
the residual-cost reasoning the materializer does.

### 4. Pipe syntax `{W|U|B|R|G}` for any-color slots,
       resolved through `PendingChoiceMana`

Birds of Paradise's "add one mana of any color" is parsed as a
single produced-mana entry with `Options: ["W","U","B","R","G"]`.
At activation time, single-option slots drop straight into the
pool; multi-option slots queue a `PendingChoiceMana` choice the
controller resolves via `resolve_choice` with a `{color}` body.

**Why:** Reuses the S14 PendingChoice pipeline rather than
inventing a parallel "color picker" subsystem. The choice modal
the client already renders for Thoughtseize handles mana picks
with one branch addition. Single-option fast path keeps basic
lands one-tap-no-modal.

**Auto-tap shortcut:** when the auto-tapper plans a multi-option
slot, the materialization step picks the color greedy against
the cost requirements **without** queuing a PendingChoice — the
auto-tap contract is "no further player decisions." See
decision 8 below.

### 5. Single `activate_mana_ability` action, not a per-card
       wire shape

One action with `{instance_id, ability_idx}` covers basic lands,
Sol Ring, Arcane Signet, Birds of Paradise, and every future
mana rock. The server resolves the index against
`ManaAbilitiesForCard(card)` which returns either the catalog-
declared abilities (Sol Ring's `[{C}{C}]`) or a synthetic
basic-land shape (Forest → `[{G}]`).

**Why:** The mana-ability surface is small and uniform — tap to
add some mana with optional restrictions. A per-card wire shape
would explode the action catalog and make the dispatcher's
switch statement load-bearing. The single action also makes the
client's right-click "tap for mana" menu trivial: render one
button per `card.mana_abilities[i]`.

**Mana abilities don't use the stack** (CR 605.3) — everything
runs synchronously inside `ActivateManaAbility` under the write
lock.

### 6. Commander identity proxy via `distinctColorsInManaCost`,
       not the layer system

Arcane Signet's "add one mana of any color in your commander's
identity" needs the controller's commander identity. S15 derives
identity by parsing the commander's printed `ManaCost` and
collecting the distinct colors — sandbox proxy. The full
identity rule (per-card oracle text scan + color indicators +
hybrid + phyrexian + reminder text) is S17.

**Why:** The layer system is two sprints out. Hand-rolling the
identity from `ManaCost` covers 95% of real commanders correctly
(Atraxa is `{G}{W}{U}{B}` in cost AND identity; Edgar Markov
is `{W}{B}{R}` in cost AND identity). The misses are
double-faced commanders, color-indicator-only cards, and
"adds {U}" reminder text — all rare in practice. Documenting
the gap up front keeps S17 honest about what it has to fix.

### 7. Backtracking auto-tapper with restriction-first
       heuristic + 10k-node budget

`Game.AutoTapForCost(controller, cost, xValue)` returns a
`[]uuid.UUID` plan. Algorithm: greedy-with-backtracking,
restriction-first heuristic. Each available source contributes
slots from its parsed `ProducedMana`; for each colored
requirement the solver picks the most-restrictive un-used
source whose first matching slot can cover it; on dead-end it
backtracks to the previous requirement. After all colored
requirements land, generic recruits fill from remaining slots,
preferring colorless producers (Sol Ring) over any-color
(Birds) over plain colored (basics).

**Budget cap:** 10,000 node expansions. A pathological manabase
(12+ duals + filter lands + tri-lands) can blow the search tree;
the cap returns `(nil, false)` cleanly so the client falls back
to manual tapping. Real Commander manabases (38 lands + rocks)
resolve in microseconds.

**Why backtracking instead of LP/ILP:** the search space is
tiny (≤ 40 sources × ≤ 10 slots), the algorithm is debuggable,
and the heuristic + budget keep worst-case behavior well-bounded.
A linear-programming formulation would be more elegant but pulls
in a solver dependency for ~1 ms p99 work.

**What this does NOT do:** filter-land sub-payment (S17),
Cavern tribe-locking (later), phyrexian self-pay (S17),
hybrid-color preferences (greedy picks the first matching half),
X-mid-cast sliders with live recompute (auto-tapper consumes
the announced XValue verbatim).

### 8. Atomic auto-tap-and-cast under one write lock,
       greedy color-picking on materialization

When `cast_spell` arrives with `auto_tap: true`,
`applyAutoTapLocked` runs the planner, then immediately taps
each card in the plan and drops produced mana into the pool —
all under the same `g.mu.Lock()` the cast already holds. A
plan failure returns `*InsufficientManaError` **before** any
permanent taps, preserving all-or-nothing semantics.

For multi-option slots (Birds, Signet) the materializer
**bypasses `PendingChoiceMana`** and picks a color greedy
against the still-unsatisfied cost requirements: walk the
pending requirement list, find one whose options intersect
this slot, consume it, drop that color into the pool. When
no requirement matches, the slot drops its first option as
generic-eligible mana.

**Why bypass the choice queue:** auto-tap-and-cast's contract is
"no further player decisions." Queuing a color pick would defeat
the whole point of the affordance. The greedy pick is correct
because the planner already proved a satisfying assignment
exists; the materializer just realizes one such assignment.

**Why all under one lock:** the server's invariant is "one
mutation = one snapshot." If the auto-tap, mana drop, and cost
spend split into separate mutations, a competing action between
them could leak the auto-tapped mana to a different cast. The
single-lock atomicity also lets clients trust that an
`insufficient_mana` error after `auto_tap: true` means
"nothing changed" — no half-tapped board to clean up.

### 9. Structured `*InsufficientManaError` with `Missing []string`,
       not string parsing

The strict gate returns `*InsufficientManaError{Missing: ["{R}",
"{1}"]}` (typed, with `errors.As` round-trip preserved). The
WebSocket layer at `ws/hub.go` materializes this into an
`error` frame with `code: "insufficient_mana"`, `missing: [...]`,
`card_id: "..."`. The client renders the missing symbols
verbatim into the override toast and the auto-tap preview
modal's "missing" line.

**Why:** Parsing `err.Error()` for the missing symbols would be
fragile — a future change to error wording would silently break
the client UI. Typed structured errors make the contract
explicit. `Unwrap` returns the sentinel `ErrInsufficientMana` so
existing callers using `errors.Is` keep working.

### 10. Function-var hooks (`CatalogManaAbilities`) instead of
        method calls (cycle break)

`game/effect_hooks.go` declares `var CatalogManaAbilities
func(oracleID string) []ManaAbilityShape`. The `effects`
package's `init()` writes the function pointer. The `game`
package can't import `effects` (cycle: `game → effects → game`),
so the hook is the seam.

**Why:** Same dependency-inversion pattern as S14's `OnETB` /
`OnResolve` hooks. Tests in the `game` package install a
synthetic hook (`withCatalogHook`) so they can exercise
catalog-bound behaviour without importing the real catalog.

### 11. Sandbox simplifications kept explicit

- `commanderIdentityFor` reads from printed `ManaCost`, not the
  layer system (decision 6).
- Mana cost parsing accepts an empty `ManaCost` as costless
  rather than rejecting — Scryfall ingestion gaps don't wedge
  sandbox casts. An `EventCostWarning` fires for visibility.
- Mana abilities are tap-cost-only in S15. Sacrifice-cost
  abilities (Lotus Petal) are explicitly rejected; the spec
  exists but the activation path returns `ErrInvalidParam` so
  a future Lotus Petal entry fails loudly. *(Superseded: the
  activation path since S21 sub-PR 1, the PLANNER by the
  2026-09-22 amendment below.)*
- The auto-tap-preview lock UI is **modal-internal** — clicking
  a battlefield land before opening the modal does not pre-lock
  it. A "battlefield-level lock mode" is plausible follow-up
  work; sandbox is fine without it.

## Consequences

- **Three new mana-bearing catalog cards** ship with mana
  abilities: Sol Ring (`{C}{C}`), Arcane Signet
  (`{W|U|B|R|G}` narrowed by commander identity), and Birds of
  Paradise (`{W|U|B|R|G}` raw).
- **Strict-mode cast** rejects insufficient-mana with a
  structured frame the client renders into a toast carrying
  `Cast anyway` and `Auto-tap & cast` buttons.
- **Auto-tap-and-cast** atomically taps the planned permanents
  and casts the spell in one snapshot — Cyclonic Rift overload
  from a 6-Island Bant manabase is one keypress.
- **`AutoTapPreviewModal`** lets the player inspect the plan
  before committing, with per-row lock toggles that re-fetch
  the preview with that source excluded. Enter confirms; ESC
  cancels.
- **Manual `activate_mana_ability`** still works as a sandbox
  override — players can hand-tap a Mountain even with strict
  mode off, and the resulting mana sits in the pool until the
  next step boundary.
- **`docs/protocol.md`** documents the new wire fields:
  `PlayerView.mana_pool`, `CardView.mana_cost` +
  `mana_abilities`, `mana_pick` PendingChoice kind,
  `cast_spell` flags (`strict`, `force_cast`, `auto_tap`,
  `locked_sources`), the `insufficient_mana` error frame, and
  the `GET /games/:id/auto-tap-preview` endpoint.
- **S17 (layer system)** picks up the commander-identity
  refactor — the proxy in S15 is the explicit hand-off.
- **S19 (triggered abilities)** picks up "whenever you tap a
  land for mana" triggers — the EventManaAbilityActivated +
  EventManaAdded event log is the integration point.

---

## Amendment (2026-09-22, #1215): the planner may pay a cost that eats the source, and pays it last

Decision 11's third bullet said "sacrifice-cost abilities are explicitly
rejected". The ACTIVATION half of that stopped being true in S21 sub-PR 1,
which made a Treasure, a Lotus Petal and an Eldrazi Spawn really crack for
mana. The PLANNER half stayed, as one line at the top of `autoTapAbilityFor`:

```go
if !a.TapCost || a.SacrificeCost {
    continue
}
```

That folds two unlike clauses into one exclusion.
`ManaAbilityShape.SacrificeOther` — Ashnod's Altar's "Sacrifice a creature" —
asks WHICH permanent dies, and naming one is a decision the auto-tapper's
standing contract ("no further player decisions and no hidden costs") forbids
it to make. `ManaAbilityShape.SacrificeCost` eats the SOURCE: it names nothing,
asks nothing, and `ActivateManaAbility` pays it off the source's own ID. The
half that asks nothing was excluded with the half that asks, and the result was
the #273 failure shape on the commonest token in the format — three Treasures
and no untapped lands read as unpayable to the cast preview, to the strict cast
gate and to the legal-move enumerator, so no bot ever cracked a Treasure to
cast anything.

**What changed.**

1. **The exclusion splits.** `a.SacrificeOther != nil` still drops the source;
   `a.SacrificeCost` alone does not. `!a.TapCost` stays: a `tapPlan` is a list
   of permanents to TAP, so a sacrifice-ONLY ability (a Gold token, an Eldrazi
   Spawn or Scion) has no slot in the plan's shape yet and is still hand-
   activated. That gap is named rather than closed here.

2. **The executor pays it, in the activation's component order.**
   `materializePlanLocked` resolves the sacrifice through the same
   `validateSacrificeCostLocked` all three cost sites share, then pays it with
   `payCostSacrificesLocked` AFTER the tap and BEFORE the mana — the order
   `ActivateManaAbility` uses, and for its reasons: the tap has to happen while
   the permanent is still on the battlefield, and CR 605.3b makes the payment
   and the production one atomic step. The payment is re-validated at
   materialisation the way the gate, the summoning sickness and the counter
   cost already are, because a plan can arrive stale; and the source's
   CONTROLLER is re-checked for the first time, because CR 701.21a lets you
   sacrifice only what you control and a stale plan could otherwise have eaten
   a permanent that changed hands.

3. **The dies-triggers are NOT drained inside the executor.** Every caller
   already drains at the point a player would next receive priority —
   `CastSpell`, `ActivateCatalogAbility` and the special-action verb each end
   with `runStateChecksLocked`. That is what puts a Blood Artist trigger ABOVE
   the spell the Treasure was cracked to cast rather than under it; draining
   mid-announcement would have put it under.

4. **A self-sacrificing source is the LAST resort.** A new outermost key,
   `tapSource.Sacrifices`, sorts such a source behind every ordinary source in
   the coloured pass (`autoTapLocked`'s sort) and behind even a frozen one in
   the generic recruiter (`orderUnusedByGenericPreference`). Cracking a
   Treasure to pay a generic pip a basic could have paid spends a resource the
   player never agreed to spend, which is the bar the life-cost and
   counter-adding exclusions are held to — and a Treasure is a five-colour
   source, so without the key the any-colour generic tier would have recruited
   it AHEAD of the untapped Mountain beside it. Behind Frozen rather than in
   front of it, because a permanent that misses one untap comes back and a
   cracked Treasure does not.

Decision 8 is otherwise unchanged: a plan failure still returns
`*InsufficientManaError` before anything is tapped, and the whole pass is still
one write lock. What IS new under that lock is a battlefield exit — the
auto-tap payment can now destroy a permanent — so the undo stack's pre-mutation
clone is load-bearing for a case it never used to cover, and
`internal/ws/undo_sacrificed_mana_source_test.go` holds it.

The `/autotap` preview endpoint (`writeAutoTapPreview` →
`AutoTapForCostExcluding`) and the legal-move enumerator
(`legal.canPayExcluding` → `AutoTapForCostForEffectExcluding`) flip with the
planner and needed no code of their own, which is the point of their being one
solver with three readers.

**Still open**, named here rather than discovered later:

- A mana ability whose cost is a sacrifice with NO `{T}` (Gold, Eldrazi Spawn,
  Eldrazi Scion) is still not plannable. `tapPlan` models "permanents to tap",
  and a sacrifice-only source needs the plan to say "and this one is only
  cracked" — including for a source that is already tapped, which
  `gatherTapSources` skips today.
- ~~The ordering hint ADR 0068's 2026-09-22 amendment declares
  (`Spec.WantsManaFrom`, `tapSource.Wanted`) has to compose with the
  last-resort tier once it lands.~~ **Closed in this same PR**, which rebased
  onto #1212 — see the composition note below.

### The composition with #1212's source wish

#1212 landed first and gave the planner a SOURCE WISH: a card whose text reads
which mana paid for it (`Spec.WantsManaFrom` → `tapSource.Wanted`) prefers a
source it can read back. Its own header declared that hint **inert**, because
`autoTapAbilityFor` refused every sacrifice-cost ability and so a Treasure — the
source eighteen printed cards ask about — was never a candidate.

This PR makes it live, and the two features then meet on exactly one kind of
permanent: a Treasure is both the commonest WISHED source and a
self-sacrificing one. So the order of the two keys is the whole of the
composition, and it is:

```
Wanted → Sacrifices → Frozen → (restrictiveness | tier, slotCnt)
```

**The wish is read first**, in both comparators. Two reasons, and the first is
decisive: a wish ranked under the tier is a wish that never fires, because the
tier's whole job is to route around the exact source the wish names — the hint
would have gone straight from "inert because a Treasure cannot be planned" to
"inert because a planned Treasure is always avoided". And #1212 originally
placed `Wanted` *below* `Frozen`; raising it above is the consistent completion
rather than a second decision, since a wish allowed to beat "this permanent
will be DESTROYED" must also beat "this permanent misses one untap", which is
strictly the cheaper of the two.

One sentence for the whole order: **the card being cast gets the source it asks
for, and where it asks for nothing the planner spends the cheapest thing on the
board.** Hired Hexblade auto-taps the Treasure and draws; an ordinary creature
on the same board takes the lands and the Treasure survives.

One consequence worth naming, because #1212's code comment asserted the
opposite: `materializePlanLocked`'s `SourceKinds: manaSourceKindsOf(tappedForMana)`
used to be justified with "the auto-tapper never plans a sacrifice cost, so the
permanent is still here". It no longer is — the plan's own sacrifice runs
between the tap and the mint — so the pre-tap copy is now load-bearing on the
auto-tap path exactly as it always was on `ActivateManaAbility`'s. The code was
already right; only its reason changed.

## Amendment (2026-09-23, #1228): the planner may spend a card that is not on the battlefield

Decision 7 called the candidate set "every available mana ability on the
battlefield", and for four sprints that sentence had no exceptions worth
naming. CR 113.6 supplies two:

```
Simian Spirit Guide   Exile this card from your hand: Add {R}.
Elvish Spirit Guide   Exile this card from your hand: Add {G}.
```

CR 605.1a makes those mana abilities — they could add mana, they are not
loyalty abilities, they target nothing — so they resolve immediately, with no
stack and no priority window, and can be activated in the middle of paying for
a spell. And CR 113.6 is what lets them work at all from a hand: the ability
says where it functions, which is the saying the rule asks for. A Spirit Guide
is the ritual every deck that runs one runs it as, and to the auto-tapper it
was invisible: a hand full of them read as "missing {R}" to the cast preview,
to the strict gate and to every bot.

Five decisions.

### 1. The candidate set grows a second WALK, not a second solver

`gatherTapSources` keeps the battlefield loop it has had since S15 and gains
`gatherManaZoneSources` beside it, over `game.supportedManaAbilityZones`. Both
append to one `[]tapSource` and the solver, the budget, the backtracking and
the generic recruiter are untouched — a Spirit Guide is a source with one
`{R}` slot and nothing about the search knows where it came from.

The walk is over a declared list rather than over `p.Hand` for the reason
`legal.abilityZones` and `protocol.stampZoneAbilities` are each one list:
adding a zone should be adding it in one place. Today the list is the HAND
alone, and `effects.Register` refuses anything else at boot — a zone no
consumer walks is a declaration the engine silently ignores, and the card would
register, look complete on the catalog page and never produce a mana.

The per-card gates the hand walk does NOT apply are each a rule rather than a
shortcut: no `Tapped` check and no `{T}` (the payment is the card leaving the
zone), no `CanActivateManaAbilities` (Arrest and Cursed Totem restrict a
permanent, and layer 6 has nothing to say about a card in a hand — the same
omission `ActivateManaAbility` and `legal.activatedMoves` make on this arm), no
summoning sickness (CR 302.6 is about a permanent you control) and no `Frozen`
tier (a card about to be exiled will not miss an untap). The activation gate,
the `Condition`, the CR 903.4f narrowing and the CR 106.12b production window
are all asked exactly as the battlefield loop asks them, through the same
helpers, because none of them is a fact about the battlefield.

### 2. A third tier, below the second: `tapSource.LeavesHand`

#1215 made a sacrifice-self source the LAST thing the planner reaches for.
This is the tier below it:

```
Wanted → LeavesHand → Sacrifices → Frozen → (restrictiveness | tier, slotCnt)
```

One sentence: **a Treasure is a resource the player already put on the table
and a card in hand is a spell they have not played yet, so the planner spends
the Treasure first.** An untapped Mountain beats both, and a frozen Mountain
beats both too — missing one untap is a permanent coming back next turn.

`LeavesHand` sits ABOVE `Sacrifices` in both comparators, which is how "below"
is spelled: the keys apply in order and the first to differ decides, so testing
the hand bit first is what puts every hand source behind every Treasure.

The wish still wins, unchanged and for the unchanged reason: #1212's hint is an
explicit instruction from the card being cast and the tiers are the planner's
own thrift. No printed card wishes for mana from a card in hand today, so the
two do not yet meet; the order is stated rather than discovered later.

### 3. Two pickers, not one picker with a zone parameter

`autoTapAbilityFor` stays the battlefield picker and gains exactly one line —
the CR 113.6 predicate, so a Spirit Guide that somehow reached the battlefield
is not planned as a source there. `autoManaExileAbilityFor` is its
non-battlefield sibling, and like it is shared by the planner and the executor
so the two can never disagree about which ability index a planned card is going
to be spent for.

Two functions rather than one with a `zone` argument because the two exclude for
different reasons. The battlefield picker's seven exclusions are about a payment
the planner may not decide or may not afford. This one's demands are about what
a card in a hand even IS: the ability has to function from this zone, and it has
to pay by exiling itself — which has to be the WHOLE cost.

That last demand is stated in the picker rather than inferred from the boot
check two packages away, and it is the one line here that is really about
safety: a mana ability off the battlefield with NO cost would be a source the
planner could spend without limit. Every other component — a mana or life cost,
a rider, a discard, restricted output, a spent exhaust ability — is excluded for
exactly the reason the battlefield picker excludes it, and none of them becomes
safe because the source is in a hand. What survives the filter is "Exile this
card from your hand: Add {R}" and nothing else the catalog can express.

### 4. The executor RE-FINDS the card; the plan carries no zone

`plannedTap` is still `{CardID, OneColor}`. `materializePlanLocked` looks each
planned card up on the battlefield and, failing that, in the piles a mana
ability may function from, and takes the arm that matches where it actually is.

Carrying the zone on the plan would be a second opinion about where a card is,
and the whole discipline of this executor is that a plan can arrive stale: the
gate, the sickness, the counter cost and the sacrifice are all re-asked, and a
source that has changed drops silently rather than stranding what the plan had
already spent. A card that has left the hand since the plan was made is simply
not found, which is the same answer a permanent that left the battlefield gets.

Three things differ on the non-battlefield arm and all three are rules: the
"you" is the card's OWNER (CR 108.4), the production is not a tap for mana
(CR 106.12a) — so `produceManaLocked` is told so and no triggered mana ability
fires, because CR 605.1b wants a permanent tapped for mana and Wild Growth has
nothing to attach to a card in hand — and `producedManaPreviewLocked` is
therefore priced with `fromTap: false`, so Mana Reflection's "if you tap a
permanent for mana" does not double a Spirit Guide.

### 5. The undo stack covers a HAND exit now

#1215 made the auto-tap payment able to destroy a permanent and named the undo
stack's pre-mutation clone as load-bearing for a case it never used to cover.
This is the same sentence one zone over: the payment can now take a card out of
a player's hand, and a take-back has to put it back.
`internal/ws/undo_exiled_mana_source_test.go` holds it, beside
`undo_sacrificed_mana_source_test.go` that holds the other.

**Still open**, named here rather than discovered later:

- The sacrifice-only source (Gold, Eldrazi Spawn) named in the #1215 amendment
  is unchanged and still not plannable. It is a nearer neighbour than it was —
  this amendment proves a plan entry need not be a tap — but the fix is
  `tapPlan` growing a payment kind, not another zone.
- `AutoTapForCost*` returns `[]uuid.UUID`, so a plan containing a hand card
  reaches the `/autotap` preview and the client as an ID like any other. The
  client highlights planned permanents and simply finds nothing for a card in
  hand; showing "and this card out of your hand" in the preview is a UI
  improvement this PR did not make.
