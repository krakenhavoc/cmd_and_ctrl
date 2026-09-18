package game

import (
	"github.com/google/uuid"
)

// activated.go — S21 sub-PR 2: activated abilities in the catalog
// (CR 602), the fourth ability type after triggered (S19), static
// (S16) and replacement (S17).
//
// Before this, a catalog card's activated ability didn't exist:
// ActivateAbility put a labelled item on the stack with no cost
// check and no effect, and players resolved it in their heads. A
// sacrifice outlet ("Sacrifice a creature: deal 1 damage") can't
// work that way — the cost IS the game action.
//
// The shape mirrors the rest of the catalog surface. A card declares
// its abilities; the engine validates timing and cost, pays the
// cost, and pushes a stack item carrying the same Effect closure
// S19's triggered abilities use. Resolution is therefore already
// written: resolveTopAbilityLocked runs Effect with the live game,
// after the CR 608.2b target re-check.
//
// Mana abilities stay separate (CR 605.3b — they don't use the
// stack) and keep their own ManaAbilityShape path. An ability that
// would be a mana ability but whose cost sacrifices a DIFFERENT
// permanent (Ashnod's Altar) fits neither surface cleanly and is
// deferred; the note is in ADR 0020.

// AbilityCost is what a player pays to activate an ability
// (CR 602.1a). Every field is additive: Goblin Bombardment is
// {SacrificeOther: creature-spec}, Krenko is {Tap: true}, a
// hypothetical "{2}, {T}, Sacrifice a creature:" would set all
// three.
type AbilityCost struct {
	// Tap requires the source to be untapped, taps it, and — for a
	// creature source — enforces summoning sickness (CR 302.6).
	Tap bool

	// SacrificeSelf sacrifices the source as part of the cost.
	SacrificeSelf bool

	// SacrificeOther sacrifices other permanents the activator
	// controls, chosen at announce and matched against this spec
	// ("Sacrifice a creature"). The activator names them in
	// ActivateAbilityParams.SacrificeIDs. The source itself is a
	// legal choice when the spec admits it — Carrion Feeder can eat
	// itself, as in paper.
	//
	// The spec's Min == Max is how many (#747): "Sacrifice two
	// artifacts" is a clause with a count of 2, built by
	// effects.SacrificeN. Every one-permanent constructor stamps 1.
	SacrificeOther *TargetSpec

	// Mana is a printed cost string ("{1}{B}") paid from the pool
	// under the same strict / permissive rules casting uses. Empty
	// means no mana component.
	Mana string

	// Life is a life payment (CR 119.4). Paying life is legal at any
	// total above the payment; the SBA loop handles the rest.
	Life int

	// Loyalty is the loyalty-counter component of a planeswalker's
	// loyalty ability (CR 606.4): +N adds N counters to the source,
	// −N removes N, and [0] neither. Nil means "this is not a
	// loyalty ability", which is every other ability in the catalog.
	//
	// A POINTER, not a plain int, because [0] is a real printed cost
	// (Jace's [0]: Brainstorm) and has to be distinguishable from
	// "no loyalty component" — the difference decides whether the
	// activation burns the turn's once-per-turn window.
	//
	// ADR 0020 deliberately kept loyalty OUT of AbilityCost and let
	// the S13.1 ActivateLoyalty sandbox action carry it. ADR 0032 §7
	// re-audited that call: the sandbox action moves a counter and
	// runs no effect, so no planeswalker in the game could do
	// anything, which is what #329 and #334 report. Reversing the
	// exclusion is what makes a loyalty ability an ordinary CR 602
	// activation — same stack item, same Effect closure, same
	// targeting.
	//
	// Three rules ride along with a non-nil Loyalty and are enforced
	// in ActivateCatalogAbility rather than asked of each card:
	//
	//	CR 606.2  the source must be a planeswalker you control
	//	CR 606.6  a −N cost needs at least N loyalty counters
	//	CR 606.3  sorcery speed, once per turn per planeswalker
	//
	// SorcerySpeed therefore does not need to be set alongside it.
	Loyalty *int

	// Crew is the crew number of a Vehicle's crew ability (CR
	// 702.122a): "Tap any number of untapped creatures you control
	// with total power N or more". Zero means "not a crew cost",
	// which is every ability that is not printed on a Vehicle.
	//
	// A plain int rather than a pointer, unlike Loyalty: crew 0 is
	// not a printed cost, so zero and absent are the same thing.
	//
	// Three things make crew different from every other cost here
	// and are enforced in ActivateCatalogAbility rather than asked of
	// each card:
	//
	//	CR 702.122b  the creatures tapped are NOT the source, so
	//	             Tap must stay false — a crewed Vehicle is
	//	             untapped and can attack.
	//	CR 702.122b  summoning sickness does not apply. Tapping a
	//	             creature to crew is not paying a {T} cost, so a
	//	             creature that arrived this turn may crew.
	//	CR 702.122a  power is read at the moment the cost is paid,
	//	             from the post-layer effective value, so an
	//	             anthem and +1/+1 counters both count.
	//
	// The activator names the creatures in
	// ActivateAbilityParams.CrewIDs. Any number is legal as long as
	// the total clears the bar, and a single 5-power creature crews a
	// Vehicle that says crew 3 — the printed text is a floor, not an
	// exact amount.
	Crew int

	// RemoveCounters is a "remove N counters" component (#625): from
	// the source ("Remove a gold counter from this artifact"), from
	// another permanent the activator controls ("remove a loyalty
	// counter from a planeswalker you control"), or of any kind
	// ("Remove a counter from a creature you control"). Nil means no
	// counter component. See CounterRemovalCost in counter_cost.go.
	//
	// The activator names the permanent in
	// ActivateAbilityParams.CounterSourceIDs and, for the any-kind
	// form, the kind in CounterKind — both at announce, beside the
	// sacrifice and crew picks. Two rules are enforced in
	// ActivateCatalogAbility rather than asked of each card:
	//
	//	CR 118.3 / 602.2b  paid at activation, after everything
	//	                   else is validated; the removal is not a
	//	                   replaceable event, so nothing doubles or
	//	                   halves it.
	//	CR 606             removing a planeswalker's loyalty counter
	//	                   this way is not activating a loyalty
	//	                   ability — no sorcery-speed window, and the
	//	                   once-per-turn flag is left alone.
	//
	// "Rather than pay" on an activated ability (Heart of Kiran) is
	// modelled as a SECOND ability entry whose cost is this component,
	// not as an alternatives slot on AbilityCost: the client already
	// lists abilities separately, and the entry is only payable while
	// a real planeswalker holds a real counter, which is the whole of
	// what #259 asks.
	RemoveCounters *CounterRemovalCost

	// MinX is the floor the printed text puts on the announced X —
	// Helm of Obedience's "X can't be 0" is MinX: 1. Zero means the
	// ordinary floor of zero, which is what every other {X} cost
	// charges.
	//
	// There is deliberately no MaxX here. A cost has no printed
	// ceiling on X; what the activator can actually pay is the
	// ceiling, and that is the mana check's job, not the card's.
	// The legal-move enumerator's Options.MaxX is a search bound on
	// the BOT, not a rule, and lives there for that reason.
	//
	// MinX is meaningless without an {X} in Mana, and Register
	// panics at boot on a spec that sets one without the other: a
	// floor on a variable that cannot vary is a card-file mistake,
	// and it would silently make the ability unactivatable.
	MinX int
}

