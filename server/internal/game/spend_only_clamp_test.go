package game

import (
	"math"
	"testing"
)

// TestUnfoldedClampsHugeRun: the allocation size in unfolded cannot
// overflow however long the folded run claims to be (#4204).
func TestUnfoldedClampsHugeRun(t *testing.T) {
	c := ParsedCost{Required: []ColorRequirement{{Options: []string{"R"}}}}
	c.folded = reqRun{req: ColorRequirement{Options: []string{"B"}}, n: math.MaxInt}
	if got := len(c.unfolded().Required); got != 1+maxUnfoldedRun {
		t.Fatalf("len = %d, want %d", got, 1+maxUnfoldedRun)
	}
	c.folded.n = 3
	if got := len(c.unfolded().Required); got != 4 {
		t.Fatalf("len = %d, want 4", got)
	}
}
