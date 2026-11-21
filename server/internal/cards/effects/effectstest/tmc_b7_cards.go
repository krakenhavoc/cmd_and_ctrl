package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmcB7PlayerRef is a player target ref.
func tmcB7PlayerRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
}

// tmcB7TryPickPlayer answers the open pick_target prompt for `chooser`
// with a player and returns the engine's answer.
func TmcB7TryPickPlayer(t *testing.T, g *game.Game, chooser, target uuid.UUID) error {
	t.Helper()
	B04WaitForPick(t, g, chooser)
	p := LatestPickTarget(g, chooser)
	return g.ResolvePickTarget(p.ID, chooser, tmcB7PlayerRef(target))
}

// tmcB7Tap taps a battlefield permanent directly.
func TmcB7Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
}
