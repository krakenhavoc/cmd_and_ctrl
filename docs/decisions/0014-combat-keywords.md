# ADR 0014 — Combat keywords (S18)

**Status:** Implemented · 2026-04-23 (planned + shipped) · Sprint S18

## Context

S16 shipped the CR 613 layer engine. S17 shipped the CR 614
replacement-effect engine. The sandbox grants keyword strings via
Layer 6 — Lord of Atlantis appends `"flying"`/`"islandwalk"` to
`Characteristic.Abilities`, and the wire projection
(`CardView.Abilities`) already exposes them. What the engine still
does not do is *consume* those keywords: blocked attackers deal
full power to the first blocker in slice order regardless of
flying/menace/deathtouch; lifelink doesn't gain life; trample
doesn't carry; vigilance doesn't untap; summoning sickness isn't
tracked at all, so haste has nothing to bypass; flash isn't gated.

S18 wires combat keywords into the engine. Twelve core keywords go
live (flying, reach, deathtouch, lifelink, trample, vigilance,
first strike, double strike, menace, defender, haste, flash) plus
summoning sickness, a proper two-substep combat damage flow, and a
multi-blocker damage-assignment prompt (CR 510.1c). Twelve catalog
cards ship with each keyword exercised at least once. Niche keyword
families (protection, indestructible, hexproof, ward, shroud, and
the legacy/evergreen tail — banding, rampage, flanking, fear,
intimidate, shadow, exalted, annihilator, persist, undying,
tribute, prowess, cascade) are explicitly deferred — tracked on a
dedicated umbrella issue so nothing leaks between sprints.

The S16 #157 hotfix ("blocked attackers deal full power to the
first blocker; blockers deal back; SBA fires after combat") is
replaced by the proper flow here. The sandbox note on #68
acknowledged the hotfix as a temporary stand-in.

## Decisions

### 1. Keywords are strings in `Characteristic.Abilities`, read via one helper

Keywords are already a solved representation problem — Lord of
Atlantis demonstrated in S16 that a Layer 6 `StaticAbility` can
append a string to `Characteristic.Abilities` and everything from
the layer engine to the wire projection picks it up. S18 does not
introduce a richer "keyword ability" type.

What's new is the consumer side. S18 adds
[server/internal/game/keywords.go](../server/internal/game/keywords.go)
with four readers:

```go
func HasKeyword(c *Card, kw string) bool
func HasSummoningSickness(c *Card) bool
func CanBlock(attacker, blocker *Card) bool
func BlockerCountValid(attacker *Card, blockers []*Card) bool
```

Every combat-side consumer in the engine reads through these —
`DeclareAttacker`, `DeclareBlocker`, `resolveCombatDamageLocked`,
the cast-legality gate for flash. Inline `for _, a := range c.Effective().Abilities { if a == "flying" ... }`
loops are banned; they drift as soon as one caller forgets to
normalise case or spelling.

**Why a function, not a method on `Card`:** pointer-receiver methods
would force `Card` to grow combat-specific API surface and pull the
layer engine's `Effective()` assumption into every caller. A bare
function in a dedicated file localises the coupling.

### 2. Summoning sickness: `Card.SummonedThisTurn bool`, cleared at controller's untap step

Add a single bool to `Card`. Set true in the battlefield-entry
paths (the same sites S17 routes through for
enters-tapped/enters-with-counters); cleared at `StepUntap` entry
inside `runStepEntryHooksLocked` by looping the active player's
battlefield. Haste bypasses at *read time* in `HasSummoningSickness`:

```go
func HasSummoningSickness(c *Card) bool {
    return c.SummonedThisTurn && !HasKeyword(c, "haste")
}
```

**Why read-time bypass, not clear-on-ETB-with-haste:** control
changes (S24) will reset `SummonedThisTurn` for creatures that
change controllers across a turn boundary even if they already had
haste. Keeping the flag pure makes that future work simpler.

**Gates:** `DeclareAttacker` rejects sick creatures; tap-cost
activation of creature abilities (e.g. mana abilities on a
creature) rejects sick creatures. Land tap is exempt per CR 302.6
(not a creature).

### 3. Combat damage rewrite: split into two substeps (CR 510.2 + 510.3)

Today `resolveCombatDamageLocked` is one pass. S18 rewrites it as:

1. **First-strike substep** — creatures with `first strike` or
   `double strike` assign and deal damage simultaneously. SBA fires
   between substeps so dead creatures exit before the regular
   substep.
