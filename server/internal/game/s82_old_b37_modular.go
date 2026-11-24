package game

// s82_old_b37_modular.go — "a modular triggered ability" as a cause a
// counter replacement can read (Zabaz, the Glimmerwasp, #5501).

// ResolvingModularTriggerForEffect reports whether the item resolving
// right now is a modular dies trigger (CR 702.43a). A counter
// replacement's AppliesTo reads it to tell the counters a modular trigger
// puts on a creature from any other placement. Caller must hold g.mu,
// as a replacement's AppliesTo does.
func (g *Game) ResolvingModularTriggerForEffect() bool {
	r := g.resolving
	if r == nil || r.hasCard || r.item == nil {
		return false
	}
	return r.item.Body == modularMoveCountersBody.key
}
