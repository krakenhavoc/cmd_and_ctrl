package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// passTo passes priority until `seat` holds it.
func PassTo(t *testing.T, g *game.Game, seat uuid.UUID) {
	t.Helper()
	for i := 0; i < len(g.Seats); i++ {
		if ph := g.Turn.PriorityHolder; ph >= 0 && g.Seats[ph].ID == seat {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("priority never reached %s", seat)
}
