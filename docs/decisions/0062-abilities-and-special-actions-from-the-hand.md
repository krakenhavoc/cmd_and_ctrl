# ADR 0062 — Abilities and special actions from the hand

**Status:** Accepted · 2026-09-18 · S43 — Hand special actions and face-down objects
**Issue:** [#655](https://github.com/krakenhavoc/cmd_and_ctrl/issues/655)
**Numbering:** on 2026-09-18, after `git fetch origin`, every ref was
scanned for ADR files with `git log --all --format= --name-only
--diff-filter=A -- docs/decisions/`, which covers every remote branch and
every commit reachable from one. The highest number present anywhere is
**0060** (`0060-leaving-the-game.md`, on `develop`). **0061** is being
taken at the same time by the replaceable-events work, so this ADR takes
0062. 0005, 0024, 0029 and 0030 stay permanently unused per AGENTS.md §4.
**Builds on:** [ADR 0020](0020-activated-abilities.md) (CR 602 activated
abilities, `AbilityCost`, and its #743 activation-condition and #747
sacrifice-count addenda), [ADR 0021](0021-additional-costs.md) (the
discard-a-card additional cost to cast, CR 601.2h),
[ADR 0025](0025-alternative-costs.md) and the S29 `CastableZones` work in
`server/internal/game/cast_zones.go` (which zone a card may be cast
FROM), [ADR 0048](0048-cost-modification.md) (cost modification, #873),
[ADR 0018](0018-triggers-on-the-stack.md) (triggers and the harvester),
[ADR 0037](0037-unimplemented-card-signal.md) (the keyword table and the
coverage signal), and #799 / #853 (`server/internal/game/discard.go`, the
ONE discard path and its `DiscardCauseCost`).
**Implements:** cycling and typecycling (CR 702.29) on
[#660](https://github.com/krakenhavoc/cmd_and_ctrl/issues/660).
**Designs, does not implement:** foretell
[#658](https://github.com/krakenhavoc/cmd_and_ctrl/issues/658), suspend
[#659](https://github.com/krakenhavoc/cmd_and_ctrl/issues/659).
**Tracker:** [#886](https://github.com/krakenhavoc/cmd_and_ctrl/issues/886).
**Seams:** closes `docs/engine-seams.md` "Discard-a-card cost on an
activated ability", and the hand half of "Ability activatable from a
non-battlefield zone".

Line references are to `origin/develop` at `2a0a46bd`.

## Context

Three keywords act on a card that is **in a player's hand**, and the
engine has no path for either kind of hand action:

| keyword | what it is | rule |
|---|---|---|
| Cycling | an **activated ability** that functions only in hand: "[Cost], Discard this card: Draw a card" | CR 702.29a |
| Foretell | a **special action**: pay {2}, exile the card face down, any time you have priority during your turn | CR 702.143a–b, CR 116.2h |
| Suspend | a **special action**: exile the card with N time counters, whenever you could begin to cast it | CR 702.62a, CR 116.2f |

They are not the same kind of thing, and the most expensive mistake
available here is to model them as if they were. The #94 body called
cycling "an alternative cast path", and five card files in the catalog
repeat it. It is not one: nothing is cast, the card goes to the
graveyard rather than the stack, and the draw happens because an ABILITY
resolved off the stack. An alternative cost is the wrong shape in every
particular.

What is missing on `develop` at `2a0a46bd`:

- **Activated abilities work only on the battlefield.**
  `ActivateCatalogAbility` looks its source up with `findBattlefieldCard`
  (`server/internal/game/activated.go:391`). The bot enumerator loops
  over `g.Battlefield.Cards` (`server/internal/legal/abilities.go:43`).
  The view projects abilities for battlefield cards only
  (`stampActivatedAbilities`, `server/internal/protocol/view.go:1834`,
  called once at `view.go:1502` with `&view.Battlefield`).
- **`AbilityCost` cannot discard.** It carries tap, sacrifice-self,
  sacrifice-other, mana, life, loyalty, crew, remove-counters and an X
  floor (`activated.go:35-168`). `AdditionalCost.DiscardCards`
  (`additional_cost.go:33`) exists, but it is a cost to CAST A SPELL and
  only the cast path reads it.
- **No special-action verb.** `server/internal/actions/actions.go:26-104`
  has casts, activations, and sandbox verbs. Nothing there is CR 116.2.
- **Nothing tells the client a hand card has an ability.** A hand
  `CardView` carries `alternative_costs` and `exile_play`; a battlefield
  one carries `activated_abilities`.

The two halves are independent — a discard cost is just as useful on
Fauna Shaman, which sits on the battlefield — but they arrive together
because cycling needs both, and cycling is what ships on this ADR.

## Decision

### 1. An activated ability declares the zones it functions from

`ActivatedAbilityShape` grows

```go
// Zones is the set of zones this ability functions from (CR 113.6).
// Nil means the battlefield, which is every ability written before
// this field existed.
Zones []ZoneKind
```

mirrored on `effects.ActivatedAbility.Zones`, so cycling is

```go
Activated: []ActivatedAbility{Cycling("{3}")},
```

where the constructor stamps `Zones: []game.ZoneKind{game.ZoneHand}`.

**A set of zones, not a `FromHand bool`.** The three candidates were a
bool, a single `Zone ZoneKind`, and a slice. The slice wins on the next
four cards rather than on this one:

- **Reassembling Skeleton** and **Drownyard Temple** ("{1}{B}: Return
  this card from your graveyard to the battlefield") are the very next
  cards on the same seam row and want `ZoneGraveyard`. A `FromHand bool`
  buys them nothing and would have to be deleted from every card file
  that set it the day one of them is written.
- **Eternal Dragon** prints a hand ability (plainscycling) AND a
  graveyard ability, on one card, as two entries — which a per-entry
  zone handles, and which is why the zone belongs on the ABILITY and not
  on the `Spec`.
- **Greater Gargadon** activates from exile while suspended. That is a
  per-INSTANCE permission rather than a card-level one — the same
  distinction `cast_zones.go` already draws for impulse exile — so it
  will ride a grant on the card instance and be unioned with this
  declaration at the check, not replace it.
- **Two zones on one ability** is a real printed wording ("from your
  hand or your graveyard"), and a slice costs nothing to read.

`ZoneBattlefield` need not be listed and is what nil means, exactly the
way `CastableZones`' nil means hand. The parallel is deliberate: one
dimension for "where may this card be cast from", one for "where does
this ability function from", both per-declaration, both defaulting to
the overwhelmingly common answer.

**Validated at activation like any other zone-scoped rule, on the ONE
path.** `ActivateCatalogAbility` keeps its single body. What changes is
the source lookup:

1. Find the card in whatever zone holds it (`findCardAndZoneLocked`, a
   scan mirroring the existing `findCardByIDLocked`).
2. If the zone is the battlefield, everything CR 602.5 and layer 6
   already do stays exactly as it is: `RecomputeLayersIfStaleLocked`,
   the controller check, `CanActivateAbilities`.
3. Off the battlefield the **owner** is the "you" of the printed text.
   There is no controller for a card in a hand (CR 108.4), and this
   engine's hands hold only their owner's cards. Layers are not
   recomputed and CR 602.5's "its activated abilities can't be
   activated" is not consulted: Arrest applies to a permanent.
4. The ability's `Zones` must contain the zone the source is actually
   in, or the activation is refused with `ErrActivationZoneNotAllowed`
   before anything else is validated and before anything is paid.

There is **no hand-activation fork** and no second entry point. A
cycling ability activated from the battlefield fails at step 4 for the
same reason and in the same place a battlefield ability activated from
hand does.

**Costs that need a permanent are refused at `Register`, not at
runtime.** `Tap`, `SacrificeSelf`, `Crew` and `Loyalty` all name the
source as a permanent on the battlefield. An ability that declares a
non-battlefield zone and one of those components is a card-file mistake,
and `effects.Register` panics at boot naming the card and the component
— the same treatment `MinX` without an `{X}` gets.

**Split second.** Cycling is an activated ability, so it is barred while
`SplitSecondActive` (CR 702.61b) by the check that is already the first
line of `ActivateCatalogAbility` and already the first line of the
enumerator (`legal/abilities.go:39`). Nothing is added. Special actions
are the opposite case; see Decision 4.

### 2. Two discard cost components on `AbilityCost`

```go
// DiscardSelf discards the SOURCE card as part of the cost —
// cycling's "Discard this card" (CR 702.29a).
DiscardSelf bool

// DiscardCards is "Discard a card" / "Discard a creature card" /
// "Discard N cards" as a cost component.
DiscardCards *DiscardCost
```

with

```go
type DiscardCost struct {
    N     int              // how many; at least one
    Label string           // the clause as printed: "a creature card"
    Match func(Card) bool  // nil matches any card in hand
}
```

**Both, not one.** The seam row names both shapes and they behave
differently at every step. `DiscardSelf` names no card (the source is
the card), needs no chooser, needs no options on the wire, and takes the
source out of the zone the ability was activated from. `DiscardCards`
names N cards the activator chooses out of a hand that may hold more,
and its source stays where it is. Folding the first into the second as
"N = 1 matching only this card" would put a chooser, a prompt and an
options list on the wire for every cycling ability in the game, and
would admit a payload that named some other card.

`Match` is a `func(Card) bool` rather than a `*TargetSpec`:
`TargetSpec`'s predicates and its `specMatchLocked` evaluator are
written against permanents on the battlefield, and a cost does not
target anyway (CR 601.2h). `func(Card) bool` is the shape
`SearchLibrarySpec` already uses for exactly this question, one zone
over.

**One discard helper.** Both components pay through
`discardCardsLocked(..., discardOptions{cause: DiscardCauseCost, source: …})`
— the one discard path from #799, with the cause #856 gave it. Nothing
in this ADR writes a second discard loop, and that buys three things for
free:

- `EventDiscardCard` fires per card, so Marauding Mako, Scrounging
  Skyray, Magmakin Artillerist and every other discard payoff see a
  cycling exactly as they see a looting.
- The discard goes through `routeCardToZoneLocked`, so the CR 614
  replacement window runs over it (#853) and a future madness (#657) or
  Library of Leng sees a cost discard as a discard.
- `DiscardCauseCost` sets `zoneRoute.MustSettleNow`, which is what makes
  Decision 3 true.

**Order of payment.** The discard components are paid LAST — after the
mana, the life, the loyalty counters, the counter removals and the
sacrifices, at the point where `ActivateCatalogAbility` already sets
`source = nil` because the payment has invalidated the pointer. Paying
the discard earlier would invalidate the source in the middle of the
other components for a `DiscardSelf` cost.

**Announce-time choices, no new pause point.** The activator names the
cards for a `DiscardCards` cost in `ActivateAbilityParams.DiscardIDs` /
the wire's `discard_ids`, beside `sacrifice_ids`, `crew_ids` and
`counter_source_ids`. This is the announce-time cost pause that already
exists — the client's cost picker, where "Sacrifice a creature" on an
ability and the cast path's "discard a card" additional cost
(`DiscardCostModal.svelte`) are already answered — and it is reused
rather than duplicated. Every id is validated before anything is paid:
exactly N, distinct, in the activator's hand, each matching `Match`, and
never the source of an ability activated from hand.

Nothing here is a `PendingChoice`, so nothing here needs a
`PendingChoiceKind`, a `choiceGateDecisions` row or a `choiceMoves` case
(AGENTS.md §7, #730/#794). That is a consequence of Decision 3, not an
omission.

### 3. A cost settles now — CR 601.2h / CR 602.2b — including for a commander

CR 602.2b activates an ability in one indivisible step. The engine has
one place where that is expressed, and it is a bit on the event rather
than a policy in the cost code: `DiscardCauseCost` →
`zoneRoute.MustSettleNow`. The CR 614 window still runs, so a
replacement effect still SEES the discard; it settles without asking, so
the payment cannot stop halfway with an ability half-announced.

**A commander cycled from hand goes to the graveyard.** CR 903.9 offers
its owner the command zone "instead" whenever a commander would be put
into a graveyard from anywhere — the hand included, since #539 made the
window open over one exit primitive. But CR 903.9 is a **may**, and the
engine cannot ask a question in the middle of paying a cost. The rule
that resolves the tension is the one already written for a commander
pitched to Thrill of Possibility's additional cost and for
`payLifeAsCostLocked`: a cost discard does not pause, so the unanswered
"may" declines and the commander lands in the graveyard.

This is a real (small) divergence from paper, where the owner would be
asked, and it is recorded here rather than papered over. It is also
narrow: a player has to be cycling their own commander. Widening it
would mean making an announce interruptible, which risks every
activation's atomicity to fix one card in one hand. #903's
resolution-time prompt work is where that question gets reopened, if it
ever should be.

### 4. Special actions (CR 116.2): the verb, designed here, built on #658 / #659

**One `special_action` verb with a kind, not one verb per keyword.**

```go
// TypeSpecialAction is a CR 116.2 special action: a game action a
// player takes without using the stack and without passing priority.
// Params: {card_id, kind}.
TypeSpecialAction Type = "special_action"
```

with `kind` one of `foretell`, `suspend`, and later `turn_face_up`
(CR 116.2g — morph, megamorph, disguise).

One verb, because the rules say these are one category and the verb
table is where that categorisation has to be visible. Four verbs would
be four `requirePriorityHolder` calls, four `unmarshalParams` blocks and
four enumerator cases for a set of actions whose entire shared contract
— no stack, no announce, no response window, no targets — is the reason
CR 116 groups them at all. The per-kind differences are two fields wide
(when, and what it costs), which is a switch, not a verb.

**Timing is per kind, and it lives beside the kind, not in the verb.**

| kind | rule | window |
|---|---|---|
| `foretell` | CR 702.143a, CR 116.2h | any time you have priority during **your** turn; **legal under split second** (CR 702.61b) |
| `suspend` | CR 702.62a, CR 116.2f | any time you could begin to **cast** the card, which imports "can't cast" effects (CR 702.62c) and therefore **is not** legal under split second |

The two enumerators return early on `g.SplitSecondActive`
(`legal/cast.go:37`, `legal/abilities.go:39`). A special-action
enumerator MUST NOT copy that early return; it asks each kind. This is
written down because copying the two lines above it is exactly what will
happen otherwise, and the resulting bug — foretell illegal under a
Trickbind — is invisible until somebody plays a split-second card.

**The catalog declares one with a `Spec` slot, not a `Catalog*` hook**
(AGENTS.md §7 "Adding a Spec slot", #622 / #627): `Spec.SpecialActions
[]SpecialAction`, `game.CardDef.SpecialActions`, one line in
`effects.buildDef`, and the engine call site is the new verb's handler.

**The legal-enumeration hook** is a new `specialActionMoves()` on
`internal/legal`'s enumerator, called from where `activatedMoves()` is,
emitting `Move{Type: TypeSpecialAction, Kind: KindSpecialAction, …}`. It
is a new MOVE type, not a new `PendingChoiceKind`, so
`TestEveryChoiceKindIsClassifiedAndEnumerated` does not cover it — which
is precisely why it is named here: the move must be added in the same PR
as the verb, or bot seats will never take the action.

**The client menu row** is a row in the hand card's popover, beside the
`hand_abilities` rows of Decision 5 and above the `alternative_costs`
rows: "Foretell {2}", "Suspend 3—{R}". It fires `special_action`
directly with no cost picker, because neither kind's cost needs a
choice; a kind that later does reuses the picker `hand_abilities` uses.

Nothing above is built on this ADR. #658 and #659 are pure
implementation: the verb, the `Spec` slot, the enumerator method, the
two timing predicates, and — for suspend — the exile-zone triggers
Decision 7 defers.

### 5. The client learns about a hand ability from `hand_abilities`

`CardView` grows

```go
HandAbilities []ActivatedAbilityView `json:"hand_abilities,omitempty"`
```

stamped for the cards in a seat's own hand, alongside
`alternative_costs` and `exile_play`, by a `stampHandAbilities` pass
that mirrors `stampActivatedAbilities`. It reuses `ActivatedAbilityView`
whole — the cost fields, `condition_unmet`, `demands_x` and
`legal_targets` are the same questions — and the client sends the same
`activate_ability` action with the same `ability_index`.

**A separate field, not `activated_abilities` on a hand card.** The two
lists are disjoint by construction (a card in hand projects only the
abilities that function from hand; a permanent only those that function
there) but they are read by different UI: a permanent's rows are its
menu, and a hand card's rows sit in a popover next to "cast" and the
alternative-cost offers. A separate wire field means a client that has
not been taught about hand abilities shows nothing, rather than showing
a cycling row on a battlefield permanent, and it leaves room for
Decision 4's special-action rows to join without a third shape.

**The index is the ability's index in the card's full list**, not its
index among the hand-functioning ones. The engine validates the zone
after the index lookup, so the two can never disagree.

Two cost projections are added to `ActivatedAbilityView` for
Decision 2: `discard_self` (advisory — the client renders "Discard this
card") and `discard_cost_n` / `discard_cost_label` /
`discard_cost_options` (the count, the printed clause, and the cards in
hand that could pay). `discard_cost_options` is what lets the client
skip the picker when the hand holds exactly N payable cards, which is
CR 601.2h feeling right rather than being a modal with one button.

**Only the hand's owner.** Unlike a permanent's abilities, which are
public, a hand is not. `stampHandAbilities` runs over every seat's own
hand and `FilterViewFor` strips other seats' hands wholesale, as it
already does for their contents.

### 6. Bots enumerate a hand activation with the same cost solver

`legal.activatedMoves` gains a second loop, over the seat's own hand,
sharing every line of the per-ability body with the battlefield loop:
the same sorcery-speed gate, the same `Condition` call, the same
`affordableXExcluding` mana solve, the same target expansion, the same
`MaxExpansionPerSource` budget. What differs is the source list and a
zone predicate, and that is all that differs.

A discard cost's payment set is solved the way the sacrifice and counter
costs are: one cheapest payable set, one move, not one move per subset.
"Cheapest" for a discard is the hand order the seat already holds,
filtered by `Match` and excluding the source. A seat that cannot pay
produces no move — the #544 rule, which is that the enumerator must
never offer a move the engine will refuse.

### 7. `EventCycle` now; the cycle trigger from the graveyard later

Cycling a card (CR 702.29b) is a thing that happens, and Astral Slide,
Drake Haven, Fluctuator and New Perspectives all watch for it. The
engine emits `EventCycle` when a cycling ability is activated, carrying
the cycled card and its controller, and `effects.WheneverYouCycle` is
the one-line trigger constructor over it. A watcher **on the
battlefield** works with no change to the harvester at all: the
battlefield scan is the harvester's first pass.

**"When you cycle THIS card" (CR 702.29c) does not.** The card is in the
graveyard by the time the event fires — that is the rule, not an
accident — and the harvester's zone scan is the battlefield plus two
narrow special cases (`harvestCastFromStack` for "when you cast this
spell", `harvestLTB` for a card that has just left). A third narrow scan
is the right shape, and it is deliberately NOT built here:

- It needs a zone dimension on `TriggeredAbility` — the CR 113.6
  question Decision 1 answers for activated abilities, asked again for
  triggered ones — and the same dimension is what suspend's upkeep
  countdown and its last-counter cast trigger need **from exile**
  (#659). Building it for one card now means building it twice.
- `server/internal/game/triggers.go` is under concurrent change for the
  combat-damage trigger keys.

Consequence, recorded honestly: **Magmakin Artillerist keeps one
caveat.** It can be cycled and the cycling draws, but its own "When you
cycle this card, it deals 1 damage to each opponent" does not fire. Its
discard trigger does. Decree of Pain, which is nothing but a cycle
trigger, is therefore NOT added on #660. `docs/engine-seams.md` carries
the row.

### 8. The keyword table, and what the coverage signal says

"cycling" joins `canonicalKeywords` (`server/internal/game/keywords.go`)
in the same change that teaches the engine to honour it, which is the
rule ADR 0037 and that table's own comment already state. "foretell" and
"suspend" join it on #658 and #659, not before.

`NeedsCatalogEffect` is unaffected: it reads printed text, and a card
that prints "Cycling {2}" prints rules the engine will not run without a
`Spec` either way. What changes is that the cycling cards stop declaring
a caveat about it, so their `Completeness` moves from `caveats` to
`full`, and Sylvan Reclamation gets a real `Completeness` for the first
time.

## Consequences

- One activation path with a zone dimension. Reassembling Skeleton,
  Drownyard Temple and the Spirit Guides become card files rather than
  engine work; Eternal Dragon becomes a card file with two ability
  entries once the graveyard half of the seam row opens.
- One discard helper with one cause, so a cost discard is visible to
  replacement effects and to discard payoffs without either of them
  knowing that costs exist.
- One cost pause point: the client's announce-time picker. No new
  `PendingChoiceKind`, no interruptible announce.
- A commander cycled from hand is lost to the graveyard rather than
  offered the command zone. Documented, narrow, and a consequence of an
  atomic announce rather than of a missing feature.
- Magmakin Artillerist's own cycle trigger stays unimplemented until the
  triggered-ability zone dimension exists, which is the same piece of
  work suspend needs.
- `hand_abilities` is a second ability list on the wire. A client that
  ignores it loses nothing it had.

## Alternatives considered

**Cycling as an alternative cost.** What #94 and five card files assume.
Rejected: nothing is cast, the card goes to the graveyard rather than
the stack, the draw is an ability resolving, and the "when you cycle"
trigger fires from the graveyard. Every observable differs from a cast.
Modelling it as one would also make it counterable as a spell, stoppable
by "can't cast" effects, and visible to cast triggers, all of which are
wrong.

**`FromHand bool` on the ability.** Smaller today. Rejected in
Decision 1: the next four cards on the same seam row want the graveyard
and exile, and a bool would have to be deleted from every card file that
set it.

**A `cycle` verb of its own.** Rejected: cycling is an activated ability
(CR 702.29a) and it uses the stack, which is exactly what a special
action does not do. A verb would need its own cost payment, its own
stack push and its own enumerator entry, all duplicating
`activate_ability`.

**One `AbilityCost.DiscardCards` covering "this card" as N = 1 with a
`Match` on the source.** Rejected in Decision 2: it puts a chooser and
an options list on the wire for every cycling ability and admits a
payload that names the wrong card.

**A `PendingChoice` for the discard-a-card cost.** Rejected: CR 602.2b
makes an activation indivisible, so the prompt would have to pause an
announce. The engine already answers announce-time cost choices in the
client's picker — for sacrifice-as-cost on an ability, and for
discard-as-cost on a cast.

**One special-action verb per keyword.** Rejected in Decision 4: four
copies of one handler for a set of actions CR 116 groups precisely
because their contract is identical.

## Open questions

1. **The triggered-ability zone dimension** (Decision 7). `Zones` on
   `TriggeredAbility` plus a narrow "the card the event names, wherever
   it is" harvest. Wanted by Magmakin Artillerist, Decree of Pain,
   suspend's two triggers, and every "when this card is put into a
   graveyard from anywhere" the catalog has skipped. Should be its own
   issue on #886.
2. **Per-instance activation grants** — Greater Gargadon activating from
   exile while suspended. `ExilePlayPermission` is the model; the union
   with `Zones` is one line at the check, and nothing needs it before
   #659.
3. **New Perspectives' "you may pay {0} rather than pay cycling costs"**
   (#625) is an alternative cost on an activated ability. It needs an
   `Alternatives` slot on `AbilityCost` that ADR 0020 deliberately does
   not have, and it is out of scope here.

## Amendment — 2026-09-18: Decision 4 is built

Decision 4 designed the special-action verb and deliberately built
none of it. [#658](https://github.com/krakenhavoc/cmd_and_ctrl/issues/658)
(foretell, CR 702.143) and
[#659](https://github.com/krakenhavoc/cmd_and_ctrl/issues/659)
(suspend, CR 702.62) shipped it together on one PR, and it is what the
ADR said it would be, item for item:

- **One verb.** `actions.TypeSpecialAction` = `special_action`, params
  `{card_id, kind}` plus the `{strict, auto_tap}` payment pair every
  other announce carries. One engine entry,
  `game.PerformSpecialAction` (`server/internal/game/special_action.go`).
- **A per-kind timing table**, `Game.SpecialActionTimingOKLocked`, and
  it is the ONLY place either window is written down: the engine, the
  bot enumerator and the wire's `available` flag all read that one
  function.
- **The split-second warning was worth writing.**
  `legal/special_actions.go` does not open with the
  `if g.SplitSecondActive { return }` its two neighbours open with;
  it asks the table per kind. Foretell under a Trickbind is legal
  (CR 702.61b) and there is a test row for it.
- **`Spec.SpecialActions []game.SpecialAction`**, `CardDef.SpecialActions`,
  one line in `effects.buildDef`, and the verb's handler as the call
  site — the AGENTS.md §7 "Adding a `Spec` slot" recipe, unchanged.
  `effects.Register` refuses at boot a kind the engine cannot carry
  out, an unparseable cost, and a suspend with no time counters.
- **`specialActionMoves()`** on the `legal` enumerator, called where
  `activatedMoves()` is, emitting `Move{Kind: KindSpecialAction}`; a
  `payoffOf` case in `aiseat/heuristic`.
- **The client row** is `CardView.special_actions`, stamped only on
  the viewer's own hand and stripped by the same redaction that
  strips `hand_abilities`, rendered in the hand card's menu above
  "move to" and fired with no picker.

Two things Decision 4 did not say, decided while building:

- **`turn_face_up` (CR 116.2g) is reserved, not stubbed.** A kind with
  no performer is refused BEFORE its cost is paid and is never
  projected to a client, so the verb can grow a third kind without
  ever having half of one.
- **A special action's mana is spent with no `ManaSpendPurpose`.**
  It is neither a cast nor an activation, so mana restricted to
  either cannot pay for it. That is the strictly-weaker reading, which
  is the direction #259 requires.

Decision 7's open question 1 — the triggered-ability zone dimension —
was answered by #925 before either keyword needed it, and suspend's
upkeep countdown rides `TriggeredAbility.Zones = {ZoneExile}` exactly
as that issue predicted. Open question 2 (per-instance activation
grants, Greater Gargadon) is still open.

## Amendment — 2026-09-23: the special action's cost is priced through CR 601.2f (#1319)

The amendment above built the verb's payment as a bare parse-and-pay:
`sa.Cost` in, `payAbilityManaCostLocked` charges it, and the doc
comment on that call said outright that "a special action is not an
activation … nothing prices it through the CR 601.2f pass." Ranar the
Ever-Watchful's "The first card you foretell each turn costs {0} to
foretell" (deck tracker #1306) needed exactly that pass, and this is
the third door it opens — [ADR 0020](0020-activated-abilities.md)'s
2026-09-22 `#1184` note already opened the second one, for activated
abilities, in the identical shape.

**`CostQuery.SpecialAction` (`*SpecialActionCostSubject`, one field:
`Kind`)** is `CostQuery.Ability`'s twin, naming which CR 116.2 action
is being priced. **`CostModifier.SpecialActions`** is the third
partition bit, beside `.Activations`: `activeCostModifiersLocked` now
switches on which of `q.Ability` / `q.SpecialAction` is set (or
neither, for a cast) and shows a modifier only the announcements its
own bit claims. A card written as "spells cost {1} more" — Sphere of
Resistance's `AppliesTo` is nil, meaning "every spell" — could not
otherwise be trusted not to start taxing every foretell in the game
the day the field appeared, exactly the risk the #1184 note names for
activations.

**`Game.SpecialActionManaCostForEffect(actor, card, kind, sa)`** is
`AbilityManaCostForEffect`'s special-action twin and the one function
three readers share: `PerformSpecialAction` pays it,
`legal.specialActionMovesForCard` prices affordability against it
(replacing a bare `game.ParseCost(sa.Cost)` that could have offered a
move the engine then refused, #544's rule again), and
`SpecialActionView.ChargedCost` renders it on the wire beside the
printed `Cost` — `ActivatedAbilityView.ChargedManaCost`'s shape,
one door over, so a discounted row and its real price can never
disagree.

**The per-turn tally the clause reads** is `Game.ForetoldThisTurn` /
`ForetoldCountThisTurn(player)`, `Game.SpellsCastThisTurn`'s shape one
zone over. It is bumped inside `foretellLocked`, after the card has
actually landed in exile — a special action a replacement window only
PAUSED has not foretold anything yet, and one that never lands must
not spend the discount a player never got to use.

**Self modifiers stay cast-only.** `activeCostModifiersLocked`'s self
slot ("THIS SPELL costs {N} less to cast", CR 113.6d) is skipped
whenever `q.Ability` OR `q.SpecialAction` is set, for the reason the
#1184 note gives for activations: a spell's own printed reduction is
about the card as a spell on the stack, and neither a permanent's
ability nor a card's own special action is that.

`effects.SpecialActionCostsLess` (`ASpecialActionOfKind`,
`TheFirstOneThisTurn`) is the card-file constructor, `ActivationCostsLess`'s
shape one door over. Shipped on **Ranar the Ever-Watchful** (its
"create a Spirit" trigger stays a caveat, waiting on the exile-cause
event #1320 brings). See
[docs/engine-seams.md](../engine-seams.md)'s "Special actions from the
hand, and foretell" entry for the closed-seam summary.

## Amendment — 2026-09-23: plot is the fourth kind (#1342)

Plot (CR 702.170a) is a hand special action: "any time you have priority
during your main phase while the stack is empty, you may exile this card from
your hand and pay [cost]. It becomes a plotted card." It needed a row in each
of Decision 4's per-kind tables and nothing else:

| table | plot's row |
|---|---|
| `specialActionZone` | the actor's **hand**, like foretell and suspend |
| `SpecialActionTimingOKLocked` | `sorcerySpeedOpenLocked` — the owner's main phase, stack empty. It is the **keyword's** window, not the card's, so an instant or a card with flash is plotted at sorcery speed too (the reminder text's "Plot only as a sorcery"). That is where plot and suspend differ: suspend asks the card's own casting window (CR 702.62c). No split-second clause is needed, because split second only matters while a spell is on the stack and this window needs an empty one. |
| `specialActionPerformer` | `plotLocked` (`game/plot.go`): route the card to exile **face up** with `MoveCauseSpecialAction` (#1320), then call `PlotExiledCardForEffect` on what landed |

The plotted state is #1318's, unchanged: [ADR 0066's 2026-09-23
amendment](0066-granted-cast-and-play-permissions.md) gives a card in exile a
free, per-instance `CastPermission` with `TimingPlot` and a `NotBeforeTurn`
floor. A card plotted from hand and a card Aven Interrupter plotted are the
same object with the same permission. A commander whose owner sends it to the
command zone instead (CR 903.9) never landed in exile, so nothing is plotted.

The catalog line is `effects.Plot("{cost}")`: `Cost` is the plot cost and
`CastCost` stays empty, because the later cast is free and the grant sets the
price. `"plot"` joined `canonicalKeywords` in the same change, as foretell and
suspend did. The enumerator, the wire's `special_actions` row, and the
client's menu row needed no code: each walks `SpecialActionsOfferedByCard` and
asks the timing table, which is Decision 4's "one verb" claim holding for a
fourth time.

Proof cards: Djinn of Fool's Fall, Spinewoods Paladin, Beastbond Outcaster and
Plan the Heist, all `full`. Still open on this family (#1382): "when this card becomes
plotted" triggers (Longhorn Sharpshooter, Aloe Alchemist) need an event for the
plotted state. Fblthp, Lost on the Range, which plots from the top of the
library, needs the special action to reach a zone other than the hand.
