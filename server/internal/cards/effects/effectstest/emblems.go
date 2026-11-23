package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emblemsOf reads a seat's emblems off the projection the wire uses.
func EmblemsOf(g *game.Game, playerID uuid.UUID) []game.EmblemView {
	var out []game.EmblemView
	g.ReadSnapshot(func() { out = g.EmblemsForPlayer(playerID) })
	return out
}

func EffectivePTOf(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	var p, tough int
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				eff := c.Effective()
				p, tough, found = eff.Power, eff.Toughness, true
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return p, tough
}

func HasKeywordOf(g *game.Game, id uuid.UUID, kw string) bool {
	has := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			for _, k := range c.Effective().Abilities {
				if k == kw {
					has = true
				}
			}
		}
	})
	return has
}
