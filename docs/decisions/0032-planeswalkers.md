# ADR 0032 — Starting loyalty is printed card data, not catalog data

**Status:** Implemented · 2026-09-11 · Branch `fix/planeswalkers-274`
**Issue:** [#274](https://github.com/krakenhavoc/cmd_and_ctrl/issues/274)
**Extends:** [ADR 0007](0007-stack-foundation.md) §6 (loyalty resolves
immediately), [ADR 0008](0008-counter-mechanics.md) (the 704.5i SBA),
[ADR 0010](0010-card-effect-catalog.md) §8 (`Spec.StartingLoyalty`),
[ADR 0020](0020-activated-abilities.md) (loyalty excluded from
`AbilityCost`), [ADR 0027](0027-attack-triggers.md) §1 (no
planeswalker defender).

## Context

Planeswalkers have been half-built since S13.1 and nobody had written
down which half. Five ADRs mention them in passing and none is about
them; the result was a load-bearing claim that turned out to be
false, and a production bug that made every planeswalker in the game
unplayable.

The false claim is ADR 0010 §8:

> the ETB hook stamps loyalty counters via `AddCounterForEffect`
> **before** the next SBA boundary — otherwise a planeswalker with
> `StartingLoyalty = 0` … would die to the S13.2 0-loyalty SBA.

The ordering half of that was true. The *sourcing* half was the
problem: `StartingLoyalty` was a field on `effects.Spec`, so the only
planeswalkers that ever got loyalty counters were the ones somebody
had written a catalog entry for. ADR 0010 §8 also records that "The
Wandering Emperor is the only S14 planeswalker", and no planeswalker
has been added to the catalog since. So the set of playable
planeswalkers in this engine was exactly one card.

Issue #274 is what that looks like from a chair: Ian cast Teferi in a
live game on 2026-09-11 and watched him land in the graveyard. The
pinned replay (report `0774a326-3cae-4b2a-9ac7-8b7f3620afcb`, game
`fe34c746`) settles the details — it was **Teferi, Time Raveler**,
`Legendary Planeswalker — Teferi`, in hand from seq 30, on the stack
at seq 96, in Ian's graveyard at seq 98 with `counters: null` and no
`auto` bit. He never appears on the battlefield in any snapshot: he
resolved onto it with zero loyalty counters and CR 704.5i swept him
off before the next snapshot was taken.

This is a systemic failure, not a Teferi failure. It breaks the
engine's central promise — that a card with no catalog entry still
works as a manual sandbox object — for an entire permanent type.

## Decisions

### 1. Printed loyalty lives on `game.Card`, alongside power and toughness

`Card.StartingLoyalty` (`card.go:96`) is an int stamped at deck-import
time, exactly like `Power`, `Toughness`, `ManaCost` and `Colors`
already are. Scryfall's `loyalty` string is carried through
`cards.Card.Loyalty` (`index.go:95`) and parsed by
`deck.printedLoyalty` (`deck.go:261`), called from `toGameCard`
(`deck.go:292`).

**Why not keep it in the catalog:** starting loyalty is printed on the
card, in the same corner as power/toughness. It is not behaviour, so
it does not belong in a behaviour catalog. Every other printed
attribute the rules engine reads had already made this journey; loyalty
was the one that had not, and the opt-in nature of the catalog turned
that omission into "this card type does not work".

Non-numeric loyalty parses to zero, matching how `Power` handles `*`
(`deck.go:261-276`). Those cards stay a manual-sandbox case.

### 2. Double-faced cards read loyalty off the face

Scryfall puts power / toughness / loyalty on the **face** for
double-faced cards and leaves the top-level field empty, so
`CardFace.Loyalty` (`index.go:135`) exists and `printedLoyalty` falls
back to the first face that prints a number.

This matters more than it looks. `Card.TypeLine` is stamped from
Scryfall's top-level type line, which for a transforming card is the
joined `"Legendary Creature — Elf Scout // Legendary Planeswalker —
Nissa"`. `typeLineHas` is a substring check (`card.go:410-425`), so
`IsPlaneswalker()` is **true** for Nissa, Vastwood Seer — and before
this change she resolved as a 0-loyalty planeswalker and died on the
spot. Reading the face's loyalty makes her survive. It also puts
loyalty counters on what is visibly a creature, which is cosmetically
wrong and strictly better than the alternative. DFC modelling is a
real gap and is called out in §7.

### 3. The stamp moved into the `game` package, onto every entry path

`game.stampStartingLoyaltyLocked` (`effect_hooks.go:297`) does the
work, called unconditionally at the top of `fireETBHookLocked`
(`effect_hooks.go:256`) — **before** the `ETBEffectHook == nil ||
oracleID == ""` guards that used to swallow it.

`fireETBHookLocked` was chosen as the site because it is already the
single funnel every battlefield entry passes through: spell resolution
(`mutations.go:1272`), land cast (`mutations.go:669`), reanimation
(`effect_api.go:720`), token creation (`effect_api.go:1431`), the CR
614 entry pipeline (`entry_choice.go:326`), manual `move_card`
(`mutations.go:2664`), and `effect_api.go:1130`. Seven doors, one
lock. Nothing was renamed, so no call site changed.

Running before the nil-hook guard is the point: loyalty now works in a
server built without the `effects` blank import, and in the `game`
package's own tests, where the catalog hooks are nil by construction.
That configuration is exactly the shape of a non-catalog card, which
is why the regression test for #274 registers no hooks at all.

### 4. `Spec.StartingLoyalty` survives as a fallback, not as the source

`effects.fireOnETB` no longer stamps anything (`wire.go:269-274`).
Instead `init` registers `game.CatalogStartingLoyalty`
(`wire.go:28`), and the stamp consults it **only when the card carries
no printed value** (`effect_hooks.go:53`, `effect_hooks.go:297-328`).

Precedence is printed-then-catalog, with an "already has loyalty
counters" guard in front of both. Without that guard a deck-imported
The Wandering Emperor would enter with 6 — printed 3 plus catalog 3.
`TestPrintedLoyaltyNotDoubleStampedWithCatalog` pins it.

**Why keep the catalog path at all:** tokens, test fixtures and the
demo seed never go through `deck.toGameCard`, so they have no printed
data to read. Deleting the field would have broken
`TestWanderingEmperorEntersWithStartingLoyalty` for no gain. Keeping
it costs one hook.

### 5. The stamp goes through `AddCounterForEffect`, not a map write

`effect_hooks.go:329` calls `AddCounterForEffect` rather than writing
`Counters["loyalty"]` directly, so a planeswalker entering under
Doubling Season gets its loyalty doubled through the CR 614
replacement pipeline (`RepEventCounter`). A direct write would have
silently skipped every counter replacement effect.

### 6. CR 704.5i is unchanged and stays unchanged

The 0-loyalty SBA (`mutations.go:1836-1839`) was never the bug and was
not touched. A planeswalker with nothing printed and nothing in the
catalog still enters at zero and still dies —
`TestPlaneswalkerWithoutPrintedLoyaltyStillDies` exists specifically
so that "fix #274" never quietly becomes "planeswalkers are
immortal".

### 7. What is implemented, what half-works, and what is absent

This is the audit the codebase did not have. Everything below is as of
this branch.

**Works.**

- **Type identity.** `IsPlaneswalker()` (`card.go:383`) and
  `IsPermanent()` (`card.go:396`, which includes planeswalkers per CR
  110.4).
- **Zone routing.** A resolving planeswalker spell takes the permanent
  branch to the battlefield (`mutations.go:1243-1272`). This was never
  broken; #274 only looked like a routing bug.
- **The loyalty stamp**, per §1–§5 above, on all seven entry paths.
- **CR 704.5i**, the 0-loyalty SBA (`mutations.go:1836-1839`), and its
  battle-defense twin.
- **Loyalty as a counter.** `CounterLoyalty` is in the registry
  (`counter_types.go:29`, `:83`) and renders client-side as a purple
  `L n` pip (`counterTypes.ts:43` + `CounterPips.svelte:22-41`) plus a
  green bottom-right badge (`Card.svelte:132-134`, `:301-315`,
  `:501-507`). Two renderings of the same number in two colours is a
  cosmetic wart, not a bug.
- **Sorcery-speed cast timing**, by accident: the client's
  `timing.ts:170-174` falls through to the non-instant permanent
  branch, which is the right answer.

**Half-works.**

- **Loyalty abilities.** Two separate half-paths, neither complete,
  and the confusion between them is why the decklist triage and the
  docs pass appeared to contradict each other. Both were right.

  `ActivateLoyalty` (`mutations.go:1531`) is the S13.1 sandbox path.
  It **does** enforce the gates: sorcery speed (CR 606), once per turn
  per permanent via `LoyaltyActivatedThisTurn` (`mutations.go:1546`,
  cleared at `game.go:553`), split-second, and battlefield presence.
  What it does not do is **anything else**. `delta` is a raw int the
  client supplies, `label` is explicitly discarded (`mutations.go:1585`
  — `_ = label`), and no effect of any kind runs. It moves a counter.
  It also does not check that the target is a planeswalker (any
  battlefield card will take loyalty counters), does not check that
  the activating player controls it, and does not enforce CR 606.3's
  "you cannot remove more loyalty than is there" — the comment at
  `mutations.go:1572-1574` deliberately allows negative so the SBA can
  see the intent.

  `ActivateAbility` / `ActivatedAbilityShape` (`activated.go:65`) is
  the S21 pipeline that actually puts an ability on the stack and runs
  an `Effect` (`activated.go:84`). It cannot express a loyalty ability
  because `AbilityCost` (`activated.go:35`) has exactly `Tap`,
  `SacrificeSelf`, `SacrificeOther`, `Mana` and `Life` — no loyalty
  component. ADR 0020 excluded it on purpose.

  **So the answer to "what is missing to make a loyalty ability run
  its effect" is: a `Loyalty int` component on `AbilityCost`, its
  payment step, and the `LoyaltyActivatedThisTurn` gate wired into
  `ActivateAbility`.** The effect machinery, the stack item, the
  targeting and the sorcery-speed flag all already exist. No catalog
  card declares a loyalty ability today — `Loyalt` appears in
  `spec.go` only as `StartingLoyalty` (`spec.go:63`, `:69`).

  On the client this is worse than half: `activate_loyalty` is not in
  the `ActionType` union (`protocol.ts:60-95`), so sending it is a
  TypeScript compile error, and `canActivateLoyalty` (`timing.ts:209`)
  is dead code reached only by its own unit tests. The only way a
  player changes loyalty today is the right-click quick-counter menu
  (`contextMenu.logic.ts:68-72`), which is a sandbox escape hatch with
  no gates at all.

**Absent.**

- **Attacking a planeswalker.** Confirmed: there is no planeswalker-
  defender concept anywhere. `Card.AttackingTarget` is documented as
  "the player ID" (`card.go:158-162`) and `DeclareAttacker`
  (`mutations.go:3326`) rejects anything that is not a seated player
  outright — `if g.playerByIDLocked(targetPlayerID) == nil { return
  ErrPlayerNotFound }` (`mutations.go:3335-3337`). Passing a
  planeswalker's instance ID is an error, not a legal declaration.
  The client matches end to end: attack targets come from
  `view.seats` only (`Game.svelte:530-537`,
  `PlayerPanel.svelte:183-185`, `contextMenu.logic.ts:492-507`) and
  combat arrows resolve their destination via `[data-seat-id]`
  (`CombatArrows.svelte:43-49`, `:160-161`), so a card-valued target
  would draw a dangling arrow. Block-side too: incoming attacks are
  detected with `c.attacking_target === viewerID`
  (`Game.svelte:483-486`). ADR 0027 §1 recorded this gap for Hellrider
  and it is unchanged.
- **Damage removing loyalty (CR 120.3c).** Damage aimed at a
  planeswalker lands in `DamageMarked`
  (`effect_api.go:216-232`, `mutations.go:2178-2184`) and is then read
  by nobody: the lethal-damage SBA is behind an `IsCreature()` guard
  (`mutations.go:1812`) and the planeswalker branch reads only loyalty
  counters (`mutations.go:1836-1839`). The marked damage is wiped at
  cleanup (`game.go:736`). Concretely: **Lightning Bolt targeting a
  planeswalker is a no-op.** Announce-time targeting permits it —
  `TargetMode: "any"` is documented as "player / creature /
  planeswalker / battle" (`spec.go:78-80`) and `DealDamage.Apply`
  routes any battlefield card to `DealDamageToCreatureForEffect`
  (`primitives.go:36-47`) — so it is worse than a refusal: it looks
  like it worked.
- **Damage redirection.** Correctly absent. The redirect-to-
  planeswalker rule was removed from the CR in 2018; the modern rules
  have you attack or target the planeswalker directly. Nothing should
  be built here; the two items above are its replacements.
- **Planeswalker uniqueness / the legend rule (CR 704.5j).** Not
  implemented for any card type — `IsLegendary` does not exist and the
  only mention of the legend rule in the server is a comment
  (`token_copy.go:139`).
- **Loyalty as a layered characteristic.** Deliberate and correct.
  `Characteristic` (`characteristic.go:31-40`) has no `Loyalty` field
  and says why: loyalty is a counter (CR 121), not a layered
  characteristic (CR 613). Note that `docs/sprints.md:1304` claims
  otherwise and is stale.
- **Client type awareness.** `cardTypes.isPlaneswalker`
  (`cardTypes.ts:32-34`) exists and has zero importers —
  `Card.svelte:132` re-implements it with its own regex.
  Planeswalkers are bucketed into a battlefield row literally
  labelled `"enchant / artifact"`
  (`cardTypes.ts:51-55`, `PlayerPanel.svelte:245`). `targeting.ts` has
  no `"planeswalker"` mode (`targeting.ts:29-38`). There is no
  planeswalker icon and no e2e coverage.

## Consequences

- Every planeswalker in the Scryfall index is now castable and sticks
  to the battlefield with the right number of loyalty counters,
  catalog entry or not. This is the engine's non-catalog promise
  restored for a permanent type that never had it.
- Reanimating a planeswalker now gives it loyalty. That failure was
  written down in
  [hashaton-the-cheater.md](../decklists/hashaton-the-cheater.md)
  as a known casualty of the same hook and is fixed as a side effect
  of §3, with `TestReanimatedPlaneswalkerGetsPrintedLoyalty` pinning
  it.
- The index grows one string field per card. The bulk dump is streamed
  and only decoded fields are retained (`index.go:159-190`), so the
  cost is ~40k short strings, almost all empty.
- Engine coverage (`server/internal/game/planeswalker_test.go`): the
  #274 regression itself; the 704.5i guard still firing on a
  loyalty-less walker; the catalog fallback; the no-double-stamp
  precedence rule; reanimation; and manual `move_card`.
- Ingestion coverage (`server/internal/deck/deck_test.go`,
  `TestPrintedLoyaltyStampedOnGameCard`): top-level loyalty, the DFC
  face fallback, non-numeric loyalty, and a creature staying at zero.
- No client change. Loyalty counters already render; there was nothing
  to display that was not already displayed once the counters exist.

## What this deliberately does not do

Each of these is S27 ([#92](https://github.com/krakenhavoc/cmd_and_ctrl/issues/92))
scope and none is made harder by this change.

- **Loyalty abilities that do anything.** See §7. The shortest path is
  a `Loyalty` component on `AbilityCost` plus the once-per-turn gate
  moved into `ActivateAbility`, after which loyalty abilities are
  ordinary S21 activated abilities and the existing `ActivateLoyalty`
  action can be retired rather than extended.
- **Attacking planeswalkers.** Needs `DeclareAttacker` to take a
  polymorphic defender and `AttackingTarget` to stop meaning "player
  ID", plus client work on arrows and block detection.
- **Damage removing loyalty (CR 120.3c).** Should land with, or
  before, attacking planeswalkers — combat damage to a walker is the
  main way loyalty leaves.
- **The legend rule (CR 704.5j).** Not a planeswalker-specific
  problem; it is missing for legendary creatures too.
- **Double-faced cards.** §2 is a workaround, not modelling. A DFC has
  one type line joining both faces and no notion of which face is up.

## Note on stale documentation

Three written claims were wrong or misleading when this branch
started, and two of them cost real time:

- **ADR 0010 §8** implies starting loyalty is handled. It is handled
  *for catalog cards*, which at the time meant one card. That
  omission is this ADR's subject.
- **`docs/sprints.md:1304`** lists `Loyalty` as a field on
  `Characteristic`. It is not, by design — see
  `characteristic.go:28-30`.
- **`docs/sprints.md:733`** ticks `canActivateLoyalty` as delivered.
  The function exists (`timing.ts:209`); nothing calls it, and the
  action it gates is not in the client's action union at all.