2. **Regular substep** — creatures with `double strike` or
   *without* `first strike` assign and deal damage simultaneously.
   SBA fires after.

This is the CR 510 flow. `collectFirstStrikeCombatants` and
`collectRegularCombatants` read live battlefield state each call
so creatures killed in the first substep don't participate in the
second.

### 4. Damage assignment prompt — new `damage_assignment` PendingChoiceKind

CR 510.1c: the attacker's controller divides combat damage among
blockers, assigning at least lethal to each before assigning any
to the next. S18 adds a fifth `PendingChoiceKind` alongside S13's
`discard_from_hand` / S15's `mana_pick` / S17's
`replacement_order` / `optional_replacement`.

**Single-stage prompt** (versus two-stage). The CR 509.2 blocker
ordering step is merged into the same prompt: the client can
reorder blockers *and* assign amounts in one panel; server
validates the ordered prefix-lethal rule atomically.

**Why one stage:** two-stage doubles the client round-trips
without adding any info the server needs earlier. If a user
reports friction ("I wanted to decide the order first"), the
prompt can be split later — the wire payload is extensible. The
existing three PendingChoice kinds have proven robust at one
stage.

Trample is an extra parameter in the assignment payload —
`trampleToPlayer` — which the server accepts only if the attacker
has trample AND every blocker in the order got at-least-lethal.

### 5. Deathtouch via `Card.MarkedLethalByDeathtouch`, not damage-count short-circuit

CR 702.2c: a creature with deathtouch that deals any non-zero
damage to a creature causes that creature to be destroyed at the
next SBA. Two implementation paths:

- **A — short-circuit** `DamageMarked >= Toughness` in the SBA to
  `DamageMarked >= 1 || DeathtouchHit`.
- **B — flag** `Card.MarkedLethalByDeathtouch bool`, set on damage
  from a deathtouch source; SBA treats as lethal.

**Chose B.** Cleaner intent: "this creature is marked for death"
is an explicit state, not a recomputation from damage events. Also
simpler cleanup: the flag zeroes at `StepCleanup` alongside
`DamageMarked`. A short-circuit splits lethal logic across two
predicates; a flag keeps it in one.

### 6. Lifelink applies universally to the damage source, not just combat

CR 702.15: "Damage dealt by a source with lifelink causes its
controller to gain that much life." Covers *all* damage — combat,
burn, triggered-ability, everything. S18 routes lifelink through
the same code path that marks damage, so
`DealDamageToCreatureForEffect` and `DealDamageToPlayerForEffect`
(S17 catalog-effect helpers) also credit life. This is important:
Vampire Nighthawk's combat lifelink and a hypothetical "deals 2
damage to target creature" lifelink source should both gain life.

