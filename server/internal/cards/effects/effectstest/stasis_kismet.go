package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stasis_kismet_test.go covers the S17 sub-PR 4 replacements:
//   - Stasis cancels StepUntap so the cursor skips straight past
//     it to StepUpkeep.
//   - Kismet taps opponents' creatures / artifacts / lands on
//     entry — verified for spell-resolve and land-play entry sites.
//   - Fetched-land enters-tapped finishers for Cultivate / Path /
//     Solemn Simulacrum via the new TappedOnEntry primitive flag.

const (
	StasisOracle = "a8cf1379-0195-4e11-b994-481ef1284245"
	KismetOracle = "81fdd1c4-d43b-4f8b-8712-7c2bf45a3e0b"
)

// walkToUpkeepOfSeat0 advances the cursor to seat 0's upkeep on a
// later turn, which is when Stasis's own sacrifice trigger fires.
func WalkToUpkeepOfSeat0(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 80; i++ {
		if g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepUpkeep && g.Turn.Round >= 2 {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep iter %d: %v", i, err)
		}
	}
	t.Fatalf("never reached seat 0 upkeep: turn=%d seat=%d step=%v",
		g.Turn.Round, g.Turn.ActiveSeat, g.Turn.Step)
}
