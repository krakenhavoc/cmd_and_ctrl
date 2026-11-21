package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// warp_test.go — S29, #324. Warp is three shipped mechanisms in a
// trench coat (an alternative cost, a delayed trigger, an exile-play
// grant), so the tests are about the seams between them rather than
// about any one of them:
//
//  1. The discount is real — the creature enters for the warp cost.
//  2. The take-back is real — it is gone at the end step, and gone
//     to EXILE rather than to a graveyard, so it can come back.
//  3. "On a later turn" is real. This is the one a naive
//     implementation gets wrong, because an unbounded zone-bound
//     grant is live the instant it is stamped — and it is stamped
//     during the end step of the turn the creature was warped in.

// weftstalkerWarpOracle is the same key pirates_batch2_test.go
// already names; spelled again here so this file reads on its own.
const WeftstalkerWarpOracle = "926d52a5-4db1-46ce-9567-17c28bf56ae7"

// advanceThroughEndStep walks the cursor to the end step, which is
// where the warp trigger fires, and then lets it resolve.
func AdvanceThroughEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	PassPriorityAroundTable(t, g)
}
