package game

import (
	"strconv"

	"github.com/google/uuid"
)

// CatalogKey returns the effect-catalog key for a card's ACTIVE
// face (ADR 0034 §5).
//
// Scryfall issues one oracle_id per CARD, not per face, so the two
// halves of Sea Gate Restoration collide onto a single spec slot —
// and effects.Register PANICS on a duplicate oracle ID, deliberately,
// because a collision otherwise means one card silently shadowing
// another. Rather than weaken that panic, the key becomes composite:
//
//	face 0  →  "<oracle_id>"        (bare — unchanged)
//	face N  →  "<oracle_id>#N"
//
// Face 0 keeping the bare ID is what makes this a no-op for all ~241
// registered specs and every single-faced card in the game. A back
// face registers under "<oracle_id>#1" and cannot collide with
// anything, so Register's duplicate check stays exactly as strict as
// it was — it simply now has a second, distinct key to reject
// duplicates within.
//
// Register, Lookup, Has and Spec are untouched: they already take and
// hold opaque strings. The work was swapping the ~25 production call
// sites from a bare c.OracleID to CatalogKey(c), which is mechanical
// — and because CastSpell has already called SetFace by the time any
// of them run, announce-time and battlefield-time hooks both resolve
// to the correct half with no further plumbing.
//
// The cost, stated plainly: a call site that FORGETS CatalogKey
// silently resolves to face 0's spec rather than erroring.
func CatalogKey(c Card) string {
	if c.ActiveFace == 0 || c.OracleID == "" {
		return c.OracleID
	}
	return c.OracleID + "#" + strconv.Itoa(c.ActiveFace)
}

// CatalogKeyForFace is CatalogKey for a face other than the one
// currently active — used by the legal-move enumerator, which has to
// price BOTH halves of a modal DFC without mutating the card.
func CatalogKeyForFace(oracleID string, face int) string {
	if face == 0 || oracleID == "" {
		return oracleID
	}
	return oracleID + "#" + strconv.Itoa(face)
}

// effect_hooks.go holds the per-slot function variables the engine
// reads the catalog through. Since #622 the catalog sets none of them:
// carddef.go gives each a default that reads its slot off the one
// precomputed CardDef, and they remain as per-slot test seams. The
// history below is kept for the shape of the boundary.
//
// Originally: the function-variable slots that the S14
// card-effect catalog populates from its own init() block. The
// game package can't directly import server/internal/cards/effects
// (effects needs to reach into *Game for mutations, so that
// direction is the one that builds cleanly). Dependency inversion
// via exported `var` callbacks breaks the cycle:
//
//   game  ──  EffectResolver / ETBEffectHook / IsCatalogCard  ──┐
//                                                                │ populates at init
//   effects (imports game) ───────────────────────────────────────┘
//
// main.go blank-imports effects so the init fires at server boot.
// With no import, the hooks stay nil and the resolution path falls
// back to today's manual sandbox behaviour — the catalog is opt-in
// at the server build level too.
//
// Added in S14 sub-PR 3.

// EffectResolver is invoked from resolveTopOfStackLocked after the
// target re-check and before zone routing. Implementations look up
// the spell's oracle ID (stable across printings) in the catalog
// and run the registered OnResolve callback if present. Nil means
// "no catalog wired" — the resolution path skips the call and
// relies on the manual sandbox behaviour.
//
// Implementations MUST NOT take g.mu — the resolution path already
// holds it. Use *ForEffect helpers on *Game.
var EffectResolver func(g *Game, item *StackItem, oracleID string) error

// ETBEffectHook fires after a permanent crosses into the
// battlefield from any source (land cast, spell resolution,
// MoveCardByID into battlefield). Implementations look up the
// card's oracle ID in the catalog and run the registered AsEnters
// callback + stamp StartingLoyalty for planeswalkers.
var ETBEffectHook func(g *Game, cardID uuid.UUID, oracleID string) error

