package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerChosenPlayer answers the open as-enters prompt with `seat`,
// through the wire resolver rather than by stamping the field — the
// prompt is the half of the card this file is testing.
func AnswerChosenPlayer(t *testing.T, g *game.Game, chooser, seat uuid.UUID) {
	t.Helper()
	c := PendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatalf("no choose-a-player prompt is open: %+v", g.PendingChoices)
	}
	idx := -1
	for i, opt := range c.PickOptions {
		if opt.Player == seat {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("seat %v is not offered by %+v", seat, c.PickOptions)
	}
	if err := g.ResolveOptionPick(c.ID, chooser, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
}
