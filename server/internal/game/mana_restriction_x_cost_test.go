package game

import "testing"

// ManaRestrictXCost (#4834, Elementalist's Palette): the mana pays a
// cost that contains {X} — a spell whose mana cost has one, an
// alternative cost that has one, an ability whose cost has one — and
// nothing else.
func TestManaRestrictXCostAdmitsOnlyCostsWithX(t *testing.T) {
	tags := []string{ManaRestrictXCost}
	if !ManaSpendForCast(Card{ManaCost: "{X}{U}{U}", TypeLine: "Sorcery"}).allows(tags) {
		t.Error("refused a spell with {X} in its mana cost")
	}
	if ManaSpendForCast(Card{ManaCost: "{2}{U}", TypeLine: "Sorcery"}).allows(tags) {
		t.Error("admitted a spell without {X}")
	}
	withX := ManaSpendForAbility(Card{TypeLine: "Artifact"})
	withX.CostHasX = true
	if !withX.allows(tags) {
		t.Error("refused an ability whose cost has {X}")
	}
	if ManaSpendForAbility(Card{TypeLine: "Artifact"}).allows(tags) {
		t.Error("admitted an ability whose cost has no {X}")
	}
	unlock := ManaSpendContext{Purpose: SpendPurposeUnlock, CostHasX: true}
	if unlock.allows(tags) {
		t.Error("admitted a payment that is neither a cast nor an activation")
	}
}