// CatalogStartingLoyalty returns the catalog's declared starting
// loyalty for an oracle ID, or 0 when the card has no catalog entry
// (or is not a planeswalker). It is a FALLBACK only: the printed
// value on Card.StartingLoyalty, stamped by the deck importer from
// Scryfall, always wins. The catalog answer exists for cards that
// never went through deck import — tokens, test fixtures, and the
// demo seed.
//
// Nil hook means "no catalog wired", which is the normal state in
// the game package's own tests and in a server built without the
// effects blank import. Loyalty still works in that configuration;
// that is the point of moving the stamp here (issue #274).
var CatalogStartingLoyalty func(oracleID string) int

// IsCatalogCard reports whether an oracle ID is present in the
// card-effect catalog. Used by the view layer (protocol.CardView)
// to stamp the `auto` bit so the client can render the auto badge.
// Nil is treated as "no catalog wired" → every card looks manual.
var IsCatalogCard func(oracleID string) bool

// CatalogTargetMode returns the registered card's announce-time
// target prompt shape (see effects.Spec.TargetMode) or empty
// string when no catalog entry matches. Nil hook always returns
// empty. Serialised onto CardView.TargetMode for the client's
// cast-targeting UI.
var CatalogTargetMode func(oracleID string) string

// ManaAbilityShape is the minimal mana-ability surface the game
// package consumes. Mirrors effects.ManaAbility but lives in `game`
// to avoid an import cycle (the effects package already imports
// `game`). The S15 dispatcher reads this on every
// activate_mana_ability call to look up the ability's cost shape +
// produced-mana string.
type ManaAbilityShape struct {
	TapCost bool
	// SacrificeCost sacrifices the SOURCE as part of the cost
	// (Treasure, Lotus Petal, an Eldrazi Spawn).
	SacrificeCost bool
	// SacrificeOther sacrifices one OTHER permanent the activator
	// controls, matched against this spec — Ashnod's Altar's
	// "Sacrifice a creature". The activator names it in
	// ManaAbilityParams.SacrificeIDs.
	//
	// Distinct from SacrificeCost because the two compose: a card
	// could in principle eat itself and something else. The source
	// is a legal choice when the spec admits it, exactly as on an
	// activated ability's SacrificeOther (Carrion Feeder eats
	// itself), so validation rejects naming the same permanent
	// twice rather than assuming they differ.
	//
	// Added in the S21 mana-cost pass.
	SacrificeOther *TargetSpec

	// LifeCost is a life component in the activation cost (CR
	// 118.8) — Mana Confluence's "{T}, Pay 1 life: Add one mana of
	// any color". Mirrors AbilityCost.Life, which CR 602 activated
	// abilities have carried since S21. Validated before anything
	// is paid and paid after the tap, so an attempt at too low a
	// life total fails without tapping the source.
	//
	// A life COST is not the same thing as a Rider that loses life:
	// a cost is checked and paid up front and makes the ability
	// unactivatable when it can't be met, while a rider is part of
	// the ability's effect and happens no matter what. Ancient Tomb
	// ("Add {C}{C}. This land deals 2 damage to you") is a rider and
	// can be activated at 1 life; Mana Confluence is a cost and
	// cannot be activated at 0.
	//
	// Added in the S22 mana-ability-rider pass.
	LifeCost int

	// ManaCost is a mana component in the activation cost — the
	// Signet cycle's "{1}, {T}: Add {W}{U}", Cabal Coffers' "{2},
	// {T}". Scryfall brace grammar, parsed with ParseCost.
	//
	// Paid out of the controller's pool BEFORE the source taps, and
	// validated alongside every other component first, so a Signet
	// activated on an empty pool fails without tapping. There is no
	// auto-tap here: ActivateManaAbility will not tap other
	// permanents to fund a mana ability, because a mana ability
	// resolves with no priority window (CR 605.3a) and the player
	// has to have floated the mana deliberately.
	//
	// This closes the last of S15's "mana / life / counter
	// sub-costs land with later sprints" note (#352 sub-gap 1).
	ManaCost string

	// Condition gates activation — "Activate only if you control
	// five or more lands" (Temple of the False God), "Activate only
	// if you control three or more artifacts" (Mox Opal). Checked
	// before any cost is validated or paid; a false return is
	// ErrConditionNotMet and nothing is spent or tapped.
	//
	// READ-ONLY and runs under g.mu, which ActivateManaAbility
	// holds for write and the auto-tapper holds for read. Inspect
	// g.Battlefield / *ForEffect accessors; a public locking
	// mutator deadlocks.
	//
	// Nil means "no gate", which is nearly every mana ability.
	//
	// Added in the S32 mana-pipeline pass (#352 sub-gap 5).
	Condition func(g *Game, controller, source uuid.UUID) bool

	// ProducedFunc computes the produced-mana string at activation
	// time, for abilities whose output the printed text derives
	// from the board rather than naming:
	//
	//   - DERIVED colours — Exotic Orchard ("any color that a land
	//     an opponent controls could produce"), Reflecting Pool,
	//     Fellwar Stone, Mox Amber. Returns a pipe string.
	//   - SCALED amounts — Cabal Coffers ("{B} for each Swamp you
	//     control"), Gaea's Cradle. Returns the slot repeated.
	//
	// Wins over Produced when non-nil. Returning "" produces no
	// mana at all, which is the printed behaviour for Exotic
	// Orchard with no opponent lands and for Gaea's Cradle with no
	// creatures — the ability is still activatable and the source
	// still taps.
	//
	// Same locking contract as Condition: read-only, under g.mu.
	//
	// Added in the S32 mana-pipeline pass (#352 sub-gaps 3 and 4).
	ProducedFunc func(g *Game, controller, source uuid.UUID) string

	// Restrictions are the tags stamped onto every ManaToken this
	// ability produces — "spend this mana only to cast a creature
	// spell" (Ancient Ziggurat), "only to cast colorless Eldrazi
	// spells or activate abilities of colorless Eldrazi" (Eldrazi
	// Temple). Build them with the ManaRestrict* constructors in
	// mana_restriction.go, which is also where the spend-time
	// matching lives.
	//
	// Empty for ordinary mana. A non-empty list makes this ability
	// invisible to the auto-tapper (autoTapAbilityFor) — restricted
	// mana is a decision the planner cannot make for the player.
	//
	// Added in the S32 mana-pipeline pass (#352 sub-gap 2).
	Restrictions []string

	// RestrictionsFunc computes the spend restrictions at activation
	// time, for an ability whose restriction names something chosen
	// rather than printed: Cavern of Souls' "spend this mana only to
	// cast a creature spell of THE CHOSEN TYPE".
	//
	// Wins over Restrictions when non-nil. Returning nil produces
	// UNRESTRICTED mana, so a card whose restriction cannot be
	// computed yet must return an impossible tag rather than nothing
	// — that is the #259 direction, and the one asymmetry in this
	// slot worth stating out loud. Cavern of Souls with no type named
	// yet returns a tag naming the empty tribe, which the matcher
	// refuses, so the mana is unspendable rather than free.
	//
	// Same locking contract as Condition and ProducedFunc: read-only,
	// under g.mu. Added in S26.
	RestrictionsFunc func(g *Game, controller, source uuid.UUID) []string

	Produced string
	Label    string

	// Rider is the post-production half of a mana ability whose
	// oracle text continues past the "Add …" clause — the painland
	// cycle's "This land deals 1 damage to you", Ancient Tomb's 2.
	// It runs immediately after the produced mana lands in the
	// controller's pool, in printed order, as part of the same
	// atomic mana-ability resolution (CR 605.3a — no stack, no
	// priority window in between).
	//
	// `source` is the activating permanent's instance ID and may
	// already have left the battlefield when the ability also had a
	// sacrifice cost, so the rider gets the ID rather than a *Card
	// and must not assume the permanent is still there.
	//
	// Runs under g.mu held in write mode (ActivateManaAbility holds
	// it): use *ForEffect helpers only, never a public locking
	// mutator. A rider that changes life totals does NOT need to run
	// its own state-based-action pass — ActivateManaAbility runs one
	// on the way out whenever a rider fired.
	//
	// Nil for the overwhelming majority of mana abilities.
	//
	// Added in the S22 mana-ability-rider pass.
	Rider func(g *Game, controller, source uuid.UUID) error

	// IgnoreCommanderIdentity opts a multi-option ("pipe") produced
	// string out of the commander-identity narrowing that
	// ActivateManaAbility otherwise applies to every such slot.
	//
	// That narrowing exists for Arcane Signet and Command Tower,
	// whose printed text really does say "in your commander's color
	// identity". City of Brass and Mana Confluence say "any color"
	// flatly, and the painland / Talisman duals name two specific
	// colors — none of them should be narrowed. Setting this keeps
	// the printed option set intact.
	//
	// Added in the S22 mana-ability-rider pass.
	IgnoreCommanderIdentity bool
}

