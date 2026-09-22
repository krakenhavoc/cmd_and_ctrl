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
	// CR 708.2a, ADR 0069 decision 4: a face-down permanent has NO
	// TEXT — no triggered, activated, mana, static or replacement
	// abilities, no cost modifiers, no "as enters" hook, no printed
	// keywords, no catalog entry at all. catalogDef already reads the
	// empty key as "this card has no entry", and every Catalog*
	// reader already handles that, so the whole of CR 708.2a is this
	// one predicate at the one place a Card becomes a catalog key.
	//
	// It is HERE rather than in CatalogAbilityKey (the documented
	// "what does this permanent DO" accessor) because four ability
	// readers deliberately bypass that accessor, each for a good
	// reason — the layer pass's static gather, the LTB trigger's
	// last-known identity, HasKeyword, and the ETB hook — and each
	// would otherwise have needed its own face-down arm.
	//
	// Exile is deliberately NOT suppressed: FaceDownIsPermanent is
	// false for the two exile kinds, so a foretold card keeps the
	// CastableZones, AlternativeCosts and Targets its cast out of
	// exile needs (#658). Nothing off the battlefield runs a trigger,
	// static or replacement off a catalog entry (CR 113.6), so
	// keeping it costs nothing.
	if c.FaceDownIsPermanent() {
		return ""
	}
	base := c.OracleID
	// #521: a TOKEN has no printing and therefore no oracle ID, so
	// its printed abilities are registered under a synthetic token
	// key instead (token_key.go). The oracle ID WINS when there is
	// one, which is CR 707.2: a token copy carries the copied card's
	// oracle ID and must keep resolving to that card's entry, not to
	// a token's.
	if base == "" {
		base = c.TokenKey
	}
	if c.ActiveFace != 0 && c.OracleID != "" {
		base = c.OracleID + "#" + strconv.Itoa(c.ActiveFace)
	}
	// CR 707.9a, #665: an ability a copy effect GRANTED is part of
	// the object's copiable values and has no oracle ID of its own,
	// so it rides in the key. Every card in the game but a granted
	// copy takes the nil-slice fast path and gets the bare key back
	// unchanged; catalogDef answers the composite one by merging the
	// grant's abilities into the card's. See copy_grants.go.
	if len(c.GrantedAbilities) == 0 {
		return base
	}
	return catalogKeyWithGrants(base, c.GrantedAbilities)
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
	// SacrificeOther sacrifices OTHER permanents the activator
	// controls, matched against this spec — Ashnod's Altar's
	// "Sacrifice a creature". The activator names them in
	// ManaAbilityParams.SacrificeIDs; the spec's Min == Max is how
	// many (#747), 1 for every one-permanent constructor.
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

	// TapOthers taps OTHER untapped permanents the activator
	// controls as part of the cost — Springleaf Drum's "{T}, Tap an
	// untapped creature you control", Jaspera Sentinel's "{T}, Tap
	// another untapped creature you control", Heritage Druid's
	// three Elves. Nil means no such component.
	//
	// The SAME component AbilityCost.TapOthers carries (#758), with
	// the same validator and the same payer, because it is the same
	// cost — a card that printed it on a mana ability and on a
	// CR 602 ability would be paying one clause two ways otherwise.
	// The activator names the permanents in
	// ManaAbilityParams.TapIDs.
	//
	// The auto-tapper never plans a source whose mana ability
	// carries one (autoTapAbilityFor): tapping a creature the
	// player was keeping back to block is a decision the planner
	// cannot weigh, the same bar a life cost fails.
	TapOthers *TapOthersCost

	// LifeCost is a life component in the activation cost (CR
	// 119.4) — Mana Confluence's "{T}, Pay 1 life: Add one mana of
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
	// resolves with no priority window (CR 605.3b) and the player
	// has to have floated the mana deliberately.
	//
	// This closes the last of S15's "mana / life / counter
	// sub-costs land with later sprints" note (#352 sub-gap 1).
	ManaCost string

	// RemoveCounters is a "remove N counters" component of the
	// activation cost — Vivid Creek's "{T}, Remove a charge counter
	// from this land", Ramos's "Remove five +1/+1 counters", Mage-
	// Ring Network's "Remove any number of storage counters" (#789).
	//
	// It is the SAME type an activated ability's cost carries
	// (AbilityCost.RemoveCounters), not a parallel one, and that is
	// the whole design: one CounterRemovalCost, two owners, so the
	// validator (validateCounterRemovalLocked), the candidate walk
	// (CounterCostOptionsForEffect), the move enumerator, the
	// protocol view and the client's picker are each written once.
	// The S15 note above finally has no "counter" left in it.
	//
	// Validated with every other component before any is paid, and
	// paid after the tap and the life, so an activation that cannot
	// pay fails with the source still untapped.
	//
	// The AUTO-TAPPER only plans a source whose counter cost it can
	// both decide and pay: the self form, a printed kind, a fixed
	// count, and enough counters on the permanent right now. Every
	// other shape is a decision, and the planner makes none — see
	// autoTapAbilityFor and manaCounterCostPlannable.
	RemoveCounters *CounterRemovalCost

	// AddCounter is a cost that puts a counter on the source. No
	// printed mana ability has one today; the slot exists because
	// the component is declared once and owned by both ability
	// kinds, and leaving it off here would mean a second place that
	// has to learn about counter costs later. Auto-tap never plans
	// an ability that has one.
	AddCounter *CounterAddCost

	// DiscardCards is a "discard N cards" component of the activation
	// cost — Skirge Familiar's "Discard a card: Add {B}" (#1213).
	//
	// The SAME game.DiscardCost an activated ability's cost carries
	// (#660), with the same options walk, the same validator and the
	// same payer, for the reason SacrificeOther, RemoveCounters and
	// TapOthers above are each one struct with two owners: a clause
	// declared twice is a clause that can be paid two ways.
	//
	// It pays through the ONE discard door (discardCardsLocked with
	// DiscardCauseCost), so EventDiscardCard fires once per card, the
	// CR 614 window runs over the exit and madness (CR 702.35a) sees
	// it — without any of them learning that mana abilities exist.
	// CR 601.2h's indivisible step is expressed by
	// zoneRoute.MustSettleNow, as it is for every other cost discard.
	//
	// The activator names the cards in ManaAbilityParams.DiscardIDs.
	// There is no DiscardSelf twin: that component discards the
	// SOURCE (cycling, CR 702.29a) and a mana ability's source is a
	// permanent, which is not in a hand to discard.
	//
	// The AUTO-TAPPER never plans an ability that has one, the same
	// bar the life cost and the tap-others cost fail: which card to
	// pitch is a decision, and the planner makes none. A hand-clicked
	// Skirge Familiar is a mana source; an auto-tapped one is not.
	DiscardCards *DiscardCost

	// ProducedForPaid computes the produced-mana string from what
	// the cost actually PAID, for an ability whose output the
	// printed text derives from the payment rather than from the
	// board (#789):
	//
	//   - Mage-Ring Network, "Add {C} for each storage counter
	//     removed this way" — returns "{C3}" for three.
	//   - Crucible of the Spirit Dragon's "Add X mana", and any
	//     future "for each counter / each life paid" clause.
	//
	// Wins over ProducedFunc, which wins over Produced. Returning ""
	// produces no mana, which is the printed behaviour for a variable
	// removal that removed nothing.
	//
	// The `paid` record is the SAME PaidCost a stack item carries
	// (paid_cost.go); a mana ability has no stack item to hang it on
	// (CR 605.3b), so it is handed over directly and lives only for
	// the length of the activation.
	//
	// Same locking contract as Condition and ProducedFunc: read-only,
	// under g.mu. CR 106.7's "could produce" reader evaluates it with
	// the LARGEST record the source could pay right now, because
	// that is what "could" means — a Mage-Ring Network with three
	// storage counters could produce {C}, and one with none could
	// not.
	ProducedForPaid func(g *Game, controller, source uuid.UUID, paid PaidCost) string

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

	// DerivesFromOtherSources marks a ProducedFunc that asks OTHER
	// permanents what THEY could produce — Exotic Orchard,
	// Reflecting Pool, Fellwar Stone. It is the recursion guard:
	// ProducibleManaLocked (CR 106.7) evaluates every other
	// ProducedFunc and skips these, because two Exotic Orchards
	// facing each other would otherwise recurse until the stack ran
	// out. CR 106.6b answers the circular case with "no mana" and so
	// does the guard.
	//
	// A declaration rather than something inferred at run time: the
	// alternative is a re-entrancy counter, which on a snapshotted
	// struct is undo state nobody wants and off it is a data race
	// between two games in one process.
	//
	// Added in S44 (#782).
	DerivesFromOtherSources bool

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

	// Exhaust marks an EXHAUST mana ability (#1183):
	//
	//	Exhaust — {G}, {T}: Add three mana of any one color.
	//	(Activate each exhaust ability only once.)
	//
	// The twin of ActivatedAbilityShape.Exhaust, reading the same
	// record through the same key, and one declarative bit for the
	// same reason: the keyword IS the rule. Loot, the Pathfinder is
	// the one printed card, and prints it three times — once on a mana
	// ability and twice on ordinary ones.
	//
	// Keyed by the LABEL (activation_tally.go), so a mana ability that
	// sets this must have one; effects.Register refuses a blank label
	// and a duplicate at boot, across BOTH ability lists, because the
	// two share one key space on one object.
	//
	// The extra reader a mana ability has is the AUTO-TAPPER. A spent
	// exhaust ability is not a mana source — gatherTapSources will not
	// plan it and materializePlanLocked will not tap it — because a
	// planner that spent one behind the player's back would be taking
	// the ability away to pay for something the player was not asked
	// about. ProducibleManaLocked answers CR 106.7 the same way, so a
	// Reflecting Pool is not priced on mana the spent source can never
	// make again.
	Exhaust bool

	// Rider is the post-production half of a mana ability whose
	// oracle text continues past the "Add …" clause — the painland
	// cycle's "This land deals 1 damage to you", Ancient Tomb's 2.
	// It runs immediately after the produced mana lands in the
	// controller's pool, in printed order, as part of the same
	// atomic mana-ability resolution (CR 605.3b — no stack, no
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

	// NarrowToCommanderIdentity intersects a multi-option ("pipe")
	// produced string with the controller's commander's colour
	// identity before the pick is offered.
	//
	// Set it ONLY when the printed text says "any color in your
	// commander's color identity": Command Tower, Arcane Signet,
	// Commander's Sphere, Path of Ancestry. Every other multi-option
	// source — Birds of Paradise, Treasure, City of Brass, the
	// painlands and guildgates — keeps its printed option set, with
	// the identity's colours listed first (owner decision 2026-09-17;
	// see manaPickOptionsFor).
	//
	// CR 903.4f (#844): with no commander, or a commander whose colour
	// identity is colourless, the intersection is empty and the ability
	// adds NO mana — no pick is offered, and nothing enumerates the
	// activation (ManaAbilityAddsNoMana).
	//
	// Replaced IgnoreCommanderIdentity, its inverse, when the default
	// flipped from narrowing to ordering.
	NarrowToCommanderIdentity bool
}

