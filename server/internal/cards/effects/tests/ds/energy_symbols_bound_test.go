package testsds

import (
	"runtime"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3086: EnergySymbols is a stack label; at the energy ceiling it must
// not build a string the size of the total.
func TestEnergySymbolsAreBoundedAtTheCeiling(t *testing.T) {
	if got := effects.EnergySymbols(2); got != "{E}{E}" {
		t.Errorf("two: %q", got)
	}
	if got := effects.EnergySymbols(0); got != "" {
		t.Errorf("zero: %q", got)
	}
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	got := effects.EnergySymbols(game.MaxCount)
	runtime.ReadMemStats(&after)
	if n := strings.Count(got, "{E}"); n > game.MaxProducedMana+1 {
		t.Errorf("%d symbols at the ceiling, want at most %d", n, game.MaxProducedMana+1)
	}
	if grew := after.TotalAlloc - before.TotalAlloc; grew > 8<<20 {
		t.Errorf("EnergySymbols(MaxCount) allocated %d bytes", grew)
	}
}
