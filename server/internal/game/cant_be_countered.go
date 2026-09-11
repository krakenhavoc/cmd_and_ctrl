package game

import "github.com/google/uuid"

// cant_be_countered.go — S23: the "This spell can't be countered"
// rider (Supreme Verdict, Cavern of Souls' grant, Thrun).
//
// Deliberately NOT a keyword and not a continuous effect. It is a
// per-card static statement about a spell on the stack, read at
// exactly one moment — when something tries to counter it — so it
// gets the same shape every other card-level static declaration in
// this engine has: a catalog hook, consulted at the choke point.
//
// The scope is honest about being narrow. Only the spell's OWN
// printed rider is modelled. Cavern of Souls-style GRANTS ("creature
// spells you control can't be countered") need a continuous effect
// over the stack, which the layer system does not reach — layers
// apply to permanents (ADR 0012), and a spell on the stack is not
// one. A grant joins this file when that gap is closed, not before.

// CatalogCantBeCountered is the catalog hook the effects package
// wires at init, mirroring CatalogTargetSpec / CatalogModeSpec /
// CatalogAdditionalCost. Nil means no card declares the rider.
var CatalogCantBeCountered func(oracleID string) bool

// spellCantBeCounteredLocked reports whether the spell on the stack
// prints "this spell can't be countered".
//
// Caller must hold g.mu.
func (g *Game) spellCantBeCounteredLocked(spellID uuid.UUID) bool {
	if CatalogCantBeCountered == nil || g.Stack == nil {
		return false
	}
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == spellID {
			oracle := CatalogKey(g.Stack.Cards[i])
			return oracle != "" && CatalogCantBeCountered(oracle)
		}
	}
	return false
}
