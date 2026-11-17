package game

// emblem_cost_modifiers.go — an emblem's cost-changing statics (CR 114.3,
// CR 601.2f). Saheeli, Filigree Master's emblem: "Artifact spells you cast
// cost {1} less to cast." An emblem's abilities function in the command
// zone, so the cast price reads them as one more source beside the
// battlefield, the planar pass and eminence (activeCostModifiersLocked).

// appendEmblemCostModifiersLocked adds every emblem's cost modifiers that
// price q's kind of announcement, each bound to the emblem as its source:
// "you" is the emblem's controller, its owner, who never changes (CR 114.2).
//
// Caller must hold g.mu.
func (g *Game) appendEmblemCostModifiersLocked(out []boundCostModifier, q CostQuery) []boundCostModifier {
	if CatalogCostModifiers == nil {
		return out
	}
	g.forEachEmblemLocked(func(src *Card) bool {
		for _, m := range costModifiersOf(src) {
			if !m.pricesAnnouncement(q) {
				continue
			}
			out = append(out, boundCostModifier{modifier: m, source: *src})
		}
		return true
	})
	return out
}
