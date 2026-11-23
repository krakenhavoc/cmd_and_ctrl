package game

import "testing"

// ManaRestrictKickedSpell (Volshe Tideturner): the mana pays a spell cast
// with its kicker paid (CR 702.33a) and nothing else.
func TestManaRestrictKickedSpellAdmitsOnlyAKickedCast(t *testing.T) {
	tags := []string{ManaRestrictKickedSpell}
	card := Card{ManaCost: "{1}{R}", TypeLine: "Creature — Bear"}
	if ManaSpendForCast(card).allows(tags) {
		t.Error("admitted a cast that names no optional costs")
	}
	if ManaSpendForCastParams(card, CastSpellParams{}).allows(tags) {
		t.Error("admitted an unkicked cast")
	}
	if ManaSpendForAbility(Card{TypeLine: "Artifact"}).allows(tags) {
		t.Error("admitted an activation")
	}
	if !(ManaSpendContext{Purpose: SpendPurposeCast, KickedSpell: true}).allows(tags) {
		t.Error("refused a kicked cast")
	}
	both := []string{ManaRestrictAnyOf(
		[]string{ManaRestrictCast, ManaRestrictAnyType("Instant", "Sorcery")},
		[]string{ManaRestrictCast, ManaRestrictKickedSpell},
	)}
	if !(ManaSpendContext{Purpose: SpendPurposeCast, KickedSpell: true, Types: []string{"Creature"}}).allows(both) {
		t.Error("the combined restriction refused a kicked creature spell")
	}
	if !(ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Sorcery"}}).allows(both) {
		t.Error("the combined restriction refused an unkicked sorcery")
	}
	if (ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Creature"}}).allows(both) {
		t.Error("the combined restriction admitted an unkicked creature")
	}
}
