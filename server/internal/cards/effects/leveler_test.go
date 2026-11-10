package effects

import "testing"

// A closed band holds from N1 to N2 inclusive (CR 711.2a); an open one
// from N3 up (CR 711.2b).
func TestLevelBandHoldsOnItsRangeOnly(t *testing.T) {
	closed := LevelBand{From: 2, To: 3}
	open := LevelBand{From: 4}
	for n, want := range map[int][2]bool{
		0: {false, false}, 1: {false, false}, 2: {true, false}, 3: {true, false},
		4: {false, true}, 5: {false, true}, 40: {false, true},
	} {
		if got := closed.holds(n); got != want[0] {
			t.Errorf("LEVEL 2-3 with %d counters: %v, want %v", n, got, want[0])
		}
		if got := open.holds(n); got != want[1] {
			t.Errorf("LEVEL 4+ with %d counters: %v, want %v", n, got, want[1])
		}
	}
	if closed.symbol() != "Level 2-3" || open.symbol() != "Level 4+" {
		t.Errorf("symbols %q and %q", closed.symbol(), open.symbol())
	}
}

// Level up is "[Cost]: Put a level counter on this permanent. Activate
// only as a sorcery." (CR 702.87a).
func TestLevelUpKeywordIsASorcerySpeedRow(t *testing.T) {
	ab := LevelUpKeyword("{3}")
	if ab.Label != "Level up {3}" || !ab.SorcerySpeed || ab.Effect == nil {
		t.Errorf("label %q, sorcery speed %v, effect set %v", ab.Label, ab.SorcerySpeed, ab.Effect != nil)
	}
}
