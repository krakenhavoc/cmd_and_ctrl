package protocol

import "testing"

// #5405: an extra_phase line for several added steps counts them
// ("you get that many additional upkeep steps", Obeka).
func TestDescribeAddedPhasesCountsSteps(t *testing.T) {
	for _, tc := range []struct {
		label  string
		amount int
		want   string
	}{
		{"step:end", 1, "an additional end step"},
		{"step:upkeep", 1, "an additional upkeep step"},
		{"step:upkeep", 3, "three additional upkeep steps"},
		{"combat,combat", 2, "two additional combat phases"},
	} {
		if got := describeAddedPhases(tc.label, tc.amount); got != tc.want {
			t.Errorf("describeAddedPhases(%q, %d) = %q, want %q", tc.label, tc.amount, got, tc.want)
		}
	}
}