The life-gain happens after the mark, in the same locked call. No
replacement-pipeline queue needed — life gain is not itself
replaceable by S18 cards (S17's `RepEventLife` replacements still
see the +N, they just don't distinguish lifelink-sourced).

### 7. Trample overflow = `power - sum(blocker.lethal_remaining)` when all blockers at-least-lethal

Standard CR 702.19b math. Single-blocker case:
`overflow = attacker.Power - max(0, blocker.CurrentToughness - blocker.DamageMarked)`.
Multi-blocker case: the `damage_assignment` prompt accepts a
`trampleToPlayer` parameter; the server validates that total
assigned damage equals attacker power AND every blocker got at
least their lethal remaining before spillover.

Without trample, multi-blocker assignments must sum to exactly
`attacker.Power` across the blockers. The prompt payload's
`allow_trample` flag lets the client hide the
"trample to player" input when trample isn't granted.

### 8. Flash needs a keyword reader that works on cards *in hand*

`card.Effective()` only maintains layer-engine output for
battlefield cards — the layer engine runs over the battlefield
slice. A creature in hand has no `Effective()` cache, so a naive
`HasKeyword(card)` would miss the printed "Flash" on an Ambush
Viper sitting in hand.

Two paths:

- **A** — generalise `Effective()` to fall back to
  `printedCharacteristic()` for non-battlefield cards. Feasible,
  but pulls the layer system into zones it wasn't built for.
- **B** — add a dedicated `Spec.PrintedKeywords []string` slot.
  `HasKeyword` branches: battlefield → `Effective().Abilities`;
  anywhere else → `CatalogPrintedKeywords(card.OracleID)`.

**Chose B.** Keywords are a special-enough consumer of ability
data (they're the only thing we look up on cards off-battlefield
today) that generalising `Effective()` is over-broad. The
`PrintedKeywords` slot feeds two places:

1. The battlefield path via a layer-6 `StaticAbility` auto-generated
   at catalog load time (`AppliesTo: target == source`,
   `Apply: append ... PrintedKeywords`). One per catalog card with
   keywords; deduped against any hand-written `Spec.Static`.
2. The off-battlefield path via direct
   `CatalogPrintedKeywords(oracleID)` lookup inside `HasKeyword`.

This keeps flash-gating server-side without wire-side changes.
"Flash" badges in the hand view are a follow-up; the server-side
fix is sufficient for S18's exit criteria.

### 9. Menace enforced at `DeclareBlocker` close-out, not per-decl

Menace requires at least two blockers total (or no blockers).
Checking it on each `declare_blockers` action would false-trigger
on the first of a planned two-blocker block.

S18 enforces menace when the declare-blockers step closes (the
transition to `StepCombatDamage`), via a validator at the top of
`assignAndDealCombatDamageLocked`. If exactly one blocker is
assigned to a menace attacker, clear that blocker's `BlockingTarget`
(blocker effectively "didn't block") and emit a server-side
`slog.Warn` for diagnostics. The attacker becomes unblocked.

**Why not reject at declare time:** asymmetric to the player's
mental model — they're halfway through choosing blockers and get
a red error. "Close-out validates the final state" matches CR
509's step semantics.

### 10. Champion of Lambholt deferred to S19

#68 originally listed "Champion of Lambholt block-restriction
clause" as an S18 task (the counter half is a triggered ability,
which lives in S19). Shipping only the block-restriction half
ships a card with no in-engine way to pump itself — players would
have to hand-feed counters via the S17 shift-click debug
affordance. **S19 ships both halves together.** Memory note on
#69 records the carry-in.

### 11. Baneslayer Angel ships without protection clauses

Protection from Demons / Dragons is CR 702.16; it's scoped out of
S18 (S24 umbrella with Mind Control's control-change
replacement). Baneslayer ships with flying + first strike + lifelink
— three of S18's four most-used keywords in one card. The absent
clauses are noted in the card file and the #68 decision log.