// CatalogManaAbilities returns the registered mana abilities for
// the given oracle ID (one entry per `effects.Spec.ManaAbilities`
// element), or nil when no catalog entry exists / the entry has no
// mana abilities. The cards/effects package populates this hook at
// init time alongside EffectResolver / ETBEffectHook. Nil hook ⇒
// engine falls back to the synthetic basic-land ability path.
//
// Added in S15 sub-PR 2.
var CatalogManaAbilities func(oracleID string) []ManaAbilityShape

// CatalogStaticAbilities returns the registered static abilities
// for the given oracle ID (one entry per `effects.Spec.Static`
// element), or nil when no catalog entry exists / the entry has no
// statics. The cards/effects package populates this hook at init
// time alongside the other catalog hooks. Nil hook ⇒ no card has a
// static ability ⇒ the layer engine's recompute pass leaves
// effective characteristics equal to printed.
//
// Used by Game.activeStaticAbilitiesLocked at every recompute pass
// (snapshot-driven, lazy via the layerVersion counter). Added in
// S16 sub-PR 3.
var CatalogStaticAbilities func(oracleID string) []StaticAbility

// CatalogReplacements returns the registered replacement effects
// for the given oracle ID (one entry per
// `effects.Spec.Replacements` element), or nil when no catalog
// entry exists / the entry has no replacements. The cards/effects
// package populates this hook at init time alongside the other
// catalog hooks. Nil hook ⇒ no card has a replacement effect ⇒
// gatherActiveReplacementsLocked skips the catalog leg and returns
// only built-ins + test replacements.
//
// Used by Game.gatherActiveReplacementsLocked at every apply-loop
// iteration. Unlike static abilities, replacement effects fire
// PRE-event — the engine walks the battlefield to find applicable
// effects BEFORE any rule-visible mutation runs. See
// replacements.go. Added in S17 sub-PR 2.
var CatalogReplacements func(oracleID string) []ReplacementEffect

