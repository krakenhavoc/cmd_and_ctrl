package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2B07PickIfAsked answers a pending target prompt for chooser with the
// card, when the engine asked (a lone legal target is chosen for you).
func Cmdr2B07PickIfAsked(t *testing.T, g *game.Game, chooser, card uuid.UUID) {
	t.Helper()
	if LatestPickTarget(g, chooser) != nil {
		PickCard(t, g, chooser, card)
	}
}
