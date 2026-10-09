package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// net_removal.go prices removal net of what its target's controller
// gets back (#2679, NetRemoval).
//
// targetsValue prices every opposing permanent a spell points at as
// removed: its value × RemovalConfidence × leaderBoost. Two kinds of
// removal hand something back, and that price could see neither.
//
// A COMMANDER comes back. A commander destroyed, exiled, bounced or
// tucked may go to the command zone instead (CR 903.9a, 903.9b), and
// its owner casts it again for an additional {2} for each previous cast
// from there (CR 903.8). What the removal takes is that tax and the
// turn the body is away, not the body. So an opposing commander that
// its opponent owns is priced at
//
//	min(its value, 2 × CommanderTax + DamageToOpponent × power)
//
// before the same RemovalConfidence and leaderBoost: CommanderTax is
// what Strength already charges a seat per cast from the command zone,
// counted twice for the {2}, and one turn of the commander is the
// damage it does not deal while it is away, at the unit the attack
// planner prices a point of damage with. Nothing from the catalog is
// needed: `is_commander` is on the wire. A commander the bot owns that
// an opponent controls keeps its full price, since it comes back to the
// bot.
//
// A declared GIFT comes back. A target entry's `returns` (#2679) says
// what the target's controller is given when it is removed: Rapid
// Hybridization's 3/3, Swords to Plowshares' life, Path to Exile's land.
// It is valued as that seat would value it on its board, and charged on
// the same scale as the removal, value × RemovalConfidence ×
// leaderBoost, so the move is worth the removal less the gift:
//
//   - a creature token at its body (CreatureValue of a token of the
//     printed size);
//   - life equal to the target's power at what Strength counts it;
//   - a land at ManaSource.
//
// Only a target an opponent controls is netted. With NetRemoval off,
// both are priced as plain removal.

// commanderRemovalValue is what removing permanent c takes from its
// controller, when c is an opposing commander that comes back: the tax
// and one turn of the body, never more than its value. `full` is the
// value the removal would otherwise be priced at, before
// RemovalConfidence and leaderBoost.
func (st *state) commanderRemovalValue(cfg Config, c *protocol.CardView, full float64) float64 {
	if !cfg.NetRemoval || c == nil || !c.IsCommander || c.Controller == st.me || c.Owner == st.me {
		return full
	}
	back := 2*st.w.CommanderTax + cfg.DamageToOpponent*float64(max(c.Power, 0))
	return min(full, back)
}

// removalReturnsValue is what the move's declared returns give the
// controllers of the permanents it removes, priced on the removal's
// scale (RemovalConfidence × leaderBoost) so the caller subtracts it
// from the removal's price. Zero with NetRemoval off, for a pick with
// no `returns`, and for a permanent the bot controls.
func (p *Policy) removalReturnsValue(st *state, targets []targetRef, entryOf func(targetRef) *protocol.TargetPurposeView) float64 {
	if !p.cfg.NetRemoval {
		return 0
	}
	var v float64
	for _, t := range targets {
		if t.Kind == "player" {
			continue
		}
		e := entryOf(t)
		if e == nil || e.Returns == nil {
			continue
		}
		c := st.bf[t.ID]
		if c == nil || c.Controller == st.me {
			continue
		}
		v += st.returnValue(c, e.Returns) * p.cfg.RemovalConfidence * st.leaderBoost(p.cfg, c.Controller)
	}
	return v
}

// returnValue is what one declared return is worth to the controller of
// the removed permanent c.
func (st *state) returnValue(c *protocol.CardView, r *protocol.TargetReturnView) float64 {
	var v float64
	if r.CreatureTokens > 0 {
		token := protocol.CardView{Power: r.TokenPower, Toughness: r.TokenToughness, IsToken: true, Controller: c.Controller}
		v += float64(r.CreatureTokens) * st.w.CreatureValue(&token)
	}
	if r.LifeEqualToPower && c.Power > 0 {
		if e := st.evals[c.Controller]; e != nil {
			v += st.w.lifeChange(e, c.Power)
		}
	}
	v += float64(r.Lands) * st.w.ManaSource
	return v
}
