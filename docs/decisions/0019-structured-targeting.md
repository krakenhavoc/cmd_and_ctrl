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

## Out of scope (next sub-PRs)

- **Trigger target picking.** The S19 ETB-destroy cards still
  auto-pick; a `pick_target` pending choice built from the same
  `LegalTargets` walk replaces `pickFirstOpponentNonland` and the
  `HasLegalTarget` warning.
- **Multi-target** (`Min`/`Max` > 1, "up to N", distribute) and
  **AllowSameTarget**.
- **Modes and X** — `ModeSpec`, the mode picker, X live validation.
- **Hexproof / shroud / protection** as target-legality
  modifiers — the predicate hook is where they'll go.
