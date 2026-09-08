# ADR 0019 — Structured targeting

**Status:** Implemented (sub-PR 1) · 2026-09-04 · Branch `feat/s20-target-predicates`

## Context

Until S20, "targeting" was a client hint. `Spec.TargetMode`
("creature", "player", …) told the Svelte picker which surfaces to
make clickable; the server accepted whatever ID came back. Negate
could counter a creature spell, Path to Exile could exile a land,
and a resolving spell only checked that its target still *existed*
(CR 608.2b's existence half), never that it was still a legal
target. The S19 trigger cards went further and auto-picked targets
server-side ("first opponent artifact") because there was no
picker at all for abilities.

The sprint's exit criterion — cast Doom Blade, see only non-black
creatures — needs the predicate to live on the server (it's the
rules authority and the only place with full information), and the
client to be told the answer rather than re-deriving it.

## Decisions

### 1. `game.TargetSpec` is the truth; `TargetMode` is derived

A card declares `Spec.Targets *game.TargetSpec`: which zones a
card target may live in, whether players qualify, a `CardOK` /
`PlayerOK` predicate over candidates (live game + caster +
candidate), a `Label` for the banner, and `Min`/`Max` (1/1 for
now). `TargetMode` is derived from `Targets.Mode` for the wire;
cards without a spec keep the S13.1 free-form behaviour — no
legality check — so nothing regresses for un-migrated cards.

### 2. Predicates are a small composable library in the catalog package

`effects/targets.go`: type predicates (`Creature`, `Artifact`,
`Instant`, …, all post-layer), colour (`OfColor`, `NonBlack`,
`Colorless`), stats (`PowerLE`, `ManaValueLE`), control/ownership
(`YouControl`, `OpponentControls`, `YouOwn`, `NotSelf`), plus
`And`/`Or`/`Not`. Constructors (`TargetAny`, `TargetCreature`,
`TargetPermanent`, `TargetSpell`, `TargetPlayer`,
`TargetCardInGraveyard`) fix the zone set and mode so a card file
reads like oracle text. Predicates run under the game lock and must
stay pure: the legal-target walk calls them per candidate on every
snapshot.

### 3. Colour comes from Scryfall, with a mana-cost fallback

`Card.Colors` is stamped at deck import from Scryfall's computed
`colors` (handles indicators, Devoid, hybrid). `EffectiveColors()`
falls back to the coloured symbols in `ManaCost` when the field is
empty (tokens, test fixtures). Layer 5 colour-changing effects
aren't modelled; `EffectiveColors` is the seam when they are.

### 4. Legality is enforced twice, with the same predicate

- **Announce (CR 601.2c):** `CastSpell` validates count and each
  ref against the spec → `ErrIllegalTarget` (the card stays in
  hand). The client picker only offers legal targets, so this is a
  backstop against stale snapshots and hand-built payloads.
- **Resolution (CR 608.2b):** `spellAllTargetsIllegalLocked` runs
  the predicate from the item's controller's point of view when a
  spec exists, so a Doom Blade target that turned black — or a
  creature that stopped being one — counters the spell by game
  rules, not only a target that left. Ability items (triggers)
  keep the existence check until they carry specs.

### 5. The legal set rides the snapshot, not a round trip

`CardView.legal_targets` (`{players, cards}`) is stamped in
`ViewOfGame` for every hand / command-zone card with a spec, from
its owner's point of view, and stripped from opponents' views in
`FilterViewFor` (even for revealed hand cards). The client
targeting store reads it into two sets; surfaces (battlefield
cards, player portraits, stack items) ask `isLegalCardTarget` /
`isLegalPlayerTarget`; legal cards get a green ring; the banner
shows "N legal". `canCastFromHand` denies a targeted spell whose
set is empty ("No legal target"). Cost: a few predicate walks per
snapshot for the handful of targeted cards in a hand — negligible
at eight seats.

Why not a `query_targets` action: a round trip before every cast
adds latency to the most common click in the game, and the same
data is what the S13.3 greyed-illegal-actions affordance needs on
every render anyway.

### 6. Triggers pick on the board too (sub-PR 2)

`TriggeredAbility.Targets` gives a trigger the same clause. The
harvester computes the legal set the moment the trigger fires (CR
603.3d: targets are chosen as the ability is put on the stack); an
empty set removes the trigger with no prompt at all — no more
"Yes" that silently does nothing, which is what the S19
`HasLegalTarget` warning papered over. A "you may" prompt still
comes first; on "yes" (or immediately for a mandatory trigger) a
`pick_target` pending choice carries the frozen legal set. The
client doesn't render it as a modal: it enters the ordinary
board-click targeting flow with the prompt's set (graveyard
targets via the zone browser), and the click answers
`resolve_choice {target}`. `ResolvePickTarget` re-validates the
pick against the live board, stamps it on the built item together
with the spec, and the item's resolution runs the same CR 608.2b
re-check spells get.

