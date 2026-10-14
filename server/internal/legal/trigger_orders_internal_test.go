package legal

import (
	"fmt"
	"slices"
	"testing"
)

func TestTriggerOrdersOfferEveryOrderOfThree(t *testing.T) {
	for n, want := range map[int]int{0: 1, 1: 1, 2: 2, 3: 6} {
		got := triggerOrders(n)
		if len(got) != want {
			t.Errorf("n=%d: %d orders, want %d", n, len(got), want)
		}
		if n > 0 && !slices.Equal(got[0], identity(n)) {
			t.Errorf("n=%d: first order %v, want the collected order", n, got[0])
		}
	}
}

// Past three, every trigger can be put first and put last, the collected
// order stays first, and nothing repeats.
func TestTriggerOrdersPutEachTriggerFirstAndLast(t *testing.T) {
	for _, n := range []int{4, 5, 7} {
		got := triggerOrders(n)
		if !slices.Equal(got[0], identity(n)) {
			t.Errorf("n=%d: first order %v, want the collected order", n, got[0])
		}
		if len(got) > 2*n+2 {
			t.Errorf("n=%d: %d orders, want at most %d", n, len(got), 2*n+2)
		}
		seen := map[string]bool{}
		for _, o := range got {
			if k := fmt.Sprint(o); seen[k] {
				t.Errorf("n=%d: %v offered twice", n, o)
			} else {
				seen[k] = true
			}
		}
		for i := 0; i < n; i++ {
			var first, last bool
			for _, o := range got {
				first = first || o[0] == i
				last = last || o[n-1] == i
			}
			if !first || !last {
				t.Errorf("n=%d: trigger %d first=%v last=%v, want both", n, i, first, last)
			}
		}
	}
}

func identity(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}
