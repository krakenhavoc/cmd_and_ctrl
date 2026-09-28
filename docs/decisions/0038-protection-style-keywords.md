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

## Amendment (2026-09-27, #1560): waiving hexproof, and ward that doesn't trigger

Nowhere to Run (#1565, the Edea deck) prints two statics this ADR had
no shape for: "Creatures your opponents control can be the targets of
spells and abilities as though they didn't have hexproof" (CR 702.11)
and "Ward abilities of those creatures don't trigger" (CR 702.21).
Kaya, Bane of the Dead prints the first one for players as well.

### A1. A waiver at the choke point, not a layer-6 removal

"As though it didn't have hexproof" is not "loses hexproof". Glaring
Spotlight's ruling says so directly: the creature keeps the ability,
and only the targeting is affected. Removing `"hexproof"` from
`Characteristic.Abilities` would work today, because targeting is
hexproof's only reader. It would still be wrong in three ways. The
badge would disappear from the creature's own controller's view. A
second waiver would be timestamp-ordered against every other layer-6
grant. And a Shadowspear that removed hexproof would become the same
object as a Nowhere to Run that waived it. Shadowspear's "lose
hexproof" stays a layer-6 `RemoveKeywordsMod`.

So the waiver is a new catalog slot, `Spec.HexproofBypasses
[]game.HexproofBypass`. `game/hexproof_bypass.go` reads it live off
the battlefield, keyed by `CatalogAbilityKey`, the way `BlockRules`
and `PlayerKeywords` are read. Nothing is stored, so undo, the
snapshot and restore points have nothing new to carry. A source that
has lost all its abilities waives nothing (CR 613.1f).

### A2. Only hexproof, and only when hexproof is what refuses

`canBeTargetedBy` (keywords.go) checks shroud first, then hexproof,
then protection, as before. It asks for a waiver only when hexproof
alone would refuse the target, so a board with no hexproof on it
never walks the battlefield looking for one. Shroud (CR 702.18) and
protection (CR 702.16b) are separate refusals, and a hexproof waiver
does not affect them. The exported `CanBeTargetedBy` has no `*Game`
and so honours no waiver. Both targeting call sites in `targets.go`
now use `g.canBeTargetedByLocked`.

The announce gate (CR 601.2c) and the resolution re-check
(CR 608.2b) go through the same function, so decision 1 covers them
both again. If Nowhere to Run leaves while a spell aimed at a
hexproof creature is on the stack, the target becomes illegal. That
is the card's 2024-09-20 ruling, and the resolution path needed no
change for it. The bot's enumerator reaches targeting through
`LegalTargetsForEffect` and gets the waiver for free, as decision 5
predicted.

### A3. Whose spells: `YoursOnly`

The printed cards differ on one word. Glaring Spotlight, Kaya and
Detection Tower say "spells and abilities **you control**". Nowhere to
Run says only "spells and abilities", so any player's spell may
target the creature, including one cast by the creature's other
opponents. `HexproofBypass.YoursOnly` carries that difference as data.
The effects constructors name it (`BySpellsAndAbilities` /
`BySpellsAndAbilitiesYouControl`), so a card file cannot leave it out
without choosing.

The affected set is read live on every check. A static ability has
no locked set (CR 611.3a). A creature that enters or changes control
afterwards is covered for exactly as long as it matches.

### A4. The player half

`HexproofBypass.Player` waives a player's hexproof (CR 702.11d). It
is read by `canPlayerBeTargetedByLocked`, the player half of the same
choke point, in the same place: only once player hexproof would
refuse. Kaya, Bane of the Dead is the card that needs it.

### A5. Ward suppression lives on the trigger

`Spec.WardSuppressions []game.WardSuppression` is read by
`Game.WardSuppressedForEffect`. The one caller is the trigger
condition in `effects.WardGranted`. Every ward in the catalog uses
that condition: printed `Ward`, a granted ward (Lavaspur Boots), an
emblem's (Teferi Akosa), and a face-down permanent's. Three details
matter:

- It is asked about the **warded permanent**, never the trigger's
  source. A granted ward's trigger belongs to the Equipment or
  emblem, but "ward abilities of those creatures" is about the
  creature that has the ward.
- It is asked **when the ability would trigger**, and at no other
  time. Removing Nowhere to Run later does not make ward trigger
  retroactively (the card's second ruling). A ward trigger that is
  already on the stack still resolves.
- It suppresses **ward** only. Diffusion Sliver's "counter it unless
  its controller pays {2}" works like ward but is not a ward ability,
  and it is not built on `WardGranted`. Leaving it alone is what the
  printed text says.

### A6. Not built here

- **A waiver with a duration.** Detection Tower's "{1}, {T}: Until
  end of turn, your opponents and creatures your opponents control
  with hexproof can be the targets…" would be a `ScopedEffect` mod
  kind that the same two readers walk. No card in a tracked deck needs
  it yet, and it is the only missing piece for Detection Tower.
- **"Lose hexproof and can't have hexproof"** (Arcane Lighthouse,
  Archetype of Endurance). This is a layer-6 removal plus a rule that
  stops the ability being gained again. It is a different seam.
- **Glaring Spotlight.** Its static is this seam. Its second ability
  says "creatures you control … can't be blocked this turn", and the
  card's ruling covers creatures that arrive after the ability
  resolves. `RestrictUntilEOT` locks its set at resolution, so the
  card would ship weaker than printed. It is left out until
  "can't be blocked" can be read over a live set.

## Amendment (2026-09-28, #1651): a waiver with a duration, and "can't have hexproof"

A6 left two shapes unbuilt. This amendment builds both.

### B1. A turn-scoped waiver is a `ScopedEffect` mod, read at the same choke point

Detection Tower: "{1}, {T}: Until end of turn, your opponents and
creatures your opponents control with hexproof can be the targets of
spells and abilities you control as though they didn't have hexproof."

This is A1's waiver, created by a resolving ability (CR 611.2) instead
of printed on a permanent. It is an ADR 0041 data record with a new mod
kind, `waiveHexproof` (`game.WaiveHexproofMod`), under a new affected
scope, `opponentsAndTheirCreatures`. The kind has its own reader,
`readerTargeting`. The layer adapter skips it, because it is not a
characteristic. `hexproofBypassedLocked` and
`playerHexproofBypassedLocked` walk `Game.ScopedEffects` after the
battlefield statics. So the announce gate, the CR 608.2b re-check and
the bot's enumerator all get it from the same two functions, as A2 said
they would.

Three details are fixed by the printed text:

- **The beneficiary is the record's `Controller`.** That is the
  ability's controller as it resolved (CR 611.2c reads "you" once).
  Every card that prints a timed waiver says "spells and abilities you
  control", so the kind always behaves like A3's `YoursOnly`. A
  third player gets nothing, and the kind has no flag to say
  otherwise. If a card ever prints an unrestricted timed waiver, it
  gets a new kind.
- **The set is live.** A waiver changes no characteristic and no
  control, so CR 611.2c does not lock it (the same reading #1571 and
  #1650 used for attack requirements and restrictions). A creature
  that an opponent casts after the Tower resolves is covered, and so
  is one that changes control to an opponent. The scope matches
  opponents' **creatures** only. A hexproof noncreature permanent
  (for example, an animated manland that stops being a creature) is
  not covered, because the Tower does not name it.
- **The player half.** The new scope is the first one that names
  players. `scopeCoversPlayer` answers for it. Every other scope
  covers no player, so no existing record changes meaning.

The record is data. Undo copies it with the registry, and the snapshot
writes it verbatim. The cleanup sweep ends it like every other
until-end-of-turn record. No new state was needed. The restore-time
key check refuses an unknown scope or kind, as ADR 0041 P4 requires.

### B2. "Can't have" is a post-layer-6 strip, recorded on the characteristic

Arcane Lighthouse: "Until end of turn, creatures your opponents control
lose hexproof and shroud and can't have hexproof or shroud." The
Archetype cycle prints the static form: "Creatures your opponents
control lose <keyword> and can't have or gain <keyword>."

"Can't have" is a CR 101.2 "can't". It beats every "can", whatever the
timestamps. A layer-6 removal alone is not enough, because layer 6
runs in timestamp order (CR 613.7). A grant sorted after the removal,
such as a Heroic Intervention cast afterwards or a catalog creature's
own `PrintedKeywords` static, appends the keyword again and keeps it.
That was the declared gap on Archetype of Aggression and Archetype of
Courage.

There were two ways to model it:

1. **A flag that the keyword reader checks.** `HasKeyword` would
   answer false for a keyword the object can't have, and the
   keyword would stay in `Characteristic.Abilities`.
2. **A strip after layer 6.** Effects record what the object can't
   have, and when the layer-6 bucket finishes, the engine removes
   those keywords from `Abilities`.

This amendment uses (2). Under (1), the keyword would still be in
`Abilities`, so the wire would ship a hexproof badge on a creature
that has no hexproof. Every reader that walks the list instead of
calling `HasKeyword`, such as the protection-quality walk and the
view, would need to learn the rule, and a missed reader would be a
silent leak. Under (2), every downstream reader sees the right list
without being changed. That is the argument `materialiseControlLocked`
made for layer 2.

How it works:

- **`Characteristic.CantHave []string`** holds the keyword tokens the
  object can't have. It works like `Restrictions`. A layer-6 effect
  appends to it. Nothing clears it, including a CR 613.1f "loses all
  abilities". The can't-have belongs to the effect's source, not to
  the object.
- **`enforceCantHaveLocked`** runs once, right after the layer-6
  bucket (`layerPassLocked`). It removes every listed token from
  `Abilities`, compared case-insensitively. Layer 7 and everything
  after the pass see the stripped list. The timestamp of the
  effect that recorded the can't-have does not matter, because the
  strip runs after the whole bucket. That is the point of the rule.
- **CR 613.6 still applies to the source.** The recording happens
  inside the ordinary layer-6 `Apply`. So an Archetype that has lost
  its abilities records nothing, and its opponents' creatures can have
  the keyword again. This needed no extra code.
- **The static form** is `effects.LoseAndCantHave(applies, keywords…)`,
  which is one layer-6 `StaticAbility`. **The scoped form** is the
  mod kind `cantHaveKeywords` (`game.CantHaveKeywordsMod`), which is a
  layer-6 mod whose `Apply` records the tokens. Arcane Lighthouse
  changes characteristics, so CR 611.2c locks its set. It is a
  pinned record over the opponents' creatures as the ability
  resolves. A creature that comes under an opponent's control later
  keeps its hexproof.
- **"Lose" needs no separate removal.** The strip removes the keyword
  whatever added it, including the printed baseline. "Loses X and
  can't have X" is therefore one record or one static, not two.

`CantHave` is part of the characteristic's snapshot shape. It appears
on last-known information like every other field there. It is
`omitempty`, so every existing fixture renders byte-for-byte as before.
`clone` copies it, and `sameCharacteristic` compares it.

### B3. Cards

- **Detection Tower** (B1) and **Arcane Lighthouse** (B2, scoped) ship
  `full`.
- **Archetype of Endurance**, **Archetype of Imagination** and
  **Archetype of Finality** (B2, static) ship `full`.
- **Archetype of Aggression** and **Archetype of Courage** are moved
  onto `LoseAndCantHave`. Their caveats described exactly this gap,
  so they now ship `full`.