// CatalogPrintedKeywords returns the printed keyword list for the
// given oracle ID (one entry per `effects.Spec.PrintedKeywords`
// element), or nil when no catalog entry exists / the entry has no
// printed keywords. Populated at init time by the cards/effects
// package alongside the other catalog hooks.
//
// Two consumers:
//   - On-battlefield: wire.go synthesizes a self-only Layer 6
//     StaticAbility that appends each entry to the card's own
//     Characteristic.Abilities, so HasKeyword naturally picks them
//     up via Effective().Abilities.
//   - Off-battlefield: HasKeyword (keywords.go) reads directly from
//     this hook when the card has no `effective` cache (e.g. a card
//     in hand needs flash gating before cast). The layer engine
//     only maintains Effective() for battlefield cards.
//
// Nil hook ⇒ no catalog wired ⇒ off-battlefield keyword reads
// return false for every card. Added in S18 sub-PR 2.
var CatalogPrintedKeywords func(oracleID string) []string

// CatalogTriggers returns the registered triggered abilities for the
// given oracle ID (one entry per `effects.Spec.Triggered` element),
// or nil when no catalog entry exists / the entry has no triggers.
// The cards/effects package populates this hook at init time
// alongside the other catalog hooks. Nil hook ⇒ no card has an
// auto-fire trigger ⇒ the S19 harvester returns immediately on
// every event (the layerVersionBump path keeps running normally).
//
// Used by triggerHarvester.OnEvent in triggers.go on every event
// emit. Per-event walk is linear in battlefield size; the hook is
// the inner-loop oracle lookup that happens per card. Added in S19
// sub-PR 1.
var CatalogTriggers func(oracleID string) []TriggeredAbility

