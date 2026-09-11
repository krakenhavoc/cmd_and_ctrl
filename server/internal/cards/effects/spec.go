// Package effects is the S14 card-effect catalog: declarative Forge-
// style specifications for the ~30 starter Commander cards that
// resolve automatically on the stack. Entries are keyed by Scryfall
// ID, registered at package init() time in one-file-per-card under
// this directory.
//
// Design pillars (pinned in docs/decisions/0010-card-effect-catalog.md):
//   - Declarative DSL in Go structs, no oracle-text parsing.
//   - Opt-in per Scryfall ID — non-catalog cards keep today's manual
//     sandbox posture.
//   - Effects run under the resolution write lock; primitives call
//     *Locked helpers on game.Game only. A primitive that calls a
//     public locking mutator deadlocks — enforce via review, Go can't
//     encode "locked context" at the type level.
//   - Events emitted per primitive application (see game.EmitEvent).
//
// This file declares the Spec type. Registry, Context, and primitive
// types live in neighbouring files.
package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spec is the canonical declaration for one card in the catalog.
// Populated once at package init() time via Register; looked up at
// resolution time via Lookup. Absence in the registry ⇒ non-catalog
// ⇒ today's manual sandbox behaviour.
type Spec struct {
	// OracleID is the Scryfall oracle-level card identifier — stable
	// across printings (every printing of Lightning Bolt shares one
	// oracle_id). Required. The registry panics if two Specs collide
	// on this key so a copy-paste duplicate fails at server boot
	// (the earliest, most visible failure mode).
	OracleID string

	// Name is the human-readable card name. Present in the Spec so
	// logs / panics / tests have a nice handle without bouncing
	// through the Scryfall index.
	Name string

	// OnResolve runs when the spell's stack item resolves (after
	// target re-check, before zone routing). Nil means "no effect
	// on resolve" — the card still routes to the battlefield /
	// graveyard per type, but no automation fires. Typical usage
	// for vanilla creatures with no ETB triggers (Birds of Paradise
	// in S14).
	//
	// Returning an error emits an EffectError event and allows
	// resolution to continue — partial failure does not wedge the
	// stack.
	OnResolve func(item *game.StackItem, ctx *Context) error

	// OnETB fires after a permanent moves from the stack (or
	// battlefield source) into the battlefield. S14 direct-call
	// path; S19 re-routes through the listener registry without
	// per-card changes. Nil is common (most catalog cards have no
	// ETB effect).
	OnETB func(card *game.Card, ctx *Context) error

	// StartingLoyalty is the loyalty counter count a planeswalker
	// enters the battlefield with. 0 means "not a planeswalker" or
	// "planeswalker with 0 starting loyalty" (which the SBA would
	// immediately kill — never the real case). The resolution path
	// stamps this via AddCounter when the card crosses into the
	// battlefield. Used by The Wandering Emperor in S14.
	StartingLoyalty int

	// TargetMode tells the client what to prompt for at cast time.
	// Serialized to CardView.target_mode so the Svelte cast flow
	// can enter a targeting state before firing cast_spell with
	// populated targets[]. Empty string means "no target prompt,
	// cast immediately." Valid values:
	//
	//   ""             no prompt (Pyroclasm, Wrath, vanilla permanents)
	//   "any"          player / creature / planeswalker / battle
	//                  (Lightning Bolt, Shock, Helix)
	//   "player"       seated player only
	//   "creature"     battlefield creature only
	//   "stack_spell"  an item currently on the stack (Counterspell,
	//                  Negate, Swan Song)
	//   "card_in_graveyard" a card in any graveyard (Regrowth, Eternal
	//                       Witness)
	//
	// Target count is always 1 in S14; multi-target (Arcing Lightning
	// style "distribute 3 damage") is S22 territory.
	//
	// S20: prefer Targets. When Targets is set, TargetMode is derived
	// from Targets.Mode and this field is ignored. Cards still on a
	// bare TargetMode get the S13.1 free-form picker with no
	// server-side legality check.
	TargetMode string

	// Targets is the S20 structured targeting clause: which zones and
	// players the single target slot accepts, and the predicate a
	// candidate must pass. Drives the client's legal-target set, the
	// announce-time check (CR 601.2c) and the resolution re-check
	// (CR 608.2b). Build it with the constructors in targets.go —
	// TargetAny(), TargetCreature("target non-black creature",
	// NonBlack()), TargetSpell(…), TargetCardInGraveyard(…). Nil
	// means no structured targeting.
	Targets *game.TargetSpec

	// ManaAbilities is the list of activated mana abilities the card
	// exposes from the battlefield. Each entry is one tap-or-cost-
	// for-mana ability — Sol Ring's "{T}: Add {C}{C}", Birds of
	// Paradise's "{T}: Add one mana of any color", Arcane Signet's
	// commander-identity-restricted variant. Mana abilities do NOT
	// use the stack (CR 605.3); they resolve synchronously when the
	// activate_mana_ability action fires. The Index a client sends in
	// the action payload is the position in this slice.
	//
	// Empty / nil for cards that have no mana abilities (the vast
	// majority of S15 catalog). Basic lands get a synthetic default
	// ability derived from their TypeLine when no catalog entry
	// declares ManaAbilities — see Game.ActivateManaAbility for the
	// fallback shape.
	//
	// Added in S15 sub-PR 2.
	ManaAbilities []ManaAbility

	// Static is the list of continuous-effect static abilities the
	// card contributes to the layer engine while on the battlefield
	// (CR 613). Each entry declares its `Layer`, `SubLayer` (for
	// Layer7PT only), an `AppliesTo` predicate evaluated per
	// candidate target on every recompute pass, and an `Apply`
	// function that mutates the candidate's `Characteristic` in
	// place. The engine recomputes from scratch on every relevant
	// event (battlefield zone change, counter change) — so AppliesTo
	// re-runs against the current board state every time.
	//
	// Empty / nil for cards with no static abilities (the S15
	// majority). When non-empty, the card's effects only apply
	// while it's on the battlefield (CR 113.6 default — non-
	// battlefield-zone statics are deferred to a later sprint).
	//
	// Reaches directly into `game.StaticAbility` rather than a
	// per-package adapter type because the function shapes already
	// reference `game.Card` / `game.Game` / `game.Characteristic` —
	// no information would survive a parallel struct.
	//
	// Added in S16 sub-PR 3.
	Static []game.StaticAbility

	// Replacements is the list of CR 614 replacement effects the
	// card contributes while on the battlefield. Each entry
	// declares a `Watches` pre-filter (which event kinds it cares
	// about), an `AppliesTo` predicate evaluated per candidate
	// event, a `Replace` function that mutates or cancels the
	// event, and metadata (Controller, SelfReplacement, Label) for
	// the CR 616 order-choose prompt.
	//
	// Empty / nil for cards with no replacement effects (the S16
	// majority). When non-empty, the replacements fire only while
	// the card is on the battlefield (CR 113.6 default).
	//
	// Reaches directly into `game.ReplacementEffect` rather than a
	// per-package adapter type because the function shapes already
	// reference `game.ReplacementEvent` / `game.Game` / `game.Card` —
	// parallels the S16 `Static []game.StaticAbility` pattern.
	//
	// Added in S17 sub-PR 2. Populated by catalog cards starting
	// in sub-PR 3 (Doubling Season, Hardened Scales, Branching
	// Evolution).
	Replacements []game.ReplacementEffect

	// PrintedKeywords is the list of combat keywords printed on the
	// card — entries like "flying", "reach", "deathtouch", "lifelink",
	// "trample", "vigilance", "first strike", "double strike",
	// "menace", "defender", "haste", "flash". Canonical lowercase
	// tokens; see AGENTS.md §7 "Adding a combat-keyword card" for the
	// complete table.
	//
	// Feeds two consumers:
	//
	//   1. On-battlefield: wire.go synthesizes a self-only Layer 6
	//      StaticAbility per card that appends each entry to the
	//      card's own `Characteristic.Abilities`, so the keyword
	//      lands in `card.Effective().Abilities` alongside grants
	//      from other cards' static abilities (Lord of Atlantis).
	//   2. Off-battlefield: game.HasKeyword (keywords.go) falls back
	//      to `CatalogPrintedKeywords(oracleID)` when the card has
	//      no `effective` cache — needed for flash gating on a
	//      card still in hand.
	//
	// Keyword behaviour itself is engine-side (flying block
	// restriction, trample overflow, etc.); card files just declare
	// the strings.
	//
	// Empty / nil for cards with no printed combat keywords (the
	// S17 majority). Added in S18 sub-PR 2.
	PrintedKeywords []string

	// Triggered is the list of CR 603 triggered abilities the card
	// declares for the S19 auto-fire dispatcher. Each entry watches a
	// set of EventKinds, predicates whether a specific event triggers
	// the source, and builds the StackItem to enqueue onto
	// g.PendingTriggers — same queue the manual S13.1
	// AnnounceTrigger feeds. The harvester (game.triggerHarvester)
	// walks the battlefield + LKI map on every event emit and runs
	// each declared TriggeredAbility's Watches → AppliesTo → Build
	// pipeline.
	//
	// Empty / nil for cards with no auto-fire triggers (the S18
	// majority). Sub-PR 1 ships the dispatcher framework with zero
	// catalog declarations — sub-PRs 3-7 fill in cards. Manual
	// AnnounceTrigger remains for non-catalog cards and "hidden info"
	// triggers (cards in hand with cast-replacement triggers).
	//
	// Reaches directly into `game.TriggeredAbility` rather than a
	// per-package adapter type — same rationale as Static and
	// Replacements.
	//
	// Added in S19 sub-PR 1.
	Triggered []game.TriggeredAbility

	// Modes is the S20 sub-PR 4 modal-spell clause ("Choose one —").
	// Each option carries its oracle bullet and, when the bullet
	// targets, its own TargetSpec; the engine derives the cast's
	// effective target clause from the chosen options and the client
	// shows a mode picker before targeting. OnResolve reads the
	// choice back with ctx.HasMode(i). Build it with ChooseOne /
	// ChooseN in modes.go. Nil for non-modal cards.
	//
	// Sub-PR 4 limit: at most one chosen option may target, so a
	// "choose two" card may carry targets on at most one option
	// (Register panics otherwise). Per-mode target slots are the
	// deferred multi-target work.
	Modes *game.ModeSpec

	// AdditionalCost is the S21 sub-PR 5 "As an additional cost to
	// cast this spell, …" clause (CR 601.2f). Only the discard shape
	// exists today: Thrill of Possibility, Big Score, Unexpected
	// Windfall. The caster picks the cards in a client prompt that
	// opens BEFORE targeting, matching the order costs are paid in,
	// and the engine pays the cost with the spell already on the
	// stack so discard payoffs (Mary Read and Anne Bonny, Marauding
	// Mako) trigger above it. Build it with DiscardCost(n). Nil for
	// cards with no additional cost.
	AdditionalCost *game.AdditionalCost

	// AlternativeCosts is the S22 "you may cast this spell for its
	// overload / evoke / cleave cost" clause (CR 118.9) — a cost paid
	// INSTEAD of the mana cost, not alongside it like AdditionalCost.
	// Build the entries with Overload / Evoke / Cleave in
	// alternative_cost.go, never by hand: each keyword bundles a
	// text rewrite with its price, and the rewrite is the half a
	// card file would forget.
	//
	//	AlternativeCosts: []game.AlternativeCost{Overload("{4}{R}")},
	//
	// A slice because a card can offer more than one (spree, the
	// modal-cost cards). Nil for nearly every card. OnResolve reads
	// the choice back with ctx.PaidAltCost("overload").
	AlternativeCosts []game.AlternativeCost

	// TapCost is the S22 "tap permanents you control to help pay"
	// cost component — convoke (CR 702.51) and waterbend, which are
	// the same mechanic under two names. Unlike the other cost slots
	// this one does not add a demand, it SPENDS against one: each
	// permanent tapped pays for {1}, or (convoke only) for one mana
	// of that permanent's colour.
	//
	// Build it with Convoke() or Waterbend("{X}") in tap_cost.go,
	// never by hand: the keyword carries the pool of legal permanents
	// and the colour rule with it, and a card file that got the
	// colour rule wrong would ship a card stronger than printed.
	//
	//	TapCost: Convoke(),
	//	TapCost: Waterbend("{X}"),
	//
	// Nil for nearly every card. The caster's picks ride cast_spell
	// as `tap_ids`; tapping nothing is always legal.
	TapCost *game.TapPermanentsCost

	// Activated is the list of CR 602 activated abilities the card
	// offers from the battlefield — the fourth ability type, added
	// in S21 sub-PR 2. Each entry declares its cost (tap, sacrifice
	// this / sacrifice another matching a spec, mana, life), an
	// optional target clause validated like a spell's, and the
	// Effect that runs when the ability resolves off the stack.
	//
	// Mana abilities do NOT belong here — they don't use the stack
	// (CR 605.3a) and keep their own ManaAbilities slot.
	//
	// Build the entries with the constructors in activated.go:
	//
	//	Activated: []ActivatedAbility{{
	//		Label: "Sacrifice a creature: This enchantment deals 1 damage to any target.",
	//		Cost:  SacrificeACreature(),
	//		Targets: TargetAny(),
	//		Effect: ...,
	//	}},
	Activated []ActivatedAbility

	// NoMaxHandSize declares the printed static "You have no maximum
	// hand size" (Thought Vessel, Reliquary Tower, Spellbook,
	// Venser's Journal). True while the permanent is on the
	// battlefield; the controller skips the CR 402.2 cleanup-step
	// discard entirely.
	//
	// This is deliberately NOT a `Static` entry. Every other
	// continuous effect in the catalog modifies a characteristic of
	// an OBJECT, which is what the CR 613 layer engine models —
	// game.StaticAbility's Apply takes a *Characteristic and a
	// target *Card, and there is no seat in that signature. "You
	// have no maximum hand size" modifies a PLAYER, so it has no
	// characteristic to sit in and no layer to sit at.
	//
	// Rather than grow a parallel player-layer pipeline for one
	// clause, the engine DERIVES the answer: at cleanup it asks the
	// battlefield whether the active player controls any permanent
	// with this bit set (game.Game.EffectiveMaxHandSizeLocked, fed
	// by the game.CatalogNoMaxHandSize hook). Nothing is written to
	// Player.MaxHandSize, so nothing has to be restored when the
	// permanent leaves — which is what makes two copies, and one of
	// two leaving, come out right without any bookkeeping.
	//
	// Issue #338.
	NoMaxHandSize bool

	// Battle is a battle's printed battle data — its defense and its
	// subtype (CR 310). Nil for every card that is not a battle,
	// which is nearly all of them.
	//
	// A FALLBACK, like StartingLoyalty: the printed value on
	// game.Card.StartingDefense, stamped by the deck importer from
	// Scryfall's per-face `defense`, always wins. See BattleSpec in
	// battles.go for why it is declared anyway.
	//
	// Added in S27.
	Battle *BattleSpec
}