// CatalogWantsDistinctColors reports whether a spell READS the
// colours of the mana that paid for it — converge (CR 702.86),
// sunburst (CR 702.44), and nothing else today. It is the switch that
// picks the colour-maximising payment strategy at the cast gate
// (#761, ADR 0068 §5).
//
// A declaration on effects.Spec rather than something inferred from
// the oracle text, for the reason DerivesFromOtherSources is one: a
// text scan quietly stops matching when a card words the clause
// differently, and the failure is silent and stronger-than-nothing in
// the wrong direction (a converge spell that counts one colour).
//
// Nil hook ⇒ no catalog wired ⇒ every payment keeps the default
// colourless-first order, which is what the game package's own tests
// expect.
var CatalogWantsDistinctColors func(oracleID string) bool

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

// CatalogTriggerDoublers returns the CR 603.2d doublers declared by a
// catalog card. It is a separate slot so game-package tests can stub the
// count without importing the effects package.
var CatalogTriggerDoublers func(oracleID string) []TriggerDoubler

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
	// S27: a Saga enters with a lore counter (CR 714.3). Same
	// reasoning as the loyalty stamp above, and the same placement:
	// it is printed-rules behaviour keyed on the card's subtype, so
	// it runs before the oracle-ID / nil-hook guards and applies to
	// Sagas the catalog has never heard of.
	g.sagaEntersWithLoreCounterLocked(cardID)

	// S27: a battle enters with its printed defense counters and
	// chooses a protector (CR 310.4, 310.9a). Same placement and same
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
// loyalty doubled (CR 122.6), which a direct map write would skip.
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
