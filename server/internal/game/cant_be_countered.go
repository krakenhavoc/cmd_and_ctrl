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
// Two sources answer it. The spell's OWN printed rider (the catalog
// hook), and since #1547 the MANA that paid for it: Cavern of Souls',
// Delighted Halfling's and Boseiju's "that spell can't be countered"
// is a spend rider on the token (mana_spend_rider.go), stamped Applied
// on the item's payment record when the token paid. Both are read
// here and nowhere else, so every counter verb — CounterTargetForEffect,
// the to-zone and to-library counters, and the put_in_library prompt's
// counterableSpellOnStackLocked — honours both through one gate.
//
// Still not modelled: a static GRANT over the stack ("creature spells
// you control can't be countered", Allosaurus Shepherd) — a continuous
// effect on spells, which the layer system does not reach (ADR 0012).
// Nor the manual CounterSpell sandbox action, which honours neither
// source by design: it is the table's override, not a rules verb.

// CatalogCantBeCountered is the catalog hook the effects package
// wires at init, mirroring CatalogTargetSpec / CatalogModeSpec /
// CatalogAdditionalCost. Nil means no card declares the rider.
var CatalogCantBeCountered func(oracleID string) bool

// spellCantBeCounteredLocked reports whether the spell on the stack
// prints "this spell can't be countered", or was paid for with mana
// whose rider says so (#1547).
//
// Caller must hold g.mu.
func (g *Game) spellCantBeCounteredLocked(spellID uuid.UUID) bool {
	// #1547: the mana that paid for it said so.
	if g.StackMeta[spellID].SpellCantBeCounteredByMana() {
		return true
	}
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
