package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// harmful_target.go prices a cast or an activation that aims a clause
// declared Harmful at a permanent of the bot's own (#4314).
//
// A clause is Harmful (effects.Harmful, game.TargetSpec.Harmful) when its
// effect hurts the permanent it picks and the clause may only be aimed at
// the caster's own: Simulacrum's "deals that much damage to target
// creature you control". targetsValue prices every permanent of the bot's
// at the same flat OwnPermanentTarget, so the bot aimed it at whichever
// creature came first, its best one included. Such a pick is priced with
// opponentsChoiceValue, as the pick_target branch of choices.go does
// (#4249): a permanent of its own costs what losing it does, so the least
// valuable legal one scores highest. A pick of an opponent's permanent, or
// of a player, keeps the price it always had.

// clauseHarmful reports whether clause `slot` of a statement is declared
// Harmful: the matching entry of `clauses` when the statement has several,
// else the one `legal_targets` for slot 0.
func clauseHarmful(first *protocol.LegalTargetsView, clauses []protocol.LegalTargetsView, slot int) bool {
	if len(clauses) > 0 {
		return slot >= 0 && slot < len(clauses) && clauses[slot].Harmful
	}
	return slot == 0 && first != nil && first.Harmful
}

// castHarmful is clauseHarmful for one pick of a cast, read from the same
// statement castEntryFor reads its purpose from.
func castHarmful(c *protocol.CardView, cp castParams, t targetRef) bool {
	if c == nil {
		return false
	}
	if c.Modes != nil && len(cp.Modes) > 0 {
		if t.Mode < 0 || t.Mode >= len(cp.Modes) {
			return false
		}
		opt := cp.Modes[t.Mode]
		if opt < 0 || opt >= len(c.Modes.Options) {
			return false
		}
		o := &c.Modes.Options[opt]
		return clauseHarmful(o.LegalTargets, o.Clauses, t.Slot)
	}
	if t.Mode != 0 {
		return false
	}
	if cp.AlternativeCost != "" {
		for i := range c.AlternativeCosts {
			if ac := &c.AlternativeCosts[i]; ac.Key == cp.AlternativeCost && ac.LegalTargets != nil {
				return clauseHarmful(ac.LegalTargets, ac.Clauses, t.Slot)
			}
		}
	}
	return clauseHarmful(c.LegalTargets, c.Clauses, t.Slot)
}

// rowHarmful is clauseHarmful for one pick of an activated row.
func rowHarmful(src *protocol.CardView, index int, t targetRef) bool {
	if t.Mode != 0 {
		return false
	}
	row := rowAt(src, index)
	return row != nil && clauseHarmful(row.LegalTargets, row.Clauses, t.Slot)
}

// pricedTargetsHarm is pricedTargetsValue with the picks of a Harmful
// clause that name a permanent of the bot's own priced by what losing it
// costs. Every other pick goes through pricedTargetsValue unchanged, so a
// card that declares nothing is priced exactly as before.
func (p *Policy) pricedTargetsHarm(st *state, targets []targetRef, harmful func(targetRef) bool, entryOf func(targetRef) *protocol.TargetPurposeView, self, source *protocol.CardView) (float64, bool) {
	var own, rest []targetRef
	for _, t := range targets {
		if t.Kind == "card" && harmful(t) {
			if c := st.bf[t.ID]; c != nil && c.Controller == st.me {
				own = append(own, t)
				continue
			}
		}
		rest = append(rest, t)
	}
	if len(own) == 0 {
		return p.pricedTargetsValue(st, targets, entryOf, self, source)
	}
	v, priced := p.pricedTargetsValue(st, rest, entryOf, self, source)
	return v + st.opponentsChoiceValue(p.cfg, st.me, own), priced
}
