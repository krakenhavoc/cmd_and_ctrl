package coverage

import (
	"reflect"
	"testing"
)

// The embedded fixture is the checked-in directory, file for file in
// meaning: one freshness guard (TestOracleFixtureIsCurrent) covers both.
func TestPrintedOracleIsTheFixture(t *testing.T) {
	want, err := LoadOracleFixture(OracleFixtureDir)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	got := PrintedOracle()
	if len(got) == 0 {
		t.Fatal("the embedded oracle fixture parsed to nothing")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("the embedded fixture (%d cards) differs from %s (%d cards)", len(got), OracleFixtureDir, len(want))
	}
}

func TestPhraseMatcherIsWordBounded(t *testing.T) {
	m := PhraseMatcher([]string{"warp", "free cast"})
	for s, want := range map[string]bool{
		"Warp isn't implemented.":         true,
		"The FREE CAST is not offered.":   true,
		"It has no warpath.":              false,
		"Nothing about this card at all.": false,
	} {
		if got := m(s); got != want {
			t.Errorf("PhraseMatcher(%q) = %v, want %v", s, got, want)
		}
	}
}
