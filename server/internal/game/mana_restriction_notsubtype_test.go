package game

import "testing"

// A negated subtype ("a non-Human creature spell", Animal Attendant)
// admits only a cast whose object lacks the subtype; a changeling
// creature spell has every creature type (CR 702.73a), so it is refused.
func TestNotSubtypeRestrictionAdmitsOnlyACastWithoutTheSubtype(t *testing.T) {
	tags := []string{ManaRestrictCast, ManaRestrictType("Creature"), ManaRestrictNotSubtype("Human")}
	cases := []struct {
		name string
		ctx  ManaSpendContext
		want bool
	}{
		{"bear", ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Creature"}, Subtypes: []string{"Bear"}}, true},
		{"human", ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Creature"}, Subtypes: []string{"Human", "Soldier"}}, false},
		{"changeling", ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Creature"}, Subtypes: []string{"Shapeshifter"}, AllCreatureTypes: true}, false},
		{"activation", ManaSpendContext{Purpose: SpendPurposeActivate, Types: []string{"Creature"}, Subtypes: []string{"Bear"}}, false},
		{"unknown purpose", ManaSpendContext{Types: []string{"Creature"}, Subtypes: []string{"Bear"}}, false},
	}
	for _, c := range cases {
		if got := c.ctx.allows(tags); got != c.want {
			t.Errorf("%s: allows = %v, want %v", c.name, got, c.want)
		}
	}
	// The negated tag alone must not admit an activation.
	if (ManaSpendContext{Purpose: SpendPurposeActivate}).allows([]string{ManaRestrictNotSubtype("Human")}) {
		t.Error("notsubtype alone admitted an activation")
	}
}
