package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paused_exit_tap_cost_test.go — #1427 with real cards. A commander
// that an effect is returning to its owner's hand stays on the
// battlefield while its owner answers CR 903.9b (ADR 0115: a destroyed
// one no longer pauses), and no cost may TAP it in that window: one
// card per tap shape — Birds of Paradise ({T} mana), Smuggler's Copter
// (crew), Bennie Bracks (a spell's convoke) and Survivors' Encampment
// ("{T}, Tap an untapped creature you control" on a mana ability).
// Without the gate each of them spent the bounced commander while
// its owner was still deciding. The engine half is
// game/paused_exit_tap_cost_test.go; the moving costs are #1445's
// paused_exit_cost_test.go.

// toMainPhase walks to the precombat main before anything is
// bounced: the CR 903.9b prompt blocks AdvanceStep, and each refusal
// here must come from the cost, not from the walk.
func ToMainPhase(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}
