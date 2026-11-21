package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// awakenedLand reads the land after the layers: its counters, its P/T, and
// whether it is an Elemental Land Creature with haste.
func AwakenedLand(t *testing.T, g *game.Game, land uuid.UUID) (counters, power, toughness int, elementalLandCreature, hasty bool) {
	t.Helper()
	power, toughness, landCreature, hasty := EarthbentBody(t, g, land)
	elemental := false
	g.ReadSnapshot(func() {
		if c := FindBattlefieldCardForTest(g, land); c != nil {
			elemental = c.HasSubtype("Elemental")
		}
	})
	return CountersOn(g, land, game.CounterPlusOne), power, toughness, landCreature && elemental, hasty
}
