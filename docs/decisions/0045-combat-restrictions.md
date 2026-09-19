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
  `BlockerCountValid` validates a block count at the declare-blockers
  close-out for menace, and `DeclareAttackers` receives the entire
  attacking set in one call. A count limit belongs beside
  `BlockerCountValid` as a set-shaped predicate, with the per-permanent
  bits in this ADR left alone. The enumerator agreement is harder for
  those than for these, because a per-move enumerator has to reason
  about a set — which is the open question to answer before writing
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
`declaration_limit`, `not_defending`, `tapped`, and a reserved
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
  more entry there.

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
  shape would work, but no card in this wave needs it.
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
and this is its answer. Decision 13's own PR has not landed yet; this
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
  Patrol (afflict, CR 702.131) and Grazilaxx watch this and need no
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
