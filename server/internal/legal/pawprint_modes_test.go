package legal

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pawprint_modes_test.go — CR 700.2i, #4786: the enumerator offers a
// pawprint spell every selection its budget allows and none it doesn't,
// the same bound the announce gate holds it to.

func TestModeSelectionsKeepsAPawprintSpellWithinItsBudget(t *testing.T) {
	ms := &game.ModeSpec{
		Prompt: "Choose up to five {P} worth of modes", Min: 0, Max: 5, Repeatable: true, Pawprints: 5,
		Options: []game.ModeOption{{Label: "A.", Pawprints: 1}, {Label: "B.", Pawprints: 2}, {Label: "C.", Pawprints: 3}},
	}
	got := modeSelections([]int{0, 1, 2}, 0, 5, true, 1000, ms.WithinPawprints)
	// Multisets of {1, 2, 3} {P} adding up to 5 or less, the empty one
	// included: 12 without the three, 4 with it.
	if len(got) != 16 {
		t.Errorf("got %d selections, want 16: %v", len(got), got)
	}
	seen := map[string]bool{}
	for _, sel := range got {
		if n := game.PawprintsOf(ms, sel); n > 5 {
			t.Errorf("offered %v, which spends %d {P}", sel, n)
		}
		key := ""
		for _, m := range sel {
			key += string(rune('a' + m))
		}
		if seen[key] {
			t.Errorf("offered %v twice", sel)
		}
		seen[key] = true
	}
	// The walk is over non-decreasing option indexes, so a multiset is
	// keyed in that order.
	for _, want := range []string{"", "aaaaa", "bc", "abb", "ac", "aab"} {
		if !seen[want] {
			t.Errorf("never offered %q", want)
		}
	}
}
