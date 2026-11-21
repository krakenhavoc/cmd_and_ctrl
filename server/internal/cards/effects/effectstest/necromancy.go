package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nmAdvanceTo moves the turn forward until it reaches `step`, with the
// stack empty.
func NmAdvanceTo(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for i := 0; i < 32 && g.Turn.Step != step; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Step != step {
		t.Fatalf("never reached %s (at %s)", step, g.Turn.Step)
	}
}

// nmCleanupTriggers counts the delayed triggers waiting for a cleanup step.
func NmCleanupTriggers(g *game.Game) int {
	n := 0
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.At == game.StepCleanup {
			n++
		}
	}
	return n
}
