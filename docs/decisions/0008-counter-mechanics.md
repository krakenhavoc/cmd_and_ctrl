# ADR 0008 — Counter mechanics (S13.2)

**Status:** Accepted · 2026-04-21 · Sprint S13.2

## Context

S13.1 shipped the four canonical Commander SBAs (lethal damage, 0
toughness, 0 life, 21 commander damage, empty-library draw) plus the
priority + stack loop, but explicitly deferred the **counter SBAs**
(planeswalker loyalty 0, battle defense 0, +1/+1 / -1/-1 cancel,
poison ≥ 10, saga final chapter) to S13.2. S13.2 also closes the
counter surface end-to-end: a per-player named-counter map (poison,
energy, experience, rad, plus homebrew), the cleanup-step damage
clear (CR 514.2), the shared counter-type registry that gives the
client iconography hooks, and the basic counter-pip overlay UI.

## Decisions

### 1. Five counter SBAs in the existing loop

`stateBasedActionsLocked` from S13.1 grew four new branches and one
ordering rule:

- **704.5q +1/+1 / -1/-1 cancel runs FIRST** — before the lethal-
  damage / 0-toughness destruction passes. A 2/2 with one +1/+1, one
  -1/-1, and 1 marked damage doesn't die: the counters cancel first,
  leaving a 2/2 with 1 damage. Pinned by
  `TestS132SBAPlusMinusCounterCancel` so a future refactor doesn't
  silently reorder them.
- **704.5i** — planeswalker with 0 loyalty counters → owner's
  graveyard. The S13.1 placeholder-creature exemption doesn't apply
  here; planeswalkers always die at 0.
- **704.5v/w** — battle with 0 defense counters → owner's graveyard.
  Symmetric with the planeswalker SBA.
- **704.5c** — player with ≥ 10 poison counters loses. Reads the
  unified `Player.Counters["poison"]` map; falls back to the legacy
  `Player.Poison` int field if the map is empty (so pre-S13.2
  replays decode correctly).

Saga final-chapter sacrifice (CR 704.5s) is the last SBA the plan
called out; the SBA *half* is trivial (any saga whose `lore` counter
≥ its final chapter index is sacrificed) but requires per-card
final-chapter metadata that the engine doesn't have yet. Deferred to
S14+ alongside the effect catalog. Documented as out-of-scope here.

### 2. `Player.Counters` map alongside legacy `Poison` / `Energy` ints

**Decision:** Add `Player.Counters map[string]int` keyed by counter
name (poison, energy, experience, rad, homebrew). Keep
`Player.Poison` and `Player.Energy` as int fields synchronised with
the corresponding map entries via `setPlayerCounterLocked` so the
S10 `set_poison` / `set_energy` actions and any pre-S13.2 replay
path keep working.

**Why both:** The unified map is the future shape — generic, sparse,
extensible to homebrew. The legacy ints are load-bearing for the
existing SetPoison / SetEnergy actions and the S10 PlayerHeader
copy that reads `seat.poison` / `seat.energy` directly. Mirroring
both costs almost nothing and avoids a wire-breaking migration.

### 3. `add_player_counter` as the canonical mutator

**Decision:** New `add_player_counter` action takes `{ name, delta }`
and routes through `Game.AddPlayerCounter`, which writes the unified
map AND mirrors to the legacy int field for poison / energy. Drives
`runStateChecksLocked` after the mutation so 10 poison or future
counter-driven losses fire immediately.

The old `set_poison` / `set_energy` actions also now call
`runStateChecksLocked` so a SetPoison(10) wins the game on the same
boundary an `add_player_counter("poison", 10)` would.

### 4. Cleanup-step damage clear (CR 514.2)

**Decision:** Extend the existing `runStepEntryHooksLocked` cleanup
branch (which already auto-advances the cursor per S13) to zero
`Card.DamageMarked` on every battlefield card before the auto-
advance. The lethal-damage SBA from S13.1 reads `DamageMarked`, so
clearing it here means per-turn damage doesn't carry over.

Order matters: the clear runs BEFORE the auto-advance. A creature
that would have died to lethal damage at end-of-turn already died via
the SBA loop earlier; cleanup is for survivors.

### 5. Counter-type registry as a single source of truth

**Decision:** New `server/internal/game/counter_types.go` with
named constants for every counter type the engine references
directly (`CounterPlusOne`, `CounterMinusOne`, `CounterLoyalty`,
`CounterDefense`, `CounterPoison`, `CounterEnergy`,
`CounterExperience`, `CounterRad`, plus a few common utility
counters: charge, stun, shield, lore). New
`client/src/lib/counterTypes.ts` mirrors the strings and adds
iconography (colour + glyph + abbr) for the pin set.

Unknown counter names round-trip through both engine and wire
without validation — homebrew counters get a neutral pip colour and
the first 3 characters of the name as the abbr.

`PoisonLethal` (10) is exported as a named constant on both sides
so the SBA threshold is easy to find and audit.

### 6. Client UX: counter pips on cards + per-player counter row

