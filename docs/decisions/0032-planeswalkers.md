# ADR 0032 — Starting loyalty is printed card data, not catalog data

**Status:** Implemented · 2026-09-11 · Branch `fix/planeswalkers-274`
**Amended:** 2026-09-11 · Branch `feat/loyalty-abilities` — §8 below
**Issue:** [#274](https://github.com/krakenhavoc/cmd_and_ctrl/issues/274),
[#329](https://github.com/krakenhavoc/cmd_and_ctrl/issues/329),
[#334](https://github.com/krakenhavoc/cmd_and_ctrl/issues/334)
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

- **Loyalty abilities.** *Fixed — see §8. The audit below stands as
  written and its prediction held; read it for why §8 looks the way
  it does.* Two separate half-paths, neither complete,
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

### 8. Loyalty abilities are ordinary CR 602 activated abilities

*Amendment, 2026-09-11, branch `feat/loyalty-abilities`. Closes
[#329](https://github.com/krakenhavoc/cmd_and_ctrl/issues/329) and
[#334](https://github.com/krakenhavoc/cmd_and_ctrl/issues/334).*

§7 predicted the shortest path and it held up exactly. Nothing in
the audit needed revising: `ActivateLoyalty` really did discard its
label and run no effect, `ActivatedAbilityShape` really did have the
whole stack-item pipeline waiting, and the missing piece really was
one cost component.

**What the replays show.** Two reports, two different Teferis, one
symptom.

- #329 (report `c3ecbf98-8bf7-4f57-9e90-16bf388e19c1`, 337
  snapshots) is **Teferi, Who Slows the Sunset** — not Time Raveler,
  which is worth recording because it means the bug was never
  card-specific. He is on the battlefield from seq 327 with
  `counters: {loyalty: 4}`, so §1–§5 of this ADR are working. His
  `activated_abilities` array is empty in every snapshot, and his
  `tapped` bit alternates `null` / `true` across seq 327–335 as the
  reporter clicks him and untaps him again. That is the whole bug
  report, in the data: "when I click on him to choose one of his
  abilities it just tapped him."
- #334 (report `2b4ce9cc-7be7-49bd-a44f-19abdcb2fb52`, 152
  snapshots) is **Teferi, Time Raveler**, loyalty 4, again with
  `activated_abilities: []` and the same tapped / untapped flip at
  seq 150–152. Neither client log contains a single
  `activate_loyalty` frame, because the client could not send one.

**`AbilityCost.Loyalty *int`** (`activated.go`) is the component.
A pointer, not an int, because `[0]` is a real printed cost (Jace's
`[0]: Brainstorm`) and has to stay distinguishable from "not a
loyalty ability" — the difference decides whether the activation
spends the turn's once-per-turn window. Catalog files write it as
`Cost: LoyaltyCost(-3)`.

**Its presence carries the rules, so no card has to remember them.**
`ActivateCatalogAbility` derives four gates from a non-nil
`Loyalty`:

| Rule | Enforcement |
| --- | --- |
| CR 606.1 | source must be a planeswalker → `ErrNotAPlaneswalker` |
| CR 606.2 | activator must control it → the pre-existing `ErrCardCallerMismatch` |
| CR 606.3 | a `−N` needs N counters → `ErrInsufficientLoyalty` |
| CR 606.5 | sorcery speed, and once per turn per planeswalker → `ErrSorcerySpeedRequired`, `ErrLoyaltyAlreadyActivated` |

CR 606.3 is checked in the validate-everything-first block, before
any cost is paid, so a refused activation leaves the loyalty
untouched and does **not** burn the turn's window. Paying down to
exactly zero is legal and is not an overpayment — the 704.5i SBA of
§6 takes it from there, which
`TestLoyaltyAbilityMayPayDownToZero` pins.

**Payment goes through `applyCounterLocked`, not
`AddCounterForEffect`** — the opposite of §5, and deliberately so.
§5 is about a planeswalker *entering*, which is an effect, so
Doubling Season applies. Paying a loyalty cost is a cost payment (CR
121.1 / 606.2), not an effect, so counter-doubling replacements must
**not** apply: Doubling Season really does nothing to a `+1`. Routing
the payment through the CR 614 pipeline would have silently made it
double.

**The once-per-turn gate moved.** `LoyaltyActivatedThisTurn`
(`game.go:141`) and its turn-boundary flush (`game.go:554`) were
already right; they were just consulted only by the sandbox action.
Both paths now set and read it, so a manual `+1` and a catalog `−3`
on the same planeswalker on the same turn correctly conflict.

**ADR 0020's exclusion is reversed**, and this is the note it
deserves. 0020 kept loyalty out of `AbilityCost` on the grounds that
the S13.1 `ActivateLoyalty` action already carried it. That was true
and it was not enough: the action moves a counter and runs no
effect, so the exclusion quietly meant "no planeswalker in this
engine can ever do anything". The reversal costs one nullable field
and buys the entire card type. The field comment on
`AbilityCost.Loyalty` carries the same note for anyone reading the
code instead of the ADRs.

**`ActivateLoyalty` was hardened, not retired.** §7 suggested
retiring it. It stays, because it is the manual path for the
thousand-odd planeswalkers with no catalog entry — the same bargain
manual `tap` strikes for every card the catalog cannot express, and
#329's own Teferi is one of them. What changed is that it is no
longer a counter faucet pointed at the whole table: it now rejects
non-planeswalkers, rejects cards the activator doesn't control, and
enforces CR 606.3 instead of driving loyalty negative. The old
comment arguing for negative loyalty ("so the SBA can see the
intent") is gone: the SBA reads `loyalty <= 0`, so clamping and
refusing are indistinguishable to it, and refusing is what the rules
say.

**Two planeswalkers ship.**

- **Teferi, Time Raveler** (`teferi_time_raveler.go`), the card #334
  is named after. His `−3` is complete — loyalty payment, "up to
  one target artifact, creature, or enchantment" bounce, and the
  draw. His `+1` is registered with its loyalty gain and **no
  effect**, because "until your next turn, you may cast sorcery
  spells as though they had flash" needs an "as though" cast
  permission the gate at `mutations.go:597` has no hook for, and an
  "until your next turn" duration that `turn_scoped_statics.go:97`
  already records as inexpressible. Registering it anyway is a
  judgement call: a planeswalker that can only ever tick *down* is
  a worse lie than one whose plus ability at least protects him, the
  label the player reads is the printed text, and the gates around
  it are real. His static ("each opponent can cast spells only any
  time they could cast a sorcery") needs the same missing machinery
  and is absent.
- **The Wandering Emperor** (`wandering_emperor.go`), complete in
  all three abilities, because S21 had already built every primitive
  she needs — `WhiteSamuraiToken` was added for her `+1` and had sat
  unused ever since, and PR #314's until-end-of-turn statics make
  the `−2` expressible as `BoostUntilEOT` + `GrantKeywordUntilEOT`.
  Her "activate loyalty abilities at instant speed the turn she
  enters" clause is a per-permanent override of CR 606.5 that the
  gate has no hook for, so she is slower than printed on the turn
  she lands.

**Client.** Three fixes, one per symptom.

- `activate_loyalty` joined the `ActionType` union, so the client
  can send the action the server has exposed since S13.1. It is not
  decorative: the card menu offers a manual loyalty section on a
  planeswalker the catalog does not know, with the costs that walker
  can legally pay right now (+2 / +1 / [0], and −1 down to its
  current loyalty), and those rows fire `activate_loyalty`. The
  ungated add / remove-counter rows stay where they were — those are
  the escape hatch for fixing a mistake, and gating them would
  defeat their purpose.
- `canActivateLoyalty` (`timing.ts:209`) had no callers at all. It
  now gates the loyalty rows in the card menu, and its
  once-per-turn check reads the real thing: `CardView` carries
  `loyalty_activated`, stamped from `LoyaltyActivatedThisTurn`. The
  `alreadyActivated` argument survives as an override for an
  optimistic local update and ORs with the server's flag, so a stale
  `false` can never re-open a row the server has closed.
- **Clicking a planeswalker no longer taps it.** `PlayerPanel`'s
  click router fell through every branch to `onTapToggle`;
  `battlefieldClickIntent` (`contextMenu.logic.ts`) now sends a
  planeswalker to the ADR 0028 card menu instead, whether or not it
  has catalog abilities, because even a bare walker's menu carries
  the loyalty rows. Tap and untap remain available there for the
  rare effect that wants them.

**Two surgical fixes fell out.** `abilityLegalTargets`
(`view.go:1921`) never copied a clause's `Min` / `Max` onto the wire,
so every activated ability shipped as `0..0` — "unbounded, confirm
with nothing picked" — and `targeting.ts`'s `beginForAbility`
compensated by hard-coding `1 / 1`, with a comment naming this exact
fix. Teferi's `−3` is the first clause whose `Min` is genuinely 0, so
the compensation had to go: the count is stamped server-side and read
client-side, and `isMultiPick` now treats `min < 1` as "needs a Done
button" so "up to one target" can be answered with none.

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
  *(§8 is where the client work landed: the click router, the menu's
  loyalty rows, and the `Min` / `Max` the ability picker had been
  faking.)*

## What this deliberately does not do

Each of these is S27 ([#92](https://github.com/krakenhavoc/cmd_and_ctrl/issues/92))
scope and none is made harder by this change.

- ~~**Loyalty abilities that do anything.**~~ **Shipped** on
  `feat/loyalty-abilities` — §8. The predicted shortest path was the
  path taken, with one correction: `ActivateLoyalty` was hardened
  rather than retired, because it is still the only way to drive the
  planeswalkers the catalog has never heard of.
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
  action it gates is not in the client's action union at all. *As of
  §8 the tick is finally earned: the predicate gates the card menu's
  loyalty rows and `activate_loyalty` is in the union.*
