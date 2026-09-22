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

**Status:** Accepted · 2026-09-17 · tracked on
[#743](https://github.com/krakenhavoc/cmd_and_ctrl/issues/743). This
status covers this section only. The decisions above and the #625
addendum are unchanged and stay accepted.
**Decided by the owner (2026-09-17):** mana abilities get the same
`condition_unmet` flag, on `ManaAbilityView`, in the same server and
client PRs. Temple of the False God and Mox Opal grey out like a
non-mana ability whose condition fails (§9). The options considered are
kept in [Decided questions (#743)](#decided-questions-743) at the end
of the section.

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
  (`protocol/view.go:1049`). `ManaAbilityView` has no condition field
  either (`view.go:1145`).

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

#### 9. Both ability views carry `condition_unmet`, evaluated for the controller

Add `ConditionUnmet bool` (`json:"condition_unmet,omitempty"`) to
**`ActivatedAbilityView` and `ManaAbilityView`** (owner decision, see
[Decided questions (#743)](#decided-questions-743)). It is true when
the ability has a condition and that condition is false right now. It is
absent when there is no condition, or when the condition holds. The flag
is negative because wire booleans use `omitempty`, which would drop a
positive `condition_met: false` from the JSON.

It is evaluated once per ability, with the permanent's controller as
`controller`:

- **Non-mana abilities:** in `viewOfActivatedAbilities`. That function
  already computes every sacrifice, crew and target option from the
  controller's side.
- **Mana abilities:** in `stampActivatedAbilities` (`view.go:1645`),
  beside `stampManaSacrificeOptions` (`:1665`, `:1717`). That pass has
  the game handle and the parsed controller. `viewOfManaAbilities`
  (`:2969`) takes only the card and cannot evaluate a closure. The
  closure is the `ManaAbilityShape.Condition` the activation, the
  auto-tapper and the enumerator already call, so no second copy of the
  rule is written.

Neither list is stripped per viewer. One value is right for every
viewer: the condition is about the controller, and only the controller
can open the menu.

Every viewer receives the flag, so **a condition must read only public
information**. Counts of permanents, graveyard cards and cards in hand,
life totals, the turn and the step are all public. A future condition
that needs hidden information, such as which cards are in a hand, must
not be written as a `Condition` unless the flag is first stripped for
other viewers. None of the 29 census cards needs hidden information.
Neither does any shipped mana-ability condition: they read permanents
controlled (Mox Opal, Temple of the False God, Shrine of the Forsaken
Gods, the Shapeshifter tokens from Springleaf Parade), counters on the source
(Gemstone Mine, Runaway Steam-Kin) and library size (Millikin).

The client greys the row the way it greys `sorcery_speed`: one arm in
`abilityBlocked` (`client/src/lib/contextMenu.logic.ts`) and one in
`ManaAbilityMenu.svelte`, with the reason "activation condition not
met". Both predicates already take mana and non-mana rows through one
cost-shaped type (`AbilityCost`, `CostShaped`), so `condition_unmet` is
added to that type once and the same arm greys both kinds of row.
`protocol.ts` gets the field on `ActivatedAbilityView` and
`ManaAbilityView`. The arm runs after the sorcery-speed and loyalty arms,
so those rows keep their more specific reasons, and after the existing
`cant_activate_mana` restriction check in `abilityItems`, which still
wins for a mana row. The row's label is already the full printed
ability, including the instruction, so the reason does not repeat the
clause.

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
- Mana and non-mana conditions share one error, one signature, one set
  of helpers and one wire flag. A helper written for one kind of ability
  works for the other, and the menu greys both kinds of row the same way.
- Temple of the False God, Mox Opal and the other shipped mana abilities
  with a condition stop showing a clickable row that the server refuses.
  Their conditions now also run on every view build, not only in the
  activation, the auto-tapper and the enumerator.

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
   - the `condition_unmet` field on `ActivatedAbilityView` and
     `ManaAbilityView`, documented in `docs/protocol.md` at both;
   - the §10 helpers the cards below need;
   - the cards: Sanctum of Eternity, Tectonic Edge, Weathered Wayfarer and
     Bonders' Enclave;
   - the regenerated census;
   - `docs/engine-seams.md`: the activation-condition row moves to Closed.

   The server refuses a failed activation whether or not the client greys
   the row, so the cards can ship in this PR.
2. **Client PR.**
   - `condition_unmet` on both ability views in `protocol.ts`;
   - the greying arm in `abilityBlocked` and in `ManaAbilityMenu.svelte`,
     reached by mana and non-mana rows alike;
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
  - Non-mana: Tectonic Edge's ability with an opponent at three lands,
    then at four.
  - Mana: Temple of the False God's `mana_abilities[0]` with the
    controller at four lands carries the flag, and at five it does not.
    Mox Opal with two artifacts carries it, and with three it does not.
  - A mana ability with no condition (a basic land) never carries it.
- **Client.** `abilityBlocked` returns the condition reason for
  `condition_unmet`, on a mana row and on a non-mana row. A
  sorcery-speed row keeps its own, more specific reason. A mana row that
  is both restricted (`cant_activate_mana`) and condition-failed shows
  the restriction.
- **Catalog soak.** `go test ./internal/aiseat -run TestCatalogSoak` with
  the new cards in the catalog shows no refused move.

### Decided questions (#743)

Answered by the owner on 2026-09-17. The chosen option is marked
**(chosen)**; the recommendation text is kept for the record.

1. **Should mana abilities get the same flag?** `ManaAbilityView` has no
   condition field. So Temple of the False God with four lands, or Mox
   Opal with two artifacts, shows a clickable row that the server then
   refuses.
   - **(a) (chosen)** Add `condition_unmet` to `ManaAbilityView` in the
     same two PRs, with the same closure and the same client arm.
   - **(b)** Leave mana abilities alone and track the gap separately.

   **Recommendation: (a).** It takes a few lines on each side. Without it,
   the same menu greys one kind of row whose condition fails and leaves the
   other clickable.

   Applied in §9, *Consequences*, the implementation plan and the test
   plan.

## Addendum (2026-09-17): sacrifice costs of N permanents (#747)

**Status:** Accepted · 2026-09-17 · tracked on
[#747](https://github.com/krakenhavoc/cmd_and_ctrl/issues/747). This
status covers this section only. It amends §2 and §3 above for every
sacrifice cost site. [ADR 0021](0021-additional-costs.md) has a short
matching addendum for the additional-cost site.
**Decided by the owner (2026-09-17):** the picker is a multi-select that
confirms at exactly N, plus a "Choose for me" button that fills the
selection in the enumerator's §15 order (tokens first, lowest mana
value, source last). The button never confirms (§16).
**Decided by the lead on the owner's standing guidance (2026-09-17):**
Transmutation Font's "with different names" is out of this work. The
Font keeps its caveat, and a seam row covers set-level restrictions on a
sacrifice cost (§17, *Out of scope*). The options considered are kept in
[Decided questions (#747)](#decided-questions-747) at the end of the
section.

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

**The order is one shared helper in `game`.** The enumerator's three
sites and the protocol views (§16) sort the same candidate list with the
same function, for example `game.SacrificePaymentOrderForEffect(source,
ids)`; the name is for the implementation to settle. It lives in `game`
because `legal` and `protocol` both import `game`, and neither may own a
rule the other copies. Every key it reads (token, mana value, instance
ID, board position) is public.

The mana move's label (`abilities.go`, `len(sacs) == 1`) becomes a
comma-joined list of the sacrificed names.

#### 16. Wire and client: the existing `min`/`max`, a multi-select picker, and "Choose for me"

No new wire field. `sacrifice_options.min` and `sacrifice_options.max`
carry N on `activated_abilities[i]`, `mana_abilities[i]` and
`additional_cost`. `sacrifice_ids` is already a list on `activate_ability`,
`activate_mana_ability` and `cast_spell`. `docs/protocol.md` documents
both at the three places.

**`sacrifice_options.cards` is sent in §15's order.** `sacrificeCostOptions`
(`protocol/view.go:2883`) and the additional-cost stamp (`view.go:1482`)
sort the controller's candidates with the shared helper before they go
on the wire. The client has no token or mana-value field to sort by, and
it must not derive the rule (#429). For N ≥ 2, the first N entries of
the list are exactly the set the enumerator offers a bot for the same
board. At N = 1 the enumerator still offers every candidate, and the
list's order is the only change.
`docs/protocol.md` documents the order at the three places.

`SacrificeCostModal.svelte` takes the count from `options.max`. It
becomes a multi-select that confirms only when exactly N permanents are
chosen, and `onConfirm` takes `string[]` (owner decision). At N = 1 it
works as it does today.

**"Choose for me"** (owner decision) is a button in the modal when
N ≥ 2. It replaces the current selection with the first N entries of
`sacrifice_options.cards`, in wire order. It does not confirm: the player
can still change any pick, and Confirm and Enter stay the only ways to
pay. It is disabled when there are fewer than N options, which the menu
already prevents from opening. The client does no ordering of its own.

The menu's "nothing to sacrifice" check in `abilityBlocked` changes from
"no options" to "fewer options than `min`". Its reason names the count,
for example "needs 3 Foods (you have 2)".

#### 17. Cards

- **Caveats removed:** Savvy Hunter, Samwise Gamgee, Sai, Master
  Thopterist, Magda, the Hoardmaster, and Magda, Brazen Outlaw.
- **Transmutation Font keeps its caveat** (lead decision): its "three
  artifact tokens with different names" is a restriction on the set,
  which this addendum does not build (*Out of scope*).
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
  (Transmutation Font). Decided out (see
  [Decided questions (#747)](#decided-questions-747)). It could reuse a
  set-level `Validate` in the style of #682, but the enumerator would then
  have to search for a valid set rather than take the first N, and "Choose
  for me" could no longer take the first N entries. The cards PR adds a
  "Set-level restriction on a sacrifice cost" row to
  `docs/engine-seams.md`, with Transmutation Font as its first card.
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
- The order of `sacrifice_options.cards` becomes part of the wire
  contract. A human's "Choose for me" and a bot's payment are the same
  set on the same board, because both come from one helper.
- Magda, Brazen Outlaw's "Sacrifice five Treasures" is one click and a
  confirm, not five clicks and a confirm.

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
   - the shared §15 order helper in `game`;
   - the enumerator at all three sites;
   - `sacrifice_options.cards` sorted by the helper at all three views;
   - the `SacrificeN` and `SacrificeNCost` constructors;
   - `docs/protocol.md`, including the option order.

   The tests use fixture cards. No shipped card declares N ≥ 2 yet, so a
   human cannot hit an N-cost the client cannot pay.
2. **Client PR.** The multi-select `SacrificeCostModal` with its
   "Choose for me" button, used by the ability picker and `Board.svelte`'s
   cast-time picker; the count-aware arm in `abilityBlocked`; and vitest
   cases.
3. **Cards PR.**
   - the five caveat removals (Transmutation Font keeps its caveat);
   - Priest of Forgotten Gods and Kuldotha Forgemaster;
   - the regenerated census;
   - `docs/engine-seams.md`: the "Sacrifice cost of more than one
     permanent" row moves to Closed, and a new "Set-level restriction on
     a sacrifice cost" row lists Transmutation Font.

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
- **Protocol.**
  - `sacrifice_options.min` and `max` are N on all three views.
  - `sacrifice_options.cards` is in §15's order on all three views: a
    token before a nontoken, a lower mana value before a higher one, and
    the source last when the clause admits it.
  - **Parity:** for an N = 2 fixture on the same board, the first N
    entries of the view's list equal the sacrifice set of the
    enumerator's move, at each of the three sites.
- **Client.**
  - The modal cannot confirm below or above N. `abilityBlocked` names
    the shortfall.
  - "Choose for me" selects the first N options in wire order and
    replaces an existing partial selection.
  - "Choose for me" never calls `onConfirm`. After it, changing one pick
    and confirming sends the changed set.
  - The button is absent at N = 1.
- **Catalog soak** with the cards PR applied.

### Decided questions (#747)

Answered on 2026-09-17: question 1 by the owner, question 2 by the lead
on the owner's standing guidance. The chosen option is marked
**(chosen)**; the recommendation text is kept for the record.

1. **Picker ergonomics for identical tokens.** Magda, Brazen Outlaw's
   "Sacrifice five Treasures" means five clicks and a confirm in a plain
   multi-select.
   - **(a)** A plain multi-select that confirms at exactly N.
   - **(b) (chosen)** (a), plus a "Choose for me" button that fills the
     selection in the enumerator's §15 order (tokens first, lowest mana
     value, source last). The player can still change the picks before
     confirming, and the button never confirms.
   - **(c)** When the pool has exactly N candidates, preselect all of
     them. Not chosen.

   **Recommendation: (b).** It is one button, it uses the order the bots
   already use, and it never confirms for the player. (c) fits in
   alongside it if wanted.

   Applied in §15 (the shared order helper), §16 (the wire order and the
   button), *Consequences*, the implementation plan and the test plan.
2. **Transmutation Font's "three artifact tokens with different names":
   in this work or not?**
   - **(a) (chosen)** Out. The Font keeps its caveat, and "a set-level
     restriction on a sacrifice cost" becomes its own seam row.
   - **(b)** In. The clause gains a set-level validator (#682 style), and
     the enumerator searches for a valid set instead of taking the first
     N.

   **Recommendation: (a).** One card needs it. It turns §15's "take the
   first N" into a search, and the addendum stays a count and nothing
   else.

   Applied in §17, *Out of scope* and the cards PR.

## Addendum (2026-09-18): the rest of counter costs (#789)

**Status:** Accepted · 2026-09-18 · S44 — Mana and cost components. Tracked
on [#789](https://github.com/krakenhavoc/cmd_and_ctrl/issues/789). Finishes
the seam the [#625 addendum](#addendum-625-counters-as-a-cost) opened, whose
"Still out of scope" list is now empty.

### Context

#625 shipped `AbilityCost.RemoveCounters` in three printed shapes and named
what it did not cover: counter costs on MANA abilities, a variable count, a
removal split across permanents, and a cost that ADDS a counter. Those four
blocked Vivid Creek and Vivid Grove (#451), Ramos (#391), Crucible of the
Spirit Dragon (#402), Mage-Ring Network (#449) and Tekuthal (#304) outright,
and shipped Devoted Druid (#301), Iron Spider (#393) and Hopeful Initiate
(#460) with caveats.

### 18. Four more answers to one question, not four more components

The new shapes are not new components. They are more answers to the question
the existing one already asks — *which counters come off, and from where* —
so `CounterRemovalCost` gains two flags rather than growing two siblings:

| Shape | Printed | Declared as |
|---|---|---|
| self | "Remove a gold counter from this artifact" | `RemoveCountersFromThis("gold", 1)` |
| other permanent | "remove a loyalty counter from a planeswalker you control" | `RemoveCountersFrom(…)` |
| any kind | "Remove a counter from a creature you control" | `Counter == ""` |
| **variable** | "Remove any number of storage counters from this land" | `RemoveCountersXFromThis("storage", 0)` — `Variable` |
| **among** | "Remove two +1/+1 counters from among artifacts you control" | `RemoveCountersAmong("+1/+1", 2, …)` — `Among` |
| **any-kind among** (#943) | "Remove three counters from among other artifacts, creatures, and planeswalkers you control" | `RemoveCountersAmong("", 3, …)` — `Counter == "" && Among` |

One type means one validator (`validateCounterRemovalLocked`), one candidate
walk (`CounterCostOptionsForEffect`), one enumerator arm, one protocol view
and one client picker. The alternative — a sibling type per shape — is four
validators that must agree about "you control it and it is not targeted",
which is four chances to disagree; #544's lesson is that the enumerator and
the engine disagreeing is the expensive bug.

The constructor names still read like the printed text
(`RemoveCountersXFromThis`, `RemoveCountersAmong`), because a card file
should say what the card says. The flags are the engine's business.

### 19. One component, two owners: `ManaAbilityShape.RemoveCounters`

The same `*CounterRemovalCost` now hangs off a mana ability, with the same
meaning and the same validator. `ManaAbilityCost.RemoveCounters` is built
from the same constructors (`RemoveCountersFromThis("charge", 1).RemoveCounters`),
and `ManaAbilityParams` carries the same three payment fields
`ActivateAbilityParams` does.

`AddCounter` rides along on both for the same reason, even though no printed
mana ability has one: a component declared once and owned by one ability
kind is a component the other kind has to learn about later.

Payment order on a mana ability is the CR 602 path's, minus the components
it does not have: **mana → tap → life → counters removed → counter added →
sacrifice**. Counters before the sacrifice for #625's reason — a self
removal has to find the source still on the battlefield.

This empties the last of S15's "mana / life / counter sub-costs land with
later sprints" note on `ManaAbilityCost`.

### 20. The auto-tapper plans a counter cost only when it can both decide and afford it

`autoTapAbilityFor`'s contract has always been "no further player decisions
and no hidden costs". Applied to counters, that is three conditions, checked
by `manaCounterCostPlannable` — one predicate shared by the planner
(`gatherTapSources`) and the executor (`materializePlanLocked`), for the
reason `manaTapBlockedBySickness` is shared: when the planner's copy is the
laxer one, the executor strands whatever the plan had already tapped.

- The counters come off the **source**. "From a creature you control" and
  "from among artifacts you control" both ask *which permanent*.
- The kind and the count are **printed**. An any-kind cost asks *which
  kind*; a variable cost asks *how many*, and the answer changes how much
  mana arrives.
- The permanent **holds enough right now**. This is the "never plan a Vivid
  land with no charge counters" rule, and the executor re-asks it, because
  a plan can arrive stale.

Vivid Creek and Vivid Grove pass all three, which is the point: the
commonest counter-cost mana ability in the game auto-taps like any other
land until its charge counters run out, and then quietly stops being a
five-colour source — exactly as it stops being one in paper.

A cost that ADDS a counter is never planned. It spends a resource the player
never agreed to spend, like a life cost.

### 21. A variable count is announced, and reaches the effect through the paid-cost record

`Variable` makes `N` a FLOOR rather than an amount, the way `MinX` is a
floor on an announced X, and for the same reason: what a cost's X can be is
bounded by what the payer can actually pay and by nothing else. There is
deliberately no maximum on the declaration — the permanent's own counters
are the ceiling, and the validator enforces it.

The count the activator announces is the payment itself, so it rides the
payment fields rather than a second X slot. It reaches the effect through
`StackItem.Paid.CountersRemoved` — **the same record [ADR 0068](0068-the-mana-spent-on-a-spell.md)
introduces for the mana half**, designed once and landing in the same PR.
An activated ability reads `ctx.CountersRemoved()`; a mana ability has no
stack item (CR 605.3b), so its record is handed to
`ManaAbilityShape.ProducedForPaid`, which is how "Add {C} for each storage
counter removed this way" knows how many came off. That callback wins over
`ProducedFunc`, which wins over `Produced`; CR 106.7's "could produce"
reader evaluates it with the largest payment the source could make right
now, because "could" is about the possible.

### 22. One payment shape on the wire, validated as a set

`counter_source_ids` (already a list since #625) is joined by
`counter_counts`, the per-permanent split, and the existing `counter_kind`:

```
self / other, printed count   counter_source_ids (or nothing)
among                         counter_source_ids + counter_counts, totalling N
variable                      counter_counts, at or above the floor
any kind                      + counter_kind
```

A fixed single-permanent payment still sends exactly what a #625 client
sends, so nothing that already speaks this payload has to change.

**#943 adds the sixth line, and one field.** An among cost whose kind is
EMPTY (Tekuthal, Inquiry Dominus' "Remove three counters from among other
artifacts, creatures, and planeswalkers you control") is paid in whatever
kinds the permanents hold, so the kind question is asked once per part:

```
any-kind among                counter_source_ids + counter_counts
                              + counter_kinds, one per permanent
```

`counter_kinds` is an EXTENSION of the triple, not a rival encoding of it:
it is sent only when the parts actually differ in kind, and a payment of
one kind still says so in `counter_kind` — byte-for-byte what a #625 client
sends. A `counter_picks: [{id, kind, count}]` list was the alternative and
was rejected for that reason: it is a second shape for every payment, which
means every existing client, every existing server test and the enumerator
all have to learn it to buy nothing for the five shapes that were already
right.

The set's identity becomes **(permanent, KIND)** rather than the permanent:
one creature carrying a +1/+1 and a shield counter may pay with both, which
is two parts of one payment. For a printed kind the two keys are the same
thing, so a fixed-kind among payment that names one permanent twice is
refused exactly as before.

Nothing new is projected for it. A `counter_cost_options` row has always
been a (permanent, kind) pair, so the option list already answers "which
kinds, off which permanent", and the client's many-pick modal — a stepper
per row — is already the kind selector. The enumerator's one-payment rule
is unchanged too: drain the fullest first, now across kinds as well as
permanents.

The among payment is validated as a SET, the crew shape: every permanent
distinct, controlled by the activator, matched by the clause without the
targeting gate (CR 601.2h), holding at least the count named against it, and
the counts totalling exactly N. Any failure refuses the whole activation
with no counter removed, per §3.

The enumerator offers ONE among payment (drain the fullest permanents
first) and ONE variable payment (every counter the permanent holds), for the
reason `crewPayment` and `sacrificePayments` offer one: the sets differ only
in which permanents are drained, and a policy has nothing to choose between
them with. `MoveCost.Counters` gains an entry per permanent drained, so a
bot sees what the payment costs it and not only what the card charges.

### 23. Adding a counter is its own small type, and CR 118.3 is a predicate

`AbilityCost.AddCounter` is a `CounterAddCost{Counter, N}`, always on the
source: no printed card pays a cost by putting a counter on something else,
and inventing the clause would mean inventing a picker for it.

Two rules ride with it, both already this file's:

- **Not replaceable.** `applyCounterLocked`, not `AddCounterForEffect`.
  Paying a cost is not an effect (CR 121.1), so Doubling Season does NOT
  double Devoted Druid's -1/-1 — which would double the price of a card
  that is meant to be pure upside.
- **Ordering.** Paid at announce with everything else (CR 118.3 / 602.2b),
  after the removal half and before the sacrifices.

**The "can't have counters" refusal is `canPlaceCounterLocked`**, one
predicate the validator, the legal enumerator and the protocol view all
read, so a greyed row, a skipped move and a refused activation are the same
answer. Today it answers no for exactly one reason — the permanent is not on
the battlefield under the payer's control — because the engine models no
Solemnity-class prohibition yet. That is stated rather than hidden: this
predicate is the one place such a static plugs in, and the refusal it drives
is already wired end to end.

### Still out of scope

- ~~**An any-kind removal spread across permanents.**~~ **Closed by #943**
  (§22 above): `Counter == "" && Among` is a shape, the kind is named per
  permanent on `counter_kinds`, and `effects.Register`'s refusal is gone.
  Tekuthal, Inquiry Dominus ships with it — with one caveat that is not
  this seam: "if you would proliferate, proliferate twice instead", a
  replacement of a keyword action (CR 701.34) the engine has no seam for.
  Its `{U/P}` symbols are payable either way since #971.
- **A counter cost as an ADDITIONAL cost to cast a spell.** The component
  lives on `AbilityCost`; `AdditionalCost` has its own shape.
- **A prohibition to refuse against.** §23.

## Addendum (2026-09-22): exhaust, and the activation tally (#1181)

**Status:** Accepted · 2026-09-22 · tracked on
[#1181](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1181). This
status covers this section only; every decision above stays accepted
and unchanged.

### Context

Exhaust is an activated-ability modifier printed on 41 cards across
`dft`, `tla`, `tle`, `fra`, `ytdm` and `yecl`:

> **Exhaust — {4}: Earthbend 4.** *(Activate each exhaust ability only
> once.)*

The engine had a shape for the ability and none for the "only once".
Two things already on `Game` look as though they should answer it and
neither does:

- `ActivatedAbilityShape.Condition` (the #743 addendum above) is the
  right GATE and the wrong question. A condition is a predicate over
  the board; "have I already done this" is a fact about the past, and
  nothing recorded it.
- `TurnTally` (#586, `game/turn_tally.go`) counts what a source's
  abilities **resolved** or **triggered** this turn. Both halves are
  wrong for exhaust. An exhaust ability countered on the stack is
  still spent, so the count has to be of ANNOUNCEMENTS; and exhaust
  never refreshes, while `TurnTally` is emptied on every turn advance
  by construction.

The seam doc's [Per-source activations-this-turn
count](../engine-seams.md) row (Quirion Ranger, Wirewood Symbiote,
boast) wants the same missing number in a different scope. That is why
this is one record and not two.

### Decision 1: one record, `Game.Activations`, with two scopes

`game/activation_tally.go`:

```go
type ActivationTally struct {
    Ever map[string]int `json:"ever,omitempty"` // whole game, never reset
    Turn map[string]int `json:"turn,omitempty"` // emptied on the turn advance
}
```

Both maps are keyed by `ObjectTallyKey(source, Card.ObjectEpoch,
label)` — the key `ResolvedThisTurn` and `TriggeredThisTurn` already
use. `Ever` is what exhaust reads (`ActivatedThisGame`); `Turn` is
what "Activate only once each turn" and boast will read
(`ActivatedThisTurn`) when their cards are written, and it is here now
because the two are one write at one call site.

Written in ONE place: `ActivateCatalogAbility`, beside
`notePlayerActivationLocked`, at the announce. Read in three, all
through one function — `Game.AbilityExhausted(source, shape)`.

**Mana abilities are deliberately not counted.** They take the other
entry point (`ActivateManaAbility`, CR 605.3a — no stack, no
priority), so `effects.ManaAbility` carries no `Exhaust` field and the
combination is unspellable rather than silently ignored. One printed
card wants it — **Loot, the Pathfinder**'s "Exhaust — {G}, {T}: Add
three mana of any one color" — and is not in the catalog for that
reason.

### Decision 2: the key is the object and the ability's LABEL

Per ABILITY, because a card may print three exhaust abilities (Loot,
the Pathfinder) and each is activatable once; per OBJECT, because
CR 400.7 says a permanent that changed zones is a new object with no
memory of its previous existence.

The label rather than the ability's index: the index is a position in
`ActivatedAbilitiesForCard`'s FILTERED list, and an
[ADR 0071](0071-designations-that-switch-abilities-on.md) designation
switching an ability on renumbers everything behind it. A label is
what the card printed. `effects.Register` refuses an exhaust ability
with a blank label, two exhaust abilities on one card sharing a label,
and a label that prints "Exhaust" without the bit (or the bit without
the word) — all at boot.

Three printed consequences fall out of the key rather than being
coded:

| board | answer | why |
| --- | --- | --- |
| flicker the permanent | its exhausts are available again | exile and return is two epoch bumps |
| phase it out | they would not be | phasing is not a zone change (CR 702.25f) and must not bump the epoch when it is built |
| copy it | the copy has its own | a new instance is a new key (CR 707.2: what a permanent has done is not copiable) |

Phasing is not modelled in this engine. The rule the day it is: **a
phase-out does not bump `Card.ObjectEpoch`.**

Nothing is deleted at the battlefield exit, for the reason
`battlefield_exit.go` already gives about `TurnTally`'s per-ability
counts: the epoch IS the forgetting, so the old object's entries are
unreachable rather than absent.

### Decision 3: the card side is one bit

`ActivatedAbilityShape.Exhaust bool` / `effects.ActivatedAbility.Exhaust`,
carried through `activatedShapes`. No per-card logic, no per-card
condition closure, exactly as `Cycling` is one bit.

It is **not** `Condition` and **not** `ActiveWhen`:

- a `Condition` is CR 602.1b and may be true again tomorrow; the view
  says `condition_unmet` and the client says "activation condition not
  met";
- an `ActiveWhen` designation means the ability is not on the
  permanent at all, so it is absent from `ActivatedAbilitiesForCard`;
- an exhausted ability is still printed and still shown, greyed, with
  a different reason — it is never available again for this object.

A card that prints both sets both. **Bitter Work** ("Exhaust — {4}:
Earthbend 4. Activate only during your turn.") is the card, and the
order in `ActivateCatalogAbility` is timing → exhaust → condition, so
a second activation on your own turn is refused as exhausted rather
than as badly timed.

### Decision 4: one reader, three refusals

`Game.AbilityExhausted` is read by:

1. `ActivateCatalogAbility` — refuses with `ErrAbilityExhausted`,
   after the timing check and before X, targets and every cost, so
   nothing is paid;
2. `legal.activatedMoves` — does not enumerate the move;
3. `protocol.viewOfActivatedAbilities` — stamps `exhausted` on
   `ActivatedAbilityView`, and the client greys the row with
   "already activated (exhaust)".

That is #544's rule expressed as a property of there being one reader
rather than three copies of a rule that agree today.

`ErrAbilityExhausted` is its own error and `exhausted` its own wire
flag, both because the recovery differs: a condition can hold again on
the next turn, an exhaust only if the permanent becomes a new object.

### Consequences

- `Game.Activations` is classified `carried` in the snapshot plan
  (#1020's `snapshot_drift_test.go`), so the round-trip property test
  enforces it. It has to be: "have I used this yet" is game state a
  player can lose a game over, and a restore that dropped it would
  hand every exhaust ability on the board back.
- It rewinds with `Clone` / `RestoreFrom` for the mirror reason — an
  undo that kept the activation would take an ability away for the
  rest of the game on the strength of something that no longer
  happened.
- `Ever` is never flushed, so it grows by one entry per ability ever
  activated. That is the cost of "for the whole game" and it is
  bounded by the length of the game.

### Cards

**Prowcatcher Specialist** (the keyword and nothing else),
**Greenbelt Guardian** (an exhaust ability beside a repeatable one —
the per-ability proof on a printed card), **Bitter Work** (exhaust and
a printed condition, over earthbend) and **Ba Sing Se** (an earthbend
activation that is deliberately NOT exhaust: it prints "Activate only
as a sorcery" and repeats every turn). All four `full`.

### Still out of scope

- **The mana-ability half.** ~~([#1183](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1183))~~ **Closed** — see the note of 2026-09-22 below.
- **Cards that read the record from outside** ([#1184](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1184)). Rangers' Refueler and
  Afterburner Expert ("Whenever you activate an exhaust ability, …")
  want an event or a watch, not this map; Elvish Refueler ("you may
  activate exhaust abilities as though they haven't been activated")
  wants a permission that overrides the gate; Boom Scholar ("Exhaust
  abilities of other permanents you control cost {2} less") wants a
  `CostModifier` that can see the bit. None is built.
- **The per-turn row itself.** `ActivatedThisTurn` exists and has no
  reader: Quirion Ranger, Wirewood Symbiote, Varragoth, Broadside
  Bombardiers and boast are one `Condition` each away and are not in
  this PR.
- **Avatar Kuruk**'s "Exhaust — Waterbend {20}: Take an extra turn
  after this one" still waits on extra turns (#753) and on the
  waterbend cost, neither of which is this seam.

## Note (2026-09-22, #1183): the mana half, and the two write sites

**Status:** Accepted · 2026-09-22 · tracked on
[#1183](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1183). This
note closes the first bullet of "Still out of scope" above; every
decision in the addendum stays accepted and unchanged.

### Context

Decision 1 above says mana abilities are deliberately not counted:
they take the other entry point (`ActivateManaAbility`, CR 605.3a — no
stack, no priority, no announcement to hang a record on), so
`effects.ManaAbility` carried no `Exhaust` field and the combination
was **unspellable** rather than silently ignored. One printed card
wants it, and it is a Commander card people play:

> **Loot, the Pathfinder** — `{2}{G}{U}{R}`, Legendary Creature —
> Beast Noble, 2/4
> Double strike, vigilance, haste
> **Exhaust — {G}, {T}: Add three mana of any one color.**
> Exhaust — {U}, {T}: Draw three cards.
> Exhaust — {R}, {T}: Loot deals 3 damage to any target.

It is also the printed example Decision 2 is shaped for — three
exhaust abilities on one card, each activatable once — so it is the
card most worth reaching.

### Decision 1: the card side is the same bit, twice

`ManaAbilityShape.Exhaust` / `effects.ManaAbility.Exhaust`, carried
through `buildDef`'s mana projection. No new record, no new key, no
per-card logic — the twin of `ActivatedAbilityShape.Exhaust`.

The PREDICATE is one function and the two public readers are thin:

```go
func (g *Game) AbilityExhausted(source uuid.UUID, ab ActivatedAbilityShape) bool
func (g *Game) ManaAbilityExhausted(source uuid.UUID, ab ManaAbilityShape) bool
// both -> g.exhaustedLocked(source, ab.Exhaust, ab.Label)
```

Two entry points rather than one generic one because the two ability
kinds are two structs everywhere else in the engine; what must not
drift is the rule, and the rule is written once.

`effects.Register`'s boot checks now run over **both** ability lists
with **one** `seen` set. That shared set is the point: a mana ability
and an activated ability on the same card write to the same key space
on the same object, so two exhaust abilities that shared a label
across the kinds would share one use, and two per-list sets would not
have caught it.

### Decision 2: TWO write sites, because the auto-tapper spends abilities

The CR 602 path has one write site (`ActivateCatalogAbility`). The
mana path has two, and this is the whole reason #1183 is its own issue
rather than a line in #1181:

1. **`ActivateManaAbility`** — the hand click. The key is taken before
   anything is validated or paid (a Lotus Petal-shaped sacrifice cost
   ends the object and carries `Card.ObjectEpoch` with it), the gate
   is read from it immediately, and the record is written at the point
   every gate has passed and the first payment is about to be made.
   CR 605.3a makes the activation one indivisible step with no
   priority window inside it, so there is no later "announcement
   finished" to hang the write on — and an activation that begins
   paying has happened. That is #1181's "an exhaust ability countered
   on the stack is still spent", spelled for a path with no stack.
2. **`materializePlanLocked`** — the AUTO-TAPPER's executor, which
   taps a permanent and mints its mana **directly** rather than
   routing through `ActivateManaAbility`. A plan that spent an exhaust
   ability without recording it would hand the player the ability
   straight back.

Both writes are **unconditional**, exactly as the CR 602 one is:
`Ever` is what exhaust reads and `Turn` is the per-turn count the
"Activate only once each turn" cards will read, and both are one write
at one call site.

### Decision 3: five readers, and the planner is one of them

`Game.ManaAbilityExhausted` is read by:

1. `ActivateManaAbility` — refuses with `ErrAbilityExhausted` before
   any cost is validated, so a second click taps nothing;
2. `legal.manaMoves` — does not enumerate the move;
3. `protocol`'s `ManaAbilityView` — stamps `exhausted`, under the SAME
   wire name the activated view uses, so the client's row predicate is
   structural and `ABILITY_EXHAUSTED` needed no sibling string;
4. the **auto-tapper**, both halves, through the one picker the
   planner (`gatherTapSources`) and the executor
   (`materializePlanLocked`) share. `autoTapAbilityFor` became a
   method for this: one of its exclusions is now a fact about the game
   rather than about the ability shape. Putting it inside the picker
   rather than beside the sickness and gate checks at the two call
   sites is deliberate — a card whose FIRST mana ability is a spent
   exhaust still auto-taps the second, which a per-source check
   outside the picker would have got wrong;
5. `ProducibleManaLocked` — CR 106.7's "could produce".

### Decision 4: CR 106.7 answers no, and that is a declared narrowing

`producible_mana.go`'s own docblock says costs and timing are not
asked about: a tapped Island still offers `{U}`, a Temple of the False
God its controller cannot activate still offers `{C}`, a false
`Condition` is irrelevant. Read strictly, "Activate each exhaust
ability only once" is an activation restriction of that same family,
so CR 106.7 would still count a spent exhaust ability.

**It answers no anyway**, and the reason is the direction of the
error. The readers of CR 106.7 are Exotic Orchard, Reflecting Pool and
Fellwar Stone, and their answers price casts: a Pool deriving a colour
from a Loot whose exhaust is already gone is a colour the auto-tapper
cannot actually produce, and the executor would tap the Pool for
nothing on the way to a cast it cannot pay. Answering no is one colour
short — the WEAKER-than-printed direction, which is the same direction
the recursion guard in that file is already short in — and it keeps
CR 106.7 and the auto-tapper saying the same thing about the same
permanent.

It is the only place in that file where a permanent can be "could
produce nothing" for a reason that is not about its output, and it is
written down there as well as here.

### Consequences

- The activation record now covers **every** activation the engine
  performs. `ActivatedThisTurn` consequently has a mana-side number
  too, which the "Activate only once each turn" seam row can read when
  its cards are written.
- Nothing about the snapshot changes: `Game.Activations` was already
  `carried` (#1020) and already rewound with `Clone` / `RestoreFrom`.
  The mana path writes to the same maps, which the tests re-assert on
  this path because it writes its own record and could have got the
  key wrong on its own.
- Nothing on the wire moves in a breaking direction: `exhausted` is
  one more `omitempty` flag on `ManaAbilityView`, absent for every
  mana ability but a spent exhaust one.

### Cards

**Loot, the Pathfinder**, `full`. Its second and third abilities are
ordinary activated abilities on existing primitives (draw three; three
damage to any target), and "Add three mana of any one color" is
`OneColorOfAmount(3)` — ONE colour pick that adds three tokens (#742),
not three independent picks.

The auto-tapper never plans Loot's mana ability, and that is **not** a
simplification of exhaust: a mana ability with a MANA component in its
cost (`{G}` here) is excluded from planning outright, because the
planner would have to solve a second cost to fund the first
(`autoTapAbilityFor`, unchanged since S32). The player floats the `{G}`
and clicks, which is how the card is played on paper. The exhaust
refusal sits in that same picker, so the day a planner learns to fund
a mana cost, a spent Loot is already refused.

### Still out of scope

Unchanged from the addendum above: the cards that read the record from
OUTSIDE (#1184) and Avatar Kuruk's extra turn (#753).
