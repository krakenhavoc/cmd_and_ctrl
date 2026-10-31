package game

import "testing"

// ColorsAny is an any-of colour clause read off the card's effective colours
// (#4615): one named colour is enough, a colourless card never matches, and
// it composes with the type clauses as a further AND.
func TestPermissionFilterColorsAnyMatchesAtLeastOneNamedColour(t *testing.T) {
	f := PermissionFilter{CreatureOnly: true, ColorsAny: "RG"}
	for _, tc := range []struct {
		name string
		card Card
		want bool
	}{
		{"red creature", Card{TypeLine: "Creature — Bear", Colors: []string{"R"}}, true},
		{"green creature", Card{TypeLine: "Creature — Bear", Colors: []string{"G"}}, true},
		{"red-blue creature", Card{TypeLine: "Creature — Bear", Colors: []string{"U", "R"}}, true},
		{"blue creature", Card{TypeLine: "Creature — Bear", Colors: []string{"U"}}, false},
		{"colourless creature", Card{TypeLine: "Artifact Creature — Golem"}, false},
		{"red sorcery", Card{TypeLine: "Sorcery", Colors: []string{"R"}}, false},
		{"colour from the cost", Card{TypeLine: "Creature — Bear", ManaCost: "{2}{G}"}, true},
	} {
		if got := f.Matches(tc.card); got != tc.want {
			t.Errorf("%s: Matches = %v, want %v", tc.name, got, tc.want)
		}
	}
	if !(PermissionFilter{}).Matches(Card{TypeLine: "Creature — Bear", Colors: []string{"U"}}) {
		t.Error("the zero filter stopped matching a blue creature")
	}
}
