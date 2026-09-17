# ADR 0020 — Activated abilities in the catalog

**Status:** accepted (S21 sub-PR 2)
**Supersedes:** nothing. Extends [ADR 0018](0018-triggers-on-the-stack.md).

## Context

By the end of S20 the catalog could express four of the five ways a
card does something: spells resolve (`OnResolve`), triggers fire
(`Triggered`, S19), statics apply continuously (`Static`, S16) and
replacements rewrite events (`Replacements`, S17). The fifth —
CR 602 activated abilities — did not exist.

`Game.ActivateAbility` accepted a card ID and a free-text label and
pushed a stack item with no cost and no effect. Players paid the
cost by hand (tapping the permanent themselves, dragging the
sacrificed creature to the graveyard) and resolved the effect in
their heads. That was tolerable for "{T}: add mana" — which has its
own path anyway — and untenable for the S21 goal, because an
aristocrats deck IS its sacrifice outlets. "Sacrifice a creature:
deal 1 damage" isn't an effect with a cost attached; the cost is
half the card, and the dies-triggers it causes are the whole
strategy.

## Decisions

### 1. One shape, mirroring the rest of the catalog

`Spec.Activated []ActivatedAbility`, converted at the wire hook into
`game.ActivatedAbilityShape` — the same import-cycle dodge
`ManaAbilityShape` uses. An entry is a label, a cost, an optional
target clause, a sorcery-speed flag and an `Effect` closure.

The effect is the *same* `func(g *Game, item *StackItem) error`
S19 gave triggered abilities. That was the whole point of ADR 0018's
closure: resolution was already written. `resolveTopAbilityLocked`
runs the CR 608.2b re-check and then the effect, and it cannot tell
an activated item from a triggered one.

### 2. Costs are a struct, not a string

`AbilityCost{Tap, SacrificeSelf, SacrificeOther, Mana, Life}` —
additive fields rather than a parsed cost line. Parsing "{2}, {T},
Sacrifice a creature:" from oracle text would be a second
mini-language to maintain against the same 20-odd shapes the
constructors already cover (`TapCost()`, `SacrificeACreature()`,
`Plus(...)`).

`SacrificeOther` reuses `TargetSpec` for its "what may I sacrifice"
predicate, because the shape is identical. It is emphatically not
targeting: a sacrifice cost doesn't target, so hexproof and
"can't be the target of" never apply to it, and the client picks it
from a plain list instead of the board-click targeting flow. The
controller restriction (CR 701.21a) lives in the engine's cost
validation rather than the spec.

### 3. Validate everything, then pay everything

Every cost component is checked before any is paid. A tapped Krenko
fails without sacrificing; an illegal target fails without tapping.
This is the same discipline S21 sub-PR 1 applied to sacrifice-cost
mana abilities, and it matters more here because the costs are
heterogeneous — a partial payment would leave the board in a state
no rule describes.

Payment order within the commit is mana → tap → life → sacrifice.
Sacrifices go last because they move cards, which invalidates the
`*Card` pointer the tap needed.

### 4. Costs are paid at announce, so their triggers sit above

Paying a cost happens during activation (CR 601.2h), not at
resolution. So sacrificing a creature to Goblin Bombardment puts the
creature's dies-trigger on the stack *above* the Bombardment
ability, and it resolves first. Blood Artist drains before the
damage lands. The engine gets this for free by running the state
checks at the end of activation, after the item is already on the
stack — but the ordering is load-bearing for the whole archetype, so
there's a test that asserts it directly rather than trusting the
side effect.

### 5. The free-form path stays

`ActivateAbility` (label-only, no cost, no effect) still handles
non-catalog cards, which is most of the corpus. The action layer
picks the path by whether the payload names an `ability_index`. A
card gains a real ability by being added to the catalog, not by a
migration.

## Consequences

- Sac outlets work end to end: Goblin Bombardment, Carrion Feeder.
- Tap abilities enforce summoning sickness (CR 302.6), which the
  free-form path never did — Krenko can't tap the turn he lands.
  `CardView.summoning_sick` now rides the wire so the menu can grey
  the entry instead of failing the click.
- The activation menu is the existing right-click popover, with the
  activated abilities listed under the mana abilities.

## Out of scope

- **Mana abilities with a non-self sacrifice cost** (Ashnod's Altar:
  "Sacrifice a creature: Add {C}{C}"). It's a mana ability, so it
  must not use the stack (CR 605.3b), but its cost needs the
  sacrifice picker this ADR builds for stack-using abilities. It
  fits neither surface cleanly; it wants a third path where the
  activation collects a cost choice and then resolves immediately.
- **X in an activation cost** ({X}: …), **counters as a cost**
  (`Remove a +1/+1 counter:`), and **"activate only once each
  turn"** — all wanted by cards further down the S21 list, none
  needed by the three cards here.
- **Loyalty abilities** keep their own `ActivateLoyalty` path.

## Addendum (#625): counters as a cost