// DemandsX reports whether the ability's mana component contains
// {X}, and is therefore an ability the activator announces a value
// for at CR 602.2b.
//
// X lives in the MANA component and nowhere else. A spell can grow
// an X outside its printed cost (Toxic Deluge's "pay X life" rides
// AdditionalCost.PayLifeX, waterbend's rides TapPermanentsCost), so
// the cast path has three places to ask. An ability has one, and
// keeping it that way is what lets every consumer — the view, the
// enumerator, the client — derive "does this prompt for X" from the
// cost string rather than from a flag a card file could forget to
// set.
func (c AbilityCost) DemandsX() bool {
	return c.XSlots() > 0
}

// XSlots is how many {X} tokens the mana component carries. Usually
// 1; Treasure Vault's "{X}{X}" is 2, and the total generic demand is
// XSlots * the announced X. An unparseable cost reports 0 — the
// activation path rejects it separately rather than guessing.
func (c AbilityCost) XSlots() int {
	if c.Mana == "" {
		return 0
	}
	cost, err := ParseCost(c.Mana)
	if err != nil {
		return 0
	}
	return cost.XSlots
}

// FloorX is the smallest legal announcement for this cost: MinX when
// the card prints a floor, zero otherwise. Zero for a cost with no
// {X} at all, which is the only value the engine accepts there.
func (c AbilityCost) FloorX() int {
	if !c.DemandsX() || c.MinX < 0 {
		return 0
	}
	return c.MinX
}

