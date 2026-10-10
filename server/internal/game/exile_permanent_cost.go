package game

import (
	"slices"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// exile_permanent_cost.go — #1600: exiling permanents you control as
// part of a cost (CR 118.3, CR 602.2b, CR 605.3a, CR 701.13a).
//
//	The Soul Stone      "{6}{B}, {T}, Exile a creature you control:
//	                    Harness The Soul Stone."
//	Altar of Bhaal      "{2}{B}, {T}, Exile a creature you control:
//	                    Return target creature card from your graveyard
//	                    to the battlefield."
//	City of Shadows     "{T}, Exile a creature you control: Put a
//	                    storage counter on this land."
//	Food Chain          "Exile a creature you control: Add X mana of
//	                    any one color, where X is 1 plus the exiled
//	                    creature's mana value." (a MANA ability)
//
// The battlefield sibling of ExileCost (exile_cost.go), which exiles
// CARDS out of the activator's hand or graveyard, and of ExileSelf,
// which exiles the source. This one names OTHER permanents the
// activator controls, so it is a picking problem, and it is the one
// AbilityCost.SacrificeOther and ReturnToHandCost solve: the
// activator's own permanents only, matching the clause, exactly Count
// of them, each once — one candidate walk that the engine, the protocol
// view's picker and the legal-move enumerator all read (#544).
//
// It differs from both, observably:
//
//   - **Not SacrificeOther with a destination.** A sacrifice goes
//     through sacrificePermanentLocked: EventSacrifice, and a dies
//     trigger when the permanent is a creature. Exiling is a plain zone
//     change (CR 701.13a). Nothing dies (exile is not a graveyard), so
//     "whenever a creature dies" never sees it; nothing is sacrificed,
//     so "whenever you sacrifice" never sees it either. The permanent
//     leaves the battlefield through the one exit primitive, so
//     "whenever a permanent leaves the battlefield" always does.
//   - **Not ReturnToHandCost.** The destination is exile, and the
//     permanents it moved are recorded on PaidCost.Exiled, so "the
//     exiled creature" (Food Chain's mana value) can be answered.
//
// THE CLAUSE IS DATA, not a TargetSpec: a card type, read off the
// permanent's current characteristics (Card.HasCardType). Every printed
// clause the component covers is "a <card type> you control", and a
// TargetSpec would have brought its predicate closures into Game's
// reach — ADR 0041's closure ratchet (closure_fields_test.go) admits no
// new func-typed route, and asks that a new component be data. A clause
// a card type cannot say ("another nontoken artifact creature", Curie,
// Emergent Intelligence) grows a field here when its card is built.
//
// And what it shares with every cost component that moves cards: the
// move does not TARGET (CR 601.2h / 602.2b), so hexproof, shroud and
// protection never apply; it goes through routeCardToZoneLocked with
// MustSettleNow, because a cost is one indivisible step; and a
// commander named to it is asked CR 903.9 BEFORE anything is paid
// (#1397, cost_commander_choice.go), the answer riding the move.

// ExilePermanentsCost is the "exile N <permanents> you control"
// component of an ability's cost. Two owners, like ExileCost,
// SacrificeOther and TapOthers: AbilityCost.ExilePermanents (a CR 602
// ability — The Soul Stone) and ManaAbilityShape.ExilePermanents (a
// CR 605 mana ability — Food Chain).
//
// The zero value, and nil, demand nothing — so the activation paths can
// ask without a guard.
type ExilePermanentsCost struct {
	// Count is how many permanents must be exiled, or the floor with
	// OrMore. A count bounded by a TOTAL ("one or more other artifacts
	// you control with total mana value X", Fabrication Foundry) is an
	// announcement this component does not model.
	Count int

	// CardType is the card type the permanent must have — "creature"
	// for "Exile a creature you control" — compared case-insensitively
	// against its CURRENT types, so an animated land is a creature and
	// a creature an effect turned into a noncreature artifact is not.
	// Empty is "a permanent you control". effects.Register refuses a
	// value that is not a permanent card type.
	CardType string

	// ExcludeSource is the printed word "another". No catalog card
	// prints it on this clause yet; when it is clear, the source pays
	// if it has the card type.
	ExcludeSource bool

	// FromGraveyard is craft's second zone (CR 702.167b, ADR 0137): a
	// material named WITHOUT the word "card" ("Craft with artifact")
	// may be a permanent the activator controls OR a card with that
	// quality in the activator's own graveyard, and one payment may mix
	// the two ("Exile the two from among creatures you control and/or
	// creature cards in your graveyard", Visage of Dread). A graveyard
	// card is judged on its characteristics there, which for a
	// double-faced card are its front face's (CR 712.8a).
	//
	// Ability-only: effects.Register refuses it on a mana ability, and
	// beside an ExileCards component, which could then name the same
	// graveyard card twice.
	FromGraveyard bool

	// Subtype is a subtype the material must have — "Craft with Island"
	// (Waterlogged Hulk), "Craft with Cave" (Kaslem's Stonetree): a
	// craft quality may be a subtype (CR 702.167b). Read through
	// Card.HasSubtype, so a changeling is every creature type. Empty
	// names no subtype; set beside CardType it narrows that type.
	Subtype string

	// OrMore makes Count a floor rather than the count (ADR 0137's
	// 2026-10-10 amendment): "Craft with one or more creatures", "four
	// or more red instant and/or sorcery cards". The activator names as
	// many materials as they like from Count up, and the list's length
	// is the announcement (CR 602.2b) — the shape of an open sacrifice
	// count (SacrificeCostBounds, #1213).
	OrMore bool

	// GraveyardOnly narrows FromGraveyard to its second zone alone: a
	// material named WITH the word "card" ("four or more red instant
	// and/or sorcery cards", Ore-Rich Stalactite) is a card in the
	// activator's graveyard and never a permanent (CR 702.167b).
	// effects.Register refuses it without FromGraveyard.
	GraveyardOnly bool

	// CardTypes is "<type> and/or <type>": a material has at least one
	// of them. Beside CardType, never instead of it (Register refuses
	// both). A type no permanent has ("instant", "sorcery") is accepted
	// only with GraveyardOnly, since only a card could have it.
	CardTypes []string

	// Color is a colour the material must have, as a one-letter code
	// ("R" for "red instant and/or sorcery cards"), read through
	// Card.HasColor.
	Color string

	// ShareCardType is a rule over the chosen SET: every material
	// shares at least one card type with all the others — Eye of Ojer
	// Taq's "two that share a card type". An artifact creature and an
	// artifact land share artifact.
	ShareCardType bool

	// EachSubtype is the other set rule: one material for each listed
	// subtype, one-to-one, so Count is its length — Throne of the Grim
	// Captain's "a Dinosaur, a Merfolk, a Pirate, and a Vampire". A
	// changeling fills any one entry but never two (the sacrifice set
	// rule's matching, assignSacrificeSet, #2526).
	EachSubtype []string

	// Label is the clause as printed, without the verb — "a creature
	// you control" — shown above the client's picker.
	Label string
}

// ExilePermanentsCostBounds is how many materials one payment may
// name: Count to Count for a fixed clause, Count with no ceiling (hi 0)
// for OrMore. Nil-safe: 0 / 0, which a caller reads as "no component".
func ExilePermanentsCostBounds(ec *ExilePermanentsCost) (lo, hi int) {
	if ec.Empty() {
		return 0, 0
	}
	if ec.OrMore {
		return ec.Count, 0
	}
	return ec.Count, ec.Count
}

// exileCountLegal reports whether naming `named` materials is a legal
// count for `ec`.
func exileCountLegal(ec *ExilePermanentsCost, named int) bool {
	lo, hi := ExilePermanentsCostBounds(ec)
	return named >= lo && (hi == 0 || named <= hi)
}

// Empty reports whether the cost demands nothing. Nil-safe.
func (c *ExilePermanentsCost) Empty() bool {
	return c == nil || c.Count < 1
}

// matches reports whether `card` has the clause's card type, colour
// and subtype. The set rules (ShareCardType, EachSubtype) are about the
// payment as a whole and are judged by exileSetSatisfied; a candidate
// for an EachSubtype clause must still have one of its subtypes.
func (c *ExilePermanentsCost) matches(card *Card) bool {
	if c.CardType != "" && !card.HasCardType(strings.ToLower(c.CardType)) {
		return false
	}
	if len(c.CardTypes) > 0 && !slices.ContainsFunc(c.CardTypes, func(t string) bool {
		return card.HasCardType(strings.ToLower(t))
	}) {
		return false
	}
	if c.Color != "" && !card.HasColor(c.Color) {
		return false
	}
	if len(c.EachSubtype) > 0 && !slices.ContainsFunc(c.EachSubtype, card.HasSubtype) {
		return false
	}
	return c.Subtype == "" || card.HasSubtype(c.Subtype)
}

// cardTypesOf is the card types `card` has, lowercased, in CR 205.2a's
// order — what ShareCardType compares.
func cardTypesOf(card *Card) []string {
	var out []string
	for _, t := range ChoosableCardTypes {
		if lt := strings.ToLower(t); card.HasCardType(lt) {
			out = append(out, lt)
		}
	}
	return out
}

// SharedCardTypes is the card types every card in `cards` has,
// lowercased, in CR 205.2a's order: what "two that share a card type"
// (Eye of Ojer Taq) checks, and what Apex Observatory's "a card type
// shared among two exiled cards used to craft it" offers. Nil for no
// cards.
func SharedCardTypes(cards []Card) []string {
	if len(cards) == 0 {
		return nil
	}
	shared := cardTypesOf(&cards[0])
	for i := 1; i < len(cards) && len(shared) > 0; i++ {
		var kept []string
		for _, t := range shared {
			if cards[i].HasCardType(t) {
				kept = append(kept, t)
			}
		}
		shared = kept
	}
	return shared
}

// exileSetSatisfied judges the clause's set rule over the materials of
// one payment, as the objects they are now. A clause with no set rule
// is satisfied by anything (each material's own check already ran); the
// count is the caller's question. Nil `cards` (a pick that is no
// material) never satisfies a rule.
func (c *ExilePermanentsCost) exileSetSatisfied(cards []*Card) bool {
	if !c.ShareCardType && len(c.EachSubtype) == 0 {
		return true
	}
	if cards == nil {
		return false
	}
	if c.ShareCardType {
		vals := make([]Card, len(cards))
		for i, cd := range cards {
			vals[i] = *cd
		}
		if len(SharedCardTypes(vals)) == 0 {
			return false
		}
	}
	if len(c.EachSubtype) > 0 {
		if len(cards) != len(c.EachSubtype) {
			return false
		}
		order := make([]int, len(cards))
		for i := range order {
			order[i] = i
		}
		if assignSacrificeSet(c.subtypeFits(cards), order) == nil {
			return false
		}
	}
	return true
}

// subtypeFits is the EachSubtype matching table: row i, column j says
// whether candidate j has the i-th subtype.
func (c *ExilePermanentsCost) subtypeFits(cards []*Card) [][]bool {
	fits := make([][]bool, len(c.EachSubtype))
	for i, st := range c.EachSubtype {
		fits[i] = make([]bool, len(cards))
		for j, cd := range cards {
			fits[i][j] = cd.HasSubtype(st)
		}
	}
	return fits
}

// exileMaterialLocked is the object `id` names as a material of `ec`
// for `playerID`: a card in their own graveyard when the clause reaches
// there, else a permanent on the battlefield (never, for a
// graveyard-only clause). Nil when it is neither. It does not judge the
// clause.
//
// Caller must hold g.mu.
func (g *Game) exileMaterialLocked(playerID uuid.UUID, ec *ExilePermanentsCost, id uuid.UUID) *Card {
	if gc := g.graveyardMaterialLocked(playerID, ec, id); gc != nil {
		return gc
	}
	if ec.GraveyardOnly {
		return nil
	}
	return findBattlefieldCard(g, id)
}

// exileMaterialsLocked resolves `ids` to the objects they name as
// materials of `ec`, or nil when any of them is not one.
//
// Caller must hold g.mu.
func (g *Game) exileMaterialsLocked(playerID uuid.UUID, ec *ExilePermanentsCost, ids []uuid.UUID) []*Card {
	out := make([]*Card, 0, len(ids))
	for _, id := range ids {
		c := g.exileMaterialLocked(playerID, ec, id)
		if c == nil {
			return nil
		}
		out = append(out, c)
	}
	return out
}

// graveyardMaterialLocked returns the card `id` names in `playerID`'s
// own graveyard when the clause reaches there (FromGraveyard, ADR
// 0137), or nil. "Your graveyard" is the activator's: a card in another
// player's graveyard never pays, whoever controlled it last.
//
// Caller must hold g.mu.
func (g *Game) graveyardMaterialLocked(playerID uuid.UUID, ec *ExilePermanentsCost, id uuid.UUID) *Card {
	if ec == nil || !ec.FromGraveyard {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	for i := range p.Graveyard.Cards {
		if p.Graveyard.Cards[i].InstanceID == id {
			return &p.Graveyard.Cards[i]
		}
	}
	return nil
}

// ExilePermanentsOptionsForEffect is the set of permanents that could
// pay `ec` right now for an ability on `sourceID` activated by
// `playerID`, in battlefield order — the ONE walk the protocol view
// stamps, the legal enumerator pays out of and the validator accepts
// from (#544). A hexproof permanent is on it: paying a cost does not
// target. Whether the options add up to Count is the payment's
// question; this never enumerates subsets.
//
// Caller must hold g.mu (read or write).
func (g *Game) ExilePermanentsOptionsForEffect(playerID, sourceID uuid.UUID, ec *ExilePermanentsCost) []uuid.UUID {
	if ec.Empty() {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if ec.GraveyardOnly || c.Controller != playerID || (ec.ExcludeSource && c.InstanceID == sourceID) || !ec.matches(c) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	// ADR 0137: craft's graveyard half, after the permanents, in pile
	// order.
	if ec.FromGraveyard {
		if p := g.playerByIDLocked(playerID); p != nil && p.Graveyard != nil {
			for i := range p.Graveyard.Cards {
				c := &p.Graveyard.Cards[i]
				if c.InstanceID == sourceID || !ec.matches(c) {
					continue
				}
				out = append(out, c.InstanceID)
			}
		}
	}
	return out
}

// ExilePermanentsPayable reports whether `ec` could be paid at all
// right now (CR 118.3).
//
// Caller must hold g.mu (read or write).
func (g *Game) ExilePermanentsPayable(playerID, sourceID uuid.UUID, ec *ExilePermanentsCost) bool {
	if ec.Empty() {
		return true
	}
	return g.ExilePermanentsPaymentForEffect(playerID, ec, g.ExilePermanentsOptionsForEffect(playerID, sourceID, ec)) != nil
}

// ExilePermanentsPaymentForEffect is ONE legal payment of `ec` out of
// `candidates` (already in payment order) with the fewest materials the
// clause allows, or nil when the candidates cannot pay it (CR 118.3).
// With no set rule it is the first Count. A ShareCardType clause takes
// the first Count of the earliest card type (CR 205.2a order) that
// enough candidates have; an EachSubtype clause is the one-to-one
// matching, trying a candidate that fits fewer entries first, so a
// changeling is kept for the entry nothing else fills.
//
// Caller must hold g.mu (read or write).
func (g *Game) ExilePermanentsPaymentForEffect(playerID uuid.UUID, ec *ExilePermanentsCost, candidates []uuid.UUID) []uuid.UUID {
	if ec.Empty() || len(candidates) < ec.Count {
		return nil
	}
	if ec.ShareCardType {
		if groups := g.ExileSharedTypeGroupsForEffect(playerID, ec, candidates); len(groups) > 0 {
			return append([]uuid.UUID(nil), groups[0].Candidates[:ec.Count]...)
		}
		return nil
	}
	if len(ec.EachSubtype) > 0 {
		cards := g.exileMaterialsLocked(playerID, ec, candidates)
		if cards == nil {
			return nil
		}
		fits := ec.subtypeFits(cards)
		versatility := make([]int, len(cards))
		for j := range cards {
			for i := range fits {
				if fits[i][j] {
					versatility[j]++
				}
			}
		}
		order := make([]int, len(cards))
		for j := range order {
			order[j] = j
		}
		sort.SliceStable(order, func(a, b int) bool { return versatility[order[a]] < versatility[order[b]] })
		cols := assignSacrificeSet(fits, order)
		if cols == nil {
			return nil
		}
		out := make([]uuid.UUID, len(cols))
		for i, j := range cols {
			out[i] = candidates[j]
		}
		return out
	}
	return append([]uuid.UUID(nil), candidates[:ec.Count]...)
}

// ExileSharedTypeGroup is one card type and the candidates that have
// it, in the order given.
type ExileSharedTypeGroup struct {
	CardType   string
	Candidates []uuid.UUID
}

// ExileSharedTypeGroupsForEffect splits `candidates` by card type for a
// ShareCardType clause: one group per card type at least Count of them
// have, in CR 205.2a's order. Count materials from one group are a
// legal payment, and a payment is legal only if it lies inside some
// group. Nil for a clause without the rule.
//
// Caller must hold g.mu (read or write).
func (g *Game) ExileSharedTypeGroupsForEffect(playerID uuid.UUID, ec *ExilePermanentsCost, candidates []uuid.UUID) []ExileSharedTypeGroup {
	if ec.Empty() || !ec.ShareCardType {
		return nil
	}
	cards := g.exileMaterialsLocked(playerID, ec, candidates)
	if cards == nil {
		return nil
	}
	var out []ExileSharedTypeGroup
	for _, t := range ChoosableCardTypes {
		lt := strings.ToLower(t)
		grp := ExileSharedTypeGroup{CardType: lt}
		for j, c := range cards {
			if c.HasCardType(lt) {
				grp.Candidates = append(grp.Candidates, candidates[j])
			}
		}
		if len(grp.Candidates) >= ec.Count {
			out = append(out, grp)
		}
	}
	return out
}

// ExileSubtypeGroupsForEffect splits `candidates` by the EachSubtype
// entry each could fill, for the protocol view's each_of: a candidate
// that fits two entries (a changeling) appears under both. Nil without
// the rule.
//
// Caller must hold g.mu (read or write).
func (g *Game) ExileSubtypeGroupsForEffect(playerID uuid.UUID, ec *ExilePermanentsCost, candidates []uuid.UUID) []SacrificeSetGroup {
	if ec.Empty() || len(ec.EachSubtype) == 0 {
		return nil
	}
	cards := g.exileMaterialsLocked(playerID, ec, candidates)
	if cards == nil {
		return nil
	}
	fits := ec.subtypeFits(cards)
	out := make([]SacrificeSetGroup, len(ec.EachSubtype))
	for i, st := range ec.EachSubtype {
		out[i].Label = subtypeArticle(st) + " " + st
		for j, id := range candidates {
			if fits[i][j] {
				out[i].Candidates = append(out[i].Candidates, id)
			}
		}
	}
	return out
}

// subtypeArticle is "a" or "an" before a subtype's name.
func subtypeArticle(word string) string {
	if word != "" && strings.ContainsRune("AEIOUaeiou", rune(word[0])) {
		return "an"
	}
	return "a"
}

// validateExilePermanentsCostLocked checks an ExilePermanents component
// without moving anything — ADR 0020 §3's "validate everything, then
// pay everything", so a refused activation never leaves a board half
// exiled. In the order the errors matter:
//
//   - IDs sent for an ability with no such component are refused
//     rather than ignored, as an unexpected sacrifice_ids or return_ids
//     is (ErrInvalidParam).
//   - Exactly Count permanents, each named once (ErrInvalidParam).
//   - None of them ALSO moved by another component of the same payment
//     (`alsoSpent`: the sacrifices, the returns, the source when the
//     cost exiles or sacrifices it). One permanent pays one component
//     (CR 118.3) (ErrInvalidParam).
//   - Never the source when the clause says "another" (ErrInvalidParam).
//   - On the battlefield (ErrCardNotFound), controlled by the activator
//     (ErrCardCallerMismatch), of the clause's card type
//     (ErrIllegalTarget, the code the sacrifice and return clauses use
//     for a permanent the clause does not admit).
//
// A TAPPED permanent is a legal pick: exiling is not tapping.
//
// Caller must hold g.mu.
func (g *Game) validateExilePermanentsCostLocked(playerID, sourceID uuid.UUID, ec *ExilePermanentsCost, ids, alsoSpent []uuid.UUID) error {
	if ec.Empty() {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if !exileCountLegal(ec, len(ids)) {
		return ErrInvalidParam
	}
	spent := make(map[uuid.UUID]bool, len(alsoSpent))
	for _, id := range alsoSpent {
		spent[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] || spent[id] || (ec.ExcludeSource && id == sourceID) {
			return ErrInvalidParam
		}
		seen[id] = true
		// ADR 0137: a craft material may be a card in the activator's
		// own graveyard. It is theirs by being there, so the clause is
		// the only question.
		if gc := g.graveyardMaterialLocked(playerID, ec, id); gc != nil {
			if gc.InstanceID == sourceID || !ec.matches(gc) {
				return ErrIllegalTarget
			}
			continue
		}
		// A graveyard-only clause ("four or more … cards") is never
		// paid with a permanent.
		if ec.GraveyardOnly {
			return ErrIllegalTarget
		}
		c := findBattlefieldCard(g, id)
		if c == nil {
			return ErrCardNotFound
		}
		// "You control" is the cost's own clause (CR 118.3: you can
		// only pay with what is yours to pay with).
		if c.Controller != playerID {
			return ErrCardCallerMismatch
		}
		if !ec.matches(c) {
			return ErrIllegalTarget
		}
	}
	// ADR 0137's amendment: the clause's rule over the whole set ("two
	// that share a card type", "a Dinosaur, a Merfolk, a Pirate, and a
	// Vampire"), judged once every pick is known to be a material.
	if !ec.exileSetSatisfied(g.exileMaterialsLocked(playerID, ec, ids)) {
		return ErrIllegalTarget
	}
	return nil
}

// movedSourceAlso is the `alsoSpent` list validateExilePermanentsCostLocked
// checks a payment against: every permanent another component of the
// same payment moves (`lists` — the sacrifices, which already hold the
// source when the cost sacrifices it, and the returns), plus the source
// when the cost exiles it.
func movedSourceAlso(sourceID uuid.UUID, exileSelf bool, lists ...[]uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	if exileSelf {
		out = append(out, sourceID)
	}
	for _, l := range lists {
		out = append(out, l...)
	}
	return out
}

// payExilePermanentsCostLocked exiles the validated permanents, one
// route each, so a leaves-the-battlefield watcher sees every one of
// them. Through the ONE exit door (routeCardToZoneLocked) with cause
// cost and MustSettleNow: a cost is one indivisible step (CR 601.2h /
// 602.2b, and CR 605.3b for a mana ability), so the move may not stop
// on a prompt. A commander named here was asked CR 903.9 before the
// payment began (askCostCommanderLocked, #1397); `answers` carries what
// its owner said, and a "yes" puts it in the command zone instead —
// the cost is still paid.
//
// Returns the instance IDs it moved, for PaidCost.Exiled. The battlefield
// exit has already written each permanent's last-known information
// (rememberDepartingPermanentLocked), so "the exiled creature's mana
// value" is read as it last existed there (LastKnownPermanentForEffect).
//
// Call only after validateExilePermanentsCostLocked has passed, and
// BEFORE a CR 602 ability's stack item is built, so the leaves-triggers
// this queues are drained by the closing state-check pass and sit ABOVE
// the ability (CR 603.3b). On a mana ability they are drained on the
// way out, with the mana already in the pool.
//
// A permanent that vanished between validation and payment is skipped,
// as payReturnToHandCostLocked skips one.
//
// Caller must hold g.mu.
func (g *Game) payExilePermanentsCostLocked(playerID, sourceID uuid.UUID, ids []uuid.UUID, answers map[uuid.UUID]bool) ([]uuid.UUID, error) {
	var moved []uuid.UUID
	for _, id := range ids {
		// ADR 0137: a craft material in the activator's graveyard goes
		// through the same door, as a card leaving a graveyard rather
		// than a permanent leaving the battlefield.
		if findBattlefieldCard(g, id) == nil {
			if z := g.findCardZoneLocked(id); z == nil || z.Kind != ZoneGraveyard || z.Owner != playerID {
				continue
			}
		}
		if _, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:        id,
			Dst:           ZoneExile,
			Actor:         playerID,
			Source:        sourceID,
			Cause:         MoveCause{Kind: MoveCauseCost, Controller: playerID},
			MustSettleNow: true,
			// #1397: the owner's CR 903.9 answer, asked before the
			// payment by askCostCommanderLocked.
			commanderAnswer: commanderAnswerFor(answers, id),
		}); err != nil {
			return moved, err
		}
		moved = append(moved, id)
	}
	return moved, nil
}

// PermanentCardTypes are the card types a permanent can have (CR 110.4)
// — the values ExilePermanentsCost.CardType may name, lowercase. Read by
// effects.Register's boot check.
var PermanentCardTypes = []string{"artifact", "battle", "creature", "enchantment", "land", "planeswalker"}
