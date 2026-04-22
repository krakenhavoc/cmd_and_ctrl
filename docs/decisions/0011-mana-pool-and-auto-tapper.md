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
  a future Lotus Petal entry fails loudly.
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
