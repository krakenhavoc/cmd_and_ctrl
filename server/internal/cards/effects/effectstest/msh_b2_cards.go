package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mshB2PickIfAsked answers a target pick the engine opened, if it opened
// one: a lone legal target may be chosen for the player.
func MshB2PickIfAsked(t *testing.T, g *game.Game, chooser, id uuid.UUID) {
	t.Helper()
	if LatestPickTarget(g, chooser) != nil {
		PickCard(t, g, chooser, id)
	}
}
