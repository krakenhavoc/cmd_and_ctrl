package heuristic

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// fuel.go prices a card the seat would SPEND to a cost — the blue card
// Force of Will pitches, the Island Daze returns, the five cards an Uro
// exiles to escape (#1013, ADR 0033 §1's amendment).
//
// THE GAP IT CLOSES. `Evaluate` prices the battlefield and the seats,
// and a card in a graveyard was worth nothing to it. Two things
// followed, and they were the same missing number:
//
//   - `legal` could only enumerate ONE payment per offer
//     (maxEnumeratedCostPayments was 1), because the payments were
//     indistinguishable to the policy and a wider search would have
//     spent the whole per-source budget ranking them by index. An Uro
//     escaping over a graveyard holding a second Uro, a Snapcaster
//     target and three lands ate whichever three were oldest.
//   - `valueOfCast` priced a non-hand cast as costing no card at all
//     (#673), which was the right answer to the wrong question: what
//     it really spends is the graveyard, and charging a HAND card
//     there was what made every flashback score below passing. The
//     gap was written down in docs/bot.md as a known one.
//
// ONE function answers both, which is the point: the enumerator reads
// it through `legal.Options.OrderCostFuel` to decide which payment to
// offer, and `valueOfCast` reads it to price the payment it was
// offered. A second scorer for the ordering would eventually disagree
// with the first about what an escape costs, and the bot would choose
// a payment it then priced as a mistake.
//
// WHAT A CARD OFF THE BATTLEFIELD IS WORTH, in three cases:
//
//   - IN HAND. A resource (`Weights.Hand`, the number Evaluate already
//     charges per card in hand) plus what the card would do if it
//     resolved. Pitching a second Force of Will costs more than
//     pitching an Island, and the evaluation says so.
//   - IN A GRAVEYARD OR EXILE, and the seat can still CAST it. Escape,
//     flashback, a granted impulse — the view says so on the card
//     itself (`CastableHere`, or an alternative-cost offer stamped for
//     this seat), so the policy needs no oracle text and no catalog
//     lookup. Worth a discounted card-in-hand: it is a real card, but
//     it needs its own cost and its own window, and it may be gone
//     before either arrives.
//   - IN A GRAVEYARD OR EXILE, and it is not a cast surface. Two
//     floors, and the gap between them is the arm that makes an Uro
//     eat the lands rather than the spells: `Config.FuelFloor` for a
//     LAND, which does nothing at all from a graveyard without a
//     Crucible, and `Config.FuelIdle` for anything else, which is the
//     graveyard synergy the policy cannot see — delve, a flashback
//     granted later, a Snapcaster target. Neither is zero: a card
//     nobody can use today is still one tomorrow might want, and a
//     floor of zero would make it the first thing every cost ate.
//
// It does NOT know what the SPELL being cast is, the same limit
// TargetOrder declares: it cannot prefer to pitch a card the spell
// would rather have in the graveyard. Ranking by what the seat LOSES
// is the whole of its power, and losing least is right for every cost.

// CostFuelPricer is implemented by Policy, so the runner threads the
// price into the enumerator. See aiseat.CostFuelPricer.
var _ aiseat.CostFuelPricer = (*Policy)(nil)

// CostFuelPrice implements aiseat.CostFuelPricer. `in` carries the View
// and the Seat; Moves is empty, because this is called in order to
// build them.
//
// One state per decision, captured by the closure, so the enumerator's
// per-candidate call is a map lookup and some arithmetic rather than a
// re-scan of the zones.
func (p *Policy) CostFuelPrice(in aiseat.Input) legal.CostFuelOrder {
	p.mu.Lock()
	st := p.newState(in)
	p.mu.Unlock()
	return func(c legal.TargetCandidate) float64 {
		return p.fuelValue(st, c.ID.String())
	}
}

