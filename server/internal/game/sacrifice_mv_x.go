package game

// sacrifice_mv_x.go — "Sacrifice a creature with mana value X" as a cost
// component of an activated ability (Sidisi, Regent of the Mire: "{T},
// Sacrifice a creature you control with mana value X other than Sidisi:
// Return target creature card with mana value X plus 1 from your
// graveyard to the battlefield").
//
// X is announced with the activation (CR 602.2b), the way Lazav's
// "mana value X" target and Shining Shoal's pitch read it, and the
// sacrificed permanent's mana value must equal it: TargetSpec.
// ManaValueEqualsX on the SacrificeOther clause. DemandsX counts it, so the
// view and the client ask for X as for any {X}. The same X binds the
// ability's target clauses (ManaValueEqualsX, ManaValueEqualsXPlus), which
// judge it at announce and again at resolution (CR 608.2b).

// SacrificeManaValueX reports whether a sacrifice clause is "with mana
// value X".
func SacrificeManaValueX(spec *TargetSpec) bool {
	return spec != nil && spec.ManaValueEqualsX
}

// sacrificeManaValueXAdmits is true for a clause that is not "mana value
// X", and otherwise when the permanent's mana value is the announced x
// (CR 202.3: {X} in a permanent's cost counts as 0).
func sacrificeManaValueXAdmits(g *Game, spec *TargetSpec, c Card, x int) bool {
	if !SacrificeManaValueX(spec) {
		return true
	}
	return boundAdmitsAt(g, spec, c, x)
}

// SacrificeManaValueXAdmitsForEffect is sacrificeManaValueXAdmits for the
// legal-move enumerator. Pure read.
func (g *Game) SacrificeManaValueXAdmitsForEffect(spec *TargetSpec, c Card, x int) bool {
	return sacrificeManaValueXAdmits(g, spec, c, x)
}
