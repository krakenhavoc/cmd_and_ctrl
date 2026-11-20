package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hmlUntapSkips is the permanent's next-untap-step marks.
func HmlUntapSkips(t *testing.T, g *game.Game, id uuid.UUID) []game.UntapSkip {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("battlefield card %s is missing", id)
	}
	return c.NextUntapSkips
}