// ActivatedAbilityShape is one activated ability on a permanent, as
// the game package consumes it. Mirrors effects.ActivatedAbility,
// which lives in the catalog package (the same import-cycle dodge
// ManaAbilityShape uses).
type ActivatedAbilityShape struct {
	// Label is the oracle text of the ability, shown in the
	// activation menu: "Sacrifice a creature: Goblin Bombardment
	// deals 1 damage to any target."
	Label string

	Cost AbilityCost

	// Targets is the ability's target clause, validated at announce
	// and re-checked at resolution exactly as a spell's is. Nil for
	// untargeted abilities.
	Targets *TargetSpec

	// Modes is the CR 700.2 mode clause of a modal activated ability
	// ("{4}, {T}: Choose one — double the counters on target
	// permanent; double the counters you have"). The same
	// game.ModeSpec a modal spell and a modal triggered ability
	// declare — one struct, three owners (#764, ADR 0065 §3).
	//
	// Chosen at ACTIVATION with the targets (CR 602.2b), in the same
	// message, because activating an ability is one indivisible step
	// and there is nobody to prompt: the player who activates is the
	// player who chooses. Validated exactly as a cast's modes are;
	// each chosen occurrence contributes its option's target clauses
	// to the announcement, in mode order.
	//
	// Nil for every ability that is not modal, which is nearly all of
	// them. Added by #764.
	Modes *ModeSpec

	// SorcerySpeed marks "activate only as a sorcery" (CR 602.5d).
	SorcerySpeed bool

	// Condition is the ability's other activation instructions
	// (CR 602.1b): "Activate only if an opponent controls four or
	// more lands" (Tectonic Edge), "Activate only during your turn"
	// (Sanctum of Eternity). Nil means none, which is nearly every
	// ability. ADR 0020's #743 addendum.
	//
	// The contract is ManaAbilityShape.Condition's, word for word:
	// READ-ONLY, and it runs under g.mu — held for write by
	// ActivateCatalogAbility, for read by the view and the legal
	// enumerator — so it reads *ForEffect accessors and g.Turn /
	// g.Seats directly, never a public locking accessor
	// (g.ActivePlayer deadlocks the activation). `controller` is the
	// activating player, the "you" of the printed text; `source` is
	// the permanent's instance ID, enough to read a per-source count
	// later without changing the signature.
	//
	// Checked once, in ActivateCatalogAbility, right after the timing
	// check and before X, the loyalty checks, the costs and the
	// targets, so a false return is ErrConditionNotMet with nothing
	// paid (CR 602.5). Never re-checked at resolution: an activation
	// instruction is not part of the effect (CR 602.1b).
	//
	// Every viewer receives the view's condition_unmet flag, so a
	// condition must read only public information (board counts,
	// graveyard and hand sizes, life, the turn and the step).
	//
	// SorcerySpeed stays a separate flag: some cards print both
	// (Speaker of the Heavens), and the two fail with different
	// errors.
	Condition func(g *Game, controller, source uuid.UUID) bool

	// Effect runs at resolution against the live game. Same contract
	// as TriggeredAbility's stack items: never capture a *Card,
	// read what you need off the item and the game.
	Effect func(g *Game, item *StackItem) error
}

// CatalogActivatedAbilities returns the registered activated
// abilities for an oracle ID, or nil. Populated by the effects
// package at init, alongside the other catalog hooks.
var CatalogActivatedAbilities func(oracleID string) []ActivatedAbilityShape

// ActivatedAbilitiesForCard returns the activated abilities a card
// offers right now. Catalog-only: unlike mana abilities there's no
// synthetic fallback, because there's no ability every card of some
// type implicitly has.
func ActivatedAbilitiesForCard(c Card) []ActivatedAbilityShape {
	// S24 layer 6: a permanent an ability-removing effect applies to
	// offers nothing, and that has to be checked BEFORE the
	// instance-carried list as well as before the catalog. A Clue
	// token's "{2}, Sacrifice this: Draw a card" is an activated
	// ability like any other, and Darksteel Mutation takes it away
	// exactly as it takes away Sol Ring's.
	//
	// This one accessor is why internal/legal needs no change:
	// legal.Move enumeration, the client's context menu
	// (protocol.CardView), the lobby's ability lookup and
	// ActivateCatalogAbility itself all read through here, so the
	// enumerator can never offer a bot a move the engine will refuse
	// — the #544 hung-table failure mode.
	if c.HasLostAllAbilities() {
		return nil
	}
	// S21 sub-PR 4: intrinsic abilities win — a token has no oracle
	// ID for the catalog to key on, and Food / Clue / Blood ARE
	// their activated ability.
	if len(c.ActivatedAbilities) > 0 {
		return c.ActivatedAbilities
	}
	if CatalogActivatedAbilities == nil || c.OracleID == "" {
		return nil
	}
	return CatalogActivatedAbilities(CatalogAbilityKey(c))
}

