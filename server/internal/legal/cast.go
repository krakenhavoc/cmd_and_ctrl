package legal

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castParams is the cast_spell wire payload this package emits. Every
// cast is sent strict + auto_tap: the bot pays for what it casts, and
// affordability is decided here so the engine never has to reject.
type castParams struct {
	InstanceID string `json:"instance_id"`
	FromZone   string `json:"from_zone,omitempty"`
	// AlternativeCost is the key the cast claims (CR 118.9) — the
	// offer a granted permission synthesises for a non-hand cast
	// (ADR 0066), since a zone a permission prices cannot be cast
	// from without paying that price.
	AlternativeCost string `json:"alternative_cost,omitempty"`
	// AltCostIDs are the cards paid to the NON-MANA half of the
	// claimed cost (CR 601.2b) — Force of Will's pitched blue card,
	// Daze's Island, escape's N other cards from the graveyard.
	// Exactly CardPaymentCount entries, or none.
	AltCostIDs   []string     `json:"alt_cost_ids,omitempty"`
	Targets      []targetWire `json:"targets,omitempty"`
	Modes        []int        `json:"modes,omitempty"`
	XValue       int          `json:"x_value,omitempty"`
	DiscardIDs   []string     `json:"discard_ids,omitempty"`
	SacrificeIDs []string     `json:"sacrifice_ids,omitempty"`
	// OptionalCosts are the optional additional costs this move pays
	// (CR 601.2b, ADR 0073), as positions in the card's OptionalCosts
	// slice, repeated once per payment for a multikicker.
	OptionalCosts []int `json:"optional_costs,omitempty"`
	Strict        bool  `json:"strict,omitempty"`
	AutoTap       bool  `json:"auto_tap,omitempty"`
	// Face is the printed face being cast or played (ADR 0034).
	// Omitted — the front — for every single-faced card.
	Face int `json:"face,omitempty"`
}

// castZones are the five places a cast or a land play can come out
// of, in the order the moves are emitted. ONE walk since #673: the
// zone a card is in decides which surfaces have to be checked, never
// which kinds of cost are available, because "where may I cast this
// from" and "what may I pay for it" are separate questions the engine
// answers in separate functions (cast_zones.go).
func (e *enumerator) castZones() []struct {
	z    *game.Zone
	kind game.ZoneKind
	from string
} {
	g, p := e.g, e.p
	return []struct {
		z    *game.Zone
		kind game.ZoneKind
		from string
	}{
		{p.Hand, game.ZoneHand, "hand"},
		{p.Command, game.ZoneCommand, "command"},
		{p.Graveyard, game.ZoneGraveyard, "graveyard"},
		{g.Exile, game.ZoneExile, "exile"},
		{p.Library, game.ZoneLibrary, "library"},
	}
}

// castMoves enumerates land plays and spell casts from EVERY zone the
// seat can cast out of — hand, the command zone, its graveyard, exile
// and the top of its library — at every price the engine would accept
// for each (#673).
//
// Two questions, two engine functions, and the whole of this method is
// asking them in a loop:
//
//   - MAY this card be cast from this zone: the card's own
//     declaration (CastableZonesFor — flashback, escape, Gravecrawler),
//     or a granted CastPermission (ADR 0066 — Snapcaster, cascade,
//     impulse exile, warp's recast, a foretold or suspended card).
//   - AT WHAT PRICE: game.CastOffersForLocked, which lists the printed
//     mana cost and every alternative cost claimable from that zone,
//     already filtered to the ones payable right now.
//
// Both are the functions CastSpell validates with, so a bot can never
// be offered a cast the engine will refuse, at a zone it will refuse,
// or at a price it will not charge.
func (e *enumerator) castMoves() {
	g := e.g
	if g.SplitSecondActive {
		return
	}
	speed := sorcerySpeedOpen(g, e.seat)
	// #500: the allowance is the player's, not a literal one — a
	// controlled Exploration or a one-turn grant raises it. Same
	// helper the engine's own refusal reads, so the enumerator can
	// never offer a land play CastSpell will reject.
	landOwed := g.LandDropsRemainingLocked(e.seat) > 0
	// The fast negative, and the same one the view takes: in almost
	// every game nothing grants anything. It no longer gates the
	// GRAVEYARD walk, because a card's own flashback or escape opens
	// that zone with nothing granted — but exile and the library are
	// reachable only through a grant (cast_zones.go), so skipping
	// them costs nothing and an enumeration runs on every decision.
	anyGrant := g.AnyCastPermissionsForEffect()

	for _, zone := range e.castZones() {
		if zone.z == nil {
			continue
		}
		cards := zone.z.Cards
		switch zone.kind {
		case game.ZoneExile:
			if !anyGrant {
				continue
			}
		case game.ZoneLibrary:
			// CR 401.5: only the top card is ever open, and the top is
			// the LAST element. Checking one card rather than walking
			// the library also keeps this loop from touching hidden
			// information it has no business reading.
			if !anyGrant || len(cards) == 0 {
				continue
			}
			cards = cards[len(cards)-1:]
		}
		for _, c := range cards {
			e.castMovesFromZone(c, zone.kind, zone.from, speed, landOwed, anyGrant)
		}
	}
}

