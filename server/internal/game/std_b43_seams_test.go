package game

import "testing"

// Seams built for Standard set batches 43-46 (#4452).

// ManaRestrictFaceDownSpell (Tin Street Gossip): the mana pays a spell cast
// face down (CR 708.4) and nothing else.
func TestManaRestrictFaceDownSpellAdmitsOnlyAFaceDownCast(t *testing.T) {
	tags := []string{ManaRestrictFaceDownSpell}
	down := Card{ManaCost: "{1}{R}", TypeLine: "Creature — Bear"}
	down.SetFaceDown(FaceDownDisguised)
	if !ManaSpendForCast(down).allows(tags) {
		t.Error("refused a face-down cast")
	}
	if ManaSpendForCast(Card{ManaCost: "{1}{R}", TypeLine: "Creature — Bear"}).allows(tags) {
		t.Error("admitted a face-up cast")
	}
	if ManaSpendForAbility(Card{TypeLine: "Artifact"}).allows(tags) {
		t.Error("admitted an activation")
	}
	if (ManaSpendContext{Purpose: SpendPurposeTurnFaceUp}).allows(tags) {
		t.Error("admitted a turn-face-up payment on its own")
	}
	both := []string{ManaRestrictAnyOf([]string{ManaRestrictFaceDownSpell}, []string{ManaRestrictTurnFaceUp})}
	if !(ManaSpendContext{Purpose: SpendPurposeTurnFaceUp}).allows(both) {
		t.Error("the combined restriction refused a turn-face-up payment")
	}
}

// PermissionFilter.PowerCapped (Assemble the Players): a creature card whose
// power is N or less, and nothing that is not a creature.
func TestPermissionFilterPowerCapped(t *testing.T) {
	f := PermissionFilter{CreatureOnly: true, PowerCapped: true, PowerAtMost: 2}
	for _, tc := range []struct {
		name  string
		power int
		want  bool
	}{{"0", 0, true}, {"2", 2, true}, {"3", 3, false}} {
		c := Card{TypeLine: "Creature — Bear", Power: tc.power, Toughness: 9}
		if got := f.Matches(c); got != tc.want {
			t.Errorf("power %s: Matches = %v, want %v", tc.name, got, tc.want)
		}
	}
	if (PermissionFilter{PowerCapped: true, PowerAtMost: 2}).Matches(Card{TypeLine: "Artifact", Power: 0}) {
		t.Error("a noncreature card matched a power clause")
	}
	if !(PermissionFilter{CreatureOnly: true, PowerCapped: true}).Matches(Card{TypeLine: "Creature — Bear", Power: 0, Toughness: 1}) {
		t.Error("a power 0 creature does not match PowerAtMost 0: the flag, not the number, says the clause is there")
	}
}
