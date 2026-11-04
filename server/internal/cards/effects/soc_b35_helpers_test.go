package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// socB35Token makes a token as an effect would and lets its enters triggers
// reach the stack.
func socB35Token(t *testing.T, g *game.Game, owner uuid.UUID, row string) {
	t.Helper()
	socB11Token(t, g, owner, row)
	g.RunStateChecksForTest()
}