// castMovesFromZone expands ONE card sitting in ONE zone: its faces,
// its land play, and one call into castMovesForCard per price the
// engine would let this seat announce.
func (e *enumerator) castMovesFromZone(c game.Card, kind game.ZoneKind, from string, speed, landOwed, anyGrant bool) {
	g := e.g
	// CastPermissionForLocked answers nil unless the window is open
	// for this seat, so there is no second liveness test here — one
	// function reads the duration (#945).
	var perm *game.CastPermission
	if anyGrant {
		perm = g.CastPermissionForLocked(e.seat, c, kind)
	}
	// The per-card fast negative for the three zones that are a cast
	// surface only sometimes. CastOffersForLocked would answer with an
	// empty list anyway — this just declines to build one for every
	// card in a thirty-card graveyard on every bot decision.
	switch kind {
	case game.ZoneGraveyard, game.ZoneExile, game.ZoneLibrary:
		if perm == nil && !castableFromZoneAnyFace(c, kind) {
			return
		}
	}
	// ADR 0034: a modal DFC is two playable objects sharing one
	// instance, and since #719 so is an ADVENTURE card — CR 715.3 lets
	// the caster choose the creature or the Adventure — so enumerate
	// each face as its own move and let the bot pick between them.
	// CastableFaces returns [0] for everything else, so this loop runs
	// once for every single-faced card and the enumeration is
	// unchanged for them.
	//
	// A grant that NAMES faces NARROWS the choice to exactly those (a
	// defeated Siege's back, CR 715.4's "cast the creature from
	// exile"), which is the same rule faceForCastLocked applies — so
	// the enumerator cannot offer a half the announce path refuses.
	faces := c.CastableFaces()
	if granted, ok := perm.GrantsFaces(e.seat); ok {
		faces = granted
	}
	for _, face := range faces {
		// The face is materialised onto a COPY, exactly as CastSpell
		// does, so all the type, cost and catalog reads below see the
		// chosen half without any of them learning about faces.
		card := c
		card.SetFace(face)
		if card.IsLand() {
			e.landPlayMove(card, kind, from, speed, landOwed, perm)
			continue
		}
		for _, offer := range g.CastOffersForLocked(e.seat, card, kind, perm) {
			// The bot POLICY on top of the rule, and the only thing
			// this package adds to the engine's answer: CR 119.4 lets
			// a player pay life down to exactly zero, the next
			// state-based check then kills them, and a bot offered
			// that line would take it.
			if offer != nil && offer.Life > 0 && e.p.Life <= offer.Life {
				continue
			}
			e.castMovesForCard(card, from, speed, perm, offer)
		}
	}
}

// castableFromZoneAnyFace reports whether ANY castable face of the
// card declares `kind` a cast surface. One question per face rather
// than one for the card, because an MDFC's halves are separate
// catalog entries and only the back may print flashback.
func castableFromZoneAnyFace(c game.Card, kind game.ZoneKind) bool {
	for _, face := range c.CastableFaces() {
		probe := c
		probe.SetFace(face)
		if game.CardCastableFromZone(game.CatalogKey(probe), kind) {
			return true
		}
	}
	return false
}

// landPlayMove emits the one move for playing a land out of `kind`,
// when CR 305 allows it.
func (e *enumerator) landPlayMove(card game.Card, kind game.ZoneKind, from string, speed, landOwed bool, perm *game.CastPermission) {
	// CR 305: main phase, empty stack, your turn, and the per-turn
	// land-play allowance. The engine enforces all four since #500;
	// the check stays here so a bot is never OFFERED a move that would
	// be refused.
	if !speed || !landOwed {
		return
	}
	// CR 903.4 is a permission to CAST a commander, not to play a land
	// out of the command zone, and no other rule opens that door.
	if kind == game.ZoneCommand {
		return
	}
	// CR 305.1: playing a land is not casting, so a cast-only
	// permission strands it.
	if perm != nil && perm.CastOnly {
		return
	}
	label := "Play " + card.Name
	if kind != game.ZoneHand {
		label += " from " + from
	}
	e.add(Move{
		Type:   TypeCastSpell,
		Player: e.seat,
		Kind:   KindLand,
		Label:  label,
		Source: card.InstanceID,
		Params: mustJSON(castParams{
			InstanceID: card.InstanceID.String(),
			FromZone:   from,
			Face:       card.ActiveFace,
		}),
	})
}

// castMovesForCard expands one non-land card at ONE announced price
// into concrete casts: every legal (modes × targets × cost payment)
// combination the seat can afford, capped at MaxExpansionPerSource.
//
// `offer` is the CR 118.9 cost this expansion pays — nil for the
// printed mana cost, otherwise one of the entries
// game.CastOffersForLocked listed for this card in this zone. The
// caller loops the offers; each is its own set of moves, because a
// flashed-back Faithless Looting and a hard-cast one are different
// prices with different consequences, exactly as a kicked and an
// unkicked cast are (ADR 0073 §9).
func (e *enumerator) castMovesForCard(card game.Card, from string, speed bool, perm *game.CastPermission, offer *game.AlternativeCost) {
	// ADR 0073 §9: a card with optional additional costs is several
	// casts, not one — an unkicked Burst Lightning and a kicked one
	// are different moves at different prices with different effects,
	// and a bot that was only ever offered the cheap one would never
	// kick anything. Each announced set walks the whole expansion
	// below on its own.
	//
	// optionalCostSets is [nil] for every card that offers none,
	// which is every card in the catalog before #664 — so this loop
	// runs exactly once for them and the enumeration is unchanged.
	for _, chosen := range optionalCostSets(game.OptionalCostsFor(game.CatalogKey(card)), e.opts.MaxExpansionPerSource) {
		e.castMovesPayingOptional(card, from, speed, perm, offer, chosen)
	}
}

