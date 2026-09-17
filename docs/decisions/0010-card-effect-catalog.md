# ADR 0010 — Card-effect catalog foundation (S14)

**Status:** Accepted · 2026-04-21 · Sprint S14

## Context

S13's rules graft (S13.1 – S13.5) landed priority, stack, SBAs,
cleanup-discard, and per-card visibility. Everything on the
stack still resolves **manually**: players type in the life
loss after Lightning Bolt, hand-pick the card going to
graveyard from Mind Rot, click through cleanup themselves.
S14 closes the smallest viable slice of that gap — ~30 of
the most-played Commander cards resolve end-to-end with zero
manual intervention, while every other card keeps today's
Cockatrice-style manual posture.

The sprint deliberately does not ship a general rules engine:
no oracle-text parser, no layer system, no replacement
pipeline, no mana pool enforcement, no keyword engine. Those
are Phase-7 sprints S15 – S19 in their own right. S14 ships the
infrastructure they all plug into (events, listeners,
primitives, a declarative DSL) plus the starter 30 cards that
exercise 13 of 15 primitives.

## Decisions

### 1. Declarative Forge-style DSL in Go structs

Each catalog card is a ~10 – 50-line file under
`server/internal/cards/effects/` containing exactly one
`init()` that calls `effects.Register(Spec{...})`. No
oracle-text parsing — the author hand-translates "deal 3
damage to any target" into `DealDamage{Amount: 3, Target:
item.Targets[0].ID}.Apply(ctx)`.

**Why:** Oracle text is famously weird (Umezawa's Jitte's
second ability, Dark Depths' counter-cap, anything with
"as...enters"). Scryfall's own parser is heuristic and lossy.
A hand-written DSL keeps the engine small and the per-card
semantics exact; the trade is a per-card cost of ~15 LoC,
paid once.

**Cross-reference:** Forge and MTG Online use the same pattern
for largely the same reasons. The precedent was explicit in the
S14 planning round.

### 2. Catalog membership keyed by `oracle_id`, not `scryfall_id`

Scryfall has two stable IDs per card: `scryfall_id` is unique
per **printing** (Lightning Bolt has 30+), `oracle_id` is
stable across printings. The catalog keys on `oracle_id` so
the deck importer can stamp any printing of Lightning Bolt
and the catalog still hits.

**Gotcha that bit us in sub-PR 4:** we initially keyed on
`ScryfallID` and rebased to `OracleID` mid-flight. Any new
effect file uses `OracleID`; the per-card tests pass oracle
IDs into `castCatalogSpell`.

### 3. Opt-in per oracle ID; Cockatrice-style fallback preserved

`effects.Lookup(oracleID)` returns `(Spec, false)` for any
card not in the catalog. The resolution path in
`resolveTopOfStackLocked` calls `fireEffectResolverLocked`
(which no-ops on a nil lookup) and then routes to the
battlefield / graveyard as before. Non-catalog cards are
indistinguishable from their S13 behaviour.

A `TestNonCatalogSpellStaysSandbox` canary pins this — a spell
with an empty oracle ID resolves to graveyard with no life
change. Every sub-PR is expected to keep that test green.

### 4. Dependency inversion via function-variable hooks

The `cards/effects` package must reach into `*game.Game` for
mutations. The `game` package would naturally want to call
`effects.Lookup` from `resolveTopOfStackLocked`. That's a
cycle.

**Decision:** declare function-var slots in the `game`
package that `effects` populates at init time:

```go
// server/internal/game/effect_hooks.go
var EffectResolver  func(g *Game, item *StackItem, oracleID string) error
var ETBEffectHook   func(g *Game, cardID uuid.UUID, oracleID string) error
var IsCatalogCard   func(oracleID string) bool
var CatalogTargetMode func(oracleID string) string
```

```go
// server/internal/cards/effects/wire.go
func init() {
    game.EffectResolver = resolveSpell
    game.ETBEffectHook = fireOnETB
    game.IsCatalogCard = Has
    game.CatalogTargetMode = ...
}
```

A blank import of `cards/effects` from `cmd/server/main.go`
triggers the registration. Without the import the hooks stay
nil and the entire server reverts to S13 manual behaviour —
the catalog is **opt-in at the build level too**. Useful for
minimal test binaries.

### 5. Effects run under the resolution write lock

`resolveTopOfStackLocked` holds `g.mu` as a write lock. It
calls `EffectResolver` which calls the registered `OnResolve`
which calls primitive `Apply` methods which call
`g.XxxForEffect(...)` helpers. **None** of those helpers may
take `g.mu` themselves — that would deadlock.

Mechanical consequence: every effect helper on `*Game` is
named `XxxForEffect` (e.g. `PlayerByIDForEffect`,
`DealDamageToPlayerForEffect`). The naming is the guardrail.
Go can't encode "locked context" at the type level; review
enforces the rule. A primitive that reaches into
`ctx.Game` and calls a public locking mutator (like
`g.CastSpell`) deadlocks — documented on every helper.

### 6. 15 composable primitives

The primitive list (`effects/primitives.go`):

```
DealDamage GainLife DrawCards DiscardCards MillCards
DestroyTarget ExileTarget BounceToHand
TapTarget UntapTarget
CounterTarget AddCounter CreateToken
ReturnFromGraveyard SearchLibrary
```

Each is a struct with an `Apply(ctx *Context) error` method.
Cards compose them declaratively:

```go
// Lightning Helix — "Deal 3 damage to any target. You gain 3 life."
OnResolve: func(item *game.StackItem, ctx *Context) error {
    target := item.Targets[0].ID
    if err := (DealDamage{Source: ctx.Source(), Target: target, Amount: 3}).Apply(ctx); err != nil {
        return err
    }
    return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
}
```

S14 catalog exercises 13 of 15 primitives at least once.
`TapTarget` / `UntapTarget` ship with unit-test-only coverage
so S15 (mana) has the hooks ready.

### 7. `Spec.OnETB` is a direct-call shim in S14

`OnETB` fires from `fireETBHookLocked`, which is called from
every battlefield-entry site in the game package (land cast,
spell resolution, admin move-to-battlefield). Synchronous
call inside the same mutation lock as the ETB event emission.

S19's "ability auto-fire" sprint migrates `OnETB` through the
listener registry. The catalog Specs stay unchanged: the
migration is a refactor in `wire.go` only. Eternal Witness's
graveyard fetch and Solemn Simulacrum's basic-land fetch
already express the shape S19 will consume.

### 8. Planeswalker loyalty via `Spec.StartingLoyalty`

Planeswalkers enter the battlefield with a printed loyalty
value (CR 306.5b). The Spec carries it as a plain int; the
ETB hook stamps loyalty counters via `AddCounterForEffect`
**before** the next SBA boundary — otherwise a planeswalker
with `StartingLoyalty = 0` (never the real case, but keep
the guard correct) would die to the S13.2 0-loyalty SBA.

Activated loyalty abilities stay manual in S14. S13.1's
`activate_loyalty` action is the affordance. S19 wires real
+1 / -1 / -2 effects through a `Spec.Abilities` field —
again, the per-card Spec stays source-stable.

The Wandering Emperor is the only S14 planeswalker.
Starting loyalty 3. Samurai token template already lands
(`WhiteSamuraiToken` in `tokens.go`) so the S19 wiring has
nothing to add.

### 9. Event log + pull-based listener registry

New `server/internal/game/events.go`:
`Event{Kind, Actor, Source, Target, Amount, CardID, OldZone, NewZone, Seq}`.
Every rules-visible mutation calls `Game.EmitEvent(ev)` under
lock. Events live on the `Game` struct (deep-copied by
`clone.go` for undo); a per-emit `notifyListenersLocked`
walks the listener registry.

**S14 registers zero production listeners.** The registry is
scaffolding for S19 triggered abilities. A `noOpListener` in
tests exercises the wiring.

Events are exposed on the wire (`GameView.events`, bounded
rolling window) so the client can surface auto-toast
notifications ("Alice took 3 from Lightning Bolt") as a
visual nudge against double-applying after an auto-resolve.

### 10. `PendingChoices` for asynchronous effect-driven decisions

Two discard systems now coexist:

- **`DiscardPending map[uuid.UUID]int`** (S13.4) — cleanup-step
  over-max discard. Chooser is always the over-full player;
  client pops `DiscardPromptModal`. Only useful when the
  discarder picks from their own hand.

- **`PendingChoices []PendingChoice`** (S14) — generic async
  decisions where the chooser is not necessarily the source-
  zone owner. Used by Thoughtseize (caster picks from the
  target's revealed hand). Client pops `ChoicePromptModal` for
  any pending choice whose `Chooser == viewerID`.

Thoughtseize flow:
1. Spell resolves. `QueueDiscardFromRevealedHand(caster,
   target, ...)` reveals `target`'s hand sticky (via S13.5
   KnownBy) and queues a PendingChoice addressed to `caster`.
2. Caster's client shows the revealed hand in the modal;
   caster picks one card and submits `resolve_choice`.
3. Server `ResolvePendingChoice` moves the picked card to
   `target`'s graveyard and drains the queue.

Mind Rot uses the older shape (`DiscardChoiceForEffect` writes
into `DiscardPending` — target picks from their own hand). Same
UI surface, different wire, both additive.

Future effects wanting "chooser != discarder" or pick-from-
zone-X semantics can reuse `PendingChoices` by minting a new
`PendingChoiceKind` + dispatch case. No per-card plumbing.

**Amendment (2026-09-17, #651): an effect's discard is a pending
choice; `DiscardPending` is cleanup-only.**

"Both additive" was wrong, and the paragraph above described
behaviour the code did not have. The two obligations do not have
the same shape, and sharing a map cost the engine both of them:

- **Nothing waited.** `DiscardPending` is not a `PendingChoice`, so
  the turn-structure verbs never looked at it. A Mind Rot or a loot
  could be passed straight through: spells resolved and steps
  advanced while the discard was still owed. It is part of the
  resolving effect (CR 608.2c) and must finish inside it.
- **Cleanup erased it.** `populateDiscardPendingLocked` sets
  `g.DiscardPending = nil` on cleanup entry and then writes only the
  active player's hand-size count, so an effect discard still owed
  was silently dropped.

`QueueDiscardChoiceForEffect(DiscardPrompt{...})`
(`server/internal/game/effect_api.go`) now queues a
`PendingChoiceChooseCards` addressed to the **discarding** player over
their own hand, with `Zone: ZoneHand` — the pick Sylvan Library and
`PutFromHandOntoBattlefield` already use. Not a third discard system
and not a new kind: the live-zone re-check, the set-level `Validate`
hook, the `internal/legal` enumerator case and the client's
`ChoicePromptModal` all came with it. What makes it a *discard* is the
continuation, which moves the picks to the graveyard and emits one
`EventDiscardCard` per card before running `DiscardPrompt.Then`. The
prompt then blocks `advance_step` / `pass_priority` / `pass_turn`
through #730's gate (ADR 0018 §6 amendment) with no extra machinery.
`DiscardChoiceForEffect(player, n)` survives as a thin wrapper for the
plain case, which is most of the catalog.

Three rules the shape settles:

- **Order.** A loot ("draw a card, then discard a card") draws in the
  statement above the prompt, so the prompt is built from the
  post-draw hand and a card just drawn is a legal pitch. A rummage
  ("discard a card, then draw") puts its second half in `Then`, which
  runs once the cards are in the graveyard — the Scry contract, for
  the same reason. Syphon Mind is the catalog's one rummage and the
  reason its declared caveat is gone.
- **Empty hand queues nothing**, and `Then` still runs: CR 701.8a
  discards as many as you can, and "discard your hand, *then* draw
  three" draws three from an empty hand. A prompt with no candidates
  and a floor of one is one nobody can answer (#544).
- **Random is not a choice.** `DiscardRandomForEffect` (CR 701.8b)
  and "discard your hand" stay synchronous moves with no prompt.

`DiscardPending` keeps only what CR 514.1 uses it for: the active
player's cleanup-step hand-size discard, drained by
`discard_selection`, rendered by `DiscardPromptModal`, enumerated by
`legal.cleanupDiscardMoves` — whose "Discard to hand size" label is
true again now that a Mind Rot never lands in that map.

The cost, paid knowingly: an effect discard carries a continuation, so
`GameSnapshot.Restorable()` is false while one is open, exactly as it
already was for a scry, a search or an optional trigger. The
alternative — a serialisable discard nothing waits for — is the bug
this amendment closes.

### 11. Sticky reveal preserved through view filtering

S13.5's `KnownBy` set on `Card` is sticky — once revealed to
a viewer, a hand card stays known until a shuffle / mulligan
clears it. The view-filter path had a gap: `hideZoneContents`
stripped every opponent hand card wholesale, so Thoughtseize's
reveal never reached the caster.

**Fix:** a new `keepKnownInHandZone` path branches on viewer
identity in `FilterViewFor`:
- Seated viewer (`viewerID != ""`): opponent hand zones keep
  cards where `KnownByYou == true`; unknown cards are dropped
  but the `Count` stays accurate.
- Spectator / admin (`viewerID == ""`): opponent hand zones
  stay fully hidden via `hideZoneContents` (the isKnower
  short-circuit would otherwise leak every hand).

Client `Hand.svelte` composes revealed cards (face-up) +
synthesised face-down placeholders for the `count - revealed`
remainder. The compact-opponent-hand setting still caps
placeholders; revealed cards always render regardless.

### 12. Kept-but-gated manual actions

Actions like `draw_card` and `untap_all` still exist on the
wire for sandbox use, replay compatibility, and "fix wedged
state" admin moves. They do not conflict with the auto-resolve
path because the auto-action runs in the step-entry hook
(S13) and the manual action is a no-op during that exact
step by design.

The same posture applies to effect-catalog cards: if a player
manually clicks `change_life -3` after Lightning Bolt
auto-resolves, the server applies both. The UX nudge against
double-apply is the event-log toast + the gold-leaf AUTO
badge on the card — no server-side de-dup.

### 13. Target-picker flow (two-click cast for targeted catalog spells)

`Spec.TargetMode string` declares the announce-time prompt
shape (`"any"`, `"player"`, `"creature"`, `"stack_spell"`,
`"card_in_graveyard"`, or empty for no prompt). The view layer
mirrors it onto `CardView.target_mode`; the client writes a
`TargetingState` into a Svelte store when the viewer clicks a
targeted catalog card; the board-level click interceptor
fires `cast_spell` with a populated `targets[]` when the
second click lands on a legal surface (player portrait,
battlefield creature, stack item).

Target legality validation at **announce** is S20 smart-cast
territory. S14 validates at **resolve** via CR 608.2b
("countered by game rules"): if every targeted slot is
illegal at resolution time, the spell fizzles to graveyard.

### 14. Stack-target arrows (visual affordance)

Stack items with card targets render a gold dashed SVG arrow
from the stack item to the target card (or the target stack
item, for Counterspell / Negate / Swan Song). The wiring
reuses the existing `CombatArrows` component + the
`data-stack-item-id` / `data-instance-id` data attributes;
`rectIn` has a two-step fallback (`[data-instance-id]` →
`[data-stack-item-id]`) so stack-to-stack arrows resolve.

Gold dashed stroke distinguishes these from combat (red
solid) arrows.

### 15. Dev deck-validation bypass

S14's 30-card catalog is hard to rehearse against a full
99-card deck. New `CMDCTRL_DEV_SKIP_DECK_VALIDATION=1` env
var relaxes the deck validator on the `POST /games/{id}/decks`
path. Sets a `dev_skip_validation` warning on the response so
the client can surface "you are in dev mode" if it cares.

Never used in production; the flag is explicit opt-in and
the lobby response is self-documenting.

### 16. Sandbox simplifications, documented per-card

Each catalog card's file carries an explicit comment listing
what it does NOT do at sandbox quality:

- **Enters-tapped** deferred to S17 (replacement pipeline).
  Path to Exile's land, Cultivate's first land, Solemn
  Simulacrum's land all enter **untapped** in S14.
- **"You choose" tutors** deferred to S20 (smart-cast UI).
  Tutors auto-pick the first library match; Regrowth /
  Eternal Witness auto-pick the top of the graveyard.
- **Target-at-ETB** (Acidic Slime's "destroy target
  artifact/enchantment/land", Solemn's dies-trigger)
  deferred to S19 (ability auto-fire pipeline). Acidic Slime
  is explicitly **not shipped** in S14.
- **Thoughtseize "non-land, non-Thoughtseize"** predicate
  stays unenforced — the caster picks any card in the
  revealed hand.
- **Vampiric Tutor "put on top"** simplified to "put in hand";
  revisit alongside replacement timing in S17.

Every deferral is one-line-grep-able from
`server/internal/cards/effects/*.go`.

## Consequences

- **30 catalog cards** ship auto-resolving end-to-end:
  Lightning Bolt, Shock, Lightning Helix, Pyroclasm, Wrath of
  God, Damnation, Day of Judgment, Counterspell, Negate, Swan
  Song, Divination, Harmonize, Sign in Blood, Glimpse the
  Unthinkable, Thoughtseize, Mind Rot, Swords to Plowshares,
  Path to Exile, Unsummon, Demonic Tutor, Vampiric Tutor,
  Cultivate, Regrowth, Eternal Witness, Solemn Simulacrum,
  Sol Ring, Arcane Signet, Birds of Paradise, The Wandering
  Emperor.
- 13 of 15 primitives exercised by live cards. Tap / Untap
  ship with unit-test coverage only.
- `GameView.events` rolling window exposes the authoritative
  per-game event log for client toasts + future analytics.
  Pre-S14 replay snapshots decode as `events: nil`.
- `GameView.pending_choices` materialises the async-decision
  queue on the wire, per-viewer-redacted via the same
  `KnownBy` projection used for hand reveal.
- S19 (triggered-ability auto-fire) plugs into the listener
  registry without per-card changes. OnETB migrations become
  a refactor in `wire.go`.
- S15 (mana pool + auto-tapper) plugs into the catalog via
  activated-ability specs on the existing `Spec` struct.
- Acidic Slime and Solemn Simulacrum's dies-trigger are
  explicit S19 handoffs — the Spec exists, the trigger wiring
  lands with the listener pipeline.
- Deck importer now stamps both `ScryfallID` (printing) and
  `OracleID` (stable) so every deck works with the catalog
  regardless of which printing the player imported.
