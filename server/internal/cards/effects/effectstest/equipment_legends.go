package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// battlefieldCardCopy returns a copy of a battlefield card with its
// effective characteristic freshly resolved — the read CurrentPower
// and CurrentToughness need.
func BattlefieldCardCopy(t *testing.T, g *game.Game, cardID uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == cardID {
				out, found = c, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", cardID)
	}
	return out
}