// maxEnumeratedCostPayments caps how many ways the enumerator will
// offer to pay the CARD-shaped half of one alternative cost — which
// blue card Force of Will pitches, which five cards an Uro escapes
// with. Not a rule; a policy, documented in docs/bot.md beside
// maxEnumeratedRepeats.
//
// ONE, and the reason is ADR 0033 §1's corollary rather than
// squeamishness about combinatorics: a variable in a COST must not
// become an arity of the target/mode cross product. Escape-five over
// a twenty-card graveyard is 15,504 payments, every one of which
// would have to be priced and every one of which is the same spell
// with the same targets — so a cap of twelve would spend the entire
// per-source budget on twelve indistinguishable Uros and never offer
// the second target of anything.
//
// "Indistinguishable" is the load-bearing word, and it is a statement
// about the POLICY rather than about Magic. The heuristic prices the
// battlefield and the seats; a card in a graveyard or a hand has no
// value in its evaluation at all (aiseat/heuristic/score.go), so it
// cannot tell two escape payments apart and would pick between them
// by index. When a policy learns to price the cards a cost eats, this
// constant is where the search is widened.
const maxEnumeratedCostPayments = 1

// maxEnumeratedRepeats caps how many times the enumerator will offer
// to pay one REPEATABLE optional cost (multikicker) in a single
// announcement. Not a rule — multikicker is unbounded in paper — but
// the expansion here is already modes × targets × cost payments, and
// multiplying it by the seat's available mana is how a bot decision
// stops terminating. Documented as policy in docs/bot.md.
const maxEnumeratedRepeats = 3

// optionalCostSets is the bot's announcement policy for optional
// additional costs (ADR 0073 §9): decline everything, or pay exactly
// ONE of the offered costs — once, or up to maxEnumeratedRepeats
// times for a repeatable one.
//
// NOT enumerated, deliberately and for the reason escape's card
// component is not: paying TWO different optional costs at once
// (Thornscape Battlemage's "Kicker {R} and/or {W}") is a power-set
// search whose every member also needs its own affordability probe,
// and no card in the catalog offers two. A bot simply does not take
// that line yet; it is never offered one it cannot pay for.
func optionalCostSets(costs []game.AdditionalCost, budget int) [][]int {
	if len(costs) == 0 {
		return [][]int{nil}
	}
	out := [][]int{nil}
	for i := range costs {
		reps := min(costs[i].MaxPayments(), maxEnumeratedRepeats)
		for k := 1; k <= reps; k++ {
			if budget > 0 && len(out) >= budget {
				return out
			}
			set := make([]int, k)
			for j := range set {
				set[j] = i
			}
			out = append(out, set)
		}
	}
	return out
}

// optionalCostLabel spells the announced optional costs into the
// move's label — " (Kicker {4})", " (Multikicker {G} ×3)" — so the
// kicked and unkicked casts of one card are distinguishable in the
// move log and in a bot-eval trace. Empty for a move that pays none.
func optionalCostLabel(optional []game.AdditionalCost, chosen []int) string {
	if len(chosen) == 0 {
		return ""
	}
	counts := make(map[int]int, len(chosen))
	for _, i := range chosen {
		counts[i]++
	}
	out := ""
	for i := range optional {
		n, ok := counts[i]
		if !ok {
			continue
		}
		if out != "" {
			out += ", "
		}
		out += optional[i].Label
		if n > 1 {
			out += fmt.Sprintf(" ×%d", n)
		}
	}
	if out == "" {
		return ""
	}
	return " (" + out + ")"
}

// costPaymentDemands sums what one cast owes in CARDS across its
// CR 601.2f payment plan (ADR 0073 §4): the mandatory additional cost
// plus each announced optional one. Reports the total discard count
// and the single sacrifice clause to expand.
//
// `ok` is false when the plan carries TWO sacrifice clauses, which
// this package does not enumerate — see the call site.
func costPaymentDemands(mandatory *game.AdditionalCost, optional []game.AdditionalCost, chosen []int) (discards int, sacrifice *game.TargetSpec, ok bool) {
	add := func(c *game.AdditionalCost) bool {
		if c == nil || c.Empty() {
			return true
		}
		discards += c.DiscardCards
		if c.Sacrifice == nil {
			return true
		}
		if sacrifice != nil {
			return false
		}
		sacrifice = c.Sacrifice
		return true
	}
	if !add(mandatory) {
		return 0, nil, false
	}
	for _, i := range chosen {
		if i < 0 || i >= len(optional) {
			continue
		}
		if !add(&optional[i]) {
			return 0, nil, false
		}
	}
	return discards, sacrifice, true
}