// ActivateAbilityParams carries the announce-time choices for a
// catalog activated ability.
type ActivateAbilityParams struct {
	// SacrificeIDs names the permanents paid to a SacrificeOther
	// cost: exactly the clause's count (SacrificeCostCount), each
	// once. Their order does not matter — the permanents leave as
	// one simultaneous exit (#747).
	SacrificeIDs []uuid.UUID

	// CrewIDs names the creatures tapped to pay a Crew cost, in the
	// order the activator picked them. Any number is legal; what
	// matters is that their total effective power clears the crew
	// number (CR 702.122a).
	CrewIDs []uuid.UUID

	// CounterSourceIDs names the permanent a RemoveCounters cost
	// removes from (#625). Exactly one for the "from a planeswalker
	// you control" form; empty (or the source's own ID) for the
	// self form. A slice for the same wire-shape reason SacrificeIDs
	// is one.
	CounterSourceIDs []uuid.UUID

	// CounterKind is the kind a "remove a counter" cost of ANY kind
	// removes (Fain, the Broker), chosen at announce with the
	// permanent. Optional for a cost that prints its kind — if sent,
	// it must repeat that kind.
	CounterKind string

	// Targets are the ability's targets, validated against the
	// ability's clause list — or, for a modal ability, against the
	// clauses of the chosen modes (#764).
	Targets []TargetRef

	// Modes are the mode indexes announced for a modal activated
	// ability (CR 602.2b, CR 700.2), in the order chosen; a repeated
	// index is legal only when the ability's ModeSpec is Repeatable
	// (CR 700.2d). Empty for every non-modal ability, and non-empty
	// for one is rejected rather than ignored. Added by #764.
	Modes []int

	// XValue is the value announced for an {X} in the ability's mana
	// component (CR 602.2b). Chosen as part of ACTIVATING the
	// ability — after the source and the mode, before any cost is
	// paid — and locked from that moment: it rides onto the stack
	// item and nothing later can change it, which is what makes an
	// X ability's effect (Helm of Obedience's mill, Treasure Vault's
	// token count) a fact about the announcement rather than about
	// the board at resolution.
	//
	// The same slot on the cast path is CastSpellParams.XValue, and
	// the two are validated the same way: non-negative, zero unless
	// the cost actually has an {X}, and at least the cost's printed
	// floor.
	XValue int

	// Strict / AutoTap mirror CastSpellParams: they gate the mana
	// component of the cost the same way a cast is gated.
	Strict  bool
	AutoTap bool
}