// CatalogNoMaxHandSize reports whether the given oracle ID is a
// permanent whose controller has no maximum hand size (Thought
// Vessel, Reliquary Tower, Spellbook, Venser's Journal). Populated
// at init time by the cards/effects package from
// `effects.Spec.NoMaxHandSize`. Nil hook ⇒ no catalog wired ⇒ every
// player keeps whatever Player.MaxHandSize says.
//
// This is the one continuous effect in the catalog that is
// PLAYER-scoped rather than card-scoped. The layer engine (CR 613)
// only models characteristics of objects, so there is no
// characteristic for "you have no maximum hand size" to modify and
// no layer for it to sit in. Rather than invent a player-layer
// pipeline for a single clause, the value is DERIVED: the cleanup
// step asks the battlefield at the moment it needs an answer (see
// Game.EffectiveMaxHandSizeLocked).
//
// Deriving instead of writing to Player.MaxHandSize is what makes
// the leave case correct for free. A "set on enter, restore on
// leave" design has to answer "restore to what?", and gets two
// things wrong that a real game hits: two Thought Vessels, where the
// first to leave would restore the cap while the second is still
// out; and a player whose maximum was changed by something else in
// between, whose real value the restore would clobber. Derivation
// has no stored value to strand, so neither case exists. It also
// keeps undo correct without touching clone.go — there is no new
// state to clone.
//
// Issue #338.
var CatalogNoMaxHandSize func(oracleID string) bool

// EffectiveMaxHandSizeLocked returns the hand-size cap that actually
// applies to `p` right now (CR 402.2): NoMaxHandSize when the player
// controls any battlefield permanent granting "you have no maximum
// hand size", otherwise the player's own Player.MaxHandSize.
//
// Player.MaxHandSize remains the BASE value and is never written by
// this path, so the set_max_hand_size sandbox action and a Thought
// Vessel compose the obvious way: the permanent wins while it is
// there, and the base value is untouched underneath it.
//
// Caller must hold g.mu (read or write).
func (g *Game) EffectiveMaxHandSizeLocked(p *Player) int {
	if p == nil {
		return DefaultMaxHandSize
	}
	if p.MaxHandSize == NoMaxHandSize {
		return NoMaxHandSize
	}
	if CatalogNoMaxHandSize == nil {
		return p.MaxHandSize
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != p.ID || c.OracleID == "" {
			continue
		}
		// CatalogAbilityKey: "you have no maximum hand size" is a
		// static ability, and a Thought Vessel that has lost all its
		// abilities gives the cap back.
		if CatalogNoMaxHandSize(CatalogAbilityKey(*c)) {
			return NoMaxHandSize
		}
	}
	return p.MaxHandSize
}

// fireEffectResolverLocked invokes the registered EffectResolver
// if non-nil, emits EventEffectError on failure, and swallows the
// error so the resolution path keeps moving. Caller must hold g.mu.
func (g *Game) fireEffectResolverLocked(item *StackItem, oracleID string, cardID uuid.UUID) {
	if EffectResolver == nil || oracleID == "" {
		return
	}
	if err := EffectResolver(g, item, oracleID); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   cardID,
			ErrorMsg: err.Error(),
		})
	}
}