// castMovesPayingOptional is castMovesForCard for ONE announced set
// of optional additional costs — the unkicked cast, or the kicked
// one. `chosen` is nil for every card that offers none.
func (e *enumerator) castMovesPayingOptional(card game.Card, from string, speed bool, perm *game.CastPermission, offer *game.AlternativeCost, chosen []int) {
	g, p := e.g, e.p
	// #662: the spell IS its own source (CR 702.16b), so every legal
	// set below is computed against the card's colour and type. An
	// enumerator that passed only the seat would offer the bot a
	// pro-red creature for its red spell and the server would refuse
	// the move — the #347 / #544 failure mode.
	castSrc := game.SourceObject(e.seat, &card)

	// Timing (CR 307.1 / 702.8): instants and flash any time the seat
	// holds priority; everything else needs the sorcery-speed window.
	// ADR 0066: a granted permission may say otherwise, and the engine
	// reads the same field in the same order, so the two cannot drift.
	requiresSorcerySpeed := !card.IsInstant() && !game.HasKeyword(&card, "flash")
	if perm != nil {
		switch perm.Timing {
		case game.TimingFlash:
			requiresSorcerySpeed = false
		case game.TimingSorcery:
			requiresSorcerySpeed = true
		}
	}
	if requiresSorcerySpeed && !speed {
		return
	}

	// The clause this cast announces under, which is the offer's
	// business as much as the card's: overload DELETES the target
	// clause ("change 'target' to 'each'") and cleave SWAPS it for a
	// wider one, and the engine reads the same function at announce
	// (CR 601.2c) and again at resolution. An enumerator that offered
	// an overloaded Cyclonic Rift with a target would have every one
	// of those moves refused with ErrInvalidParam.
	cardSpec := game.TargetSpecUnderAlternativeCost(game.TargetSpecFor(game.CatalogKey(card)), offer)

	// A card the catalog marks as targeted the S13.1 way (free-form
	// target_mode, no structured spec) cannot be enumerated: the
	// engine demands a target but nothing says which are legal. Under
	// an offer that clears the clause there is no target to demand,
	// so the cast is enumerable after all.
	if game.TargetModeFor(game.CatalogKey(card)) != "" && cardSpec == nil && !offer.Clears() {
		return
	}

	// Cost. A cost the parser can't read is not enumerable: since
	// #289 the engine rejects such a cast with ErrUnparseableCost
	// (split and adventure cards import a joined "{1}{R} // {1}{U}"),
	// so offering the move would hand the client an action that is
	// guaranteed to fail. Enumerating it free — what this used to
	// do, mirroring the engine's old silent downgrade — is worse:
	// it advertises a free spell that isn't one.
	//
	// #696: THE engine's pricer, called with the announcement this
	// move will actually send. It settles the CR 118.9 swap (the
	// permission's price, when it has one, IS the cost this cast pays
	// — ADR 0066), the commander tax, the mana half of the announced
	// optional costs (ADR 0073 §3) and the "spend mana as though any
	// colour" fold, in that order, so a bot is never offered a cast at
	// a price the engine will not charge. A copy of the first three
	// used to live here.
	//
	// #673: `offer` is the entry game.CastOffersForLocked listed for
	// this card in this zone — the card's own flashback or overload as
	// readily as the one a permission synthesises — and the pricer
	// takes it off the announcement below, so both kinds reach the
	// same walk.
	fromZone := game.ZoneHand
	switch from {
	case "command":
		fromZone = game.ZoneCommand
	case "graveyard":
		fromZone = game.ZoneGraveyard
	case "exile":
		fromZone = game.ZoneExile
	case "library":
		fromZone = game.ZoneLibrary
	}
	// The announcement, built once and reused for every repricing
	// below — the target-set loop reprices against it and changes
	// nothing but Targets.
	announce := game.CastSpellParams{
		FromZone:        from,
		AlternativeCost: offerKey(offer),
		OptionalCosts:   chosen,
		// ADR 0034: `card` has already had SetFace applied by the
		// caller, so this is the face the move announces.
		Face: card.ActiveFace,
	}
	price, err := e.g.PriceCastForEffect(e.seat, card, announce)
	if err != nil {
		return
	}
	// CastPrice.Base, not .Total: the convoke / waterbend subtraction
	// is a PAYMENT (CR 601.2h) and this package enumerates no tap
	// payments, so the cost it searches an X against must still carry
	// its {X} slot. Reading Total would settle X into generic and
	// Chord of Calling would only ever be offered at X=0.
	cost := price.Base
	// CR 118.6: a spell with no mana cost (Ancestral Vision, Living
	// End) can't be cast by paying it, and ParseCost reads that empty
	// string as a free {0}. Every move this function builds pays the
	// printed cost, so none of them is legal; the engine refuses the
	// cast with ErrNoManaCost.
	if game.HasNoManaCost(card) && offer == nil && (perm == nil || perm.Cost == "") {
		return
	}
	optional := game.OptionalCostsFor(game.CatalogKey(card))
	// S28: the board's cost modifiers (CR 601.2f). Same reasoning as
	// the parse gate above — a move enumerated at the printed price
	// while a Sphere of Resistance sits on the table is a move the
	// engine will reject for insufficient mana, and a bot that keeps
	// picking rejected moves stalls. A modifier the engine refuses to
	// price (ErrCostModifier) makes the cast unenumerable for the
	// same reason an unparseable cost does.
	//
	// Applied below rather than read off price.Total because this
	// package prices the SAME cast once per candidate target set (a
	// per-target surcharge, ADR 0048 addendum §14), and they all apply
	// to the one pre-modifier total above.
	//
	// Priced at X=0 even though affordableX is about to search for a
	// bigger X. The only modifier kind X can change the answer for is
	// a CostFloor (Trinisphere), and X=0 is the branch where the
	// floor applies — so the search starts from the most expensive
	// reading and can only narrow the X it offers. Conservative in
	// the direction that never advertises an unaffordable move.
	// #760, ADR 0073 §7: the one announce-time cast gate, the twin of
	// the split-second check at the top of castMoves and beside it
	// for the same reason — a bot offered a move the engine will
	// refuse keeps picking it and stalls. THE SAME FUNCTION CastSpell
	// calls, so the two cannot disagree about what is banned.
	//
	// The announced optional costs ride along because the gate takes
	// the announcement: CR 601.3a lets a choice made while proposing
	// the spell lift a ban, and this is the choice that has been made
	// by now.
	if err := g.CastGateLocked(e.seat, card, fromZone, announce); err != nil {
		return
	}
	// ADR 0048 addendum §14: when something on the board or the card
	// itself prices by target (Fireball's surcharge, Price of Fame's
	// discount), one price up front is not the price — and a
	// nil-targets price cannot even be used as a gate, because a
	// target-reading REDUCTION makes the real cast cheaper than it.
	// So the up-front price and its early return run only when
	// nothing reads targets, which is every board without such a
	// card; otherwise each (modes, targets) set below is priced on its
	// own and carries its own X.
	spend := game.ManaSpendForCast(card)
	perTarget := e.g.CastPriceReadsTargetsForEffect(card)
	// Additional costs (CR 601.2f). Read before the X search because
	// one of them can PRICE X: Toxic Deluge's "pay X life" is the
	// whole of its X, and the mana cost it prints has no {X} slot at
	// all (#957). The payments themselves are expanded below.
	addCost := game.AdditionalCostFor(game.CatalogKey(card))
	// #810: the one X rule. A spell whose whole effect is X (Fireball,
	// Stroke of Genius) is not offered at X=0, where it would resolve
	// for nothing; a spell with a fixed rider still is. #957 is the
	// same rule reaching the non-mana half of the price: the floor is
	// the same one, and xCeilingFromCost is what a "pay X life" cost
	// can pay for (CR 119.4, less one so the seat survives its own
	// sweep). Both live in x.go.
	xFloor := enumeratedXFloor(game.CatalogKey(card), 0)
	xLifeCeiling := xCeilingFromCost(addCost, p.Life)
	x := 0
	if !perTarget {
		priced, err := e.g.ApplyCostModifiersForEffect(cost, game.CostQuery{
			Card:       card,
			Controller: e.seat,
			FromZone:   fromZone,
		})
		if err != nil {
			return
		}
		var ok bool
		x, ok = e.announcedX(priced, spend, xFloor, xLifeCeiling)
		if !ok {
			return
		}
	}

	// Modes → each choice of modes yields its own clause list, and
	// each clause its own picks (#764). Options with no legal target
	// are dropped before any combination is built, so the budget is
	// never spent on selections the engine would refuse (ADR 0065
	// §6).
	modeSpec := game.ModeSpecFor(game.CatalogKey(card))
	modeSets := [][]int{nil}
	if modeSpec != nil {
		modeSets = e.legalModeSets(castSrc, modeSpec)
		if len(modeSets) == 0 {
			return
		}
	}

	// The additional cost's PAYMENTS (CR 601.2f), read above. Discards
	// choose from the rest of the hand; a sacrifice chooses from the
	// seat's own permanents matching the clause. A pay-X-life needs no
	// payment set of its own — the announced X is the payment, and it
	// was priced with the rest of X above.
	//
	// ADR 0073: the demands are summed across the mandatory cost AND
	// whichever optional ones this move announces, so a kicked
	// Gatekeeper of Malakir expands its kicker's sacrifice exactly as
	// a mandatory one would. The wire lists are flat and the engine
	// walks them in the same order.
	discards, sacrifice, ok := costPaymentDemands(addCost, optional, chosen)
	if !ok {
		// Two card-shaped sacrifice clauses on one cast (a mandatory
		// one AND a kicker's). Not enumerated: the two pools have to
		// be searched together and written into one flat list, and no
		// card in the catalog asks for it. The bot does not take the
		// line; it is never offered one it cannot pay.
		return
	}
	// CR 601.2b's card component of the ALTERNATIVE cost, which is a
	// different list on the wire from the additional cost's (they are
	// paid at different steps and one of them vanishes when the offer
	// is declined — see CastSpellParams.AltCostIDs).
	//
	// The candidates come from the engine's own acceptance predicate,
	// so a payment this builds is a payment
	// validateAlternativeCostPaymentLocked accepts; and the whole
	// search collapses to ONE payment by policy — see
	// maxEnumeratedCostPayments.
	altCostSets := [][]uuid.UUID{nil}
	if want := offer.CardPaymentCount(); want > 0 {
		pool := g.AltCostCandidatesLocked(e.seat, card.InstanceID, offer)
		altCostSets = combinations(pool, want, want, maxEnumeratedCostPayments)
		if len(altCostSets) == 0 {
			// Unreachable through CastOffersForLocked, which already
			// dropped an offer with too few candidates. Kept because
			// this function is the one that writes the payment: a
			// future caller that skips the offer filter must not be
			// able to emit a cast with an unpayable cost.
			return
		}
	}
	discardSets := [][]uuid.UUID{nil}
	sacrificeSets := [][]uuid.UUID{nil}
	if discards > 0 {
		var pool []uuid.UUID
		for _, h := range p.Hand.Cards {
			if h.InstanceID != card.InstanceID {
				pool = append(pool, h.InstanceID)
			}
		}
		discardSets = combinations(pool, discards, discards, e.opts.MaxExpansionPerSource)
		if len(discardSets) == 0 {
			return
		}
	}
	if sacrifice != nil {
		// Cost, not target — see SpecCandidatesForEffect.
		lt := g.SpecCandidatesForEffect(e.seat, sacrifice)
		var pool []uuid.UUID
		for _, id := range lt.Cards {
			if c := findBattlefield(g, id); c != nil && c.Controller == e.seat {
				pool = append(pool, id)
			}
		}
		// #747: N from the clause, one payment per cast for N ≥ 2,
		// nothing offered when the caster controls fewer than N.
		sacrificeSets = e.sacrificePayments(pool, sacrifice, uuid.Nil)
		if len(sacrificeSets) == 0 {
			return
		}
	}

	budget := e.opts.MaxExpansionPerSource
	for _, modes := range modeSets {
		// The budget is spent MODES-outermost: every mode selection
		// gets at least one target set before any gets a second, so a
		// bot is never offered only the first bullet of a charm
		// (ADR 0065 §6).
		steps := game.AnnouncedClauses(cardSpec, modeSpec, modes)
		// #619, CR 601.2c. Crackle with Power's target count IS X
		// ("deals five times X damage to each of up to X targets"),
		// and the announce path refuses any cast where the two
		// disagree. So for such a step the enumerator does not pick an
		// X and a target list independently — it picks the targets,
		// and the X it announces is how many it picked.
		xSteps := stepsCountedByX(steps)
		if len(xSteps) > 0 {
			if cost.XSlots == 0 {
				// The step's X is announced by a cost this package
				// cannot price — Waterbender's Restoration's
				// waterbend {X}, paid by tapping artifacts and
				// creatures. The only announcement the enumerator
				// could make for it is X=0, which buys no targets and
				// a spell that does nothing, and any larger one would
				// be a move the engine refuses for an unpaid cost. So
				// the cast is not enumerable, the same answer an
				// unparseable cost gets.
				continue
			}
			// The arities to generate are bounded by the largest X
			// the seat could announce. Under a per-target price that
			// is not known until each set is priced, so open the step
			// to MaxX and let the affordability check below drop what
			// cannot be paid; the budget caps the expansion either
			// way.
			bound := x
			if perTarget {
				bound = e.opts.MaxX
			}
			if bound < 1 {
				continue
			}
			steps = openXCountedSteps(steps, bound)
		}
		targetSets := e.legalStepSets(castSrc, steps, budget)
		if len(targetSets) == 0 {
			continue
		}
		for _, targets := range targetSets {
			setX := x
			if perTarget {
				// §14: priced with this set's targets. An unaffordable
				// set is skipped before any budget is spent on it, so
				// a Fireball the seat can pay for at one target is not
				// crowded out by the three-target sets it cannot.
				priced, err := e.g.ApplyCostModifiersForEffect(cost, game.CostQuery{
					Card:       card,
					Controller: e.seat,
					FromZone:   fromZone,
					Targets:    targets,
				})
				if err != nil {
					continue
				}
				var ok bool
				setX, ok = e.announcedX(priced, spend, xFloor, xLifeCeiling)
				if !ok {
					continue
				}
			}
			if len(xSteps) > 0 {
				// setX is the largest announcement this seat can pay
				// for; a set that filled more X-counted slots than
				// that is a cast it cannot make. The cost is monotonic
				// in X, so the comparison is the whole affordability
				// check.
				k, ok := announcedXCount(steps, xSteps, targets)
				if !ok || k < 1 || k > setX {
					continue
				}
				setX = k
			}
			for _, altPaid := range altCostSets {
				for _, discards := range discardSets {
					for _, sacs := range sacrificeSets {
						if budget <= 0 {
							return
						}
						budget--
						label := "Cast " + card.Name
						switch from {
						case "command":
							label += " from the command zone"
						case "graveyard", "exile", "library":
							label += " from " + from
						}
						if setX > 0 {
							label += fmt.Sprintf(" for X=%d", setX)
						}
						// #673: the price. A flashed-back Faithless
						// Looting, an escaped Uro and a hard-cast one
						// are otherwise the same line in the move log,
						// and a bot eval that cannot tell them apart
						// cannot explain why the bot escaped.
						label += altCostLabel(g, offer, altPaid)
						// ADR 0073: the kicked and unkicked casts are
						// otherwise the same line in the move log, and a
						// bot eval that cannot tell them apart cannot
						// explain why the bot kicked.
						label += optionalCostLabel(optional, chosen)
						label += targetLabel(g, targets)
						e.add(Move{
							Type:   TypeCastSpell,
							Player: e.seat,
							Kind:   KindCast,
							Label:  label,
							Source: card.InstanceID,
							// CR 119.4: the life half of the offer is a
							// price Params cannot name, so a policy
							// reading only the payload would price
							// Force of Will's pitch as free. See
							// MoveCost.
							Cost: moveCost(offerLife(offer), 0),
							Params: mustJSON(castParams{
								InstanceID:      card.InstanceID.String(),
								FromZone:        from,
								AlternativeCost: offerKey(offer),
								AltCostIDs:      idStrings(altPaid),
								Targets:         wireTargets(targets),
								Modes:           modes,
								XValue:          setX,
								DiscardIDs:      idStrings(discards),
								SacrificeIDs:    idStrings(sacs),
								OptionalCosts:   chosen,
								Strict:          true,
								AutoTap:         true,
								// ADR 0034: `card` has already had
								// SetFace applied by the caller, so
								// ActiveFace IS the face this move casts.
								Face: card.ActiveFace,
							}),
						})
					}
				}
			}
		}
	}
}

