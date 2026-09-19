package game

import "github.com/google/uuid"

// cast_cost.go — which cost a cast is actually paying, and the one
// thing that settles on its own: CR 107.3b's value of X.
//
// A cast pays exactly one mana cost, and the engine already chooses
// it in one place (printedCostLocked): the alternative cost claimed
// at announce (CR 118.9), an exile grant's own price (cascade's
// "{0}", airbend's "{2}"), or the cost printed in the corner. Which
// one it is decides more than the number of mana owed, so the choice
// is a value the announce path can ask questions of rather than a
// string it parses twice.

// CastCost is the pair the announce path needs: what the card
// prints, and what this cast actually owes before the commander tax,
// the cost modifiers and convoke touch it.
//
// Both halves are cost STRINGS in Scryfall brace notation, because
// the question CR 107.3b asks is about the shape of the cost
// ("does it have an {X} slot") rather than about its total.
type CastCost struct {
	// Printed is the card's printed mana cost, for the face being
	// cast. Empty for a card with no mana cost (CR 202.1b).
	Printed string

	// Paid is the cost string this cast pays instead: the claimed
	// alternative cost, the exile grant's override, or Printed when
	// neither applies.
	Paid string
}

// CastCostFor picks the cost this cast pays, in the same order
// printedCostLocked applies it: the resolved alternative cost
// replaces the printed cost (CR 118.9), and a granted permission's
// own flat price replaces whatever was chosen, because the permission
// is the reason the cast is happening at all.
//
// Pure. `alt` is the offer the announce path already resolved — the
// card's own, or the one a permission synthesised (ADR 0066), which
// is why this takes the resolved offer rather than a key: by the time
// a Snapcaster'd Brainstorm is priced, the catalog has nothing to say
// about "flashback" on that card.
//
// `grant` is nil for an ordinary cast. A permission that carries an
// AltCostKey has already become `alt`, so only a FLAT price (airbend's
// "{2}", cascade's "{0}") is read here.
func CastCostFor(card Card, alt *AlternativeCost, grant *CastPermission) CastCost {
	paid := card.ManaCost
	if alt != nil {
		paid = alt.ManaCost
	}
	if grant != nil && grant.Cost != "" {
		paid = grant.Cost
	}
	return CastCost{Printed: card.ManaCost, Paid: paid}
}

// LocksXAtZero reports whether CR 107.3b fixes X at 0 for this cast:
//
//	"If a player is casting a spell that has an {X} in its mana cost,
//	 the value of X isn't defined by the text of that spell, and an
//	 effect lets that player cast that spell while paying neither its
//	 mana cost nor an alternative cost that includes X, then the only
//	 legal choice for X is 0. This doesn't apply to effects that only
//	 reduce a cost, even if they reduce it to zero."
//
// Two reads of one pair of strings, and every free-cast path falls
// out of them:
//
//   - The printed cost has an {X} slot, so there is an X to announce
//     at all (CR 601.2b). A spell whose X lives somewhere OTHER than
//     the mana cost is untouched — Toxic Deluge's "pay X life"
//     additional cost and Waterbender's Restoration's waterbend {X}
//     are costs of their own, not the mana cost, so 107.3b has
//     nothing to say about them and their X is still announced on a
//     free cast.
//   - The cost being paid has no {X} slot. Cascade's "{0}" grant, a
//     Siege's free cast, "you may cast it without paying its mana
//     cost" as an AlternativeCost with an empty ManaCost, and an
//     alternative cost priced without X ("you may pay {R} rather
//     than pay this spell's mana cost") are all one answer here. An
//     alternative cost that DOES carry X ("{X}{R}, discard a card")
//     still asks, because the rule's exception is written about the
//     cost, not about the keyword granting it.
//
// Cost REDUCTIONS are outside this by construction: a modifier
// subtracts from the cost that was chosen here and never replaces
// it, so the {X} slot survives Ghalta's discount all the way down to
// {0} and the caster still announces X.
//
// CR 107.3c — a spell whose own text defines X — would be a third
// clause here, and there is deliberately none: no card in the
// catalog defines the X in its mana cost from its text, and the
// engine has no seam that could express it. If one arrives it
// attaches to the first read (the printed cost's X stops being an
// announcement), not to the second.
//
// A printed cost the parser cannot read (split and adventure cards
// import a joined "{1}{R} // {1}{U}") locks nothing. The engine
// refuses such a cast outright when it pays the printed cost, and
// when a grant prices it there is no {X} the parser can see to fix.
func (c CastCost) LocksXAtZero() bool {
	printed, err := ParseCost(c.Printed)
	if err != nil || printed.XSlots == 0 {
		return false
	}
	paid, err := ParseCost(c.Paid)
	if err != nil {
		return false
	}
	return paid.XSlots == 0
}