// fireETBHookLocked runs the two things that must happen whenever a
// card crosses onto the battlefield: the printed starting-loyalty
// stamp (CR 306.5b) and the registered ETBEffectHook. Used by every
// code path that places a card on the battlefield — land cast,
// spell resolution, reanimation, token creation, manual admin move.
// Caller must hold g.mu.
//
// The loyalty stamp runs FIRST and runs unconditionally — before
// the oracle-ID / nil-hook guards below. Issue #274: it used to
// live inside the catalog's hook (the catalog's as-enters wrapper), so it was
// skipped for any card the catalog didn't know, and the CR 704.5i
// SBA then swept the 0-loyalty planeswalker into the graveyard on
// the next priority-grant boundary.
func (g *Game) fireETBHookLocked(cardID uuid.UUID, oracleID string) {
	g.stampStartingLoyaltyLocked(cardID, oracleID)
	// S27: a Saga enters with a lore counter (CR 714.2b). Same
	// reasoning as the loyalty stamp above, and the same placement:
	// it is printed-rules behaviour keyed on the card's subtype, so
	// it runs before the oracle-ID / nil-hook guards and applies to
	// Sagas the catalog has never heard of.
	g.sagaEntersWithLoreCounterLocked(cardID)

	// S27: a battle enters with its printed defense counters and
	// chooses a protector (CR 310.4, 310.5). Same placement and same
	// reasoning as the loyalty stamp above: printed rules keyed on
	// the card's type, so it runs before the oracle-ID / nil-hook
	// guards and applies to battles the catalog has never heard of.
	g.stampBattleEntryLocked(cardID, oracleID)
	if ETBEffectHook == nil || oracleID == "" {
		return
	}
	if err := ETBEffectHook(g, cardID, oracleID); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   cardID,
			ErrorMsg: err.Error(),
		})
	}
}

// stampStartingLoyaltyLocked puts a planeswalker's starting loyalty
// counters on it as it enters the battlefield (CR 306.5b). Caller
// must hold g.mu.
//
// Source precedence:
//
//  1. Card.StartingLoyalty — printed data, stamped by the deck
//     importer from Scryfall's `loyalty` field. This is the normal
//     path for every card a player actually brought to the game.
//  2. CatalogStartingLoyalty(oracleID) — the effects catalog's
//     Spec.StartingLoyalty, for cards that never went through deck
//     import (tokens, fixtures, the demo seed).
//
// The already-has-loyalty guard is what keeps the two from adding
// up: a deck-imported catalog planeswalker would otherwise enter
// with double its printed loyalty. It also makes the call idempotent
// for the entry paths that fire the hook more than once.
//
// Counters go through AddCounterForEffect rather than being written
// directly so the CR 614 counter-replacement pipeline still sees
// them — a planeswalker entering under Doubling Season gets its
// loyalty doubled (CR 121.3), which a direct map write would skip.
//
// Timing matters: this runs as part of the battlefield-entry path,
// which is strictly before the next priority-grant boundary, so the
// 0-loyalty SBA at mutations.go:1836 never observes the walker in
// its unstamped state.
func (g *Game) stampStartingLoyaltyLocked(cardID uuid.UUID, oracleID string) {
	var card *Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			card = &g.Battlefield.Cards[i]
			break
		}
	}
	if card == nil || !card.IsPlaneswalker() {
		return
	}
	if card.Counters[CounterLoyalty] > 0 {
		return
	}
	loyalty := card.StartingLoyalty
	if loyalty <= 0 && CatalogStartingLoyalty != nil && oracleID != "" {
		loyalty = CatalogStartingLoyalty(oracleID)
	}
	if loyalty <= 0 {
		// Nothing printed and nothing in the catalog. Leave it at
		// zero and let CR 704.5i do its job — inventing a loyalty
		// value here would make planeswalkers unkillable.
		return
	}
	if err := g.AddCounterForEffect(cardID, CounterLoyalty, loyalty); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   cardID,
			ErrorMsg: err.Error(),
		})
	}
}

// IsAutoCard is the exported wrapper for the view layer. Returns
// true when the card's oracle ID is in the catalog. Called from
// protocol.viewOfCard when stamping CardView.Auto. Nil hook returns
// false — no catalog means no auto.
func IsAutoCard(oracleID string) bool {
	if IsCatalogCard == nil || oracleID == "" {
		return false
	}
	return IsCatalogCard(oracleID)
}

// TargetModeFor returns the catalog's declared target prompt mode
// for the given oracle ID, or empty string if the card isn't in
// the catalog / has no target prompt. Called from protocol.viewOfCard
// when stamping CardView.TargetMode.
func TargetModeFor(oracleID string) string {
	if CatalogTargetMode == nil || oracleID == "" {
		return ""
	}
	return CatalogTargetMode(oracleID)
}