// ActivateCatalogAbility activates ability `index` on a permanent
// the player controls: validates timing and cost, pays the cost,
// and pushes a stack item carrying the ability's effect (CR 602.2).
// The activator retains priority, as with any announce.
//
// Costs are validated in full before ANY of them is paid, so a
// half-paid activation can't strand the board — the same discipline
// the sacrifice-cost mana ability path uses.
//
// Caller must NOT hold g.mu.
func (g *Game) ActivateCatalogAbility(playerID, cardID uuid.UUID, index int, params ActivateAbilityParams) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.SplitSecondActive {
		return ErrSplitSecondActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	// S24: the CR 602.5 restriction rides on the effective
	// characteristic, so the layers have to be fresh before the
	// source is judged. Hoisted above the controller check rather
	// than tucked in beside the restriction test, because
	// Card.Controller is itself layer 2's materialised output — a
	// Mind Control that resolved a moment ago has to be visible to
	// "do you control this permanent" too. Fast-path no-op when
	// nothing has changed.
	g.RecomputeLayersIfStaleLocked()
	source := findBattlefieldCard(g, cardID)
	if source == nil {
		return ErrCardNotFound
	}
	if source.Controller != playerID {
		return ErrCardCallerMismatch
	}
	// CR 602.5: "its activated abilities can't be activated"
	// (Arrest, Faith's Fetters). Checked before the index lookup so
	// the answer does not depend on which ability was named, and
	// before any cost validation so nothing is paid. Loyalty
	// abilities are activated abilities (CR 606.1) and are covered
	// here too. Mana abilities take the other entry point and the
	// other bit — see restrictions.go on why that split exists.
	if !CanActivateAbilities(source) {
		return ErrCantActivate
	}
	abilities := ActivatedAbilitiesForCard(*source)
	if index < 0 || index >= len(abilities) {
		return ErrInvalidParam
	}
	ab := abilities[index]

	// --- timing -------------------------------------------------
	// CR 606.3: a loyalty ability is sorcery-speed whether or not
	// the catalog entry bothered to say so — the loyalty component
	// carries the restriction, so a card can't forget it.
	if (ab.SorcerySpeed || ab.Cost.Loyalty != nil) && !g.sorcerySpeedOpenLocked(playerID) {
		return ErrSorcerySpeedRequired
	}
	// CR 602.1b / 602.5: the ability's "Activate only if …" and
	// "Activate only during …" instructions (#743). A player can't
	// begin to activate an ability whose instructions forbid it, so
	// this stops the activation before X, targets or any cost — and
	// it is never checked again at resolution. When both this and the
	// timing check fail, the timing error wins; that order is ours,
	// not the rules'.
	if ab.Condition != nil && !ab.Condition(g, playerID, cardID) {
		return ErrConditionNotMet
	}

	// --- validate every cost before paying any ------------------
	// CR 602.2b: X is announced as part of activating, with the
	// targets and before anything is paid. Same three checks the
	// cast path makes at CR 601.2b, in the same order.
	//
	// The "no {X}, so X must be zero" branch is rejected rather than
	// ignored on purpose, exactly as an unexpected sacrifice_ids is:
	// a client that sends an X for Krenko is confused about which
	// ability it is firing, and silently dropping the field would
	// hide that from whoever has to debug it.
	if params.XValue < 0 {
		return ErrInvalidParam
	}
	if !ab.Cost.DemandsX() {
		if params.XValue != 0 {
			return ErrInvalidParam
		}
	} else if params.XValue < ab.Cost.FloorX() {
		// "X can't be 0" (Helm of Obedience). A floor is part of the
		// cost, so announcing under it is an illegal announcement,
		// not a cheap one.
		return ErrInvalidParam
	}
	if ab.Cost.Loyalty != nil {
		// CR 606.2: loyalty abilities live on planeswalkers. The
		// controller check above already covers "a planeswalker you
		// control".
		if !source.IsPlaneswalker() {
			return ErrNotAPlaneswalker
		}
		// CR 606.3: once per turn per planeswalker. The flag was
		// S13.1's and only the sandbox action consulted it; this is
		// the path that matters now.
		if g.LoyaltyActivatedThisTurn[cardID] {
			return ErrLoyaltyAlreadyActivated
		}
		// CR 606.6: you can't activate a −N ability with fewer than
		// N loyalty counters. Paying down to exactly 0 is legal and
		// the 704.5i SBA sweeps the walker afterwards.
		if n := *ab.Cost.Loyalty; n < 0 && source.Counters[CounterLoyalty] < -n {
			return ErrInsufficientLoyalty
		}
	}
	if ab.Cost.Tap {
		if source.Tapped {
			return ErrAlreadyTapped
		}
		// CR 302.6: a creature's {T} ability needs it to have been
		// under your control since your most recent turn began.
		// Non-creature sources (Krenko is a creature; an artifact
		// with {T} isn't) are never sick.
		if source.IsCreature() && HasSummoningSickness(source) {
			return ErrSummoningSick
		}
	}
	sacrifices, err := g.validateSacrificeCostLocked(playerID, cardID, ab.Cost, params.SacrificeIDs)
	if err != nil {
		return err
	}
	crew, err := g.validateCrewCostLocked(playerID, ab.Cost, params.CrewIDs)
	if err != nil {
		return err
	}
	counters, err := g.validateCounterRemovalCostLocked(playerID, cardID, ab.Cost, params.CounterSourceIDs, params.CounterKind)
	if err != nil {
		return err
	}
	if ab.Cost.Life > 0 && p.Life < ab.Cost.Life {
		// CR 119.4 forbids paying more life than you have. Paying
		// down to exactly 0 is legal; the SBA loop ends the game
		// after.
		return ErrInvalidParam
	}
	// CR 602.2b / 700.2: the modes are announced with the targets, in
	// that order — the chosen bullets are what decide which target
	// clauses the activation even has (#764).
	if err := validateModes(ab.Modes, params.Modes); err != nil {
		return err
	}
	if ab.Modes == nil && len(params.Modes) > 0 {
		return ErrInvalidParam
	}
	steps := AnnouncedClauses(ab.Targets, ab.Modes, params.Modes)
	if len(steps) == 0 && len(params.Targets) > 0 {
		return ErrInvalidParam
	}
	xSteps := resolveStepCountsFromX(steps, params.XValue)
	params.Targets = assignAnnouncedSlots(steps, params.Targets)
	for _, i := range xSteps {
		if n := stepTargetCount(steps[i], params.Targets); n != params.XValue {
			return ErrInvalidParam
		}
	}
	if err := g.validateAnnouncedTargetsLocked(playerID, steps, params.Targets); err != nil {
		return err
	}

	// --- pay ----------------------------------------------------
	if ab.Cost.Mana != "" {
		// S32 (#352): "activate abilities of colorless Eldrazi" is a
		// restriction on the SOURCE permanent, so the spend context
		// is built from it.
		// An ability whose cost includes {T} cannot tap its own
		// source for mana to help pay itself — the {T} and the mana
		// are components of the same cost, and the source can only
		// be tapped once. Without the exclusion the auto-tapper is
		// free to spend Treasure Vault's own "{T}: Add {C}" on its
		// "{X}{X}, {T}" ability, which is a free mana on every
		// activation of every tap ability with a mana component.
		var excluded map[uuid.UUID]bool
		if ab.Cost.Tap {
			excluded = map[uuid.UUID]bool{cardID: true}
		}
		if err := g.payAbilityManaCostLocked(p, cardID, ab.Cost.Mana, params, ManaSpendForAbility(*source), excluded); err != nil {
			return err
		}
	}
	if ab.Cost.Tap {
		source.Tapped = true
		g.EmitEvent(Event{Kind: EventTapCard, Actor: playerID, CardID: cardID})
	}
	for _, id := range crew {
		// The crewing creatures tap, the Vehicle does not (CR
		// 702.122b) — which is the whole point, since a Vehicle that
		// tapped to crew itself could never attack.
		if c := findBattlefieldCard(g, id); c != nil {
			c.Tapped = true
			g.EmitEvent(Event{Kind: EventTapCard, Actor: playerID, CardID: id})
		}
	}
	// #793: the cost path — CR 602.2b activates an ability in one
	// indivisible step, so the payment runs the CR 614 window
	// (CR 119.4) but never stops to ask a CR 616 ordering question.
	if ab.Cost.Life > 0 {
		if err := g.PayLifeForEffect(cardID, playerID, ab.Cost.Life); err != nil {
			return err
		}
	}
	if ab.Cost.Loyalty != nil {
		// applyCounterLocked, not AddCounterForEffect: paying a cost
		// is not an effect (CR 121.1 / 606.2), so counter-doubling
		// replacements do NOT apply to a loyalty ability's + cost.
		// Doubling Season really does nothing here, and routing
		// through the CR 614 pipeline would silently make it.
		if err := g.applyCounterLocked(cardID, CounterLoyalty, *ab.Cost.Loyalty); err != nil {
			return err
		}
		if g.LoyaltyActivatedThisTurn == nil {
			g.LoyaltyActivatedThisTurn = make(map[uuid.UUID]bool)
		}
		g.LoyaltyActivatedThisTurn[cardID] = true
	}
	// #625: a "remove N counters" component. After life and loyalty,
	// before sacrifices — a self-form removal on a source that is
	// also sacrificed has to find the source still on the
	// battlefield. Never replaceable and never a loyalty activation;
	// payCounterRemovalLocked says why.
	if err := g.payCounterRemovalLocked(counters); err != nil {
		return err
	}
	// Sacrifices last: they move cards, which invalidates `source`.
	// One payment is one simultaneous exit (#747, CR 603.10a).
	if err := g.payCostSacrificesLocked(sacrifices); err != nil {
		return err
	}
	source = nil

	// --- announce -----------------------------------------------
	itemID := uuid.New()
	if g.StackMeta == nil {
		g.StackMeta = make(map[uuid.UUID]*StackItem)
	}
	item := &StackItem{
		ID:           itemID,
		Kind:         StackItemActivated,
		Controller:   playerID,
		Owner:        playerID,
		SourceCardID: cardID,
		Label:        ab.Label,
		Targets:      append([]TargetRef(nil), params.Targets...),
		Modes:        append([]int(nil), params.Modes...),
		// CR 602.2b: X was announced above and is locked here. The
		// effect reads it back through Context.X(), the same
		// accessor an X spell's OnResolve uses, and the wire ships
		// it on the stack item so the table can see what was paid.
		XValue:     params.XValue,
		Effect:     ab.Effect,
		targetSpec: ab.Targets,
		modeSpec:   ab.Modes,
		Seq:        g.nextStackSeqLocked(),
	}
	g.StackMeta[itemID] = item
	// #628: activating an ability is a player decision, so it
	// restarts the CR 726 loop run. The announce emits EventTrigger
	// rather than an event of its own — the same kind a triggered
	// ability announces with — so the notch is here rather than in
	// turnTallyListener, which cannot tell the two apart.
	g.notePlayerDecisionLocked()
	g.EmitEvent(Event{
		Kind:   EventTrigger,
		Actor:  playerID,
		Source: cardID,
		CardID: cardID,
	})
	// CR 602.2b / 115.7: the ability's targets were chosen as it was
	// put on the stack. S22, for "whenever ~ becomes the target of a
	// spell or ability" (Monk Gyatso) — the clause names abilities as
	// well as spells, so the activated path has to emit too.
	g.emitBecameTargetLocked(playerID, cardID, itemID, params.Targets)
	// The cost may have queued dies-triggers (a sacrifice outlet
	// feeding Blood Artist). Drain them so they sit ABOVE the
	// ability on the stack, which is where paying a cost puts them.
	g.runStateChecksLocked()
	return nil
}