// CastPrice is the whole of what one announced cast owes, and the
// ONE answer the engine gives to "what does this cost?" (#696).
//
// It exists because three consumers have to agree with CastSpell to
// the symbol — the auto-tap preview endpoint, the auto-tapper itself
// and the bot enumerator — and before this they re-derived pieces of
// the price from the card. A mirrored copy of the commander tax and
// the cost modifiers knows nothing about the alternative cost claimed
// at announce, a granted permission's flat override, the "spend mana
// as though any colour" fold or the mana half of an optional
// additional cost, so every non-hand and alternative-cost cast
// disagreed with the engine in one direction or the other: a preview
// that disabled a cast the engine would have allowed, or planned taps
// for one it would refuse.
type CastPrice struct {
	// CastCost is the pair of cost STRINGS CR 118.9 chose between:
	// what the card prints, and what this cast pays instead (the
	// claimed alternative cost, or a granted permission's own flat
	// price). Embedded, so LocksXAtZero rides along — CR 107.3b is a
	// question about exactly this pair.
	CastCost

	// Card is the card AS THE CAST ANNOUNCES IT: the chosen face
	// already materialised (ADR 0034), which is the value every
	// announce gate in CastSpell reads and the one every number below
	// was computed from. A caller that goes on to ask the card a
	// question — ManaSpendForCast's spend context, say — must ask
	// this copy rather than the one it passed in, or it will read the
	// front face of a cast that announced the back.
	Card Card

	// Base is the total at CR 601.2f BEFORE the board's cost
	// modifiers and before convoke or waterbend spends anything
	// against it: the chosen cost string, plus the commander tax (CR
	// 903.8), plus the mana half of the announced optional additional
	// costs, with a grant's "spend mana as though any colour" fold
	// applied.
	//
	// The number the announce-time tap budget is measured against —
	// and the one the bot enumerator prices from, because it re-runs
	// the cost modifiers once per candidate target set and needs the
	// cost they apply TO. Reading Total there instead would fold the
	// {X} slot of a convoke spell away (tapPermanentsAdjusted settles
	// X into generic) and a Chord of Calling would only ever be
	// enumerated at X=0.
	Base ParsedCost

	// Total is what the payment charges: Base with the cost modifiers
	// applied and the permanents named in CastSpellParams.TapIDs
	// subtracted. Identical to what applyCastCostLocked will demand
	// of the mana pool, which is the entire contract.
	Total ParsedCost
}

// PriceCast prices an announced cast exactly as CastSpell will charge
// it — the alternative cost claimed at announce, a granted
// permission's override, the commander tax, the optional additional
// costs' mana, the board's cost modifiers and the convoke/waterbend
// subtraction — and is the only supported way to ask that question
// from outside the package.
//
// `params` is the announcement: the same CastSpellParams the caller
// would send to CastSpell. FromZone, AlternativeCost, OptionalCosts,
// Face, XValue, Targets and TapIDs all change the answer; everything
// else is ignored. An empty FromZone reads as the hand, exactly as
// the cast path reads it.
//
// The Phyrexian strike (CR 107.4f) is deliberately NOT applied here:
// it is a claim about how the announced cost is paid rather than part
// of the cost, applyCastCostLocked strikes it after this number is
// settled, and PhyrexianLifePlan is the shared helper that does it.
//
// Takes the read lock. Callers already under g.mu want
// PriceCastForEffect.
func (g *Game) PriceCast(playerID uuid.UUID, card Card, params CastSpellParams) (CastPrice, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.priceCastLocked(playerID, card, params)
}

// PriceCastForEffect is PriceCast for a caller that already holds
// g.mu — the bot enumerator, which runs inside ReadSnapshot, and the
// view layer. Read-only, same contract as ApplyCostModifiersForEffect.
func (g *Game) PriceCastForEffect(playerID uuid.UUID, card Card, params CastSpellParams) (CastPrice, error) {
	return g.priceCastLocked(playerID, card, params)
}

// priceCastLocked is the body of both. One walk of printedCostLocked
// and one of its modifier-and-tap tail, so Base and Total can never
// describe two different casts. Caller must hold g.mu.
func (g *Game) priceCastLocked(playerID uuid.UUID, card Card, params CastSpellParams) (CastPrice, error) {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return CastPrice{}, ErrPlayerNotFound
	}
	// ADR 0034, and the same order CastSpell settles it in: the
	// permission is read off the UNFACED card (a grant belongs to the
	// instance, not to a face), and the face it names — or the one
	// the caller asked for, narrowed by the grant's Faces (CR 715.4,
	// ADR 0066 amendment) — is then materialised onto the copy
	// everything downstream prices. Without this the preview would
	// price the front face of a modal DFC whose back was announced,
	// an adventure's creature half at the adventure's price, and a
	// Siege's granted back-face cast at the front's.
	//
	// No turn is passed: the permission's window was checked by
	// CastPermissionForLocked, which is #945's one liveness test.
	srcKind, _ := castZoneFromWire(params.FromZone)
	grant := g.CastPermissionForLocked(playerID, card, srcKind)
	face, ok := faceForCastLocked(card, params.Face, grant, playerID)
	if !ok {
		return CastPrice{}, ErrInvalidFace
	}
	params.Face = face
	card.SetFace(face)
	base, chosen, err := g.printedCostLocked(p, card, params)
	if err != nil {
		return CastPrice{}, err
	}
	total, err := g.costAfterModifiersLocked(base, p, card, params)
	if err != nil {
		return CastPrice{}, err
	}
	return CastPrice{CastCost: chosen, Card: card, Base: base, Total: total}, nil
}