// ActivatedAbility is one activated ability on a permanent. Mirrors
// game.ActivatedAbilityShape; the wire hook converts. Added in S21
// sub-PR 2.
type ActivatedAbility struct {
	Label        string
	Cost         game.AbilityCost
	Targets      *game.TargetSpec
	SorcerySpeed bool
	Effect       func(g *game.Game, item *game.StackItem) error
}

// ManaAbility is one mana-producing activated ability on a permanent.
// Cost expresses what the controller pays to activate (today: tap +
// optional sacrifice; future: pay-life, sub-mana). Produced is the
// mana the ability adds to the controller's pool; uses the same
// brace-notation grammar the Scryfall mana_cost field does, plus a
// pipe (`|`) inside a single brace pair to mean "controller picks
// one of these colors" — Birds of Paradise prints `"{W|U|B|R|G}"`.
// Label is the menu copy the client renders ("Add {C}{C}",
// "Add one mana of any color"); empty falls back to a generated
// label.
type ManaAbility struct {
	Cost     ManaAbilityCost
	Produced string
	Label    string

	// Rider is everything the oracle text says AFTER the "Add …"
	// clause, as one callback: the painland cycle's "This land deals
	// 1 damage to you", Ancient Tomb's "deals 2 damage to you". It
	// runs immediately after the produced mana lands in the pool,
	// inside the same atomic mana-ability resolution (CR 605.3a).
	//
	// Build one with PainRider(n) rather than by hand — that helper
	// is the whole reason this slot exists so far.
	//
	// A rider is NOT a cost: it happens whether or not the player
	// could "afford" it, and a source with a damage rider stays
	// activatable at 1 life. Use ManaAbilityCost.Life for a real
	// cost ("{T}, Pay 1 life:").
	//
	// Runs under the resolution write lock — *ForEffect helpers
	// only, never a public locking mutator.
	//
	// Added in the S22 mana-ability-rider pass.
	Rider func(g *game.Game, controller, source uuid.UUID) error

	// IgnoreCommanderIdentity keeps a pipe-syntax Produced string
	// ("{U|R}", "{W|U|B|R|G}") at its printed width instead of
	// letting the engine intersect it with the controller's
	// commander colour identity.
	//
	// Set it whenever the printed text does not actually say "in
	// your commander's color identity" — City of Brass and Mana
	// Confluence ("any color"), the painland and Talisman duals
	// (two named colours). Leave it false for Command Tower,
	// Arcane Signet, Commander's Sphere and Path of Ancestry, whose
	// text is the reason the narrowing exists.
	//
	// Added in the S22 mana-ability-rider pass.
	IgnoreCommanderIdentity bool

	// ProducedFunc computes Produced at activation time instead of
	// declaring it. Two card families need it and they are the same
	// mechanism (#352 sub-gaps 3 and 4):
	//
	//   - DERIVED colours: Exotic Orchard ("one mana of any color
	//     that a land an opponent controls could produce"),
	//     Reflecting Pool, Fellwar Stone, Mox Amber. Return a pipe
	//     string — "{W|U|G}".
	//   - SCALED amounts: Cabal Coffers ("{B} for each Swamp you
	//     control"), Gaea's Cradle. Return the slot repeated —
	//     "{B}{B}{B}".
	//
	// Wins over Produced when non-nil. Returning "" adds no mana,
	// which is the printed behaviour when the derivation finds
	// nothing — Exotic Orchard on an empty opposing board, Gaea's
	// Cradle with no creatures. The source still taps.
	//
	// READ-ONLY, and it runs under g.mu: ActivateManaAbility holds
	// the write lock and the auto-tapper holds the read lock. Read
	// g.Battlefield and the *ForEffect accessors; a public locking
	// mutator deadlocks. Build one with ProducedFromLands,
	// ProducedRepeated or a sibling in mana_derivation.go rather
	// than by hand.
	ProducedFunc func(g *game.Game, controller, source uuid.UUID) string

	// Condition gates activation — "Activate only if you control
	// five or more lands" (Temple of the False God), "…three or
	// more artifacts" (Mox Opal). Checked before any cost is
	// validated, so a failed gate taps nothing and spends nothing
	// (CR 602.5a). Same read-only-under-the-lock contract as
	// ProducedFunc.
	//
	// Added in the S32 mana-pipeline pass (#352 sub-gap 5).
	Condition func(g *game.Game, controller, source uuid.UUID) bool

	// Restrictions are the "spend this mana only on …" tags stamped
	// onto every token this ability produces — Ancient Ziggurat,
	// Eldrazi Temple, Shrine of the Forsaken Gods, the coloured half
	// of Delighted Halfling. Build them with the game package's
	// ManaRestrict* constructors; the spend-time matcher lives in
	// game/mana_restriction.go.
	//
	// All tags must hold for the token to be spendable (AND), and an
	// unknown tag denies. A restricted ability drops out of auto-tap
	// planning — see autoTapAbilityFor.
	//
	// Declaring this WITHOUT the engine honouring it would ship
	// every card in the group stronger than printed, which is the
	// #259 rule; the two halves landed together in #352.
	Restrictions []string
}

