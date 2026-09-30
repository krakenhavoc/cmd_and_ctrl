package game

import "testing"

// Powerstone (#1727): "can't be spent to cast a nonartifact spell" pays
// for an artifact spell and for an activation, and for nothing else.
func TestNotNonartifactSpellRestriction(t *testing.T) {
	pool := ManaPool{{Color: "C", Restrictions: []string{ManaRestrictNotNonartifactSpell}}}
	cost, _ := ParseCost("{C}")

	artifactSpell := ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Artifact", "Creature"}}
	activation := ManaSpendContext{Purpose: SpendPurposeActivate, Types: []string{"Creature"}}
	if !pool.CanPayFor(cost, 0, artifactSpell) {
		t.Error("Powerstone mana refused an artifact spell")
	}
	if !pool.CanPayFor(cost, 0, activation) {
		t.Error("Powerstone mana refused an activated ability")
	}
	if pool.CanPayFor(cost, 0, creatureSpendContext()) {
		t.Error("Powerstone mana paid for a nonartifact creature spell")
	}
	if pool.CanPayFor(cost, 0, instantSpendContext()) {
		t.Error("Powerstone mana paid for an instant")
	}
	if pool.CanPayFor(cost, 0, ManaSpendContext{}) {
		t.Error("an unknown purpose must refuse restricted mana")
	}
}
