# ADR 0019 — Structured targeting

**Status:** Implemented (sub-PR 1) · 2026-09-04 · Branch `feat/s20-target-predicates`

## Context

Until S20, "targeting" was a client hint. `Spec.TargetMode`
("creature", "player", …) told the Svelte picker which surfaces to
make clickable; the server accepted whatever ID came back. Negate
could counter a creature spell, Path to Exile could exile a land,
and a resolving spell only checked that its target still *existed*
(CR 608.2b's existence half), never that it was still a legal
target. The S19 trigger cards went further and auto-picked targets
server-side ("first opponent artifact") because there was no
picker at all for abilities.

The sprint's exit criterion — cast Doom Blade, see only non-black
creatures — needs the predicate to live on the server (it's the
rules authority and the only place with full information), and the
client to be told the answer rather than re-deriving it.

## Decisions

### 1. `game.TargetSpec` is the truth; `TargetMode` is derived

A card declares `Spec.Targets *game.TargetSpec`: which zones a
card target may live in, whether players qualify, a `CardOK` /
`PlayerOK` predicate over candidates (live game + caster +
candidate), a `Label` for the banner, and `Min`/`Max` (1/1 for
now). `TargetMode` is derived from `Targets.Mode` for the wire;
cards without a spec keep the S13.1 free-form behaviour — no
legality check — so nothing regresses for un-migrated cards.

### 2. Predicates are a small composable library in the catalog package

`effects/targets.go`: type predicates (`Creature`, `Artifact`,
`Instant`, …, all post-layer), colour (`OfColor`, `NonBlack`,
`Colorless`), stats (`PowerLE`, `ManaValueLE`), control/ownership
(`YouControl`, `OpponentControls`, `YouOwn`, `NotSelf`), plus
`And`/`Or`/`Not`. Constructors (`TargetAny`, `TargetCreature`,
`TargetPermanent`, `TargetSpell`, `TargetPlayer`,
`TargetCardInGraveyard`) fix the zone set and mode so a card file
reads like oracle text. Predicates run under the game lock and must
stay pure: the legal-target walk calls them per candidate on every
snapshot.

### 3. Colour comes from Scryfall, with a mana-cost fallback

`Card.Colors` is stamped at deck import from Scryfall's computed
`colors` (handles indicators, Devoid, hybrid). `EffectiveColors()`
falls back to the coloured symbols in `ManaCost` when the field is
empty (tokens, test fixtures). Layer 5 colour-changing effects
aren't modelled; `EffectiveColors` is the seam when they are.

### 4. Legality is enforced twice, with the same predicate

- **Announce (CR 601.2c):** `CastSpell` validates count and each
  ref against the spec → `ErrIllegalTarget` (the card stays in
  hand). The client picker only offers legal targets, so this is a
  backstop against stale snapshots and hand-built payloads.
- **Resolution (CR 608.2b):** `spellAllTargetsIllegalLocked` runs
  the predicate from the item's controller's point of view when a
  spec exists, so a Doom Blade target that turned black — or a
  creature that stopped being one — counters the spell by game
  rules, not only a target that left. Ability items (triggers)
  keep the existence check until they carry specs.

### 5. The legal set rides the snapshot, not a round trip

`CardView.legal_targets` (`{players, cards}`) is stamped in
`ViewOfGame` for every hand / command-zone card with a spec, from
its owner's point of view, and stripped from opponents' views in
`FilterViewFor` (even for revealed hand cards). The client
targeting store reads it into two sets; surfaces (battlefield
cards, player portraits, stack items) ask `isLegalCardTarget` /
`isLegalPlayerTarget`; legal cards get a green ring; the banner
shows "N legal". `canCastFromHand` denies a targeted spell whose
set is empty ("No legal target"). Cost: a few predicate walks per
snapshot for the handful of targeted cards in a hand — negligible
at eight seats.

Why not a `query_targets` action: a round trip before every cast
adds latency to the most common click in the game, and the same
data is what the S13.3 greyed-illegal-actions affordance needs on
every render anyway.

### 6. Triggers pick on the board too (sub-PR 2)

`TriggeredAbility.Targets` gives a trigger the same clause. The
harvester computes the legal set the moment the trigger fires (CR
603.3d: targets are chosen as the ability is put on the stack); an
empty set removes the trigger with no prompt at all — no more
"Yes" that silently does nothing, which is what the S19
`HasLegalTarget` warning papered over. A "you may" prompt still
comes first; on "yes" (or immediately for a mandatory trigger) a
`pick_target` pending choice carries the frozen legal set. The
client doesn't render it as a modal: it enters the ordinary
board-click targeting flow with the prompt's set (graveyard
targets via the zone browser), and the click answers
`resolve_choice {target}`. `ResolvePickTarget` re-validates the
pick against the live board, stamps it on the built item together
with the spec, and the item's resolution runs the same CR 608.2b
re-check spells get.

The prompt can't be cancelled from the client — the server owns a
trigger that needs a target — and it pre-empts any cast-targeting
prompt in flight.

### 7. Modes carry their own target clause (sub-PR 4)

A modal card declares `Spec.Modes` — a `game.ModeSpec` with the
prompt ("Choose one"), the option labels, `Min`/`Max`, and per
option an optional `TargetSpec`. The card-level `Targets` stays
nil: Rakdos Charm's first bullet targets a player, its second an
artifact, its third nothing, and there is no single clause that
describes the card. The engine derives the cast's effective spec
from the chosen options (`castTargetSpec`), so announce validation
and the resolution re-check are the same code paths a Doom Blade
uses; the option's spec is what the picker highlights.