// affordableXFrom reports whether the seat can pay cost right now —
// from the floating pool, or by the auto-tapper's plan — and, for an
// {X} cost, the largest X it can pay, at or above `floor`, up to MaxX.
// Mirrors the engine's auto_tap + strict path: the pool is consulted
// first, then a plan is sought for the WHOLE cost (the engine does not
// net floating mana against the plan).
//
// The floor is the announcement's lower bound, and it has two sources,
// both settled by enumeratedXFloor (x.go) before the call: a printed
// "X can't be 0" (Helm of Obedience), and #810's rule that a move
// whose whole effect is X is not worth offering at X=0. A seat that
// cannot pay for the floor has no move at all rather than a free one.
//
// The scan still starts at the floor and still breaks on the first
// unaffordable value, because the cost is monotonic in X: every
// extra point of X buys the same XSlots generic symbols. Nothing
// here enumerates a RANGE — exactly one X comes back, so X never
// enters an expansion cross product (see activatedMoves for why
// that matters).
func (e *enumerator) affordableXFrom(cost game.ParsedCost, spend game.ManaSpendContext, floor int) (int, bool) {
	return e.affordableXExcluding(cost, spend, floor, nil)
}

// affordableXExcluding is affordableXFrom with sources the payment
// may not use — the one caller is an activated ability whose cost
// includes {T}, which cannot tap its own source for mana.
func (e *enumerator) affordableXExcluding(
	cost game.ParsedCost,
	spend game.ManaSpendContext,
	floor int,
	excluded map[uuid.UUID]bool,
) (int, bool) {
	if cost.XSlots == 0 {
		// No {X}: the floor is meaningless and X is always zero.
		return 0, e.canPayExcluding(cost, 0, spend, excluded)
	}
	if floor < 0 {
		floor = 0
	}
	best, ok := -1, false
	for x := floor; x <= e.opts.MaxX; x++ {
		if e.canPayExcluding(cost, x, spend, excluded) {
			best, ok = x, true
			continue
		}
		break
	}
	return best, ok
}

