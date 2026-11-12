package game

// life_cost_bar.go — "Players can't pay life to cast spells or to
// activate abilities that aren't mana abilities." (Karn's Sylex, #5660).
//
// A world rule its source carries on itself as the restriction bit
// PlayersCantPayLifeToCastOrActivate (restrictions.go), the way
// PermanentsCantPhaseIn is carried: derived from the battlefield on
// every question, stored nowhere, so two sources compose and one of
// them leaving revokes nothing the other still says.
//
// WHAT IT STOPS. CR 101.2: an effect that says something can't happen
// beats the rule or effect that says it can, and CR 118.3 has a player
// unable to pay a cost they lack the means for, so every life component
// of a spell's total cost (CR 601.2f: its mana cost or alternative
// cost, plus additional costs and cost increases) and of a non-mana
// activation's (CR 602.2b) is unpayable while the rule is live:
//
//   - a printed "Pay N life" on an activated ability, and "Pay X life"
//     above X = 0;
//   - an additional cost's "pay N life" or "pay X life", an alternative
//     cost's life, and a targeting tax paid in life;
//   - the life half of a Phyrexian mana symbol (CR 107.4f) and of a
//     life-for-mana grant (K'rrik), in a spell's or such an ability's
//     cost.
//
// WHAT IT LEAVES ALONE, because the printed text names two things only:
// a mana ability's cost (CR 605.1a: Mana Confluence, a filter land paid
// under K'rrik), and every payment that is neither a cast nor an
// activation: a ward cost, a "pay 2 life unless" at resolution, a land
// entering ("you may pay 2 life"), an attack tax, a special action.
// Paying 0 life is no payment (CR 119.4 lets any player pay 0), so "pay
// X life" at X = 0 stays legal.
//
// ONE READER, asked beside CanPayLifeLocked at each announce gate, by
// the legal-move enumerator and by the view, so a payment the engine
// refuses is never offered.

// cantPayLifeToCastOrActivateLocked reports whether a permanent on the
// battlefield carries the PlayersCantPayLifeToCastOrActivate world
// rule, read off the layered view like every restriction gate. A
// phased-out source is not in the battlefield slice and stops nothing
// (CR 702.26b).
//
// Caller must hold g.mu (read or write).
func (g *Game) cantPayLifeToCastOrActivateLocked() bool {
	if g == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if Restricted(&g.Battlefield.Cards[i], PlayersCantPayLifeToCastOrActivate) {
			return true
		}
	}
	return false
}

// CantPayLifeToCastOrActivateForEffect is the exported reader, for the
// legal-move enumerator and the view.
//
// Caller must hold g.mu (read or write).
func (g *Game) CantPayLifeToCastOrActivateForEffect() bool {
	return g.cantPayLifeToCastOrActivateLocked()
}

// CanPayLifeToCastOrActivateLocked is CanPayLifeLocked for a life
// component of a spell's cost or of a non-mana activation's: CR 119.4
// and CR 119.8, and not at all while "players can't pay life to cast
// spells or to activate abilities that aren't mana abilities" is live.
// A non-positive amount is payable by anybody, as it is there.
//
// Caller must hold g.mu (read or write).
func (g *Game) CanPayLifeToCastOrActivateLocked(p *Player, amount int) bool {
	if amount <= 0 {
		return true
	}
	if g.cantPayLifeToCastOrActivateLocked() {
		return false
	}
	return g.CanPayLifeLocked(p, amount)
}
