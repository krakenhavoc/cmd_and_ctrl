package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dftCGPickRefused asserts that the open target pick refuses each of
// the given cards.
func DftCGPickRefused(t *testing.T, g *game.Game, chooser uuid.UUID, bad ...uuid.UUID) {
	t.Helper()
	pick := LatestPickTarget(g, chooser)
	if pick == nil {
		t.Fatal("no target pick")
	}
	for _, id := range bad {
		if err := g.ResolvePickTarget(pick.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
			t.Fatalf("%s was a legal target", id)
		}
	}
}