// `spend` is the #352 spend context — what the mana would be paid
// for — so restricted mana in the pool counts toward a cast it may
// legally fund and toward no other.
func (e *enumerator) canPay(cost game.ParsedCost, x int, spend game.ManaSpendContext) bool {
	return e.canPayExcluding(cost, x, spend, nil)
}

// canPayExcluding is canPay with sources the auto-tapper may not
// reach for. It has to mirror exactly what the engine excludes, or
// the enumerator's answer and the engine's answer disagree — which
// is the #544 failure mode, one cost component over.
func (e *enumerator) canPayExcluding(
	cost game.ParsedCost,
	x int,
	spend game.ManaSpendContext,
	excluded map[uuid.UUID]bool,
) bool {
	if e.p.ManaPool.CanPayFor(cost, x, spend) {
		return true
	}
	_, ok := e.g.AutoTapForCostForEffectExcluding(e.seat, cost, x, excluded)
	return ok
}

// legalModeSets lists every distinct mode selection of size
// Min..Max, excluding any with more than one targeted option (the
// engine's castTargetSpec rejects those).
func (e *enumerator) legalModeSets(src game.TargetSource, ms *game.ModeSpec) [][]int {
	// ADR 0065 §6, "prefer the modes that have legal targets": an
	// option whose clause cannot be filled is dropped before any
	// combination is built, so the budget never goes on a selection
	// the engine would refuse at announce.
	options := e.g.ChoosableModeOptionsForEffect(src, ms)
	if !game.EnoughChoosableModes(len(options), ms) {
		return nil
	}
	hi := ms.Max
	if hi <= 0 || (!ms.Repeatable && hi > len(options)) {
		hi = len(options)
	}
	budget := e.opts.MaxExpansionPerSource
	var out [][]int
	add := func(sel []int) bool {
		out = append(out, append([]int(nil), sel...))
		return len(out) < budget
	}
	if ms.Min == 0 {
		if !add(nil) {
			return out
		}
	}
	lo := ms.Min
	if lo < 1 {
		lo = 1
	}
	// CR 700.2d: the all-one-option selections first, so a
	// repeatable spec whose only legal option is one mode is not
	// crowded out by mixed multisets.
	if ms.Repeatable {
		for _, opt := range options {
			for n := lo; n <= hi; n++ {
				sel := make([]int, n)
				for i := range sel {
					sel[i] = opt
				}
				if !add(sel) {
					return out
				}
			}
		}
	}
	var rec func(start int, cur []int) bool
	rec = func(start int, cur []int) bool {
		if len(cur) >= lo && len(cur) <= hi {
			if !(ms.Repeatable && len(cur) == 1) && !add(cur) {
				return false
			}
		}
		if len(cur) == hi {
			return true
		}
		for i := start; i < len(options); i++ {
			if !rec(i+1, append(cur, options[i])) {
				return false
			}
		}
		return true
	}
	rec(0, nil)
	return out
}