The "counters as a cost" line under *Out of scope* is now in scope, in
part. `AbilityCost.RemoveCounters` is a `game.CounterRemovalCost{Counter,
N, From}`: remove N counters as part of paying the cost. It covers three
printed shapes, and nothing else.

| Shape | Printed | Declared as |
|---|---|---|
| self | "Remove a gold counter from this artifact" | `From == nil` — `RemoveCountersFromThis("gold", 1)` |
| other permanent | "remove a loyalty counter from a planeswalker you control" | `From` is a `TargetSpec` — `RemoveCountersFrom("loyalty", 1, "a planeswalker you control", Planeswalker())` |
| any kind | "Remove a counter from a creature you control" | `Counter == ""` — `RemoveCountersFrom("", 1, "a creature you control", Creature())` |

**The choice is made at announce.** The permanent rides
`ActivateAbilityParams.CounterSourceIDs` (omitted for the self form) and,
for the any-kind form, the kind rides `CounterKind`, beside the
sacrifice and crew picks. No pending prompt is involved.

**§2 still holds: it is a predicate, not targeting.** `From` reuses
`TargetSpec` the way `SacrificeOther` does, and the engine matches it
with `specMatchLocked(..., false)`. A hexproof or shrouded permanent you
control pays. "You control" is enforced by the engine, not the spec.
`game.CounterCostOptionsForEffect` is the single candidate walk (built on
`SpecCandidatesForEffect`, not `LegalTargetsForEffect`) that the
protocol view and the move enumerator both read, so the client, the bots
and the engine agree on which permanent pays.

**§3 still holds: validate everything, then pay everything.**
`validateCounterRemovalCostLocked` runs in the validate block beside the
crew check: exactly one permanent for the other form, controlled by the
activator, matching `From`, holding at least N of the kind (else
`ErrInsufficientCounters`). A bad target, sacrifice or crew ID on the
same activation removes no counter. Payment order is now
mana → tap → crew → life → loyalty → **counters** → sacrifice. Counters
come off before sacrifices so that a self-form removal on a source that is
also sacrificed still finds the source.

Three rules come with the component and are enforced by the engine rather than by each card:

- **Not replaceable.** The removal goes through `applyCounterLocked`, not
  `AddCounterForEffect`. It is a cost, not an effect, so nothing that
  doubles or modifies counters applies. This is the loyalty cost's rule.
- **Not a loyalty activation.** Removing a planeswalker's loyalty counter
  to pay another permanent's cost does not stamp
  `LoyaltyActivatedThisTurn` and has no sorcery-speed window. The walker
  can still activate its own loyalty ability that turn. A walker paid
  down to 0 dies to the CR 704.5i state-based action the activation
  already runs, with the ability on the stack.
- **No timing of its own.** Heart of Kiran's crew-by-counter is instant
  speed, like crew.

**A 0/0 that pays with its last counter dies, and a `*` creature does
not, by the toughness rule in [ADR 0007 §7](0007-stack-foundation.md)
as amended by #683, not by a rule of the cost's.** The payment goes
through `applyCounterLocked` like any other counter removal, and the
state-based check the activation already runs applies that rule with
the ability on the stack. What the cost adds is that the case becomes
ordinary play: Mikaeus, the Lunarch paying his team pump with his last
counter, or Fain, the Broker spending a 0/0's last counter. A card cast
for X=0 never had a counter, so it is still skipped; that gap is noted
on the X cards, and declared to players on Mikaeus.

**"Rather than pay" on an activated ability is a second ability entry,
not an alternatives slot.** Heart of Kiran lists "Crew 3" and "Crew —
remove a loyalty counter from a planeswalker you control" as two
abilities with the same effect. The client already lists abilities
separately. What #259 requires is that the second entry has a real
cost, and it does: it cannot be activated without a planeswalker that
holds a counter.

`effects.Register` panics on `N <= 0`, and on an any-kind cost with
`N > 1`, because one kind choice cannot say how "remove two counters"
was paid when the two could be different kinds. `effects.Plus` merges
the field. Without that, `Plus(TapCost(), RemoveCountersFromThis(...))`
would silently drop the counter and the ability would be free.

Bots see the price on `legal.MoveCost.Counters` (`{card_id, counter,
n}`), priced against the permanent the counters come off, not the
move's source. That follows the #74 and #547 precedent: a cost the wire
payload cannot name rides the move.

**Still out of scope:**

- A removal **split across several permanents**: Iron Spider, Stark
  Upgrade's "Remove two +1/+1 counters from among artifacts you control".
  One permanent per payment cannot express it, and the card keeps its
  caveat.
- A cost that **adds** a counter: Devoted Druid's "Put a -1/-1 counter on
  this creature".
- Counter costs on **mana abilities** (`ManaAbilityCost`).

`docs/engine-seams.md`'s counter-cost row lists what is still waiting.

## Addendum (2026-09-17): activation conditions (#743)

