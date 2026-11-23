package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// socB14PT is a battlefield permanent's power and toughness as the game
// compares them: layers and +1/+1 counters included.
func SocB14PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	var p, tg int
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				p, tg, found = c.CurrentPower(), c.CurrentToughness(), true
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return p, tg
}
