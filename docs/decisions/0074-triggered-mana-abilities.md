# ADR 0074 — Triggered mana abilities: a trigger that never reaches the stack, and the colour it can read

**Status:** Accepted · 2026-09-18 · S44 — Mana and cost components
**Issues:** [#763](https://github.com/krakenhavoc/cmd_and_ctrl/issues/763)
(triggered mana abilities, CR 605.1b / CR 605.4a, and the produced-colour seam
row it merges), tracker
[#887](https://github.com/krakenhavoc/cmd_and_ctrl/issues/887)
**Numbering:** on 2026-09-18 every ADR filename that exists on any fetched ref
was collected with `git log --all --diff-filter=A --name-only --pretty=format:
-- docs/decisions/` after `git fetch origin` (302 remote heads). The highest
number present anywhere was `0073-optional-additional-costs-and-the-cast-gate.md`
(the #664/#760 branch, PR #988), so this ADR takes **0074**. The scan was
repeated immediately before the push. 0005, 0024, 0029 and 0030 stay
permanently unused per AGENTS.md §4.
**Builds on:** [ADR 0040](0040-mana-pipeline.md) (the mana pipeline — this adds
the fourth way mana reaches a pool), [ADR 0011](0011-mana-pool-and-auto-tapper.md)
(the pool and the auto-tapper), [ADR 0018](0018-triggers-on-the-stack.md) /
[ADR 0007](0007-stack-foundation.md) (every OTHER trigger, which does use the
stack), [ADR 0071](0071-designations-that-switch-abilities-on.md)
(`Designation.Active` gating and the `…ForCard` accessor shape this copies),
[ADR 0010](0010-card-effect-catalog.md) / #622 (the one `CardDef` lookup),
[ADR 0036](0036-attachments.md) (Auras — four of the five cards here are one)
**Amends:** [ADR 0040](0040-mana-pipeline.md) — the mana pipeline gains a
production hook and `EventManaAdded` gains the colour it adds;
[ADR 0018](0018-triggers-on-the-stack.md) — "every trigger goes on the stack"
is no longer true, and this names the one exception and why it cannot be
written as an ordinary `TriggeredAbility`
**Designs, does not implement:** mana-production *replacement* (CR 106.12b —
Nyxbloom Ancient, Mana Reflection: its own engine-seams row), turn-scoped mana
triggers (High Tide, Bubbling Muck — left to
[#663](https://github.com/krakenhavoc/cmd_and_ctrl/issues/663)'s registry),
triggers on mana produced without a tap (§3). Counting triggered mana in the
auto-tap PLAN was listed here until the [amendment of
2026-10-30](#amendment-of-2026-10-30-4716-the-auto-tap-planner-counts-triggered-mana)
replaced §7 and built it
**Amended:** 2026-10-14, [a spell-made mana trigger for its controller
only](#amendment-2026-10-14-3504-a-spell-made-mana-trigger-for-its-controller-only)
(#3504). 2026-10-30, accepted the same day: [the auto-tap planner counts triggered
mana](#amendment-of-2026-10-30-4716-the-auto-tap-planner-counts-triggered-mana)
(#4716), which replaces §7 (Decision 6) and was built on 2026-10-31 ([as
built](#as-built-2026-10-31)).

---

## 1. The problem

CR 605.1b makes a triggered ability a **mana ability** when three things are
true at once: it triggers off an activated mana ability resolving (or off mana
being added), it does not target, and it could add mana. CR 605.4a then says a
mana ability does not use the stack — it simply resolves, with no priority
window for anybody.

The engine had no such path. Two separate facts made Wild Growth unwritable:

- **Every trigger goes on the stack.** The harvester appends each built item to
  `g.PendingTriggers` (`game/triggers.go`), so a Wild Growth written as an
  ordinary `TriggeredAbility` would give both players a priority window that
  CR 605.4a does not allow, and would let the extra `{G}` arrive *after* the
  spell that needed it had already been paid for.
- **No colour on the event.** `EventManaAdded` said "the token's color rides
  Amount as 0". Nothing could tell a Mana Flare which type the land produced.
  That is the separate engine-seams row "Mana-ability trigger carrying produced
  colour" (Forsaken Monument, Ultima), which #763 merges into this one because
  it is the same pipeline.

And the mana itself is produced in **three** places, not one: the hand-clicked
`ActivateManaAbility`, the auto-tapper's executor `materializePlanLocked`, and
`ResolveManaChoice` — which is the *only* one of the three that knows the
colour a Birds-of-Paradise-style source produced.

## 2. Decision 1 — a separate catalog slot, not a `TriggeredAbility` with a flag

Triggered mana abilities live on **`Spec.ManaTriggers []game.ManaTrigger`**,
read by the engine through **`CatalogManaTriggers`** (the per-slot hook) and
**`game.ManaTriggersForCard(c)`** (the gate-applied accessor), exactly as
`Spec.Triggered` is read through `CatalogTriggers` / `TriggersForCard`.

A flag on `TriggeredAbility` was the obvious alternative and is wrong, for a
reason that is structural rather than stylistic: *nothing* on
`TriggeredAbility` survives the change. `Watches []EventKind`, `AppliesTo(ev
Event, …)`, `Build(…) *StackItem`, `OptionalPrompt`, `Targets`,
`HasLegalTarget`, `OncePerBatch`, `FromStack` and `Zones` all exist to get an
item onto the stack correctly — and a mana trigger has no item, no targets (CR
605.1b forbids them), no prompt and no batch. A flagged `TriggeredAbility`
would be a struct with nine fields the harvester must learn to skip, and the
skip would be one `if` away from a Wild Growth landing on the stack.

What it keeps is the two things that are not about the stack:

- **`ActiveWhen Designation`**, evaluated in `ManaTriggersForCard` and nowhere
  else, so a gated-off mana trigger is never matched (ADR 0071).
- **`CatalogAbilityKey`**, so a CR 613.1f ability-removing effect (Song of the
  Dryads on the Aura, Kenrith's Transformation) takes the trigger away. This is
  the "the Aura losing its abilities stops the trigger" case #763 asks for, and
  it is free: it is the key, not a check.

```go
type ManaTrigger struct {
	Label      string
	AppliesTo  func(prod ManaProduced, source *Card, g *Game) bool
	Produced   func(prod ManaProduced, source *Card, g *Game) string
	ActiveWhen Designation
}
```

`Produced` returns a string in the `ParseProducedMana` grammar — the *same*
grammar a mana ability's `Produced` uses, pipes and per-colour amounts
included. So "adds an additional `{G}`" is `"{G}"`, Overgrowth is `"{G}{G}"`,
Fertile Ground is `"{W|U|B|R|G}"`, and "one mana of any type that land
produced" is a pipe built from the payload. Returning `""` adds nothing, which
is the answer for every clause whose input is not there yet.

There is deliberately **no recipient callback**. The mana goes to
`prod.Controller` — the player whose pool the original mana landed in — because
that is what every printed card says ("its controller adds", "that player
adds", and Mirari's Wake's "add", which reaches the same player via its own
`AppliesTo`). A card that ever differs is when the field gets added.

## 3. Decision 2 — one payload, one firing function, three call sites

```go
// ManaProduced is one completed "tapped for mana" production.
type ManaProduced struct {
	Source     Card      // the permanent tapped, as it was when it produced
	Controller uuid.UUID // whose pool the mana landed in
	Colors     []string  // the mana types added, in the order added
}
```

`Source` is a **value copy** rather than a pointer, because a tap-cost ability
may also sacrifice its source (Lotus Petal) and the predicate still has to be
able to ask what it was.

`(*Game).fireManaTriggersLocked(prod, pending)` is the one function. It walks
the battlefield in order, asks `ManaTriggersForCard` for each permanent,
applies `AppliesTo`, and hands `Produced`'s string to the shared adder — in
**two passes**: every condition is judged against the board as it was when the
mana was produced, and only then does anything resolve. That is the CR 603.2
split, and it is also what keeps the walk safe, since adding mana emits events
and an event runs listeners. It is called from exactly three places and never
anywhere else:

| Site | When it fires | `pending` |
|---|---|---|
| `ActivateManaAbility` | after the rider, if the activation put at least one token in a pool directly | `nil` |
| `ResolveManaChoice` | when the answered pick was a mana ability's (`PendingChoice.ManaTapped`) | `nil` |
| `materializePlanLocked` | once per planned source, after its slots are minted | the cast's remaining colour requirements |

**Two rules make "once per production" fall out of that table instead of
needing a token to track it.** A mana ability either mints its mana directly
(a Forest, Sol Ring) or queues a pick (Birds, a dual, a chosen-colour land, a
Gilded Lotus) — no printed card does both in one ability. So the activation
fires only when it minted something directly, and the pick fires when it is
answered; each real card takes exactly one of the two branches. An ability that
did both would fire twice; it is unreachable, and it is written down here
rather than guarded against, because the guard would be per-activation state on
`Game` that the snapshot would then have to carry.

**`AddMana` from a resolving spell never fires** (CR 106.12a: Dark Ritual is
not "tapped for mana"). `AddManaForEffect` does not call the firing function,
and the picks it queues leave `ManaTapped` false, which is the whole of the
rule — there is no second place to keep in step.

**Only a TAP fires.** `ManaProduced` is built only where the mana ability's
cost included `{T}`. Every printed triggered mana ability says "tapped for
mana" (Wild Growth, Mana Flare, Crypt Ghast's Swamp, Forsaken Monument's
permanent), so this is the narrowest shape that covers the printed cards, and
it means an Ashnod's Altar or a Treasure cannot be mistaken for one. Widening
it is a payload field and a call site, on the day a card asks.

**Triggered mana does not re-trigger.** Not a depth counter: the adder in §4
never calls the firing function, so there is nothing to recurse. That is also
the rule (a Mana Flare does not see Wild Growth's `{G}`, because nothing was
tapped for it), so the implementation and the rule are the same sentence.

## 4. Decision 3 — the colour choice inside a trigger reuses the two modes the pipeline already has

"Adds an additional one mana of any color" (Fertile Ground) and "one mana of
any type that land produced" (Mana Flare, Mirari's Wake) both make the
**trigger** the thing with a colour choice. The engine already has exactly two
answers to "a mana slot with several colours", and this reuses both rather than
inventing a third:

- **On a player's own activation** (`pending == nil`): queue the same
  `PendingChoiceMana` a Birds of Paradise activation queues.
- **Inside the auto-tapper** (`pending != nil`): pick greedily against the
  cast's still-unpaid colour requirements with `pickColorForSlot`, the same
  function `materializePlanLocked` uses for the source's own slots. The
  auto-tapper's contract is "no further player decisions", and a prompt
  appearing halfway through a cast would break it. Since the [amendment of
  2026-10-30](#amendment-of-2026-10-30-4716-the-auto-tap-planner-counts-triggered-mana)
  (item 5) the colour the plan booked for the pick comes first, and the
  greedy pick is the fallback when the plan booked none.

One body serves all of it: `AddManaWithOptionsForEffect`'s slot walk was
extracted into `addManaSlotsLocked(p, source, produced, narrow, pending)`, and
the trigger adder is that function with a non-nil `pending` in the auto-tap
case. There is no parallel "add mana" path; there is one, with a mode.

The token's `Source` is the **trigger's** permanent (the Aura), not the land:
the mana comes from Wild Growth's ability.

## 5. Decision 4 — `EventManaAdded` carries the colour

`Event.Colors` already exists (`#761` put the distinct colours of a payment on
`EventManaSpent`). `EventManaAdded` now carries the one colour it added, at all
four emit sites — the two in `mutations.go`, the one in `ResolveManaChoice` and
the one in `AddManaForEffect`. No new field, no wire schema change, and the
"the token's color rides Amount as 0" note in `events.go` is gone.

This closes the engine-seams row "Mana-ability trigger carrying produced
colour" on its own terms: a listener that only wants to know what colour
arrived no longer has to read the pool.

## 6. Decision 5 — what is NOT in scope, stated

- **Mana-production replacement** (CR 106.12b — Nyxbloom Ancient, Mana
  Reflection). That is a replacement effect on the production, not a trigger
  after it, and it keeps its own engine-seams row.
- **Turn-scoped mana triggers** (High Tide, Bubbling Muck — "until end of turn,
  whenever a player taps an Island for mana…"). The slot here is catalog data
  read off a permanent; a turn-scoped one needs a cloned, snapshotted registry
  cleared at cleanup, which is the question
  [#663](https://github.com/krakenhavoc/cmd_and_ctrl/issues/663) is already
  holding. Left there on purpose rather than answered twice.
- **CR 106.7 "could produce"** does not count a triggered mana ability. An
  Exotic Orchard facing a Wild-Growth-enchanted Island sees `{U}`, not
  `{U}{G}`. CR 106.7 is about "an ability of that permanent", and Wild Growth's
  ability is the Aura's — so this is also the correct reading, not only the
  convenient one.
- **Mana produced without a tap** — §3.

## 7. Decision 6 — the auto-tap PLANNER does not count triggered mana

> **Superseded 2026-10-30** ([#4716](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4716), accepted the same day,
> built 2026-10-31 in [#4775](https://github.com/krakenhavoc/cmd_and_ctrl/pull/4775)):
> the planner counts what a triggered mana ability would add, judged by the
> trigger's own condition. See [the amendment](#amendment-of-2026-10-30-4716-the-auto-tap-planner-counts-triggered-mana)
> and [as built](#as-built-2026-10-31). This section no longer describes the
> code. It is kept as written, the decision the amendment replaced, so read its
> present tense as 2026-09-18.

The planner (`gatherTapSources` / the solver) models only what a source's own
mana ability produces. A land enchanted with Wild Growth is planned as a
one-mana source; the extra `{G}` arrives when the executor taps it and floats.

The consequence is stated plainly: **the tapper taps more lands than it needed
to, and the surplus sits in the pool until the step ends.** That is weaker than
printed and it is safe, which is the direction ADR 0011 and #259 both take. The
alternative — teaching the solver that some sources produce extra mana whose
colour depends on a trigger that may or may not apply — would put a second,
approximate copy of the trigger's `AppliesTo` inside the planner, and the two
copies drifting is exactly the failure `autoTapAbilityFor` and
`manaTapBlockedBySickness` are shared to prevent.

The executor is *not* in the same position and does not need to be: it fires
the real trigger, so the mana that arrives is the mana the cards say.

## 8. Consequences

- One new catalog slot, one new engine accessor, one new firing function, one
  new payload struct, and one bool on `PendingChoice`. No new event kind, no
  new prompt kind, no wire schema change.
- `docs/engine-seams.md` closes two rows with one entry.
- A card author writing "whenever … is tapped for mana" now has a recipe
  (AGENTS.md §7 (now docs/adding-cards.md)) and must not reach for `Spec.Triggered`. The distinction is
  observable in one line: if the ability adds mana and does not target, it is a
  `ManaTrigger`; Electro's and Fire Nation Palace's "add mana" triggers fire on
  a cast and on an attack, so they are ordinary stack triggers (CR 605.5a) and
  stay where they are.

## 9. Cards this unblocks in this PR

| Card | What it proves |
|---|---|
| Wild Growth | the mechanism end to end: an Aura's fixed `{G}`, no stack item, no priority window |
| Overgrowth | more than one token from one trigger |
| Utopia Sprawl | composition with #742's stored `Card.ChosenColor`, and an "Enchant Forest" host predicate |
| Fertile Ground | a colour CHOICE inside the trigger — a prompt by hand, a greedy pick under the auto-tapper |
| Mana Flare | the produced COLOUR being read (`ManaProduced.Colors`), and a trigger that fires on an OPPONENT's land |

## Amendment (2026-10-14, [#3504](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3504)): a spell-made mana trigger for its controller only

`DelayedTrigger.ManaTapSubtype` / `ManaAdds` (built by `effects.TurnManaTrigger`, read by `fireManaTriggersLocked`) is the
spell-made form of this mechanism: "until end of turn, whenever a player taps a Swamp for mana, that player adds an additional {B}"
(Bubbling Muck, High Tide). Pyre Rhymer // Molten Tide prints "whenever **you** tap a Mountain", so the record gains
`ManaTapControllerOnly`: when set, a production whose pool owner (`ManaProduced.Controller`) is not the record's
`Controller` is skipped, and an opponent's Mountain neither triggers it nor receives the mana. The default is off, so the
two older cards are unchanged. The field is an additive snapshot key (`manaTapControllerOnly`, shape v8 recorded).

Back-out: drop the `ManaTapControllerOnly` check in `fireManaTriggersLocked` and
`TestControllerOnlyManaTriggerFiresForYourMountainsOnly` and `TestPyreRhymerMoltenTideAddsForYourMountainsOnly` fail; the
High Tide / Bubbling Muck tests keep passing.

## Amendment of 2026-10-30 ([#4716](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4716)): the auto-tap planner counts triggered mana

**Status:** Accepted (owner decisions 2026-10-30) · 2026-10-30 · S78 — Launch week: player reports, flakes and first-game polish (tracker [#4648](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4648)). This amendment changes no code. The owner wants it built next, in the [slices](#slices) below.
**Owner decision (2026-10-30, on #4716):** option 1. Amend Decision 6 so the auto-pay planner counts the mana a triggered mana ability would add (Crypt Ghast, Wild Growth and the like). It checks the real trigger's own condition, so the plan and its preview are right before anything is tapped, and nothing is over-tapped or sacrificed (CR 605.1b, 605.4a). Calls this amendment makes on its own are under [Calls made here](#calls-made-here).
**Owner decisions (2026-10-30):** the owner accepted the amendment the same day and chose the recommended option on all three of its questions. The answers are under [Owner decisions](#owner-decisions-2026-10-30) and are binding. The options not chosen are kept under [Questions for the owner (answered)](#questions-for-the-owner-answered).
**Replaces:** §7 (Decision 6), and the header's "counting triggered mana in the auto-tap PLAN" under "Designs, does not implement".
**Composes with:** [ADR 0013](0013-replacement-effects.md) §5ab Decision 3 and its amendment of 2026-10-16 (#3939): the planner already prices CR 106.12b replacements through the replacement's own predicate, and this does the same for triggers. [ADR 0011](0011-mana-pool-and-auto-tapper.md) (the auto-tapper, its tiers and the #1285 preview). [ADR 0118](0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md) §1 (every payer plans through `autoTapTopUpLocked`).
**Why an amendment and not a new ADR:** it reverses one decision of this ADR, and the mechanism is this ADR's own firing function, called by one more reader. It adds no seam, no event, no prompt and no wire field. That is the in-file amendment convention (ADR 0011, ADR 0013 and ADR 0136 all amend in place). [ADR 0136](0136-planning-the-turns-mana.md) explains when a new ADR is right instead: when the change adds a new decision step with its own measurement and config. This one does not.

### What happened

In game `c16a7f14` on launch day, the reporter cast Veinripper ({3}{B}{B}{B}) with Crypt Ghast on the battlefield: "Whenever you tap a Swamp for mana, add an additional {B}." Auto-pay tapped every Swamp and also sacrificed a Treasure. The planner counts one mana per Swamp and reaches for a Treasure only when the lands fall short, so the board most likely had five Swamps and no other untapped source. Under Crypt Ghast, three Swamps pay six mana. The pinned replay (`GET /bugreport/8fe8b219-…/replay`, admin) can confirm the exact board.

This is Decision 6 working as written. The planner priced each Swamp as one {B}, found the lands one mana short, and reached the sacrifice tier (ADR 0011, #1215). The executor then fired the real trigger. Every Swamp added {B}{B}, the cost was paid with mana to spare, and the surplus floated until the step ended. The ordering is not at fault: the Treasure is already the planner's last ordinary tier. The planner was wrong about how much the lands make.

Decision 6 called this "weaker than printed and safe". It has two failures, not one:

- **It over-spends.** It taps lands the cost did not need. When the shortfall reaches the last tiers, it cracks a Treasure, exiles a Spirit Guide or spends Vivi's once-per-turn ability (ADR 0011's #1215, #1228 and #1621 tiers). None of those come back.
- **It under-reports.** A cast that only the trigger mana makes payable has no plan at all. Under strict payment (ADR 0118) it is refused and offered Cast anyway, the legal-move list does not offer it, and the ready ring and the bots do not see it. Three Swamps and Crypt Ghast read as three mana to a five-drop. That is the failure #273 and #779 were about: a payable cast read as unpayable.

### Decision 6, as amended

**The auto-tap planner counts the mana that triggered mana abilities would add when it taps a source. It judges each one by that trigger's own condition, through the same function the executor fires them with.**

#### 1. The planner asks the trigger itself; there is no second copy

`fireManaTriggersLocked` already works in two passes (§3): pass one judges every trigger against the production, and pass two resolves what pass one found. Pass one is split out as a read-only function, `manaTriggerFiringsLocked(prod ManaProduced) []manaTriggerFiring`. It walks the spell-made records (`DelayedTrigger.ManaTapSubtype`, `ManaTapControllerOnly`: High Tide, Bubbling Muck, Molten Tide) and then the battlefield, calls `ManaTriggersForCard`'s `AppliesTo` and `Produced`, and returns each firing's source, label and output string. The executor calls it and then resolves the list, as it does today. The planner calls it with a hypothetical production (item 2) and does not resolve anything. One predicate, two readers. This is ADR 0013 §5ab Decision 3's pattern (`producedManaPreviewLocked` walks the replacement's own `AppliesTo` and `Replace`), applied to triggers.

A cheap gate goes first: `manaTriggersExistLocked`, one battlefield walk asking whether any permanent has a mana trigger and whether any delayed mana record exists, like `producesManaReplacementsExistLocked`. On nearly every board the answer is no and the planner runs exactly as before.

**Considered and rejected: a declared shape on the trigger** (for example a `Plan` field such as "adds {B} when the tapped land is a Swamp"). The shape would be a second copy of `AppliesTo`, which is the drift Decision 6 was written to avoid. The conditions are also not data. Regal Behemoth asks whether you are the monarch. Sasaya counts the other lands with the same name. Caged Sun and Utopia Sprawl read a chosen colour, and Caged Sun also reads the produced colour. Forsaken Monument and Ultima ask whether {C} was produced. Winter's Night asks for a snow land. A grammar that could say all of that would be a second rules engine, and every new card would have to be written twice.

**What it costs.** `AppliesTo` and `Produced` now also run under a read lock (`ReadSnapshot`), and on a production that has not happened yet. Their doc already says "READ-ONLY: use *ForEffect accessors" (`ManaTrigger`, `mana_trigger.go`). That rule now carries weight, and the doc comment says so: the callbacks run under `g.mu` in either mode. For card authors it means three things. Read the tapped permanent through `prod.Source`, never through a fresh battlefield lookup. Read nothing from the mana pool. Write nothing. Every catalogued `ManaTrigger` card today already follows them; slice 1 adds a test that checks every one.

#### 2. The hypothetical production is the one the executor will make

The planner builds the `ManaProduced` that `materializePlanInLocked` would build for the same source:

- **`Source`** is a copy of the permanent as the executor copies it (`tappedForMana`), marked tapped when the ability taps.
- **`Controller`** is the player whose pool the executor mints into: the payer. That is still true when another player is making the payer's choices (ADR 0166, ADR 0167). It is never the seat that clicked.
- **`Colors`** are the colours after the CR 106.12b window, as the planner already prices them (`priceProducedSlotsLocked`). These are the colours `produceManaLocked` returns at the tap, which the executor passes on (`addedColors`). So under Mana Reflection a Swamp is {B}{B} and Crypt Ghast adds one {B}: it is one production, so it is one trigger (CR 106.12a; Forsaken Monument's ruling, "if you tap a permanent for more than one {C}, you add only one additional {C}").
- **Only an ability whose cost includes {T}** (CR 106.12): the executor's own `tapped` bit. A Treasure ({T}, sacrifice) is tapped for mana. A Gold or an Eldrazi Spawn (sacrifice only) and a Spirit Guide (exiled from the hand) are not, and no trigger is counted for them, just as none fires.
- **A production that adds nothing fires nothing**, as it does today.

#### 3. What the triggers add becomes slots in the planner's grammar

Each firing's `Produced` string is parsed with `ParseProducedMana`. Each slot is priced through the CR 614 window with `fromTap` false, which is the window the trigger's adder opens (`addManaSlotsLocked` mints with `ManaFromTap` false). A replacement keyed to a tap therefore does not double it (CR 106.12b; Mana Reflection's ruling of 2020-08-07: a triggered mana ability "won't be affected by Mana Reflection"), and any replacement that does apply to it is counted. The slots are appended after the source's own slots, and a new field, `tapSource.TriggerSlots`, says how many there are. The costed planner's per-slot list (`plannedTap.Slots`, #2455) and the executor's walk over the printed slots keep their meaning.

| Board | The source as planned |
|---|---|
| Swamp, Crypt Ghast | `{B}`, `{B}` (one tap pays two pips, like a Sol Ring) |
| Forest, Wild Growth | `{G}`, `{G}` |
| Forest, Overgrowth | `{G}`, `{G}`, `{G}` |
| Forest, Fertile Ground | `{G}`, `{W\|U\|B\|R\|G}` (a pick, item 5) |
| Forest, Utopia Sprawl with no colour chosen | `{G}` (`Produced` returns `""`) |
| Swamp, Mana Reflection, Crypt Ghast | `{B}`, `{B}`, `{B}` |
| Swamp, an opponent's Mana Flare | `{B}`, `{B}` |
| Swamp, an opponent's Crypt Ghast | `{B}` (its "you" is not the payer) |

The solver, the generic recruit, the fit swap and the hand demand (ADR 0136's #3604 amendment) and the tiers all read slots, so they see the trigger's slots with no change of their own. A Swamp under Crypt Ghast is a two-mana source to all of them. Nothing changes a source's tier. A Treasure is still the last ordinary tier, and the planner now reaches it only when the lands really fall short.

#### 4. When the answer depends on a pick, the source is expanded per pick

For a source with a pick (a dual land, Birds of Paradise, a filter land's two picks), the planner judges each combination of its picks.

- **If every combination gets the same extra mana**, the extra slots are appended to the one candidate. Crypt Ghast on a Watery Grave adds {B} whichever colour the land makes, because the land is a Swamp either way.
- **If the combinations differ**, the source becomes one candidate per combination with its picks fixed. The candidates are mutually exclusive alternatives under `tapPlan.hasCard`, which is #779's and #2558's expansion. Mirari's Wake, Zendikar Resurgent, Mana Flare and Vorinclex add "one mana of any type that land produced", so they depend on the pick. So does Caged Sun (only when the land made the chosen colour), and so do Forsaken Monument and Ultima (only when {C} was made). The plan entry carries the picks its candidate fixed: `OneColor` for one pick, which the executor finds with `firstSlotOffering` (the #1222 path), and per-slot colours for more than one. The executor then makes the colours the plan priced, rather than working them out again.

**Rejected shortcut:** modelling "one mana of any type that land produced" on a `{W|U}` land as a second `{W|U}` slot. That would let one dual land and a Mirari's Wake pay {W}{U}. The card cannot do that: the Wake adds the type the land made. It is stronger than printed, which is the wrong direction (AGENTS.md §7).

How far to expand a source with several picks was the amendment's first question. [Owner decision 1](#owner-decisions-2026-10-30): every combination, capped at 36.

#### 5. A trigger's own choice is booked, and the executor makes what was booked

Fertile Ground's "one mana of any color", and Mana Flare after a land that made two types, are choices the trigger makes. They become multi-option trigger slots, and the solver books them like any other slot. The plan entry carries the colour booked for each trigger slot, in firing order (`plannedTap.TriggerColors`). The executor passes them to the adder in place of the greedy `pickColorForSlot`. It falls back to the greedy pick when the booked colour is not on offer (a stale plan), or when nothing was booked because the slot only paid generic.

Why not stay greedy, as today: the trigger fires right after its own land, so the greedy walk can see a pip that a later source in the plan was booked for. Take a cost of {U/R}{B}, a Forest under Fertile Ground and one Swamp. The plan books the trigger for {U/R} and the Swamp for {B}. The greedy walk picks {B} for the trigger, because {B} is the more restrictive pip. Then the Swamp's {B} has nothing to pay, and {U/R} goes unpaid. That is the #273 disagreement between the planner and the executor, and carrying the booked colour is how `OneColor` (#779) and `DifferentColors` (#2558) already answer it.

**Who chooses** is unchanged. For every printed card in the family, the player who adds the mana chooses, and that player is the payer. Fertile Ground says "its controller adds", meaning the enchanted land's controller. Mana Flare's ruling of 2023-09-01 says "you choose one type that was produced". The auto-tapper choosing for the payer is its standing contract ("no further player decisions"). It decides nothing for another player.

#### 6. When the trigger's controller is not the payer

The walk covers every permanent on the battlefield, as `fireManaTriggersLocked`'s does, not only the payer's. A trigger is controlled by its source's controller (CR 603.3a), but whose taps it sees is in its text, and its `AppliesTo` already says so:

- **An opponent's Mana Flare** ("whenever a player taps a land for mana, that player adds …", the text Heartbeat of Spring and Zhur-Taa Ancient also print) is counted. The mana reaches the payer and nobody can respond to it (CR 605.4a). The battlefield is public, so the plan uses no hidden information.
- **An opponent's Crypt Ghast or Mirari's Wake** ("whenever **you** tap …") is not counted, because its `AppliesTo` compares `prod.Controller` with its own controller. That is the executor's answer too, with no rule of the planner's own.
- **A Wild Growth an opponent controls on the payer's land** is counted: "its controller adds" means the enchanted land's controller.
- **The mana always goes to `prod.Controller`** (Decision 1: there is no recipient callback). The planner never counts mana for anyone but the payer. A card that ever sends the mana or the choice to someone else needs that field first, and it is not plannable until then.

#### 7. Every payer gets it, and the preview needs no change of its own

`autoTapTopUpLocked` is the one planning entry point (ADR 0118 §1). It serves the cast, an activation, a special action, an attack tax, a pay-unless prompt, the enumerator's affordability probe and the `/auto-tap-preview` endpoint. So the payment, the move list, the strict gate, the ready ring and the bots all agree on what is payable.

The preview (`GET /games/{id}/auto-tap-preview`) returns the plan itself (`plan`, plus the `sources` that `describePlanLocked` reads through the executor's own pickers, ADR 0011 #1285). On the #4716 board, the drag highlight and the preview rows name three Swamps and no Treasure, before anything is tapped. No wire field changes. The rows do not also say what the triggers add ([owner decision 3](#owner-decisions-2026-10-30)).

The preview is advisory. The cast plans again at the moment it pays (`applyAutoTapLocked`), so a Crypt Ghast destroyed between the drag and the release changes the payment. The preview was right for the board it was asked about.

#### 8. The clone check still guards the payment, and its fallback is today's plan

`planFundsLocked` runs the real executor on a clone before anything is tapped (#2461). The planner judges every source against the board before the payment, and the executor judges each source at its own tap. The two differ only when an earlier step of the same payment changes a trigger's condition. The likely case is a sacrifice in the plan that removes a trigger's permanent, or the land an Aura enchants.

When a plan that counts trigger mana fails the clone check, the payer plans once more without trigger mana (exactly today's planner) and checks that plan instead. Every board that pays today still pays, and trigger mana is counted whenever it will really arrive. The fallback can tap more than the first plan only on a board where the counted mana would not have arrived, and on such a board the extra tap is needed.

#### 9. What does not change

- **CR 106.7, "could produce".** Decision 5's third bullet stands. Trigger mana is not something the land can produce (Wild Growth's ruling, 2004-10-04), so an Exotic Orchard facing a Wild Growth Island still sees {U}. The view's `produced` and `color_options` do not include trigger mana. The planner's slots are a planning model, not a CR 106.7 answer.
- **Triggers that use the stack** (CR 605.5a). Electro's and Fire Nation Palace's "add mana" triggers fire on a cast and on an attack, and they resolve after the payment, when a player would next receive priority (CR 603.3). They are never counted. The planner reads only `ManaTriggersForCard` and the delayed mana records, so this holds by construction.
- **Replacement doublers** (CR 106.12b: Mana Reflection, Nyxbloom Ancient, Damping Sphere, Hall of Gemstone) are already priced by ADR 0013 §5ab Decision 3 and #3939. Item 2 judges triggers after that pricing, and item 3 prices the trigger's output through the same window, in the order the executor uses.
- **Triggered mana does not re-trigger** (Decision 2), so the planner counts no chains: the Wild Growth {G} that Mana Flare does not see is not counted twice.
- **Mana can still float.** A land cannot be half-tapped. Paying {G} with a Wild Growth Forest still leaves a {G} floating. What ends is tapping a second land, or cracking a Treasure, for mana the first land was already going to make.
- **Riders on the tap.** Winter's Night's "doesn't untap" rider and an opponent's Vorinclex trigger are not tiers today, and they still are not. Counting their lands' mana only means fewer of those lands get tapped.
- **The heuristic bot's own mana model** (ADR 0136 §2) reads the view, not the planner. Teaching it trigger mana is [#4770](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4770) ([owner decision 2](#owner-decisions-2026-10-30)).
- **Snapshots, events and the wire.** The plan is transient. No snapshot key, event kind or prompt is added.

### Rules

Each number was checked against the pinned text, the Comprehensive Rules effective September 25, 2026 (`MagicCompRules 20260925.txt`).

- **CR 106.12**: "To 'tap [a permanent] for mana' is to activate a mana ability of that permanent that includes the {T} symbol in its activation cost." Only a {T} ability is a tap, so a Gold, a Spawn or a Spirit Guide never gets trigger mana in the plan, as none fires in the pool (item 2).
- **CR 106.12a**: an ability that triggers on a permanent being "tapped for mana", or tapped for mana of a type, "triggers whenever such a mana ability resolves and produces mana or the specified type of mana". The trigger is judged on the production as it resolves, which is after the replacement window and once per production (item 2).
- **CR 106.12b**: a replacement that applies when a permanent is tapped for mana "modifies the mana production event while such an ability is resolving and producing mana". That is Mana Reflection. It is a replacement, not a trigger, and it is already priced. It does not touch the trigger's own mana, which is not a tap (item 3).
- **CR 605.1b**: a triggered ability is a mana ability if it does not require a target, "triggers from the activation or resolution of an activated mana ability … or from mana being added to a player's mana pool", and could add mana. The planner counts exactly what the engine fires, which is the tap form (§3, "Only a TAP fires").
- **CR 605.4a**: a triggered mana ability "doesn't go on the stack, so it can't be targeted, countered, or otherwise responded to. Rather, it resolves immediately after the mana ability that triggered it, without waiting for priority." The extra mana is in the pool in the middle of the payment, before the cost is paid, and nobody can take it away between the tap and the mana. Counting it is paying the way a player pays.
- **CR 601.2g and 601.2h**: mana abilities are activated before the cost is paid, and then "the player pays the total cost". Choosing which mana abilities to activate is the planner's job, and a player makes that choice already counting the mana the triggers will add.
- **CR 605.5a**: a triggered ability that could produce mana "but triggers from an event other than activating a mana ability", or has a target, is not a mana ability and uses the stack. It is never counted (item 9).
- **CR 106.7**: what a permanent "could produce" counts "any type of mana that an ability of that permanent would produce". Wild Growth's ability belongs to the Aura, so the land's answer does not change (item 9).
- **CR 603.3a**: "A triggered ability is controlled by the player who controlled its source at the time it triggered." The trigger's text, through its `AppliesTo`, decides whose taps it sees. The mana goes to the player the text names (item 6).

### Edge cases

| Case | What the planner does | Why |
|---|---|---|
| A trigger that could be responded to | It never counts one. A stack "add mana" trigger is not in `Spec.ManaTriggers`. | CR 605.5a. Its mana arrives after the payment. |
| The trigger's source leaves between the preview and the cast | The cast plans again when it pays. The preview reflected the board when it was asked. | Item 7. |
| A step of the same payment removes a trigger's source or host | The clone check catches it, and the payer falls back to today's plan. | Item 8. |
| Several triggers on one tap (Wild Growth and Utopia Sprawl on one Forest, Crypt Ghast and an opponent's Mana Flare on one Swamp, two Crypt Ghasts) | It counts all of them, in battlefield order, each judged on the same production. | The executor fires all of them in that order (§3). |
| A doubler and a trigger together | It doubles the printed mana, then judges the trigger once on the doubled production. The trigger's own mana is not doubled. | CR 106.12a, 106.12b; Mana Reflection's ruling. |
| "Any color" (Fertile Ground, Blighted Burgeoning) | Adds a five-option slot. The plan books it, and the executor makes the booked colour. | Item 5. |
| "Any type that land produced" on a land with a pick | Adds one candidate per pick. | Item 4. Modelling it as a free pick is stronger than printed. |
| "Any type that land produced" after a land made two types (a filter land's `{U}{R}`) | A trigger pick over the two types, booked like Fertile Ground's. | Mana Flare's and Mirari's Wake's rulings: "you choose one type that was produced". |
| A trigger whose output is empty (Utopia Sprawl with no colour chosen, Regal Behemoth with no monarch) | Counts nothing. | `Produced` returns `""`, and the executor adds nothing. |
| A spell-made turn trigger (High Tide, Bubbling Muck, Molten Tide's "you") | Counted, through the same pass one. | They are data records on `DelayedTrigger`, judged by the same function. |
| A source that is locked, excluded, summoning sick, or not plannable for any other reason | It has no candidate, so no trigger is counted. | Nothing is tapped, so nothing triggers. |
| Mana the plan does not need (paying {G} with a Wild Growth Forest) | Taps one Forest, and a {G} floats. | A land cannot be half-tapped (CR 106.4). |

### Calls made here

- **Ask the trigger, not a declared shape.** The owner's decision says the planner "checks the real trigger's own condition", and ADR 0013 §5ab Decision 3 already did the same for replacements (item 1).
- **An opponent's trigger that gives the payer mana is counted.** The mana arrives whatever anyone wants (CR 605.4a), and the battlefield is public (item 6).
- **The plan carries the colours it booked for a trigger's pick and for a pick it expanded,** and the executor makes them. Staying greedy can leave a hybrid pip unpaid (item 5).
- **The fallback is today's planner,** asked only when the clone check refuses a plan that counted trigger mana. Nothing that pays today stops paying (item 8).
- **Nothing changes a source's tier.** Trigger mana changes what a source makes, not what spending it costs. A Treasure is still the last ordinary tier (item 3).
- **One production per source, judged once,** exactly as the executor fires (item 2).
- **CR 106.7 stays as Decision 5 wrote it,** and the view's `produced` does not change (item 9).

### Owner decisions (2026-10-30)

The owner accepted this amendment on 2026-10-30 and chose the recommended option on all three questions below. The answers are binding, and the slices are written to them. The owner wants the build next.

1. **Every combination of picks is expanded, capped at 36** (question 1, option 1; item 4). Above the cap a source gets no colour-dependent extra, and a test names the shape.
2. **The bot's mana model learns trigger mana in a follow-up, [#4770](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4770)** (question 2, option 1): an additive view field, read by the bot behind a `Config` switch and measured in the arena (ADR 0136 §8). None of this amendment's slices changes the bot.
3. **The preview gets no new field** (question 3, option 1; item 7). Slice 4 is not planned.

### Questions for the owner (answered)

Each question lists the recommended option first. Where one option is more CR-faithful, that is the recommendation. The owner chose the recommended option for all three on 2026-10-30 ([Owner decisions](#owner-decisions-2026-10-30)). The questions are kept with the options not chosen.

1. **A source with more than one pick, whose trigger mana depends on the picks** (a filter land such as Flooded Grove, `{G|U}{G|U}`, under Mirari's Wake, Zendikar Resurgent, Mana Flare or Vorinclex).
   1. **(Recommended, the most CR-faithful)** Expand every combination of picks, as item 4 describes, with a cap that only bounds the search: 36 combinations. That covers every source with two picks (a filter land is 4, two five-colour picks are 25). Above the cap, the source gets no colour-dependent extra, and a test names the shape: Cascading Cataracts' "five mana in any combination of colors" under Mirari's Wake is the printed example. The plan counts what the cards add on every board short of that.
   2. Expand only a source with a single pick, and give no colour-dependent extra to any source with more than one pick. This is the decline-the-source posture of #2558 and #3939. It is simpler, but a filter land under Mirari's Wake is planned as two mana, and the planner may tap one land more than it needs.

   **Answered: option 1, as recommended (owner decision 1).**

2. **The heuristic bot's own turn plan** ([ADR 0136](0136-planning-the-turns-mana.md) §2) reads `produced` off the view, so it will still count a Swamp under Crypt Ghast as one {B}. CR 106.7 keeps trigger mana out of `produced`, so that field cannot carry it.
   1. **(Recommended)** Open a follow-up issue for an additive view field: the mana a tap of your own permanent would add through triggers, judged by the same pass one. The bot's mana model reads it behind a `Config` switch, measured in the arena as ADR 0136 §8 requires. Until then the bot's turn plan stays conservative. Once slice 2 lands, the enumerator offers every cast the trigger mana pays for, so the bot already casts those one at a time.
   2. Leave the bot's model as it is. It keeps under-planning two-spell turns on such boards, and the plan-miss counter does not show it, because the model never planned those turns.

   **Answered: option 1, as recommended (owner decision 2). The follow-up is [#4770](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4770).**

3. **Should the preview say what the triggers add?**
   1. **(Recommended)** No new field. The preview's rows say what paying with each source costs (ADR 0011, #1285), and trigger mana costs nothing. The drag highlight and the rows are already right, because they read the plan. The preview modal opens only for a refused cast that was not auto-tapped (ADR 0118).
   2. Add an additive `adds` field to each preview source: the mana its tap makes, triggers included (for example "{B}{B}"). The row would show it, so a player can see why three Swamps pay a six-drop. It is one wire field and one client row, built as a fourth slice.

   **Answered: option 1, as recommended (owner decision 3).**

### Slices

Each slice targets `develop`, references S78 and #4716, and runs `scripts/go-docker.sh test ./...` (the packages named are the ones that change). Slice 2 changes the engine, so it also runs the branch E2E. Slices 1 and 3 do not need it.

| # | Slice | Files | Tests | Model |
|---|---|---|---|---|
| 1 | **The judge.** Pass one of `fireManaTriggersLocked` becomes the read-only `manaTriggerFiringsLocked`. Add the `manaTriggersExistLocked` gate. Change the `ManaTrigger` doc to "either lock mode". No behaviour change. | `game/mana_trigger.go`, `cards/effects/mana_triggers.go` (doc), `docs/adding-cards.md` (the authors' three rules, item 1) | Existing `mana_trigger*_test.go` pass unchanged. `TestManaTriggerFiringsMatchWhatFires`: with stubbed triggers, delayed records included, the judge's list equals what the firing resolves. `TestEveryCatalogManaTriggerIsReadOnly` (`cards/effects`): for every `Spec` with `ManaTriggers`, on a board built for it, calling the judge under `ReadSnapshot` leaves the game's snapshot byte-identical. | Sonnet |
| 2 | **The planner counts it.** Trigger slots on candidates (`tapSource.TriggerSlots`). Per-pick expansion with the picks carried (`OneColor`, per-slot colours) and owner decision 1's cap of 36 combinations. Trigger output priced with `fromTap` false. `plannedTap.TriggerColors`, taken by the executor's adder. The fallback in every payer that runs `planFundsLocked`. | `game/autotap.go`, `game/autotap_topup.go`, `game/autotap_costed.go`, `game/mutations.go`, `game/mana_trigger.go`, `game/produce_mana.go` | The engine tests and the #4716 regression under [Tests](#tests-that-prove-it). Existing `autotap*_test.go` and `mana_trigger*_test.go` pass; a test whose expectation changes says why in the PR. | Opus |
| 3 | **Cards and docs.** Remove the "ADR 0074 §7" simplification from the comments of the 10 card files that cite it, and from the `mutations.go` and test comments. Replace the "Declared" paragraph of the triggered-mana recipe in `docs/adding-cards.md`. Add a closed-seam fragment under `docs/engine-seams/closed/`, never the generated block. The amendment is already Accepted (owner decisions of 2026-10-30), so this slice does not change its status. | `cards/effects/{blighted_burgeoning,crypt_ghast,fertile_ground,forsaken_monument,mana_flare,overgrowth,ultima_origin_of_oblivion,utopia_sprawl,wild_growth,winters_night}.go`, `cards/effects/mana_trigger_cards_test.go`, `game/mana_trigger_test.go`, `game/mutations.go` (comments only), `docs/adding-cards.md`, `docs/engine-seams/closed/` | `go test ./internal/roadmap/` (the closed-seams fragment) and `./internal/docsguard/` | Sonnet |
| 4 | *Not planned* ([owner decision 3](#owner-decisions-2026-10-30)). It would have added the preview's `adds` field and its row. | — | — | — |

### Tests that prove it

**The #4716 regression** (`cards/effects`, real catalog): `TestCryptGhastAutoPayTapsThreeSwampsForSixMana`. The board is the report's as far as it shows it: Crypt Ghast, five Swamps and a Treasure. Plan and pay {3}{B}{B}{B}, which is Veinripper's cost. Veinripper is not catalogued, so the cost is the input, and the test goes through the auto-tap payer (`applyAutoTapLocked`'s path), not only the planner. The test asserts:

- exactly three Swamps are tapped;
- two Swamps are untapped;
- the Treasure is still on the battlefield;
- the pool is empty after the payment.

A second case asks the `/auto-tap-preview` endpoint on the same board (`lobby`) and expects three Swamps in `plan` and no Treasure in `sources`.

**Engine tests** (`game`, stubbed catalog, beside `mana_trigger_test.go`):

- `TestPlannerCountsAFixedTrigger`: a Wild Growth-style {G} on a Forest, with two Forests out. {G}{G} taps one Forest.
- `TestPlannerJudgesTheTappedLandsSubtype`: a Swamp-only trigger adds nothing to an Island. A Swamp-typed dual gets the extra with no expansion.
- `TestPlannerCountsAnOpponentsThatPlayerTrigger` and `TestPlannerIgnoresAnOpponentsYouTrigger`: item 6.
- `TestPlannerExpandsAPickWhenTheTriggerReadsTheType`: a `{W|U}` land with "any type that land produced". It pays {W}{W} alone and does not pay {W}{U} alone.
- `TestTriggerPickMakesTheBookedColour`: {U/R}{B} with a Forest under a Fertile Ground-style trigger and one Swamp. The payment succeeds and no mana floats. It fails with the greedy pick.
- `TestDoublerThenTriggerCountsOnce`: a doubler and a Crypt Ghast-style trigger on one Swamp give three {B}, not four.
- `TestTriggerCountsOncePerProduction`: a "{C} was produced" trigger beside a `{C}{C}` source adds one {C}.
- `TestNoTriggerManaWithoutATap`: a sacrifice-only source and a hand source get no trigger slots.
- `TestPlannerFallsBackWhenTheTriggerManaWillNotArrive`: a plan whose own sacrifice removes the trigger's host fails the clone check and pays with today's plan.
- `TestPlanUnchangedWithoutManaTriggers`: on a board with no mana trigger, the plan is identical with the change and without it (the gate).
- `TestCouldProduceIgnoresTriggeredMana`: `ProducibleManaLocked` for an Exotic Orchard facing a Wild Growth Island is {U} (CR 106.7, item 9).

**Enumerator** (`legal`): `TestCastPayableOnlyWithTriggerManaIsOffered`. Crypt Ghast and three Swamps offer a {4}{B} spell under strict payment, and the cast succeeds.

**Catalog cards** (`cards/effects`): Mirari's Wake on a dual land, Caged Sun when the land did and did not make the chosen colour, Forsaken Monument with Sol Ring, Utopia Sprawl with no colour chosen, and Sasaya counting its same-name lands.

**Back-outs.** If the trigger pricing in the gather is dropped, the #4716 regression fails with five Swamps and the Treasure spent. If the executor goes back to the greedy pick, `TestTriggerPickMakesTheBookedColour` fails. If the fallback is dropped, `TestPlannerFallsBackWhenTheTriggerManaWillNotArrive` fails. If the gate is dropped, nothing fails; the gate only saves work, and `TestPlanUnchangedWithoutManaTriggers` shows that it is safe to drop.

### As built (2026-10-31)

Slice 1 is [#4772](https://github.com/krakenhavoc/cmd_and_ctrl/pull/4772) and slice 2 is [#4775](https://github.com/krakenhavoc/cmd_and_ctrl/pull/4775). Slice 3 brought the card comments, the recipe in `docs/adding-cards.md` and the closed-seam fragment in line with them, and changed no behaviour. Four things slice 2 settled that the sections above leave open or give differently:

- **The sort tiebreak.** `sortTapSources` scores a source's restrictiveness (`restrictivenessScore`) on its own slots only; its trigger slots are left out. A tie goes to the source with more trigger slots (`tapSource.TriggerSlots`). Scoring them would rank a Forest under Wild Growth (two slots) behind a plain Forest (one), and a cost of {G}{G} would tap both. With the tiebreak the Wild Growth Forest pays it alone. On a board with no trigger slot both keys are the old score and the order is unchanged.
- **Costed plans count trigger mana too.** The #2455 planner (`autotap_costed.go`) books against a source's trigger slots as well. `plannedTap.Slots` stays the source's own slots and the trigger's booked colours go to `plannedTap.TriggerColors`. When trigger mana funds a Signet's cost, the token names the trigger's permanent (`tapSource.TriggerFrom`), not the land, as the executor mints it.
- **A source gets no trigger mana in three shapes**, and is planned as it was before the amendment. Above the 36-combination cap of [owner decision 1](#owner-decisions-2026-10-30), whether or not the combinations differ. With two or more picks whose trigger mana depends on them, when a CR 106.12b production replacement changed its slots, because the planner's slot *k* is then no longer the executor's slot *k* and the picks cannot be pinned per slot. And a "N mana of one color" source with a further pick, a shape no printed card has. A source with one pick is carried as `OneColor` and is not affected by the replacement rule. When every combination gets the same extra mana, it is appended to the one candidate, as item 4 says.
- **The booked-pick example is {1}{U/R}{B}, not {U/R}{B}.** Item 5's example and `TestTriggerPickMakesTheBookedColour` differ by the {1}. With a cost of {U/R}{B} the Forest's own {G} would have nothing to pay and would float, which fails the test's "no mana floating" check for a reason that has nothing to do with the pick. With {1}{U/R}{B} the Forest's {G} pays the {1}, the trigger is booked for {U/R} and the Swamp for {B}; the greedy pick takes {B} and leaves {U/R} unpaid, as item 5 says.