Two consequences follow. First, an untargeted chosen mode must
arrive with no targets and a targeted one with exactly its count —
the modal card is never free-form. Second, sub-PR 4 supports **one
targeted option per cast**: "choose two" cards whose options each
target (Cryptic Command, Kolaghan's Command) need per-mode target
slots on the wire and on `StackItem.Targets`, which is the same
plumbing multi-target needs, so it ships with that work rather than
as a special case here. `Register` panics if a card declares more
than one targeted option with `Max > 1`, so the limit fails at boot
rather than at the table.

On the wire the owner's hand card carries `modes` with each
option's `target_mode` + `legal_targets`, computed from the same
snapshot pass as `legal_targets`. The client's cast flow is now X
prompt → mode picker → board targeting → `cast_spell {modes,
targets, x_value}`; each step is its own small surface rather than
one consolidated dialog, because each has exactly one thing to ask.
`canCastFromHand` greys a modal card when fewer castable options
than `Min` remain (targeted options need a legal target to count).

Effects read the choice back with `ctx.HasMode(i)` and resolve the
chosen bullets in printed order (CR 608.2c).

### 8. One clause, N slots; partial illegality resolves (sub-PR 5)

A multi-target clause is the same `TargetSpec` with `Min`/`Max`
set — "two target creatures" is one predicate chosen twice, "up to
two" is `0..2`, "any number" is `1..0` (unbounded). Announce
rejects duplicates unless `AllowSame` (CR 115.3), and preserves the
order the player clicked in, so a positional clause ("2 damage to
any target and 1 damage to any other target") reads slots by index.

The spell's item now remembers the spec it was announced under,
the way trigger items already did. That makes the per-slot check
cheap: `Context.IsTargetLegal` runs the announced predicate, not
just an existence check, and `Context.LegalTargets()` is the
surviving subset. The all-illegal short-circuit still fizzles the
spell before `OnResolve`; with one target left the spell resolves
and does what it can (CR 608.2b) — Ashes to Ashes still deals its
5 to the caster with one creature gone.

On the wire `legal_targets` and `pick_target` carry `min`/`max`.
The client keeps one targeting store: at `max 1` the first click
completes the prompt as before; otherwise clicks toggle into a pick
list (gold ring on picked cards / portraits), the banner shows
"n/max picked" and a Done button that enables at `min`, and Enter
confirms. Trigger prompts share the flow and answer with
`resolve_choice {targets: [...]}`.

## Out of scope

- **Per-mode target slots** for "choose two" cards whose options
  each target (Cryptic Command, Kolaghan's Command) — needs
  `StackItem.Targets` grouped by mode and a two-step picker.
- **Divided damage** (`Distribution` already rides the item; no
  UI or predicate yet) and cost-per-target clauses (Fireball's
  `{1}` per extra target: the X prompt runs before targets are
  known).
- **Hexproof / shroud / protection** as target-legality
  modifiers — the predicate hook is where they'll go.

## Amendment (2026-09-22, #1196): retargeting a spell or ability on the stack (CR 115.7)

Everything above is about the targets an announcement **chooses**.
CR 115.7 is about the targets an announcement **already has**: Deflecting
Swat, Bolt Bend, Misdirection, Ricochet Trap and Imp's Mischief all
reach onto the stack and rewrite a `StackItem.Targets` that somebody
else picked. `docs/engine-seams.md`'s *Stack-item retarget* row is that
gap, and it is the one row the 2026-09-16 audit singled out as
undercounted.

The whole of it is one observation: **retargeting is the announce gate
run a second time, for a different chooser.** Nothing about which
targets are legal changes when a Swat redirects a Bolt — the same
clause, the same `TargetSource`, the same protection and hexproof
check. What changes is *who is answering the question*. So this ships
as an entry point over the existing walk, not as a second walk.

### 1. `RetargetStackItemForEffect` is the one entry point

`game.RetargetStackItemForEffect(itemID, chooser, policy, next)`
([retarget.go](../../server/internal/game/retarget.go)) takes the
**whole new target list**, one-for-one with the item's current one, and
re-runs `validateAnnouncedTargetsLocked` over it with the item's own
announced clause list (`itemAnnouncedClauses`, ADR 0065 §1) and the
item's own `TargetSource` (`stackItemSourceLocked`). Everything the
announce gate enforces is therefore enforced here without being
restated: per-clause `CardOK` / `PlayerOK`, zone membership, `Min` /
`Max`, `AllowSame`, `Distinct`, and — through `targetLegalLocked` —
`CanBeTargetedBy`, which is protection, hexproof and shroud.

Three things are the retarget's own, and they are the only three:

- **One-for-one.** `next` must be the same length as the item's list
  and each ref must answer the same `(Mode, Slot)` step. CR 115.7
  changes *which* objects are targeted, never *how many*, and never
  which clause a slot belongs to. A mismatch is `ErrInvalidParam`.
- **A change count.** `RetargetChangeOne` ("change the target of…",
  CR 115.7b) permits at most ONE slot to differ; `RetargetChooseNew`
  ("choose new targets for…", CR 115.7c) permits every slot to differ.
- **An unchanged slot is waived.** CR 115.7c lets a player leave a
  target unchanged *even if it is now illegal*, so the legality check
  is run on the slots that CHANGED and waived on the ones that did
  not. That is the one new parameter on the announce walk:
  `validateAnnouncedTargetsWithLocked(src, steps, targets, waive)`,
  with `validateAnnouncedTargetsLocked` staying the nil-waiver form
  every announce path already calls.

### 2. The chooser is not the controller

This is the decision the rest falls out of. The `TargetSource` handed
to the check is the **item's** — `Controller: item.Controller`, and the
item's own card as the source object — while the player answering the
prompt is the **retargeting effect's** controller. A Deflecting Swat
pointed at an opponent's "target creature you control" can only move it
to a creature *that opponent* controls, and protection is still tested
against the redirected spell's colour rather than against the Swat.
Keeping those two players in different variables is the whole of it;
the catalog's ~300 `Targets:` clauses are untouched, exactly as they
were by #662.

### 3. What a retarget does not touch

`PaidCost`, `Modes`, `XValue`, `AltCost`, `Foretold`, `CastFromZone`
and `IsCopy` all stay: a retarget rewrites the choice of targets and
nothing else about the announcement (CR 115.7 changes no other
announce-time choice, and CR 608.2b re-reads the rest unchanged).

`Distribution` is the one field that is neither kept nor dropped but
**remapped**. It is keyed by target ID, so a slot that moves would
otherwise leave its portion of "divide 4 damage as you choose among
one, two or three targets" stranded on the old key. CR 115.7c says the
division can't be changed, so the portion travels with the slot:
`newDist[next[i].ID] = oldDist[old[i].ID]`, positionally.

### 4. The offer, and why "nothing happens" is an outcome

`OfferRetargetForEffect` is the card-facing half. It computes, per
slot, the legal set for that slot's clause **minus the target already
there** — "each target can be changed only to *another* legal target" —
and skips a slot whose alternatives are empty, which is CR 115.7a's
"if a target can't be changed to another legal target, the original
target is unchanged, even if that target is illegal". A spell or
ability with no targets at all cannot be retargeted and is
`ErrInvalidParam`; the card's own clause keeps it off the picker in the
first place.

`RetargetChangeOne` is refused on an item with more than one real
target. Every printed "change the target" also prints "with a single
target", so the restriction costs nothing and is enforced rather than
assumed — a multi-slot "change *a* target" would have to ask which slot
first, and there is no card to design that prompt against.

### 5. `retarget` is a prompt kind of its own, not a `pick_target`

The CR 707.10c copy re-target rides `pick_target` (S30, #95) on the
reasoning that the QUESTION is identical. That reasoning does not
survive the move to a stack item, for one reason: **the departure
table**. `pick_target` is `{reassign: true}` — a trigger's CR 603.3d
pick is about the whole board and a survivor can answer it. A retarget
is not: nobody else gets to aim another player's Deflecting Swat, and
CR 800.4 leaves the targets unchanged when the player who would have
changed them is gone. So `PendingChoiceRetarget` takes its own row,
`{}` — dropped, and the drop does nothing, which is exactly the
printed outcome.

It blocks the table (`choiceGateDecisions`, deny-by-default agreeing
with the explicit row): the retargeting spell is mid-resolution and the
item underneath it is still on the stack waiting to find out what it
points at.

The prompt carries **data, not a continuation**: the item id, the
policy, whether declining is allowed, and which slot is being asked.
Answering it rewrites an object that is already on the stack, so there
is no closure to resume and nothing for `ContinuationCensus` to count —
a paused retarget is a restorable snapshot, which is not true of any
other prompt in this family. A multi-slot "choose new targets" walk
applies each slot's answer as it is given and asks the next; a walk cut
short (the chooser leaves) leaves the slots already changed changed,
each of which was legal on its own.

The wire and the client reuse the `pick_target` projection and the
board picker unchanged — same `{players, cards, min, max}` shape, same
green-ring flow — and `internal/legal` gets its own case so the bot can
answer, ordered through #1014's `Options.OrderTargets` and offering the
decline when `min` is 0.

### 6. The copy's re-target moves onto the same check

`resolveCopySpellTargetsLocked` validated with `validateTargetsLocked`,
which is the announce gate — so it required every ref to be legal,
including one the player left alone, and it never checked that the
number of targets was the same. Both are wrong for CR 707.10c, which
is CR 115.7c by reference. It now calls the shared
`retargetCheckLocked` with `RetargetChooseNew` and the original item's
targets as the "old" list. The copy still BUILDS a new object rather
than rewriting one, so what is shared is the check, not the
application — which is the honest seam between the two.

### Out of scope

- **Targeting an ability on the stack.** `TargetSpell` walks
  `ZoneStack`, which holds spell cards; an ability item has no card
  there. Deflecting Swat and Bolt Bend therefore ship with the "or
  ability" half caveated (#259), and Misdirection, Ricochet Trap and
  Imp's Mischief — which print "target spell" — do not need it. This
  is ADR 0065's existing open item, not a new one.
- **A multi-slot "change a target"** — see §4.
- **Changing a target TO a named object** (Spellskite's "change a
  target of target spell or ability to this creature"). It is a
  different shape: no choice is offered, and the new target is the
  source itself.
- **Imp's Mischief's older printing**, whose "that spell's controller
  may have you lose life instead" was a second prompt addressed
  across the table. The current oracle text is an unconditional life
  loss and is what ships.

## Amendment (2026-09-23, #1211): targeting an ability on the stack (CR 115.4)

The amendment above ends with **"targeting an ability on the stack"**
in its Out of scope, and so does ADR 0065 §*Still open*, and so does
the `Modal and multi-target clauses` seam row. This closes it.

The gap was never about the RULE. CR 115.4 is one sentence — "some
spells and abilities can target abilities on the stack; such a spell
or ability can only target an activated or triggered ability" — and
the engine has been able to *counter* one since S13.1
(`counterAbilityLocked`). What it could not do was **name** one: every
target clause walked ZONES, the stack zone holds spell cards, and an
ability item is an entry in `Game.StackMeta` with a synthetic id and
no card anywhere.

#1223 built the half of the answer it needed for Strionic Resonator
(`TargetSpec.Abilities` + `AbilityOK`, recorded as decisions 12–15 of
[ADR 0043](0043-copy-effects.md)'s 2026-09-22 amendment). This
amendment says what generalising it to the whole catalog cost, which
was almost nothing, and states the three decisions that were actually
open.

### 1. A stack clause is TWO enumerations over ONE id space

`TargetSpellOrAbility` (`server/internal/cards/effects/stack_targets.go`)
sets `Zones: {ZoneStack}` **and** `Abilities: true` on one spec. The
zone walk finds the spell cards; `abilityItemsBySeqLocked` finds the
ability items; both append to `LegalTargets.Cards`, and both answer a
`TargetRef{Kind: TargetCard}`.

There is no fifth `TargetRefKind` and there will not be one. The wire
projection, the client picker, `internal/legal`'s enumerator, the
CR 601.2c gate and the CR 608.2b re-check all key on *"a uuid that has
to still be there"*, which is exactly as true of a stack item as of a
card; a new kind would have had to be taught to every one of them to
say the same thing. The two id spaces overlap **by design** — a spell
item's id IS its card's instance id — and `specMatchLocked` resolves
the ambiguity by checking `StackMeta` first and excluding
`StackItemSpell` there, so a clause admitting both still resolves each
ref to exactly one object.

**A mana ability is unreachable for free.** CR 605.3b keeps it off the
stack, so there is no item to enumerate and Stifle's, Disallow's and
Voidslime's printed "(Mana abilities can't be targeted.)" needs no
predicate, no check and no caveat.

### 2. The predicate vocabulary is over the ITEM, not over the card

A "target spell or ability" clause narrows both halves at once — Bolt
Bend's "with a single target" is a fact about the *announcement*, and
a spell and an ability have the same announcement record. So the
predicate the catalog writes is a `StackItemPredicate` over
`*game.StackItem`, and `TargetSpellOrAbility` runs it on **both**
halves: the ability half directly, the spell half after looking the
card's own item up in `StackMeta`.

That is what lets the three printed shapes be three call sites and no
new machinery:

| Printed | Written |
|---|---|
| "target spell or ability" | `TargetSpellOrAbility(label)` |
| "target spell or ability **with a single target**" | `TargetSpellOrAbility(label, ItemHasASingleTarget())` |
| "target **activated ability, triggered ability, or legendary spell**" | `TargetSpellOrAbility(label, AnyStackItem(AnAbilityItem(), ASpellItem(Legendary())))` |

`HasASingleTarget()` (the `CardPredicate`, #1196) stays for the cards
that print "target **spell** with a single target" — Misdirection,
Ricochet Trap, Imp's Mischief — and both read the same
`StackItemTargetCountForEffect`, so they cannot drift.

The spell half REFUSES a stack card with no `StackMeta` entry, which
is strictly narrower than the old `TargetSpell` walk and deliberately
so: the only object in that state is a spell that is *mid-resolution*
(the meta is deleted before the card is routed away), and nothing may
target it.

### 3. The CR 702 keyword gate stays off, and CR 701.5 is a deletion

An ability on the stack is neither a permanent nor a player, so
nothing about it can have hexproof, shroud or protection, and
`specMatchesLocked` applies `CanBeTargetedBy` only to the zone walk.
Protection still governs the ability's own SOURCE by way of
`retargetSourceLocked` when the ability is retargeted, which is the
one place it can matter.

Countering one is a **deletion, not a zone change** (CR 701.6a: "an
ability that's countered doesn't go anywhere"). `counterAbilityLocked`
drops the `StackMeta` entry and recomputes split second; there is no
`routeCardToZoneLocked` call, no CR 903.9 window, no flashback
override and no destination to redirect — which is why
`CounterTargetToZoneForEffect` treats an ability exactly as
`CounterTargetForEffect` does and Dest is meaningless for one.

Two things were wrong on that path and are fixed here rather than
worked around at the card:

- **`CounterAbility` (the public, locking entry) was a second copy of
  the body**, the #529 shape that let the two counter paths drift on
  flashback. It now delegates, so there is one deletion.
- **The emitted `EventCounterSpell` claimed the countered ability's
  own source as `Source`**, which `events.go` documents as *the
  counter itself*. The game log reads that field as "who countered
  it" and therefore printed *"Llanowar Elves countered
  `<uuid>`"* — the victim in the attacker's place and an unresolvable
  id in the victim's. The event now carries `Target` (the item),
  `CardID` (the permanent whose ability it was) and `Label`, and
  `protocol/log.go` renders the label, because a countered ability has
  no card of its own to name.

### 4. What needed nothing

Stated because the measurement is the point: `RetargetStackItemForEffect`,
`OfferRetargetForEffect` and `retargetCheckLocked` (#1196) are written
over `StackMeta` and `itemAnnouncedClauses`, both of which serve an
ability item already — so Deflecting Swat's and Bolt Bend's "or
ability" half needed **one word in a clause constructor** and not one
line of retarget code. `CounterTargetForEffect` has discriminated on
`item.Kind` since S13.1. `internal/legal`'s target enumerator and the
protocol's `stampLegalTargets` are id-agnostic. The stack overlay
already lights up any item whose id is in the legal set.

The client's share is **three strings**: the `stack_ability` and
`stack_item` target modes beside `stack_spell` (the banner's sentence
and `isTargetingStack`'s membership), and the cast flow's mode
allowlist, which is an allowlist and would otherwise have fired
`cast_spell` with no targets rather than opening the picker.

### Out of scope

- **Trickbind's second sentence** — "activated abilities of that
  permanent can't be activated this turn" is a per-permanent
  activation restriction with a turn duration, which
  [ADR 0073](0073-optional-additional-costs-and-the-cast-gate.md) §*A
  restriction with a DURATION* already prices. The card is not shipped.
- **Spellskite** — changing a target TO a named object, unchanged from
  the #1196 amendment's Out of scope.
- **"Counter target ability unless its controller pays"** — the
  `counter_unless_paid` guard is keyed to a spell's stack card
  (`server/internal/game/counter_unless_paid.go:105`); no catalog card
  in the gap needs it yet.
- **A per-KIND target mode on the wire.** `stack_ability` and
  `stack_item` are hints for the banner's sentence, not legality; the
  legal set is still the only thing the picker obeys.

## Amendment (2026-09-24, #1559): rules over the chosen SET of targets (CR 601.2c)

Everything above judges ONE candidate at a time: the clause's zones,
its `CardOK` / `PlayerOK` predicate, and the CR 702 keyword gate. Some
printed clauses constrain the picks against each other — "any number
of target creature cards that each have a **different mana value**"
(Agadeem's Awakening), "two target creatures **controlled by different
players**" (Run Away Together), "up to six target creature cards with
**different names**" (Behold the Sinister Six!). Whether such a pick is
legal depends on what else was picked, so no per-candidate predicate
can say it. Before this amendment every such card either skipped the
catalog or shipped a caveat that enforced the rule at resolution — a
mis-picked cast accepted and then wasted.

### 1. The rule is a pairwise-distinct KEY, not a predicate over the set

`TargetSpec.Different *TargetDifference{Label, Key}`: no two picks of
the clause may share the value `Key` reads off each card. The issue
proposed a `GroupOK func(chosen []Card) bool`; this is narrower on
purpose, because four readers have to answer the same question and
only one of them can run a Go predicate over a set:

| Reader | What it does with the key |
|---|---|
| announce gate (`validateAnnouncedTargetsWithLocked`) | refuses a set with a repeated key |
| CR 608.2b re-check (`TargetStillLegalForEffect`) | re-judges the surviving picks |
| enumerator (`internal/legal`) | never builds a set with a repeated key |
| client picker | greys a candidate whose key a pick already holds |

A key is evaluable by all four from one shipped map
(`LegalTargetsView.different.keys`); an arbitrary predicate would have
made the enumerator search subsets and left the client unable to grey
anything. Every printed "different X" clause found in the dump is a
pairwise-distinct key: mana value, controller, name. The catalog builds
them with `effects.EachDifferentManaValue()`, `EachDifferentController()`
and `EachDifferentName()`, attached with `.EachDifferent(…)`.

### 2. Announce: refused with the rule in the message

A set that breaks the rule is `ErrIllegalTarget` wrapped with the
printed rule — `illegal target: those targets must each have a
different mana value` — so the toast says which rule was broken
rather than "illegal target" about a set whose members are each fine.
The check runs over the whole list; under a CR 115.7 retarget, two
UNCHANGED slots that already collide are left alone (CR 115.7c lets
them stay even if illegal), but a changed slot may not land on a key
another slot holds, and the retarget offer leaves such cards out.

### 3. Resolution: judged over the survivors, and a collision drops BOTH

CR 608.2b says a target must still meet the targeting requirements; it
does not say how a requirement BETWEEN targets is apportioned. The
choice, stated so nobody reopens it:

- a pick that became illegal on its own (it left, it stopped being a
  creature) is dropped first and conflicts with nothing — so one card
  exiled in response to Agadeem's Awakening costs exactly that card;
- among the picks still individually legal, two that now share a key
  are **both** illegal. Neither is preferred: the constraint is
  symmetric, "the first one you named" is not a rule of the game, and
  it is the weaker of the available readings, never the stronger. A
  Run Away Together whose two creatures end up under one controller
  returns neither.

The re-check is `setRuleConflictLocked`, reached from
`TargetStillLegalForEffect`, so `Context.LegalTargets()`,
`IsTargetLegal` and the all-illegal fizzle all see it with no card-side
code.

### 4. Fillability counts keys, not candidates

"Two target creatures controlled by different players" with every
creature on one side of the table has two candidates and no legal
pair. The CR 603.3d check and the cast offer (`anyClauseUnfillableLocked`,
`queuePickTargetStepLocked`) count distinct keys, so such a trigger is
removed rather than prompting for an answer nobody can give.

### 5. "Mana value X or less" is a flag the engine binds, like CountFromX

Agadeem's Awakening's other half reads the X announced at CR 601.2b,
and a `CardOK` predicate is handed the game, the caster and the
candidate — never the announcement. `TargetSpec.ManaValueAtMostX` is
resolved exactly as `CountFromX` is: `bindStepsX` writes the announced
X onto the per-announcement clause COPY at cast, and
`itemAnnouncedClauses` binds it again from `StackItem.XValue` for the
CR 608.2b re-check. An unbound copy (the hand snapshot, built before X
exists) does not apply it; the view ships each candidate's mana value
beside `mana_value_at_most_x` and the client narrows by the X it
collected. `effects.Register` refuses the flag on a trigger's or an
activated ability's clause: a trigger announces no X, and the bot's
activation enumerator does not bind one.

### 6. What is NOT this seam

**"For each opponent, up to one target creature that player
controls"** (Molten Primordial) was filed against #1559 and triaged out
of it: it is one clause PER OPPONENT, each binding its pick to a
player, and `TriggeredAbility.TargetsFrom` already builds that list.
The difference is observable — a creature that changes hands in
response is an illegal target under the binding, even when its new
controller is another opponent — so Molten Primordial ships on
per-opponent clauses, not on `EachDifferentController`. **Windgrace's
Judgment** prints the same sentence on a SPELL, whose clause list is
static; it uses `EachDifferentController` and keeps a narrowed caveat
for exactly that difference.

### Out of scope

- **A set rule on a COST** — Transmutation Font's "sacrifice three
  artifact tokens with different names". The sacrifice validator is a
  separate walk (`docs/engine-seams.md`'s set-of-sacrificed-permanents
  row); `TargetDifference` is the shape it could reuse, and it does not
  read it yet.
- **Rules that are not pairwise-distinct** — "total mana value 4 or
  less" over a set of targets. `ChooseCardsPrompt.Validate` covers it
  for a non-targeting pick (#998); no printed TARGET clause in the
  catalog's reach needs it.
- **Binding a spell's pick to a chosen player** (Windgrace's Judgment's
  remaining caveat) — a per-target record on the stack item nothing
  else needs yet.

## Amendment (2026-09-28, #1723): an exact mana value, and X binding for activated abilities

§5 above closed with "`effects.Register` refuses the flag on a
trigger's or an activated ability's clause: a trigger announces no X,
and the bot's activation enumerator does not bind one." Lazav, the
Multifarious's `"{X}: Lazav becomes a copy of target creature card in
your graveyard with mana value X, ..."` needed both halves of that
sentence to stop being true for an ACTIVATED ability specifically —
triggers still announce nothing.

### 1. `ManaValueEqualsX`, the exact twin of `ManaValueAtMostX`

"With mana value X" is not "X or less": `TargetSpec.ManaValueEqualsX`
is the same announced-X mechanism (`xBound` / `xBoundSet`, written by
`bindStepsX` at announce and again from `StackItem.XValue` at
resolution) with `xBoundAdmits` comparing `==` instead of `<=`.
`effects.Register` refuses a clause that sets both — they are
different printed clauses, never one card's, and a spec with both
would have the second call silently overridden by the first the
reader checks.

**This is NOT monotonic in X, and that is the whole reason it needed
its own enumerator answer (§3).** Raising X does not grow the legal
set the way "or less" does — it usually REPLACES it, with a
completely different (possibly empty) set of cards.

### 2. X binding for an activated ability: the ability's OWN cost, not the spell path

CR 602.2b announces an activated ability's X the same way CR 601.2b
announces a spell's, so the engine machinery needed nothing new: the
announce path (`ActivateCatalogAbility`) already called `bindStepsX(steps,
params.XValue)` before this, and the resolution re-check
(`itemAnnouncedClauses`) already read `StackItem.XValue` for
`item.targetSpec`, an ability item's included — both were written
generically over "the announcement," never specifically over "the
cast." An X-bound clause on an activated ability was refused not
because the plumbing was spell-only, but because
`effects.Register`'s `checkNoXBound` panicked before any card could
reach it.

The fix is at that one boot check: `checkNoXBound` takes an
`xAvailable bool` the caller supplies — `a.Cost.DemandsX()` (the same
predicate `#1213`'s two-X-claimants check already reads: `{X}` in the
mana cost, or a variable sacrifice/tap count) for an activated
ability's clause, and its per-mode clauses; always `false` for a
trigger's, which still announces nothing whatever the card's other
abilities cost. A spell's top-level `Targets` was never routed through
this check at all — nothing changed there.

### 3. The enumerator cannot pick ONE X here — #544's rule over a pair

Every other `{X}` ability in `internal/legal` (`x_abilities_test.go`)
offers exactly one move, at the largest affordable X — sound for "or
less," because a bigger X only ever admits a superset. It is UNSOUND
for "exactly X": the largest affordable X (say 5, from five untapped
lands) can have zero graveyard creatures at mana value 5 while mana
value 0, 1 and 3 each have one, and "offer the largest X" would offer
NOTHING for an ability that plainly has three legal activations.

`abilityMovesForSource` (`internal/legal/abilities.go`) answers this by
trying every X from the announcement's floor (`enumeratedXFloor`, same
as the ordinary case) up to the largest affordable one
(`basePay.xValue`, already solved once for the ordinary case) and
keeping only the `(X, target)` pairs `legalStepSets` actually returns —
an X with no matching graveyard card contributes no move, exactly as a
target clause with no legal candidate always has. This is the pairwise
form of #544's rule ("never offer a move the engine refuses"), over the
announcement's two free variables instead of one.

Two things the ladder deliberately does NOT try to solve, because no
catalog card needs them yet and guessing would risk offering something
illegal: an ability whose PRICE reads its targets (`perTarget`), and an
X announced by a sacrifice or tap count rather than by mana
(`ab.Cost.XSlots() == 0`). Both cases are `continue`d out of — under-
enumerated, never over-enumerated.

### Cards

**Lazav, the Multifarious** (new, `CompletenessFull`). Its duration-
copy half — `BecomeCopy{Duration: CopyIndefinite}` — is
[ADR 0043](0043-copy-effects.md)'s amendment of the same date, not this
one; this ADR is only the target clause.

### Tests

`server/internal/game/target_set_test.go`'s
`TestManaValueEqualsXBindsToTheAnnouncedX` is `xBoundAdmits` /
`bindStepsX` in isolation, the same shape
`TestManaValueAtMostXBindsToTheAnnouncedX` already used.
`server/internal/cards/effects/lazav_the_multifarious_test.go` is the
card-level announce gate and the CR 608.2b re-check (a target's mana
value changing between announce and resolution).
`server/internal/legal/lazav_the_multifarious_test.go` is §3's
non-monotonic ladder, with `dispatchAll` as the soundness check every
enumerator test in that package carries.

### Out of scope (still)

Everything §6's "not this seam" list and the original "Out of scope"
section named is still out of scope. This amendment only widens WHO
may set `ManaValueAtMostX` / `ManaValueEqualsX`, not what either flag
can express.

## Amendment (2026-10-01, #1743): a retarget whose destination is fixed (CR 115.7b)

The #1196 amendment's Out of scope ended with **"Changing a target TO a
named object"** — Spellskite's "change a target of target spell or
ability to this creature" — and #1211's repeated it. Two printed
siblings say the same thing from an enters trigger: Mizzium Meddler
("you may change a target of target spell or ability to this
creature") and Hydroelectric Specimen ("you may change the target of
target instant or sorcery spell with a single target to this
creature"), which shipped with its trigger caveated for exactly this
gap.

It is the same observation one more time: **the destination is a
candidate answer, and the gate already knows how to judge an answer.**
So the variant is a field on the existing offer and a branch in the
existing answer, not a second retarget.

### 1. `RetargetOffer.To` pins the destination

`game.RetargetOffer` gains `To TargetRef`; its zero value (`Kind ""`)
is the free choice every earlier card prints. Card side,
`effects.ChangeTargets` gains `To` and `ToSource` — the latter is the
printed "to this creature", resolved to the ability's source
permanent at resolution.

A slot is **eligible** when the item's target list with that ONE slot
replaced by `To` passes `retargetCheckLocked` under
`RetargetChangeOne` — the check a free "change the target" answer
passes (`pinnedRetargetSlotsLocked`). So each of Spellskite's
2020-08-07 rulings is the gate's behaviour rather than new code:

- **Spellskite must be a legal target for the slot**, judged for the
  ITEM's controller (§2 of the #1196 amendment): the slot's clause and
  zones, and `CanBeTargetedBy` — so an opponent's spell cannot be
  pulled onto a hexproof Spellskite, and nobody's onto a shrouded one.
- **"If changing one target … would make other targets … illegal,
  that target can't be changed"**: the whole new list is validated, so
  `To` may not join a clause that already names it (unless
  `AllowSame`), nor land where a `Distinct` clause or a #1559 set rule
  forbids.
- **A slot `To` already holds is not a change** (CR 115.7a's
  "another legal target") and is never eligible.
- **No eligible slot is an outcome, not an error**: "You can activate
  Spellskite's ability even if Spellskite isn't a legal target … or
  even if that spell or ability has no targets. In this case, no
  targets are changed." Nobody is asked.
- **Spellskite leaving the battlefield** before the ability resolves
  changes nothing. A destroyed Spellskite is in no battlefield clause's
  zones, so the gate refuses it; a Spellskite that left and came back
  is a NEW OBJECT with the same instance ID (CR 400.7), which every
  clause would accept, so `ToSource` reads the ability's
  `SourcePermanent()` (#1418) and offers nothing when it has `Left`.

A pinned offer takes only `RetargetChangeOne` (`ErrInvalidParam`
otherwise — no card prints "choose new targets … to X"), and, unlike a
free one, over an item with any number of targets: the free variant's
single-target restriction exists because a multi-slot "change a
target" would have to ask which slot AND where, and a pinned one has
already answered where.

### 2. The one question is WHICH slot

"If the spell or ability has multiple instances of the word 'target,'
you choose which one target you're changing to Spellskite as
Spellskite's ability resolves" — and the same for several targets
under one instance (Deepglow Skate). So:

- one eligible OBJECT and a mandatory change: it is made, no prompt;
- otherwise — two or more eligible objects, or a printed "you may"
  (Mizzium Meddler, Hydroelectric Specimen) — a `PendingChoiceRetarget`
  whose options are the objects CURRENTLY in the eligible slots, with
  `min` 0 for a "you may" and 1 otherwise.

The prompt carries the new `PendingChoice.RetargetTo`, and that field
is the whole difference on the answer side: `ResolveRetarget` hands a
pinned prompt to `resolvePinnedRetargetLocked`, which reads the one
ref as "the slot currently holding this object" and changes it to
`RetargetTo`. Eligibility is recomputed against the board as the
answer arrives; with nothing eligible any more the prompt is dropped
and nothing moves, and an answer naming an object in no eligible slot
is `ErrIllegalTarget` with the prompt left open.

**It stays one prompt kind.** The question shape (one ref out of a
server-computed set, decline when `min` is 0), the departure row
(`{}` — dropped, targets unchanged), the board picker and the
`internal/legal` case are all exactly right as they stand; what the
options MEAN is the server's business. No wire field changes: the
client never needed `RetargetSlot` either.

`RetargetTo` is carried by the snapshot (additive within v7: its zero
value is the free retarget every older file holds; a pre-#1743 binary
cannot hold a game with a pinned prompt correctly, but every card that
can open one has catalog abilities that binary lacks, which restore
already flags as `AbilitiesLostOnRestore`).

### 3. Two instances of "target" naming the same object

CR 601.2c lets two DIFFERENT instances of "target" choose the same
object unless one says "another". The board can point at an object but
not at an instance of a word, so two eligible slots holding one object
are one option, and choosing it changes the EARLIER slot. It matters
only when the two clauses do different things to that object (Soul's
Fire aimed at one creature twice), and the result is weaker than
printed, never stronger. Spellskite and Mizzium Meddler declare it as a
caveat; Hydroelectric Specimen cannot reach it ("with a single
target") and is `full`.

### Cards

**Spellskite** (new, caveat §3), **Mizzium Meddler** (new, caveat §3),
**Hydroelectric Specimen** (caveat cleared, `full`).

### Tests

`server/internal/game/retarget_pinned_test.go` — legal destination
moves with no prompt (card and player), every refusal the gate makes
(predicate, hexproof for the item's controller, shroud, already the
target, left the battlefield), hexproof on the item controller's OWN
creature allowed, the which-slot prompt and its refusals, a change
that would break another slot, the same object in two clauses, the
"you may" decline, the destination leaving under an open prompt, the
item gone, the policy restriction, and the prompt surviving a
snapshot round trip. `server/internal/legal/retarget_pinned_choice_test.go`
— every enumerated answer accepted and moving its own slot, and the
optional decline. `server/internal/cards/effects/pinned_retarget_test.go`
— the three cards on a real board, including {U/P} paid with life and
with {U}, an activated ability redirected, Arc Trail's two instances,
and Spellskite destroyed or flickered in response.

### Out of scope (still)

- **"Change the target to enchanted creature if able"** (Captured by
  the Consulate) and **"to the player with the lowest result"**
  (Ricochet) — both can use `RetargetOffer.To`; neither card is in a
  batch yet, and Ricochet also needs its die-roll trigger.
- **Muck Drubb**'s "target spell that targets only a single creature"
  is a clause predicate over what the slot NAMES, not how many slots
  there are — the same gap the random-retarget row (#1815) records for
  Chef's Kiss.
- **A free multi-slot "change a target"** — still no printed card.

## Amendment (2026-10-07, #1863): targets described by what the source is doing in combat

"Target creature it's blocking" (Wall of Vipers, Goblin Snowman),
"target creature with flying blocked by this creature" (Whip Vine) and
"target creature that's blocking equipped creature" (Plasma Caster) name
a creature by its relation to the object whose ability this is. A
`CardPredicate` is handed the caster and the candidate and cannot say
it; `TargetSpec.RelativeToSource` (#2146) already established the
answer for power and toughness: **the relation is data on the clause,
judged in the one shared walk beside `CardOK`**, so the legal set, the
announce check (CR 601.2c), the resolution re-check (CR 608.2b), the
client's `legal_targets` and the bot's enumerated moves are one answer.
This amendment adds the second kind of relation to the same walk.

### 1. `TargetSpec.CombatWithSource`

`game.SourceCombat{Of, Role}`. The zero value says nothing.

- **`Of`** is whose combat the clause reads: `CombatOfSource` ("this
  creature") or `CombatOfHost` (the permanent the source is attached to,
  for an Equipment's "equipped creature" or an Aura's "enchanted
  creature"). Plasma Caster is the only host card today; Auras with the
  same wording would use the same value.
- **`Role`** is how the candidate relates to that creature: `CombatBlockedBy`
  (the candidate is an attacker the reference creature is blocking) or
  `CombatBlocking` (the candidate blocks the reference creature, which is
  itself attacking). The reference's blocked set is
  `Card.BlockedAttackers`, so a creature that blocks several attackers
  (Palace Guard) offers each of them.
- A creature that has left combat is no longer blocked or blocking (CR
  506.4): `CombatBlockedBy` requires the candidate to still be an
  attacker, and `CombatBlocking` requires the reference to still be one.
  That is what makes a Wall of Vipers whose target left combat in
  response do nothing at all (CR 608.2b), the Wall included.

### 2. A source that has left, and a source that was never there

An ability resolves after its source has gone (CR 608.2), and the
target is then judged by what the source last was (CR 608.2h).
`PermanentInfo` gains `Blocking []uuid.UUID`, written beside `Attacking`
when a permanent departs, and `TargetSource` carries the whole record
(`Departed`, `DepartedID`) for the one case `stackItemSourceLocked`
already identifies. So a Goblin Snowman destroyed in response to its own
ability still deals its damage to the creature it was blocking.

Two cases admit **nothing**, never everything: a walk with no source at
all (`SourceChooser`, a cost payment), and `CombatOfHost` whose host is
not on the battlefield (an unattached Equipment, or a host that has
left, which is read live rather than from its own record). Both are
weaker than printed, which is the direction a simplification may err
(#259). A TRIGGER whose source is already gone as its targets are chosen
has no record to read either; no card needs it yet.

### 3. Why not a closure, and why not `TargetsFrom`

A spec rides a paused prompt and a stack item, and a func there is a
restore-point blocker (`closure_fields.txt`), which is why
`RelativeToSource` is data and this follows it. A `TargetsFrom` that
reads the source's blockers reads the board when the trigger is built,
which ADR 0041 P9 forbids for a restorable trigger and which would
freeze the answer against a block removed in response.

### Cards

Wall of Vipers, Whip Vine and Goblin Snowman use `Of: source, Role:
blocked_by`; Plasma Caster uses `Of: host, Role: blocking`. Old Man of
the Sea needed no new targeting shape: "power less than or equal to this
creature's power" is the #2146 `PowerNoGreater` comparison, and its
duration is ADR 0109 §3's conjunction. All five are `Full`.

### Tests

`game/targets_combat_source_test.go` pins each (Of, Role) pairing, the
departed record, the no-source case and the left-combat case on the bare
walk; `legal/combat_relative_targets_test.go` pins that the enumerator,
the wire's `legal_targets` and the engine's dispatch agree; each card has
announce, resolution and response tests in
`cards/effects/combat_relative_targets_test.go`.

### Out of scope (still)

- **A trigger whose source has left before its target is chosen** reads
  no combat record (see §2).
- **Relations other than blocking** ("creature that's attacking you",
  "creature dealt damage by this creature") are different data and are
  not claimed by this field; `AttackingYou` already covers the first as
  a plain predicate because it names the caster, not the source.

## Amendment (2026-10-08, #2182): an ability that still resolves if its target becomes illegal (CR 608.2b)

CR 608.2b removes a spell or ability whose targets are ALL illegal on
resolution. Gilded Drake prints an exception: "This ability still
resolves if its target becomes illegal." Without it the Drake's "if you
don't or can't make an exchange, sacrifice this creature" never ran when
its only target left, so the card shipped stronger than printed.

**Decision.** `TargetSpec.ResolvesIfIllegal` (set with
`.StillResolves()` on the card's target clause) exempts an item from the
all-targets-illegal removal in `spellAllTargetsIllegalLocked`. It is a
clause-level field, so it rides the spec that every stamped, cloned and
restored item already re-derives (no new `StackItem` field, no snapshot
change), and it is read off the announced clauses, so it works for a
spell, a trigger or a modal option alike. Any flagged clause exempts the
item.

**The effect's half of the contract.** The engine no longer removes the
item, so the effect must treat an illegal target as unaffected:
`Context.IsTargetLegal` still answers the per-slot CR 608.2b re-check.
`ExchangeControl.ApplyAndReport` says whether an exchange happened
(CR 701.12b: false when either object is gone), which makes "if you
don't or can't" a plain branch.

**Out of scope.** Nothing here changes which targets are legal at
announce (CR 601.2c) or the partial-illegal case, which already resolved.
Only Gilded Drake prints the sentence in the current Scryfall dump.

## Amendment (2026-10-13, #3259): a card in exile as a target, and "exiled with this" (CR 607.2a)

"Put **target** creature card exiled with The Darkness Crystal onto the
battlefield" names a card in exile by its link to the object whose
ability this is. Two halves were missing, and the card shipped with a
caveat (the card was chosen as the ability resolved, so it could not be
responded to with the choice in view). `TargetSpec.Zones` already took
`ZoneExile` and the engine walked it, but nothing could answer the clause:
the client's picker did not open for it, and `CardOK` is handed the caster
and the candidate, never the source.

### 1. `card_in_exile`, the exile sibling of `card_in_graveyard`

`effects.TargetCardInExile(label, preds...)` declares `Zones {ZoneExile}`
and `Mode "card_in_exile"`. Nothing in the legal-set walk changed for it:
the legal set, the announce check (CR 601.2c) and the resolution re-check
(CR 608.2b) were already one walk over `Zones`. The client learns the mode
(`opensTargetPicker`, the banner sentence, the free-form fallback) and
answers it through the zone browser's exile tab, whose cards already route
clicks to the targeting store. A card is in one player's exile pile, so
the viewer opens the pile that holds it, as for graveyards; the banner
says so. `effects.FaceUp()` is Pull from Eternity's "face-up".

### 2. `TargetSpec.ExiledWithSource`

A bool on the clause, built with `effects.ExiledWithThis(spec)`. It is
judged in `sourceAdmits`, beside `RelativeToSource` and `CombatWithSource`,
for the reason those are data: a spec rides a paused prompt and a stack
item, and a func there is a restore-point blocker. The comparison is
`Card.ExiledWith == {source id, source epoch}`:

- the live source's own `ObjectEpoch`, so a second Crystal's exiles are
  not offered and a Crystal that left and came back (a new object, CR
  400.7) has no claim on what its earlier self exiled;
- a source that has left is read from its last-known record
  (`TargetSource.Departed`, CR 608.2h), so an ability already on the stack
  still reaches the cards its source exiled;
- a walk with no source object (`SourceChooser`, a snapshot-only source)
  and the zero link admit **nothing**, never everything.

New `TargetSource.ObjectRef` is the {instance, epoch} read. No snapshot
shape, no wire field and no closure changed: `TargetSpec` is re-derived
from the catalog, and `Card.ExiledWith` was already carried by the
snapshot (ADR 0066's 2026-10-07 amendment).

### 3. A face-down exiled card is never offered to someone who has not seen it

A clause over exile can reach foretold, hidden or plotted cards. The legal
set crosses the wire and the bot's move list, and a predicate that said
"creature card" of a card its chooser may not look at would tell them its
type. The shared walk and the announce check now drop a face-down card in
exile the chooser is not a knower of (`hiddenExileCard`). Face-up exile is
public and unaffected.

### 4. Who stamps the link

Only a replacement exile stamps `Card.ExiledWith` today
(`ReplacementEvent.ExiledWith`), which is The Darkness Crystal's shape. An
ability-resolution exile (Duplicant, Angel of Serenity) is still found
through the effects package's event-log record, and ADR 0155's planned
stamp is unchanged and still its own work. A clause that wants those cards
waits for it; this amendment adds the clause, not the stamp.

### 5. The bot

The enumerator needed nothing: it reads the same legal set. The heuristic
priced an exile target as 0, so it could not tell a 6/4 from a 1/1 for
"take it" effects; `cardTargetValue` now prices a card in exile as
`cardValue`, whoever owned it. A hate effect over exile (Pull from
Eternity) scores the same, which costs the bot at worst one cheap card
spent on a good target.

### Cards

The Darkness Crystal (`Full`: the ability now targets) and Pull from
Eternity (`Full`). No other catalog card had a "choose on resolution"
caveat or comment standing in for an exile target; the Bronzebeak
Foragers, Gelatinous Cube and Nyla family wait on the ability-exile stamp
(section 4), not on this clause.

### Tests

`game/targets_exiled_with_test.go` pins the bare walk (object identity,
stale epoch, no source, the resolution re-check, a departed source, a
hidden card, a Clone); `legal/exile_targets_test.go` pins that the
enumerator, the wire's `legal_targets` / `target_mode` and the engine's
dispatch agree; `cards/effects/kuja_b_crystal_test.go` and
`pull_from_eternity_test.go` pin announce, resolution and response;
`aiseat/heuristic/exile_target_test.go` pins the pricing; the client
suite pins the mode vocabulary.

### Out of scope (still)

- The ability-exile `Card.ExiledWith` stamp (ADR 0155 PR 3).
- A pile-level hint on the table ("2 legal targets in this exile pile").
  The dock's "N legal" count and the sentence "a card in exile (open a
  player's exile pile)" say it; a per-pile ring would be client polish.

## Amendment (2026-10-15, #3846): "target ... exiled this way" -- a reflexive trigger's target is limited to its payload (CR 603.12)

Espers to Magicite: "Exile each opponent's graveyard. When you do, choose up to one target creature card exiled this
way." The reflexive trigger's own target clause had no way to say "exiled this way": the set is built by the parent
while it resolves, so no property of a candidate card names it, and the card shipped choosing at resolution (so the
table could not respond to the choice). `ExiledWithSource` is no help because the parent is a SPELL, and a spell is
not a permanent: `SourceObjectForEffect` refuses it, so there is no CR 607.2a object to stamp an exile onto.

`TargetSpec.AmongPayload` (bool) is the clause, built with `effects.AmongTheCardsNamed(spec)` and declared on the
reflexive body's `ReflexiveBody` registration like any reflexive clause. The set is `TargetSpec.AmongCards`, derived
data that a card never sets: `QueueReflexiveTriggerForEffect` fills it from the trigger's payload (`ReflexiveTrigger.Cards`)
when the trigger is created, and restore fills it again from the item's payload when it re-derives the clause
(`stampReflexivePayload`), so a table holding the trigger on the stack is still a restore point. It is judged in
`sourceAdmits` beside `ExiledWithSource` and `PaidForSource`, so the legal set, the pick (CR 603.3d), the resolution
re-check (CR 608.2b), the picker and the bot's moves are one answer. A clause with the flag and an empty payload
admits nothing. No wire change: the picker sees the ordinary legal set. Espers pairs it with
`TargetCardInExile(..., Creature()).WithCount(0, 1)` ("up to one").

Back-out: drop the flag in `AmongTheCardsNamed` and `TestEspersTargetIsLimitedToTheCardsExiledThisWay` fails (a
creature card exiled before the spell is offered); stop stamping the payload on restore and
`TestEspersTargetedTriggerSurvivesASnapshotRestore` fails.

## Amendment (2026-10-15, #3860): which instance of "target" changes, and a clause relative to the trigger's event

Two gaps in the targeting machinery, both found by the zero-caveats pass on six cards.

**1. A retarget whose destination is fixed asks WHICH instance of "target" (CR 115.7b, CR 601.2c).**
The 2026-10-01 amendment above declared a caveat: when the same object fills two instances of the word "target"
(Azorius Ploy at one creature twice; CR 601.2c allows it unless the second says "another"), the object is one thing to
click, and the engine changed the EARLIER slot. Which slot becomes Spellskite matters when the two sentences do
different things, and Spellskite's 2020-08-07 ruling puts the choice with its controller.

`finishPinnedRetargetLocked` (game/retarget.go) is now the one place a pinned retarget is carried out once the OBJECT is
settled — whether the controller named it in the `retarget` prompt or it was the only one the board could offer. One
slot holds the object: it changes, no prompt. Several do: a `PendingChoiceOptionPick` with one option per slot, each
labelled "Change <the clause's own words>" (with "(first)" / "(second)" when the words repeat) and showing the creature
in `Cards`. The continuation is a KEYED option pick (`retarget/pinned-slot`, option_pick_keyed.go) carrying the stack
item, the object and the destination as plain ids, so a table waiting on the answer is still a restore point, as the
object prompt before it always was. The slots are RECOMPUTED when the answer arrives, never trusted from the offer; a
prompt that ends unanswered (its chooser left, CR 800.4a) changes nothing, which is CR 115.7a's "a change that is not
made leaves the target alone". No wire change: `option_pick` already has a client, a bot policy (the first option is
the always-legal one) and an enumerator row. Spellskite and Mizzium Meddler lose their caveat and ship `full`.

**2. A target clause that is relative to the trigger's event: "target artifact that opponent controls".**
`TargetSpec` clauses could be relative to their SOURCE (`RelativeToSource`, `CombatWithSource`, `ExiledWithSource`) but
not to the EVENT that fired the ability. A mode's clause is fixed on the card before the trigger exists, so Parapet
Thrasher's "Destroy target artifact that opponent controls" — "that opponent" being the player the Dragons just dealt
combat damage to — offered every opponent's artifacts and wasted the bullet at resolution.

`TargetSpec.ControlledByThatPlayer` (bool, built with `effects.ControlledByThatPlayer(spec)`) is the clause, judged in
`sourceAdmits` beside the other relative gates. The event is `TargetSource.Trigger`, set by `WithTrigger` at every point
a trigger's walk builds its source — the harvest's fillable check, the mode prompt, each pick step and its answer — and
by `stackItemSourceLocked` from the item's carried context, so the legal set, the announce check (CR 601.2c), the
resolution re-check (CR 608.2b), the picker and the bot's moves are one answer. A mode with no such artifact is not on
offer at all (CR 603.3c), which leaves the other two bullets for the turn's remaining triggers. A walk with no trigger
admits NOTHING, never everything. It is data on the spec, not a closure, for `RelativeToSource`'s reason: a spec rides a
paused prompt and a stack item.

Back-out: take the `min(len(slots), 1)` shortcut in `finishPinnedRetargetLocked` and
`TestPinnedRetargetSameObjectInTwoClausesAsksWhichSlotChanges`, `TestSpellskiteAsksWhichInstanceOfTargetChangesToIt` and
`TestMizziumMeddlerAsksWhichInstanceOfTargetChanges` fail; drop `thatPlayerAdmits` from `sourceAdmits` and
`TestParapetThrasherOffersNoArtifactBulletWithoutThatOpponentsArtifact` fails; drop the item's trigger from
`stackItemSourceLocked` and the resolution re-check fails (`TestParapetThrasherDestroysTheDamagedOpponentsArtifact`).

## Amendment (2026-10-16, #1815): targets chosen at random

Every target so far was chosen by a player: CR 115.7's retarget through the `retarget` prompt, a trigger's CR 603.3d
target through `pick_target`. Chef's Kiss ("Copy it, then reselect the targets at random for the spell and the copy.
The new targets can't be you or a permanent you control.") and Witch Hunt ("target opponent chosen at random") need the
ENGINE to pick. Both picks are the same act: a uniform draw over the legal set the prompt would have offered, from the
keyed "pick" stream (ADR 0054) keyed to the player and the source, so an undo replays it. Nothing new is persisted.

**1. Reselecting a spell's targets at random (CR 115.7).** `Game.ReselectTargetsAtRandomForEffect(RandomRetarget)`
(game/random_retarget.go) is the 2026-09-22 retarget with the chooser replaced by the draw; legality is unchanged and
judged for the ITEM's controller. Three readings, each from the rulings:

- The draw is over ALL legal targets, the current one included. "Reselect" is not CR 115.7a's "change to another":
  Grip of Chaos's reminder "(Select from among all legal targets.)" and its ruling "it may still choose the original
  one". The other reading (exclude the current target) would make the card stronger than printed.
- Each instance of the word "target" is drawn on its own, in announce order ("a new target will be reselected at random
  for each instance of the word target", Chef's Kiss 2021-06-18). A candidate is kept only when the list with it in that
  slot passes `retargetCheckLocked`, so the duplicate rule, Distinct, the set rules and the relations hold and the final
  list is legal (CR 115.7e).
- A slot with nothing that could be selected keeps its target, even an illegal one ("If there are no legal targets that
  could be selected, the targets remain unchanged"). `RandomRetarget.Exclude` is "can't be you or a permanent you
  control": that player and the permanents they control are never drawn.

The rewrite goes through `RetargetStackItemForEffect`, so the new targets become targets (CR 115.7) and a division
travels with its slot (CR 115.7f). "Targets only a single permanent or player" is CR 115.9c, not `HasASingleTarget`'s
slot count: `Game.StackItemSoleTargetForEffect` counts DIFFERENT objects and players, and the card predicate checks
that one's current state (CR 115.9b). `CopySpellKeepingTargetsForEffect` makes "copy it" (CR 707.10, no new targets)
and returns the copy's id so the card can reselect for it next.

**2. "Target … chosen at random" on a trigger (CR 603.3d).** `TargetSpec.AtRandom` (`effects.ChosenAtRandom`) is data
on the clause, for `RelativeToSource`'s reason. The trigger target walk (`queuePickTargetStepLocked`) computes the
step's legal set exactly as for a prompt and, for a random clause, draws one target from it instead of queuing
`pick_target`; with no legal target a required clause removes the ability as before. A reflexive trigger's clause is
walked the same way. `effects.Register` refuses the flag on a spell's or an activated ability's clause (their targets
are announced by the player in one message, CR 601.2c / 602.2b; Goblin Test Pilot waits on that, #4051) and on any
clause but one undivided target, the only shape printed.

**Owner decision not needed.** The rulings settle each reading above, and each is the rules-faithful option.

**Cards:** Chef's Kiss, Witch Hunt, Scab-Clan Giant, Knight Rampager and Power Pack, all `full`. Grip of Chaos waits on
a trigger for a triggered ability put on the stack (#4050); Explosion of Riches is card work (#4052).

Back-out: leave the current target out of the draw and `TestChefsKissCanRedrawTheCurrentTarget` fails; drop the
exclusion and `TestChefsKissNeverPicksAPermanentYouControl` fails; move a target with nothing selectable and
`TestChefsKissKeepsTargetsWhenNothingElseCanBeSelected` fails; queue the `pick_target` prompt for a random clause and
`TestWitchHuntGoesToAnOpponentChosenAtRandom` fails; draw from a private source and
`TestChefsKissReplaysTheDrawAfterUndo` fails.

## Amendment (2026-10-16, #2055): a target chosen by an opponent ("target … of an opponent's choice")

Ten Commander-legal cards print a target "of an opponent's choice" (Nova Pentacle, Evangelize, Preacher, Arena, Magus
of the Arena, Cuombajj Witches, Diaochan, Volcanic Offering, Mausoleum Turnkey, Karplusan Minotaur). Until now every
target was chosen by the player putting the spell or ability on the stack. The rulings agree on the shape: as the spell
or ability is put on the stack its CONTROLLER chooses an opponent, and THAT opponent chooses the target (Nova Pentacle,
Evangelize, Preacher; CR 601.2c, via CR 602.2b for an activation and CR 603.3d for a trigger). The controller still
controls it and is still the player targeting, so an opponent cannot name their own hexproof creature (Cuombajj
Witches). With several such clauses the controller may name the same opponent or different ones (Volcanic Offering),
and the choice of opponent is part of the announcement: a retarget keeps it, and "they control" then means the chosen
opponent (Magus of the Arena, Evangelize).

**1. Data on the clause.** `TargetSpec.ChosenByOpponent` (built with `effects.OfAnOpponentsChoice`) marks the clause;
`TargetSpec.ControlledByChooser` (`effects.TheyControl`) is "… they control". Legality is judged for the controller
in the one shared walk exactly as for any other clause; `chooserAdmits` beside the other relative gates reads the opponent
bound onto the per-announcement clause copy (`bindStepChoosers`, like the X bound). Unbound — the legal set a snapshot
ships before anyone is named — "they control" admits any opponent's creature: the superset, as an unbound X is.
`StackItem.TargetChoosers` (`[]TargetChooser{Mode, Slot, Player}`) records who chose each such clause; it is carried by
clone, snapshot (additive under v9, omitted when empty) and a CR 707.10 copy, and `itemAnnouncedClauses` binds it, so
the CR 608.2b re-check and a CR 115.7 retarget read "they control" against the opponent the announcement named.
`effects.Register` refuses "they control" alone, an opponent's clause on a mode, a gift or an optional cost's rewrite,
and one that divides, relates, sets a set rule, is bounded by X or is "up to" — no printed card is any of those.

**2. A spell or an activation is PARKED, the shape #1397 built.** The controller's payload names targets for their own
clauses only (`StepsChosenByController`); a ref aimed at an opponent's clause is refused. `castSpellLocked` and
`activateCatalogAbilityLocked` validate the announcement with the opponents' unanswered clauses left out
(`splitOpponentTargets`), then park it where the CR 903.9 commander question parks, before anything is paid or moved
(CR 601.2c precedes CR 601.2h). For the first unanswered clause the engine lists the opponents who could fill it (an
opponent with no creature cannot choose "a creature they control"; none ⇒ the announcement is refused, as a spell with no
legal target is). One ⇒ that opponent is named. More ⇒ the controller is asked with an ordinary choose-a-player option
pick. The named opponent then gets an ordinary `pick_target` prompt over the clause's legal set (judged for the
controller, bound to them), with `PickTargetFor` naming the controller. Each answer makes the whole announcement again
with what has been chosen attached (`CastSpellParams.opponentPicks`, unexported, set only by the resume), so the re-run
re-validates everything against the board as it is then; the final run pays and puts the item on the stack like any
other. A re-run that no longer validates is the controller's log line, not the answering opponent's error. A second
announcement of the same spell or ability while one is parked is refused (`ErrChoicePending`). The frame
(`opponentTargetFrame`) is plain data, so the closure ratchet gains no route; a table waiting on it is not a restore
point (`ContinuationCensus.ChoiceResumeFrames`), as no parked announcement is.

**3. A trigger walks it.** A trigger's targets are already a server-side walk, one `pick_target` per clause
(`pickTargetFrame`). An opponent's step adds the choose-an-opponent question in front of it (that option pick carries
the walk as a marker, so the drain holds the batch behind it and an ended turn drops it with the walk) and sends the
step's prompt to the named opponent; `ResolvePickTargets` judges the answer for the trigger's controller with the named
opponents bound, and the item is stamped with `TargetChoosers`. No opponent able to choose a required clause removes the
trigger (CR 603.3d).

**4. Departures (CR 800.4).** A controller who leaves takes their announcement with them: an opponent's open pick for it
is withdrawn and nothing runs (CR 800.4a, 800.4d). A choosing opponent who leaves is handled by the existing
reassignment of a `pick_target` (CR 800.4g's "another opponent if possible"); the answer records whoever answered.

**5. The wire, the client and the bot.** `LegalTargetsView.chosen_by_opponent` marks the clause: the client's walk skips
it (every other step keeps its own slot), and a statement made only of such clauses opens no walk, so the announcement is
sent at once. The opponent's prompt is an ordinary `pick_target` whose `pick_target.chosen_for` names the controller,
answered on the board like any other. The enumerator names no target for an opponent's clause and offers the
announcement only while some opponent could complete it (`OpponentStepFillableForEffect`). A bot choosing for somebody
else's effect gives up the least: its own permanent is priced at what losing it costs and a card the controller gets back
at what it is worth to them; anything else is priced as its own pick would be, which aims it away from itself.

### Cards

Nova Pentacle, Evangelize, Arena, Magus of the Arena, Cuombajj Witches, Volcanic Offering and Mausoleum Turnkey (Full).
Left out: Preacher (the ability does nothing if Preacher was untapped at ANY point after activation; the engine reads
"remains tapped" as the ability resolves, the gap Vedalken Shackles' family shares), Diaochan, Artful Beauty ("activate
only during your turn, before attackers are declared" has no activation window), and Karplusan Minotaur ("whenever you
lose a coin flip"). Filed as #4061, #4062 and #4063.

### Tests

`cards/effects/opponent_target_cards_test.go` drives every card through the real announce, prompt and resolve path:
nothing paid while parked, a payload naming the opponent's target refused, only opponents who could choose offered,
hexproof judged for the controller, "they control" re-checked at resolution, two clauses with two opponents, and the
trigger walk. `legal/opponent_target_test.go` holds the enumerator to the controller's own clause, the view to
`chosen_by_opponent` and `chosen_for`, and the opponent's owed pick to their own creatures. The client's
`opponentChoiceTarget.test.ts` holds the walk.

Back-out: drop `chooserAdmits` from `sourceAdmits` and the Evangelize and Arena tests offer every opponent's creature;
skip the park and the Nova Pentacle test sees the cost paid before the target; drop `bindStepChoosers` from
`itemAnnouncedClauses` and Evangelize takes a creature its opponent no longer controls.

## Amendment (2026-10-19, [#4128](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4128)): "for each opponent, … target … that player controls", required

Blatant Thievery ("For each opponent, gain control of target permanent that player controls") needs one REQUIRED pick per
opponent (CR 601.2c). The "up to one" form already existed (`perOpponentClauses`, Windgrace's Judgment): one clause per possible
opponent bound to a seat, every clause optional so a seat that is not at the table offers nothing. A required clause cannot
borrow that shape. At a two-player table the clauses for the second and third seat could never be filled, so the spell would be
uncastable; and making them "up to one" would let the caster skip an opponent who has a legal permanent, which is stronger
than printed.

**Decision.** `TargetSpec.OpponentSlot int` (1-based, zero is no slot) binds a clause to the n-th seat after the chooser's in
turn order, as DATA. Two readings, both through one place (`game/opponent_slot.go`):

- **Who may be picked.** `opponentSlotAdmits`, folded into `sourceAdmits` beside `chooserAdmits`: the candidate must be
  controlled by the live player in that slot. A seat with no player, or an eliminated one, admits nothing. Because it lives in
  the shared walk, the legal set, the announce check, the CR 608.2b re-check, the client's picker and the bot's moves are one
  answer, and a permanent that changed hands in response is illegal for its slot even when it went to another opponent.
- **Whether the clause is required.** `Game.ClauseMinForEffect` is the clause's `Min`, except 0 when the slot's seat holds no
  live opponent (an eliminated player is no longer an opponent, CR 800.4a). The announce gate, `anyClauseUnfillableLocked`
  (the cast offer), the wire view (`viewOfTargetClause` stamps `legal_targets.min` and each clause's `min`) and the enumerator
  (`legalTargetSetsFrom`) all ask it. A live opponent who controls nothing targetable leaves `Min` at 1, so the spell is
  uncastable, which is what the rule says.

`effects.ForNthOpponent(n, spec)` binds a clause; `perOpponentRequiredClauses(noun, preds…)` is the three-clause statement.
No func-typed field was added to any struct reachable from `Game`: the closure ratchet is untouched.

**Card.** Blatant Thievery, `full`.

**Tests.** `cards/effects/roonb_cards_test.go` and `roonb_cards4_test.go` (four live opponents; a pick from the wrong opponent;
an eliminated opponent; a live opponent with nothing to take; a pick that changed hands in response; a two-player table),
`legal/opponent_slot_test.go` (the enumerator offers one pick per live opponent, none for an eliminated one, and no cast when a
live opponent has nothing; every move is dispatched), `protocol/opponent_slot_view_test.go` (the per-clause `min`).

Back-out: drop `opponentSlotAdmits` from `sourceAdmits` and `TestRoonbBlatantThieveryRequiresAPickPerOpponentFromThatOpponent`
fails; read `c.Min` instead of `clauseMinLocked` in the announce gate and the eliminated-opponent and two-player tests fail;
drop the view stamp or the enumerator's read and the protocol and legal tests fail.

## Amendment (2026-10-20, #4051): an activated ability's target chosen at random

The 2026-10-16 amendment (#1815) let the engine draw a TRIGGER's target and refused the flag on an activated ability,
whose targets the player announces in one message (CR 602.2b, then CR 601.2c). Goblin Test Pilot ("{T}: This creature
deals 2 damage to any target chosen at random") is that case. Its rulings settle every reading: "The target is chosen
at random as part of activating the ability. Players can respond to the ability knowing what the target of the ability
is", "all possible legal targets (including creatures and players) must have an equal chance of being chosen", and the
Pilot is a legal target of its own ability.

**1. The announcement draws it.** `activateCatalogAbilityLocked` slots the payload's refs as before, then
`announceRandomTargetsLocked` (game/random_target_announce.go) fills each random step: a payload ref aimed at one is
refused (`ErrInvalidParam`, no payload may choose a random target), and the step gets one target drawn uniformly from
its legal set by `drawStepTargetLocked`, the trigger walk's draw, on the "pick" stream keyed to the activator and the
source (ADR 0054), so an undo replays it. With nothing to draw the activation is refused (CR 601.2c). Everything after
that (the announce gate, the division, the costs, CR 601.2h) is unchanged. The draw rides the announcement as
`ActivateAbilityParams.randomTargets`, unexported like ADR 0109 §7's random discards and for the same reason: a run
parked on a CR 903.9 question and made again aims at the same target. It is logged as the trigger's is ("<clause>:
<name> (at random)") once the ability is on the stack, so a refused announcement logs nothing.

**2. Nobody picks it.** The clause is a step nobody answers, like an opponent's clause (#2055): `StepsChosenByController`
leaves it out, the enumerator names no target for it and keeps only the prefixes `RandomStepFillableForEffect` can
draw for, and the view marks it `at_random` so the client's walk skips it. An ability whose only clause is random opens
no walk and is activated at once. `abilityClauseView` (a lone ability clause) now carries `chosen_by_opponent` as
well, which Nova Pentacle's view was missing.

**3. Scope.** `effects.Register` allows the flag on an activated ability's lone clause, the one printed shape. A second
clause is refused because the payload names it before the draw, so a relation or Distinct between them could not be
judged; a mode's clause and a spell's stay refused.

**Owner decision not needed.** The rulings fix when the draw is made, over what, and that it is public; each choice
above is the rules-faithful one.

**Cards:** Goblin Test Pilot, `full`.

Back-out: let the payload name the random clause and `TestGoblinTestPilotRefusesANamedTarget` fails; draw from a private
source and `TestGoblinTestPilotReplaysTheDrawAfterUndo` fails; leave the Pilot or players out of the legal set and
`TestGoblinTestPilotDrawsAnyTargetAtRandom` fails; enumerate the clause and `TestGoblinTestPilotEnumeratesNoTarget`
fails.

## Amendment (2026-10-23, #1997): a trigger's target chosen by the triggering player

"Target nonartifact creature that player controls of their choice" (The Abyss, Magus of the Abyss) is chosen by the player the trigger's event names, the player whose upkeep it is, even when that is not the ability's controller. `TargetSpec.ChosenByTriggerActor` (`effects.OfThatPlayersChoice`) is the clause. The #2055 amendment's opponent's choice asks the controller to name an opponent first. This clause asks nothing first, because the event names the chooser. So nothing is recorded on the item: the CR 608.2b re-check reads the same event. `ControlledByChooser` reads "that player controls" against the event's player. Legality is still judged for the controller. Triggers only. [ADR 0109](0109-rule-gates-land-types-mana-and-cost-components.md)'s amendment of 2026-10-23, second part, has the detail.

Back-out: skip the event-player gate in `chooserAdmits` and `TestTheAbyssTargetIsChosenByTheUpkeepPlayer` and `TestMagusOfTheAbyssOnItsControllersUpkeepMustChooseItself` fail; send the prompt to the controller and `TestTheAbyssTargetIsChosenByTheUpkeepPlayer` fails.

## Amendment (2026-10-30, #4616): a ceiling on the total of the chosen targets

The 2026-09-24 amendment (#1559) left one family out of scope: "Rules that are not pairwise-distinct — 'total mana value
4 or less' over a set of targets". The Super Hero Civil War (Marvel Super Heroes, #4455) prints one, chapter I: "Gain
control of up to two target creatures with total mana value 6 or less for as long as this Saga remains on the
battlefield". A key cannot say it: whether a 3 fits depends on what the other pick adds up to, not on a value two picks
share. Every claim below was checked on `origin/develop` at `2edf3b5dd`, and every rule against the pinned Comprehensive
Rules (`MagicCompRules 20260925.txt`): CR 601.2c (targets are announced), CR 603.3d (a trigger's are announced the same
way), CR 608.2b (each is checked again on resolution), CR 115.7d and 115.7e (a retarget), CR 202.3a and 202.3e (mana value),
CR 708.2a (a face-down permanent).

### 1. `TargetSpec.Total`, the third set rule

`game.TargetTotal{Label, Stat, Max}`, attached with `.WithTotal(…)` and built with `effects.TotalManaValueAtMost(n)`:
the picks' total of one statistic may not exceed `Max`. `Stat` is a closed enum (`TargetTotalManaValue`), not a func,
for TargetSameness's reason: the clause is reachable from a stack item and a paused pick_target frame, and a func there
is a new closure route (testdata/closure_fields.txt is untouched). The value is `Game.ManaValueForEffect`, the read
`ManaValueLE` uses: {X} is zero off the stack, an object with no mana cost is zero, and a card whose cost the engine
cannot read has no value and fits no total, as it meets no "mana value N or less". One read differs: a face-down
permanent is zero (CR 708.2a gives it no mana cost), where `ManaValueForEffect` would read the cost of the card it
hides, and the view would then tell every seat what that card is. The same four readers answer it:

| Reader | What it does with the total |
|---|---|
| announce gate (`validateAnnouncedTargetsWithLocked`) | sums the step's picks; refuses `illegal target: those targets must have total mana value 6 or less`, and refuses a pick with no value |
| CR 608.2b re-check (`TargetStillLegalForEffect` → `totalRuleBrokenLocked`) | sums the surviving picks (§2) |
| enumerator (`internal/legal`, `setRuleKeys.total`) | never extends a set past the ceiling; a lone candidate takes the set walk, so "choose nothing" is still offered when none fits |
| client picker (`breaksSetRule`) | greys a candidate whose value would pass the ceiling, from `total: {label, max, values}`; the dock shows "total 4 of 6" |

Fillability (CR 603.3d and the cast offer) counts the smallest values first (`totalCountLocked`): no set of k picks has a
smaller total than the k smallest. The one-slot retarget offer narrows to the cards that fit beside the staying slots,
and the gate judges the final list (CR 115.7e), so an unchanged slot is counted but not refused on its own (CR 115.7d).

### 2. Resolution: the survivors are summed, and over the ceiling NONE is legal

CR 608.2b re-checks every target against the targeting requirements, and the total is one of them. What happens when a
target's mana value has changed so the total now exceeds 6 is settled by the printed rulings on this family. Nethroi,
Apex of Death (2020-04-17) and Reunion of the House (2025-04-04), whose totals can grow while the targets sit in a
graveyard: "the total power of the creature cards may become greater than 10. If this happens, the entire selection of
targets is illegal. In this case, Reunion of the House won't resolve and none of its effects will happen." So:

- a pick that became illegal on its own (it left, stopped being a creature, gained protection) is dropped first and adds
  nothing, as §3 of the 2026-09-24 amendment drops it. A creature that leaves in response costs exactly that creature,
  and removing a pick can only lower the total;
- the picks still individually legal are summed, and when the sum exceeds the ceiling EVERY one of them is illegal. With
  one clause that is every target, so the chapter does not resolve (CR 608.2b), which is the rulings' "nothing is
  returned". None is preferred: the same reading §3 gives a collision under a key;
- a survivor whose value can no longer be read fits no total and is illegal itself; the others are summed without it.

This is a ruling Wizards has published, so no owner decision was needed.

### 3. Scope

`effects.Register` (`checkTotalRule`) refuses a total with no statistic or Label, a negative ceiling, a clause that can
pick a player, a total beside a key rule or `EachOf`, a total with an X (count or bound), on an opponent's or the
triggering player's choice, and on a sacrifice clause. No printed card needs any of those yet, and the fillability count
and the enumerator answer one set rule per clause. "Total power N or less" (Nethroi, Reunion of the House) and "total
mana value X or less" (Orcus, Rampaging Yao Guai) are a new `TargetTotalStat` and an X binding, each its own change.

**Card.** The Super Hero Civil War, `full`.

**Tests.** `game/target_total_test.go` (the announce gate, an unreadable cost, the resolution re-check over survivors,
fillability, the retarget offer and gate, the values), `cards/effects/the_super_hero_civil_war_test.go` (every chapter
through the real catalog, the total growing before resolution, a target leaving), `cards/effects/target_set_test.go`
(the boot checks), `legal/target_total_test.go` (every offered set within the total and dispatched, a lone candidate over
it, the view's `total`), `aiseat/super_hero_civil_war_bot_test.go` (the heuristic answers the prompt within the total and
takes what it names) and `client/src/lib/targetSet.test.ts` (the picker and the dock).

Back-out: skip the total in the announce gate and `TestTotalRefusesASetOverTheTotalAtAnnounce` fails; drop
`totalRuleBrokenLocked` from the re-check and `TestTotalIsRejudgedOverTheSurvivorsAtResolution` and
`TestTheSuperHeroCivilWarChapterITakesNothingWhenTheTotalGrows` fail; drop `setRuleKeys.total` and
`TestSuperHeroCivilWarEnumeratesOnlySetsWithinTheTotal` fails; let a lone candidate take the single-slot answers and
`TestSuperHeroCivilWarLoneCandidateOverTheTotalStillHasAnAnswer` fails.
