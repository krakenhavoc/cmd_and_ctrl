package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mind_control_test.go pins the S24 layer-2 control change. The card
// is two lines; everything asserted here is the engine behind it.

const MindControlOracle = "5912546a-acc2-448c-b042-64bdac5ec129"

// controllerOf reads a battlefield card's controller through a
// snapshot, which forces the recompute — so it reads the
// post-layer-2 answer, which is the whole point of materialising.
func ControllerOf(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	var out uuid.UUID
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c.Controller, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

func SummoningSickOf(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	var out bool
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				out = game.HasSummoningSickness(&g.Battlefield.Cards[i])
				return
			}
		}
	})
	return out
}
