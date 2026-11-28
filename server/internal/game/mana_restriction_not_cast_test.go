package game

import "testing"

// Thran Turbine: "This mana can't be spent to cast spells" pays for an
// activated ability and for nothing that casts.
func TestNotCastRestriction(t *testing.T) {
	pool := ManaPool{{Color: "C", Restrictions: []string{ManaRestrictNotCast}}}
	cost, _ := ParseCost("{C}")

	if !pool.CanPayFor(cost, 0, ManaSpendContext{Purpose: SpendPurposeActivate, Types: []string{"Creature"}}) {
		t.Error("the mana refused an activated ability")
	}
	if !pool.CanPayFor(cost, 0, ManaSpendContext{Purpose: SpendPurposeCumulativeUpkeep}) {
		t.Error("the mana refused a cumulative upkeep payment")
	}
	if pool.CanPayFor(cost, 0, creatureSpendContext()) {
		t.Error("the mana paid for a creature spell")
	}
	if pool.CanPayFor(cost, 0, instantSpendContext()) {
		t.Error("the mana paid for an instant")
	}
	if pool.CanPayFor(cost, 0, ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Artifact"}}) {
		t.Error("the mana paid for an artifact spell")
	}
	if pool.CanPayFor(cost, 0, ManaSpendContext{}) {
		t.Error("an unknown purpose must refuse restricted mana")
	}
}
