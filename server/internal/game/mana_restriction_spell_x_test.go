package game

import "testing"

// ManaRestrictSpellXCost (Troyan, Gutsy Explorer): the mana pays a SPELL
// that has {X} in its printed mana cost, whatever cost is paid for it, and
// nothing else (CR 106.6, CR 107.3).
func TestManaRestrictSpellXCostAdmitsOnlySpellsWithXInTheirManaCost(t *testing.T) {
	tags := []string{ManaRestrictSpellXCost}
	if !ManaSpendForCast(Card{ManaCost: "{X}{U}{U}", TypeLine: "Sorcery"}).allows(tags) {
		t.Error("refused a spell with {X} in its mana cost")
	}
	if ManaSpendForCast(Card{ManaCost: "{2}{U}", TypeLine: "Sorcery"}).allows(tags) {
		t.Error("admitted a spell without {X}")
	}
	// An alternative cost does not change the spell's own mana cost.
	alt := ManaSpendForCast(Card{ManaCost: "{X}{U}{U}", TypeLine: "Sorcery"})
	alt.CostHasX = false
	if !alt.allows(tags) {
		t.Error("refused a spell with {X} in its mana cost cast for a cost without {X}")
	}
	// An ability whose cost has {X} is not a spell.
	ability := ManaSpendForAbility(Card{TypeLine: "Artifact"})
	ability.CostHasX = true
	if ability.allows(tags) {
		t.Error("admitted an activated ability")
	}
	if (ManaSpendContext{Purpose: SpendPurposeUnlock, ManaCostHasX: true}).allows(tags) {
		t.Error("admitted a payment that is not a cast")
	}
}