// validateCrewCostLocked resolves a Crew cost into the concrete list
// of creatures to tap, without tapping anything (CR 702.122a).
//
// What it enforces, and what it deliberately does not:
//
//   - Every named card must be an untapped creature the activator
//     controls, and must be named once. Naming the same creature
//     twice would let one 3-power creature crew a 6.
//   - Total EFFECTIVE power must reach the crew number. Effective,
//     so an anthem and +1/+1 counters both count (CurrentPower), and
//     read at payment time (CR 702.122a) rather than at declaration.
//   - Summoning sickness is NOT checked. Tapping a creature to crew
//     is not paying a {T} cost, so a creature that arrived this turn
//     may crew (CR 702.122b). This is the rule most implementations
//     get wrong in the strict direction; getting it wrong the other
//     way is impossible here, because the check simply isn't made.
//   - Overshooting is legal. The printed number is a floor.
//
// The clause is "creatures you control", with no "another", so a
// Vehicle that something else has already animated is a legal
// crewer — of a different Vehicle, or in principle of itself. That
// is deliberately not special-cased: the rule says what it says, and
// self-crewing taps the Vehicle and strands it, which is a bad play
// rather than an illegal one.
//
// Caller must hold g.mu.
func (g *Game) validateCrewCostLocked(playerID uuid.UUID, cost AbilityCost, chosen []uuid.UUID) ([]uuid.UUID, error) {
	if cost.Crew <= 0 {
		if len(chosen) > 0 {
			return nil, ErrInvalidParam
		}
		return nil, nil
	}
	if len(chosen) == 0 {
		return nil, ErrInsufficientCrew
	}
	// Effective characteristics have to be fresh: a creature that
	// gained +2/+2 this turn crews for its current power, not its
	// printed one.
	g.RecomputeLayersIfStaleLocked()
	seen := make(map[uuid.UUID]bool, len(chosen))
	total := 0
	out := make([]uuid.UUID, 0, len(chosen))
	for _, id := range chosen {
		if seen[id] {
			return nil, ErrInvalidParam
		}
		seen[id] = true
		c := findBattlefieldCard(g, id)
		if c == nil {
			return nil, ErrCardNotFound
		}
		if c.Controller != playerID {
			return nil, ErrCardCallerMismatch
		}
		if !c.IsCreature() {
			return nil, ErrNotACreature
		}
		if c.Tapped {
			return nil, ErrAlreadyTapped
		}
		total += c.CurrentPower()
		out = append(out, id)
	}
	if total < cost.Crew {
		return nil, ErrInsufficientCrew
	}
	return out, nil
}

