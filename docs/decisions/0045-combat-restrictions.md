# ADR 0045 — A restriction vocabulary for declarations and activations

**Status:** Accepted · 2026-09-14 · Sprint S24 · Relates to [#76](https://github.com/krakenhavoc/cmd_and_ctrl/issues/76), [#544](https://github.com/krakenhavoc/cmd_and_ctrl/issues/544), [#546](https://github.com/krakenhavoc/cmd_and_ctrl/pull/546), [ADR 0036](0036-attachments.md), [ADR 0038](0038-protection-style-keywords.md)

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