// fuelValue prices the card with this instance ID as something the seat
// is about to spend: HIGHER means it would rather keep it.
//
// Nil-safe in the way that matters: a card in a zone this seat cannot
// read is `Weights.Unknown`, which is deliberately small and positive.
// An unreadable card is something rather than nothing, and pricing it
// at zero would make it the first thing every cost ate.
func (p *Policy) fuelValue(st *state, id string) float64 {
	if c := st.mine[id]; c != nil {
		// A card in hand, or in the command zone: a resource AND a
		// spell. Force of Will's pitch and Daze's alternative to one.
		return st.w.Hand + p.resolvedValue(st, c, 0)
	}
	if c := st.bf[id]; c != nil {
		// Daze's Island: a permanent, priced exactly as the rest of
		// the evaluation prices it, attachment roles and all.
		return st.permanentValue(c)
	}
	c := st.graveyard[id]
	if c == nil {
		c = st.exile[id]
	}
	if c == nil {
		return st.w.Unknown
	}
	if !castableFromHere(c) {
		// Nothing the seat can do with it TODAY, and the two arms are
		// the difference between an Uro eating the lands and an Uro
		// eating the spells. A land in a graveyard does nothing at all
		// without a Crucible; anything else is the graveyard synergy
		// the policy cannot see — delve, a flashback granted later, a
		// Snapcaster target — which wants a spell far more often than
		// it wants a land. Both are floors rather than values: a card
		// off the battlefield that the seat cannot cast is not a
		// resource, it is a lottery ticket.
		//
		// resolvedValue is deliberately NOT consulted here, and the
		// reason is that it would invert the answer: it prices a LAND
		// card as a mana source (permanentValue's land arm), which is
		// what the card would be worth on the battlefield and exactly
		// not what it is worth in a graveyard.
		if isLand(c) {
			return p.cfg.FuelFloor
		}
		return p.cfg.FuelIdle
	}
	return p.cfg.FuelIdle + p.cfg.FuelRecast*(st.w.Hand+p.resolvedValue(st, c, 0))
}

// castableFromHere reports whether the seat can still cast this card
// out of the zone it is in — an escape or flashback card in its own
// graveyard, a foretold or impulsed card in exile, a permission's
// grant.
//
// Read off the view's own stamps rather than re-derived: the server
// already answers "is this a cast surface for this seat" when it builds
// the projection (`CastableHere`, ADR 0066), and an offer list is
// stamped only for the seat the offers were computed for. A policy that
// re-derived it would be a second reader of the cast gate, which is the
// thing ADR 0033 §3 exists to prevent.
func castableFromHere(c *protocol.CardView) bool {
	return c.CastableHere || len(c.AlternativeCosts) > 0
}

// resolvedValue is what a card is worth once it RESOLVES: a permanent's
// body, or a mana-value proxy for an instant or sorcery whose text the
// wire does not carry.
//
// Lifted out of valueOfCast by #1013 so the fuel price and the cast
// price agree about what a card is worth by construction. `x` is the
// announced X, which is zero for a card nobody is casting yet.
func (p *Policy) resolvedValue(st *state, c *protocol.CardView, x int) float64 {
	if c == nil {
		return 0
	}
	var v float64
	switch {
	case isCreature(c) || isPermanentSpell(c):
		// What the permanent will be worth once it resolves. It
		// arrives summoning-sick and untapped; permanentValue reads
		// SummoningSick off the hand card, which is false there, so
		// discount a creature explicitly.
		pv := st.w.permanentValue(c)
		if isCreature(c) {
			pv *= st.w.SickCreature
		}
		v = pv
	default:
		// An instant or sorcery: no body, so its value is a mana-value
		// proxy for whatever it does that the wire does not describe.
		// (valueOfCast adds the TARGETS on top; the fuel price does
		// not, because a card being pitched is not being pointed at
		// anything.)
		v = p.cfg.SpellPerMana * float64(manaValue(c.ManaCost, x))
	}
	if c.IsCommander {
		v += p.cfg.CommanderBonus
	}
	if c.Unimplemented {
		// The engine will run none of this card's printed rules
		// (ADR 0037). It still costs a card.
		v -= p.cfg.SpellPerMana * float64(manaValue(c.ManaCost, x))
	}
	return v
}