// validateSacrificeCostLocked resolves the SacrificeSelf /
// SacrificeOther components into the concrete list of permanents to
// sacrifice, without moving anything. Caller must hold g.mu.
//
// #747 (ADR 0020 addendum §13): the clause's count is the spec's
// Min == Max — "Sacrifice two artifacts" is a clause with Max 2 — and
// the payment must name exactly that many permanents, each once, each
// on the battlefield under the payer's control (CR 701.21a) and each
// matching the clause. None may be the source when SacrificeSelf pays
// the source too. Any failure refuses the whole payment with nothing
// moved; all three cost sites (an activated ability, a mana ability,
// an additional cost to cast) share this one check.
func (g *Game) validateSacrificeCostLocked(playerID, sourceID uuid.UUID, cost AbilityCost, chosen []uuid.UUID) ([]uuid.UUID, error) {
	var out []uuid.UUID
	if cost.SacrificeSelf {
		out = append(out, sourceID)
	}
	if cost.SacrificeOther == nil {
		if len(chosen) > 0 {
			return nil, ErrInvalidParam
		}
		return out, nil
	}
	if len(chosen) != SacrificeCostCount(cost.SacrificeOther) {
		return nil, ErrInvalidParam
	}
	seen := make(map[uuid.UUID]bool, len(chosen))
	for _, id := range chosen {
		// One permanent pays one sacrifice. Naming it twice would let
		// a single Food pay "Sacrifice two Foods".
		if seen[id] {
			return nil, ErrInvalidParam
		}
		seen[id] = true
		c := findBattlefieldCard(g, id)
		if c == nil {
			return nil, ErrCardNotFound
		}
		// CR 701.21a — you can only sacrifice what you control.
		if c.Controller != playerID {
			return nil, ErrCardCallerMismatch
		}
		// specMatchLocked, not targetLegalLocked: sacrificing a
		// permanent to pay a cost does not target it (CR 601.2h), so
		// the CR 702 keyword gate must not apply — Carrion Feeder can
		// still eat your own hexproof creature.
		if !g.specMatchLocked(playerID, cost.SacrificeOther, TargetRef{Kind: TargetCard, ID: id}, false) {
			return nil, ErrIllegalTarget
		}
		// Paying the same permanent twice (self-sacrifice plus the
		// same card as one of the "other" picks) isn't a legal cost
		// payment.
		if cost.SacrificeSelf && id == sourceID {
			return nil, ErrInvalidParam
		}
	}
	return append(out, chosen...), nil
}