// ManaAbilityCost names the activation cost of one mana ability.
// Tap is the canonical cost ({T}). Sacrifice is reserved for future
// mana rocks like Lotus Petal — declared on the struct so the wire
// shape is stable, NOT exercised by S15's catalog. Mana / life /
// counter sub-costs land with later sprints when a catalog card
// demands them.
type ManaAbilityCost struct {
	Tap bool
	// Sacrifice sacrifices the SOURCE (Treasure, Lotus Petal).
	Sacrifice bool
	// SacrificeOther sacrifices one OTHER permanent the activator
	// controls, matched against this spec — Ashnod's Altar's
	// "Sacrifice a creature: Add {C}{C}". Build it with the same
	// constructors an activated ability's cost uses
	// (SacrificeACreature().SacrificeOther), so the two ability
	// kinds share one clause vocabulary and one client picker.
	//
	// Added in the S21 mana-cost pass, which is also what closed
	// the S15 note that sacrifice costs were "reserved for future
	// mana rocks" — they are all live now.
	SacrificeOther *game.TargetSpec

	// Life is a "Pay N life" component of the activation cost (CR
	// 118.8) — Mana Confluence's "{T}, Pay 1 life: Add one mana of
	// any color". Mirrors game.AbilityCost.Life, which CR 602
	// activated abilities have carried since S21 and the fetchlands
	// already use.
	//
	// This is the S15 note above finally coming true: life was the
	// one "later sprints when a catalog card demands them" sub-cost
	// left, and the S22 mana-rider pass is the batch that demanded
	// it.
	//
	// A COST, not a rider: validated before anything is paid, so an
	// activation at a life total below N is rejected outright and
	// the source does not tap. "Add {R}. This land deals 1 damage to
	// you" is the other thing — see ManaAbility.Rider.
	Life int

	// Mana is a mana component of the activation cost — the Signet
	// cycle's "{1}, {T}: Add {W}{U}", Cabal Coffers' "{2}, {T}",
	// the filter lands' "{W/U}, {T}". Scryfall brace grammar.
	//
	// Paid out of the controller's pool before the source taps, and
	// validated with every other component first, so an unaffordable
	// activation fails with the source untouched. No auto-tap: the
	// player floats the mana first, which is how a Signet is played
	// on paper — see ActivateManaAbility.
	//
	// This is the last of S15's "mana / life / counter sub-costs
	// land with later sprints when a catalog card demands them"
	// note. #267 took life; #352 takes mana, and the counter case
	// still has no card asking for it.
	Mana string
}

// ZeroUUID is an alias for uuid.Nil. Mostly used in tests to
// distinguish "no target / self" fields from "uninitialised" —
// callers don't need this but having it available keeps the
// effects package self-contained for tests that don't import uuid.
var ZeroUUID = uuid.Nil
