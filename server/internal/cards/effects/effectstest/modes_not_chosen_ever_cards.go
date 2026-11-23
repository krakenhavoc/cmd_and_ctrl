package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// passPriorityOnce passes priority a single time.
func PassPriorityOnce(t *testing.T, g *game.Game) {
	t.Helper()
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
}