**Status:** Proposed · 2026-09-17 · tracked on
[#743](https://github.com/krakenhavoc/cmd_and_ctrl/issues/743). This
status covers this section only. The decisions above and the #625
addendum are unchanged and stay accepted. Open questions for the owner
are at the end of the section.

### Context

A non-mana activated ability can carry one activation restriction today:
`SorcerySpeed`. Nothing can say "Activate only if an opponent controls
four or more lands" (Tectonic Edge), "Activate only during your turn"
(Sanctum of Eternity) or "Activate only if this creature is attacking"
(Glint-Horn Buccaneer). CR 602.1b calls these activation instructions.
They restrict when the ability can be activated, they "function at all
times", and they are "not part of the ability's effect".

Checked on develop at `bcac391`:

- `game.ActivatedAbilityShape` (`server/internal/game/activated.go:211`)
  and `effects.ActivatedAbility` (`cards/effects/spec.go:467-473`) have
  Label, Cost, Targets, SorcerySpeed and Effect, and nothing else.
- `ActivateCatalogAbility` (`activated.go:335`) checks timing only for
  `SorcerySpeed` or a loyalty cost (`activated.go:384`). The enumerator
  does the same (`legal/abilities.go:56`).
- `ActivatedAbilityView` has `sorcery_speed` and no condition field
  (`protocol/view.go:1049`).

Mana abilities already have this shape. `ManaAbilityShape.Condition
func(g *Game, controller, source uuid.UUID) bool`
(`game/effect_hooks.go:183-197`, #352) is read-only and runs under
`g.mu`. It is checked before any cost in four places:

- `ActivateManaAbility` (`mutations.go:3373-3374`)
- the auto-tapper (`mutations.go:1190`, `autotap.go:189`)
- the enumerator (`legal/abilities.go:383`)

A false condition returns `ErrConditionNotMet` (`errors.go:267-272`), and
`effects.ControlsAtLeast` (`cards/effects/mana_derivation.go:274`) builds
conditions.

Two shipped cards settle for less:

- **Sanctum of Eternity** uses sorcery speed in place of "during your
  turn", and declares a caveat for it (`sanctum_of_eternity.go:21-33`).
- **Glint-Horn Buccaneer** leaves out its attacking-only ability
  (`glint_horn_buccaneer.go:14-24`). That ability also needs a discard
  cost (#660).

The census counts 29 cards where this is the only core blocker.

### Decisions

#### 6. `Condition` on the activated ability, with the mana ability's contract

Add `ActivatedAbilityShape.Condition` and `effects.ActivatedAbility.Condition`,
both typed `func(g *Game, controller, source uuid.UUID) bool`. They are
carried through `effects.buildDef` like every other field of the ability
(AGENTS.md §7, "Adding a `Spec` slot"). Nil means no condition, which is
every ability in the catalog today.

The contract is the mana ability's, word for word:

- it is read-only;
- it runs under `g.mu`, held for write by the activation and for read by
  the view and the enumerator;
- it uses `*ForEffect` accessors and never a public locking accessor.

`g.ActivePlayer()` takes the read lock, so a condition that calls it
deadlocks the activation. That is why the helpers in §10 read `g.Turn`
and `g.Seats` directly, as `b23IsYourTurn` already does.

`controller` is the activating player, and "you" in the printed text
means that player. `source` is the permanent's instance ID. That is
enough to read a per-source activation count later, so the signature
does not change when the once-each-turn count lands (see *Out of scope*).

`SorcerySpeed` stays a separate flag. Some cards print both (Speaker of
the Heavens: "only if you have at least 7 life more than your starting
life total and only as a sorcery"). The two fail with different errors
and different client copy.

#### 7. Checked once, at activation, before anything is announced or paid

The check goes in `ActivateCatalogAbility`, immediately after the timing
check at `activated.go:384`. That puts it before X, the loyalty checks,
the costs and the targets:

```go
if ab.Condition != nil && !ab.Condition(g, playerID, cardID) {
    return ErrConditionNotMet
}
```

CR 602.5 says a player "can't begin to activate" an ability that is
prohibited from being activated. So a failed condition stops the
activation at this point: no later check runs and nothing is paid. The
layers are already current here, because `RecomputeLayersIfStaleLocked`
runs above the controller check. A condition that counts creatures
therefore sees the post-layer board.

When both checks fail, the error is `ErrSorcerySpeedRequired`, because
the timing check runs first. That order is our choice, not a rule, and
it only changes which error the player sees.

The condition is **not** checked again at resolution. Under CR 602.1b it
is an activation instruction, not part of the effect. Suppose Tectonic
Edge is activated while an opponent has four lands, and that opponent
sacrifices a land in response. The ability still resolves.
`resolveTopAbilityLocked` needs no change.

#### 8. The enumerator checks the same closure

In `legal/abilities.go`, `activatedMoves` checks `ab.Condition` with the
same arguments, right after the `speed` check at `:56`. This is what
`manaMoves` already does at `:383`. The enumerator then never offers a
bot an activation that the engine refuses (#544).

#### 9. The view carries `condition_unmet`, evaluated for the controller

Add `ActivatedAbilityView.ConditionUnmet bool`
(`json:"condition_unmet,omitempty"`). It is true when the ability has a
condition and that condition is false right now. It is absent when there
is no condition, or when the condition holds. The flag is negative
because wire booleans use `omitempty`, which would drop a positive
`condition_met: false` from the JSON.

It is evaluated once, in `viewOfActivatedAbilities`, with the permanent's
controller as `controller`. That function already computes every
sacrifice, crew and target option from the controller's side. The list is
not stripped per viewer (`stampActivatedAbilities`, `view.go:1645`). One
value is right for every viewer: the condition is about the controller,
and only the controller can open the menu.

Every viewer receives the flag, so **a condition must read only public
information**. Counts of permanents, graveyard cards and cards in hand,
life totals, the turn and the step are all public. A future condition
that needs hidden information, such as which cards are in a hand, must
not be written as a `Condition` unless the flag is first stripped for
other viewers. None of the 29 census cards needs hidden information.

The client greys the row the way it greys `sorcery_speed`: one arm in
`abilityBlocked` (`client/src/lib/contextMenu.logic.ts`) and one in
`ManaAbilityMenu.svelte`, with the reason "activation condition not
met". The row's label is already the full printed ability, including the
instruction, so the reason does not repeat the clause.

Greying rather than hiding follows the `sorcery_speed` precedent from
S31. The player can see that the permanent has the ability, and that it
is not available right now.

#### 10. Helpers in `cards/effects`

Each helper returns a condition closure, and each is a pure read.

| Helper | Printed | Reads |
|---|---|---|
| `ControlsAtLeast(n, match)` (exists) | "only if you control N or more …" (Bonders' Enclave: "a creature with power 4 or greater") | the battlefield, as today |
| `OpponentControlsAtLeast(n, match)` | "only if an opponent controls four or more lands" (Tectonic Edge) | true when **one** opponent controls at least N on their own. In a four-player game the opponents' lands are not added together |
| `OpponentControlsMore(match)` | "only if an opponent controls more lands than you" (Weathered Wayfarer) | the same per-opponent comparison, against the activator's own count |
| `DuringYourTurn()` | "Activate only during your turn" | `g.Seats[g.Turn.ActiveSeat].ID == controller` (the body of `b23IsYourTurn`, moved to a shared helper) |
| `DuringStep(steps...)`, `DuringYourStep(steps...)` | "only during your upkeep", "only during combat", "only before blockers are declared" | `g.Turn.Step`, plus the `DuringYourTurn` check for the "your" forms. "During combat" is the list of combat steps. "Before blockers are declared" is every step up to and including declare attackers |
| `GraveyardAtLeast(n, match)` | threshold, "seven or more cards are in your graveyard" (Cephalid Coliseum, Barbarian Ring) | the controller's graveyard |
| `SourceIsAttacking()` | "only if this creature is attacking" | the source's `AttackingTarget != uuid.Nil`, which the combat code already sets and clears (CR 508.1k, 506.4) |
| `SourceHasCountersAtLeast(kind, n)` | "only if this enchantment has four or more quest counters on it" (Luminarch Ascension) | the source's counters |
| `LifeAtLeastAboveStarting(n)` | Speaker of the Heavens | `p.Life` against `game.StartingLife` |
| `AllOf(conds...)` | a card that prints two conditions | ANDs them |

"This turn" conditions read `Game.TurnTally` (#586) through
`TurnTallyFor`. Examples are Idol of Oblivion ("if you created a token
this turn") and Lagomos, Hand of Hatred ("if five or more creatures died
this turn"). They get no helper until a second card uses the same one.

The names are for the implementation to settle. Each row sets scope
only: a helper is added together with its first card, not before.

#### 11. Cards

- **Sanctum of Eternity** gets `Condition: DuringYourTurn()`, loses
  `SorcerySpeed`, and loses its caveat. It becomes activatable during
  your combat and end step, and in response on your own turn, as
  printed.
- **The first new cards** are Tectonic Edge, Weathered Wayfarer and
  Bonders' Enclave.
- The rest of the census list becomes ordinary catalog work through the
  batch issues. Three cards are misattributed and stay blocked on their
  own seams: Barad-dûr (amass), Hydra Broodmaster (monstrosity) and
  Broadside Bombardiers (boast).

### Out of scope

- **"Activate only once each turn"** and **boast**. Boast is CR
  702.142a: "Activate only if this creature attacked this turn and only
  once each turn". Both need a count of how many times *this source's*
  ability was activated this turn. Boast also needs "attacked this turn".
  `TurnTally` counts resolutions and triggers, not activations. That gap
  is the "Per-source activations-this-turn count" row in
  `docs/engine-seams.md`, and Beledros Witherbloom's untap waits on it.
  Once the count exists, a once-each-turn check is an ordinary
  `Condition` that reads the count by `source`, with this signature.
- **Restrictions that persist through a change of control** (CR 602.5b).
  This rule matters only for restrictions that carry state, like the
  count above. A `Condition` with no state is re-evaluated against the
  current controller on every activation, which is what the printed text
  means.
- **Restrictions on abilities granted by another object** (CR 602.5c).
  A granted ability is not a catalog `ActivatedAbilityShape` today.

### Consequences

- About twenty cards become catalog work with no further engine change,
  and Sanctum of Eternity loses its caveat.
- Each view build runs every declared condition once, on the snapshot
  path. At worst each run walks the battlefield. `ControlsAtLeast` already
  adds the same cost for mana abilities.
- Mana and non-mana conditions share one error, one signature and one set
  of helpers. A helper written for one kind of ability works for the
  other.

### Alternatives considered

- **Declarative fields** (`MinLands int`, `YourTurnOnly bool`, and so
  on). Every census card would add a field, and the mana ability's
  condition is already a closure. Rejected.
- **Widening `SorcerySpeed` into a timing enum** that includes "during
  your turn". That covers two of the 29 cards and none of the board-count
  conditions. Rejected.
- **Checking the condition again at resolution.** That is wrong under CR
  602.1b, and the player would pay for an activation that then does
  nothing. Rejected.
- **Hiding an ability whose condition fails, instead of greying it.** An
  alternative-cost offer with a failed `Condition` is hidden (ADR 0048
  §7). But an offer is one way to cast a spell, and a menu row is an
  ability the permanent has. Greying matches `sorcery_speed`.

### Implementation plan

1. **Server PR.**
   - the field on both structs, and the `buildDef` line;
   - the check in `ActivateCatalogAbility`, and the same check in the
     enumerator;
   - the `condition_unmet` view field, documented in `docs/protocol.md`;
   - the §10 helpers the cards below need;
   - the cards: Sanctum of Eternity, Tectonic Edge, Weathered Wayfarer and
     Bonders' Enclave;
   - the regenerated census;
   - `docs/engine-seams.md`: the activation-condition row moves to Closed.

   The server refuses a failed activation whether or not the client greys
   the row, so the cards can ship in this PR.
2. **Client PR.**
   - `condition_unmet` in `protocol.ts`;
   - the greying arm in `abilityBlocked` and in `ManaAbilityMenu.svelte`;
   - vitest cases.

### Test plan

- **Game package, failed condition.** Use an ability whose cost includes
  mana, {T} and a sacrifice. A failed condition returns
  `ErrConditionNotMet` and changes nothing:
  - the mana pool is unchanged;
  - the source stays untapped;
  - the permanent named for the sacrifice stays on the battlefield;
  - the counters for a counter cost stay on;
  - no stack item and no `EventTrigger` appear.
- **Game package, both checks.** A condition that holds lets the
  activation go through. On an ability with both sorcery speed and a
  condition, each check failing alone refuses the activation with its own
  error.
- **No resolution check.** Activate Tectonic Edge while an opponent has
  four lands, then have that opponent sacrifice one in response. The
  ability still resolves (CR 602.1b).
- **`DuringYourTurn`.** Refused on an opponent's turn. Allowed in your
  own combat, in your end step, and with an item on the stack. The Sanctum
  of Eternity regression test changes from "sorcery speed only" to this.
- **`OpponentControlsAtLeast`, four players.** Three opponents with two
  lands each do not meet "four". One opponent with four lands does.
- **`SourceIsAttacking`.** True while the creature is declared as an
  attacker. False before attackers are declared, and after combat ends.
- **Enumerator.** On the same board as the engine test, the ability is
  missing from the legal moves while the condition fails, and present
  once it holds.
- **Protocol.** `condition_unmet` is present only while the condition
  fails, and never present on an ability with no condition.
- **Client.** `abilityBlocked` returns the condition reason for
  `condition_unmet`. A sorcery-speed row keeps its own, more specific
  reason.
- **Catalog soak.** `go test ./internal/aiseat -run TestCatalogSoak` with
  the new cards in the catalog shows no refused move.

### Open questions for the owner

1. **Should mana abilities get the same flag?** `ManaAbilityView` has no
   condition field. So Temple of the False God with four lands, or Mox
   Opal with two artifacts, shows a clickable row that the server then
   refuses.
   - **(a)** Add `condition_unmet` to `ManaAbilityView` in the same two
     PRs, with the same closure and the same client arm.
   - **(b)** Leave mana abilities alone and track the gap separately.

   **Recommendation: (a).** It takes a few lines on each side. Without it,
   the same menu greys one kind of row whose condition fails and leaves the
   other clickable.

## Addendum (2026-09-17): sacrifice costs of N permanents (#747)

**Status:** Proposed · 2026-09-17 · tracked on
[#747](https://github.com/krakenhavoc/cmd_and_ctrl/issues/747). This
status covers this section only. It amends §2 and §3 above for every
sacrifice cost site. [ADR 0021](0021-additional-costs.md) has a short
matching addendum for the additional-cost site. Open questions for the
owner are at the end of the section.

### Context

Every sacrifice cost pays **exactly one** permanent besides the source.
So "Sacrifice two artifacts", "Sacrifice three Foods" and "Sacrifice five
Treasures" cannot be written. Checked on develop at `bcac391`:

- **Three cost sites share one validator.** The sites are
  `AbilityCost.SacrificeOther` (`game/activated.go:43-49`),
  `ManaAbilityShape.SacrificeOther` (`game/effect_hooks.go:134-147`) and
  `AdditionalCost.Sacrifice` (`game/additional_cost.go:48`). All three call
  `validateSacrificeCostLocked` (`activated.go:651`): the activation at
  `activated.go:443`, the mana ability at `mutations.go:3399`, and the
  additional cost at `additional_cost.go:140`.
- **The validator accepts only one pick.** It refuses anything else with
  `if len(chosen) != 1 { return nil, ErrInvalidParam }` (`activated.go:662`).
- **The wire and payment already handle a list.** `ActivateAbilityParams.SacrificeIDs`
  is a slice "so the wire shape survives contact with Altar of
  Dementia-style" costs (`activated.go:274-279`), and every payment loop
  already ranges over the IDs.
- **The enumerator hard-codes one.** It calls `combinations(pool, 1, 1, …)`
  at `legal/abilities.go:148`, `:410` and `legal/cast.go:202`, although
  `combinations` (`cast.go:440`) takes a lower and upper bound.
- **The client picker holds one choice.** `SacrificeCostModal.svelte`
  keeps a single `chosen` ID (`:27`).

Every sacrifice clause is built by `sacrificeSpec`
(`cards/effects/activated.go:116`), which is `TargetPermanent`. That
constructor stamps `Min: 1, Max: 1` (`cards/effects/targets.go:306`). The
card-specific helpers (`b12SacrificeAGoblin`, `b27SacrificeAForest`,
`b29SacrificeALand`) all call it. The three sacrifice views already put
the clause's `Min` and `Max` on the wire: `abilityClauseView`
(`protocol/view.go:2892`) for abilities and mana abilities, and
`viewOfLegalTargets` (`view.go:1513`) for the additional cost.

Six catalog cards leave the ability out and declare a caveat: Savvy
Hunter, Samwise Gamgee, Sai, Master Thopterist, Magda, the Hoardmaster,
Magda, Brazen Outlaw, and Transmutation Font. Transmutation Font also
needs "with different names".

### Decisions

#### 12. The count lives on the sacrifice clause: `Min == Max == N`

A cost of "sacrifice N" is a sacrifice clause whose `TargetSpec` has
`Min == Max == N`. The constructors:

```go
// "Sacrifice two artifacts" (Sai), "Sacrifice three Foods" (Samwise).
func SacrificeN(n int, label string, preds ...CardPredicate) game.AbilityCost {
    return game.AbilityCost{SacrificeOther: sacrificeSpec(label, preds...).WithCount(n, n)}
}
```

A mana ability takes `SacrificeN(...).SacrificeOther`, the same way it
takes `SacrificeACreature().SacrificeOther` today. An additional cost
gets `SacrificeNCost(n, label, preds...)`. `SacrificeACreature()` and the
existing helpers are unchanged, because a count of 1 is what they
already build.

For a sacrifice clause, `Min` and `Max` count the permanents sacrificed.
§2 still holds: this is a predicate, not targeting. The field comment on
`TargetSpec.Min` says so.

This is chosen over a separate `SacrificeCount int` on each of the three
cost structs, for three reasons:

- **One place, not three.** Every site already carries the
  `*TargetSpec`, so the validator, the enumerator and all three views get
  the count with no new field.
- **No silent merge bug.** A new `AbilityCost` field would also need a
  line in `effects.Plus`. If that line were forgotten,
  `Plus(TapCost(), SacrificeN(3, …))` would silently cost one Food.
  That is stronger than printed (#259), the same hazard the #625
  addendum found for counters. On the clause, the count cannot be
  separated from the thing it counts.
- **The wire already carries it.** All three views already send the
  clause's `Min` and `Max` as `sacrifice_options.min` and `max`. Today
  both are 1.

`effects.Register` panics on a sacrifice clause where any of these holds:

- `Min != Max`, or `Min < 1`;
- `CountFromX` is set;
- `AllowSame` is set;
- `Players` is set.

The variable forms are refused until their seam exists (see *Out of
scope*), so no card can declare one and have it quietly read as a fixed
count. Every clause in the catalog today is 1/1, so the guard changes no
shipped card. The first sub-PR proves that with the catalog-wide
registration test.

#### 13. Validation: exactly N, distinct, yours, matching

`validateSacrificeCostLocked` reads `n := spec.Max` and requires:

1. `len(chosen) == n`, else `ErrInvalidParam`.
2. No ID twice, else `ErrInvalidParam`.
3. Each permanent is on the battlefield (`ErrCardNotFound`), controlled
   by the payer (`ErrCardCallerMismatch`, CR 701.21a), and matches the
   clause through `specMatchLocked(..., false)` (`ErrIllegalTarget`). The
   `false` means no targeting check, so hexproof never matters.
4. None is the source when `SacrificeSelf` is also set, else
   `ErrInvalidParam`. This check exists today.

Any failure refuses the whole activation or cast with nothing paid (§3,
and ADR 0021 §2). All three sites share the validator, so each site gets
all four checks with no further change.

#### 14. Payment: one event per permanent, observed as one simultaneous exit

The payment order does not change: mana → tap → crew → life → loyalty →
counters → sacrifice. Each permanent still goes through
`sacrificePermanentLocked`, which emits one `EventSacrifice`. So "whenever
you sacrifice a permanent" and dies triggers fire once per permanent.

**The sacrifices of one payment are wrapped in
`beginSimultaneousExitLocked`**, together with the source when
`SacrificeSelf` is set. That wrapper is the batch that
`destroyPermanentsLocked` (`game/simultaneous.go:195`) already uses for
board wipes. The trigger harvester consults the batch for every event
kind (`triggers.go:278`), `EventSacrifice` included.

Here is why the batch is needed. Suppose Priest of Forgotten Gods
sacrifices Blood Artist and a Goblin together. Blood Artist must see both
deaths: leaves-the-battlefield and sacrifice triggers look back in time
(CR 603.10a). Paid one at a time, the outcome depends on slice order. If
Blood Artist goes first, it is gone before the Goblin dies, and one drain
is lost.

`beginSimultaneousExitLocked` has no indestructible filter of its own.
That filter lives in `DestroyPermanentsForEffect`, which this path does
not call. So an indestructible permanent can still be sacrificed, as CR
701.21a requires.

The payment loop in `payAdditionalCostLocked` and the mana ability's
payment loop open the same batch. The batch also fixes today's
two-permanent case, where `SacrificeSelf` and `SacrificeOther` on one
ability are paid in order. With a batch the result is the same as the
player picking the best order, which CR 601.2h allows ("in any order").

#### 15. The enumerator: N from the clause, and one payment per move when N ≥ 2

All three enumerator sites call `combinations(pool, n, n, cap)`, with `n`
taken from the clause. When the pool has fewer than N permanents,
`combinations` returns nil (`cast.go:441`), so the ability or spell is not
offered (#544).

- **N = 1:** unchanged. Every candidate is its own move, up to
  `MaxExpansionPerSource`.
- **N ≥ 2:** one sacrifice set per move, not every combination. The set
  is the first N candidates in a policy-neutral order:
  1. tokens before nontokens;
  2. then lower mana value;
  3. then the ability's own source last, when the clause admits it;
  4. then board order, so the result is stable.

  Three reasons:
  - **The budget.** The expansion loops put targets outside and
    sacrifice sets inside, and they share one budget of 12. Ten Treasures
    choose five is 252 sets. Enumerating them would spend the whole budget
    on the first target and never reach the second, which is the #544
    failure again.
  - **The choice is about loss.** The sets differ only in which
    permanents are lost. `crewPayment` answers the same kind of choice
    with one answer (`legal/abilities.go:319`).
  - **Neutral.** The ordering favours no policy. `legal` must not import
    `aiseat` (#687). #687 orders *targets* by threat, which is a different
    question.

The mana move's label (`abilities.go`, `len(sacs) == 1`) becomes a
comma-joined list of the sacrificed names.

#### 16. Wire and client: the existing `min`/`max`, and a multi-select picker

No new wire field. `sacrifice_options.min` and `sacrifice_options.max`
carry N on `activated_abilities[i]`, `mana_abilities[i]` and
`additional_cost`. `sacrifice_ids` is already a list on `activate_ability`,
`activate_mana_ability` and `cast_spell`. `docs/protocol.md` documents
both at the three places.

`SacrificeCostModal.svelte` takes the count from `options.max`. It
becomes a multi-select that confirms only when exactly N permanents are
chosen, and `onConfirm` takes `string[]`. At N = 1 it works as it does
today. The menu's "nothing to sacrifice" check in `abilityBlocked`
changes from "no options" to "fewer options than `min`". Its reason names
the count, for example "needs 3 Foods (you have 2)".

#### 17. Cards

- **Caveats removed:** Savvy Hunter, Samwise Gamgee, Sai, Master
  Thopterist, Magda, the Hoardmaster, and Magda, Brazen Outlaw.
  Transmutation Font depends on open question 2.
- **Added:** Priest of Forgotten Gods and Kuldotha Forgemaster. After
  that, the fixed-N census cards (Hedron Detonator, Whisper, Blood
  Liturgist, Teysa, Orzhov Scion, and the rest) go through the batch
  issues.

### Out of scope

- **Variable counts:** "Sacrifice X Treasures", and "one or more" (Radiant
  Lotus). These need `Min != Max` or `CountFromX`, an announced count, and
  a record of the count that was paid. §12's registration guard refuses
  them until then.
- **Effects that read what was sacrificed** ("the sacrificed creature's
  power"). These need last-known information on the stack item.
- **Restrictions on the set as a whole**, such as "with different names"
  (Transmutation Font). This could reuse a set-level `Validate` in the
  style of #682, but the enumerator would then have to search for a
  valid set rather than take the first N. See open question 2.
- **Ward's sacrifice cost** (`effects.WardCost.Sacrifice`) is a separate
  shape on a separate path.
- The misattributed census cards stay out:
  - Liliana, Dreadhorde General's sacrifice is an effect on resolution,
    not a cost;
  - Phyrexian Soulgorger needs cumulative upkeep (#567);
  - Jolene, the Plunder Queen also needs a replacement effect for token
    creation.

### Consequences

- Fourteen cards become catalog work, and five shipped caveats go away.
- A clause's `Min` and `Max` now have two meanings, depending on the
  field that holds the spec: a target count, or a sacrifice count. The
  registration guard and the field comment keep the second meaning to
  the one shape this addendum supports.
- Bots with a sacrifice-N cost always offer the cheapest-looking payment.
  A policy cannot choose to sacrifice a better creature. That is the same
  trade crew made.

### Alternatives considered

- **`SacrificeCount int`** on `AbilityCost`, `ManaAbilityShape` and
  `AdditionalCost`. That is three fields, one `Plus` line that must not be
  forgotten, and a count stored apart from its clause. Rejected (§12).
- **A `SacrificeCostSpec{Spec, N}` wrapper.** It would change the type of
  every existing `SacrificeOther` and `Sacrifice` field and every call site
  that reads one, for information the spec can already hold. Rejected.
- **Enumerating every combination for N ≥ 2.** This exhausts the
  expansion budget on the first target (§15). Rejected.
- **Paying the sacrifices one at a time, with no batch.** The
  trigger count would depend on the order of IDs on the wire. Rejected.

### Implementation plan

1. **Server PR, no catalog card changes.**
   - the registration guard;
   - the validator reads N;
   - the batch at all three payment sites;
   - the enumerator at all three sites;
   - the `SacrificeN` and `SacrificeNCost` constructors;
   - `docs/protocol.md`.

   The tests use fixture cards. No shipped card declares N ≥ 2 yet, so a
   human cannot hit an N-cost the client cannot pay.
2. **Client PR.** The multi-select `SacrificeCostModal`, the count-aware
   arm in `abilityBlocked`, and vitest cases.
3. **Cards PR.** The five or six caveat removals, Priest of Forgotten Gods,
   Kuldotha Forgemaster, the regenerated census, and
   `docs/engine-seams.md`'s "Sacrifice cost of more than one permanent" row
   moved to Closed.

### Test plan

- **N at every cost site.** Run a fixture N = 2 cost at each of the
  three sites: an activated ability, a mana ability, and an additional
  cost to cast. Each pays both permanents and emits two `EventSacrifice`.
- **Refused payments.** Each of these refuses with nothing paid, and each
  test asserts the pool, the tap state and the battlefield afterwards:
  - one ID for an N = 2 cost;
  - three IDs;
  - the same ID twice;
  - an opponent's permanent;
  - a permanent that does not match the clause;
  - the source named when `SacrificeSelf` is also set.
- **Simultaneity.** Blood Artist and a Goblin sacrificed together to an
  N = 2 cost drain twice, in both wire orders. Add a regression for the
  existing `SacrificeSelf` plus `SacrificeOther` case.
- **Registration guard.** `Register` panics on a sacrifice clause with
  `Min != Max`, with `CountFromX`, or with `Min` 0. The catalog-wide
  registration test still passes.
- **Enumerator.**
  - A pool of N-1 offers no move.
  - A pool of N offers one.
  - Tokens are chosen before nontokens.
  - A targeted N = 2 ability still reaches its second target.
  - Every offered move is accepted by dispatch.
- **Protocol.** `sacrifice_options.min` and `max` are N on all three
  views.
- **Client.** The modal cannot confirm below or above N. `abilityBlocked`
  names the shortfall.
- **Catalog soak** with the cards PR applied.

### Open questions for the owner

1. **Picker ergonomics for identical tokens.** Magda, Brazen Outlaw's
   "Sacrifice five Treasures" means five clicks and a confirm in a plain
   multi-select.
   - **(a)** A plain multi-select that confirms at exactly N.
   - **(b)** (a), plus a "Choose N for me" button that fills the selection
     in the enumerator's §15 order (tokens first, lowest mana value
     first). The player can still change the picks before confirming.
   - **(c)** When the pool has exactly N candidates, preselect all of
     them.

   **Recommendation: (b).** It is one button, it uses the order the bots
   already use, and it never confirms for the player. (c) fits in
   alongside it if wanted.
2. **Transmutation Font's "three artifact tokens with different names":
   in this work or not?**
   - **(a)** Out. The Font keeps its caveat, and "a set-level restriction
     on a sacrifice cost" becomes its own seam row.
   - **(b)** In. The clause gains a set-level validator (#682 style), and
     the enumerator searches for a valid set instead of taking the first
     N.

   **Recommendation: (a).** One card needs it. It turns §15's "take the
   first N" into a search, and the addendum stays a count and nothing
   else.