**Decision:** New `CounterPips.svelte` renders a stacked column of
chips at the top-right of every battlefield card showing each named
counter and its count. Pip colour comes from the registry. Damage
marked lands as a separate red badge at bottom-right.

`PlayerHeader.svelte` grew a row of counter markers: poison and
energy keep their existing dedicated chips (legacy markers for
backwards visual compat), and a new generic loop renders any other
non-zero `seat.counters` entry (experience ⭐, rad ☢, homebrew •).
Self-only +/- buttons for poison and energy now route through
`add_player_counter` so the SBA loop runs on every click.

**Out of scope this sprint:** the counter-inventory popover (right-
click → "Counters" menu with per-type +/- rows for arbitrary types)
and animated counter placement. The pip overlay + marker row are
enough to verify counters are working at the table.

## Consequences

- Counter SBAs round-trip end-to-end: planeswalker / battle drops
  to graveyard at 0; +1/+1 + -1/-1 cancels cleanly with the right
  ordering; 10 poison loses the game.
- Wire grew `PlayerView.counters` (sparse map). Pre-S13.2 clients
  ignore it; S10 `seat.poison` / `seat.energy` ints stay populated.
- New `add_player_counter` action lands on the wire with the
  player-scoped guard (the affected player adjusts their own
  counters; admins bypass).
- Cleanup-step damage clear is now an engine guarantee — sub-PR
  S13.4 can build interactive discard on top without re-doing this.
- Saga SBA stays deferred until S14+ effect-catalog work supplies
  per-card final-chapter metadata.
- Counter-type registry gives S14+ a stable taxonomy for the
  effect catalog to bind against (any "put +1/+1 counter on ~"
  parser can name-key into the same constants the SBA loop reads).

## Amendment (2026-09-28, #1664): every P/T counter kind changes power and toughness

**Context.** §1 and the S13.2 gap analysis read "counters change P/T"
as "+1/+1 and -1/-1 change P/T": `Card.PowerForComparison` and
`Card.CurrentToughness` added `Counters["+1/+1"]` and subtracted
`Counters["-1/-1"]`, and nothing else. CR 122.1a covers every
"+X/+Y" and "-X/-Y" counter, and older cards print several: -2/-1
(Contagion), +1/+0 and +0/+1 (Dwarven Armorer), -0/-1 (Wall of
Roots), +1/+2 (Armor Thrull), +2/+2 (Baron Sengir). Such a counter was
stored, drawn as a pip, and changed nothing — Wall of Roots made {G}
every turn for free forever.

**Decision.**

1. **One parser.** `game.ParsePTCounter(name)` (`pt_counters.go`) is
   the only place that decides whether a counter name is a P/T counter
   and what it is worth. The grammar is closed: `[+-]N/[+-]M`, both
   signs present, one to three decimal digits a side. "+X/+X", "1/1",
   "loyalty" and any homebrew name are not P/T counters, so an unknown
   name can never be mistaken for a stat change. `PTCounterDelta` sums a
   counter map through it.
2. **The counter step sums every kind.** The CR 613.4c step — which this
   engine applies after the layer system, in `PowerForComparison` /
   `CurrentToughness`, rather than inside layer 7 — reads
   `PTCounterDelta`. The view, combat, the SBAs and every card that
   reads "its power" go through those two helpers already, so they
   follow. The catalog's three last-known-P/T readers (dies triggers
   and off-battlefield reads, `b13LastKnownPower`,
   `b17LastKnownPowerOffBattlefield`, `b29LastKnownToughnessOffBattlefield`)
   read the same sum off the event log via `b13LastKnownPTDelta`.
3. **CR 704.5q stays two kinds.** Annihilation is +1/+1 against -1/-1
   and nothing else. A +1/+0 and a -1/-0 on one creature both stay; so
   do a +1/+1 and a -2/-1. The §1 SBA is unchanged and now says so.
4. **"+1/+1 counters" printed on a card still means that kind.**
   Hardened Scales, Branching Evolution, Conclave Mentor, "the number
   of +1/+1 counters on it" and every other kind-specific reader keep
   naming `CounterPlusOne`: a +1/+2 counter is not a +1/+1 counter, and
   Hardened Scales does not add to one. Doubling Season ("one or more
   counters") doubles any kind, as printed.
5. **Proliferate's automatic pick** treats a P/T kind as harmful when it
   takes more than it gives (a -2/-1, a -0/-1), and helpful otherwise.
6. **The client** pip for a P/T kind other than the two pinned ones
   shows its whole name ("+1/+0", not "+1/") in the +1/+1 green or the
   -1/-1 red. The wire is unchanged: `CardView.counters` already carried
   the name as-is, and `power` / `toughness` already carried the sum.

**Not changed.** The bot heuristic's `counterRemovalValue` still prices
only +1/+1 / -1/-1 removal specially; any other kind is the generic
rate. The heuristic package may not import `game` (ADR 0033 §3), and no
catalogued card yet removes an odd P/T counter as a cost (Balduvian
Hydra is the one that would).

**Cards.** Contagion, Wall of Roots (the first printed mana ability
whose cost adds a counter), Dwarven Armorer, Armor Thrull and Lightning
Serpent, all `full`.
