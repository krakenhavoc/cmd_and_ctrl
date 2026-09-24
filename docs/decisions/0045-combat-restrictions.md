# ADR 0045 — A restriction vocabulary for declarations and activations

**Status:** Accepted · 2026-09-14 · Sprint S24 · Relates to [#76](https://github.com/krakenhavoc/cmd_and_ctrl/issues/76), [#544](https://github.com/krakenhavoc/cmd_and_ctrl/issues/544), [#546](https://github.com/krakenhavoc/cmd_and_ctrl/pull/546), [ADR 0036](0036-attachments.md), [ADR 0038](0038-protection-style-keywords.md)
**Addendum (Accepted, 2026-09-17):** [game-aware block legality](#addendum-2026-09-17-game-aware-block-legality--landwalk-conditional-restrictions-and-block-counts-705-750) for [#705](https://github.com/krakenhavoc/cmd_and_ctrl/issues/705) and [#750](https://github.com/krakenhavoc/cmd_and_ctrl/issues/750). §1–§6 below stand, except where the addendum's header says it supersedes them.

## Context

Every continuous effect the catalog could express before this change
either changed a characteristic (P/T, types, colours, name) or granted
a keyword. There was no way to say **"can't"**.

That is one sentence, and it was the entire reason five catalogued and
uncatalogued cards were wrong:

- **Whispersilk Cloak** shipped granting shroud and not "can't be
  blocked", with the caveat naming the cause: *"Evasion that is not a
  keyword ability lives in the declare-blockers legality check, and the
  engine has no restriction vocabulary there."*
- **Carrion Feeder** shipped with its printed drawback unenforced —
  *"'This creature can't block' isn't enforced — the Feeder can
  block."* A card strictly **stronger** than printed, which inverts
  this repo's standing convention.
- **Pacifism, Arrest and Faith's Fetters** were not in the catalog at
  all. The S24 status comment on #76 listed the gap as *"the cleanest
  next sub-PR in this seam"*.

`CanBlock` read the CR 702 evasion keywords and nothing else;
`DeclareAttacker` checked tapped, summoning sickness and defender and
nothing else; `ActivateCatalogAbility` had no concept of a permanent
whose abilities are switched off.

## Decisions

### 1. A restriction is a field, not a keyword string

`Characteristic.Restrictions` is a `game.Restriction` bitmask, written
by ordinary `StaticAbility.Apply` funcs and OR-accumulated.

The alternative — appending `"can't attack"` to
`Characteristic.Abilities`, which would have needed no new field and no
new wire key — is wrong three times over:

1. **It is not an ability the creature has.** Pacifism's restriction
   belongs to the *Aura*. CR 613 gives restrictions no layer at all for
   exactly this reason: they are not characteristics.
2. **It would render as a keyword badge**, which is the one failure
   mode [ADR 0038](0038-protection-style-keywords.md) exists to
   prevent, in reverse — a badge for something that is not a keyword.
3. **"Enchanted creature loses all abilities" would switch Pacifism
   off.** Darksteel Mutation and Song of the Dryads are the next cards
   in this seam (#76's open list) and they clear
   `Characteristic.Abilities` in layer 6. A creature that is both
   Mutated and Pacified can still not attack, and a string in the
   abilities list would have silently got that backwards.
   `TestLosingAllAbilitiesDoesNotClearRestrictions` pins it ahead of
   the card landing.

Because the field is only ever OR'd into and nothing clears it, the
layer a restriction is written in is unobservable and so is its
timestamp. The primitives use layer 6 because that is where the text
sits on the card; nothing depends on that choice.

### 2. Five bits, and the taxonomy is the point

| Text | Bit | Restricts |
|---|---|---|
| "~ can't attack" | `CantAttack` | the creature |
| "~ can't block" | `CantBlock` | the creature |
| "~ can't be blocked" | `CantBeBlocked` | the **defender's** options |
| "activated abilities can't be activated" | `CantActivate` | non-mana activations |
| "…unless they're mana abilities" | `CantActivateMana` | mana activations |

The shapes are not interchangeable, and the difference is *who* is
restricted rather than what the text says. `CantBeBlocked` is carried
on the attacker because that is the permanent the effect is attached
to, and read inside `CanBlock` because that is the one predicate that
sees both cards.

The activation pair is two bits rather than one because **Faith's
Fetters prints the difference** — Arrest stops mana abilities, Fetters
explicitly does not. Two bits also map one-to-one onto the engine's two
activation entry points, so neither gate needs to know the other's card
text.

### 3. One predicate per declaration, shared with the enumerator

This is the decision that matters most, and it is a direct consequence
of [#544](https://github.com/krakenhavoc/cmd_and_ctrl/issues/544).

`internal/legal` must never enumerate a move the engine refuses. When
it does, and the seat is a bot, the result is not a bad move: a seat
that owes a decision is enumerated that decision's answers and nothing
else, so a deterministic policy re-picks the same rejected move
forever and the table stops. That happened on a live game against a
human on 2026-09-14.

So the restriction is not checked in two places that agree by
inspection. There is **one function per declaration**, and both sides
call it:

- `game.AttackerEligible(card, seat)` — the whole attacker-side list
  (creature, controlled by seat, untapped, not already declared, no
  defender, not summoning sick, no "can't attack"). The engine's bulk
  `DeclareAttackers` runs it; `legal.combatMoves` runs it. It is
  modelled on `BlockerEligible`, which #328 introduced for the same
  reason.
- `game.CanBlock(attacker, blocker)` — evasion keywords plus both
  block-side restriction bits. `DeclareBlocker` runs it,
  `legal.combatMoves` runs it, and `seatOwesBlockDecisionLocked` (the
  #328 auto-pass guard) runs it.
- `game.CanActivateAbilities` / `game.CanActivateManaAbilities` — the
  activation pair, run by the three engine activation paths and by
  both enumerator sites.

`internal/legal/restrictions_test.go` asserts the invariant with
`dispatchAll` — every enumerated move is accepted by the dispatcher —
*and* asserts that the engine really does refuse the withheld move, so
a future enumerator cannot satisfy the test by going silent on both
sides.

### 4. The auto-tapper is a planner and has to agree too

`gatherTapSources` skips a permanent whose mana abilities can't be
activated. It plans a payment that `ActivateManaAbility` then executes,
so a source the executor will refuse must never be planned — the
failure mode is a half-paid cost with earlier permanents already
tapped. Same argument #352 made for condition-gated abilities.

### 5. Restrictions are checked at declaration, and only there

CR 508.1c / 509.1b/c: a declaration that violates a restriction is
illegal when it is made. Nothing is undone afterwards. CR 506.4 lists
what removes a permanent from combat — leaving the battlefield,
changing control, ceasing to be a creature — and "acquired a
restriction" is not on it, so a creature pacified mid-combat keeps
attacking and deals its damage. Pinned by
`TestRestrictionIsCheckedOnlyAtDeclaration`.

### 6. The client reads the restriction set; it does not derive it

`CardView.restrictions` is a `[]string` of stable snake_case tokens,
separate from `abilities` because the two render differently: a keyword
gets a badge, a restriction disables a control and supplies the reason.
`attackAll.ts` uses it so the "attack with all N" count is honest, and
the ability menu uses it so an Arrested permanent's rows grey with a
reason instead of opening a rejection toast.

Nothing in the client re-derives who can attack. #429 spent a PR
deleting client-side rules duplication and this does not start a new
one.

## What this deliberately does not express

Stated here because the shapes are named on the roadmap and the next
person needs to know they were considered, not overlooked.

- **"Can't attack unless its controller pays {2}"** — Propaganda,
  Ghostly Prison, Norn's Annex. This is a *cost* to attack, not a
  prohibition, and a bit cannot carry a cost. It wants an attack-cost
  pipeline shaped like the cast-cost one and keyed on the **defending
  player** rather than on the attacker. Nothing here blocks it; it is
  simply a different mechanism.
- **Count restrictions** — **Silent Arbiter** ("no more than one
  creature can attack each combat and no more than one creature can
  block each combat") and **Crawlspace** ("no more than two creatures
  can attack you each combat"), both named on the roadmap as blocked
  on this seam. They are **reachable now and this design does not
  preclude them**, but they are not bits: a count is a property of the
  whole declaration, not of one permanent. The seam is already there —
  `DeclareBlockers` receives an entire block declaration in one call
  and refuses an illegal COUNT before storing any of it
  (`blockerBoundsLocked`, the addendum's Decisions 12-13), and
  `DeclareAttackers` receives the entire attacking set in one call. A
  count limit belongs beside `blockerBoundsLocked` as a set-shaped
  predicate — one more entry in `checkBlockDeclarationLocked` — with
  the per-permanent bits in this ADR left alone. The enumerator
  agreement is harder for those than for these, because a per-move
  enumerator has to reason about a set — which is the open question to answer before writing
  them, not after.
- **Goad** ("can't attack you or a planeswalker you control") and
  **landwalk** ("can't be blocked as long as defending player controls
  an Island") are *conditional on the other side of the pairing*.
  Neither is a property of one permanent. Goad's home is
  `canAttackTargetLocked`, which already takes (attacker's controller,
  target) and is already shared with `AttackTargetsForEffect`;
  landwalk's is `CanBlock`, once `CanBlock` can see the game, which
  today it cannot. Lord of Atlantis's islandwalk caveat stays declared
  for that reason.

## Consequences

- Five cards ship or are completed: Pacifism, Arrest, Faith's Fetters,
  Rogue's Passage, plus Whispersilk Cloak and Carrion Feeder promoted
  to `CompletenessFull` with their caveats retired.
- `Characteristic` grows one field, which `snapshot.go` carries by
  value and `TestEmbeddedDomainTypesStayPureData` already guards.
- The wire grows one additive, omitempty `CardView` field.
- A restriction is now the cheapest possible thing to add to a card:
  `RestrictAttached(game.CantAttackOrBlock)` on an Aura,
  `RestrictSelf(...)` on a creature, `RestrictUntilEOT{...}` from a
  spell or ability.

## Addendum (2026-09-17): game-aware block legality — landwalk, conditional restrictions and block counts (#705, #750)

**Status:** Accepted · 2026-09-17 · Wave 2 of the card-coverage audit ·
tracked on [#705](https://github.com/krakenhavoc/cmd_and_ctrl/issues/705)
(landwalk) and [#750](https://github.com/krakenhavoc/cmd_and_ctrl/issues/750)
(conditional restrictions and block counts). This status covers this
section only.
**Numbering:** an addendum, not a new ADR, so no number is taken. On
2026-09-17 all 208 remote branches were checked (after
`git fetch origin '+refs/heads/*:refs/remotes/origin/*' --prune`), and
none changes this file.
**Supersedes:** [ADR 0014](0014-combat-keywords.md) §9
("Menace enforced at `DeclareBlocker` close-out, not per-decl"), and the
"Count restrictions" and "landwalk" bullets of *What this deliberately
does not express* above.
**Leaves a slot for:** protection's blocking check, CR 702.16f
([#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662)),
in the same function ([Decision 9](#9-one-order-of-checks-with-a-reserved-slot-for-protection)).
**Overlaps:** [#715](https://github.com/krakenhavoc/cmd_and_ctrl/issues/715)'s
menace item ([Decision 13](#13-the-stored-declaration-is-always-legal-and-the-menace-close-out-is-deleted)).
**Coordinates with:** [ADR 0059](https://github.com/krakenhavoc/cmd_and_ctrl/pull/824)
(turn machinery): `TurnScopedBlockRules` is swept inside its
`sweepTurnEndLocked` ([Decision 11](#11-rules-with-a-parameter-are-block-rules-read-from-their-sources-at-check-time)).
**Owner decisions:** answered 2026-09-17 and recorded under
[Decided (2026-09-17)](#decided-2026-09-17). Multi-blocker blocks are
staged on the board, attackers the selected blocker can't block are
dimmed, and the first wave is option (b), with fear, intimidate,
shadow, horsemanship and skulk moved to
[#825](https://github.com/krakenhavoc/cmd_and_ctrl/issues/825).

Line references are to `origin/develop` at `684f2786`. CR citations are
to the Aug 7, 2026 edition. Oracle text was read from the local Scryfall
dump.

### Context

#### Block legality can't see the game

`CanBlock(attacker, blocker *Card)` (`server/internal/game/keywords.go:310-330`)
reads the two restriction bits on the two cards and flying against
flying or reach. Nothing else. It has no `*Game`, so anything that
depends on the rest of the table has nowhere to go:

- **Landwalk** (CR 702.14c): "can't be blocked as long as the defending
  player controls at least one land with the specified land type". The
  defending player's lands aren't one of the two cards.
  `restrictions.go:76-83` and §"What this deliberately does not express"
  above both put landwalk's home in `CanBlock` "once `CanBlock` can see
  the game".
- **Rules with a parameter**: "can't be blocked except by Walls"
  (Prowler's Helm), "can't be blocked by creatures with power 2 or less"
  (Legolas Greenleaf), "creatures with power less than this creature's
  power can't block creatures you control" (Champion of Lambholt, which
  reads a third card live). `Restriction` is a `uint8` of five bits
  with no parameters (`restrictions.go:99-129`).
- **Rules that read the attacker's own layer-7 values**: Thieves'
  Tools' "can't be blocked as long as its power is 3 or less". A
  restriction is written in layer 6 (`effects/restrictions.go:47-67`),
  before power exists, so it can't be a conditional bit either.

Three callers must always agree (§3): `DeclareBlocker`
(`mutations.go:4362`, check at `:4398`), the move enumerator
(`legal/combat.go:94-116`, check at `:104`) and the #328 auto-pass
signal `seatOwesBlockDecisionLocked` (`blockers.go:99-141`, check at
`:135`). All three already run with the game lock held and layers
fresh: `DeclareBlocker` under the write lock after
`RecomputeLayersIfStaleLocked` (`mutations.go:4374`), the other two
inside `ReadSnapshot`.

#### Menace is enforced by silently undoing the block at damage

`BlockerCountValid` (`keywords.go:332-356`) knows one count rule, menace's
minimum of two, and **only `keywords_test.go` calls it**. Menace is
really enforced at the top of `assignAndDealCombatDamageLocked`
(`mutations.go:4570-4605`): any menace attacker with exactly one live
blocker has that blocker's `BlockingTarget` cleared, with no log line
and no warning. ADR 0014 §9 chose this on purpose, so that a player
halfway through a two-creature block doesn't see an error.

What that costs, checked on `684f2786`:

1. **The illegal block has effects before it's undone.** `DeclareBlocker`
   accepts the lone block, emits `EventBlock` (`mutations.go:4406-4418`),
   logs it and draws the arrow. It holds through the whole declare
   blockers priority window, and "whenever this blocks" triggers fire on
   it (Savvy Hunter, `effects/savvy_hunter.go:31-34`). CR 509.1 and
   733.1 say an illegal declaration is reversed and "no abilities
   trigger … as a result of an undone action".
2. **A legal block is undone too.** The close-out reads the live
   battlefield, so a menace attacker blocked by two creatures, one of
   which is destroyed before damage, loses its remaining block
   (#715's third probe). CR 509.1b: a change after a legal block
   "doesn't affect that block".
3. **The enumerator and the bots offer the lone block**
   (`legal/combat.go:103-115` has no count check), and the #328 signal
   says so openly (`blockers.go:74-80`).
4. **A maximum has no home.** Hungering Hydra, Vorrac Battlehorns and
   Alpha Authority print "can't be blocked by more than one creature".
5. **ADR 0045's own text is wrong about it.** Lines 170-173 above,
   `restrictions.go:69-71`, `blockers.go:76`, `keywords.go:334` (which
   also cites CR 702.110; menace is 702.111), `indestructible.go:13`,
   `effects/boggart_brute.go:5-8` and `AGENTS.md:1003-1007` all say
   `BlockerCountValid` runs at the declare-blockers close-out.

§"What this deliberately does not express" left one question open:
"the enumerator agreement for a set … is the open question to answer
before writing them". [Decision 14](#14-one-option-generator-shared-by-the-enumerator-and-the-328-signal)
answers it.

#### Blocking is incremental in this engine

CR 509.1 makes the declaration one turn-based action. Here, blocking is
a per-pair verb any defending seat may send while the declare blockers
priority window is open (`blockers.go:5-27`), with no "done" step. The
enumerator offers per-creature moves and a bot composes a declaration
by re-enumerating after each one (`aiseat/heuristic/combat.go:13-22`).
The client's two-click flow sends one `declare_blocker` per pair
(`client/src/routes/Game.svelte:700-708`). Any count rule has to work
with that, or change it.

Two throwaway probe tests on `684f2786`, deleted afterwards, showed
how lax the per-pair verb is:

| probe | rules | engine |
|---|---|---|
| a creature is declared blocking attacker A, then re-pointed at attacker B | "whenever this blocks" triggers once (CR 509.3a) | two `EventBlock` events for the same blocker |
| 3 players; seat 2's creature blocks an attacker aimed at seat 1 | illegal (CR 509.1a: the defending player's creatures) | accepted |
| a tapped creature blocks | illegal (CR 509.1a) | accepted |

The client reaches all three: the context menu offers "Re-declare
blocker" and lists every attacker not controlled by the card's
controller (`client/src/lib/contextMenu.logic.ts:788-805`), and the
two-click flow lets a tapped creature be selected
(`PlayerPanel.svelte:286-293`).

#### The client can't say why

A refused block comes back as `bad_request` with the generic
`ErrIllegalBlock` text, "blocker cannot legally block this attacker"
(`errors.go:302`, `ws/hub.go:925-990`), in the `lastError` banner
(`Game.svelte:1307-1312`). Nothing on the board shows which attackers a
selected creature can block before the click.

### Decisions

Decisions are numbered on from §6.

#### 7. Block legality is a method on `Game`, lock-held and read-only

```go
// Caller holds g.mu (read or write) and layers are fresh.
// Never recomputes layers, never writes.
func (g *Game) BlockPairRefusalLocked(attacker, blocker *Card) BlockRefusal
func (g *Game) CanBlockLocked(attacker, blocker *Card) bool // == BlockOK
```

The free function `CanBlock(attacker, blocker)` is **deleted**, not kept
as a wrapper. A second entry point that can't see the game is exactly
the drifting copy §3 exists to prevent, and `unused` would flag it
anyway. All three callers, the tests and the card-file comments move in
PR 1.

- **The defending player is derived, not passed.** It is
  `defendingPlayerForAttackLocked(attacker.AttackingTarget)`
  (`attack_target.go:102`): the attacked player, a planeswalker's
  controller, or a battle's protector, one attacker at a time
  (CR 506.2, 509.1a). An attacker that isn't attacking has no defending
  player, and [Decision 13](#13-the-stored-declaration-is-always-legal-and-the-menace-close-out-is-deleted)
  refuses the block for that reason before any rule reads it.
- **Read-only is a contract with a test**, not a hope. The enumerator
  and the view projection call it from inside `ReadSnapshot`, and
  `RecomputeLayersIfStaleLocked` is a write. PR 1 adds
  `TestBlockLegalityDoesNotMutate`: encode a snapshot, run the pair
  function, the set validator and the option generator over every pair
  on a busy board, encode again, and compare the bytes.
- **Card code gets the same `*Game`** that `StaticAbility.AppliesTo`
  already receives. A rule that writes through it is a bug that the
  test above catches on the cards it covers.

#### 8. A refusal carries a reason, and the reason reaches the player

```go
type BlockRefusal struct {
	Reason BlockReason // "" means legal
	Source uuid.UUID   // the permanent whose text refuses it, when there is one
	N      int         // the bound, for count reasons
}
```

`BlockReason` values are stable snake_case tokens, like
`Restriction.Names()`: `cant_block`, `cant_be_blocked`, `flying`,
`landwalk`, `cant_be_blocked_by`, `cant_be_blocked_except_by`,
`cant_block_attacker`, `too_few_blockers`, `too_many_blockers`,
`declaration_limit`, `not_defending` (claimed by #1339, Decision 34), `tapped`, and a reserved
`protection`.

The engine returns a `*BlockRefusedError` that wraps `ErrIllegalBlock`,
so `errors.Is` callers keep working. `classifyActionError` turns it into
a new error code, `illegal_block`, with a sentence built server-side
("Cold-Eyed Selkie has islandwalk and you control an Island"), plus two
additive omitempty `ErrorPayload` fields: `reason` (the token) and the
existing `card_id` (the blocker). The client never builds the sentence,
so it never re-derives a rule (§6).

#### 9. One order of checks, with a reserved slot for protection

`BlockPairRefusalLocked` runs these in order and returns the first
refusal. The order decides only which reason is reported, never whether
the pair is legal.

| # | check | reads | rule |
|---|---|---|---|
| 1 | the restriction bits: `CantBlock` on the blocker, `CantBeBlocked` on the attacker | both cards | CR 509.1b (§2) |
| 2 | evasion keywords: flying, then landwalk ([Decision 10](#10-landwalk-is-a-closed-token-set-read-from-effective-characteristics)). Fear, intimidate, shadow, horsemanship and skulk join this slot in [#825](https://github.com/krakenhavoc/cmd_and_ctrl/issues/825) (owner question 3) | both cards, the defending player's lands | CR 702.9b, 702.14c |
| 3 | **protection, reserved for #662**: an attacker with protection from a quality the blocker has | both cards | CR 702.16f |
| 4 | block rules from permanents and from until-end-of-turn effects ([Decision 11](#11-rules-with-a-parameter-are-block-rules-read-from-their-sources-at-check-time)) | the pair, the rule's source, anything else | CR 509.1b |

Slot 3 is an empty, commented case in PR 1, so #662 adds one call in a
known place and doesn't re-open the ordering. Protection reads the
blocker's effective colours, types and controller, all of which this
function already has, so the blocking half of #662 needs nothing more
from this addendum. Its targeting half is a different refactor (#662's
table).

Count rules are not in this function. They belong to a set of blocks,
not a pair ([Decision 12](#12-block-counts-are-bounds-on-a-declaration-not-pair-checks)).

#### 10. Landwalk is a closed token set, read from effective characteristics

- **Tokens.** `islandwalk`, `swampwalk`, `forestwalk`, `mountainwalk`,
  `plainswalk` and `nonbasic landwalk` join `canonicalKeywords`. The
  rarer variants (CR 702.14a/c: `legendary landwalk`, `snow swampwalk`,
  `desertwalk`, `artifact landwalk`) join with their first card, as the
  table's closedness requires (`keywords.go:44-53`).
- **One reader.** `landwalkRequirement(token) (landwalkSpec, bool)`
  parses the token into a land subtype and/or a supertype and whether
  it is negated ("nonbasic" is "without the basic supertype",
  CR 205.4c). The pair check asks: does the defending player control at
  least one land, by effective characteristics, that matches? Several
  landwalk abilities are checked separately and don't cancel
  (CR 702.14d).
- **Effective, both sides.** A granted islandwalk (Lord of Atlantis,
  layer 6) is read through `HasKeyword`. A land's types are read
  through `Card.IsLand` / `HasSubtype`, which are post-layer-4, so
  Urborg, Tomb of Yawgmoth ("each land is a Swamp in addition to its
  other land types", already in the catalog) turns on swampwalk, and a
  Blood Moon Mountain turns on mountainwalk. `Card` has no supertype
  reader today, so PR 3 adds `Card.HasSupertype` over
  `Characteristic.Supertypes`.
- **The importer needs no change.** It keeps Scryfall keywords that are
  in the table after lower-casing, so Cold-Eyed Selkie's
  `["Landwalk", "Islandwalk"]` imports as `islandwalk`. An uncatalogued
  card that prints landwalk starts being enforced in the same PR, which
  is the table's rule.
- **The badge already works.** `KeywordBadgeRow.svelte` shows an
  unknown token as a three-letter text badge with the full name as its
  tooltip ("ISL"). An icon is optional polish.

#### 11. Rules with a parameter are block rules, read from their sources at check time

A rule like "can't be blocked except by Walls" isn't a characteristic of
the attacker, so it doesn't go into `Characteristic`. The engine derives
it when it checks a block, the way ADR 0048 §1 derives cost modifiers:

```go
type BlockRule struct {
	// Exactly one of these is set.
	Pair  func(g *Game, attacker, blocker, source *Card) bool // true refuses the pair
	Count func(g *Game, attacker, source *Card) (min, max int) // 0 means no bound
	Limit func(g *Game, source *Card) (maxBlockers int)        // whole declaration; built with its first card (Decision 12)

	Reason BlockReason
	Label  string
}
```

- **Where they come from.** `Spec.BlockRules []game.BlockRule`, wired
  through a new `CatalogBlockRules` hook and read by walking the
  battlefield with `CatalogAbilityKey` (AGENTS.md §7: new readers of a
  "what does this permanent do" hook use that key). Until-end-of-turn
  rules (Gingerbrute) go in a new `Game.TurnScopedBlockRules`, next to
  `TurnScopedStatics`.
- **Swept at turn end, inside ADR 0059's `sweepTurnEndLocked`.**
  `ClearTurnScopedBlockRulesLocked` empties the registry wholesale (a
  fresh slice, as `ClearTurnScopedReplacementsLocked` does) and joins
  the non-interactive cleanup sweep that
  [ADR 0059](https://github.com/krakenhavoc/cmd_and_ctrl/pull/824)
  Decision 6 moves out of the `StepCleanup` case. So a rule also ends
  when a departed active player's turn ends through the rotation seam,
  and when the sandbox's `PassTurn` ends a turn. The registry carries no
  turn stamp, so it needs no entry in ADR 0059 Decision 10's conversion
  list. If this addendum's PR 4 lands before ADR 0059's sub-PR 1, the
  call sits beside `ClearExpiredTurnScopedStaticsLocked` in cleanup, and
  ADR 0059's sub-PR 1 moves it with the others.
- **#755 is not expected to fold it in.** #755 gives
  `TurnScopedStatics` and `TurnScopedReplacements` durations longer than
  a turn. No card in this wave, and none found while drafting, prints a
  block rule that outlasts the turn, so `TurnScopedBlockRules` stays an
  until-end-of-turn registry. #755 extends it only if such a card
  arrives, and it would use #755's duration model rather than a second
  one.
- **Losing abilities works without extra code.** Legolas Greenleaf's
  rule is Legolas's own ability. If Legolas loses all abilities,
  `CatalogAbilityKey` returns `""` and the rule is gone. Prowler's
  Helm's rule is the Equipment's ability. The equipped creature losing
  its abilities doesn't touch it, as §1's third argument requires for
  Pacifism.
- **Scopes are builders, not fields.** `effects/block_rules.go` provides
  the attacker scopes `OnSelf`, `OnAttached` and `OnMatching(pred)`,
  the blocker scope `BlockerMatching(pred)` and the controller scope
  `ControlledBySourceController`. On top of those sit the rule
  builders: `CantBeBlockedExceptBy(pred)`, `CantBeBlockedBy(pred)`,
  `CantBlockAttackers(pred)` (Gornog's shape), `MaxBlockers(n)`,
  `MinBlockers(n)`, `CantBeBlockedWhile(pred)` (Thieves' Tools), and
  `…UntilEOT` twins that register a turn-scoped rule for a snapshotted
  set (CR 611.2c), as `RestrictUntilEOT` does.
- **A builder ships with its first card** (owner policy, 2026-09-17:
  every new seam path ships with at least one real card). The first
  wave uses `CantBeBlockedExceptBy`, `CantBeBlockedBy`, `MaxBlockers`,
  `CantBeBlockedWhile`, one `…UntilEOT` twin (Gingerbrute's) and the
  scopes they need. `CantBlockAttackers` (Gornog) and `MinBlockers`
  (Pathrazer of Ulamog) are not built until their cards ship. Each
  builder lands in the PR that ships its first card: PR 4 for
  `CantBeBlockedExceptBy` on `OnAttached` (Prowler's Helm),
  `MaxBlockers` (Hungering Hydra) and Gingerbrute's `…UntilEOT` twin,
  PR 6+ for the rest.
- **Everything is read live at declaration.** Predicates read effective
  characteristics after layer 7: the blocker's power for Legolas,
  Champion of Lambholt's own power for its threshold. A pump after the
  block is declared changes nothing (§5, CR 509.1b).
- **Persistence.** `TurnScopedBlockRules` holds closures. It is counted
  in `ContinuationCensus` and marked `dropped` in the drift test,
  exactly like `TurnScopedStatics` (`snapshot_drift_test.go:124`).
  `Spec.BlockRules` is catalog data and needs nothing.

The `Restriction` bits stay as they are. They are the cheap, common
case, and the client reads them (§6).

#### 12. Block counts are bounds on a declaration, not pair checks

`BlockerCountValid` is deleted. In its place:

```go
// min 0 = no minimum; max 0 = no maximum.
func (g *Game) blockerBoundsLocked(attacker *Card) (min, max int, src uuid.UUID)
```

- **Menace** is built in as a minimum of 2 (CR 702.111b). Rules add their
  own bounds: Pathrazer of Ulamog's minimum of 3, Hungering Hydra's
  maximum of 1. The effective minimum is the largest minimum and the
  effective maximum is the smallest maximum.
- **A minimum greater than the maximum makes the attacker unblockable**
  (a menace creature wearing Vorrac Battlehorns). The option generator
  offers nothing for it, and every block is refused: one blocker as
  `too_few_blockers`, two or more as `too_many_blockers`.
- **Whole-declaration limits** (Silent Arbiter: "no more than one
  creature can block each combat") are `BlockRule.Limit`, checked
  against every stored block in the combat. No first-wave card uses it,
  so under the owner's policy the `Limit` field and its check are
  **not built in PR 4**. They land with Silent Arbiter, which also needs
  the attack-side count limit that Decision 18 leaves out. The validator
  in Decision 13 is written as a list of set checks, so the limit is one
  more entry there. *(Built by [#1507](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1507):
  [Decision 43](#decision-43-blockrulelimit-is-a-third-set-check-in-the-block-validator).)*

#### 13. The stored declaration is always legal, and the menace close-out is deleted

The engine never holds a block it would refuse. Count rules can be
checked before a block is stored because a block that needs several
creatures is declared in one action:

```go
type BlockDeclaration struct{ Blocker, Attacker uuid.UUID }
func (g *Game) DeclareBlockers(decls []BlockDeclaration) error
```

- **The new wire action is `declare_blockers`**, with params
  `{blocks: [{blocker, attacker}, …]}`. `declare_blocker` becomes
  exactly a one-entry `declare_blockers`. Both are one `room.Apply`,
  so one undo takes the whole action back, as with `declare_attackers`.
- **The set is validated as it will be after the action.** The engine
  starts from the blocks already stored, applies the new entries (an
  entry for a creature that is already blocking re-points it), and
  checks: CR 509.1a for each entry (the blocker is an untapped creature
  controlled by the attacker's defending player, and the attacker is
  attacking), the pair function for each new or changed pair, the
  bounds for **every attacker whose set of blockers changed** (including
  one that loses a re-pointed blocker), and the whole-declaration
  limits once they exist (Decision 12).
- **All or nothing.** If anything is refused, nothing is stored, no
  event is emitted, and the first refusal is returned. This is
  different from `DeclareAttackers`, which skips ineligible entries,
  because a two-creature menace block is legal only as a pair.
- **Events come after the commit**, one `EventBlock` per new pairing in
  entry order. A re-point emits no second `EventBlock`: a creature
  blocks once per combat for "whenever this blocks" (CR 509.3a). A
  wrong "becomes blocked" trigger on the attacker the creature left is
  a separate bug (see [Stale comments and follow-ups](#stale-comments-and-follow-ups)).
- **CR 509.1a is now enforced by the rules verb.** Until now the verb was
  lax, as the attacker side still is for sandbox hand-forcing (§3). A
  count rule needs the stored set to be the real declaration, though:
  a tapped creature or another player's creature would count as one
  of menace's two, and as Silent Arbiter's one. Pre-emptive blocks
  against a creature that isn't attacking (the sandbox note on
  `DeclareBlocker`, `mutations.go:4360-4361`) are refused too. The
  sandbox keeps `clear_combat` and undo.
- **What isn't re-checked.** Attackers that the action doesn't touch are
  not re-checked. An attacker that gains menace after one legal block
  keeps that block (CR 509.1b), and so does one whose second blocker
  dies (CR 509.1h, #715).
- **The menace close-out in `assignAndDealCombatDamageLocked`
  (`mutations.go:4573-4605`) is deleted in the same PR** that makes
  declaration refuse the lone block. It can't go earlier, because
  without the declaration check a lone block would stand, and a card
  would play stronger than printed for the defender. #715 keeps its
  "blocked" flag and damage work, drops its menace item, and sets the
  flag in the same commit function this decision adds. Whichever of
  the two PRs lands second rebases onto the other.
- **Nothing is checked at step exit**, so ADR 0014 §9's close-out has
  no successor. ADR 0014 §9's reason ("they're halfway through choosing
  blockers and get a red error") is answered in the client by
  [Decision 16](#16-the-client-shows-server-data-and-never-derives-a-block-rule).

#### 14. One option generator, shared by the enumerator and the #328 signal

This answers §"What this deliberately does not express"'s open question.

```go
type BlockOption struct{ Blocks []BlockDeclaration }
// Every option passes the same validator DeclareBlockers runs.
func (g *Game) BlockOptionsLocked(seat uuid.UUID, perAttackerCap int) []BlockOption
```

- **Singles**: one entry for each eligible blocker and each attacker it
  may join alone. That means the pair is legal, the attacker's maximum
  isn't reached, and the minimum is at most 1 or the attacker is already
  blocked by at least its minimum.
- **Minimum groups**: for an unblocked attacker whose minimum is `m ≥ 2`,
  combinations of exactly `m` pair-legal eligible blockers, in
  battlefield order, lexicographic, capped at `perAttackerCap`. Blockers
  beyond the minimum join later as singles. Before an option is
  returned, the validator from Decision 13 runs on it, so a
  whole-declaration limit or a maximum can never produce an option the
  engine refuses.
- **The enumerator** maps a single to `declare_blocker` and a group to
  `declare_blockers`, both `KindBlock`, with `perAttackerCap` =
  `Options.MaxExpansionPerSource` (12). `dispatchAll` holds soundness.
  Completeness is not promised (`legal.go:9-18`): on a wide board, a
  bot may not be offered the best pair. That trade is recorded under
  Consequences.
- **The #328 signal** becomes `len(g.BlockOptionsLocked(seat, 1)) > 0`.
  A defender whose only creature faces a lone menace attacker no longer
  owes a decision, and the note at `blockers.go:74-80` is deleted rather
  than corrected. `aiseat.Runner.shouldHoldForBlockers` already asks the
  enumerator, so it agrees automatically.
- **The heuristic** scores a group move as one block by `m` creatures.
  It kills the attacker if their total power reaches its toughness, and
  it loses the blockers the attacker's power can kill, cheapest first.
  The random tier needs nothing. The model tier's escalation already
  groups `KindBlock`.

#### 15. Requirements (CR 509.1c): a slot and a checkpoint, built with their first card

No first-wave card prints a blocking requirement. Lure ("all creatures
able to block enchanted creature do so"), "blocks each combat if able"
and Tromokratis are all out of scope. The design is fixed now so that
Decisions 12-14 don't rule it out:

- **Requirements are counted over a complete declaration, never per
  pair.** A partial declaration always obeys fewer requirements, so no
  per-pair refusal can express them.
- **The validator is a pure function of a hypothetical set**
  (Decision 13 builds the proposed set before it checks anything). CR
  509.1c's "maximum possible number of requirements that could be obeyed
  without disobeying any restrictions" is a search over
  restriction-legal sets, and that search needs to evaluate sets
  without storing them.
- **The checkpoint is the defending seat's `pass_priority` in declare
  blockers.** Every step exit requires every seat's pass, so it is the
  one moment the engine knows that the seat is done. A pass that obeys
  fewer requirements than the maximum is refused with a reason. The
  search returns a witness set, the enumerator offers the missing
  blocks as one `declare_blockers` move, and it withholds `pass` until
  the maximum is met, so a bot can always make progress (#544). Blocks
  aren't priority-gated, so a `declare_blockers` from a seat that has
  already passed in this round is checked against the same maximum.
- **Blocking costs (CR 509.1d)** don't count toward the maximum
  (509.1c), and nothing in the engine models them.

#### 16. The client shows server data and never derives a block rule

§6 applies unchanged. Everything the client needs is stamped by the
server:

- `CardView.blockable_attackers []string`: for each creature that
  `BlockerEligible` passes during declare blockers, the attackers
  `CanBlockLocked` allows. It holds pair legality only, with no counts.
  It is public, because both the attackers and the blockers are on the
  board.
- `CardView.blockers_min` / `blockers_max` on each attacking creature,
  omitted when 0.
- The `illegal_block` error from Decision 8 for anything the client
  still gets wrong.

How the board uses these (owner-decided, [questions 1 and 2](#decided-2026-09-17)):

- **Multi-blocker blocks are staged on the board.** Clicking a blocker
  and then an attacker whose `blockers_min` is 2 or more (menace), and
  which isn't already blocked by at least that many, stages the pair
  instead of sending it: a dashed arrow is drawn, and the
  combat-hint banner reads "needs N more blocker(s)". Each further
  blocker clicked onto that attacker stages another dashed arrow. When
  the staged count reaches the minimum, the client sends **one**
  `declare_blockers` with the whole group, and the arrows become the
  server's solid ones. Clicking a staged blocker again unstages it, and
  leaving declare blockers (or the attacker leaving combat) discards
  the staging. Staging is client state only: nothing is sent, logged or
  undone until the group is complete, so the player never sees an error
  for a block they haven't finished choosing.
- **Attackers the selected blocker can't block are dimmed.** In block
  mode, once a blocker is selected, attackers not in its
  `blockable_attackers` are dimmed. A click on a dimmed attacker is
  still sent, and the banner explains it with the server's
  `illegal_block` sentence, so the reason always comes from the server.
  The same banner carries the selection and staging instructions.
  Attackers with a minimum show "needs N blockers" on hover.
- **No per-pair reason on the wire** (option (c) was not chosen), no log
  line beyond today's block entries, and no reveal-strip cue.

The staging and dimming rules are one pure client module
(`client/src/lib/blockStaging.ts`) reading only the three stamps, so §6
holds: the client never re-derives a block rule.

#### 17. The bot's attack-side estimate stays an estimate

`heuristic.couldBlock` (`aiseat/heuristic/combat.go:299-307`) guesses,
while planning attacks, who could block. No declaration exists yet, so
there is no move list to read. It may approximate, because it only
ranks moves the enumerator has already made legal. PR 3 teaches it
landwalk from the public view (the defender's lands). PR 2 fixes its
restriction read: it looks for `"can't block"` in `abilities`, which
never contains that string. The restriction is in `restrictions` as
`cant_block`.

#### 18. Out of scope, stated

- Tromokratis's "can't be blocked unless all creatures defending player
  controls block it". This is a set rule that counts non-blockers, and
  it follows the requirements slot.
- "Can't attack or block alone". No attributed card.
- Blocking an additional creature (Brave the Sands, whose caveat stays),
  since `BlockingTarget` is a single ID.
- Attack-side count limits (Crawlspace, Silent Arbiter's first line).
  `DeclareAttackers` already receives the set, and the same all-or-nothing
  shape would work, but no card in this wave needs it. *(Built by
  [#1507](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1507):
  [Decision 44](#decision-44-the-attack-side-twin-is-gameattacklimit-a-scope-and-a-number).)*
- Creatures put onto the battlefield blocking (CR 509.4b) and "becomes
  blocked" by an effect.
- Goad (§"What this deliberately does not express" still applies).
- Protection itself (#662). This addendum only reserves its slot.
- Flanking (Sidar Kondo of Jamuraa) and Gornog's Coward type change.

### Consequences

- **A lone block on a menace attacker is refused when it's made**, with a
  reason, and never fires a trigger. Tests that relied on the silent
  revert are rewritten: `TestCombatMenaceRevertsSingleBlocker`
  (`combat_test.go:189`) becomes a declaration-time refusal test.
- **Blocking gets stricter at the rules verb.** Tapped blockers, blocks by
  a player who isn't the defending player, and pre-emptive blocks are
  refused. Any test or sandbox flow that used them has to change. The
  enumerator was already this strict, so bots don't change.
- **The wire grows**: one action (`declare_blockers`), one error code
  (`illegal_block`) with an omitempty `reason`, and three omitempty
  `CardView` fields. `docs/protocol.md` and `protocol.ts` are updated in
  the PRs that add them.
- **The engine and the bots always agree on blocks.** Singles and
  minimum groups come from one generator that runs the engine's own
  validator.
- **A bot may miss the best group on a wide board**, because groups are
  capped at 12 per attacker in battlefield order. The alternative is
  either an uncapped combinatorial move list or strategy inside the
  engine.
- **Landwalk turns on for imported cards** in the same PR as the lords, per
  the closed-table rule.
- **`docs/engine-seams.md`**: the "Conditional blocking restrictions" row
  (line 116) closes with PR 4. The landwalk half of the "Landwalk, fear,
  intimidate …" row (line 120) closes with PR 3, and the rest of that
  row stays open, tracked on
  [#825](https://github.com/krakenhavoc/cmd_and_ctrl/issues/825)
  (PR 3 adds it to the row's Tracked column).
- **Two card caveats change.** Lord of Atlantis's islandwalk caveat goes.
  Its first caveat ("Only Merfolk YOU control get +1/+1") is already
  wrong on develop: `tribal_test.go:90` asserts that the opponent's
  Merfolk is pumped. It goes in the same PR.

### Alternatives considered

- **Keep the step-exit close-out, add a log line.** Rejected. Triggers
  have already fired on the block by then (CR 733.1), and the
  close-out undoes legal blocks too (#715).
- **Refuse the lone menace block in `DeclareBlocker` without a group
  action.** Rejected. A per-pair client or bot could then never block
  a menace creature, which plays stronger than printed for the attacker.
- **Provisional blocks** that the server accepts but doesn't count until
  the minimum is met, dropped at step exit if still short. Rejected.
  The engine would need a third block state that the view, the arrows,
  the log, #715's flag and the triggers all have to understand. It also
  keeps a silent step-exit drop for a defender who blocks again after
  passing, while the active player still holds priority.
- **`declare_blockers` replaces the seat's whole set** (so removals are
  expressible). Rejected. Taking back a block that was already announced
  would leave its triggers on the stack. Undo already exists for that,
  and it restores the triggers too.
- **A `Characteristic.BlockRules` field written in the layer pass.**
  Rejected. It would hold closures, which breaks
  `TestEmbeddedDomainTypesStayPureData` and the by-value snapshot of
  `Characteristic`. It would also make a rule the attacker's ability,
  so "loses all abilities" would strip Prowler's Helm's effect, which
  §1 rejects for Pacifism.
- **A precomputed `BlockContext` passed to a pure two-card function.**
  Rejected. The rules read arbitrary state (the defending player's
  lands, a third card's live power, whole-declaration counts), so the
  context would be the game under another name, and every caller would
  have to build it the same way.
- **Pass the defending player explicitly.** Rejected for the engine
  callers, because the attacker already names it and a second source
  could disagree. The bot's attack planner is the one place without a
  declaration, and Decision 17 keeps it an estimate.
- **Enforce requirements per pair.** Impossible (Decision 15).

### PR split

**PR 1 — engine: the game-aware signature. No behaviour change.**
- `BlockPairRefusalLocked`, `CanBlockLocked`, `BlockRefusal`,
  `BlockReason`, `BlockRefusedError`. Delete `CanBlock`. Move the three
  callers and all tests. Leave the protection slot as a commented empty
  case.
- `classifyActionError` → `illegal_block` with the sentence, and
  `ErrorPayload.reason`.
- `TestBlockLegalityDoesNotMutate`.
- Fix-list items 1-4 below.
- Checks: `go test ./internal/game/... ./internal/legal/... ./internal/ws/... ./internal/aiseat/...`,
  then `go test ./...` and `make lint`.

**PR 2 — block declarations as sets; menace at declaration.** Depends on PR 1.
- `DeclareBlockers`, the `declare_blockers` action, `DeclareBlocker` as a
  one-entry call, the CR 509.1a checks, the post-commit events, and no
  re-announce on a re-point.
- `blockerBoundsLocked` with menace built in. Delete `BlockerCountValid`
  and its test.
- `BlockOptionsLocked`, the enumerator's singles and groups, and the #328
  signal on top of it.
- Delete the menace close-out. Rewrite `TestCombatMenaceRevertsSingleBlocker`.
- Heuristic group scoring and the `couldBlock` restriction read.
- The `declare_blockers` entry in `docs/protocol.md` and `protocol.ts`.
- Fix-list items 5-9 below.

**PR 3 — landwalk (#705).** Depends on PR 1. Can run in parallel with PR 2.
- Tokens, `landwalkRequirement`, `Card.HasSupertype`, slot 2 of the pair
  function.
- Cards: the landwalk group of the first wave (owner question 3, option
  (b)): Lord of Atlantis, Elvish Champion, Goblin King, Master of the
  Pearl Trident, Cold-Eyed Selkie and Trailblazer's Boots. Lord of
  Atlantis loses both caveats. Every landwalk path (a basic land type,
  a granted keyword, nonbasic landwalk) ships with a real card here.
- `couldBlock` learns landwalk. Update the engine-seams row.
- Fix-list item 10.

**PR 4 — block rules (#750, engine).** Depends on PR 2.
- `game.BlockRule` (`Pair` and `Count`; no `Limit`, Decision 12),
  `Spec.BlockRules`, `CatalogBlockRules`, `TurnScopedBlockRules` with its
  census entry, drift-test row and `ClearTurnScopedBlockRulesLocked` in
  the turn-end sweep (Decision 11), pair slot 4, and bounds read from
  rules.
- `effects/block_rules.go` with the scopes, and **one reference card per
  path** so none lands without a real card: Prowler's Helm (a catalog
  pair rule, `CantBeBlockedExceptBy` on `OnAttached`), Hungering Hydra
  (a count rule, `MaxBlockers(1)`) and Gingerbrute (the turn-scoped
  registry, through its `…UntilEOT` twin). The other builders land with
  their first cards in PR 6+.
- Agreement tests in `legal/restrictions_test.go`.

**PR 5 — client.** Depends on PR 2. The max stamp is live after PR 4.
- The `blockable_attackers`, `blockers_min` and `blockers_max` stamps
  (server side).
- The board behaviour decided in owner questions 1 and 2 (Decision 16):
  staged multi-blocker blocks with dashed arrows and "needs N more
  blocker(s)", one `declare_blockers` when the minimum is met, dimmed
  attackers for the selected blocker, and the `illegal_block` banner
  text.
- The context menu offers only attackers in `blockable_attackers`.

**PR 6+ — cards (#750's first wave).** The rest of the pair and count
cards listed under [Card first wave](#card-first-wave). The other evasion
keywords are not here: they are
[#825](https://github.com/krakenhavoc/cmd_and_ctrl/issues/825), with its
own PR and soak run. Each card follows AGENTS.md §7: declared
completeness, caveats only ever weaker than printed, and a re-check of
every other clause.

### Test plan

Engine (`internal/game`):

1. **Pair function parity.** Every case in today's `TestCanBlock`
   (`keywords_test.go:205`) gives the same answer through
   `CanBlockLocked`.
2. **Read-only.** `TestBlockLegalityDoesNotMutate` (Decision 7).
3. **Landwalk.** For each basic type: blocked with no land of the type,
   unblockable with one. Nonbasic landwalk against a basic Island (legal
   to block) and a nonbasic land (not). Urborg makes swampwalk bite.
   A granted islandwalk (Lord of Atlantis) is enforced, and it stops when
   the Lord leaves. Two landwalk abilities are checked separately.
   Landwalk against a planeswalker attack reads the planeswalker
   controller's lands, and against a battle, the protector's.
4. **Menace at declaration.** A lone block is refused with
   `too_few_blockers`, emits no `EventBlock`, and doesn't trigger Savvy
   Hunter. A two-creature `declare_blockers` is accepted. A third
   blocker then joins as a single. Re-pointing one of the two away is
   refused.
5. **No step-exit revert.** Menace blocked by two, one destroyed before
   damage: the survivor keeps the block and takes the damage. Coordinate
   with #715's version of this test.
6. **Counts.** Maximum 1: the second single is refused. Minimum 3:
   only groups of 3 are offered. Menace plus maximum 1: no option, and
   every block is refused.
7. **CR 509.1a.** A tapped blocker, another player's blocker, and a
   block against a creature that isn't attacking are each refused.
8. **No double announce.** A re-point emits no second `EventBlock`.
9. **Block rules** (PR 4): one test per rule kind with a stubbed
   rule, and one per builder in the PR that ships it. A threshold changed by
   a pump after declaration leaves the block standing (§5). A rule from
   a source that loses all abilities stops applying. An Equipment's rule
   survives the host losing its abilities. An until-end-of-turn rule
   ends at cleanup and is counted in `ContinuationCensus`, and (once
   ADR 0059's seam exists) it also ends when the active player leaves
   mid-turn.
10. **All or nothing.** One refused entry among three stores none of
    them.

Enumerator and bots:

11. `legal/restrictions_test.go`: every enumerated block move is accepted by
    `dispatchAll`, and the engine really refuses each withheld move, for
    landwalk, menace, a minimum of 3, a maximum of 1 and each pair
    builder.
12. The #328 signal and `BlockOptionsLocked(seat, 1)` agree on every
    fixture above. A lone creature against a lone menace attacker owes
    no decision.
13. The group cap: 8 eligible blockers against a menace attacker give
    at most 12 group moves, all dispatchable.
14. `aiseat`: the heuristic picks a group move when the pair kills the
    attacker, and the random tier answers group moves. The catalog soak
    (#601) with the first-wave cards added has no stalls in declare
    blockers.

Wire and client:

15. `ws`: `classifyActionError` maps each `BlockReason` to
    `illegal_block` with a non-empty sentence.
16. `protocol`: the three `CardView` stamps are present only during
    declare blockers and match `CanBlockLocked` and `blockerBoundsLocked`.
17. Client: `blockStaging.ts` (Decision 16) is a pure module with
    vitest coverage: staging and unstaging, the "needs N more" count, one
    `declare_blockers` payload exactly when the minimum is met, discard on
    leaving the step, and the dimmed set for a selected blocker. The
    dashed arrows and dimming are checked by hand until #689.

### Card first wave

Oracle text below was checked against the dump. Catalog status is
`origin/develop` at `684f2786`. Each card PR re-checks the card for gaps
outside this seam. The list is owner question 3's option (b), decided
2026-09-17.

**Landwalk (PR 3), all unblocked by this seam alone:**

| card | oracle text (the relevant clause) | today |
|---|---|---|
| Lord of Atlantis | "Other Merfolk get +1/+1 and have islandwalk." | in catalog, `CompletenessCaveats`, with two caveats: islandwalk inert, and a stale anthem caveat |
| Elvish Champion | "Other Elf creatures get +1/+1 and have forestwalk." | in catalog (`tribal_lords.go:56-60`), forestwalk omitted, completeness unreviewed |
| Goblin King | "Other Goblins get +1/+1 and have mountainwalk." | in catalog (`tribal_lords.go:63-66`), mountainwalk omitted, completeness unreviewed |
| Master of the Pearl Trident | "Other Merfolk creatures you control get +1/+1 and have islandwalk." | not in catalog |
| Cold-Eyed Selkie | "Islandwalk … Whenever this creature deals combat damage to a player, you may draw that many cards." | not in catalog |
| Trailblazer's Boots | "Equipped creature has nonbasic landwalk. … Equip {2}" | not in catalog |

**Pair rules (Prowler's Helm and Gingerbrute in PR 4, the rest in PR 6+):** Legolas Greenleaf ("can't be blocked by creatures
with power 2 or less"), Prowler's Helm ("Equipped creature can't be
blocked except by Walls"), Shifting Sliver ("Slivers can't be blocked
except by Slivers"), Champion of Lambholt ("Creatures with power less
than this creature's power can't block creatures you control", plus its
+1/+1 counter trigger), Gingerbrute ("{1}: This creature can't be
blocked this turn except by creatures with haste"), Thieves' Tools
("Equipped creature can't be blocked as long as its power is 3 or
less", plus its Treasure on entering), Wrecking Ball Arm ("base power
and toughness 7/7 and can't be blocked by creatures with power 2 or
less"; "Equip legendary creature {3}" already exists on Blackblade
Reforged).

**Counts (Hungering Hydra in PR 4, the rest in PR 6+):** Hungering Hydra ("can't be blocked by more than one
creature"), Vorrac Battlehorns ("has trample and can't be blocked by
more than one creature"), Alpha Authority ("has hexproof and can't be
blocked by more than one creature"), Goblin War Drums ("Creatures you
control have menace", the one card whose whole text is a group-block
test).

**Not in the wave:** fear, intimidate, shadow, horsemanship and skulk
cards ([#825](https://github.com/krakenhavoc/cmd_and_ctrl/issues/825)).
Pathrazer of Ulamog (annihilator 3) and Signal Pest
(battle cry) need card-side attack triggers first. Tromokratis, Silent
Arbiter, Crawlspace, Lure, Sidar Kondo of Jamuraa and Gornog, the Red
Reaper are excluded by Decision 18.

### Stale comments and follow-ups

Fix-list. Each item names the PR that fixes it.

1. `keywords.go:283-296`: the `CanBlock` doc lists "menace, fear, shadow"
   as enforced evasion and points at `BlockerCountValid`. PR 1.
2. `keywords.go:332-344`: "Called at the declare-blockers step close-out,
   per CR 702.110". The citation is wrong (menace is CR 702.111) and so
   is the claim. The comment is deleted with the function in PR 2, and
   the citation is fixed in PR 1 in case PR 2 slips.
3. `errors.go:295-301`: the `ErrIllegalBlock` doc names menace, fear and
   shadow. PR 1.
4. `mutations.go:4393-4397`: "evasion keywords (flying, menace, fear,
   shadow, …)". PR 1.
5. `restrictions.go:64-83`: the count-restriction and landwalk bullets.
   Rewritten to point here in PR 2 (counts) and PR 3 (landwalk).
6. `blockers.go:74-80`: the menace note is deleted (Decision 14). PR 2.
7. `indestructible.go:13`: "read by `CanBlock` / `BlockerCountValid`". PR 2.
8. `effects/boggart_brute.go:5-8`: "the combat close-out validator silently
   reverts the block". PR 2.
9. `AGENTS.md:1003-1007`, ADR 0045 lines 170-173 above, `mutations.go:4564-4567`
   (the menace paragraph of the damage doc), and a "Superseded by ADR 0045
   addendum" note under ADR 0014 §9 and its summary bullet (0014 lines
   212-228 and 318-320). PR 2.
10. `effects/tribal.go:24-34`, `effects/tribal_lords.go:22-28` and
    `effects/lord_of_atlantis.go:21-26`: "landwalk is not modelled". PR 3.

Bugs found while writing this, all fixed by the PRs above unless noted:

- The rules verb accepts a tapped blocker and a block by a player who
  isn't the defending player (probed; Decision 13, PR 2). The context
  menu offers the second in multiplayer (`contextMenu.logic.ts:788-790`),
  and the two-click flow offers the first.
- Re-pointing a blocker emits a second `EventBlock`, so "whenever this
  blocks" triggers twice (probed; Decision 13, PR 2). An attacker that
  loses its only blocker to a re-point has already fired any "becomes
  blocked" trigger, and once #715's flag exists it would stay marked as
  blocked. That half is **not** fixed here and needs its own issue.
- `heuristic.couldBlock` reads `"can't block"` from `abilities`, where
  it never appears (Decision 17, PR 2).
- Lord of Atlantis's first caveat describes a limitation the card no
  longer has (PR 3).

### Decided (2026-09-17)

Answered by the owner on 2026-09-17. The options are kept, the chosen one
is marked **(chosen)**, and the recommendation text is kept for the
record.

1. **How does a player declare a block that needs two or more creatures**
   (menace, Pathrazer)?
   - (a) **(chosen)** **Stage it on the board.** Clicking a blocker and then a menace
     attacker draws a dashed arrow and "needs 1 more blocker" on the
     attention strip. When the minimum is reached, the client sends one
     `declare_blockers`. Clicking the staged blocker again removes it,
     and leaving the step discards the staging.
   - (b) **Context menu only.** "Block with…" on the attacker opens a
     checklist of the defender's creatures whose `blockable_attackers`
     include it, and confirms at the minimum. The two-click flow keeps refusing a lone menace
     block with the `illegal_block` reason.
   - (c) **Both.**

   **Recommendation: (a).** It keeps the two-click flow players already
   use, and it answers ADR 0014 §9's objection directly: nothing is sent
   until the block is complete, so no error appears mid-choice. (b) is a
   fallback that PR 5 can add cheaply later.

   **Decision: (a).** Multi-blocker blocks (menace) are staged on the
   board: dashed arrows, "needs N more blocker(s)" in the combat-hint
   banner, and one `declare_blockers` when the minimum is met. Applied in
   Decision 16 and PR 5.

2. **What does the board show before a block is refused?**
   - (a) **Nothing new.** A refused click shows the `illegal_block`
     banner with its reason.
   - (b) **(chosen)** **Dim what can't be blocked.** In block mode, once a blocker is
     selected, attackers not in its `blockable_attackers` are dimmed. A
     click on a dimmed attacker still shows the reason banner. Menace
     attackers show "needs 2 blockers" on hover.
   - (c) (b), plus a small reason label on each dimmed attacker, which
     needs a per-pair reason stamp on the wire.

   **Recommendation: (b).** Landwalk makes "why can't I block that?"
   common in a lords deck. Dimming comes from the stamp at no extra wire
   cost, and the banner still explains any click. (c) multiplies the view
   by blockers × attackers for text the banner already gives.

   **Decision: (b).** Attackers the selected blocker can't block are
   dimmed, and the banner explains clicks. Applied in Decision 16.

3. **What is in the first card wave?**
   - (a) **#705 exactly.** Lord of Atlantis, Elvish Champion and Goblin
     King, then #750's pair and count cards as listed above.
   - (b) **(chosen)** (a) plus Master of the Pearl Trident, Cold-Eyed Selkie and
     Trailblazer's Boots, which this seam alone unblocks.
   - (c) (b) plus **fear, intimidate, shadow, horsemanship and skulk** as
     canonical tokens (CR 702.36b, 702.13b, 702.28b, 702.31b, 702.118b).
     Each is a two-card check in slot 2. Imported cards that print
     them start being enforced, which closes the whole "Landwalk, fear,
     intimidate …" seam row. Those tokens have no tracker today.

   **Recommendation: (b), with (c)'s keywords as a follow-up issue.** (b)
   costs three card files on top of the engine work. (c) turns on five
   evasion rules for every imported card that prints them in one PR.
   That is correct, but it is a large change in behaviour on real decks,
   and it deserves its own PR and soak run rather than riding along with
   landwalk.

   **Decision: (b).** Fear, intimidate, shadow, horsemanship and skulk
   go to a separate follow-up issue,
   [#825](https://github.com/krakenhavoc/cmd_and_ctrl/issues/825)
   (filed 2026-09-17, unscheduled). Applied in Decision 9, the PR split
   and the card first wave.


### Evasion follow-up implementation — #825 (2026-09-17)

Fear, intimidate, shadow, horsemanship and skulk now occupy slot 2 after
flying and landwalk. Each has a stable refusal token and a server-built
explanation; restrictions combine, so a blocker must satisfy every applicable
rule. The shared check reads effective characteristics and remains read-only.

The importer and keyword-only coverage scan use the canonical table, while
the legal move list and block-decision signal use the shared pair check.
The bot's attack estimate handles the five keywords from public characteristics.
This follow-up includes one real catalog consumer per keyword; #750's
parameterized rules and block-count work remain separate.

Measured against the local Scryfall dump on 2026-09-17, the canonical-table
change also makes 28 Commander-legal, single-faced keyword-only creatures
fully automatic by import: six with fear, five with intimidate, three with
shadow, ten with horsemanship and four with skulk. This is a dated measurement,
not a count of new catalog entries; Bladetusk Boar and Furtive Homunculus are
both keyword-only creatures and explicit catalog representatives.

The public card view now carries effective colors so the bot's fear and
intimidate estimates can distinguish a color-changing effect from the printed
mana cost. Empty colors means colorless, and unknown-card redaction removes
the field with the other identifying characteristics.

---

## Amendment (2026-09-17): the block declaration is announced at its lock-in (#830)

Amends the addendum's [Decision 13](#13-the-stored-declaration-is-always-legal-and-the-menace-close-out-is-deleted),
which said "events come after the commit … a re-point emits no second
`EventBlock`" and left the other half of the re-point to its own issue.
That issue is [#830](https://github.com/krakenhavoc/cmd_and_ctrl/issues/830),
and this is its answer. Decision 13's own PR had not landed when this
amendment was written (it since has, in #750's engine PR); this
amendment is written against the per-pair `DeclareBlocker` that is
still the only block verb, and Decision 13's bulk `DeclareBlockers`
inherits it unchanged — the lock-in is where the events are emitted,
whichever verb stages the pairings.

### The bug

The defender could re-point a blocker from attacker A to attacker B
before the declaration was complete (the client's "Re-declare
blocker", `contextMenu.logic.ts:794`). The first click had already
emitted `EventBlock` naming A, so A's "becomes blocked" trigger was
already harvested, and nothing withdrew it. A ended the declaration
**unblocked** and its trigger still resolved: Cyberman Patrol's
afflict 3 fired for an attacker nobody blocked, and Grazilaxx offered
to bounce an unblocked creature.

The per-attacker readers deduplicated by walking the event log back
to the attacker's own `EventAttack` (`b18AttackerAlreadyBlocked`), so
a withdrawn block also suppressed a later real one: block A with X,
re-point X to B, then block A with Y, and Y's block was declined as
"not the first".

### Decision 19. Nothing is announced until the declaration is locked in

CR 509.1 declares blockers as ONE turn-based action and CR 509.2a puts
its triggers on the stack when the declaration is complete, before
anyone receives priority. `DeclareBlocker` therefore **stages** the
pairing on `Card.BlockingTarget` and emits nothing.
`commitBlockDeclarationLocked` (`blockers.go`) is the single place the
declaration is announced, and the single place block-declaration
triggers are harvested from — through the ordinary event harvester,
which is still keyed by event kind. There is no second harvester and
no per-card special case.

The lock-in runs at the first point play moves on inside the step:

- `runStateChecksLocked`, which is the engine's "a player would
  receive priority" boundary (a trick cast during the step, a
  resolution), and
- the two places the cursor can leave the step — `AdvanceStep` and
  `PassPriority`'s wrap — both of which call `runStateChecksLocked`
  themselves when a declaration is still staged, so the lock-in always
  happens INSIDE `declare_blockers` and never a step late.

`PassPriority`'s wrap locks in **before** choosing between advancing
and resolving: if the declaration put anything on the stack or queued
a trigger prompt, priority returns to the active player (CR 509.2a)
and the step does not advance on that pass. `AdvanceStep` locks in
above the #730 prompt gate, so an optional block trigger's yes/no
holds the cursor rather than being walked past.

**Alternative rejected: withdraw on re-point** (#830's option (b) —
delete A's not-yet-resolved instances from `PendingTriggers` and the
open prompts when a blocker leaves it). It keeps the per-click
announcement and then unpicks it, which means every trigger sink
(pending queue, prompts, the stack, the log) needs a retraction path,
and a trigger that had already resolved cannot be retracted at all. A
declaration that is never announced until it is final has nothing to
retract.

### Decision 20. Two events: one per blocker, one per blocked attacker

The old single `EventBlock` had to serve both readings, which is why
every "becomes blocked" card carried a log-walking dedupe. The lock-in
emits both, in one event batch:

- **`EventBlock`**, one per (blocker, attacker) pair of the final
  declaration. "Whenever this creature blocks" (CR 509.3a), "becomes
  blocked by a creature", and the public game log's block entry.
  Actor is the blocker's controller, CardID the blocker, Target the
  attacker — unchanged fields, so Savvy Hunter and Brimaz read it
  exactly as before.
- **`EventBecomesBlocked`**, one per blocked attacker, however many
  creatures block it (CR 506.4 — an attacking creature is blocked
  once, at the moment the declaration is complete). Actor is the
  defending player; Source, CardID and Target are all the attacker,
  the way `EventBattleDefeated` names the battle three ways. Cyberman
  Patrol (afflict, CR 702.130) and Grazilaxx watch this and need no
  dedupe; `b18AttackerAlreadyBlocked` is deleted.

Not logged: `EventBecomesBlocked` has no `protocol.LogEvent`
projection, because the per-pair `EventBlock` entries already say who
blocked what and a second line per attacker would only repeat them.

**One batch.** Nothing in the lock-in opens an event batch, so the
whole declaration carries one `Event.Batch` and a `OncePerBatch`
ability collapses it — the same property `DeclareAttacker` relies on
for Adeline (#854, ADR 0049's batch amendment).

### Decision 21. The announcements are combat-scoped state

`Game.announcedBlocks` (blocker → the attacker its `EventBlock` named)
and `Game.announcedBecameBlocked` (attackers that have had their one
`EventBecomesBlocked`) are what makes the commit idempotent and
incremental: the sandbox lets the defender keep clicking after the
lock-in, and a second blocker added to an already-blocked attacker
announces its own block and no second "becomes blocked". Both are
cleared by `clearCombatLocked`, alongside the `BlockingTarget` wipe
they describe, and both are carried by `Clone` / `RestoreFrom` and the
persisted snapshot — an undo that rewound the declaration but kept the
announcements would swallow the re-done trigger, and the reverse would
double-fire it.

### What this fixes, and what it leaves

Fixed: the attacker a blocker left never becomes blocked; a blocker
re-pointed back to where it started is one declaration; a double block
is one "becomes blocked" and one "blocks" per blocker; a later real
block of an attacker is no longer suppressed by a withdrawn one.

Narrowed but not closed, #388: the lock-in drains block triggers onto
the stack inside `declare_blockers`, so under ordinary priority play
afflict resolves before combat damage — which it did not before. A
seat that clicks `advance_step` straight out of the step still walks
past the trigger on the stack and takes the damage first. Cyberman
Patrol's caveat now says exactly that. **(Closed 2026-09-18, #914:
`advance_step` passes priority until the step ends, so the skip-ahead
route resolves the afflict inside `declare_blockers` too. The caveat
is gone and the card is `CompletenessFull`.)**

Unchanged: the attack side. `DeclareAttacker` announces per creature
as it is declared and drains immediately, which is what Adeline and
the rest of the attack triggers depend on; a re-pointed attacker's
`EventAttack` keeps naming the defender it was first declared against.
That is a stale FIELD on a trigger that is legitimately owed, not a
trigger that should not exist, so it is a different fix and is not
made here.

---

## Amendment (2026-09-17): the attack declaration is announced at its lock-in too (#859)

Amends the 2026-09-17 block-declaration amendment's closing paragraph,
which said the attack side was "unchanged" and left the stale defender
to its own issue. That issue is
[#859](https://github.com/krakenhavoc/cmd_and_ctrl/issues/859), and
this is its answer. Decisions 19-21 stand exactly as written; this is
their attack-side twin.

### The bug

The active player could re-point an attacker from one defender to
another before the declaration was complete (the client's "Re-declare
attacker", `contextMenu.logic.ts:727`). `DeclareAttacker` emitted
`EventAttack` on the FIRST click and skipped re-emission on the
re-point, so the one event the creature is owed kept naming the
defender it had already left.

The trigger itself was legitimately owed — the creature did attack
(CR 508.1) — but everything downstream read the wrong player:

- **Bodies**, which put the effect somewhere: Brimaz's Cat Soldier
  entered attacking the wrong player, Hellrider pinged the wrong seat,
  Parhelion II's Angels and General Kreat's Goblin landed on the
  wrong defender.
- **Conditions**, which decide whether the trigger exists at all:
  Curse of Opulence paid its controller for an attack that left the
  enchanted player and paid nothing for one that arrived on them;
  Kazuul and Revenge of Ravens fired for a player nobody ended up
  attacking; Guild Artisan's and Horizon Explorer's "attacked a
  player" read the first classification, so a planeswalker → player
  re-point (and the reverse) was judged on the wrong one.

The second group is why a late, live read of the defender at trigger
RESOLUTION — the shape #859 sketched — was not enough: by resolution
the harvest has already happened, and a condition evaluated against
the stale defender has produced the wrong SET of triggers, not merely
a wrong field on the right ones.

### Decision 22. `DeclareAttacker` stages; the lock-in announces

CR 508.1 declares attackers as ONE turn-based action and CR 508.2
gives the active player priority afterwards, which is where the
triggers it produced go on the stack. `DeclareAttacker` therefore
**stages** the attack on `Card.AttackingTarget` (and still taps it,
CR 508.1f) and emits nothing. `commitAttackDeclarationLocked`
(`server/internal/game/attackers.go`) is the single place an attack
declaration is announced, and therefore the single place attack
triggers are harvested from — through the ordinary kind-keyed
harvester, with no second harvester, no per-card special case, and no
change to the card constructors (`WheneverThisAttacks`,
`ThisAttacked`, `attackDeclared`, …).

It runs at the same three points Decision 19 named for blocks, and
for the same reason:

- `runStateChecksLocked`, the engine's "a player would receive
  priority" boundary, where it sits immediately after
  `commitBlockDeclarationLocked`; and
- `AdvanceStep` and `PassPriority`'s wrap, which both call it when a
  declaration is still staged, so the lock-in always happens INSIDE
  `declare_attackers` and never a step late — still ahead of blockers,
  which is the S22 contract that mattered.

The bulk `DeclareAttackers` (#318) is now a pure staging loop followed
by the same single `runStateChecksLocked`. Its observable behaviour is
unchanged — one event per declared creature, one batch, one drain —
but there is now exactly ONE place an attack declaration is announced,
whichever verb staged it.

**What the S22 contract actually promised.** "One `EventAttack` per
declared attacker", not "one per click at click time". The lock-in
keeps the first and drops the second, so Adeline still makes one batch
of Humans for a three-creature attack
(`TestAdelineStillMakesOneBatchOfHumansPerAttack`), and nothing that
reads the event needed changing. Nothing on the client reads the
per-click event either: attack arrows come from `attacking_target` on
the card view (`CombatArrows.svelte`), the attack sound is played
locally on the click (`Game.svelte:650`, `:681`), and
`combatBeats.ts` reads the damage entries. **No client change.**

**Alternative rejected: resolve the defender late** (#859's own
suggestion — keep the click-time event and have one helper resolve the
defender from the attacker's current `AttackingTarget` at trigger
resolution, CR 508.4-style). It fixes the bodies and not the
conditions, as above; it would have needed a second, defender-shaped
read in every `AppliesTo` that gates on the defender, which is the
per-card special-casing this engine avoids; and it leaves the public
game log naming a defender the attacker never attacked unless the
already-emitted record is rewritten in place.

### Decision 23. `announcedAttacks` is presence, not a pairing

`Game.announcedAttacks` (the creatures that have had their one
`EventAttack` this combat) is keyed on PRESENCE rather than on the
defender the event named, because CR 508.1 declares a creature as an
attacker ONCE: an attacker that has been announced is never announced
again this combat, however often it is re-pointed afterwards. That is
the attack side's one real difference from `announcedBlocks`, which
compares the pairing because `EventBlock` is per (blocker, attacker)
pair.

The same map is how a permanent PUT onto the battlefield attacking
stays out of the declaration (CR 506.3c — Parhelion II's Angels,
Adeline's Humans, Legion Loyalty's myriad copies): the two token entry
points mark it as announced so the lock-in's battlefield scan does not
mistake its `AttackingTarget` for a staged declaration and hand it an
attack trigger it must not have.

`announcedAttacks` is cleared by `clearCombatLocked` alongside the
`AttackingTarget` wipe it describes, and is carried by `Clone` /
`RestoreFrom` and the persisted snapshot, classified `carried` in
`clone.go`, `snapshot.go` and `snapshot_drift_test.go` next to
`announcedBlocks` — an undo across a re-point that kept it would
swallow the re-done attack trigger, and one that dropped it would
double-fire it.

### What this fixes, and what it leaves

Fixed: a re-pointed attacker announces once, naming the defender it
ends on; a re-point back to where it started is still one declaration;
a planeswalker → player re-point (and the reverse) is classified on
the final target; the public game log's "attacks" line names the
player that was actually attacked.

Changed as a side effect, and closer to CR 508.2: the whole
declaration's triggers are now harvested together, so several
differing attack triggers from one controller reach the CR 603.3b
ordering prompt as one group instead of one at a time — which is what
the bulk `DeclareAttackers` already did. An attacker staged and then
un-declared by `ClearCombat` before any priority boundary now
announces nothing at all, which it should not have before either.

Unchanged: everything about blocks, and the client.

---

## Amendment (2026-09-18): two combat damage steps with a priority window between them (#717, #716)

Amends nothing above — Decisions 1-23 are about which declarations are
legal and when they are announced, and none of them moves. This adds
the step AFTER the declarations: CR 510.4's second combat damage step,
tracked on [#717](https://github.com/krakenhavoc/cmd_and_ctrl/issues/717),
and the participation rule that goes with it,
[#716](https://github.com/krakenhavoc/cmd_and_ctrl/issues/716). It is
the answer [ADR 0053](0053-combat-damage-beats.md) Decision 5 deferred
when #187 shipped presentation-only.

Sprint S37 (combat correctness), tracked on
[#880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/880).

### The rules

CR 506.1: "there are two combat damage steps" in a combat where a
creature has first strike or double strike. CR 510.4: if at least one
attacking or blocking creature has first strike or double strike **as
the combat damage step begins**, the phase gets an extra combat damage
step, the first one for those creatures, and the second for the
creatures that had NEITHER keyword then plus the ones that have double
strike now (CR 702.4b, 702.7b, 702.7c). CR 510.3 / 510.3a give the
active player priority after each step, with the damage triggers put on
the stack first.

The engine had one `combat_damage` step running two passes back to
back, so a "whenever this deals combat damage" trigger from first-strike
damage resolved after regular damage, and nobody could respond in
between.

### Decision 24. The first-strike damage step is a step, and a turn may not have it

`StepFirstStrikeDamage` (`first_strike_damage`) is a thirteenth entry in
`turnSequence`, between `declare_blockers` and `combat_damage`, in
`PhaseCombat`. `StepCombatDamage` keeps its name, its wire value and its
meaning: it is the regular damage step, the second one when the first
happened and the only one when it did not. Every combat has it.

**It is skipped by the mechanism the engine already had for a step it
does not need.** `finishStepEntryLocked` already handles a step
cancelled by a CR 614 replacement (Stasis, Necropotence) by advancing
past it and recursing into the next step's entry — a skipped step
(CR 500.11). `runStepEntryHooksLocked` now asks `stepExistsLocked`
first and makes exactly that move for a step this turn does not have:
nothing announces, no turn-based action runs, and no player gets
priority in it. The check is ahead of the replacement window on purpose
— a step that does not exist is not a step anything can replace.

`stepExistsLocked` answers false for one step today: `first_strike_damage`
when no attacking or blocking creature has first strike or double strike.
It is a live board read, not a flag set at declare-blockers, because
CR 510.4 asks the question as the damage step begins and a creature can
gain or lose first strike in the declare-blockers priority window.

**No client special case.** The client's step list gains one entry and
its phase strip one icon; nothing on the client decides whether the step
happens.

**Rejected: a `Sub` tag on a spliced second `combat_damage`**, which is
what [ADR 0059](0059-turn-machinery.md) Decision 12 sketched. That design
belongs to ADR 0059's `TurnPlan` — a planned list of steps with
insertion — and none of that machinery exists yet; building it here would
have made #717 the vehicle for extra turns and extra phases. A step that
a turn has or does not have is the same idea one layer down, and it is
the shape the plan can absorb: when the plan lands, `first_strike_damage`
is a plan entry that is present or absent, and `stepExistsLocked` is the
predicate that decides. ADR 0059 Decision 12 is superseded on the SHAPE
and stands on everything else (both steps count for `combat_damage`'s
ordinal; no card reads it; the client's stops treat the pair as one).

### Decision 25. One participation record, taken as the first step begins

`Game.firstStrikeStepParticipants` (unexported, `map[uuid.UUID]bool`) is
the combatants that had first strike or double strike as the FIRST
combat damage step began. `firstStrikeStepParticipantSetLocked` is the
one scan that produces it, and it answers both questions CR 510.4 asks
at that instant: whether the step exists (the set is non-empty) and who
deals damage in it.

`participatesInStepLocked` reads only that record:

- first step — the creatures in the record;
- second step — the creatures NOT in the record, plus any that have
  double strike now.

That second line is #716. The old `participatesInSubstep` re-read
`HasKeyword` after the layer recompute between the passes, so a creature
whose granted first strike died with its lord looked like a
non-first-striker and dealt its damage a second time; one that gained
first strike in between was skipped by a step it had never dealt damage
in. With a real priority window between the steps those are no longer
edge cases — the window is exactly where a lord dies and a pump lands.
The live keyword read that remains is the one CR 702.7c asks for:
"plus the ones that have double strike now".

**It is combat-scoped state**, cleared by `clearCombatLocked` with the
declarations it describes, and carried by `Clone` / `RestoreFrom` and the
persisted snapshot — classified `carried` in `clone.go`, `snapshot.go`
and `snapshot_drift_test.go` next to `announcedBlocks` and
`announcedAttacks`. It is carried for one more reason than they are: the
window between the steps is a priority window, so an undo or a deploy
restore can land inside it, and a record dropped there would let every
first-striker hit twice. An older snapshot restores with an empty
record, which reads as "there was no first-strike step" — right for
every file written before the steps existed.

### What this fixes along the way

**[#702](https://github.com/krakenhavoc/cmd_and_ctrl/issues/702) — the
regular pass no longer overtakes a first-strike assignment prompt.** A
CR 510.1c multi-blocker damage-assignment prompt blocks the table
(`choice_gate.go`), and the cursor cannot leave a step while a blocking
prompt is open (#730). With the two passes in two steps, that gate is
the fix: the regular step cannot begin until the first step's assignment
is answered. `TestCombatStepFirstStrikeAssignmentPromptOrder` loses its
skip. `DamageAssignmentFrame.FirstStrike` keeps its meaning as the
frame's record of which step queued it.

**The hand-rolled event batch between the passes is gone.** #784 opened
a batch by hand in `resolveCombatDamageLocked` because the rules split a
step the cursor did not. The cursor splits it now, so the ordinary "the
cursor enters a new step" boundary produces both batches and
[ADR 0049](0049-card-engine-seam-review.md)'s boundary rule is one
sentence with nothing named beside it. Professional Face-Breaker still
makes three Treasures for (first strike, A) / (regular, A) /
(regular, B).

### What it leaves

**[#914](https://github.com/krakenhavoc/cmd_and_ctrl/issues/914) is not
fixed here**, and the new step widens it by one case: `advance_step`
walks past a non-empty stack, so a seat that clicks it in the
first-strike damage step takes regular damage before the first-strike
damage triggers resolve — the same shape as the `declare_blockers`
caveat #914 already records. Ordinary priority play is correct, because
passing priority resolves the stack. The fix is a turn-structure
decision (#914 lists three options), and the version scoped to this
amendment's steps is not a one-liner: gating `advance_step` on "the
stack is not empty" in `declare_blockers` and the two damage steps was
tried on this branch and refuses the advance in **11 existing tests**
in `internal/cards/effects` — attack and upkeep triggers that are still
on the stack when the test walks the cursor into combat damage, plus
`TestB492AdvanceStepOutOfDeclareBlockersStillTakesDamageFirst`, which
pins today's behaviour on purpose. That is a change to what the verb
MEANS, and it needs the sentinel error, the `internal/legal` rule and
the client message #914 asks for. It stays #914's call.

> **Closed (2026-09-18, #914).** The call went the other way and cost
> no sentinel, no `internal/legal` rule and no client message:
> `advance_step` now means "pass priority until the step ends" (CR
> 117.4), so it RESOLVES what the step owes instead of refusing to
> move. The 11 tests the refusal broke are not broken by the drive —
> they reach the same board, because passing is what they would have
> done by hand. Three expectations did change, all in
> `internal/cards/effects`, and each to the rules answer rather than
> to silence: Drana, Liberator of Malakir (her first-strike trigger
> grows the attackers before regular damage: 2 + 3, not 2 + 2),
> Professional Face-Breaker (the (first strike, A) Treasure is already
> made when the cursor reaches the regular step — still three), and
> `TestB492…`, flipped to "afflict, then damage" and renamed. See
> [ADR 0018 §6](0018-triggers-on-the-stack.md)'s 2026-09-18 amendment.

**[#715](https://github.com/krakenhavoc/cmd_and_ctrl/issues/715) is
unchanged**: a blocked attacker whose blockers have all left combat is
still treated as unblocked in the second step. Its test keeps its skip.
*(Fixed on 2026-09-18 by Decision 26 below, and the skip came off with
it.)*

**The beat animation is not rebuilt.** [ADR 0053](0053-combat-damage-beats.md) Decision 5 now carries a note saying what #717 changed for the client beat sequencer: it keys a combat's beats on the first damage step entry, keeps the arrow cache through the new step, and keeps everything else.

---

## Amendment (2026-09-18): an attacker stays blocked when its blockers leave (#715)

Amends the addendum's [Decision 13](#13-the-stored-declaration-is-always-legal-and-the-menace-close-out-is-deleted)
on WHERE the menace close-out lives, and adds the state the combat
damage steps had been doing without. Decisions 1-25 stand.

Sprint S37 (combat correctness), tracked on
[#880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/880).

### The rules

CR 509.1h: "a creature remains blocked even if all the creatures
blocking it are removed from combat." CR 506.4 removes a creature from
combat when it leaves the battlefield, phases out, or an effect says
so. CR 510.1c: a blocked creature with no creatures blocking it
assigns no combat damage. CR 702.19d/e: trample is the exception — a
blocked trampler with nothing blocking it assigns all its damage to
the player or planeswalker it is attacking. CR 509.1b makes block
legality a property of the declaration, checked as it is made.

The engine kept no blocked state. Both combat damage passes rebuilt
`blockersByAttacker` from the live battlefield, so "no live blocker"
and "not blocked" were the same question, and the menace close-out ran
on that same live map at damage time. Three wrong boards, all the same
bug: a chump blocker that died in the first-strike step let the
attacker hit the player in the second; a blocker destroyed in the
declare-blockers window let a blocked attacker through; and a menace
attacker that lost one of its two blockers had its whole block
reverted and hit the player for full.

### Decision 26. One blocked record, written at the lock-in and read at damage

`Game.blockedAttackers` (unexported, `map[uuid.UUID]bool`) is the set
of attackers that are BLOCKED this combat. It is the field #830's
`announcedBecameBlocked` already was, under the name the rules give
it: an attacker becomes blocked exactly once (CR 506.4) and stays
blocked (CR 509.1h), so the set that has had its `EventBecomesBlocked`
and the set that is blocked are the same set, and keeping two would be
two things to get out of step.

**Written in one place.** `commitBlockDeclarationLocked` (`blockers.go`),
the block declaration's lock-in — Decision 19's "the declaration is
complete" point. Nothing else writes it. `clearCombatLocked` drops it
with the `BlockingTarget` wipe it describes, `removeFromCombatLocked`
(#921) drops one attacker's row when an effect takes that attacker out
of combat, and the battlefield exit (#935) drops it when the attacker
leaves the battlefield. A row keyed by the ATTACKER is never touched by
anything that happens to a BLOCKER, which is CR 509.1h expressed as a
map key.

**Read in one place.** `attackerBlockedLocked`, called by
`assignAndDealCombatDamageLocked` for an attacker with no live
blockers: blocked and no blockers left → no damage at all (CR 510.1c);
blocked, no blockers left and trample → all of it to what it is
attacking (CR 702.19d/e); never blocked → the unblocked path it always
took. Both damage steps run that one function, so first strike and
double strike need nothing of their own — which is what takes the skip
off `TestCombatStepDoubleStrikeBlockerDiesInFirstStep`.

**No "becomes unblocked" event.** There is no such rules concept: the
attacker's state never changes, so there is nothing to announce.

**The menace close-out moves with it**, from the top of
`assignAndDealCombatDamageLocked` to `revertIllegalBlockCountsLocked`,
called by the lock-in before it announces anything. Same rule
(`BlockerCountValid`, CR 702.111b), asked at the moment CR 509.1b asks
it: the declaration is complete, so a COUNT can be judged. An attacker
already in `blockedAttackers` is skipped — its declaration was legal
when it was made, and legality is never re-evaluated. Three things
fall out. An illegal lone block against a menace attacker is reverted
INSIDE `declare_blockers`, where the defender can see it and block
again, instead of silently at damage. The reverted blocker's
`EventBlock` is never emitted, so "whenever this creature blocks" no
longer fires for a block the engine is about to undo. And the damage
steps no longer read `HasKeyword(atk, "menace")` at all. This is the
half of Decision 13 that does not need the bulk `DeclareBlockers`
verb: the close-out is not deleted yet, but it is no longer at damage
time, and when the set-shaped declaration lands it is
`revertIllegalBlockCountsLocked` that it replaces.

**It is combat-scoped state**, classified `carried` in `clone.go`,
`snapshot.go` and `snapshot_drift_test.go` beside `announcedBlocks`,
`announcedAttacks` and `firstStrikeStepParticipants`. An undo that
rewound the removal but dropped the blocked state, or a deploy restore
that landed between the damage steps without it, would hand the
defending player damage they had blocked. The persisted key keeps its
#830 spelling (`announcedBecameBlocked`) — identical contents and
identical lifetime, so a file written before the rename restores a
mid-combat blocked state rather than losing it.

### What this fixes, and what it leaves

Fixed: the three boards above, on both damage steps, for a blocker
that died, was bounced, or was removed from combat by an effect; and a
blocked trampler whose blockers are gone now tramples over for its
full power (CR 702.19d/e).

Unchanged: the multi-blocker CR 510.1c assignment prompt. An attacker
with two or more live blockers still queues it, and a blocker that
dies while the prompt is open is the prompt's own validation problem,
not this record's.

Not attempted: "removed from combat" as a card-facing verb (#672), and
the CR 509.1c blocking REQUIREMENTS that Decision 15 sketches. Both
read this state when they land; neither writes it.

---

## Amendment (2026-09-18): combat lasts through the end of combat step (#785)

Adds the other end of the combat's lifetime. Decisions 1-26 stand;
this one moves a single call site.

Sprint S37 (combat correctness), tracked on
[#880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/880).

### The rules

CR 511.1: the end of combat step has no turn-based action, and the
active player receives priority in it. CR 511.2: "at end of combat"
abilities trigger as the step BEGINS. CR 511.3: "as soon as the end of
combat step ends, all creatures, battles, and planeswalkers are
removed from combat."

`runStepEntryHooksLocked` called `clearCombatLocked` on ENTRY to
`end_combat`, so every attacker and blocker stopped being in combat
for the whole of the step the rules keep them in. The comment there
showed the clear had already been deferred once — from `combat_damage`,
to keep the client's arrows up through the damage step — and it had
stopped one step short of where the rules put it.

The cost was that **no creature was attacking during the end of combat
step at all**. Aetherize, Settle the Wreckage and Aetherspouts cast
there found nothing and did nothing. Desert ("{T}: this land deals 1
damage to target attacking creature. Activate only during the end of
combat step") had no legal target in the only step it can be
activated, which is why it was left out of #743. Goro-Goro, Disciple
of Ryusei's "activate only if you control an attacking modified
creature" was false there, against the printed card, and carried a
declared caveat saying so.

### Decision 27. The clear happens as the cursor LEAVES end_combat

`advanceCursorLocked` — the one seam every step transition goes
through — clears combat when the step it is leaving is
`StepEndCombat`, and `runStepEntryHooksLocked` loses its
`StepEndCombat` case entirely (the step has no turn-based action,
CR 511.1). One call site moves; nothing else about the clear changes.

**What still happens at the step's entry** is the step announcement,
so "at end of combat" triggers are harvested from `EventStepBegan`
with the creatures still in combat (CR 511.2 fires them as the step
begins, and CR 511.3 removes the creatures after). Legion Loyalty's
delayed trigger, scheduled `At: StepEndCombat`, is in the same
position and now exiles its myriad tokens while they are still
attacking — which is what a token exiled at end of combat is.

**The verbs keep their own clears.** `ClearCombat` (the sandbox verb),
`PassTurn` and the eliminated-seat rotation end a turn without the
cursor ever leaving `end_combat`, so each still calls
`clearCombatLocked` itself. #672's "remove from combat" verb and
#921's `removeFromCombatLocked` are per-permanent and untouched.

**No client change.** The combat arrows are drawn from the server's
`attacking_target` / `blocking_target` with no step gate
(`CombatArrows.svelte`), so they now come down when the cursor reaches
`postcombat_main` — which is the rules answer and the same behaviour
the deferral was protecting. [ADR 0053](0053-combat-damage-beats.md)'s
`keepArrowCache` is untouched: it governs the geometry cache of tiles
that have LEFT the battlefield, kept while a beat cue is still
playing, which is a different question from whether a live card's
arrow is drawn.

### What this fixes, and what it leaves

Fixed: every "attacking creature" reader in the catalog works in the
end of combat step — Aetherize (tested), Settle the Wreckage,
Aetherspouts, `b13AttackingCreaturesYouControl`, `b26AttackingCreatures`,
`AttackingCreature()`.

Two cards lose a caveat and become `full`, neither needing a line of
card code changed — which is the Darksteel Citadel posture paying off.
**Goro-Goro, Disciple of Ryusei**'s Dragon ability is activatable in
the step, as printed. **Desert** (#450) is the worked example: "{T}:
this land deals 1 damage to target attacking creature. Activate only
during the end of combat step" landed in roadmap batch 43 declared as
printed and unusable, because the one window the ability allows had no
attacking creature in it. It has one now, and batch 43's
`TestB43DesertPingsAnAttackerInTheEndOfCombatStep` is the assertion
that batch said was waiting for this fix.

Unchanged: the two combat damage steps, the blocked state (Decision 26)
and the declarations' announcements. They are all cleared by the same
`clearCombatLocked`, so they all now last exactly as long as the
combat does.

---

## Amendment (2026-09-23, [#1227](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1227)): an entry that puts a CARD onto the battlefield attacking, and an exported "unblocked attacker"

Ninjutsu (CR 702.49) is a hand activation, and [ADR 0020](0020-activated-abilities.md)'s
2026-09-23 amendment carries that half. The two halves that are COMBAT's are
here, because both are facts this ADR already owns — `announcedAttacks`
(Decision 25) and `blockedAttackers` (Decision 26).

### Context

Decision 25 named `announcedAttacks` as the record that keeps a permanent PUT
onto the battlefield attacking out of the declaration (CR 506.3c), and named
the two token entry points as the only things that could mint one. That was
true: `ZoneEntryOptions` carried `Controller`, `Tapped` and `FaceDown` and had
nowhere to say "attacking", so a CARD moved from a real zone took
`executeEntryToBattlefieldLocked`'s `default:` branch and arrived with
`AttackingTarget == uuid.Nil`. The CR 506.3c marking lived in the minted-token
branch alone.

Decision 26's `blockedAttackers` had the mirror-image gap: one reader
(`attackerBlockedLocked`), unexported, consulted by the combat damage steps and
by nothing in `internal/cards/effects`. A cost clause could not ask whether an
attacker was blocked, so "Return an unblocked attacker you control to hand"
could not be written at all.

### Decision 29: the attack rides the ENTRY EVENT, and both doors read it

`ZoneEntryOptions.Attacking uuid.UUID` is seeded onto
`ReplacementEvent.EntersAttacking` by `putOntoBattlefieldFromZoneLocked`,
exactly as `Tapped` is seeded onto `EntersTapped` and `FaceDown` onto
`FaceDown`. The reason is the one those two already give: the entry's ONE
settled record carries the fact, a replacement effect inspecting the entry can
see it, and a CR 616 resume that holds only the event still knows what the
entry was for.

The token door now carries the same field. `enterCreatedTokensLocked` copies the
minted token's own `AttackingTarget` onto the entry event it builds, so
`executeEntryToBattlefieldLocked` reads `ev.EntersAttacking` for a token and a
card alike and the minted-token branch has no combat code left in it. The token
object keeps its own `AttackingTarget` — that is what a replacement inspecting
the staged token reads, and what the copy is taken from — but the RULE is now
stated once.

**`DeclareAttackerWith` is not the route, and this is the load-bearing half.**
It refuses outside `StepDeclareAttackers`, it taps, it checks summoning sickness
and it emits `EventAttack`. CR 506.3c is explicit that a permanent put onto the
battlefield attacking was never declared as an attacker, so all four would be
wrong — and the fourth is the one that would be invisible until an Adeline
trigger fired off a ninja.

### Decision 30: one marking function, three facts, and #1218's fourth exit

`stampEntryAttackerLocked(id, defender)` (attackers.go) is that one place. It
stamps `Card.AttackingTarget`, marks the permanent announced
(`noteAttackAnnouncedLocked` — CR 506.3c itself), and drops the layer cache.

The third is new and it was a real hole. #1218 gave
`StaticAbility.DependsOnAttackingStatus` three bumps: the `EventAttack` arm in
the layer listener, plus `invalidateLayersForAttackChangeLocked` at the two
exits that have no event of their own (a control change removing a permanent
from combat, and combat ending). An ENTRY is the fourth, and neither door had
it — a Parhelion II Angel became an attacking creature with no event naming it,
so an Ohran Frostfang reading "attacking creature" kept the stale answer until
something unrelated invalidated. Making the marking one function is what makes
that a one-line fix for both doors instead of two.

A `defender` that is no longer attackable (an eliminated seat, a planeswalker
that died while the ability was on the stack) leaves the permanent on the
battlefield NOT attacking rather than erroring. That is
`CreateTokensAttackingForEffect`'s existing posture — the effect has resolved
and the body is the part that can still be delivered — said once for both doors.

### Decision 31: `UnblockedAttackerForEffect` is CR 509.1h plus the step plus the staged block

The exported reader is not a getter on `blockedAttackers`. It answers the
question a cost clause actually asks, and it takes three facts to answer:

1. **Attacking** — `Card.AttackingTarget` is set. Removal from combat and the
   end of combat both clear it.
2. **The cursor is at declare blockers or later in this combat.** "Unblocked"
   is not a property an attacker has beforehand: CR 509.1h decides it as PART
   of the declaration, which is why the Gatherer ruling on Ninja of the Deep
   Hours says a ninjutsu ability can be activated only once blockers have been
   declared. Without this clause the blocked record is simply empty in the
   declare-attackers step and every attacker would read as unblocked — ninjutsu
   would become a "swing and always get the best creature back" ability.
3. **Nothing is blocking it**, which is two reads rather than one. The blocked
   RECORD (Decision 26 — the fact that outlives a blocker that has since died),
   AND a block that is staged and not yet announced. The second matters because
   `DeclareBlocker` only stages the pairing and the lock-in runs at the next
   priority boundary (Decision 27): in that window the record is still empty,
   and an attacker with a blocker already pointed at it must not read as
   unblocked.

**The sandbox simplification is inherited whole and is named rather than
papered over.** Nothing in this engine's step machinery marks the
declare-blockers turn-based action COMPLETE — `blockers.go`'s file header has
said so since #328 — so an attacker at the top of the step, before the
defending seat has clicked anything, reads as unblocked. That is a window in
which ninjutsu is available a beat early. It is never a window in which it is
available after the defender has blocked, which is the direction that would
matter. *Closed by #1279 (Decision 38): the declaration now has a completion
point per defending player, and the reader waits for it.*

### What this does NOT decide

- **A generic "put onto the battlefield blocking".** CR 509.1 has no such
  entry and no printed card asks for one; `TokenEntryOptions` has no blocking
  field either. Brimaz's Cat is created attacking on one branch and created
  plain on the other.
- **Commander ninjutsu** (CR 702.49c), whose entry may also come from the
  COMMAND ZONE. `putOntoBattlefieldFromZoneLocked` is generic in its source
  zone, so the entry itself would work; what is unexamined is the commander
  bookkeeping around a command-zone exit that is not a cast. See ADR 0020's
  amendment. *Shipped by #1278 (ADR 0020's 2026-09-24 amendment): the entry
  above is used unchanged, from the command zone.*

---

## Amendment (2026-09-23, [#1317](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1317), [#1329](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1329)): ending the combat phase, and reselecting an attack

Two effects that reach into a combat already under way, both found by the
*Aang is so flashy* deck triage (tracker
[#1306](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1306)). Decisions
1-31 stand. Sprint S37 (combat correctness), tracker
[#880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/880); the phase
skip is turn machinery too, tracker
[#884](https://github.com/krakenhavoc/cmd_and_ctrl/issues/884). CR numbers
checked against the pinned edition (`MagicCompRules 20260819.txt`).

### Decision 32: `EndCombatPhaseForEffect` is CR 724.2 in order, and 724.2c is folded

`server/internal/game/end_combat.go`. One printed card does this (CR 724.2
names it: Mandate of Peace), and the rule is a fixed procedure, so the verb
is the procedure:

1. **724.2g first:** outside the combat phase (`PhaseOf(step) != PhaseCombat`)
   it returns having done nothing.
2. **724.2a:** `PendingTriggers` is emptied. That queue *is* "triggered but
   not yet put onto the stack". A trigger that fires during the steps below
   lands on it afresh and is drained in the postcombat main phase, which is
   724.2f.
3. **724.2b:** every object on the stack is exiled, the resolving one
   included. Spells go through the shared stack-exit door,
   `exitSpellFromStackLocked(id, nil, ZoneExile, false)`, which retires the
   `StackMeta` record with the card (`DropStackMeta`). That is the door
   [#1318](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1318) is about:
   a plain exile route leaves the record behind and wedges the table. #1318's
   exported "exile target spell" had not merged when this landed, so this
   uses the same unexported door directly; the two are one door and do not
   conflict. The **resolving** spell has no `StackMeta` entry any more (the
   resolution frame took it before `OnResolve`), so it is found through the
   #920 resolving slot and routed with the same zone route; the frame then
   finds it gone and routes nothing (#489's `spellMovedItselfLocked`). A
   **copy** ceases to exist at once — 724.2b says an object "not represented
   by a card" ceases to exist at the next SBA check, and no zone this engine
   has can hold one. **Ability items** are deleted, exactly as CR 800.4a's
   stack cleanup deletes them: exiled, not countered, so no
   `EventCounterSpell` and nothing watching counters fires. A commander's exit
   can pause on the CR 903.9 prompt; it then sits on the stack under its
   prompt exactly as a paused counter does, and the prompt gates the table.
4. **724.2c is folded into the check that follows.** The verb only runs inside
   a resolution, and the resolution frame runs the SBA + trigger loop the
   moment it returns, with nobody having received priority in between. A
   check here would need an SBA pass that neither drains triggers nor moves a
   departed active player's turn on, and what it would buy is SBAs taken
   before the phase jump rather than after it — observable only as a
   trigger's slot in one APNAP drain that happens either way.
5. **724.2d:** `clearCombatLocked`, then the cursor walks forward through
   `advanceCursorLocked` — the seam every transition takes (Decision 27) —
   **without** running the entry hooks of the steps it passes, and the
   postcombat main phase is entered through `runStepEntryHooksLocked` like any
   other step. The engine has no "until end of combat" duration (a grep finds
   none), so that clause has nothing to expire today.
6. **724.2e falls out of 5:** the end of combat step is never entered, so it
   is never announced and nothing harvested off `EventStepBegan{end_combat}`
   fires. A CR 603.7 delayed trigger scheduled `At: StepEndCombat` is not
   drained either; it stays queued for the next end of combat step that
   begins.

It must be the **last** instruction of the effect that calls it, because it
exiles that effect's own object. On Mandate of Peace it is.

The resolution frame needed no change. `passPriorityLocked` resolves, runs
state checks and gives the active player priority — in whatever step the
cursor is now in, which is the postcombat main phase. `AdvanceStep`'s CR 117.4
drive sees the step change and stops (`driveStepEnded`).

### Decision 33: reselecting an attack writes `AttackingTarget` and nothing else

`server/internal/game/attack_reselect.go`. CR 508.7 has four clauses that
matter here; three of them are about what does **not** happen, and they are
the reason this is not `DeclareAttackerWith`:

- **508.7a** — not removed from combat, not "attacked a second time". No
  `EventAttack`, so no "whenever ~ attacks" trigger and no second Adeline
  batch; `announcedAttacks` (Decision 23) keeps the creature, so a later
  lock-in has nothing to announce; `blockedAttackers` (Decision 26) is
  untouched, so a blocked creature stays blocked by the same blockers. The
  second sentence — "still considered to have attacked the player … chosen as
  it was declared" — holds for free: the declaration's triggers were harvested
  at the lock-in, off the defender it was declared against.
- **508.7b** — no requirements or restrictions: no summoning-sickness,
  defender or tapped check, no tap, and no Propaganda tax (ADR 0080 charges
  the DECLARATION, which this is not).
- **508.7c** — the new target is checked with `canAttackTargetLocked`
  against the **attacking creature's** controller — the same function the
  declaration uses, because 508.7c and CR 506.2 name the same set. A
  defending player flashing in Misleading Signpost can push the attack onto
  any other opponent of the attacker, never back onto the attacker's side.
- **508.7d** does not apply: CR 903.2 makes Commander's default multiplayer
  setup Free-for-All with the attack multiple players option, which is the
  only setup this engine plays.

A creature that is only **staged** (declared by click, not yet locked in) is
refused with `ErrNotAttacking`: the declaration is still the active player's
to change, and a reselection landing before the lock-in would make the lock-in
announce the reselected defender, which 508.7a forbids. No effect can resolve
in that window, so the refusal only reaches a caller that is not one.

Everything downstream reads `AttackingTarget` live and follows with no change:
the new defender's creatures are the ones the block option generator
(Decision 14) offers, and an unblocked creature's damage goes to the new
target (CR 510.1b). The layer cache is dropped, as #1218 does at every
attack-status change with no event of its own.

**The prompt is the engine's**, `QueueReselectAttackForEffect`, and it is an
ordinary `option_pick`: "Keep attacking <current>" first — the "you may", and
the branch the enumerator marks always-legal — then every other legal target,
a seat as a seat option and a planeswalker or battle as a card option. Its
answer is read off the **option picked**, not an index into a captured list,
through a third continuation beside #994's `thenSeat`: `thenSubject` hands the
frame the option's seat, else its first card, else `uuid.Nil`. #994's hazard is
real here — a seat that concedes while the question is open is pruned off it
and renumbers the options after it — and an index frame would move the attack
onto the wrong player. The picked target is re-validated by the verb, so an
answer the board has since made illegal changes nothing. No wire change:
`option_pick` already carries `Player` and `Cards` per option.

The trigger condition ("when this enters during the declare attackers step")
and the per-creature chaining (Windshaper Planetar) are catalog-side, in
`effects/attack_reselect.go`.

### What this does NOT decide

- **"Your opponents can't cast spells this turn"**, Mandate of Peace's other
  sentence, is a cast restriction created by a resolving spell with a
  duration — [#1316](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1316),
  a separate PR. The card ships `caveats` until it lands.
- **Extra combat phases.** "The next phase, usually the postcombat main phase"
  is always the postcombat main phase here, because the turn has one combat
  (the "Extra combat and main phases" seam row).
- **A trigger already parked on a prompt** — a "you may" yes/no or a
  pick_target queued by the harvester during the same resolution, before the
  process began — is not withdrawn by 724.2a. It cannot exist in practice (an
  open prompt stops priority, so nothing resolves under one), and the only
  way to reach it is a trigger off Mandate of Peace's own resolution event.
- **Capricopian** ("only the player this creature is attacking may activate
  this ability") needs an activator that is not the controller, which
  `ActivatedAbility` has no shape for. Its reselect half would work unchanged.

## Amendment (2026-09-23, [#1339](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1339)): a creature blocks only what its controller is defending against

Found while testing #1329's reselect. `DeclareBlockers` took a pairing whose
blocker belonged to a player who was **not** defending against that attacker:
at a four-seat table, seat 1's 1/1 could block the 3/3 attacking seat 2, and
seat 2 took nothing. Decisions 1-33 stand. Sprint S37 (combat correctness),
tracker [#880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/880). CR
numbers checked against the pinned edition (`MagicCompRules 20260819.txt`).

The rule is CR 802.4a — "A defending player can block only with creatures
they control. Those creatures can block only creatures attacking that player,
a planeswalker that player controls, or a battle that player protects" — with
CR 509.1a saying the same thing for one defending player. Commander plays the
attack multiple players option (CR 903.2), so there are up to three defending
players in every combat this engine runs.

### Decision 34: the declaration checks the defending player, and the generator was already right

**The check.** `checkBlockDeclarationLocked` (Decision 13's validator) gains
one per-entry check, `blockDefenderRefusalLocked`
(`server/internal/game/block_declaration.go`), run before the CR 509.1b pair
check: the blocker's controller must equal
`defendingPlayerForAttackLocked(attacker.AttackingTarget)` — the player
attacked, the planeswalker's **controller**, or the battle's **protector**
(CR 310.9d). That is the same function `blockOptionsLocked` has always used
to decide which attackers a seat is offered (Decision 14), so the verb, the
enumerator and the #328 auto-pass signal now give one answer, which is
Decision 13's "the engine never holds a block it would refuse" restored in the
one direction it had been broken.

**Not in `BlockPairRefusalLocked`.** That function is the per-pair CR 509.1b
question about two creatures, and its contract already says "who controls the
blocker" is the declaration's business. Its other callers restrict the blocker
set to the defending seat before they ask it. Putting the defender check there
would make the enumerator ask it twice and change nothing.

**The refusal** is the reserved token `not_defending` (Decision 8's list),
claimed here: `BlockReasonNotDefending`, returned only by the declaration
verbs, never by a pair query — the same shape as the two count reasons. The
sentence says where the attacker **is** pointed, because that is what tells
the player whose creatures could block it: "Grizzly Bears is attacking P3, so
only P3 can block it", "…is attacking Jace, a planeswalker P4 controls, …",
"…is attacking Invasion of Ixalan, a battle P3 protects, …". `BlockRefusedError`
gains `TargetKind` and `TargetName` to carry the permanent.

**An attacker with no defending player cannot be blocked.** Two shapes reach
that: a creature not attacking at all, and one attacking a planeswalker or
battle that has left the battlefield (CR 506.4c). The generator never offered
either; the verb now refuses both with `not_defending` ("… isn't attacking
anything, so no one can block it."). The first ends the sandbox's
"pre-emptive block", which the old doc comment on `DeclareBlockers` promised
and one test (`TestClearCombatForgetsBlockAnnouncements`) leaned on — it now
re-points the attacker before its second block, which is what "next combat"
means. The second is **weaker than printed** for the defender: CR 506.4c says
such a creature "may be blocked", and in a multiplayer game the natural reading
(CR 802.2a's last-known defender) is the walker's former controller. The engine
keeps no record of that player once the walker is gone, so nobody may block it
— the answer the generator gave before this change too. Filed as
[#1364](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1364) rather than widened here.

**A standing pairing is not re-judged.** The check is skipped when the pairing
is already in the declaration's `base`. CR 508.7a and 509.1h: an attacker
reselected onto another player after it was blocked (#1329, Decision 33) stays
blocked by the same creatures, and a declaration that repeats that pairing —
alone, or inside a set beside a new pairing — must not be refused for a block
the rules say is still standing. A **new** pairing from the old defender is
refused; the new defender's is accepted.

**The wire and the client.** The view stamps `defending_player` on every
attacking card (`stampCombatTargets`, `server/internal/protocol/view.go`),
from the same function, and omits it when there is none. The client's block
pickers — the card menu's "Declare blocker" rows, the panel and seat-summary
click-to-block, and the canvas's block-mode gate — read it through
`defendingPlayerOf` / `attackersDefendedBy` (`client/src/lib/attackTargets.ts`),
so none of them offers a block the server refuses, and the click paths now
offer the attacks on a planeswalker the viewer controls or a battle they
protect, which they missed before. The client does not re-derive who defends a
battle (ADR 0045 §6's posture, applied to the defender): a frame that predates
the field falls back to a player attack's target and to nothing otherwise.

### What this does NOT decide

- **CR 506.4c's "it may be blocked"** for an attacker whose planeswalker or
  battle has left: see above. It needs the last-known defending player
  recorded when the permanent leaves ([#1364](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1364)).
  *Closed by the amendment below, Decision 35.*
- **CR 506.3e**, a creature put onto the battlefield blocking an attacker that
  is not attacking its controller: the engine has no "enters blocking" path
  yet, so there is nothing to gate.
- **CR 802.4's APNAP order of declarations** — each defending player declaring
  all of their blocks in turn — is still not modelled: every defending seat
  declares during the one declare-blockers window, and the lock-in (Decision
  19) announces the whole table's blocks at once. Legality does not depend on
  the order (CR 802.4b: a defending player's blocks are judged ignoring the
  attackers aimed at other players), which is why this check is enough.

## Amendment (2026-09-23, [#1364](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1364)): an attacker whose planeswalker or battle has left may still be blocked

Decision 34 left one shape weaker than printed on purpose: a creature attacking
a planeswalker or battle that has left combat could be blocked by nobody,
because the engine could not say who had been defending it. This closes it.
Decisions 1-34 stand. Sprint S37 (combat correctness), tracker
[#880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/880). CR numbers
checked against the pinned edition (`MagicCompRules 20260819.txt`).

### The rules

- **CR 506.4c** — "If a creature is attacking a planeswalker or battle,
  removing that planeswalker or battle from combat doesn't remove that
  creature from combat. It continues to be an attacking creature, although it
  is not attacking any player, planeswalker, or battle. It may be blocked. If
  it is unblocked, it will deal no combat damage."
- **CR 802.2a** — the defending player of such a creature is the player it was
  attacking, the controller of the planeswalker, or the protector of the battle
  it was attacking **before that permanent was removed from combat**.
- **CR 802.4a** (Decision 34) — only that player's creatures may block it.

### Decision 35: the defending player is recorded when the attack is pointed, and only the block path reads it

**The record.** `Game.attackDefenders`, attacker → defending player, written
by one helper, `setAttackTargetLocked` (`server/internal/game/attack_target.go`),
which every write that makes a creature attack something now goes through:
the two declaration verbs (`DeclareAttackerWith`, `DeclareAttackersWith`), the
CR 506.3c entry door (`stampEntryAttackerLocked` — ninjutsu, tokens created
attacking) and the CR 508.7 reselect (`ReselectAttackTargetForEffect`, #1343).
It stores `defendingPlayerForAttackLocked(target)` as it resolves at that
moment. A target with no defending player drops the row rather than keeping an
older one, so the record always describes the attack the creature is making.

**Why at pointing time, not when the permanent leaves.** CR 802.2a wants the
defender "before it was removed from combat", and CR 506.4 removes an attacked
planeswalker or battle from combat on every change that could alter that
answer — leaving the battlefield, phasing out, a control change, ceasing to be
a planeswalker or battle. So the defender at the last pointing IS the defender
at removal, and there is no exit hook to keep in sync with every door a
permanent can leave through (the battlefield exit, phasing, a type change that
is only a layer result). The one engine gap this leans on is noted below.

**The read.** `defendingPlayerForAttackerLocked(attacker)` returns the live
`defendingPlayerForAttackLocked` answer while the target resolves and the
recorded player once it does not — unless that player has left the game
(CR 800.4a). It replaces the live read at **every block-side caller and no
other**:

| Caller | File |
|---|---|
| the one option generator (Decision 14), and so the #328 signal | `block_declaration.go` `blockOptionsLocked` |
| the declaration's defender check (Decision 34) | `block_declaration.go` `blockDefenderRefusalLocked` |
| the refusal's `Defender` | `block_legality.go` `blockRefusedErrorLocked` |
| landwalk's "defending player controls a land" (CR 702.14c) | `landwalk.go` `landwalkBlockingLandLocked` |
| the enumerator's pre-filter | `legal/combat.go`, via `DefendingPlayerForAttackerForEffect` |
| the wire's `defending_player` | `protocol/view.go` `stampCombatTargets` |

Combat damage keeps its live read (`dealCombatDamageToAttackTargetLocked`), so
an unblocked creature whose target left still deals its damage to nothing
(CR 506.4c, 510.1b). The card-side "is this creature attacking you" readers
(`b17AttackersAllAvoid`, Horn of the Mark, the opponent-politics helpers) keep
theirs too: such a creature "is not attacking any player", and the fallback
would make it one.

**The refusal sentence.** `not_defending` for such an attacker says why the
player may be surprised: "Grizzly Bears is still attacking, though what it
attacked is gone, so only P3 can block it." (`TargetKind` is
`AttackTargetNone` with a `Defender` set — the new shape.)

**Lifetime.** Combat-scoped, like `announcedAttacks` (Decision 23): cleared by
`clearCombatLocked`, dropped per object by `removeFromCombatLocked`,
`forgetCombatRecordLocked` (phasing) and `forgetPerObjectTurnStateLocked` (the
battlefield exit), and carried by `Clone` / `RestoreFrom` and the persisted
snapshot (`attackDefenders`, omitempty — a file written before it restores
empty, which is the old behaviour until combat ends; no schema bump).

**The wire and the client.** No new field. `defending_player` is now present
on such an attacker, with `attacking_target_kind` absent; the client's block
pickers already read `defending_player` first (`defendingPlayerOf`), so they
offer the block with no code change beyond the comments and a test.

### What this does NOT decide

- **A control change of an attacked planeswalker** (or a protector change of
  a battle) mid-combat. CR 506.4 removes that permanent from combat, but the
  engine does not: the live read keeps resolving it, now to the new
  controller. That is a separate gap in removal-from-combat, not in this
  record, and the record would give the right answer the day the removal is
  modelled ([#1376](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1376)).
  *Closed by the amendment below, Decision 36* (the control-change half; the
  engine has no path that changes a battle's protector mid-game).
- **A planeswalker that phases back in** in the same combat still resolves
  live as the attack's target — the same removal-from-combat gap (#1376).
  *Closed by the amendment below, Decision 36.*
- **The bot's pressure estimate** (`aiseat/heuristic` `decideBlock`) counts
  only attacks aimed at its seat as incoming damage, which is right for such a
  creature (it deals none). The block moves themselves come from the
  enumerator, so a bot is offered and may take the block.

## Amendment (2026-09-24, [#1376](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1376)): an attacked planeswalker or battle that changes control or phases out leaves combat

Decision 35 recorded who was defending an attack so that the block path could
name that player once the attacked permanent had gone. It left the other half
of CR 506.4 open: an attacked planeswalker or battle that is removed from
combat **without leaving the battlefield** stayed in combat, because nothing
touched the creatures attacking it. This closes it. Decisions 1-35 stand.
Sprint S37 (combat correctness), tracker
[#880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/880). CR numbers
checked against the pinned edition (`MagicCompRules 20260819.txt`).

### The rules

- **CR 506.4** — a permanent is removed from combat if it leaves the
  battlefield, **if its controller changes, if it phases out**, if an effect
  specifically removes it from combat, if it is a planeswalker that stops being
  a planeswalker or a battle that stops being a battle, or if it is an
  attacking or blocking creature that regenerates or stops being a creature.
  A planeswalker or battle removed from combat stops being attacked.
- **CR 506.4c** (Decision 35) — the creatures attacking it keep attacking,
  attack nothing, may be blocked, and deal no combat damage if unblocked.
- **CR 802.2a** (Decision 35) — the player who may block them is the one
  defending before the removal.
- **CR 701.19a** — a regenerating permanent is removed from combat only "if
  it's an attacking or blocking creature".

### Decision 36: the attackers are re-pointed at a reserved id that names nothing

**The shape.** When an attacked planeswalker or battle is removed from combat
and stays on the battlefield, every ANNOUNCED creature attacking it has its
`Card.AttackingTarget` rewritten to `game.AttackingNothing`, the reserved id
`00000000-0000-0000-0000-000000000506`. One helper does it,
`removeAttackedFromCombatLocked` (`server/internal/game/attack_target.go`),
called from the two removal doors that keep the permanent on the battlefield:

| Door | Call site |
|---|---|
| a control change (layer 2 materialised, CR 613.1b) | `materialiseControlLocked` (`layers.go`), right after the stolen permanent's own `removeFromCombatLocked` |
| phasing out (CR 702.26b) | `phaseOutLocked` (`phasing.go`), after the batch has moved and before the events |

The attackers change in nothing else. They stay announced (they attacked),
stay blocked or unblocked (Decision 26), and keep their
`Game.attackDefenders` row — which is why the helper writes the field directly
rather than through `setAttackTargetLocked`, whose nil-defender arm would drop
the row Decision 35 exists to keep.

**Why a sentinel rather than a flag beside the walker's id.** A creature
attacking nothing has to satisfy two families of readers at once:

- every "is this creature attacking" test, `AttackingTarget != uuid.Nil` —
  the bulk of the field's ~140 non-test reads, across the engine, the enumerator, the view, the
  bots and the catalog — must still say yes (CR 506.4c: "it continues to be an
  attacking creature");
- every reader that RESOLVES the target — combat damage
  (`dealCombatDamageToAttackTargetLocked`), the card-side "attacking you"
  readers, the view's `attacking_target_kind`, the reselect label, the bot's
  per-defender tallies — must find nothing.

An id that names no seat and no card satisfies both with no edit at any
reader. It is also exactly the state a creature whose walker DIED has been in
since S27 (its target is an instance id no longer on the battlefield), which is
the state Decision 35's block fallback already handles — so blocking needed no
change at all: the live read resolves nothing, and
`defendingPlayerForAttackerLocked` falls back to the recorded defender. A flag
beside a still-live walker id would have had to be consulted by every
resolving reader, and one that forgot would deal the damage — the failure this
issue is about. The value is a version-0 uuid, so it cannot collide with a
seat or instance id (both `uuid.New`, version 4).

**Why not the walker's id left as it was.** It still resolves: under a control
change to the new controller (who is then "being attacked" and takes the
walker's damage), and after a phase-in to the walker again. Phasing out alone
already stopped the damage while the walker was out of the battlefield slice
(ADR 0084); the rewrite is what stops it resuming.

**Only announced attackers.** A creature merely staged in the
declare-attackers step (Decision 22) is not in combat yet, and its declaration
is still the active player's to change — the reselect verb refuses it for the
same reason (Decision 33). A staged attack on a walker that changes control
keeps naming the walker.

**Not inside `removeFromCombatLocked`.** Its other caller is regeneration,
and CR 701.19a removes a regenerating permanent from combat only if it is an
attacking or blocking creature. An attacked planeswalker that is also a
creature (an animated Gideon) and regenerates stays attacked, and still takes
the damage.

**Lifetime, undo and restore.** Nothing new to carry: the sentinel is an
ordinary value of a field that `Clone`, `RestoreFrom` and the persisted
snapshot already carry, and `clearCombatLocked` wipes it with every other
`AttackingTarget` when combat ends.

**The wire.** No new field. Such an attacker ships `attacking_target` set to
the reserved id (so every client check for "is attacking" still holds), no
`attacking_target_kind`, and `defending_player` naming the seat that was
defending before the removal. The client's combat arrows find no seat for the
id and draw nothing, as they already did for a walker that died.

### Tests

`server/internal/game/attacked_leaves_combat_test.go`: a walker stolen by the
attacking player (blocks by the former controller; bystander refused; no
damage), a walker stolen by a third player (no loyalty lost, no life lost), a
battle stolen mid-combat (the protector blocks; the old and new controllers
are refused; no defense lost), a walker that phases out and back in (blocks,
no damage), a staged attack left alone, a regenerating walker-creature that
stays attacked, and the clone / undo (`RestoreFrom`) / persisted-snapshot
round trip. `legal/block_defender_lki_test.go`
`TestBlockMovesOfferAnAttackerWhoseWalkerWasStolen` (the enumerator and
`actions.Dispatch` agree); `protocol/defender_lki_view_test.go`
`TestAttackerWhoseWalkerWasStolenAttacksNothingOnTheWire`.

### What this does NOT decide

- **A planeswalker or battle that stops being one and becomes one again
  inside one combat** (CR 506.4's type clause). While it is not a planeswalker
  the live read already resolves it to nothing; only the return is wrong,
  and no catalogued card produces it. It would need a type-delta hook in the
  layer pass. Filed as [#1387](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1387).
  *Closed by the amendment below, Decision 37.*
- **A battle's protector changing mid-combat.** The engine sets
  `ProtectorPlayerID` only when the battle enters (`battle.go`), so there is
  no path to hook.
- **Ninjutsu from an attacker that attacks nothing.** `stampEntryAttackerLocked`
  hands the ninja the returned creature's target; a target that resolves to
  nothing leaves the ninja on the battlefield not attacking, as it already did
  for a walker that died. Unchanged here.

## Amendment (2026-09-24, [#1387](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1387)): an attacked planeswalker or battle that stops being one leaves combat for good

Decision 36 closed the control-change and phasing clauses of CR 506.4 and left
the type clause open. This closes it. Decisions 1-36 stand. Sprint S37 (combat
correctness), tracker [#880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/880).
CR numbers checked against the pinned edition (`MagicCompRules 20260819.txt`).

### The rules

- **CR 506.4** — a permanent is removed from combat "if it is a planeswalker
  that stops being a planeswalker or a battle that stops being a battle".
- **CR 506.4c** (Decisions 35-36) — its attackers keep attacking, attack
  nothing, may be blocked by the player defending before the removal
  (CR 802.2a), and deal no combat damage if unblocked.

The removal is a one-way event. Nothing in CR 506 puts a permanent back into
combat when its type comes back, so a walker that is a planeswalker again later
in the same combat is not attacked.

### What was wrong

`classifyAttackTargetLocked` reads effective types, so while the permanent was
not a planeswalker or battle its attackers already resolved to nothing. The
wrong part was the return: when the type-changing effect ended inside the same
combat (an "until end of turn" effect cannot end before cleanup, but a "for as
long as" one can, CR 611.2b), the attackers still named the permanent, resolved
to it again, and dealt it their damage.

### Decision 37: the layer recompute is the type-change door

**The shape.** `recomputeLayersLocked` (`server/internal/game/layers.go`) calls
`removeTypeLostAttackTargetsLocked` (`server/internal/game/attack_target.go`)
once per pass. It walks the announced attackers, takes each distinct target
that resolves to a battlefield permanent, and hands any that the pass left as
neither a planeswalker nor a battle to `removeAttackedFromCombatLocked`, the
Decision 36 helper. The attackers are re-pointed at `game.AttackingNothing` at
the moment of loss, so a later pass that restores the type finds nothing
naming the permanent. The third row of Decision 36's door table:

| Door | Call site |
|---|---|
| the permanent stopped being a planeswalker or battle (layer 4, CR 613.1d) | `recomputeLayersLocked` (`layers.go`), after the layer pass and the `lastResolvedVersion` store |

**Why the recompute, and not a before/after type diff.** The layer pass is the
only place an effective type changes, and every type change bumps the layer
version, so the recompute is guaranteed to run between a loss and any reader
that could see it. What the check needs is the post-pass state and nothing
else: an attacked permanent that is not a planeswalker or battle NOW was one
when it was attacked (the declaration verbs refuse anything else), so it has
stopped being one. No "before" snapshot of the types is kept, which is one
fewer thing to carry through `Clone`.

**Why this is cheap.** Outside combat `announcedAttacks` is empty and the call
returns at once. Inside combat it looks only at what announced attackers name:
a seat, the sentinel, or an id that no longer resolves costs a lookup and
nothing else, and each attacked permanent is classified once however many
creatures attack it. No per-card work is added to the pass itself.

**After the store, unlike the control-change door.** The rewrite bumps the
layer version when an attacking-status static is live
(`invalidateLayersForAttackChangeLocked`). A bump made before
`lastResolvedVersion.Store` would be swallowed by it; made after, it asks for
one more pass, which rewrites nothing (the rewrite changes no type) and
settles. `materialiseControlLocked` still calls the helper before the store;
no attacking-status static reads the attack's target today (Ohran Frostfang
reads only "is attacking"), so that ordering is harmless there.

**What a permanent that merely GAINS a type does.** Nothing. An animated
walker (Planeswalker Creature) or an artifact battle is still a planeswalker or
battle and stays attacked.

**Only announced attackers**, as in Decision 36: a staged declaration at a
walker that loses its type is still the active player's to change.

**A pass that never sees the loss.** The check reads the state each pass
produces. A type removed and restored between two recomputes, with no reader
in between, is never observed as lost — but then no reader ever saw the
permanent as anything but a planeswalker either, so no player could have acted
on the loss. Any board read in between (a snapshot, a legality check, combat
damage) forces the pass that sees it.

**Lifetime, undo and restore.** Nothing new to carry, for the reasons in
Decision 36: the rewrite is an ordinary value of `AttackingTarget`. A clone
taken before the loss restores the original attack.

**The wire.** Unchanged from Decision 36.

### Tests

`server/internal/game/attacked_loses_type_test.go`: a walker that loses its
type and gets it back before damage (both attackers attack nothing; the
walker's controller blocks and a bystander is refused; no loyalty or life
lost), the same for a battle (its protector blocks, the controller is refused,
no defense lost), a walker that gains the creature type and a battle that
gains the artifact type (both stay attacked and take the damage), a staged
attack left alone, and the undo round trip (a clone from before the loss
restores the attack and the walker takes the damage; a clone, undo and
persisted restore from after it keep the attackers attacking nothing).

### What this does NOT decide

- **A card that produces this.** No catalogued card removes the planeswalker
  or battle type from a permanent, so the tests build the effect from a
  floating layer-4 static. No proof cards.
- **The control-change door's ordering** relative to the store, discussed
  above. Left as it is.

---

## Amendment (2026-09-24, [#1279](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1279)): the block declaration gets a completion point

Decisions 1-37 stand; Decision 31's declared simplification is closed by this
one. Sprint S37 (combat correctness), tracker
[#880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/880).

### Context

CR 509.1 makes declaring blockers one turn-based action, taken as the step
begins and before anyone receives priority. This engine has always kept the
sandbox's shape instead: entering `declare_blockers` hands the ACTIVE player
priority, and a defender's `declare_blocker(s)` is a verb accepted while that
window is open (#328). #830 moved the announcement to a lock-in at the first
priority boundary, and Decisions 26 and 31 built on it.

What none of that had was a moment at which a defender's declaration is
DONE. `Game.blockedAttackers` was empty at the top of the step for two
reasons that look the same — the defender decided not to block, or has not
been asked — so every reader that cared had to guess:

- **Ninjutsu** (Decision 31) read an attacker as unblocked the instant the step
  began, before the defender had clicked anything — "a beat early".
- **"Becomes blocked"** was announced at whichever priority boundary came first
  after the click, so a trick the active player cast in the step locked in a
  defender's half-made declaration.
- **"Whenever ~ attacks and isn't blocked"** (CR 509.3) had no moment to fire
  on at all, so every card printing it was left out.
- **The #328 auto-pass guard** (`block_decision_seats`) kept stopping a
  defender who had already declared, every time priority came back to them in
  the step, for as long as they had an untapped creature at home.
- **The bot's block grace** held the active bot's pass while any defender
  "still had a legal block", which a defender keeping a creature home always
  does, so it sat out the whole grace.
- **The client** had no way for a defender to say "done" without holding
  priority.

### Decision 38: a completion point per defending player

**State.** `Game.blocksDeclared map[uuid.UUID]bool` — the defending players
whose declaration is complete this combat (`server/internal/game/block_completion.go`).
Cleared by `clearBlockStateLocked` with the rest of combat; carried by
Clone / RestoreFrom and the persisted snapshot (`blocksDeclared`, additive, no
schema bump — a file written before it restores every defender as pending,
which is the conservative direction).

**Status.** `BlockDeclarationStatusOf(seat)` answers `none` (not defending, or
not in the combat steps from declare blockers on), `pending` or `declared`. A
defending player is one some attacker's attack is defended by, through the
same `defendingPlayerForAttackerLocked` read the verb and the option generator
use (Decisions 34-35), so "who declares blockers" has one answer.

**The four completion points.** A defender's declaration completes:

1. **as the step begins, when they have no legal block** —
   `autoCompleteBlockDeclarationsLocked`, a new `StepDeclareBlockers` case in
   the step-entry hook. "Declared, none", with nothing to ask;
2. **when they send `finish_blocks`** (`Game.FinishBlocks`, player-scoped, NOT
   priority-gated — a defender does not hold priority while the active player
   does, the reason `declare_blocker` is not either);
3. **when they pass priority in the step** — how a table that blocks by hand
   has always said "done", and still does;
4. **as the cursor leaves the step** with them still pending — the priority
   wrap and `AdvanceStep` complete everyone left, as whatever is staged, because
   the step cannot end on an unfinished turn-based action.

**Completing announces.** `completeBlockDeclarationLocked` marks the defender,
runs the #830 lock-in — which now announces only blocks whose controller's
declaration is complete — and then emits **`EventBlockersDeclared`** (Actor =
the defender, Amount = how many creatures they blocked with, 0 for "declared,
none"). One per defender per combat, after their `EventBlock` /
`EventBecomesBlocked` in the same batch, so the blocked record is written
before anything reading the new event asks it.

**The active player gets priority back.** When a completion by (2) or (3) is
the LAST one pending, the whole CR 509.1 action is over: state-based actions
and the trigger drain run and the ACTIVE player receives priority
(CR 509.2 / 117.3a). Concretely, a two-player step now runs: active player
passes (before blocks — the sandbox still allows it), defender blocks and
passes, **active player has priority again**, both pass, the step ends. Before,
the defender's pass wrapped and ended the step, and the attacker never had a
window after blockers — the window ninjutsu is activated in. Completion points
(1) and (4) do not move priority: at (1) the active player already holds it,
and (4) is the step ending.

**What stays permissive.** `declare_blocker(s)` is NOT refused after the
seat's declaration completes. A table that blocks by hand and passed a beat
early can still put the block down (it is announced at the next boundary,
#830's "a late block announces its own block"), and every test and client that
blocks without finishing keeps working. What changes is that nothing OFFERS a
block after completion: `blockOptionsLocked` answers empty for a declared
defender, so the enumerator, the bot and the #328 signal stop asking. This is
a sandbox allowance, not a rule, and a manual late block on a ninja that
entered attacking after the declaration is exactly as possible as it was.

### Decision 39: which consumers change

| Consumer | Before | After |
|---|---|---|
| `UnblockedAttackerForEffect` (ninjutsu's cost, Decision 31) | attacking, step, nothing blocking | the same **plus the attacker's defending player has declared**; while they are pending the attacker is neither blocked nor unblocked |
| "becomes blocked" / "blocks" (`EventBecomesBlocked`, `EventBlock` — afflict, Cyberman Patrol, Grazilaxx) | announced at the first priority boundary after the click | announced when the blocking player's declaration completes; a staged block by a pending defender waits through unrelated boundaries |
| "attacks and isn't blocked" (CR 509.3) | could not be written | `effects.WhenAttacksAndIsNotBlocked` watches `EventBlockersDeclared` and asks `UnblockedAttackerForEffect` of a creature attacking the event's Actor |
| #328 auto-pass (`block_decision_seats`, `SeatOwesBlockDecision`) | "under attack and has a legal block" | "still declaring and has a legal block" — a declared defender drops out even with a creature at home |
| Legal-move enumerator (block moves) | offered while any legal block remained | offered only while the seat is still declaring |
| Bot block grace (`aiseat.Runner.shouldHoldForBlockers`) | enumerated every other seat's moves for a `KindBlock` | reads `block_decision_seats` off the frame; the hold ends when the defender finishes. No longer a correctness guard — the engine returns priority to the active player after the last declaration — only a saving of the extra round |
| Client (`Game.svelte`, `priority.ts`) | no way to finish without priority | a "Done blocking" / "No blocks" control while the viewer's seat is in `block_pending_seats`, sending `finish_blocks` |

**The wire.** `TurnView` gains `block_pending_seats` and
`blocks_declared_seats` (public, omitted outside the step); a defending seat is
in exactly one. `block_decision_seats` keeps its key with the narrowed meaning
above. New action `finish_blocks`; new event kind `blockers_declared`, a
deliberate silence in the public log (its blocks are the `block` lines, its
completion is on the turn cursor) — a "declares no blockers" line is a
reasonable follow-up that needs a log kind and client copy.

### Proof cards

Four, all `full`: **Swamp Mosquito** and **Guiltfeeder** (poison / life loss to
the defending player), **Abyssal Nightstalker** (the defender's chosen discard),
and **Eternal of Harsh Truths**, whose afflict and "isn't blocked" draw are the
two answers one completed declaration gives — exactly one of them fires.

### Tests

`server/internal/game/block_completion_test.go`: declared-none at step start
for a defender with no legal block (and the event's Amount 0); pending for one
who has not acted (attacker neither blocked nor unblocked, listed pending);
`finish_blocks` with nothing staged; a staged block announced at completion and
not at an unrelated boundary, the declaration event after the block events;
the defender's pass completing and returning priority to the active player;
`AdvanceStep` completing a pending declaration; two defenders, where only the
second completion moves priority; a declared defender offered no blocks but
still able to block by hand; the refusals; undo and snapshot round trips;
combat end forgetting who declared. `declare_blockers_test.go` and the legal
package's evasion test now pin that a creature arriving after a no-legal-block
declaration does not reopen it. Card tests in `attacks_unblocked_test.go` and
`ninjutsu_test.go` (ninjutsu refused while the defender is declaring, paid once
they finish); the wire in `blockers_view_test.go`; dispatch in
`actions/declare_blockers_test.go`; the bot grace in
`aiseat/runner_test.go`; the client helper in `priority.test.ts`.

### What this does NOT decide

- **Refusing a late block.** Decision 38 keeps the verb permissive. Refusing a
  block after the seat's declaration is complete would be the rules-exact
  posture and would make "a ninja cannot be blocked" an engine fact rather
  than a table convention; it is a one-line gate in `declareBlockersLocked`
  when someone wants it, and it would retire #830's late-block paths.
- **Priority at the step's start.** The active player still receives priority
  on entry, before any declaration, which CR 509.1 does not give them. Parking
  priority (`NoPriority`) until every defender has declared would be exact, but
  it would make `pass_priority` unavailable to the defenders who use it to
  finish and would need a bot move kind for `finish_blocks`; the priority
  return above gives the active player the window that matters without either.
- **A log line for "declares no blockers"**, above. *Closed by #1500: a
  `no_blocks` log kind, narrated from `EventBlockersDeclared` only when
  `Amount == 0` — the declaration's blocks stay on the `block` lines, which
  is the reasoning this amendment already gave.*

---

## Amendment (2026-09-24, [#750](https://github.com/krakenhavoc/cmd_and_ctrl/issues/750)): the card half of block rules

Decisions 11-13 built the engine half of block rules: slot 4 of
`BlockPairRefusalLocked`, the bounds `blockerBoundsLocked` combines, the
whole-declaration refusal in `DeclareBlockers`, and the two registries
(`CatalogBlockRules`, `Game.TurnScopedBlockRules`). No card could declare a
rule, because nothing set `CatalogBlockRules`. This amendment is the card half
Decision 11 and the PR split's PR 4 and PR 6+ describe. Decisions 1-39 stand.
Sprint S37 (combat correctness), tracker
[#880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/880). The deck that
asked for it is *Aang is so flashy*
([#1306](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1306)): Avatar
Kuruk's Spirit token prints "can't block or be blocked by non-Spirit
creatures".

### Decision 40: one slot on the Spec, one on the token template, one on `CardDef`

- **`effects.Spec.BlockRules []game.BlockRule`**, projected by `buildDef` onto
  **`game.CardDef.BlockRules`**, which carddef.go publishes through
  `CatalogBlockRules`. The hook was always keyed by `CatalogAbilityKey`, so a
  permanent that loses all abilities imposes nothing (Decision 11's "losing
  abilities works without extra code" holds with no new code).
- **A token template has the same slot** (`tokenTemplate.BlockRules`,
  `effects/token_catalog.go`), registered under the token's key like its
  triggers and statics (ADR 0083). #1249 already made `forEachBlockRuleLocked`
  skip on an empty key rather than an empty oracle ID, so a token's rule is
  found exactly as a card's is. A template whose only ability is a block rule
  is registrable: `checkTokenTemplate` counts the slot.
- **Kuruk's Spirit is its own template** (`spirit-only-spirits`), not the plain
  `"1/1 colorless Spirit"` row Forbidden Orchard makes. The Orchard's Spirit
  prints nothing, and a shared row would either give it a restriction or take
  Kuruk's away.

### Decision 41: a rule is a scope plus a rule, and the scope is never optional

Every rule is read off every permanent on the battlefield for every pair the
engine checks. A `Pair` closure that forgets to ask "is this attacker mine?"
binds every attacker at the table. So `effects/block_rules.go` builds every
rule from a **scope** that says whose creatures it binds relative to its source,
and takes the scope as an argument:

| Scope | Printed as |
|---|---|
| `OnSelf()` | "this creature", "Legolas" |
| `OnAttached()` | "equipped creature", "enchanted creature" |
| `ControlledBySourceController()` | "creatures you control" |
| `OnMatching(pred)` | "Slivers", "non-Spirit creatures" — either side of the pair |
| `PowerLessThanSource()` | "creatures with power less than this creature's power" |

| Rule | Reason on the wire | First card |
|---|---|---|
| `CantBeBlockedExceptBy(scope, allowed, label)` | `cant_be_blocked_except_by` | Prowler's Helm, Departed Deckhand, Canopy Cover |
| `CantBeBlockedBy(scope, forbidden, label)` | `cant_be_blocked_by` | Legolas Greenleaf |
| `CantBeBlockedWhile(scope, cond)` | `cant_be_blocked` | Thieves' Tools |
| `CantBlockAttackers(blockers, attackers, label)` | `cant_block_attacker` | Champion of Lambholt |
| `CantBlockOrBeBlockedBy(scope, other, label)` — both of the above, one per side | both | Avatar Kuruk's Spirit token |
| `MaxBlockers(scope, n)` | `too_many_blockers` | Vorrac Battlehorns |
| `MinBlockers(scope, n)` | `too_few_blockers` | Rampaging Ceratops |

Decision 11 named the blocker-side scope `BlockerMatching(pred)`. It is
`OnMatching(pred)`, because a scope is a predicate on a creature and does not
care which side of the pair it is asked about. `CantBlockOrBeBlockedBy` uses the
same one on both sides.

Predicates are the targeting vocabulary (`CardPredicate`), evaluated with the
rule source's controller as the caster, so `YouControl()` in a rule means "the
controller of the permanent that prints it". Everything reads effective
characteristics live when the block is checked (§5, CR 509.1b).

`CantBeBlockedWhile` is a block rule and not a `Restriction` bit behind a
layer-6 condition, because Thieves' Tools' condition is the equipped creature's
power. Power is finished in layer 7, after a layer-6 static has already run. A
block rule reads it after every layer.

### Decision 42: the until-end-of-turn twin is one applier, not one per rule

Decision 11 foresaw an `…UntilEOT` twin per builder. There is one,
`BlockRuleUntilEOT{Target | Match, Rule}`: it snapshots the affected set exactly
as `RestrictUntilEOT` does (CR 611.2c: the instance ID plus its entry stamp, so
a creature that leaves and returns is a new object and is not covered), hands
`Rule` a scope for that set, and registers the result in
`Game.TurnScopedBlockRules`. Gingerbrute's `{1}` and Departed Deckhand's `{3}{U}`
are each one line with it. A turn-scoped rule has no source permanent (the
effect outlives the ability that made it, CR 611.2b), so predicates inside it
see no caster. No printed turn-scoped rule needs one.

### What this does NOT decide

- **`BlockRule.Limit`** (Silent Arbiter's "no more than one creature can block
  each combat") is still unbuilt, as Decision 12 left it, and so is the
  attack-side twin Decision 18 puts out of scope (Crawlspace). Silent Arbiter
  needs both, so they ship together: [#1507](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1507),
  its own seam row.
- **PR 5's per-card stamps** (`blockable_attackers`, `blockers_min`,
  `blockers_max`, Decision 16) are not added. The wire already carries every
  refusal's reason and label in the `illegal_block` frame, and
  `block_decision_seats` is computed from the same option generator the
  enumerator uses, so the three agree without them.
- **Hungering Hydra and Alpha Authority** are not in this change. Both are one
  `MaxBlockers` line; the Hydra also needs "whenever this creature is dealt
  damage, put that many +1/+1 counters on it", which is a card-side trigger,
  not this seam.

### Tests

`server/internal/cards/effects/block_rules_test.go`: for each rule shape, a
legal block accepted and an illegal one refused with its reason, label and
source; the legal-move enumerator never offers a refused block and
`SeatOwesBlockDecision` agrees; the token carries its rule both ways; the
turn-scoped rule binds for the turn, is swept at its end, and does not follow a
creature that left and came back; a creature that loses its abilities loses
its rule.

---

## Amendment (2026-09-24, [#1507](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1507)): combat-wide attack and block limits

Builds what Decision 12 reserved (`BlockRule.Limit`) and what Decision 18 put
out of scope (the attack-side count limit), together, because their first card
needs both. Decisions 1-42 stand. Sprint S37 (combat correctness), tracker
[#880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/880). ADR 0080 §8
pointed here for the same two cards.

### The rules

- **CR 509.1b** lets a restriction bound the number of creatures that block.
  Silent Arbiter's "No more than one creature can block each combat" bounds
  the WHOLE combat rather than one attacker (Hungering Hydra's bound is per
  attacker, and is `Count`, Decision 12).
- **CR 508.1c** does the same for attacks: "No more than one creature can
  attack each combat" (Silent Arbiter, Dueling Grounds), "No more than two
  creatures can attack you each combat" (Crawlspace). The declaration as a
  whole must obey it. No single creature is forbidden to attack.
- A restriction is checked when attackers or blockers are declared, and only
  then (§5). A limit that arrives afterwards unmakes nothing.

### Decision 43: `BlockRule.Limit` is a third set check in the block validator

```go
// Limit returns the bound `blocker` counts toward, or 0 when this rule
// does not count it.
Limit func(g *Game, blocker, source *Card) int
```

- **Where it is judged.** `checkBlockDeclarationLocked` builds the
  assignment the action would leave behind (`after`) and checks it after the
  per-pair checks and the per-attacker bounds. The Limit check is the third
  entry in that list of set checks (`blockLimitRefusalLocked`,
  `game/block_rules.go`). It counts every blocker in the combat, whichever
  attacker it blocks and whichever defender declared it.
- **When it refuses.** A rule refuses only when it counts more blockers in
  `after` than its bound **and** more than it counts in the assignment the
  action starts from. The second half follows the same idea as the
  "attackers the action does not touch are not re-judged" rule in the
  per-attacker bounds. A Silent Arbiter flashed in after two blocks leaves
  both blocks standing, and re-pointing one of them is not refused for a
  count it does not raise.
- **The reason** is `declaration_limit`. Decision 8 reserved the token, and
  this is the first change to send it. `BlockRefusal.N` is the bound and
  `Source` is the permanent that prints it. The refusal also carries a new
  `SourceName`, because the limit is printed on a third card that the
  player has to answer. The sentence is "No more than one creature can
  block each combat (Silent Arbiter)." A rule's `Label`, when set, replaces
  the clause.
- **Several limits.** Each rule is judged on its own, so the tightest bound
  wins: a Caverns of Despair beside a Silent Arbiter allows one blocker. All
  the blockers one rule counts share one bound, the smallest any of them
  reported.
- **Multiplayer.** This engine stages each defender's declaration separately
  (Decision 38). Under a combat-wide limit, the first defender to block uses
  it up, and a later defender's block is refused. In the rules, all
  defending players declare at once and have to agree. When declarations are
  staged one at a time, first-come is the only order the engine can give.
  The option generator runs every candidate through the validator
  (Decision 14), so it stops offering the second defender anything, and the
  #328 signal releases them.
- **The constructor** is `effects.NoMoreThanNCanBlockEachCombat(n)`. It takes
  no scope, unlike every other rule constructor (Decision 41), because the
  printed line has none: it binds every creature at the table, including
  those of the permanent's own controller.

### Decision 44: the attack-side twin is `game.AttackLimit`, a scope and a number

```go
type AttackLimit struct {
	Scope AttackLimitScope // AttackLimitAttackingYou (zero value) | AttackLimitEachCombat
	Max   int
}
```

It is not a `BlockRule`, for two reasons. An attack has no blocker to hand a
rule. And the "you" family is scoped by the DEFENDING SEAT, which is the axis
`AttackTax` already keys on (ADR 0080). So the struct follows `AttackTax`:

- **The "you" is structural.** `AttackLimitAttackingYou` counts attacks on the
  limit's controller, and "you" means the PLAYER. An attack on a planeswalker
  they control does not count. This is ADR 0080's reading of Propaganda's
  same word, and the reading of every printed card in the family.
- **The zero value is the narrower scope**, so a card file that forgets to
  choose limits less than printed, never more.
- **Collection.** Limits come from `CardDef.AttackLimits` and reach the engine
  through `CatalogAttackLimits`, keyed by `CatalogAbilityKey`. A permanent
  that has lost all its abilities (CR 613.1f) therefore limits nothing.
  `Spec.AttackLimits` is built with `NoMoreThanNCanAttackEachCombat(n)` /
  `NoMoreThanNCanAttackYouEachCombat(n)`.
- **The one check** is `attackLimitRefusalLocked` (`game/attack_limits.go`).
  It applies the declaration on top of every creature attacking now and uses
  the same raise-the-count rule as Decision 43. Re-pointing an attacker from
  one opponent to a Crawlspace player counts against the Crawlspace player,
  and moving one away is never refused.
- **Both verbs enforce it,** before the CR 508.1a tax is priced, so a refused
  declaration owes nothing:
  - `DeclareAttackerWith` refuses the one creature. The verb is lax about
    eligibility for the sandbox's hand-forcing, but a limit is not
    eligibility. It is the card doing its whole job, the same reason "can't
    attack" is not relaxed.
  - `DeclareAttackersWith` refuses **all or nothing**. The verb skips
    INELIGIBLE entries, but an over-full swing has none: each creature is
    fine on its own. Which creatures stay home is the attacking player's
    choice, as ADR 0080 has it for a tax the seat can only partly afford.
- **The enumerator** asks the same function of every candidate
  (attacker, target) move and withholds the refused ones (#544). Its attacks
  are one creature at a time, so under Silent Arbiter a seat is offered
  attacks until one creature is attacking, and then none. Under Crawlspace it
  is offered attacks at that player until two creatures are attacking them,
  and it keeps being offered attacks at everyone else.
- **What counts as attacking.** Every creature with an `AttackingTarget`
  counts, including one put onto the battlefield attacking, which CR 506.3c
  says was never declared. That cannot matter to any declaration the rules
  allow: such a creature can only arrive after the lock-in, when the rules'
  declaration is already over. It can only refuse a sandbox LATE
  declaration, which is the weaker direction.

### Decision 45: the wire

- A refused block is `illegal_block` / `declaration_limit`, as in
  Decision 8, with `card_id` set to the first blocker in the declaration that
  the limit counts.
- A refused attack gets a new code, **`illegal_attack`**, with `reason`
  **`attack_limit`** and `card_id` set to the first attacker in the refused
  declaration that the limit counts. The sentence is built by the engine and
  addressed to the reader: the protected seat reads "No more than two
  creatures can attack you each combat (Crawlspace)", and anyone else reads
  that seat's name. It is a new code rather than `bad_request` so that a
  client can tell an attack the rules refuse from a malformed request. The
  client needs no change: it shows any error frame's message.

### Proof cards

**Silent Arbiter**, **Dueling Grounds** (both halves, one), **Crawlspace**
(attacking you, two) and **Judoon Enforcers** (attacking you, one, beside
trample and suspend) ship `full`. **Caverns of Despair** (both halves, two)
ships with the world-rule caveat Concordant Crossroads already declares.

### Tests

- `game/combat_limits_test.go` covers the engine with the hooks stubbed. The
  second attacker is refused, with nothing staged. The bulk verb is all or
  nothing. "Attacking you" is per defender in a four-seat game, re-pointing
  onto the Crawlspace player is refused and re-pointing off it is allowed,
  and planeswalker attacks are not counted. A late limit unmakes nothing, on
  both sides. On the block side: a block is limited across attackers and
  across defenders (the generator and the #328 signal agree), menace under a
  one-blocker combat is unblockable, and the tightest bound wins.
- `cards/effects/combat_limits_test.go` runs one test per proof card. Each
  checks that the enumerator offers EXACTLY what the verb accepts, by trying
  every candidate on a clone. It also covers a Silent Arbiter that has lost
  its abilities and a Crawlspace that does not limit its own controller.
- `aiseat/combat_limits_test.go` drives one combat move by move with the
  heuristic and with a scripted policy that attacks and blocks with
  everything, under Silent Arbiter (two seats) and Crawlspace (four seats).
  The combat ends, every enumerated move is accepted, and the limits hold.
- `ws/attack_limit_error_test.go` covers the `illegal_attack` frame and the
  classifier. `ws/block_refusal_error_test.go` picks up `declaration_limit`
  through `BlockReasons()`.

### What this does NOT decide

- **Conditional and per-player variants.** Mirri, Weatherlight Duelist
  ("as long as Mirri is tapped, no more than one creature can attack you";
  "each opponent can't block with more than one creature this combat") would
  need a `While` gate on `AttackLimit` and a per-player group on
  `BlockRule.Limit`. The Eternal Wanderer ("no more than one creature can
  attack The Eternal Wanderer") would need a third `AttackLimitScope`. Each
  is one field, and none of them is built without its card.
- **Turn-scoped attack limits.** No printed card needs one, so there is no
  `TurnScopedAttackLimits` registry.
- **Reselection** (CR 508.7, `game/attack_reselect.go`) does not go through
  the declaration verbs and does not consult the limits.
- **The client's "attack with all"** under a limit is refused whole with the
  sentence. It does not offer the tax picker's "choose attackers…" flow. See
  the follow-up issue.
