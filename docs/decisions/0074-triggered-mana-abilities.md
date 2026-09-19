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
counting triggered mana in the auto-tap PLAN (§7), triggers on mana produced
without a tap (§3)

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
  appearing halfway through a cast would break it.

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
  (AGENTS.md §7) and must not reach for `Spec.Triggered`. The distinction is
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
