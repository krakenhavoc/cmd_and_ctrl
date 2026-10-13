package game

import (
	"math"
	"testing"
)

// TestSaturatingHelpersNeverWrap pins the arithmetic the writers lean on.
func TestSaturatingHelpersNeverWrap(t *testing.T) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"add small", SaturatingAdd(2, 3), 5},
		{"add at ceiling", SaturatingAdd(MaxCount, MaxCount), MaxCount},
		{"add int max", SaturatingAdd(math.MaxInt, math.MaxInt), MaxCount},
		{"add int min", SaturatingAdd(math.MinInt, math.MinInt), -MaxCount},
		{"mul small", SaturatingMul(6, 7), 42},
		{"mul ceiling squared", SaturatingMul(MaxCount, MaxCount), MaxCount},
		{"mul int max", SaturatingMul(math.MaxInt, math.MaxInt), MaxCount},
		{"mul negative", SaturatingMul(-MaxCount, MaxCount), -MaxCount},
		{"clamp low", ClampCount(math.MinInt), -MaxCount},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}

// TestPTCounterDeltaSaturates: a huge counter count times a large printed
// modifier is the multiplication that used to wrap.
func TestPTCounterDeltaSaturates(t *testing.T) {
	p, tough := PTCounterDelta(map[string]int{"+99999999999/+1": MaxCount, CounterPlusOne: MaxCount})
	if p != MaxCount || tough != MaxCount {
		t.Errorf("PTCounterDelta = %d/%d, want the ceiling %d on both", p, tough, MaxCount)
	}
	p, _ = PTCounterDelta(map[string]int{CounterPlusOne: math.MaxInt})
	if p != MaxCount {
		t.Errorf("an int-max counter total read as %d power, want %d", p, MaxCount)
	}
}

// TestParseProducedManaIsBounded: whatever string arrives, the parse is
// bounded, and says so.
func TestParseProducedManaIsBounded(t *testing.T) {
	slots, capped, err := parseProducedMana("{G99999999999999999999}")
	if err != nil || !capped || len(slots) != MaxProducedMana {
		t.Errorf("out-of-range amount: len=%d capped=%v err=%v, want %d slots, capped", len(slots), capped, err, MaxProducedMana)
	}
	slots, capped, err = parseProducedMana(RepeatMana("{G}", 1<<40))
	if err != nil || !capped || len(slots) != MaxProducedMana {
		t.Errorf("repeated slots: len=%d capped=%v err=%v, want %d slots, capped", len(slots), capped, err, MaxProducedMana)
	}
	slots, capped, err = parseProducedMana(RepeatMana("{G}", MaxProducedMana))
	if err != nil || capped || len(slots) != MaxProducedMana {
		t.Errorf("exactly the ceiling: len=%d capped=%v err=%v, want %d slots, not capped", len(slots), capped, err, MaxProducedMana)
	}
}
