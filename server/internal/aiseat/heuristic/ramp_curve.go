package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// ramp_curve.go — #5084: a row that gives up its own permanent for
// lands (Burnished Hart's "{3}, Sacrifice Burnished Hart: Search your
// library for up to two basic land cards", Sakura-Tribe Elder's) is
// worth its lands by the turn as well as the hand.
//
// ADR 0126 §6 prices a declared purpose's lands at ManaSource each plus
// the ramp premium (rampFor), and the premium is the hand's deficit: the
// largest mana value the bot holds above the mana its sources make. In
// #4991's thopter-battle arena the Hart's activation was taken 8 and 9
// times in 16 games. Its two lands were about 2.0 and its 2/2 body about
// 3.3, with a flat 0.30 for the untapped creature on top, so the bot
// cracked it only while it held a card it could not yet cast. A deck of
// zero-drops and three-drops rarely does, and early in the game two
// lands are worth far more than a vanilla 2/2: they are mana on every
// later turn, for the cards the bot has not drawn yet.
//
// With Config.PriceRampCurve on, the premium for such a row's lands is
// the larger of the hand's deficit and the curve's: the mana a deck on
// curve wants RampCurveLead turns from now (the round plus the lead,
// capped at RampWantCap) above the mana its sources make. The curve
// counts only while the round is below RampWantCap; from then on more
// mana is worth what the hand asks of it, as before. So on round 4 with
// four lands the Hart's two lands are two mana of ramp (it wants six),
// and the bot cracks it; on round 10 with eight lands the curve has
// closed and the bot keeps its blocker.
//
// The row's untapped creature was also charged the flat tappedBlocker
// (0.30) that a non-tap row of a creature pays. A row that sacrifices its
// source has already paid for the whole creature (selfSacrificeCost, its
// board value), blocker and all, so under the switch that charge is not
// made again for these rows.
//
// The switch ships OFF. In #5084's arena (thopter-battle, all four seats
// on one policy for 16 games, then two seats with the curve against two
// without for 48) it changed few decisions: the Hart is mostly on the
// battlefield from round 7 on, when the curve has closed, and early its
// {3} competes with casts the turn plan prices higher. The curve seats
// won 21 of 96 seat-games against 27 of 96. `heuristic-rampcurve` seats
// it for the next measurement (a ramp deck, or the curated decks).
//
// Scope: only an activated row of the bot's own that sacrifices its
// source and whose declared purpose puts lands onto the battlefield. A
// fetch land keeps its own price (landfall.go), and the casts of ramp
// spells and mana rocks keep ADR 0126 §2's hand deficit.

// rampCurveRow reports whether an activated row is one the curve
// prices: the row sacrifices its source and declares lands. A fetch is
// priced in landfall.go and has had its purpose cleared by then.
func rampCurveRow(row *protocol.ActivatedAbilityView, ps purposeSet) bool {
	return row != nil && row.SacrificeSelf && ps.lands > 0
}

// curveDeficit is how much mana a deck on curve wants beyond the bot's
// sources: the round plus RampCurveLead, capped at RampWantCap, less
// the mana its sources make. Zero from round RampWantCap on, and when
// the Config does not price the curve.
func (p *Policy) curveDeficit(st *state) int {
	if !p.cfg.PriceRampCurve || st.view == nil {
		return 0
	}
	round := st.view.Turn.Number
	if round <= 0 || round >= p.cfg.RampWantCap {
		return 0
	}
	want := min(p.cfg.RampWantCap, round+p.cfg.RampCurveLead)
	return max(0, want-st.manaSources())
}

// rampCurveValue is what the curve adds to `lands` lands' ramp premium
// over the hand's deficit (rampFor): RampPerMana for each land up to the
// curve's deficit, less what rampFor already paid. Zero when the hand
// asks for as much or more, and with RampPerMana zero, so the turn
// plan's premium-free pricing (plan.go) stays premium-free.
func (p *Policy) rampCurveValue(st *state, lands int) float64 {
	if !p.cfg.PricePurposes || p.cfg.RampPerMana == 0 || lands <= 0 || st.seat == nil {
		return 0
	}
	curve := p.cfg.RampPerMana * float64(min(lands, p.curveDeficit(st)))
	return max(0, curve-p.rampFor(st, nil, lands))
}