// legalStepSets is the cartesian product of each clause's legal
// picks, in step order, capped at `budget` (#764, ADR 0065 §6). An
// announcement with no steps yields the single empty set, which is
// how an untargeted cast stays one move.
func (e *enumerator) legalStepSets(src game.TargetSource, steps []game.AnnouncedClause, budget int) [][]game.TargetRef {
	out := [][]game.TargetRef{nil}
	for i := range steps {
		clause := steps[i].Clause
		picks := e.legalTargetSets(src, &clause, budget)
		if len(picks) == 0 {
			return nil
		}
		next := make([][]game.TargetRef, 0, budget)
		for _, prefix := range out {
			for _, pick := range picks {
				if len(next) >= budget {
					break
				}
				combined := append([]game.TargetRef(nil), prefix...)
				skip := false
				for _, p := range pick {
					p.Mode, p.Slot = steps[i].Mode, steps[i].Slot
					// CR 601.2c: a Distinct clause may not repeat an
					// object an earlier clause took.
					if clause.Distinct {
						for _, seen := range prefix {
							if seen.ID == p.ID {
								skip = true
							}
						}
					}
					combined = append(combined, p)
				}
				if skip {
					continue
				}
				next = append(next, combined)
			}
		}
		if len(next) == 0 {
			return nil
		}
		out = next
	}
	return out
}

// stepsCountedByX lists the indexes of the announcement's steps whose
// target count is the announced X rather than a printed constant
// (Crackle with Power, Doppelgang, Heliod's Intervention's first
// mode). Empty — the overwhelming majority — means nothing here ties
// X to anything.
func stepsCountedByX(steps []game.AnnouncedClause) []int {
	var out []int
	for i := range steps {
		if steps[i].Clause.CountFromX {
			out = append(out, i)
		}
	}
	return out
}

