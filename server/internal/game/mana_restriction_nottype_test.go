package game

import "testing"

// #2136: a negated type is a clause about a cast. It admits a spell with
// none of the named types and refuses everything else — an activation,
// an unlock and a payment with no purpose at all.
func TestNotTypeRestrictionAdmitsOnlyACastWithoutTheType(t *testing.T) {
	tags := []string{ManaRestrictCast, ManaRestrictNotType("Creature")}
	cases := []struct {
		name string
		ctx  ManaSpendContext
		want bool
	}{
		{"instant", instantSpendContext(), true},
		{"creature", creatureSpendContext(), false},
		{"artifact creature", ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Artifact", "Creature"}}, false},
		{"activation", ManaSpendContext{Purpose: SpendPurposeActivate, Types: []string{"Artifact"}}, false},
		{"unknown purpose", ManaSpendContext{Types: []string{"Instant"}}, false},
	}
	for _, c := range cases {
		if got := c.ctx.allows(tags); got != c.want {
			t.Errorf("%s: allows = %v, want %v", c.name, got, c.want)
		}
	}
	// The negated tag alone must not admit an activation either.
	activate := ManaSpendContext{Purpose: SpendPurposeActivate, Types: []string{"Artifact"}}
	if activate.allows([]string{ManaRestrictNotType("Creature")}) {
		t.Error("nottype alone admitted an activation")
	}
	// Several types mean none of them.
	both := []string{ManaRestrictCast, ManaRestrictNotType("Creature", "Land")}
	if (ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Land"}}).allows(both) {
		t.Error("noncreature-nonland admitted a land")
	}
	if !instantSpendContext().allows(both) {
		t.Error("noncreature-nonland refused an instant")
	}
}
