# ADR 0036 — Attachments: Equipment and Auras

**Status:** Proposed · 2026-09-11 · S32 spike ([#280](https://github.com/krakenhavoc/cmd_and_ctrl/issues/280)), implementation S33 · Designs [#76](https://github.com/krakenhavoc/cmd_and_ctrl/issues/76) (S24)

## Context

There is no attachment relation in the engine. An Equipment is an
artifact that sits on the battlefield doing nothing, and an Aura is an
enchantment that resolves to the battlefield attached to nothing. Both
card types are structurally inert, and both are load-bearing in the
format:

- The top-100 Commander staples triage ranks **Swiftfoot Boots (12),
  Lightning Greaves (13) and Skullclamp (40)** — three of the top 40 —
  and calls equipment "the largest single subsystem in this list"
  (`docs/decklists/top-100-commander-staples.md:411-420`). Its build
  order puts attachment fourth, after three items that have now shipped
  (`:500-515`).
- **ADR 0033 §7** (`docs/decisions/0033-ai-bot-seat.md:263-268`) states
  plainly that of #89's six proposed bot archetypes, "Voltron and any
  Equipment or Aura strategy are unbuildable until S24 lands the
  attachment layer". Three of six are buildable today.
- The Aang decklist triage records Sword of Hearth and Home as the
  deck's only Equipment and The Mighty Thor's "whenever an Equipment
  you control enters" clause as "mostly cosmetic here"
  (`docs/decklists/aang-is-so-flashy.md:295-299`).
- **ADR 0012** deferred Mind Control's control change to S17, which
  deferred it again to S24 on attachment grounds; the layer listener
  still carries the note (`server/internal/game/layer_listener.go:29-32`).

The good news the triage did not spell out: almost none of this is new
machinery. Activated abilities with costs, sorcery-speed gating,
structured targeting with announce + resolution re-checks, the layer
engine, the SBA loop, the right-click ability menu and the legal-move
enumerator all exist and all fit. What is missing is one field, one
state-based action, one wire key and one rendering rule.

This ADR is a design spike. **No code is written here.** Every
structural claim below was checked against the tree at `origin/main`
`f26c961` and is cited with `file:line`.

## Decisions

### 1. The relation is `Card.AttachedTo TargetRef`, stored on the attached object

One field on `game.Card`, pointing *from* the Equipment or Aura *to*
its host. `TargetRef` (`server/internal/game/stack.go:70-73`) is the
existing two-field value struct — `{Kind TargetRefKind, ID uuid.UUID}`
— and its zero value (`Kind == ""`) is the "unattached" sentinel.

**Why the attached object holds the pointer, not the host.** An
Equipment or Aura is attached to at most one thing (CR 301.5c,
CR 303.4d), so one field suffices. A host may carry any number of
attachments, which on the host would be a slice — and a slice needs an
explicit deep-copy branch in `cloneCard`
(`server/internal/game/clone.go:197-231`), which is exactly the class
of thing this codebase has to remember to do by hand.

**Why `TargetRef` and not `uuid.UUID`.** Curses (Curse of Opulence,
Curse of Vengeance) enchant *players*, not permanents. A bare
`uuid.UUID` cannot say which, because player IDs and card instance IDs
are both UUIDs. `TargetRef` already encodes exactly that distinction
(`TargetPlayer` / `TargetCard`, `stack.go:49-64`), it is already what
the targeting pipeline produces, and it is already on the wire as
`TargetRefView` (`server/internal/protocol/view.go:376-379`, mirrored
at `client/src/lib/protocol.ts:353`). The attach path assigns
`item.Targets[0]` verbatim — **no conversion at all**.

**Why not `*uuid.UUID`,** as [#280](https://github.com/krakenhavoc/cmd_and_ctrl/issues/280) §1 proposes.
`cloneCard` (`clone.go:197`) begins `out := c` — a plain value copy —
and then deep-copies the four slice fields and two maps individually. A
pointer field would be **aliased** between the live game and every
snapshot on the undo stack. That aliasing is harmless *only* under the
convention "always assign a fresh pointer, never mutate through one",
which `clone.go` cannot enforce and which no test guards
(`clone_test.go` has no reflection-based field-count check). A value
field is safe by construction. The same struct already carries three
`uuid.Nil`-sentinel relations for the same reason — `AttackingTarget`
(`card.go:162`), `BlockingTarget` (`:168`), `GoadedBy` (`:176`).

**Why not a side table on `Game`.** A `map[uuid.UUID]uuid.UUID` on
`Game` would need its own branch in `cloneLocked`, its own line in
`RestoreFrom` (`clone.go:374-414`), its own invariant against every
zone move, and it would not be swept by `MoveCard`'s CR 400.7 cleanup
(`zone.go:149-157`). The field rides the card through every zone change
for free.

**Consequence for `clone.go`.** #280 §1 says "`clone.go` must carry
it." It must, and with a value-typed field **it already does** — `out
:= c` at `clone.go:198` copies it, and `RestoreFrom` adopts
`src.Battlefield` wholesale (`clone.go:380`). Zero new lines. Because
nothing structurally enforces that, S33 must add an explicit
attachment case to `clone_test.go` to pin it.

**Reverse lookup** is a linear scan of `g.Battlefield.Cards`, the same
shape as `findCardOnBattlefield` (`layers.go:244-254`). The battlefield
is tens of cards; this is not a cost worth a second data structure.

### 2. A second field: `Card.AttachedAt int64` (CR 613.7d)

CR 613.7d gives an Equipment's or Aura's continuous effect a **new
timestamp** when it becomes attached. Today `activeStaticAbilitiesLocked`
binds every static's timestamp to `src.EnteredBattlefieldAt`
(`layers.go:165`). The fix is one line at that site: prefer
`src.AttachedAt` when non-zero.

Honest accounting: this buys **nothing** for the first ten cards. Every
attachment static in scope is either a Layer 7c modify or a Layer 6
ability grant, and both are commutative — order is unobservable. It
becomes observable only when a 7b *set* meets a 7c *modify*, which no
in-scope card does. This is one `int64` (free through `cloneCard`'s
value copy) and one line, and it removes a class of latent wrongness;
but it is the first thing to cut if S33 runs long.

### 3. Equipment and Auras share the relation and almost nothing else

They differ in four ways, and flattening any of them would be a bug.

**(a) How attachment happens.** Equip is a CR 702.6 activated ability
that resolves off the stack and moves a permanent already on the
battlefield. An Aura is cast targeting (CR 303.4a) and attaches as it
enters. Different code paths entirely — decisions 4 and 5.

**(b) The restriction.** Equip targets "creature **you control**"
(CR 702.6b) — but only *at activation*. Once attached, control of the
creature may change and the Equipment stays put; only "is it still a
creature" matters afterwards (CR 301.5c). An Aura's "enchant" clause is
both a cast restriction and an ongoing legality condition, and it says
nothing about control: Pacifism legally enchants an opponent's
creature and **does not change control of it**. Mind Control-style
control change is a Layer 2 effect and is explicitly **out of scope**
(see §"What this deliberately does not do").

**(c) What happens when the host becomes illegal.** This is the reason
they cannot be one function. Equipment becomes unattached and **stays
on the battlefield** (CR 704.5m). An Aura is put into its owner's
**graveyard** (CR 704.5n). Same trigger condition, opposite outcome.

**(d) Re-attachment.** Equip may be activated again and moves the
Equipment (CR 702.6d) — the second equip is an ordinary activation that
overwrites `AttachedTo`. An Aura on the battlefield never moves itself.

Net: one relation, one SBA site with two branches, two attach paths.

### 4. Equip is an ordinary `Spec.Activated` entry — no new engine verb, no new UI

```go
Activated: []ActivatedAbility{{
    Label:        "Equip {1}",
    Cost:         game.AbilityCost{Mana: "{1}"},
    Targets:      TargetCreatureYouControl(),
    SorcerySpeed: true,
    Effect:       AttachSourceToTarget,
}},
```

Every piece already exists and was checked:

- `ActivatedAbilityShape.SorcerySpeed` (`activated.go:78-79`) is gated
  by `sorcerySpeedOpenLocked` (`mutations.go:1122-1133`): main phase,
  empty stack, caller is the active player. That is CR 702.6b's "any
  time you could cast a sorcery", and the rejection is
  `ErrSorcerySpeedRequired` (`activated.go:166-168`).
- `ActivatedAbilityShape.Targets` (`activated.go:74-76`) is validated
  at announce by `validateTargetsLocked` (`targets.go:281-303`,
  called from `activated.go:193-199`) and re-checked at resolution via
  `TargetStillLegalForEffect` (`targets.go:170-181`). An equip whose
  target dies in response therefore fizzles correctly with no new code.
- The wire already carries what the client needs:
  `ActivatedAbilityView.SorcerySpeed` and `.LegalTargets`
  (`view.go:716`, `:726`).
- The **context menu already lists it**. `abilityItems`
  (`contextMenu.logic.ts:335-344`) iterates every
  `card.activated_abilities` entry; `handleActivateAbility` →
  `continueActivation` → `beginTargetingForAbility`
  (`Board.svelte:499-507`, `:553-571`) already runs the target picker
  for an ability that declares `legal_targets`, and `fireTargets`
  (`Board.svelte:453-466`) already sends `activate_ability` with them.
  This is ADR 0028 decision 5 paying off exactly as designed: equip
  lands in the right-click menu with **zero new client wiring**.
- The **legal-move enumerator picks it up for free**. `activatedMoves`
  (`server/internal/legal/abilities.go:27-100`) is generic over
  `game.ActivatedAbilitiesForCard`, already skips sorcery-speed
  abilities outside the window (`:40-42`), already checks mana
  affordability (`:54-59`) and already expands target sets (`:69-75`).
  No enumerator change at all.

**The one client gap.** `abilityBlocked`
(`contextMenu.logic.ts:303-314`) greys a row for tap / sacrifice /
no-legal-target, but has **no sorcery-speed clause** — equip would be
the catalog's first sorcery-speed activated ability (verified: no
`SorcerySpeed: true` in `server/internal/cards/effects/` today). The
server rejects correctly regardless, so this is cosmetic: one clause
reading `a.sorcery_speed` against the existing timing helper in
`client/src/lib/timing.ts`.

`Effect` is a small helper in the `effects` package that writes
`AttachedTo` on the source and emits `EventAttach`. It must re-read the
source off the battlefield by ID rather than capturing a `*Card` — the
standing rule for stack-item effects (`activated.go:81-84`).

### 5. Auras attach at resolution, from the cast target

An Aura is cast targeting (CR 303.4a). The target is already on
`StackItem.Targets`, and it is already re-checked at CR 608.2b by
`spellAllTargetsIllegalLocked` (`mutations.go:1218`, `:1321`) — an Aura
whose only target went away is countered by game rules and routed to
the graveyard, which is correct with no new code.

The attach stamp goes in the loop that already stamps `Controller` and
`EntersTapped` after the permanent lands
(`mutations.go:1274-1282`). That is the only place in
`resolveTopOfStackLocked` holding both the moved card's battlefield
index and the `item`.

**Keyed on the card type, not on a catalog opt-in.** If the resolving
permanent's `Effective().Subtypes` contains `"Aura"` and the item has
exactly one `TargetCard`/`TargetPlayer` slot, stamp it. Attachment is a
rule of the card *type*, not a property of an individual card, and
keying on the subtype means a non-catalog Aura played in the sandbox
also attaches to whatever its player picked. The alternative — a
`Spec.Enchant *TargetSpec` opt-in — buys nothing the card's existing
`Spec.Targets` clause does not already say.

This needs a `Card.HasSubtype(string)` helper in `card.go`; there is
none today (the three existing call sites each re-loop over
`Effective().Subtypes` by hand — `lord_of_atlantis.go:78`,
`krenko_mob_boss.go:54`).

### 6. "Equipped/enchanted creature" statics are ordinary catalog statics

The existing layer engine needs no change. A card declares:

```go
Static: []game.StaticAbility{{
    Layer: game.Layer7PT, SubLayer: game.SubLayer7C_Modify,
    AppliesTo: AttachedToSource,     // new shared predicate
    Apply:     func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
        c.Power++; c.Toughness--     // Skullclamp
    },
}},
```

`AttachedToSource` is `target.AttachedTo.Kind == TargetCard &&
source.AttachedTo.ID == target.InstanceID` — the exact shape of the
existing `selfOnly` predicate (`wire.go:218-220`), which is the
precedent for a one-card-scoped static. `activeStaticAbilitiesLocked`
(`layers.go:150-170`) finds it through the normal
`CatalogStaticAbilities(src.OracleID)` walk with no modification;
`applyLayerLocked` (`layers.go:217-239`) evaluates `AppliesTo` per
battlefield card per recompute, so the predicate re-reads current
attachment every pass.

Layer 6 grants work the same way and **the keyword actually lands**:
`HasKeyword` reads `c.effective.Abilities` (`keywords.go:56-63`) and
`HasSummoningSickness` reads `HasKeyword(c, "haste")`
(`keywords.go:100-108`), so Lightning Greaves' haste half is live the
moment the static applies. See decision 10 for the shroud half.

### 7. Invalidation: new `EventAttach` / `EventUnattach` event kinds

Attaching changes a layer input, so `layerVersion` must bump or
`RecomputeLayersIfStaleLocked` (`layers.go:314-326`) fast-paths past
the change. The layer listener's own comment already pre-registers this
case: "Control changes. Mind Control / aura attach is deferred to S17.
When the first 'creatures you control' predicate that can flip
mid-game lands, add an EventControlChanged kind + bump here"
(`layer_listener.go:29-32`).

Emit two new kinds in `events.go` (the enum sits at `:45-297`) and add
them to `layerVersionBump.OnEvent` (`layer_listener.go:52-67`),
alongside the existing `EventCounterPlaced` case. An event rather than
a direct `g.layerVersion.Add(1)` because: attachment has more consumers
than the layer engine ("whenever this becomes attached" triggers), the
trigger harvester dispatches off event kinds, and events are greppable
in the replay log and in the public game log ADR 0033 §4 requires.

### 8. The unattach SBA sits at the top of `stateBasedActionsLocked`, before the destruction pre-pass

```
stateBasedActionsLocked (mutations.go:1761)
  ├─ RecomputeLayersIfStaleLocked            (:1772)
  ├─ NEW: attachment legality (704.5m/n)     ← here
  ├─ counter cancel (704.5q)                 (:1778)
  ├─ player-loss SBAs                        (:1806)
  └─ destruction pre-pass `doomed`           (:1837-1879)
```

**After the recompute** because the legality test must read post-layer
state (decision 9). **Before the destruction pre-pass** so that a
creature which becomes lethally damaged *because* its +2/+2 Aura fell
off is destroyed in the same settling, rather than surviving a pass.

Setting `fired = true` on any unattach keeps `runStateChecksLocked`
(`mutations.go:1905-1919`) looping; the emitted `EventUnattach` bumps
`layerVersion`, so the next iteration's `RecomputeLayersIfStaleLocked`
sees the corrected board. CR 704.3's "simultaneously, then re-check"
is satisfied by the existing loop — no new convergence machinery.

The two branches:

```
for each battlefield card c where c.AttachedTo.Kind != "":
    if attachmentLegal(g, c) { continue }
    emit EventUnattach
    if isAura(c) { routeBattlefieldCardToOwnerGraveyardLocked(c.ID) }  // 704.5n
    else         { c.AttachedTo = TargetRef{} }                        // 704.5m
    fired = true
```

The Aura branch reuses `routeBattlefieldCardToOwnerGraveyardLocked`
(`mutations.go:2003-2052`), which means an Aura falling off goes
through the CR 614 replacement pipeline and the commander-zone built-in
like any other permanent leaving — an enchantment commander that is an
Aura works correctly for free.

### 9. Legality reads `Effective()`, never `Card.IsCreature()`

`Card.IsCreature()` reads `Card.TypeLine` — the *printed* string
(`card.go:348-350`). PR #255's engine finding 1 recorded this
precisely: "Layer 4 type changes are inert — `Card.IsCreature()` /
`IsLand()` read `Card.TypeLine`, not `Effective().Types`."

If the unattach SBA used `IsCreature()`, an Equipment would stay
attached to something that had stopped being a creature. The SBA must
test `Effective().Types` — which is also why it has to run after the
recompute (decision 8).

**Auras reuse their own announce-time predicate.** `attachmentLegal`
for an Aura is `g.targetLegalLocked(c.Controller, TargetSpecFor(c.OracleID),
c.AttachedTo)` (`targets.go:239-274`, `:100-105`). That is *literally
the same function* the cast used at announce (`targets.go:295`) and at
the CR 608.2b re-check (`targets.go:176`) — one source of truth for
"enchant creature", and it already enforces the zone list
(`targets.go:257-266`).

**Equipment cannot reuse its equip spec,** because the equip clause
says "creature **you control**" and the ongoing requirement is only "is
a creature" (decision 3b). Equipment gets the hardcoded CR 301.5c rule:
host is a battlefield card whose `Effective().Types` contains
`"Creature"`. A first-cut Aura must therefore declare an enchant clause
with no control component — Rancor, Pacifism and Curses all read
"enchant creature" / "enchant player", so all are clean.

### 10. Lightning Greaves and Swiftfoot Boots are the honest problem in this batch

Greaves grants **haste and shroud**; Boots grants **haste and
hexproof**. Haste works (decision 6). Shroud and hexproof are both on
#176's deferred list, and nothing in the engine reads them —
`targetLegalLocked` (`targets.go:239-274`) is the single choke point
for target legality and has no keyword gate.

So both cards would ship **strictly stronger than printed**: the
protection half is the *drawback-adjacent* half that stops you
targeting your own creature, and dropping it is pure upside. That
inverts this repo's convention, which #255 states explicitly — "every
other simplification in this repo is strictly weaker."

The awkward part is that these are the format's #1 and #2 equipment
(ranks 12 and 13). **Recommendation: bundle a minimal shroud / hexproof
target gate into the same sprint** rather than declare a no-op half.
It is genuinely small — the keyword already reaches
`Effective().Abilities` through the Layer 6 grant, and the gate is one
clause in the `TargetCard` branch of `targetLegalLocked`
(`targets.go:252-271`): reject when the candidate has `"shroud"`, or
has `"hexproof"` and `caster != c.Controller`. One choke point covers
announce, resolution re-check, the client's `legal_targets` projection
and the bot enumerator simultaneously. Sized in §Sizing as sub-PR 4.

If that is cut, Boots and Greaves must be cut with it, and the first
cut ships Skullclamp, Darksteel Plate and Sword of Hearth and Home.
Darksteel Plate's indestructible is also #176-deferred, but that
omission makes the card *weaker* than printed, which is in convention.

### 11. Skullclamp's dies-trigger needs no LKI extension — the ordering already works

Skullclamp is "Whenever equipped creature dies, draw two cards." The
attachment must still be readable at the moment the creature dies. It
is, and the ordering was checked end to end:

`routeBattlefieldCardToOwnerGraveyardLocked` → `executeBattlefieldLeaveLocked`
→ `snapshotLKILocked` (`mutations.go:2125`) → `MoveCard` (`:2126`) →
`EventLTB` (`mutations.go:2137`). The trigger harvester runs
**synchronously inside `EmitEvent` under the write lock**
(`triggers.go:90-92`), and at that instant Skullclamp is still on the
battlefield with `AttachedTo.ID == ev.CardID` — the unattach SBA has
not run yet, because it runs on the next `stateBasedActionsLocked`
pass. So `AppliesTo` is a direct `source.AttachedTo.ID == ev.CardID`.

This matters because `g.lastKnownBattlefield` stores a
`Characteristic` (`triggers.go:359-366`), and `Characteristic`
(`characteristic.go:31-40`) has no attachment field. Had the ordering
gone the other way, Skullclamp would have required extending LKI.

Sword of Hearth and Home's "whenever equipped creature deals combat
damage to a player" is the precedented `Watches: [EventDealDamage]` +
`ev.Combat` shape Bident of Thassa uses
(`bident_of_thassa.go:23-26`), with the controller check swapped for
`source.AttachedTo.ID == <the damaging creature>`.

### 12. `MoveCard` clears `AttachedTo` on battlefield exit; nothing sweeps the reverse direction

`MoveCard`'s CR 400.7 cleanup (`zone.go:149-157`) already zeroes
`Tapped`, `Counters`, `BattleX/Y`, `AttackingTarget`, `BlockingTarget`,
`GoadedBy` when `src.Kind == ZoneBattlefield`. `AttachedTo` and
`AttachedAt` join that list — one line each. That is the *forward*
direction: an Equipment or Aura that leaves the battlefield stops being
attached.

The *reverse* direction — permanents attached to a host that just left
— is deliberately **not** swept eagerly. `MoveCard` is a package-level
function with two `*Zone` arguments and no `*Game`
(`zone.go:141`), so it cannot see the rest of the battlefield. And it
does not need to: a dangling `AttachedTo` fails `attachmentLegal` on
the very next SBA pass, and `runStateChecksLocked` runs at every
priority boundary. The SBA **is** the sweep. This is why CR 704.5m/n
are written as state-based actions in the first place.

The one consequence to state: between the host's LTB and the next SBA
pass there is a window where `AttachedTo` names a card not on the
battlefield. That window is exactly what decision 11 exploits, and no
snapshot can be taken inside it (SBAs run under the same write lock as
the move).

### 13. Wire: one field out, the reverse list derived on the client

```go
// CardView (view.go:502)
AttachedTo *TargetRefView `json:"attached_to,omitempty"`
```

Stamped in `viewOfCard` (`view.go:1695-1772`) beside the existing
`AttackingTarget` / `BlockingTarget` / `GoadedBy` conversions
(`view.go:1751-1759`). Mirrored in `client/src/lib/protocol.ts:582`
and documented in `docs/protocol.md:364`.

**Not redacted.** `redactCardForViewer` (`view.go:1626-1648`) zeroes
printed characteristics for cards the viewer does not know; attachment
is public battlefield state exactly like `attacking_target`, which also
survives redaction. No change to that function.

**No `attachments[]` on the host,** contra #280 §1. Shipping both
directions puts two representations of one relation on the wire, which
can disagree; and the client already does precisely this kind of
derivation — `PlayerPanel.buckets` is a `$derived.by` that partitions
`controlledCards` on every snapshot (`PlayerPanel.svelte:148-154`).
One scan of `view.battlefield.cards` per render is nothing. Ship one
field; derive the other.

### 14. Client: the host owns the stack, and there is no drag-to-attach

Rendering:

1. `bucketForBattlefield` (`client/src/lib/cardTypes.ts:51-55`) keeps
   its current signature; the filter goes in `PlayerPanel.buckets`
   (`PlayerPanel.svelte:148-154`), which drops any card whose
   `attached_to.kind === "card"` and whose host is on the battlefield.
2. `BattlefieldRow` (`BattlefieldRow.svelte:64-77`) renders each host's
   attachments inside the same `[role="listitem"]` wrapper, offset
   behind the host. The negative-margin overlap technique is already in
   this file for the land strip (`BattlefieldRow.svelte:131-143`).
3. A **Curse** (`attached_to.kind === "player"`) has no host card. First
   cut: it stays in its controller's "enchant / artifact" row with a
   badge naming the enchanted player. Drawing it in the *enchanted*
   player's panel is nicer but makes "where a card is drawn" diverge
   from `card.controller`, which is a new concept for this UI and not
   worth introducing alongside the relation itself.

**No drag-to-attach in the first cut,** and the recommendation is not
"later" but "probably never":

- Drag-release on a battlefield card is already the gesture that sets
  `battle_x` / `battle_y`. A drag-onto-a-card would overload it.
- Equip is a **cost**. It needs the mana / auto-tap flow that
  `continueActivation` (`Board.svelte:553-571`) already runs, including
  `AutoTapPreviewModal`. A drag gesture has nowhere to hang that.
- ADR 0028 decision 5's principle is "one gesture, one menu, no lost
  functionality", and decision 6 deleted the `Shift+click` chord
  specifically to avoid two mental models for one operation. A drag
  path alongside the menu path would re-create exactly that.

### 15. Undo and replay need nothing beyond the field

**Undo:** `Room.Undo` (`server/internal/ws/room.go:180-222`) calls
`Game.RestoreFrom` with a previously captured `Clone`. `cloneZone` →
`cloneCard` value-copies the card struct (`clone.go:183-231`) and
`RestoreFrom` adopts `src.Battlefield` wholesale (`clone.go:380`). A
value-typed `AttachedTo` is carried with **zero new lines**. Pin it
with a test (decision 1).

**Replay:** the replay log is a JSONL stream of fully-serialised
`GameView` snapshots, one line per `Apply`
(`server/internal/ws/room.go:315-356`), served by `GET
/games/{id}/replay` (`server/internal/lobby/http.go:202`, `:645`). It
replays *views*, not actions. Once `CardView.attached_to` exists, every
replayed frame carries attachment correctly and no replay code changes.

Correction to the issue framing: #280 §6 says "shows correctly in the
replay scrubber." There is no replay scrubber — `grep -rn scrub
client/src server/internal` finds only unrelated hits. The replay is a
downloadable JSONL, admin-gated and end-of-game-gated
(`lobby/http.go:651`, `:674`). Nothing to do here either way.

### 16. Coexistence with #279's turn-scoped statics

#279 (`feat/until-end-of-turn`, in flight) adds `Game.TurnScopedStatics`
— floating `StaticAbility` values with a duration, consulted by
`activeStaticAbilitiesLocked` alongside the battlefield walk and
cleared at `StepCleanup` next to `ClearTurnScopedReplacementsLocked`
(`game.go:765`).

**They compose through machinery that already exists.** Both an
attachment static and a UEOT static become a `ContinuousEffect` and
reach the same bucket loop. `applyLayerLocked` filters to the
`(Layer, SubLayer)` bucket, `sort.SliceStable` by `Timestamp()`
(`layers.go:227-229`), then applies in order. Rancor's +2/+0 (7c,
timestamped when it attached) and Giant Growth's +3/+3 (7c, timestamped
when it resolved) are two entries in the 7c bucket applied in
timestamp order — and since both are *modifies*, the order is
unobservable anyway. CR 613.7 is satisfied by the existing sort.

**The branches touch `layers.go` in two non-overlapping places:**

| | #279 | this ADR |
|---|---|---|
| `activeStaticAbilitiesLocked` body (`:150-170`) | appends from a second source | unchanged |
| timestamp binding (`:165`) | unchanged | prefer `AttachedAt` |
| `clone.go` | new `TurnScopedStatics` branch + `RestoreFrom` line | **nothing** |
| cleanup step (`game.go:765`) | clears the list | unchanged |

**Sequencing:** land #279 first, rebase attachments onto it. The
`layers.go` conflict is a two-line textual one, not a semantic one.
`clone.go` is not a conflict at all, because a value-typed `AttachedTo`
needs no branch there — which is a second, independent reason to reject
the pointer form from decision 1.

The one thing to verify jointly once both land: a creature carrying
both an Aura and a UEOT pump keeps the pump after the Aura falls off
mid-turn, and loses the pump at cleanup while keeping the Aura. One
test, in whichever branch lands second.

## Blast radius

Server — `game` package:

| File:line | Change |
|---|---|
| `card.go:176` (after `GoadedBy`) | `AttachedTo TargetRef`, `AttachedAt int64` |
| `card.go:403` (near `IsPermanent`) | `HasSubtype(string) bool` helper |
| `clone.go:197` `cloneCard` | **no change** — value copy carries it; add a `clone_test.go` case |
| `zone.go:149-157` `MoveCard` | clear `AttachedTo` / `AttachedAt` on battlefield exit |
| `events.go:45-297` | `EventAttach`, `EventUnattach` |
| `layer_listener.go:52-67` | bump `layerVersion` on both new kinds |
| `layers.go:165` | timestamp prefers `AttachedAt` |
| `mutations.go:1772` (after the recompute) | the CR 704.5m/n SBA, two branches |
| `mutations.go:1274-1282` | Aura attach stamp on resolution |
| `effect_api.go` (~`:538`, beside `AddCounterForEffect`) | `AttachForEffect` / `UnattachForEffect` |
| `targets.go:252-271` | shroud / hexproof gate (sub-PR 4 — decision 10) |

Server — catalog and protocol:

| File:line | Change |
|---|---|
| `cards/effects/spec.go:296` | no new `Spec` field; equip is `Spec.Activated`, the enchant clause is `Spec.Targets` |
| `cards/effects/wire.go:218` | `AttachedToSource` predicate beside `selfOnly` |
| `cards/effects/` | ~10 new card files |
| `protocol/view.go:502` | `CardView.AttachedTo *TargetRefView` |
| `protocol/view.go:1751-1759` | stamp it in `viewOfCard` |
| `protocol/view.go:1626-1648` | **no change** — attachment is public |
| `legal/abilities.go:27-100` | **no change** — equip enumerates for free |

Client:

| File:line | Change |
|---|---|
| `lib/protocol.ts:582` | `attached_to?: TargetRefView` |
| `lib/components/board/PlayerPanel.svelte:148-154` | filter attached cards out of their own bucket |
| `lib/components/board/BattlefieldRow.svelte:64-77`, `:131-143` | render attachments under the host |
| `lib/contextMenu.logic.ts:303-314` | sorcery-speed clause in `abilityBlocked` |
| `lib/components/board/Board.svelte:499-571` | **no change** — the activation + targeting flow already works |

Docs: `docs/protocol.md:364` (the `CardView` paragraph). **Not** in this
branch's scope.

## Sizing

Six sub-PRs, one sprint. Comparable in shape to S21's activated-ability
arc: a lot of surfaces, very little new machinery.

| # | Scope | Rough size | Depends on |
|---|---|---|---|
| 1 | Relation, `MoveCard` clear, both SBA branches, events + listener bump, `attached_to` on the wire, `clone_test.go` pin. Ships with zero player-visible change. | ~400 lines + tests | #279 landed |
| 2 | Equip: `AttachedToSource` predicate, `AttachForEffect`, Skullclamp, Darksteel Plate, Sword of Hearth and Home | ~250 | 1 |
| 3 | Auras: resolution-time stamp, `HasSubtype`, Rancor, Pacifism | ~200 | 1 |
| 4 | Shroud / hexproof target gate; then Swiftfoot Boots + Lightning Greaves | ~150 | 1 |
| 5 | Client render + sorcery-speed grey-out | ~200 | 1 |
| 6 | One Curse (player attachment, the `TargetPlayer` branch end to end) | ~150 | 1, 3, 5 |

`AttachedAt` / CR 613.7d (decision 2) is a cuttable half-day inside
sub-PR 1.

**Split recommendation: equipment first.** Four reasons:

1. **Equip needs no new engine verb.** Sub-PR 2 is card files plus a
   predicate. Auras need a new stamp inside `resolveTopOfStackLocked`
   (`mutations.go:1274-1282`) — the hottest path in the engine and a
   file shared with four in-flight S32 branches.
2. **The Equipment SBA branch is the forgiving one.** Getting the Aura
   branch wrong puts cards in the graveyard; getting the Equipment
   branch wrong leaves an artifact cosmetically attached. Ship the
   recoverable branch while the relation is new.
3. **The payoff is all equipment.** Ranks 12, 13 and 40 of the format.
   Rancor and Pacifism are not in the top-100 triage at all.
4. **#89's Voltron deck is an equipment deck.**

The relation itself (sub-PR 1) must be designed for both regardless —
which is exactly why `AttachedTo` is a `TargetRef` and not a card ID.

## What this unblocks

**#89's bot archetypes.** ADR 0033 §7
(`docs/decisions/0033-ai-bot-seat.md:263-268`) names three of six
decks buildable today — aggro, ramp-stompy, thin spell-based control —
and gives three reasons for the rest: Voltron and Equipment/Aura
strategies wait on attachment, Aristocrats wants more of S21/S23, Combo
is "a bad idea regardless". This ADR moves **exactly one**: Voltron,
taking the count from **3 of 6 to 4 of 6**. Aristocrats is a catalog
problem this does not touch, and Combo is a judgement call, not a gap.

The bot's *action* half is already free: `activatedMoves`
(`legal/abilities.go:27-100`) will emit equip moves with no enumerator
change, correctly gated on the sorcery-speed window (`:40-42`) and mana
affordability (`:54-59`).

The bot's *evaluation* half needs one change, and it is worth naming
now so S33 hands it over cleanly. #89's `score()` counts "non-creature
permanents × w_utility" and "creatures: Σ(power + toughness)". An
attached Equipment would be counted **twice** — once as a utility
permanent, and again through the P/T it grants, which arrives on
`CardView.power`. An attached permanent should score zero on its own
line.

**Beyond bots.** ~30 cards per #76, the top-100 triage's fourth
build-order item cleared, the `Enchantment — Aura` type made functional
for the first time, and the last structural prerequisite for Mind
Control-style control change (which becomes a Layer 2 effect registered
at attach time, per #76 — not part of this ADR).

## What this deliberately does not do

- **Control-changing Auras (Mind Control).** Layer 2 is a stub
  (`layers.go:184`), `Card.Controller` is written directly in ~a dozen
  places, and a Layer 2 effect would have to win against all of them.
  Deferred with #76's original framing intact.
- **Protection (CR 702.16).** #176 routes protection to S24 alongside
  Mind Control. Decision 10 takes only the narrow shroud / hexproof
  *target gate*, which is a legality check, not the full
  damage-prevention / can't-be-enchanted / can't-be-blocked bundle.
- **Fortifications, Reconfigure, mutate, "attach to a battle".** Same
  relation, different attach verbs. Nothing here blocks them.
- **Auras entering attached from anywhere but the stack** (Animate Dead
  returning attached, Aura Shards' "attach"). `AttachForEffect` is the
  primitive; no card in scope calls it that way.
- **Drag-to-attach.** Decision 14.
- **Equipment attached to a planeswalker or battle.** CR 301.5c says
  creatures only; the type check in decision 9 enforces it.

## Note on stale documentation

Three claims in the issue and the surrounding docs did not survive
checking, and all three change the plan:

- **#280 §1 proposes `Card.AttachedTo *uuid.UUID` and says "`clone.go`
  must carry it."** The pointer form is the one shape `cloneCard`
  (`clone.go:197-198`) silently aliases, and a `uuid.UUID` of any form
  cannot express a Curse's player host. A value `TargetRef` is safe by
  construction, admits players, and needs **no `clone.go` change at
  all**. Decision 1.
- **#280 §1 asks for a reverse `attachments[]` list on the host.** Two
  wire representations of one relation that can disagree, when the
  client already derives exactly this kind of partition
  (`PlayerPanel.svelte:148-154`). Decision 13.
- **#280 §6 refers to "the replay scrubber".** There is none. The
  replay is a downloadable JSONL of full `GameView` snapshots
  (`ws/room.go:315-356`, `lobby/http.go:645`), so it inherits
  attachment from `CardView` with no work. Decision 15.

And one that is not stale but is easy to miss: `Card.IsCreature()`
reads the printed type line, not `Effective().Types`
(`card.go:348-350`) — recorded as engine finding 1 in #255 and the
reason decision 9 exists. An unattach SBA written the obvious way would
be wrong.
