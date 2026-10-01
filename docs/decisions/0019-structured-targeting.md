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
