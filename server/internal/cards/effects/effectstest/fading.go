package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fadingToUpkeepOfSeat0 walks to the next beginning of seat 0's upkeep
// and stops there, with whatever that put on the stack still waiting.
func FadingToUpkeepOfSeat0(t *testing.T, g *game.Game) {
	t.Helper()
	FadingToUpkeepOf(t, g, 0)
}

// fadingToUpkeepOf is fadingToUpkeepOfSeat0 for any seat.
func FadingToUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == seat && g.Turn.Step == game.StepUpkeep); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// fadingTriggerWaiting reports whether a fading upkeep trigger from id
// is waiting, on the stack or about to be put there.
func FadingTriggerWaiting(g *game.Game, id uuid.UUID) bool {
	if item := TriggerOnStack(g, id); item != nil {
		return true
	}
	for _, p := range g.PendingTriggers {
		if p != nil && p.SourceCardID == id {
			return true
		}
	}
	return false
}
