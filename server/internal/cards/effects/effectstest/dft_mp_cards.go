package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- fixtures ------------------------------------------------------------------

// dftMPLibraryTop puts cards on top of p's library, the first argument
// on top, and returns their ids in the same order.
func DftMPLibraryTop(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

// dftMPPassTurn ends the current turn, which also ends every "until end
// of turn" effect.
func DftMPPassTurn(t *testing.T, g *game.Game) {
	t.Helper()
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
}
