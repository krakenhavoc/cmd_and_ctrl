package game

import "testing"

// Foretelling is a special action, so mana restricted to casting does not pay
// for it; mana restricted to foretelling does, and only that (Karfell
// Harbinger's "foretell a card from your hand or cast an instant or sorcery
// spell").
func TestForetellSpendPurposeIsPaidOnlyByForetellMana(t *testing.T) {
	foretell := SpecialActionSpendContext(SpecialActionForetell)
	if foretell.Purpose != SpendPurposeForetell {
		t.Fatalf("foretell spends as %q, want %q", foretell.Purpose, SpendPurposeForetell)
	}
	castOnly := []string{ManaRestrictCast, ManaRestrictAnyType("Instant", "Sorcery")}
	if foretell.allows(castOnly) {
		t.Error("mana restricted to casting an instant or sorcery paid for a foretell")
	}
	harbinger := []string{ManaRestrictAnyOf([]string{ManaRestrictForetell}, []string{ManaRestrictCast, ManaRestrictAnyType("Instant", "Sorcery")})}
	if !foretell.allows(harbinger) {
		t.Error("Harbinger's mana refused a foretell")
	}
	if !ManaSpendForCast(Card{ManaCost: "{1}{U}", TypeLine: "Instant"}).allows(harbinger) {
		t.Error("Harbinger's mana refused an instant spell")
	}
	if ManaSpendForCast(Card{ManaCost: "{1}{U}", TypeLine: "Creature — Bear"}).allows(harbinger) {
		t.Error("Harbinger's mana paid for a creature spell")
	}
	if ManaSpendForAbility(Card{TypeLine: "Artifact"}).allows(harbinger) {
		t.Error("Harbinger's mana paid for an activated ability")
	}
	if (ManaSpendContext{Purpose: SpendPurposeTurnFaceUp}).allows(harbinger) {
		t.Error("Harbinger's mana paid to turn a permanent face up")
	}
	if foretell.allows([]string{ManaRestrictTurnFaceUp}) {
		t.Error("turn-face-up mana paid for a foretell")
	}
}
