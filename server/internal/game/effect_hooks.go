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
func CatalogKey(c Card) string { return catalogKeyOf(&c) }

// catalogKeyOf is CatalogKey without the copy: Card is over a
// kilobyte, and copying one per catalog lookup was most of
// CatalogAbilityKey's flat time in a bot table's profile (#1498).
// Internal callers that hold a *Card use this; CatalogKey stays the
// exported by-value shape every catalog file and open branch calls.
func catalogKeyOf(c *Card) string {
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
	if c.faceDownPermanent() {
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
	// ADR 0103: a Room's two halves are ONE catalog entry, told apart
	// by the door gate rather than by the key (Option 2B), so a Room
	// keys on its bare oracle ID even while its right half is on the
	// stack. A FUSED split spell is both halves at once and keys on the
	// synthetic FusedCatalogKey, which catalogDef builds from the two.
	switch {
	case c.Fused && c.OracleID != "":
		base = FusedCatalogKey(c.OracleID)
	case c.ActiveFace != 0 && c.OracleID != "" && !hasSharedTypeLine(c):
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

// TokenArtRequest is what a token creation offers the ADR 0078 art
// resolver: the template's printed characteristics, plus who is
// making the token and at which table. Controller and GameID are
// unused by the matching rule (ADR 0078 decisions 3-4 read
// characteristics only) and exist so a later per-game or per-player
// override can sit in front of the rule without a second hook
// signature — see decision 6. Passing them from day one, even
// unused, is the whole point of the seam.
type TokenArtRequest struct {
	// Template is the token as it is about to be minted: Name,
	// TypeLine, Power, Toughness, Colors and Keywords are what the
	// matching rule reads. Everything else on it (InstanceID, Owner,
	// …) is not yet stamped when this fires and must not be read.
	Template Card
	// Controller is whose token this is. Unused today.
	Controller uuid.UUID
	// GameID is which table is making it. Unused today.
	GameID uuid.UUID
}

// TokenArtResolver returns a Scryfall PRINTING id for a token about
// to be created, or "" for "no art — render the name" (ADR 0078
// decision 8). Nil means no resolver is wired, which is the state of
// every game-package test and of a server built with no Scryfall
// index at all; tokens then behave exactly as they did before ADR
// 0078 — an empty ScryfallID and the client's name-fallback render.
//
// Set from main.go over an internal/cards/tokenart.Resolver built on
// the loaded *cards.Index (internal/game cannot import internal/cards
// directly — see the EffectResolver doc above for the same cycle).
// Called from mintTokenLocked (token_create.go), after the CR 701.7b
// replacement window has settled the token's characteristics and
// before it enters the battlefield, and only when the template does
// not already carry a ScryfallID — a token COPY (CreateTokenCopy)
// already stamped the copied card's printing id via CopiableValuesOf,
// and this must never overwrite that with a generic-Soldier printing.
var TokenArtResolver func(req TokenArtRequest) string

// ManaAbilityShape is the minimal mana-ability surface the game
// package consumes. Mirrors effects.ManaAbility but lives in `game`
// to avoid an import cycle (the effects package already imports
// `game`). The S15 dispatcher reads this on every
// activate_mana_ability call to look up the ability's cost shape +
// produced-mana string.
type ManaAbilityShape struct {
	// Zones is the CR 113.6 dimension on a MANA ability: the zones
	// this ability functions from. Nil means the battlefield and
	// nowhere else, which is every mana ability the catalog held
	// before #1228.
	//
	// `ActivatedAbilityShape.Zones`' sibling (ability_zone.go, #660)
	// and `TriggeredAbility.Zones`' (#922) and `StaticAbility.Zones`'
	// (#1221), read through the same shape of predicate —
	// ManaAbilityFunctionsFromZone — by the same four consumers: the
	// activation path, the legal-move enumerator, the view's stamp
	// and, uniquely to this ability kind, the auto-tapper.
	//
	// The one printed family is the Spirit Guides' "Exile this card
	// from your hand: Add {R}" (CR 605.1a makes that a mana ability
	// and CR 113.6 is what lets it work from a hand). ADR 0071
	// Decision 1 note 3 predicted this day from the designation side:
	// "the field plus its accessor is the same two lines on the day
	// one does".
	//
	// Only the HAND is supported — supportedManaAbilityZones — and
	// effects.Register refuses anything else at boot, for the reason
	// StaticZoneUnsupported exists: a zone no consumer walks is a
	// declaration the engine silently ignores, and the card would
	// register, look complete on the catalog page and never work.
	Zones []ZoneKind

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

	// EnergyCost is "Pay N {E}" in a mana ability's cost (ADR 0129 §5,
	// CR 107.14) — Aether Hub's "{T}, Pay {E}: Add one mana of any
	// color". Such an ability is still a mana ability (CR 605.1a).
	// Validated with the other components before anything is paid
	// (CR 118.3) and paid through payEnergyLocked after the life. No
	// mana ability prints "Pay X {E}", so there is no X form. The
	// auto-tapper plans it in the energy tier, after every plan that
	// spends no energy and before the pain tier (owner decision 2).
	EnergyCost int

	// ExertCost is "Exert this land" / "Exert this creature" in a mana
	// ability's cost (ADR 0130 §4, owner decision 3; CR 701.43a) —
	// Arena of Glory's "{R}, {T}, Exert this land: Add {R}{R}", Oasis
	// Ritualist's "{T}, Exert this creature: Add two mana of any one
	// color". Still a mana ability (CR 605.1a): it is paid with the
	// other components and resolves at once (CR 605.3b), through
	// exertLocked, keyed to the activator's next untap step. A "whenever
	// you exert" trigger it causes waits for the next time a player
	// would receive priority, like every trigger from a mana ability.
	// Always payable on the battlefield (CR 701.43b). The auto-tapper
	// never plans it: an exert is a cost the player chooses, so such a
	// row is paid only when the player activates it.
	ExertCost bool

	// ManaCost is a mana component in the activation cost — the
	// Signet cycle's "{1}, {T}: Add {W}{U}", Cabal Coffers' "{2},
	// {T}". Scryfall brace grammar, parsed with ParseCost.
	//
	// Paid out of the controller's pool BEFORE the source taps, and
	// validated alongside every other component first, so a Signet
	// activated on an empty pool fails without tapping. With
	// ManaAbilityParams.AutoTap (#2215) ActivateManaAbility tops the
	// pool up from the controller's other sources first — CR 605.3a
	// lets a player activate mana abilities while paying for one —
	// and an unpayable cost still fails with nothing tapped. The
	// auto-tapper does not plan a source with an unpriced mana cost
	// as a SOURCE (autoTapAbilityAccepts): that would be a second
	// cost to solve inside the first.
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

	// ExileCards is an "Exile N cards from your hand" component of the
	// activation cost (#1283) — Cadaverous Bloom's "Exile a card from
	// your hand: Add {B}{B} or {G}{G}". The activator names the cards
	// in ManaAbilityParams.ExileIDs.
	//
	// DiscardCards' sibling one keyword action over, and NOT a discard:
	// the cards leave through the one exit primitive with
	// MustSettleNow, fire no EventDiscardCard and are invisible to
	// madness — see game.ExileCost for why the two are two components.
	// Nor is it ExileSelf below, which exiles the SOURCE and asks
	// nothing.
	//
	// The AUTO-TAPPER never plans an ability that has one, on the
	// discard's ground: which card to exile is a decision, and the
	// planner makes none.
	ExileCards *ExileCost

	// ExilePermanents is an "Exile a creature you control" component of
	// the activation cost (#1600) — Food Chain's "Exile a creature you
	// control: Add X mana of any one color, where X is 1 plus the exiled
	// creature's mana value." The activator names the permanents in
	// ManaAbilityParams.ExilePermanentIDs.
	//
	// The SAME game.ExilePermanentsCost AbilityCost.ExilePermanents
	// carries (exile_permanent_cost.go), with the same walk, validator
	// and payer. The exiled permanents are recorded on the PaidCost
	// handed to ProducedForPaid, which is how "the exiled creature's
	// mana value" is answered (read as it last existed on the
	// battlefield, LastKnownPermanentForEffect).
	//
	// Not a sacrifice: nothing dies and no EventSacrifice fires. The
	// leaves-the-battlefield triggers it queues are drained on the way
	// out, with the mana already in the pool (CR 605.3a).
	//
	// The AUTO-TAPPER never plans an ability that has one: which
	// permanent to exile is a decision, and the planner makes none.
	ExilePermanents *ExilePermanentsCost

	// ExileSelf exiles the SOURCE CARD out of the zone the ability
	// was activated from, as part of the activation cost (#1228):
	//
	//	Simian Spirit Guide  "Exile this card from your hand: Add {R}."
	//	Elvish Spirit Guide  "Exile this card from your hand: Add {G}."
	//
	// The SAME clause AbilityCost.ExileSelf carries (#1221,
	// exile_cost.go), with the same validator and the same payer,
	// because it is the same cost — scavenge exiles a card from a
	// graveyard to put counters on something, a Spirit Guide exiles
	// one from a hand to make mana, and a component declared twice is
	// a component that can be paid two ways. The ZONE it is validated
	// against comes off the ability's own Zones, not off the
	// component: the clause names "this card", and which pile it is
	// in is CR 113.6's business.
	//
	// It pays through the one exit primitive (routeCardToZoneLocked)
	// with MustSettleNow, so a commander exiled to a Spirit Guide's
	// cost still gets its CR 903.9 answer and the CR 601.2h /
	// CR 602.2b indivisible step cannot pause on a prompt.
	//
	// Refused at BOOT on an ability that declares no non-battlefield
	// zone, and required on one that does: a mana ability off the
	// battlefield has no {T} and no permanent to sacrifice, so
	// without this it would have no cost at all and CR 106.7's
	// "could produce" would be reading a free mana source.
	//
	// The AUTO-TAPPER does plan it, as the LAST-resort tier below
	// even a sacrifice source — see tapSource.LeavesHand. A card in
	// hand is worth more than a Treasure, and both are worth less
	// than an untapped land.
	ExileSelf bool

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

	// DerivesFromOtherSources marks an ability that asks OTHER
	// permanents what THEY could produce — Exotic Orchard, Reflecting
	// Pool, Fellwar Stone. Such an ability declares DerivedMatch
	// (below) instead of ProducedFunc.
	//
	// A declaration rather than something inferred at run time: the
	// alternative is a re-entrancy counter, which on a snapshotted
	// struct is undo state nobody wants and off it is a data race
	// between two games in one process.
	//
	// Added in S44 (#782).
	DerivesFromOtherSources bool

	// DerivedMatch computes CR 106.7's "could produce" derivation
	// directly, for exactly the three abilities DerivesFromOtherSources
	// marks (#1323). ProducedFunc cannot do this itself: the
	// derivation is recursive whenever the source it asks about is
	// ITSELF derived (Fellwar Stone asking an opposing Exotic
	// Orchard), and answering that correctly — a one-way chain
	// resolves, two copy-mana lands facing each other do not recurse
	// forever (CR 106.7's own closing sentence, not "106.6b" — that
	// number does not exist in the pinned edition) — needs the
	// ancestor path threaded the whole way down. A closure with
	// ProducedFunc's fixed three-argument shape (g, controller,
	// source) has no room to carry one across the package boundary
	// into a card-side helper, so this field is a plain predicate
	// instead — "candidate is a land an opponent controls" (Exotic
	// Orchard, Fellwar Stone), "candidate is a land you control"
	// (Reflecting Pool) — and producibleManaVisitingLocked
	// (producible_mana.go, this package) walks the battlefield and
	// recurses itself, with the ancestor path as an ordinary
	// parameter.
	//
	// Wins over ProducedFunc for BOTH the CR 106.7 reader and a real
	// activation — manaAbilityProducedLocked and ActivateManaAbility
	// both check it first — so the two can never compute a different
	// answer for the same board.
	DerivedMatch func(candidate Card, controller uuid.UUID) bool

	// DerivedColorsOnly drops {C} from a DerivedMatch union: "any
	// COLOR that a land an opponent controls could produce" (Exotic
	// Orchard, Fellwar Stone) cannot make colorless mana (CR 105.1);
	// "any TYPE that a land you control could produce" (Reflecting
	// Pool) can.
	DerivedColorsOnly bool

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

	// SpendRiders are what this ability's mana does WHEN IT IS SPENT
	// (#1547, mana_spend_rider.go): "and that spell can't be countered"
	// (Cavern of Souls), "if that mana is spent on a creature spell, it
	// gains haste" (Hall of the Bandit Lord), "when that mana is spent to
	// cast a red instant or sorcery spell, copy that spell" (Pyromancer's
	// Goggles). Copied onto every token the ability mints — through the
	// colour pick too, on PendingChoice.ManaRiders — with one Production
	// id per activation.
	//
	// Unlike Restrictions, a rider does NOT hide the ability from the
	// auto-tapper: it constrains nothing about where the mana may go, so
	// a planned Goggles is still a red source, and its rider fires on the
	// spend exactly as it would for a hand-tapped one.
	SpendRiders []ManaSpendRider

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

	// OncePerTurn declares "Activate only once each turn" on a MANA
	// ability — Vivi Ornitier's "{0}: Add X mana …", Ramos's counter
	// payout (#1621).
	//
	// It is a DECLARATION whose enforcement travels with it: the one
	// constructor that sets it, effects.manaShapes, folds the gate
	// ("this object has not activated the ability labelled Label this
	// turn", ActivatedThisTurn over the per-turn activation record)
	// into Condition in the same statement. So every reader that
	// already honours Condition — ActivateManaAbility, the legal-move
	// enumerator, the view's condition_unmet, the auto-tapper's planner
	// and executor — honours this one too, without learning it exists.
	//
	// The bit itself has ONE reader, and that is why it is a bit rather
	// than only a closure: the AUTO-TAPPER. An ability that costs its
	// source nothing at all (no {T}, no sacrifice, no other component)
	// is plannable only when something bounds it, and a Condition is an
	// opaque closure the planner cannot read. OncePerTurn is the bound
	// it can read — see autoTapFreeOncePerTurn and the last-resort
	// tier it opens (ADR 0011, amendment 2026-09-30).
	//
	// Keyed by the LABEL (activation_tally.go), like Exhaust, so
	// effects.Register refuses a blank one.
	OncePerTurn bool

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

	// RiderSelfDamage is the damage Rider deals the ability's
	// controller, when Rider is exactly "This permanent deals N damage
	// to you" (effects.ManaAbility.PainToYou builds both). Zero for an
	// opaque rider, or none.
	//
	// It exists for the auto-tapper (#2392). A Rider is a closure the
	// planner cannot read, so a source with one is never planned —
	// unless it declares what the rider costs here, which makes the
	// painlands' coloured halves, Ancient Tomb and Grand Coliseum
	// plannable in the pain tier (tapSource.Pain), never for more
	// damage than leaves the player above 0 life.
	RiderSelfDamage int

	// PreRider is Rider's mirror image: everything the oracle text
	// says BEFORE the "Add …" clause, run once the cost is paid and
	// BEFORE the produced string is computed (Produced / ProducedFunc
	// / ProducedForPaid, in that precedence) — Empowered Autogenerator's
	// "Put a charge counter on this artifact. Add X mana of any one
	// color, where X is the number of charge counters on this
	// artifact." The counter has to land before X is read, or X is a
	// guess at what the placement will do rather than a fact about
	// what it did.
	//
	// A card declares Rider or PreRider, never both for the same
	// clause — whichever one matches where its own printed sentence
	// puts the non-mana instruction relative to "Add …".
	//
	// Same locking contract as Rider: runs under g.mu held for write,
	// *ForEffect helpers only. Any mutation this makes and ProducedFunc
	// then reads back MUST go through a mustSettleNow entry point —
	// AddCounterMustSettleNowForEffect for a counter placement — because
	// CR 605.3b leaves no priority window inside a mana ability's
	// resolution for a CR 616 ordering prompt (two counter doublers) to
	// occupy. The pipeline settles on the gathered order instead of
	// asking, exactly as RepEventProduceMana already does for a
	// production-side doubler stack (see produce_mana.go).
	//
	// Nil for every mana ability but Empowered Autogenerator today.
	//
	// Added in the #1370 fix.
	PreRider func(g *Game, controller, source uuid.UUID) error

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
