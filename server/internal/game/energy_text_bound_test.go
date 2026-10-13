package game

import (
	"runtime"
	"strings"
	"testing"
)

// #3086: energy totals saturate at MaxCount, so the {E} text built from
// one must be bounded the way RepeatMana bounds produced mana. Normal
// totals are unchanged.
func TestEnergyPaymentWordsAreBoundedAtTheCeiling(t *testing.T) {
	if got := EnergyPaymentWords(0); got != "0 {E}" {
		t.Errorf("zero: %q", got)
	}
	if got := EnergyPaymentWords(3); got != "{E}{E}{E}" {
		t.Errorf("three: %q", got)
	}
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	got := EnergyPaymentWords(MaxCount)
	runtime.ReadMemStats(&after)
	if n := strings.Count(got, "{E}"); n > MaxProducedMana+1 {
		t.Errorf("%d symbols at the ceiling, want at most %d", n, MaxProducedMana+1)
	}
	if grew := after.TotalAlloc - before.TotalAlloc; grew > 8<<20 {
		t.Errorf("EnergyPaymentWords(MaxCount) allocated %d bytes", grew)
	}
}
