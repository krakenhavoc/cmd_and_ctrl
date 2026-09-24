# ADR 0038 — Protection-style keywords at the targeting choke point

**Status:** Accepted · 2026-09-11 · Sprint S23 · Relates to [#332](https://github.com/krakenhavoc/cmd_and_ctrl/issues/332), [#271](https://github.com/krakenhavoc/cmd_and_ctrl/pull/271), [ADR 0036](0036-attachments.md), [ADR 0037](0037-unimplemented-card-signal.md)

## Context

`server/internal/game/targets.go` is the single choke point for both
halves of targeting legality: the announce gate (CR 601.2c) and the
resolution re-check (CR 608.2b) call the *same* function,
`targetLegalLocked`. Before this change that function read **no
keywords at all**. The complete set of non-test `HasKeyword` call
sites was summoning sickness, block legality, menace count, flash
timing, defender, vigilance, first / double strike, deathtouch,
trample and lifelink — every one of them in a combat or cast-timing
path, none in a targeting path. The predicate library in
`cards/effects/targets.go` had no keyword predicate either, so a
catalog card could not opt in even if it wanted to.

So hexproof, shroud, ward and protection were inert. Not a missing
card — a missing rule. It was already costing three things:

1. **[#332](https://github.com/krakenhavoc/cmd_and_ctrl/issues/332)
   (Lotus Field)** lists hexproof as one of its two blockers.
2. **The Wandering Rescuer ([#271](https://github.com/krakenhavoc/cmd_and_ctrl/pull/271))
   shipped an inert grant.** Its Layer 6 static appended the string
   `"hexproof"` to every other tapped creature you control and
   nothing read it. The card file said so, in a simplification note.
3. **[ADR 0036](0036-attachments.md) flagged it forward.** Lightning
   Greaves and Swiftfoot Boots would ship *strictly stronger than
   printed*, because their entire purpose is a grant nothing
   enforces — inverting the convention that every simplification in
   this repo errs weaker.

## Decisions

### 1. The gate is a rule in the engine, not a predicate on the card

Hexproof and shroud are enforced by `game.CanBeTargetedBy`, called
from `targets.go` on every card candidate. A catalog card does not
opt in and cannot opt out. The alternative — a `NotHexproof()`
predicate every `Targets:` clause remembers to compose — is a rule
that is correct only as often as 250-odd card files remember it, and
the failure mode is silent.

Because announce and re-check are the same function, teaching that
one function the rule buys both: a creature that gains hexproof in
response to a removal spell already on the stack makes that spell
fizzle, with nothing else in the engine touched.

### 2. Read through `Effective()`, never the printed field

`CanBeTargetedBy` calls `game.HasKeyword`, which on the battlefield
reads `Card.Effective().Abilities`. That is the whole reason The
Wandering Rescuer works: its hexproof is a Layer 6 grant, not
printed data, and a gate that read `Card.Keywords` directly would
have left the card exactly as inert as before.

This surfaced a second bug. `AppliesTo` on that static is gated on
`target.Tapped`, and `EventTapCard` / `EventUntapCard` did **not**
bump the layer engine's invalidation counter. The cached resolution
survived the tap, so the grant appeared only when some unrelated
event happened to invalidate. Tap state is an `AppliesTo` input, so
both events now bump (`layer_listener.go`).

### 3. The gate is battlefield-only

Both keywords are abilities of a permanent. `HasKeyword`'s
off-battlefield fallback reads the card's own printed `Keywords`
slice, so without a zone guard a creature card sitting in a graveyard
that happens to print hexproof would refuse Regrowth. The guard is in
`CanBeTargetedBy`, not in its callers, so there is one place to be
right.

### 4. Paying a cost is not targeting

Convoke's tap list, an additional sacrifice cost's option list and an
activated ability's "sacrifice another creature" clause all reuse
`TargetSpec` as a predicate over permanents — and none of them
targets. `sacrifice.go` already made this distinction in a comment
before the rule existed to distinguish from.

Rather than add an opt-out field that 250 catalog files would have to
set, the two enumerations are now named for what they do:

| | targets | keyword gate |
|---|---|---|
| `LegalTargetsForEffect` / `targetLegalLocked` | yes | applied |
| `SpecCandidatesForEffect` / `specMatchLocked(…, false)` | no | skipped |

Five call sites moved to the cost-shaped pair (`tap_cost.go`,
`activated.go`, `protocol/view.go`, `legal/abilities.go`,
`legal/cast.go`). The default — what a new card gets for free — is
the targeting one, because a forgotten opt-in silently loses the
rule while a forgotten opt-out fails loudly the first time somebody
tries to convoke a Bogle.

### 5. The bot's enumerator gets the same gate, for free

`internal/legal` reaches targeting through `LegalTargetsForEffect`,
so it inherited the filter without an edit — and its cost paths were
moved to `SpecCandidatesForEffect` in the same pass.
[#347](https://github.com/krakenhavoc/cmd_and_ctrl/pull/347) is the
precedent: a gate added to the engine and not to the enumerator
produces a bot that proposes moves the server then rejects.
`legal/protection_keywords_test.go` guards both directions.

### 6. "Hexproof from <quality>" must not import as hexproof

Scryfall's `keywords` array tags 21 cards — Knight of Grace, Garruk's
Harbinger, Sphinx of the Guildpact, six Jaheiras — with **both**
`"Hexproof"` and `"Hexproof from"`, although each prints only
"Hexproof from black" or "Hexproof from monocolored". Importing the
array verbatim would have handed all 21 full untargetability:
stronger than printed, which is the one direction this repo never
errs in.

`deck.printedKeywords` now confirms a keyword with a narrower printed
variant against the oracle text before stamping it, reusing the
`keywordLines` scan the multi-face narrowing already used. The card
then carries neither the broad nor the narrow ability, which errs
weaker. `hexproof` is the only token in that set today; shroud and
the combat keywords have no parameterised form.

### 7. Ward and protection are deliberately **not** here

Both belong to the same family and neither is a missing afternoon of
work — each is blocked on a structural fact.

**Ward is not a targeting restriction.** CR 702.21a makes it a
*triggered* ability: a warded permanent is a perfectly legal target,
and the trigger then counters the spell unless its controller pays.
Putting it in `CanBeTargetedBy` would refuse the target outright
instead of offering the payment — the wrong rule, in the wrong
place. It belongs on the trigger path, next to the existing
`PayUnless` prompt (which does exist and does have the right shape:
`QueuePayUnlessForEffect` already takes a chooser who is not the
controller of the source). What is missing is the trigger condition
("becomes the target of a spell or ability an opponent controls"),
the counter-the-spell consequence, and a home for the cost.

**Protection's test is against the source, not the controller.**
CR 702.16b compares the quality to the *source object* of the spell
or ability. `targetLegalLocked` receives a `caster uuid.UUID` and
never sees the source, so the test cannot be written here at all.
Threading a source through the announce gate, the trigger
dispatcher, the pending-choice resume and the SBA re-check is a real
refactor that lands in `CastSpell`.

Both are also **parameterised** keywords — "ward {2}", "ward—pay 3
life", "protection from red" — and `Characteristic.Abilities` is a
`[]string` of bare tokens with nowhere to put the cost or the
quality. Scryfall says only `"Ward"` and `"Protection"`; the
parameter lives in oracle text.

And protection is DEBT — **D**amage, **E**nchanting/Equipping,
**B**locking, **T**argeting. Shipping only the T while `coverage.go`
stopped flagging the card would make the unimplemented-card signal
([ADR 0037](0037-unimplemented-card-signal.md)) lie: a player would
get no warning before pointing a Lightning Bolt at their own
protection-from-red creature and watching it die. Neither keyword is
in `canonicalKeywords`, so cards printing them still flag as
unimplemented. That is the honest answer until the whole of DEBT
lands.

*Superseded 2026-09-18 (S40, #662):* the protection half of this
decision is now superseded in full by
[ADR 0072 — Protection (CR 702.16)](0072-protection.md), which
shipped it. Read that ADR, not this paragraph, for how protection is
represented, where its four DEBT checks live, and what is out of
scope. Ward is untouched and everything decision 7 says about ward
still stands.

*Amended 2026-09-16 (S30 closeout, #95):* the protection half of this
decision is superseded by the protection ADR that
[#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662) asks
for, and two statements above were wrong as written. First, the
targeting rule is CR 702.16b; 702.16e is damage prevention. Second,
it is not true that no DEBT hook receives the source. `CanBlock`
(`game/keywords.go`) is handed both creatures, and
`attachmentLegalLocked` (`game/attach.go`) is handed the Aura or
Equipment itself. Damage carries only a source ID
(`ReplacementEvent.DamageSource`), and that source may already have
left the battlefield, so it needs a last-known-information lookup.
Targeting is the hook that gets only the controller's ID, and it is
still the real refactor. Ward and the parameterised-keyword argument
are unchanged.

## Consequences

**Behaviour change, measured.** Against the 2026-09-09 Scryfall dump,
**74 Commander-legal cards** now carry hexproof and **35** carry
shroud (109 total, 125 counting non-legal printings). Across the four
tracked decklist docs the intersection is **exactly one card: Lotus
Field**, in `aang-is-so-flashy.md`. Zero shroud, zero protection. No
registered catalog Spec prints either keyword — Baneslayer Angel's
protection is the only near miss and is unaffected — and no e2e
fixture deck contains one. The behavioural blast radius on the live
playtest group is one land that does not work yet for other reasons.

**Lotus Field is still blocked.** Hexproof was one of its two
blockers; the other is "{T}: Add three mana of any one color". Each
pipe slot in `Produced` queues its own independent
`PendingChoiceMana` (`AddManaForEffect`), so three `{W|U|B|R|G}`
slots would let the controller pick three *different* colours —
"three mana of any colors", strictly stronger than printed. That is
a separate gap and the card does not ship until it closes.

**The Wandering Rescuer is now correct as printed** and its
simplification note is gone. Nothing about the card had to change,
which was the point of declaring the grant rather than omitting it.

**ADR 0036's forward flag is cleared.** Lightning Greaves and
Swiftfoot Boots can now ship with a real hexproof grant.

## S30 follow-up (#95)

Decision 7 said ward and protection were each blocked on a
structural fact rather than on time. One of the two turned out to
be true only of the *placement*, and the S30 work is recorded here
so this section stops reading as an open question.

**Ward shipped, and not here.** It is a `game.TriggeredAbility`
built by `effects.Ward(WardMana("{2}"), label)` in
`server/internal/cards/effects/ward.go`, watching
`EventBecomesTarget`, gated on `ev.Actor != source.Controller`, and
resolving into the existing `QueuePayUnlessForEffect` prompt
addressed to the *caster*. Countering on a decline is
`CounterTargetForEffect`. Nothing in `CanBeTargetedBy` changed, and
nothing should: the four observable facts that separate ward from
hexproof — the announce succeeds, the payer is not the ward
permanent's controller, paying resolves the spell, and the trigger
is itself a respondable object on the stack — are exactly the four
a targeting-gate implementation cannot produce.

One engine change was needed and it is worth naming because it is
reusable. `EventBecomesTarget` carried `Source` (the source CARD)
but not the stack item. For a spell those coincide; for an
activated or triggered ability they do not, so an effect that has
to act on "the spell or ability that targeted me" could not reach
it. `Event.StackItemID` now carries the item, threaded through all
five `emitBecameTargetLocked` call sites. Any future "counter it",
"copy it" or "change its target" effect keyed on becoming a target
needs the same field.

**Ward is still not in `canonicalKeywords`,** and the reason is the
second half of decision 7 rather than the first: the cost is a
parameter and `Characteristic.Abilities` is a slice of bare tokens.
A bare `"ward"` in the enforced table would tell the ADR 0037
coverage signal that every ward card works while the engine had no
idea what to charge. The cost lives on the catalog Spec instead, so
an uncatalogued ward card still flags as unimplemented — the honest
answer, and the same one this ADR gave for protection. Only MANA
wards are supported; "ward—pay 3 life" and "ward—sacrifice a
creature" need a pay-unless prompt whose cost is an `AbilityCost`
rather than a string, and `effects.WardCost` is the type that will
carry it.

*Update, 2026-09-16 (#95):* life and sacrifice wards have shipped,
and they did not need an `AbilityCost` pay-unless. `WardCost` gained
`Life` and `Sacrifice`. A life ward is a `PendingChoiceConfirm` with
`LifeCost` declared. A sacrifice ward is the same Confirm chained to
a `PendingChoiceChooseCards` over the payer's own permanents. Both
prompt kinds came from #552. Refraction Elemental, Sedgemoor Witch
and Vein Ripper carry them; see `effects/ward.go`.

**Protection is unchanged and still absent.** Its blocker is the
one decision 7 identified as a real refactor — the quality is
tested against the SOURCE object (CR 702.16b) and neither
`targetLegalLocked` nor the damage, block or attachment paths
receive one — and S30 did not attempt it. Protection-printing cards
continue to flag as unimplemented, which keeps the DEBT problem
(shipping only the T while the signal goes quiet) from arising.

*Update, 2026-09-18 (S40, #662):* protection shipped, in
[ADR 0072](0072-protection.md). The refactor decision 7 predicted is
what it cost: `game.TargetSource` replaced the bare `caster
uuid.UUID` on the targeting half of `targets.go` and every targeting
call site now names its source. The quality lives in a parameterised
token (`protection from red`) with one closed-grammar reader in
`game/protection.go`, which is the half decision 7 called impossible
because `Characteristic.Abilities` is a `[]string` — it turned out a
token can carry its parameter as long as exactly one reader parses
it. The DEBT-signal argument held: `canonicalKeywords` gained
`protection` in the same change that enforced all four checks, and a
quality the grammar cannot parse still mints no token and still
flags the card.

*Update, 2026-09-24 (#1417):* the damage half of protection now
judges a source that left the battlefield *before* its damage event
by the characteristics it had as it last existed there. It no longer
uses its graveyard card. See
[ADR 0072's 2026-09-24 amendment](0072-protection.md) and
[ADR 0056 Decision 12](0056-infect-wither-toxic.md). This settles
the "needs a last-known-information lookup" half of Decision 7's
damage paragraph.

**Indestructible joined the table separately, in S25**
([#380](https://github.com/krakenhavoc/cmd_and_ctrl/pull/380),
`server/internal/game/indestructible.go`). It is in this family by
reputation but not by structure, which is why it landed without
needing anything this ADR argued about: it is unparameterised, it is
read off `Effective().Abilities` like every other keyword, and its
consumers — the destruction path — are already holding the card.
