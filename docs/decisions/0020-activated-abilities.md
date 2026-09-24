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
  A counter-doubling replacement applies only to a counter placed by
  an effect (CR 614.16), so Doubling Season does NOT
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
entry point (`ActivateManaAbility`, CR 605.3b — no stack, no
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
| phase it out | they are not | phasing is not a zone change (CR 702.26d) and does not bump the epoch — built in #1199, [ADR 0084](0084-phasing.md) |
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
  after this one" still waits on extra turns (#753), which is not this
  seam. (The waterbend cost it also waited on landed in #1310 — see the
  2026-09-23 waterbend amendment at the foot of this ADR.)

## Note (2026-09-22, #1183): the mana half, and the two write sites

**Status:** Accepted · 2026-09-22 · tracked on
[#1183](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1183). This
note closes the first bullet of "Still out of scope" above; every
decision in the addendum stays accepted and unchanged.

### Context

Decision 1 above says mana abilities are deliberately not counted:
they take the other entry point (`ActivateManaAbility`, CR 605.3b — no
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
   CR 605.3b makes the activation one indivisible step with no
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

## Note (2026-09-22, #1184): the three seams that read the record from outside

The addendum above built the record and one reader, and listed the
cards that read it from OUTSIDE the ability that owns it as still out
of scope. They are in scope now, and the point of writing them down
together is that they were never one gap: **three different
mechanisms wear one keyword.**

### 1. The activation event (Rangers' Refueler, Afterburner Expert)

*"Whenever you activate an exhaust ability, …"* is a trigger over an
ANNOUNCEMENT, and before this there was no announcement to watch.
`ActivateCatalogAbility` emitted `EventTrigger` — the "an item reached
`PendingTriggers`" breadcrumb, shared with triggered abilities — and
`triggerHarvester.OnEvent` returns immediately on that kind, by
design: a trigger that fires further triggers does so at resolution,
so re-entering the harvest at announce would only spam. Even if it had
not, the event carried no ability identity at all.

`EventActivateAbility` is that announcement said in a kind anything
may watch. It carries:

- `Label` — the ability's printed label, which is the same string the
  activation record is keyed by, so an event and a record entry name
  the same thing;
- `Exhaust` — the keyword bit, read off the shape at the announce.

The bit is on the EVENT rather than looked up by each watcher, and
that is not convenience. By the time a watcher runs, a `SacrificeSelf`
or `DiscardSelf` cost may have ended the object, and an ADR 0071
designation gate may have renumbered or removed the ability; the
announcement is the only moment the fact is reliably knowable.

`EventManaAbilityActivated` gained the same two stamps, at both of its
write sites (the click and the auto-tapper's executor), because
CR 605.1a makes a mana ability an activated ability and Loot, the
Pathfinder prints an exhaust one. Two kinds and not one, because the
two paths differ in what a watcher may assume: a mana ability used no
stack and granted nobody priority.

**Deliberately wider than the two cards.** The open *"Whenever an
opponent activates an ability"* row on
[docs/engine-seams.md](../engine-seams.md) (Harsh Mentor, Runic
Armasaur) is the same event with `ByAnOpponent` in place of `ByYou`
and no exhaust test. That row closes on this shape rather than on a
second one, and the seam doc now says so.

### 2. The permission (Elvish Refueler)

*"During your turn, as long as you haven't activated an exhaust
ability this turn, you may activate exhaust abilities as though they
haven't been activated."*

This is why `Game.AbilityExhausted` and `Game.ManaAbilityExhausted`
now take the **asking player**. A permission is not a fact about the
object: the record still says the ability was activated, and every
opponent still reads it that way. So the question stopped being "is
this spent" and became "is this spent FOR YOU", and every reader had
to start naming the asker — the activation path the activator,
`internal/legal` the seat it is enumerating for, the view the
controller whose menu it is stamping, the auto-tapper the tapping
player, `ProducibleManaLocked` the permanent's controller. That is the
same set the addendum above pointed at one reader so they could not
disagree; they still read one reader, it takes one more argument.

**Not an entry point that clears the record**, which was the obvious
alternative and is wrong twice over. The permission is continuous
while its conditions hold, so there is no moment to run a clear AT;
and a clear would be visible to every player and would survive the
Refueler dying, which "as though" never does (CR 609.4 — an effect
that lets you do something as though a rule were different changes
nothing else).

Keeping the record honest is also what makes the card self-limiting
with no code: activating under the permission WRITES the record a
second time, and the second activation is itself an exhaust ability
activated this turn, so the printed condition goes false on its own.
One extra activation per turn, on your turn, is the whole card.

`Game.ExhaustAbilitiesActivatedThisTurn(player)` is the counter that
condition reads, and it is an `EventsThisTurn` scan rather than a
`PlayerTurnTally` field: this is a FILTERED question (whose
activation, and was it an exhaust one) that no counter carries, and
`Game.Activations` is keyed by object rather than by player so it
cannot answer "you". It runs only when a permission is already on the
battlefield, because the gate consults the record first and the
permission only if the record said "spent".

The card side is `Spec.ExhaustPermissions` + `game.ExhaustPermission`,
a battlefield static with the ADR 0071 designation gate and the
CR 613.1f ability-removal key, read through one accessor
(`ExhaustPermissionsForCard`). `effects.Register` panics at boot on a
nil `Applies` — it would grant the permission to every player at every
moment, which no card prints — and on a blank `Label`.

### 3. The priced activation (Boom Scholar)

*"Exhaust abilities of other permanents you control cost {2} less to
activate."* The CR 601.2f pass already existed with its ordering, its
generic floor and its negative-amount refusal; what it could not do
was look at the ABILITY. `CostQuery` already carried the SOURCE
permanent, so "of other permanents you control" was expressible.

Two fields close it:

- `CostQuery.Ability` (`AbilityCostSubject`: the label, the exhaust
  bit, and a `Mana` flag that is always false today so a predicate
  written now says which kind it means);
- `CostModifier.Activations`, which **partitions** the board's
  modifiers into the ones that price casts and the ones that price
  activations.

The partition is not bookkeeping. Sphere of Resistance's `AppliesTo`
is nil, meaning "every spell"; without the partition every `{T}`
ability in the game would have started costing `{1}` more the day
`CostQuery.Ability` appeared. A modifier prices casts or activations
and never both, because every printed clause in either family says
which it means. The self-modifier slot is skipped entirely for an
activation query: "this SPELL costs {1} less to cast" (CR 113.6d) is
about the card as a spell on the stack, not about a permanent's
ability.

`Game.AbilityManaCostForEffect` is the one function three readers
share: the payment path pays it, `internal/legal` checks affordability
against it, and a test reads it. That is #544's rule with the sign
reversed — an enumerator pricing at the printed cost would silently
HIDE legal moves rather than offer illegal ones.
`payAbilityManaCostLocked` therefore takes a `ParsedCost` now; the
attack tax (CR 508.1a is a cost to attack) still parses its own string
and stays unpriced. **The special-action path is priced too, since
#1319** — CR 116.2 is not an activation, so it does not share
`CostQuery.Ability`'s door, but it wanted the identical third one; see
[ADR 0062](0062-abilities-and-special-actions-from-the-hand.md)'s
2026-09-23 note.

### Cards

**Rangers' Refueler**, **Afterburner Expert** and **Elvish Refueler**
ship `full`. Rangers' Refueler's animation prints no duration
(CR 611.2a), so it is `BecomeArtifactCreature` —
`BecomeCreatureUntilEOT`'s body with `IndefiniteDuration`, factored
out rather than flagged — and it is not the crew ability beside it.
Afterburner Expert's trigger is `InGraveyard`, which REPLACES the zone
list (#925): a battlefield copy of "return this card from your
graveyard" would have nothing to return, and "you" there is the OWNER
(CR 108.4).

**Boom Scholar** ships `caveats`, two of them: the ability's row in
the menu still lists the printed cost (the engine charges the
discounted one — the wire's `ActivatedAbilityView.mana_cost` is the
printed string and nothing renders a `ParsedCost` back), and mana
abilities are not priced through the activation pass.

### Still out of scope

Avatar Kuruk's extra turn (#753), unchanged. Two new, both narrow:
the ability view's printed-cost display, and running the CR 601.2f
pass over a MANA ability's cost.

---

## Addendum (2026-09-22): abilities that function from the graveyard and exile (#1221)

**Status:** Accepted · 2026-09-22 · tracked on
[#1221](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1221). This
status covers this section only; every decision above stays accepted
and unchanged.

### Context

CR 113.6 is one rule and the engine had answered a quarter of it.
[ADR 0062](0062-abilities-and-special-actions-from-the-hand.md)
Decision 1 gave `ActivatedAbilityShape.Zones []ZoneKind` to the ONE
activation path — nil means the battlefield, `{ZoneHand}` is cycling —
and #922 gave `TriggeredAbility` the same field. What was missing was
not a shape. It was **every consumer past the activation path**:

- `internal/legal`'s enumerator walked the battlefield and the seat's
  own HAND and stopped there, so a bot could never unearth.
- `protocol/view.go` stamped `hand_abilities` on a hand card and
  nothing anywhere else, so a human could never see the row.
- `AbilityCost` had no component for "Exile this card from your
  graveyard", which is the cost three of the four graveyard keywords
  print.

The seam doc's ["Ability activatable from a non-battlefield
zone"](../engine-seams.md) row says the same thing from the other
side: four cards were "one `Spec` edit away", and had been since #660,
because nothing would have offered the edit to anybody.

The keyword family behind the row is what settles the design, because
it is four keywords' worth of variation over one dimension:

| keyword | CR | zone | cost | effect |
|---|---|---|---|---|
| cycling | 702.29a | hand | discard this | draw |
| unearth | 702.82a | graveyard | mana only | return it, haste, exile it later |
| scavenge | 702.96a | graveyard | **exile this** | counters equal to its power |
| embalm / eternalize | 702.128a / 702.129a | graveyard | **exile this** | a token copy with changes |

Four keywords, one activation path, one zone field. Nothing here is a
new KIND of thing; all of it is the existing kind, one zone further
out.

### Decision 24: the consumers get a zone walk, not a second entry point

`ActivatedAbilitiesForCard` is unchanged and `AbilityFunctionsFromZone`
stays the one predicate. What changes is who asks it, and about how
many piles:

```go
// internal/legal/abilities.go
func (e *enumerator) abilityZones() []abilityZone   // hand, graveyard, command, exile

// internal/protocol/view.go
func stampZoneAbilities(g, seats, exile)            // the same four
```

The enumerator's walk is deliberately shaped like `castZones`
(#1014's five-zone cast walk) and deliberately one pile shorter. The
**library is not walked**, and that is a rule rather than an omission:
CR 401.2 makes a library hidden, no printed ability functions from
one, and an enumerator that read it would be touching cards the seat
is not entitled to see to answer a question whose answer is always
"nothing". The one line it would take is named in the comment, so the
day a card prints such an ability the change is a line and not a
rediscovery.

Exile is the shared pile and gets the CR 108.4 treatment in both
walks: the "you" is read off `Card.Owner` rather than off the loop,
because `g.Exile` holds every seat's cards and the per-seat piles hold
only their own. That is the same rule `ActivateCatalogAbility` has
enforced since #660 (`source.Owner != playerID`), asked by the two
consumers that have to agree with it or fall foul of #544 in one
direction or the other.

### Decision 25: `AbilityCost.ExileSelf`, and why it is not `DiscardSelf` with a zone

Scavenge, embalm and eternalize all print the same cost clause:
"Exile this card from your graveyard". It is `DiscardSelf`'s sibling
and it is a **second bool**, not a zone parameter on one "the source
pays itself" component, because the two are different rules:

- discarding is a CR 701.8 keyword action with its own event
  (`EventDiscardCard`), its own cause (`DiscardCauseCost`) and, for
  cycling, `EventCycle` on top;
- exiling as a cost is a plain CR 406 zone change with none of that.

A card file that wanted "discard this from your graveyard" would be
writing a card that does not exist. The part they genuinely share —
"the payment IS the source, so nothing is announced, nothing is
picked and nothing goes on the wire" — they get for free by both
being a bool.

Everything else is the discard component's shape, one zone over
(`game/exile_cost.go`):

- **Validated** with the rest of the cost, before anything is paid:
  the source has to be in a GRAVEYARD, or `ErrActivationZoneNotAllowed`
  with nothing spent. "Your" needs no second check — the activation
  path has already refused a non-owner off the battlefield.
- **Paid last**, beside the discards, because it moves the source and
  invalidates every pointer the payment block held.
- **Through `routeCardToZoneLocked`** like every other exit, with
  `MustSettleNow` set for the reason `DiscardCauseCost` sets it:
  CR 601.2h / 602.2b make activating an ability one indivisible step,
  so the CR 614 window runs over the move and never stops to ask.
- **Refused at BOOT** on an ability that does not declare the
  graveyard (`effects.Register`), exactly as `DiscardSelf` is refused
  off the hand. A component that could never be paid is a card-file
  mistake, and the treatment it gets is the one `MinX`-without-`{X}`
  already gets.

### Decision 26: the effect reads its source back out of exile; no snapshot on the stack item

Scavenge needs "this card's power" and embalm needs the whole card to
copy — and by the time either effect runs, its own COST has moved the
card to exile. `StackItem` carries no last-known-information snapshot
of its source (`SourceCardID` and `SourceEpoch`, and that is all), and
this addendum deliberately does not add one.

It does not need to. `LookupCardForEffect` finds a card in whatever
zone holds it, and a card outside the battlefield has no layers
applied to it (CR 613 runs on permanents), so the printed power read
out of exile at resolution is the same number the graveyard held. A
snapshot field would be a second LKI store beside the damage event's,
for a question that already has a right answer — the same call
`targets.go` declines to make for the same reason.

The cost of that choice is stated rather than hidden: a card somehow
moved OUT of exile between the announce and the resolution (a shuffle
of exile into a library) answers `ok == false` and the ability does
nothing, which is CR 608.2a's "as much as it can" and is the weaker
direction (#259).

### Decision 27: the wire renames `hand_abilities` to `zone_abilities`, and it is scoped rather than merely hidden

Two changes to `CardView`, and the second is the one that matters.

**The rename.** `hand_abilities` was named for the only zone it had.
Now that a graveyard card carries the same rows it is `zone_abilities`
— one field, because the engine has one activation path with a zone
dimension, so the wire has one row list and the client has one reader.
A second `graveyard_abilities` beside it would have been two fields
that can disagree about a card in neither zone.

**The scoping.** Through #660 the field was written to the exported
`CardView` field and got its privacy from the HAND: `FilterViewFor`
blanks another seat's hand wholesale. A graveyard is public, so the
same code would have shipped one seat's answer to the whole table —
which is exactly the surface #1055 and #1167 took off the cast
stamps, arriving again one field over. An ability row carries
`legal_targets` and `clauses`, and hexproof, shroud and "target
opponent" all narrow a target set **by who is asking**.

So the rows ride `castOffers`, the per-seat carrier that already
exists for that question, and `publicIn` drops them: the seat whose
card it is gets the list, every other seat gets the card and no list,
a spectator gets no list. The hand's rows moved there with the
graveyard's rather than leaving two lifecycles for one field.

One wrinkle worth naming, because it is the kind of thing that breaks
silently: `stampLegalTargets` and `stampGrantedPermissions` file a
seat's entry with `stampsFor`, which REPLACES. The ability stamp
therefore merges (`stampZoneAbilitiesFor`) and runs after both — the
two passes answer different questions about the same card, and a
plain `stampsFor` here would have blanked a flashback card's announce
surface the moment the card also printed a graveyard ability.

`ActivatedAbilityView.exile_self` joins `discard_self` as the
advisory bit for the new cost component. Neither has a renderer: the
keyword's own label spells the clause out, and the field is there so
a client that wants to mark the row need not parse the label.

### Decision 28: the client offers the row where a player actually looks at a graveyard

`ZoneBrowserModal` is the only place anyone inspects a graveyard or
exile, and it rendered `<Card>` with no `onActivateAbility` — so the
pop-over every battlefield permanent and every hand card already has
was suppressed for every browsed card. It now wires the same callback
`Hand.svelte` wires, gated on the card carrying rows at all, which
for a bystander is never (the server stamped none).

Like the impulse-cast button (#874) it closes the browser and hands
the choice up to `Board`: an activation can open a target picker, an
X prompt or a mode picker, and those are Board's chain, not the
modal's. The viewer's CR 307.1 window is passed down so a
sorcery-speed row greys with a reason instead of being clickable and
refused — every keyword on this surface prints "only as a sorcery",
so without it the row would be wrong more often than right.

### Cards

**Dregscape Zombie** (unearth), **Deadbridge Goliath** (scavenge) and
**Sacred Cat** (embalm) are each the keyword and almost nothing else,
which is the point: a test that sees five counters or a Zombie Cat
token is seeing the keyword's own arithmetic rather than a card's.
Dregscape Zombie ships `caveats` for the bounce hole below; the other
two ship `full`.

`ExileInsteadOfLeavingBattlefield` is factored out of Whip of Erebos
(#296) and shared with unearth, because the clause is identical, its
failure modes are silent (a missing `NewZone != ZoneExile` guard is
an infinite loop; a missed event kind is a creature that can be
reanimated twice), and two inline copies would have drifted. Its two
declared limitations are inherited whole: the redirect is turn-scoped,
and a BOUNCE bypasses it because `BounceToHandForEffect` moves a card
without running the CR 614 pipeline.

### Still out of scope

- **Ninjutsu** (CR 702.49). It is a HAND activation, so the zone
  dimension already reaches it, but its two other halves do not
  exist: a cost component that returns an unblocked attacker you
  control, and an entry that puts a card onto the battlefield
  **attacking** (`ZoneEntryOptions` has `Tapped` and no `Attacking`;
  only the token path can do it, `entry_choice.go`'s minted-token
  branch). Filed separately. *(Delivered by the 2026-09-23 amendment
  below, #1227.)*
- **Statics that function from a graveyard** — `StaticAbility.Zones`,
  the layer pass's own half of CR 113.6c (Anger, Wonder, Brawn). The
  other row of the same seam issue, and the next PR.
- **A per-instance grant of an exile ability** (Greater Gargadon while
  suspended). ADR 0062 open question 2, unchanged: the shape is
  per-DECLARATION, and a grant over one suspended card is a different
  object.
- **Cost modification for activated abilities**, unchanged from the
  #1181 addendum.

---

## Amendment (2026-09-23, [#1227](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1227)): ninjutsu, a hand activation whose cost reads combat

The #1221 addendum above listed **Ninjutsu** under "Still out of scope" with
its two missing halves named. Both exist now —
[ADR 0045](0045-combat-restrictions.md)'s 2026-09-23 amendment has the combat
half (the attacking entry, and an exported "unblocked attacker") and
[ADR 0073](0073-optional-additional-costs-and-the-cast-gate.md)'s has the
paid-cost half — and what is left for this ADR is the part that turned out to
be nothing at all.

### Decision 29: the keyword is a constructor, and the activation path is untouched

`effects.Ninjutsu(cost)` is `effects.Cycling(cost)`'s shape with a different
cost component and a different verb:

```go
Label:  "Ninjutsu {1}{U} (…)",
Cost:   Plus(ManaCost("{1}{U}"), ReturnAnUnblockedAttacker()),
Zones:  []game.ZoneKind{game.ZoneHand},
Effect: ninjutsuEnter,
```

Not one line of `ActivateCatalogAbility` changed. The zone dimension reached
the hand in #660 (Decision 1 of [ADR 0062](0062-abilities-and-special-actions-from-the-hand.md)),
the return component landed in #1213, and the boot-time guard that refuses a
cost naming the source as a permanent on a non-battlefield ability
(`AbilityNeedsPermanentSource`) already admits `ReturnToHand`, because the
permanent it returns is not the source. That is the zone dimension paying off:
the keyword that most obviously wanted a fork needed none.

### Decision 30: the timing restriction is the COST, not a `Condition`

Ninjutsu is activatable only from the declare-blockers step onward, while an
unblocked attacker you control exists. That is a real restriction and it is
deliberately NOT expressed as `ActivatedAbilityShape.Condition`.

CR 118.3 already says a player can't begin to activate an ability whose cost
they can't pay, and outside that window no creature is an unblocked attacker —
so the cost IS the restriction, and a `Condition` would be the same sentence
said twice in two places that could drift. The three consumers all get the
right answer from the one clause: the validator refuses the activation
(`ErrIllegalTarget`), `returnCostOptions` ships an empty picker, and the
legal-move enumerator offers no move at all, because `returnPayments` returns
nil when the pool cannot reach the clause's count (#544).

The one thing that gets worse is the greyed ROW, and that is fixed on the
client rather than by bending the ability's shape — see Decision 31.

### Decision 31: one shortfall predicate, both menus

`returnShortfall(options, label)` is factored out of `contextMenu.logic.ts`'s
`abilityBlocked` and is now asked by `ManaAbilityMenu.svelte` too — the
pop-over a HAND card and the zone browser open, which had no return arm at all.

It matters more for ninjutsu than for any earlier card with the component.
Quirion Ranger's row is unpayable when you control no Forest, which is rare;
a ninjutsu row is unpayable for nearly the whole game and payable only inside
one step with an unblocked attacker on the board. A row that never greyed
would have been clickable and refused far more often than it worked.

No other client change, and none was needed: `return_options` / `return_label`
have been on `ActivatedAbilityView` since #1213 and `return_ids` on the
`activate_ability` payload since the same PR, and Board's announce chain
(`afterAbilityDiscardCost` → the return picker → `continueActivation`) never
inspects the card's zone. Zone-ness is decided once, in `abilitiesOf`.

### Cards

**Ninja of the Deep Hours** (the canonical proof — draw a card on connect),
**Ingenious Infiltrator** ("whenever a Ninja you control deals combat damage to
a player", which fires off the keyword's own arrival because the Infiltrator is
a Ninja), **Moonblade Shinobi** (an Illusion token) and **Prosperous Thief**
(the batch form — "one or more Ninja or Rogue creatures you control deal combat
damage", one Treasure per player connected with). All four `full`; all four are
one `Ninjutsu(cost)` entry plus an ordinary combat-damage trigger.

### Still out of scope

- **Commander ninjutsu** (CR 702.49c) — Yuriko, the Tiger's Shadow. The entry
  would come from the COMMAND ZONE as well as the hand, and
  `putOntoBattlefieldFromZoneLocked` is already generic in its source zone, so
  the ENTRY is free. What is not free is the commander bookkeeping around a
  command-zone exit that is not a cast, which nothing in the engine does today;
  it did not fall out, so it is not here. *Shipped by #1278 — see the
  2026-09-24 amendment below.*
- **Ninjutsu on a card whose other half needs machinery.** Fallen Shinobi
  ("you may play those cards without paying their mana costs" over another
  player's exiled cards) and Silent-Blade Oni (cast a spell from an opponent's
  hand) are ordinary catalog work behind other seams, not ninjutsu work.

---

## Addendum (2026-09-23): a MANA ability that functions from the hand (#1228)

The 2026-09-22 addendum above closed the CR 602 half of CR 113.6 and named
what was left: "`ManaAbilityShape` has no zone dimension, so the Spirit Guides
stay blocked." This closes that.

```
Simian Spirit Guide   Exile this card from your hand: Add {R}.
Elvish Spirit Guide   Exile this card from your hand: Add {G}.
```

CR 605.1a makes both of those mana abilities — they could add mana, they are
not loyalty abilities, they target nothing — so they take the OTHER entry
point. `ActivateManaAbility` found its source on the battlefield and nowhere
else, `ManaAbilityShape` had neither the zone field nor an exile-this cost, and
[ADR 0071](0071-designations-that-switch-abilities-on.md) Decision 1 note 3 had
already written the sentence this issue is the answer to: mana abilities do not
get the field, "not a principle — the field plus its accessor is the same two
lines on the day one does".

Six decisions. The auto-tapper is the seventh reader and has an amendment of
its own on [ADR 0011](0011-mana-pool-and-auto-tapper.md), because it is the
consumer with no CR 602 counterpart: nothing plans a cycling activation on the
player's behalf, and a mana source is exactly the thing the planner exists to
find.

### 1. `ManaAbilityShape.Zones`, with the same nil default and the same posture

`ActivatedAbilityShape.Zones`' sibling (#660), `TriggeredAbility.Zones`' (#922)
and `StaticAbility.Zones`' (#1221). Nil means the battlefield and nowhere else,
which is every mana ability the catalog held before this, the synthetic
basic-land ability included.

The posture is the one `AbilityFunctionsFromZone` takes and is worth restating
because it is the half that surprises: **a declared zone is not an ADDITION to
the battlefield.** A Simian Spirit Guide that got cast is a 2/2 Ape with no
abilities, which is the paper card — "from your hand" is the ability, not a
permission bolted onto one. The predicate is
`game.ManaAbilityFunctionsFromZone`, a separate function from the CR 602 one
rather than a generic over both, because the two ability kinds carry their cost
components in different shapes (`ManaAbilityShape` holds them directly,
`ActivatedAbilityShape` holds an `AbilityCost`) and a shared signature would
have to take an interface to hide a four-line loop.

What IS shared is the rule underneath: `ManaAbilityNeedsPermanentSource` is
`AbilityNeedsPermanentSource` with the mana shape's field names mapped onto the
CR 602 cost's, so "a tap cost needs a permanent" is written once and a
component that becomes unpayable off the battlefield becomes unpayable for both
kinds at once.

### 2. Only the HAND is supported, and `effects.Register` refuses the rest at boot

`game.supportedManaAbilityZones` is `{ZoneHand}`, the exact shape
`supportedStaticZones` took for the graveyard in #1221 and for the same reason:
a zone the planner's gather, the enumerator's walk and the view's stamp do not
visit would be a declaration the engine silently ignores. The card would
register, look complete on the catalog page, and never make a mana.

The graveyard, exile and the command zone are each one line in that list plus
one pile in `gatherManaZoneSources` on the day a printed card asks. Nothing
does: a scan of the Scryfall dump finds exactly two cards with a mana ability
that functions off the battlefield, and they are the two above.

### 3. `ManaAbilityCost.ExileSelf` is #1221's clause with a second owner

`AbilityCost.ExileSelf` is scavenge's and embalm's "Exile this card from your
graveyard". This is the same clause, and it is the same `bool` in a second
struct rather than a new component, for the reason `SacrificeOther`,
`RemoveCounters`, `TapOthers` and `DiscardCards` are each one type with two
owners: a component declared twice is a component that can be paid two ways.
`exile_cost.go`'s validator and payer now take the zone and the bit rather than
an `AbilityCost`, and both ability kinds call them.

The ZONE it is validated against comes off the ability's own `Zones`, not off
the component. The clause names "this card"; which pile the card is in is
CR 113.6's business, and duplicating the zone on the cost would be two places
that can disagree. (The CR 602 side keeps its hard-coded graveyard: every card
that prints it there says "from your graveyard", and changing that was not this
issue's to do.)

It pays through the one exit primitive with `MustSettleNow`, so a commander
spent as a Spirit Guide's cost gets its CR 903.9 window and settles without
pausing — CR 601.2h and CR 602.2b make paying a cost one indivisible step, and
a cost may not stop to ask a question. That is the same answer #660 gave for a
commander pitched to a cost discard.

### 4. The zone and the cost imply each other, at boot

`checkManaAbilityZones` panics in both directions:

- a non-battlefield zone with NO exile cost is a free repeatable mana source,
  which is not a card anybody printed;
- an exile cost with no non-battlefield zone has nothing to exile from.

The first is the one that matters. Every other cost component a mana ability
can carry is refused off the battlefield (there is nothing to tap, sacrifice or
put a counter on), so without this rule the only way to declare a hand mana
ability would be to declare a costless one. Stating the implication at boot is
cheaper than discovering it as an infinite mana engine, and the auto-tapper's
picker states it a second time rather than trusting a check two packages away.

### 5. One activation path, and the CR 108.4 "you"

`ActivateManaAbility` finds its source in whatever zone holds it —
`findCardAndZoneLocked`, the same helper `ActivateCatalogAbility` has used
since #660 — and branches exactly where the CR 602 path branches: on the
battlefield the CONTROLLER is "you" and `CanActivateManaAbilities` applies; off
it the OWNER is "you" (CR 108.4) and the layer-6 restriction is not asked,
because Arrest and Cursed Totem restrict a permanent.

The zone check goes after the index lookup and before everything else, so the
ability judged is the one the view and the enumerator published and a refusal
costs nothing. The exile is paid LAST, after the discards, because it moves the
source and invalidates every pointer the payment block holds — and #1212's
`ManaSourceKinds` snapshot is taken before the first payment, for the reason it
already was on the Treasure path: by the time the {R} is minted the card is in
exile. Off the battlefield the snapshot reads printed characteristics, which is
the honest answer, because nothing outside the battlefield has layers.

### 6. The wire gets `zone_mana_abilities`, and `mana_abilities` means the battlefield

`zone_abilities`' twin one ability kind over, riding the same per-seat carrier
(`castOffers`) and dropped by the same `publicIn`. A separate field rather than
a reuse of `mana_abilities`, for the reason `zone_abilities` is separate from
`activated_abilities` **and one reason more**: the two lists take different wire
verbs (`activate_mana_ability` against `activate_ability`), so a client sends
the verb that matches the row it read.

The half that is a behaviour change for existing clients:
**`mana_abilities` is now filtered to the battlefield.** It has always been
stamped on every card in every zone, which was harmless while every mana ability
functioned from the battlefield — a Forest in hand publishing "{T}: Add {G}" is
a true statement about the permanent it would become. It stops being harmless
the moment a card's mana ability does NOT function there, so the exported field
now means what it always said it meant: what does this permanent do. A Forest in
hand is unaffected; a Spirit Guide on the battlefield publishes neither list.

A hand is hidden wholesale, so the per-seat scoping is belt-and-braces today.
It is written that way because `supportedManaAbilityZones` is one entry away
from a public pile, and a public pile publishing one seat's rows is #1055 and
#1167 arriving one field over.

### Shipped on

**Simian Spirit Guide** and **Elvish Spirit Guide**, both `full`. They are the
entire printed family — a scan of the Scryfall dump for "exile this card from
your hand" finds nine other cards, and all nine are CR 602 activated abilities
that GRANT a mana ability to a land (the OTJ "Outlaw" cycle, Emrakul, the
Exigent Doom), which is a different seam. Cadaverous Bloom's "Exile a card from
your hand: Add {B}{B}" is a BATTLEFIELD mana ability whose cost exiles a card
from hand — a component `ManaAbilityCost` still does not have, and the sibling
of the discard clause #1213 added.

### Still out of scope

- **A mana ability from a GRAVEYARD, exile or the command zone.** One line in
  `supportedManaAbilityZones` and one pile in `gatherManaZoneSources`; no
  printed card asks.
- **`ManaAbilityCost.ExileCards`** — "Exile A CARD from your hand" as a cost
  (Cadaverous Bloom). `DiscardCards`' sibling, on the battlefield, and a
  different component from this one for exactly the reason `ExileSelf` is a
  different component from `DiscardSelf`.
- **A mana ability with a `Rider` off the battlefield.** The picker refuses one
  and nothing prints one; it would be a rider on an object that no longer
  exists by the time it ran.
- Everything the #1221 addendum left open and #1227 did not take — the
  per-instance exile grant (Greater Gargadon while suspended) and cost
  modification for activated abilities — is unchanged. Ninjutsu closed in the
  amendment above, which landed while this one was in flight.

---

## Amendment (2026-09-23, [#1310](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1310)): waterbend as an activated ability's cost

**Sprint:** S44 — mana and cost components. Tracker [#887](https://github.com/krakenhavoc/cmd_and_ctrl/issues/887).
Deck tracker [#1306](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1306).
The ward half (#1311) is in [ADR 0073](0073-optional-additional-costs-and-the-cast-gate.md)'s
amendment of the same date.

CR 701.67a (checked against the pinned edition, effective August 7, 2026):
*"Waterbend [cost]" means "Pay [cost]. For each generic mana in that cost, you
may tap an untapped artifact or creature you control rather than pay that
mana."* S22 built it for a SPELL (`Spec.TapCost`, `game.TapPermanentsCost`,
tap_cost.go). Three printed activated abilities waited on the other half —
Aang, Swift Savior's "Waterbend {8}: Transform Aang" (shipped at a flat {8},
weaker than printed), Katara, Water Tribe's Hope's "Waterbend {X}", and Avatar
Kuruk's "Exhaust — Waterbend {20}" — and `AbilityCost` had no waterbend
component. #758's `TapOthers` is not it: that is a fixed count a cost DEMANDS,
this is an optional discount a cost OFFERS.

### Decision 32: `AbilityCost.Waterbend` is the spell's component, and the mana stays in `Mana`

`AbilityCost.Waterbend *game.TapPermanentsCost` — the same struct a spell's
convoke and waterbend use, a third owner rather than a new kind. One reading
differs, deliberately, and it is where the mana lives:

- On a **spell** the waterbend is an additional cost, so `Extra` is ADDED to
  the printed cost.
- On an **ability** the waterbend IS the cost ("Waterbend {8}:" has no other
  mana), so the mana goes in `AbilityCost.Mana` and `Extra` names the part of it
  the taps may cover (CR 701.67b: the waterbend's own generic, never the rest of
  the total).

Keeping the mana in `Mana` is the point. X detection (`DemandsX`, `MinX`), the
CR 601.2f cost-modifier pass (Boom Scholar's exhaust discount reaches Kuruk's
{20}), the view's cost chip and every affordability check already read it, and
none of them needed to learn that waterbend exists. The only new question is
"how many permanents, and which", and it has one answer shared by three readers
(`server/internal/game/waterbend_cost.go`):

- `game.WaterbendBudget(clause, priced, x)` — the clause's generic at the
  announced X, capped by the PRICED cost's generic (a discount that already
  removed a symbol leaves nothing for another tap to pay);
- `game.WaterbendReduced(priced, x, n)` — the priced cost with `n` generic paid
  by tapping, applied AFTER the cost-modifier pass (CR 601.2f before 601.2h, the
  cast path's order);
- `Game.WaterbendOptionsForEffect` — the non-targeting candidate walk the
  validator, the view's picker and the enumerator all read (#544).

`effects.WaterbendCost(cost)` builds both halves; a card file never writes
either by hand. `Plus` SUMS a waterbend's mana with another mana component
instead of letting the later one win — "{1}{U}, Waterbend {2}" owes {1}{U}{2} —
and `Register` refuses, at boot, a clause with no pool, an unparseable cost, or
more generic (or more {X}) than the mana component charges. `Plus` also learned
the `TapOthers` component it had silently dropped since #758.

### Decision 33: the activation path — validate with the rest, subtract after pricing, tap on the stack

`ActivateAbilityParams.WaterbendIDs` (wire `waterbend_ids`, its own field rather
than `tap_ids` because an ability printing both would need to say which tap paid
which). In `ActivateCatalogAbility`:

1. **Validated with every other component**, before anything is paid, through
   tap_cost.go's own validator (controlled, untapped, artifact or creature, named
   once, no more than the budget) plus what only the activation path can see: the
   source when the cost also prints {T} (CR 118.3), and a permanent another
   component of the same cost already spends (crew, tap-another, sacrifice,
   return). The last two are refused rather than sequenced: paper lets you tap
   then sacrifice, no card prints both, and refusing is the weaker direction.
2. **Excluded from the auto-tapper** (`WithAutoTapExclusions`), so a Birds of
   Paradise named to the waterbend cannot also make the {G} for the rest.
3. **Subtracted from the priced mana**, and the remainder paid through the
   ordinary strict / permissive / auto-tap path.
4. **Tapped after the ability is on the stack**, with the tap-another taps, so a
   "becomes tapped" payoff resolves first (§4, CR 603.3b).

Tapping to waterbend is not the {T} symbol (CR 302.6), so there is no
summoning-sickness check, and a source whose cost prints no {T} may pay for its
own ability — Aang, who has flash and is often freshly arrived, may be one of
the eight.

### Decision 34: the view, the enumerator, the client

- **View.** `ActivatedAbilityView.Waterbend` is the `TapCostView` a hand card's
  convoke / waterbend already ships as `tap_cost`, so the client's one picker
  (TapCostModal) serves both. `Max` is the budget at X=0; a Waterbend {X} ships 0
  with `demands_x`, the client's cue to size it from the X it collects first.
- **Enumerator.** `legal/waterbend.go` offers ONE payment, not every subset —
  crew's discipline. Free permanents first (a non-creature artifact that makes
  no mana, which costs the seat nothing), then non-mana creatures, then mana
  sources; the smallest affordable prefix is offered, always taking the free
  ones. For {X} it offers the largest X the taps and the mana reach together,
  never below the printed floor. The heuristic prices each tapped creature as a
  spent blocker, as it prices crew.
- **Client.** Board's announce chain asks after X and the Phyrexian stepper and
  before modes and targets — the position the cast chain asks its own taps in —
  and skips the question when nothing could help.

### Cards

**Aang, Swift Savior** (`full` — the caveat is gone), **Katara, Water Tribe's
Hope** (`full` — ETB Ally, "Waterbend {X}" with `MinX(1)` and
`DuringYourTurn()`, base X/X in layer 7b with the affected set locked at
resolution). **Avatar Kuruk** keeps its extra-turn caveat and nothing else: the
cost is now expressible as `WaterbendCost("{20}")` with `Exhaust: true`, and the
ability stays unregistered only because extra turns (#753) do not exist.

### Still out of scope

- **CR 701.67c** — "whenever a player waterbends". No catalog card asks, and
  neither the cast path nor this one emits an event for it.
- **Waterbend on a MANA ability.** No printed card.
- **An ability whose waterbend is only part of a larger mana cost.** The shape
  is supported (`Plus` sums, the budget is the waterbend's own generic) and
  tested; no printed card uses it yet.

## Amendment (2026-09-23, [#1297](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1297)): "Exile N cards from your graveyard / hand" as an activated ability's cost

**Sprint:** S44 — mana and cost components. Tracker [#887](https://github.com/krakenhavoc/cmd_and_ctrl/issues/887).
The vocabulary half — one component, two owners, one new field — is in
[ADR 0073](0073-optional-additional-costs-and-the-cast-gate.md)'s amendment of
the same date (Decision 7).

#1283 built `game.ExileCost` for a MANA ability — Cadaverous Bloom's "Exile a
card from your hand: Add {B}{B} or {G}{G}" — and said in its own doc comment
that the CR 602 owner would be "a second caller and not a second
implementation". The CR 602 form is far commoner, and nearly all of it reads the
GRAVEYARD, not the hand:

```
Grim Lavamancer    {R}, {T}, Exile two cards from your graveyard: 2 damage to any target.
Moorland Haunt     {W}{U}, {T}, Exile a creature card from your graveyard: a 1/1 flying Spirit.
Tome Shredder      {T}, Exile an instant or sorcery card from your graveyard: a +1/+1 counter.
Mines of Moria     {3}{R}, {T}, Exile three cards from your graveyard: two Treasures.
Holistic Wisdom    {2}, Exile a card from your hand: return a card that shares a type with it.
```

`AbilityCost` had `ExileSelf` (#1221, the SOURCE — scavenge, embalm) and
`DiscardCards` (#660, a discard), and neither is this. Mines of Moria shipped
with the ability left out and a caveat naming the gap, the #259 posture.

### Decision 35: `AbilityCost.ExileCards` is the mana ability's component, with a pile

`AbilityCost.ExileCards *game.ExileCost` — the same struct
`ManaAbilityShape.ExileCards` carries, and the same three functions:
`ExileCostOptionsForEffect` (the candidate walk the view stamps and the
enumerator pays from), `validateExileCardsCostLocked` (exactly N, distinct, in
the pile, matching, never the source, never also a discard) and
`payExileCardsCostLocked` (the one exit primitive with `MustSettleNow`, so a
commander exiled this way still gets CR 903.9 and the CR 602.2b indivisible step
never pauses). What the component learned is WHERE: `ExileCost.From`, the hand
or the graveyard, read through `Zone()` so the zero value stays the hand and
every #1283 declaration means what it meant.

A field and not a second type because the two printed forms are one rule read
against two piles of the activator's own cards: both are plain CR 406 moves,
neither is a keyword action, neither targets (CR 601.2h), and "your" (CR 108.4)
means no other player's pile is ever readable. A second type would have been a
second validator with its own opinion about overlap and the source.

The source is never a legal pick, which is what makes a graveyard ability's
"Exile ANOTHER creature card from your graveyard" (Scrapheap Scrounger) need no
predicate of its own. Cards are constructed with `ExileFromGraveyard(n, label,
match)` / `ExileFromHand(...)`, which compose with `Plus` (and `Plus` learned the
field — a composed "{R}, {T}, Exile two cards" that dropped it would be a
Lavamancer pinging for {R} forever).

### Decision 36: the activation path — validate beside the discard, pay beside it, record what was paid

In `ActivateCatalogAbility`:

1. **Validated** right after the discard component and against it, before
   anything is paid.
2. **Excluded from the auto-tapper**: `AbilityAutoTapExclusions` takes the exile
   ids. For the HAND form this is live — a Simian Spirit Guide named to Holistic
   Wisdom's cost is a mana source (#1228), and without the exclusion the planner
   would exile it for the {2} first and leave the cost paying with a card that
   is already gone. For the graveyard form it is the list being right in advance
   (no card functions as a mana source from a graveyard today).
3. **Paid** after the discards and before the exile-self: it moves cards, never
   the source, so it keeps the discard's slot in the order.
4. **Recorded** on `PaidCost.Exiled` (the ids, in the order named), read by
   `Context.Exiled()`. Every printed reader asks about the card itself — Holistic
   Wisdom's "shares a card type with the card exiled this way", Dread Defiler's
   "the exiled card's power" — and the card is findable in exile by the same
   instance ID, so an id list is the whole answer and a count would be
   `Sacrificed`'s shape answering a question nobody on this component asks.

### Decision 37: the view, the enumerator, the bot, the client

- **View.** `activated_abilities[i]` (and `zone_abilities[i]`) carry
  `exile_cost_n` / `_label` / `_options` / `_zone`, the mana view's fields under
  the same names; the mana view gained `_zone` too. The answer is `exile_ids`,
  never `discard_ids`.
- **Enumerator.** ONE payment, not one move per subset (the discard's and crew's
  discipline), from the engine's own walk minus the discard picks, CHEAPEST FUEL
  FIRST when the policy supplies `Options.OrderCostFuel` — the price escape's
  exiled graveyard is already paid by (#1013), which is the same resource spent
  the same way. The mana owner's arm now calls the same solver.
- **Heuristic.** `activateParams.ExileIDs`, each priced with `fuelValue`, so the
  payment offered first is the one priced cheapest.
- **Client.** Board's announce chain asks after the discard and before the
  return / sacrifice / crew pickers, through `DiscardCostModal` with the verb
  changed and the pile named (`exileCost.ts`, shared with the mana path), and
  skips the question when the pile holds exactly N options.

### Cards

**Mines of Moria** (caveat removed — `full`), **Grim Lavamancer**, **Moorland
Haunt**, **Tome Shredder** and **Holistic Wisdom**, all `full`.

### Still out of scope

- **A variable count** — "Exile X cards from your graveyard" (Necropolis Fiend,
  Taigam, Sidisi's Hand, Ludevic), "one or more" (Corpseweft), and The
  Capitoline Triad's "any number … with total mana value 30 or greater". These
  are an ANNOUNCED number, the #1213 variable-sacrifice question one component
  over, and `ExileCost.N` is a fixed count; `Register` refuses a zero.
- **Craft** (CR 702.167) — "Exile this artifact, Exile a creature you control or
  a creature card from your graveyard" spans the battlefield AND the graveyard in
  one clause, and returns the card transformed; its own seam.
- **A graveyard exile on a MANA ability** (Molt Tender, Titans' Nest, Sunken
  Palace). The component expresses it (`ExileCardsFromGraveyard`) and the view
  ships the zone; no catalog card declares one yet, and the auto-tapper already
  refuses any source with an exile-cards cost.

## Amendment (2026-09-24, [#1278](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1278)): commander ninjutsu, a non-cast exit from the command zone

**Sprint:** S42 — casting from non-hand zones. Tracker
[#885](https://github.com/krakenhavoc/cmd_and_ctrl/issues/885) (the issue's own
tracker; the keyword is also hand-special-action adjacent, #886, and a combat
entry, #880).

The #1227 amendment above left **commander ninjutsu** (CR 702.49c) out with one
sentence: the entry from the command zone would work, but "the commander
bookkeeping around a command-zone exit that is not a cast" was unexamined.
Examined, it is almost entirely things that already do not happen:

> Commander ninjutsu [cost] means "[cost], Return an unblocked attacker you
> control to hand: Put this card onto the battlefield from your hand **or the
> command zone** tapped and attacking."

### Decision 38: the keyword is Ninjutsu with a second zone

`effects.CommanderNinjutsu(cost)` is `Ninjutsu(cost)` with
`Zones: {ZoneHand, ZoneCommand}` and CR 702.49c's reminder text, and shares its
effect body (`ninjutsuEnter`). As with #1227, **not one line of
`ActivateCatalogAbility` changed**: `findCardAndZoneLocked` has scanned the
command zone since #660, `AbilityFunctionsFromZone` is the CR 113.6 gate, and
the CR 108.4 "you" off the command zone is the card's owner — the same
`source.Owner != playerID` arm the hand takes. The timing restriction is still
the COST (Decision 30); nothing about the command zone changes that.

### Decision 39: one more door into the shared entry batch, keeping the ID

`Game.PutFromCommandZoneOntoBattlefieldForEffect(cardID, ZoneEntryOptions)` is
the hand door with `ZoneCommand` for `ZoneHand`: `startEntryBatchLocked` now
admits the command zone as a `BatchEntry.From`, so the arrival runs the one
CR 614 entry window, lands through `landEntryLocked`, stamps the CR 506.3c
attacker, and announces `EventZoneMove` / `EventETB` like every other put. The
card **keeps its instance ID**, as a hand or library card does, and here the
reason is load-bearing: `Player.CommanderCasts` is keyed by it, so a fresh ID
would quietly reset the CR 903.8 tax.

What the door does NOT do is the answer to the issue's question:

- **No commander tax, and nothing added to the next one.** CR 903.8 taxes
  CASTING from the command zone. The tally is bumped by the cast path alone
  (`mutations.go`, after `CastSpell` succeeds); a put never reaches it, and an
  activation's cost is not a spell's cost, so no tax is priced either.
- **No CR 903.9 prompt on the way in.** `commanderZoneReplacement`'s
  `AppliesTo` is destination-only (library, hand, graveyard, exile), so an
  arrival on the battlefield is none of its business. PR #539 moved that window
  into the shared exit primitive (`routeCardToZoneLocked`), and it applies on
  the way OUT: `Card.IsCommander` rides the card through the non-cast exit, so
  a ninjutsu'd Yuriko that is bounced, killed or exiled is offered the command
  zone exactly as a cast one is, and deals commander damage while she is out
  (CR 903.10a).
- **No colour-identity check.** CR 903.4 is deck construction and mana; it
  says nothing about which card may leave the command zone.

The return cost's bounce was already on the shared exit primitive (#1227) with
`MustSettleNow`, so returning a COMMANDER as the ninjutsu cost does not stop
for CR 903.9 — a cost cannot pause (the posture `payLifeAsCostLocked` and the
discard cost take). That is unchanged and not new here.

### Decision 40: the effect checks the OBJECT, not just the zone

With two zones, `ninjutsuEnter`'s old "is it still in its owner's hand"
check stopped being enough. A commander discarded in response to her own hand
ninjutsu takes CR 903.9's command zone — a zone the same ability also names,
with the same instance ID. By CR 400.7 she is a new object and the ability has
lost her. The effect now compares `Card.ObjectEpoch` against
`StackItem.SourceEpoch` (stamped at announce for every catalog activation) and
does nothing on a mismatch, which also guarantees the zone it reads is the zone
the ability was activated from — so plain ninjutsu can never reach the
command-zone arm. This tightens plain ninjutsu too (a ninja discarded and
returned to hand in response no longer enters), which is the printed rule.

### The rest is already built

The enumerator's `abilityZones()` has walked the command zone since #1221, and
the view's `stampZoneAbilities` ships its rows on `zone_abilities`, owner-only.
No wire change. The client's one change is where the command zone's actions
live: `CommandZone.svelte` hands `zone_abilities` to its `Card` the way
`Hand.svelte` does, so right-clicking the commander opens the same popover
(with #1227's shortfall greying) that a hand card and the zone browser open,
and the tile shows an `ability` hint beside `cast` while a row is present.

### Cards

**Yuriko, the Tiger's Shadow** (`full`) — the only non-joke printing of the
keyword. Her trigger is Dark Confidant's flip with the life lost by each
opponent, on Ingenious Infiltrator's per-Ninja combat-damage condition. The two
"Fixed commander ninjutsu" cards (The Multifaceted Phyrexian, Monet) are
playtest / Un printings whose whole point is that the tax DOES apply; neither is
catalogued.

### Still out of scope

- **"Put onto the battlefield blocking"** — unchanged from #1227.
- **A commander returned as a ninjutsu COST** goes to hand without the CR 903.9
  offer, because costs cannot pause (see Decision 39). Weaker for the player
  than printed in the rare case it matters; the same posture every cost exit
  takes. Filed as [#1397](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1397).

## Amendment (2026-09-24, [#1296](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1296)): an ability's own cost clause, and a price that reads the target

**Sprint:** S44 — mana and cost components. Tracker [#887](https://github.com/krakenhavoc/cmd_and_ctrl/issues/887).
Decisions 38–40 are the commander-ninjutsu amendment above (#1399); this
one starts at 41.

### Context

An in-app report: *"Equip costs were not paid when equipping to Vivi Ornitier"*,
with Dragonfire Blade:

```
Dragonfire Blade   Equipped creature gets +2/+2 and has hexproof from monocolored.
                   Equip {4}. This ability costs {1} less to activate for each
                   color of the creature it targets.
```

Two things were wrong, and only one of them was the card.

1. **Nothing was charged.** The client stamps its `gameplay.strictMana` setting
   on `cast_spell` and on nothing else, so every `activate_ability` a player
   clicked reached `payAbilityManaCostLocked` with neither `Strict` nor
   `AutoTap` — the sandbox paper path, which spends what the pool covers and
   otherwise waives the charge (`PaidCost.OnPaper`). The reporter had strict
   mana on (their casts were refused with `insufficient_mana` minutes earlier),
   tapped a land by hand (a raw tap, no mana), and equipped for free. This has
   been true of every activated ability since S21; the bots never saw it because
   `internal/legal` sends `strict` + `auto_tap` on every move.
2. **The discount could not be said.** The card shipped with the full {4} and a
   caveat. #1184 put activations through the CR 601.2f pass, but only for BOARD
   modifiers (`CostModifier.Activations`, Boom Scholar), and a board modifier
   has two problems here: it never sees the target, and it is gathered from the
   battlefield. The second one had already bitten: Takenuma's channel discount
   was written as a board modifier scoped to its own ability, and channel is
   activated from the HAND, so the scan never found it — its caveat said "the
   discount applies when you pay for it", and it did not.

### Decision 41: `ActivatedAbilityShape.CostModifiers` — the ability's own clause

"This ability costs {1} less to activate …" is part of the ability, so it lives
on the ability: `ActivatedAbilityShape.CostModifiers` (card side
`ActivatedAbility.CostModifiers`), the activation twin of a spell's
`SelfCostModifiers` (ADR 0048 addendum). `abilityCostQueryLocked` carries the
list on `AbilityCostSubject` (unexported — a predicate cannot reach another
modifier's hooks), and `activeCostModifiersLocked` binds each one to the
ability's source **with the activator as its controller** (CR 602.2) after the
board's activation-scoped modifiers. Consequences:

- It prices this ability and no other — the slot is the scope, so no
  `q.Card.InstanceID == q.Source.InstanceID` predicate to get wrong.
- It works wherever the ability does (CR 113.6): the channel lands price their
  discount from the hand.
- The CR 601.2f rules are the same pass: increases before reductions, the
  generic floor at zero, the negative-amount refusal. Boseiju with five legends
  still costs {G}.

`effects.Register` refuses the shapes the engine would ignore: a clause on an
ability with no mana component (the pass never runs), a `CostFloor` (no printed
ability sets a floor on its own cost — Power Artifact's "can't reduce … to less
than one mana" is a board clause about other abilities, and is still open),
`SpecialActions`, or a designation gate.

### Decision 42: the price is determined after the targets, with them

CR 602.2b runs activation through 601.2b–i, so the targets (601.2c) are chosen
before the total cost is determined (601.2f). `CostQuery.Targets` already
existed for casts (ADR 0048 addendum §13), shown only to a modifier that sets
`ReadsTargets`. The activation door now fills it:

- `Game.AbilityManaCostForTargetsForEffect(activator, source, zone, ab, targets)`
  is the pricer; `AbilityManaCostForEffect` is it with nil targets.
- `ActivateCatalogAbility` prices with `params.Targets` **after** validating
  them, in the one place it pays (and in the waterbend budget, which is sized
  against the same priced cost).
- A target-reading clause with no target (a query before one is chosen, or a
  target that has left) reads zero — for a reduction, the printed cost, the
  #259 direction.

`effects.CostsLessForTheCardItTargets(label, per)` reads the first card target
("the creature it targets" — every printed clause has one target) and sets
`ReadsTargets`; `ColorsOf` (layer-5 colours) and `CountersOf(kind)` are the two
readers the cards use. `CostsLessIfItTargets` (Price of Fame's constructor)
works in the new slot unchanged.

### Decision 43: the enumerator and the view price per target

A nil-targets price is not the price of a target-reading ability, and for a
reduction it is not even a usable gate — it is higher than the real one, so it
would hide legal moves (#544 with the sign reversed). The cast path's §14 rule,
one path over:

- `Game.AbilityPriceReadsTargetsForEffect(ab)` is true when the ability's own
  clause, or any activation-scoped board modifier, reads targets (it ignores
  `AppliesTo`, erring toward true).
- `internal/legal` solves the mana half (`abilityManaPayment`, split out of
  `abilityMovesForSource`) once up front when nothing reads targets, and
  otherwise once per announcement, skipping an unaffordable set before any
  budget is spent on it.
- The view keeps `charged_mana_cost` as the no-target price and adds
  `activated_abilities[i].target_charged_mana_costs` — one price per legal
  target, from the same pricer — when the price reads the target and the
  ability is one clause, one pick, no modes. The client shows the range in the
  menu row ("{2}–{4} depending on the target") and each price with its targets
  in the targeting banner, because an equip's single click is also its confirm.

### Decision 44: the client charges activations when the player enforces mana

`client/src/lib/manaEnforcement.ts` stamps a catalog `activate_ability`
(`ability_index` present) with `strict: true, auto_tap: true` when
`gameplay.strictMana` is on, and leaves it untouched when it is off. `auto_tap`
as well as `strict`, because a cast has the "Auto-tap & cast" override toast to
fall back on and an activation has no retry path; the engine taps for the
shortfall exactly as it does for every bot activation, special action and
attack tax, and refuses only when the board cannot pay — with "insufficient
mana to activate that ability" and no `card_id`, so the cast-only override is
never offered for it. Strict mana off keeps the paper posture it always had.

### Cards

**Dragonfire Blade** (equip discount ships; the hexproof-from-monocolored
caveat stays), **Ghostfire Blade** and **Warrior's Blades** (new, `full`),
and **Takenuma, Abandoned Mire**, **Otawara, Soaring City** and **Boseiju, Who
Endures** (`full`, via `effects.ChannelDiscountPerLegendaryCreature`).

### Still out of scope

- **Board reductions with a per-reduction floor** — Training Grounds, Power
  Artifact ("can't reduce the mana in that cost to less than one mana"). A
  `CostFloor` is a total-mana minimum (Trinisphere), not this.
- **Belt of Giant Strength** ("{X} less, where X is the power of the creature
  it targets") is one `CostsLessForTheCardItTargets` away and was not added;
  the non-target members of the family (Arm-Mounted Anchor, Crown of Gondor,
  Plate Armor, Mirror of Galadriel, …) are ordinary card work now.
- **Per-target prices for a multi-target or modal ability.** No printed card
  prices by target on one; the view omits the map rather than guess.
- **The auto-tap preview's ability branch** (`/games/:id/auto-tap-preview?ability=`)
  still prices the printed cost with no modifiers at all — stale since #1184,
  filed as [#1405](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1405).