The prompt can't be cancelled from the client — the server owns a
trigger that needs a target — and it pre-empts any cast-targeting
prompt in flight.

### 7. Modes carry their own target clause (sub-PR 4)

A modal card declares `Spec.Modes` — a `game.ModeSpec` with the
prompt ("Choose one"), the option labels, `Min`/`Max`, and per
option an optional `TargetSpec`. The card-level `Targets` stays
nil: Rakdos Charm's first bullet targets a player, its second an
artifact, its third nothing, and there is no single clause that
describes the card. The engine derives the cast's effective spec
from the chosen options (`castTargetSpec`), so announce validation
and the resolution re-check are the same code paths a Doom Blade
uses; the option's spec is what the picker highlights.

Two consequences follow. First, an untargeted chosen mode must
arrive with no targets and a targeted one with exactly its count —
the modal card is never free-form. Second, sub-PR 4 supports **one
targeted option per cast**: "choose two" cards whose options each
target (Cryptic Command, Kolaghan's Command) need per-mode target
slots on the wire and on `StackItem.Targets`, which is the same
plumbing multi-target needs, so it ships with that work rather than
as a special case here. `Register` panics if a card declares more
than one targeted option with `Max > 1`, so the limit fails at boot
rather than at the table.

On the wire the owner's hand card carries `modes` with each
option's `target_mode` + `legal_targets`, computed from the same
snapshot pass as `legal_targets`. The client's cast flow is now X
prompt → mode picker → board targeting → `cast_spell {modes,
targets, x_value}`; each step is its own small surface rather than
one consolidated dialog, because each has exactly one thing to ask.
`canCastFromHand` greys a modal card when fewer castable options
than `Min` remain (targeted options need a legal target to count).

Effects read the choice back with `ctx.HasMode(i)` and resolve the
chosen bullets in printed order (CR 700.2c).

### 8. One clause, N slots; partial illegality resolves (sub-PR 5)

A multi-target clause is the same `TargetSpec` with `Min`/`Max`
set — "two target creatures" is one predicate chosen twice, "up to
two" is `0..2`, "any number" is `1..0` (unbounded). Announce
rejects duplicates unless `AllowSame` (CR 115.3), and preserves the
order the player clicked in, so a positional clause ("2 damage to
any target and 1 damage to any other target") reads slots by index.

The spell's item now remembers the spec it was announced under,
the way trigger items already did. That makes the per-slot check
cheap: `Context.IsTargetLegal` runs the announced predicate, not
just an existence check, and `Context.LegalTargets()` is the
surviving subset. The all-illegal short-circuit still fizzles the
spell before `OnResolve`; with one target left the spell resolves
and does what it can (CR 608.2b) — Ashes to Ashes still deals its
5 to the caster with one creature gone.

On the wire `legal_targets` and `pick_target` carry `min`/`max`.
The client keeps one targeting store: at `max 1` the first click
completes the prompt as before; otherwise clicks toggle into a pick
list (gold ring on picked cards / portraits), the banner shows
"n/max picked" and a Done button that enables at `min`, and Enter
confirms. Trigger prompts share the flow and answer with
`resolve_choice {targets: [...]}`.

## Out of scope

- **Per-mode target slots** for "choose two" cards whose options
  each target (Cryptic Command, Kolaghan's Command) — needs
  `StackItem.Targets` grouped by mode and a two-step picker.
- **Divided damage** (`Distribution` already rides the item; no
  UI or predicate yet) and cost-per-target clauses (Fireball's
  `{1}` per extra target: the X prompt runs before targets are
  known).
- **Hexproof / shroud / protection** as target-legality
  modifiers — the predicate hook is where they'll go.