// openXCountedSteps returns a copy of `steps` with every X-counted
// clause opened to 1..bound, so legalStepSets expands one target set
// per arity and the caller can read the arity back off the set.
//
// The engine's own resolveStepCountsFromX pins the same clauses to a
// single announced X; this is that operation with the X still unknown,
// which is the whole difference between validating an announcement and
// building one.
//
// The floor is 1, not 0. X=0 on such a step is a spell cast with no
// targets, which for every card in the catalog that has one means a
// spell that does nothing — #810's rule, and the reason this needs no
// zero case of its own. A CountFromX card with a fixed rider would
// want one; none exists, and x.go is where that would be decided.
func openXCountedSteps(steps []game.AnnouncedClause, bound int) []game.AnnouncedClause {
	out := append([]game.AnnouncedClause(nil), steps...)
	for _, i := range stepsCountedByX(out) {
		out[i].Clause.Min, out[i].Clause.Max = 1, bound
		out[i].Clause.CountFromX = false
	}
	return out
}

// announcedXCount reads the X a target set is announcing: how many
// refs answer the X-counted steps. False when two such steps disagree,
// which is a set no announcement could cover — one X, one count
// (CR 601.2c).
//
// Matched on (Mode, Slot), the coordinates legalStepSets stamps onto
// every ref, which is exactly what the engine's stepTargetCount reads.
func announcedXCount(steps []game.AnnouncedClause, xSteps []int, targets []game.TargetRef) (int, bool) {
	k := -1
	for _, i := range xSteps {
		n := 0
		for _, t := range targets {
			if t.Kind == game.TargetSelf || t.Kind == game.TargetNone {
				continue
			}
			if t.Mode == steps[i].Mode && t.Slot == steps[i].Slot {
				n++
			}
		}
		if k >= 0 && n != k {
			return 0, false
		}
		k = n
	}
	return k, k >= 0
}

// legalTargetSets expands a target clause into concrete target
// lists: the empty list when Min is 0, then every k-subset of the
// legal candidates for k in max(Min,1)..Max, players before cards,
// stopping at budget entries.
//
// `src` is the spell or ability doing the targeting, not just the
// seat: CR 702.16b tests protection against the SOURCE, so an
// enumerator that passed only a seat would offer the bot a pro-red
// creature as a target for its red spell and the server would then
// refuse the move — the #347 / #544 failure mode.
func (e *enumerator) legalTargetSets(src game.TargetSource, spec *game.TargetSpec, budget int) [][]game.TargetRef {
	lt := e.g.LegalTargetsForEffect(src, spec)
	cands := make([]game.TargetRef, 0, len(lt.Players)+len(lt.Cards))
	for _, id := range lt.Players {
		cands = append(cands, game.TargetRef{Kind: game.TargetPlayer, ID: id})
	}
	for _, id := range lt.Cards {
		cands = append(cands, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	lo, hi := spec.Min, spec.Max
	if hi <= 0 || hi > len(cands) {
		hi = len(cands)
	}
	var out [][]game.TargetRef
	if lo == 0 {
		out = append(out, nil)
	}
	if lo < 1 {
		lo = 1
	}
	if lo > len(cands) {
		return out
	}
	for k := lo; k <= hi && len(out) < budget; k++ {
		var rec func(start int, cur []game.TargetRef)
		rec = func(start int, cur []game.TargetRef) {
			if len(out) >= budget {
				return
			}
			if len(cur) == k {
				out = append(out, append([]game.TargetRef(nil), cur...))
				return
			}
			for i := start; i < len(cands); i++ {
				rec(i+1, append(cur, cands[i]))
			}
		}
		rec(0, nil)
	}
	return out
}

// combinations returns every k-subset of pool for k in min..max, in
// pool order, up to limit entries. Returns nil when the pool is too
// small for min.
func combinations(pool []uuid.UUID, lo, hi, limit int) [][]uuid.UUID {
	if lo > len(pool) {
		return nil
	}
	if hi > len(pool) {
		hi = len(pool)
	}
	var out [][]uuid.UUID
	for k := lo; k <= hi && len(out) < limit; k++ {
		var rec func(start int, cur []uuid.UUID)
		rec = func(start int, cur []uuid.UUID) {
			if len(out) >= limit {
				return
			}
			if len(cur) == k {
				out = append(out, append([]uuid.UUID(nil), cur...))
				return
			}
			for i := start; i < len(pool); i++ {
				rec(i+1, append(cur, pool[i]))
			}
		}
		rec(0, nil)
	}
	return out
}

// offerKey is the wire key of a resolved alternative cost, or "" when
// the cast pays the printed price. Nil-safe so the emit site needs no
// guard.
func offerKey(offer *game.AlternativeCost) string {
	if offer == nil {
		return ""
	}
	return offer.Key
}

// offerLife is the life an offer charges (CR 119.4), or 0. Nil-safe.
func offerLife(offer *game.AlternativeCost) int {
	if offer == nil {
		return 0
	}
	return offer.Life
}

// altCostLabel spells the claimed alternative cost into the move's
// label — " (Flashback {2}{R})", " (Escape—{1}{G}{U}, Exile five
// other cards from your graveyard, exiling Mountain, Opt, …)".
//
// Empty for a cast that pays the printed price, which is every cast
// in almost every game. The cards paid are NAMED rather than counted:
// #673 asks for it explicitly, and an escape line whose log entry did
// not say what it ate is unreviewable.
func altCostLabel(g *game.Game, offer *game.AlternativeCost, paid []uuid.UUID) string {
	if offer == nil {
		return ""
	}
	label := offer.Label
	if label == "" {
		label = offer.Key
	}
	if len(paid) > 0 {
		names := make([]string, 0, len(paid))
		for _, id := range paid {
			names = append(names, cardName(g, id))
		}
		// The verb is the component's, not a generic "paying": a Daze
		// that logged "exiling Island" would be describing a different
		// card.
		verb := ", exiling "
		if offer.ReturnToHand != nil {
			verb = ", returning "
		}
		label += verb + strings.Join(names, ", ")
	}
	return " (" + label + ")"
}
