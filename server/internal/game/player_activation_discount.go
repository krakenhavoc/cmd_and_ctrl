package game

import "github.com/google/uuid"

// player_activation_discount.go — "activated abilities of <permanents> you
// control cost {N} less to activate" as a duration effect: The Dining Car's
// "Whenever chaos ensues, activated abilities of artifact tokens you control
// cost {2} less to activate this turn" (ADR 0216, #4870).
//
// A permanent's own discount (ActivationCostsLess on a Spec) lives on the
// permanent and ends when it does. This one is made by a resolving ability
// and outlives its source, the plane that made it, so it is a PlayerStatic
// payload with a CR 611.2 duration, bound at pricing as one more CR 601.2f
// reduction by activeCostModifiersLocked, the way a next-spell promise is.
// Built fresh per query from data; nothing else is stored.

// ActivationDiscountGrant is one such discount.
type ActivationDiscountGrant struct {
	// Amount is the generic mana each activation costs less.
	Amount int `json:"amount"`
	// ArtifactTokensOnly narrows it to abilities of artifact tokens the
	// player controls.
	ArtifactTokensOnly bool `json:"artifactTokensOnly,omitempty"`
	// Label is the clause as printed, for the event log.
	Label string `json:"label,omitempty"`
}

// GrantActivationDiscountForEffect gives `player` the discount for duration
// d. Caller must hold g.mu.
func (g *Game) GrantActivationDiscountForEffect(player uuid.UUID, grant ActivationDiscountGrant, source uuid.UUID, d Duration) {
	p := g.playerByIDLocked(player)
	if p == nil || grant.Amount <= 0 {
		return
	}
	p.Statics = append(p.Statics, PlayerStatic{
		ActivationDiscount: &grant,
		Source:             source,
		Label:              grant.Label,
		Duration:           d,
	})
}

// activationDiscountModifiersLocked binds every live discount of the
// activator's that reaches the ability q prices.
//
// Caller must hold g.mu (read or write).
func (g *Game) activationDiscountModifiersLocked(q CostQuery) []boundCostModifier {
	p := g.playerByIDLocked(q.Controller)
	if p == nil {
		return nil
	}
	var out []boundCostModifier
	for _, s := range p.Statics {
		d := s.ActivationDiscount
		if d == nil || g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		if q.Card.Controller != q.Controller {
			continue
		}
		if d.ArtifactTokensOnly && !(q.Card.IsArtifact() && q.Card.IsToken()) {
			continue
		}
		n := d.Amount
		src := q.Card
		src.Controller = q.Controller
		out = append(out, boundCostModifier{
			modifier: CostModifier{
				Kind:        CostReduction,
				Label:       d.Label,
				Activations: true,
				Amount:      func(CostQuery) int { return n },
			},
			source: src,
		})
	}
	return out
}