*Update 2026-09-16:* S24 shipped Mind Control (#392) without
protection, and S30 (#95) closed without it too. Protection is now
tracked in [#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662),
and Baneslayer Angel declares the missing clauses in its `Caveats`.

### 12. No protocol version bump

`PendingChoiceView.DamageAssignment *DamageAssignmentView` is a
new optional field alongside S17's `ReplacementOptions`. The
`damage_assignment` kind is a new discriminant value. Pre-S18
clients ignore the field and don't render the prompt; the server
doesn't queue it until an S18-capable block happens. Additive.

## Out of scope (explicit deferrals)

- **Protection** (CR 702.16) → **S24** [#76](https://github.com/krakenhavoc/cmd_and_ctrl/issues/76) alongside Mind Control. *(Update 2026-09-16: not delivered by S24 or S30; tracked in [#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662).)*
- **Indestructible** (CR 702.12), **damage-prevention shields with charges** (CR 615) → **S30** [#95](https://github.com/krakenhavoc/cmd_and_ctrl/issues/95). *(Update 2026-09-16: both shipped. Indestructible in [#380](https://github.com/krakenhavoc/cmd_and_ctrl/pull/380) (S25), `server/internal/game/indestructible.go`; charged prevention shields in [#420](https://github.com/krakenhavoc/cmd_and_ctrl/pull/420) (S30), `server/internal/cards/effects/prevention.go`.)*
- **Hexproof, shroud, ward** — umbrella issue below. *(Update 2026-09-16: all three shipped. Hexproof and shroud are enforced at targeting by [#353](https://github.com/krakenhavoc/cmd_and_ctrl/pull/353), [ADR 0038](0038-protection-style-keywords.md). Ward is a triggered pay-or-counter, `effects.Ward` in `server/internal/cards/effects/ward.go`, from [#421](https://github.com/krakenhavoc/cmd_and_ctrl/pull/421) (recovered onto `main` by [#433](https://github.com/krakenhavoc/cmd_and_ctrl/pull/433)); [#647](https://github.com/krakenhavoc/cmd_and_ctrl/pull/647) added life and sacrifice ward costs.)*
- **Legacy/evergreen tail** — banding, rampage, flanking, fear, intimidate, shadow, exalted, annihilator, persist, undying, tribute, prowess, cascade. Umbrella issue below. *(Update 2026-09-16: cascade shipped in [#426](https://github.com/krakenhavoc/cmd_and_ctrl/pull/426) (S28, recovered by [#433](https://github.com/krakenhavoc/cmd_and_ctrl/pull/433)), `server/internal/game/cascade.go`. The umbrella, [#176](https://github.com/krakenhavoc/cmd_and_ctrl/issues/176), closed on 2026-09-16, and what it left open moved: prowess to [#706](https://github.com/krakenhavoc/cmd_and_ctrl/issues/706), landwalk (never on this list, but granted inert by Lord of Atlantis) to [#705](https://github.com/krakenhavoc/cmd_and_ctrl/issues/705), regeneration to [#667](https://github.com/krakenhavoc/cmd_and_ctrl/issues/667), protection to [#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662). Banding, rampage, flanking, fear, intimidate, shadow, exalted, annihilator, persist and undying as keywords, and tribute have no tracker and get built on demand, when a card needs one. So do phasing and totem armor, which the coverage roadmap had also filed under #176; infect (with wither and toxic) is tracked in [#748](https://github.com/krakenhavoc/cmd_and_ctrl/issues/748) since 2026-09-17.)*
- **Champion of Lambholt** (both halves) → **S19** [#69](https://github.com/krakenhavoc/cmd_and_ctrl/issues/69).
- **Flash-in-hand badge** — server gating is sufficient for S18; hand-view keyword surface follows with S20 smart-cast UI.
- **Mycosynth Lattice mana-ability clauses** — originally slotted here per S17 planning. On review, the clauses ("lands tap for any color" + "no land's mana ability adds non-colorless") are mana-system work that predates the S15 mana-pool auto-tapper's final shape. Re-homed to the backlog pending a dedicated mana rewrite sprint; does not block S18 combat work.

## Consequences (as shipped 2026-04-23)

- **12 catalog entries ship**: eight vanilla-keyword creatures
  (Serra Angel, Colossal Dreadmaw, Giant Spider, Typhoid Rats,
  Youthful Knight, Fencing Ace, Lightning Elemental, Wall of
  Stone) + four multi-keyword creatures (Vampire Nighthawk,
  Baneslayer Angel, Ambush Viper, Boggart Brute). Boggart Brute
  swapped in for Dreg Mangler from the planned list after
  oracle-text review showed Dreg Mangler carries scavenge+haste,
  not menace.
- **Eighth catalog hook** `CatalogPrintedKeywords` joins the
  S14/S15/S16/S17 seven. Accessible via the function-var slot in
  `effect_hooks.go`; populated by `wire.go` at init time.
- **`Spec.PrintedKeywords []string`** new field on the catalog
  spec. Feeds two consumers: on-battlefield via a synthesized
  self-only Layer 6 `StaticAbility` generated in `wire.go` (so
  the keyword lands in `Characteristic.Abilities` alongside
  hand-written statics); off-battlefield via direct
  `CatalogPrintedKeywords` lookup inside `HasKeyword` (required
  for flash gating on a hand-resident Ambush Viper).
- **`server/internal/game/keywords.go`** new file with four
  helpers: `HasKeyword`, `HasSummoningSickness`, `CanBlock`,
  `BlockerCountValid`. Single canonical reader surface — inline
  `for _, a := range Effective().Abilities` loops are banned in
  combat code.
- **`Card.SummonedThisTurn bool`** new field set on every
  battlefield entry by the layer listener (where
  `EnteredBattlefieldAt` is stamped); cleared in
  `untapAllForLocked` at the start of the controller's untap
  step. Haste is a read-time bypass in `HasSummoningSickness`,
  not a clear-on-ETB.
- **`Card.MarkedLethalByDeathtouch bool`** new field flagged by
  damage from a deathtouch source; read by the lethal-damage
  SBA (destroys regardless of toughness, per CR 702.2c); cleared
  at `StepCleanup` alongside `DamageMarked`.
- **`DeclareAttacker` gates on summoning sickness** (returns
  `ErrSummoningSick`) and defender (returns `ErrDefender`); calls
  `RecomputeLayersIfStaleLocked` first so mid-turn haste grants
  take effect immediately.
- **`resolveCombatDamageLocked` rewritten** into two substeps
  (CR 510.2 first-strike + CR 510.3 regular). SBA runs between;
  dead creatures don't participate in the second pass. Double
  strike participates in both; first strike only the first.
- **`assignAndDealCombatDamageLocked`** is the per-substep
  assigner. Unblocked → player; single blocker → full power
  with trample overflow when blocker at-least-lethal;
  multi-blocker → queue `PendingChoiceDamageAssignment` prompt.
- **Menace close-out** (CR 702.111): a single blocker against a
  menace attacker is silently reverted (clear
  `BlockingTarget`); attacker becomes unblocked.
- **`PendingChoiceDamageAssignment`** new kind + resume path.
  `DamageAssignmentFrame` carries attacker ID, blocker IDs,
  attacker power, trample/deathtouch/first-strike flags, AND
  cached `SourceLifelink` + `SourceController` so lifelink still
  fires if the attacker died to blocker damage before the
  prompt resolves.
- **`ResolveDamageAssignment`** validates permutation of
  blocker IDs, sum == attacker power, prefix-lethal-in-order
  (relaxed to 1 under deathtouch), trample-only-if-trample.
- **`markCombatDamageFromFrameLocked` + `markCombatDamageToPlayerFromFrameLocked`**
  are the damage routers for the resume path — sources keyword
  state from the frame, not the (possibly dead) attacker card.
- **Lifelink applies universally** (CR 702.15): creature-damage
  AND player-damage pipelines credit the source's controller.
  Fires via `applyLifelinkLocked` in the direct pipeline and
  `applyLifelinkFromFrameLocked` in the resume path.
- **Wire**: `PendingChoiceView.DamageAssignment *DamageAssignmentView`
  (`attacker_card_id`, `blocker_card_ids`, `attacker_power`,
  `allow_trample`, `has_deathtouch`) — additive, no version bump.
- **`actions/actions.go` `TypeResolveChoice` dispatcher leg**
  for `damage_assignment`: routes on `Assignments` / `TrampleToPlayer`.
- **Client `ChoicePromptModal.svelte` `damage_assignment` branch**
  renders reorder buttons (▲/▼) + per-blocker damage inputs +
  trample-to-player input when `allow_trample`. Submits
  `{assignments:[{blocker_id, amount}], trample_to_player}`.
- **Client `KeywordBadgeRow.svelte`** new component — bottom-edge
  row of abbreviated keyword badges (FLY / RCH / FS / DS / DT /
  LL / TR / VIG / MEN / DEF / HST / FLS) with full-word
  tooltips. Unknown keywords (e.g. Lord of Atlantis's
  `islandwalk`) fall back to a 3-char truncated label.
- **`CardView.abilities` added on the client** type (wire field
  was already present server-side since S16 sub-PR 4).
- **S19 inherits**: Champion of Lambholt both halves; the S18
  machinery (HasKeyword, two-substep combat, damage-assignment
  prompt) is the substrate.
- **S24 inherits**: protection (CR 702.16) — Baneslayer Angel
  ships here without protection clauses. *(Update 2026-09-16: S24
  did not deliver it; tracked in #662.)*
- **S30 inherits**: indestructible (CR 702.12) — reads the SBA
  branch from the opposite side of `MarkedLethalByDeathtouch`.
  Damage-prevention shields with charges (CR 615) continue to
  build on the S17 hook + turn-scoped pattern.
- **No protocol version bump** — every wire addition is
  additive.

## Decision log (planning round, 2026-04-23)

1. Keywords represented as strings in `Characteristic.Abilities`; no richer keyword type.
2. `Spec.PrintedKeywords` separate from `Spec.Static` so flash works on hand cards.
3. Summoning sickness is a per-card bool cleared at controller's untap step; haste bypass is read-time.
4. Combat damage rewritten to two substeps (510.2 + 510.3); double-strike participates in both.
5. Deathtouch implemented via `Card.MarkedLethalByDeathtouch`, not SBA short-circuit.
6. Lifelink applies to all damage from the source, not combat damage only.
7. Damage-assignment prompt single-stage (order + amounts merged).
8. Menace enforced at declare-blockers step close-out, not on each declare action.
9. Champion of Lambholt deferred to S19 (both halves together).
10. Baneslayer Angel ships without protection clauses.
11. Mycosynth Lattice mana-ability clauses dropped from S18 scope (re-homed to backlog).
12. No protocol version bump; `damage_assignment` PendingChoice + view field additive.