// payAbilityManaCostLocked charges an activated ability's mana
// component through the same pool / auto-tap machinery a cast uses.
// Permissive mode (the default) leaves the pool alone and emits a
// cost warning, matching how S15 treats an unaffordable cast.
// Caller must hold g.mu.
// `spendCtx` describes the ability's SOURCE permanent, which is what
// a restricted token is matched against when the restriction says
// "activate abilities of …" (#352).
func (g *Game) payAbilityManaCostLocked(p *Player, sourceID uuid.UUID, costStr string, params ActivateAbilityParams, spendCtx ManaSpendContext, excluded map[uuid.UUID]bool) error {
	cost, err := ParseCost(costStr)
	if err != nil {
		return ErrInvalidParam
	}
	// The announced X multiplies into the generic demand exactly as
	// it does for a cast: cost.Generic + cost.XSlots*x. Treasure
	// Vault's "{X}{X}" has two slots, so X=3 costs six.
	x := params.XValue
	if !params.Strict && !params.AutoTap {
		if !p.ManaPool.CanPayFor(cost, x, spendCtx) {
			g.EmitEvent(Event{
				Kind:   EventCostWarning,
				Actor:  p.ID,
				Source: sourceID,
			})
			return nil
		}
		p.ManaPool.SpendManaFor(cost, x, spendCtx)
		return nil
	}
	if params.AutoTap && !p.ManaPool.CanPayFor(cost, x, spendCtx) {
		plan, ok := g.autoTapLocked(p.ID, cost, x, excluded)
		if !ok {
			return &InsufficientManaError{Missing: p.ManaPool.MissingFor(cost, x, spendCtx)}
		}
		g.materializePlanLocked(p, plan, cost)
	}
	if !p.ManaPool.CanPayFor(cost, x, spendCtx) {
		return &InsufficientManaError{Missing: p.ManaPool.MissingFor(cost, x, spendCtx)}
	}
	p.ManaPool.SpendManaFor(cost, x, spendCtx)
	return nil
}
