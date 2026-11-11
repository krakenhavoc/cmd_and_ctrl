package game

import "testing"

// #5368: "Spend this mana only to cast a spell from your graveyard"
// (Lord of the Forsaken). A positive tag, the graveyard twin of
// ManaRestrictCastFromExile: only a cast whose zone is a graveyard.
func TestCastFromGraveyardRestriction(t *testing.T) {
	pool := ManaPool{{Color: "C", Restrictions: []string{ManaRestrictCastFromGraveyard}}}
	cost, _ := ParseCost("{C}")
	spell := creatureSpendContext()

	for _, tc := range []struct {
		name string
		ctx  ManaSpendContext
		want bool
	}{
		{"a cast from the graveyard", withCastFrom(spell, ZoneGraveyard), true},
		{"a cast from hand", withCastFrom(spell, ZoneHand), false},
		{"a cast from exile", withCastFrom(spell, ZoneExile), false},
		{"a cast from the command zone", withCastFrom(spell, ZoneCommand), false},
		{"a cast from an unknown zone", spell, false},
		{"an activation", ManaSpendContext{Purpose: SpendPurposeActivate}, false},
		{"an unknown purpose", ManaSpendContext{}, false},
	} {
		if got := pool.CanPayFor(cost, 0, tc.ctx); got != tc.want {
			t.Errorf("%s: can pay = %v, want %v", tc.name, got, tc.want)
		}
	}
}
