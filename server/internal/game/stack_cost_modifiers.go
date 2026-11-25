package game

import "github.com/google/uuid"

// stack_cost_modifiers.go — #6222: a spell's cost-changing static that
// functions while the spell is on the stack (CR 113.6b). Kaervek's
// Torch: "As long as Kaervek's Torch is on the stack, spells that target
// it cost {2} more to cast." One more source list into the CR 601.2f
// gather (activeCostModifiersLocked), beside the battlefield, eminence,
// planes and emblems.

// appendStackCostModifiersLocked adds the OnStack modifiers of every
// spell on the stack that prices q's kind of announcement, each bound to
// the spell as its source with the spell's controller as "you" (CR
// 108.4). The spell being priced is skipped: it is on the stack while it
// is cast (CR 601.2a), and its own "this spell costs" clause is the self
// slot, not this.
//
// Caller must hold g.mu.
func (g *Game) appendStackCostModifiersLocked(out []boundCostModifier, q CostQuery) []boundCostModifier {
	if CatalogCostModifiers == nil || g.Stack == nil {
		return out
	}
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == q.Card.InstanceID {
			continue
		}
		mods := costModifiersOf(&g.Stack.Cards[i])
		if len(mods) == 0 {
			continue
		}
		src := g.Stack.Cards[i]
		if item := g.StackMeta[src.InstanceID]; item != nil && item.Controller != uuid.Nil {
			src.Controller = item.Controller
		}
		for _, m := range mods {
			if !m.OnStack || !m.pricesAnnouncement(q) {
				continue
			}
			out = append(out, boundCostModifier{modifier: m, source: src})
		}
	}
	return out
}
