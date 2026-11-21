package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kili_the_resourceful_test.go — Kíli's "you may pay {0} rather than pay
// the equip cost of the first equip ability you activate each turn"
// (CR 118.9, #4475, game/equip_cost_alternative.go) and his draw.

const (
	KiliTheResourcefulOracle = "840de40b-590e-4bf0-8d16-22b0eff61202"
	KiliBonesplitterOracle   = "452e3f5f-ce17-4682-966b-5cc100210aee"
	KiliBlackbladeOracle     = "dca51281-fb21-45b6-beb4-1f13397caee2"
)

// twinOf is the {0} twin of the equipment's equip row `of`, as the engine
// offers it to `me` right now, and whether there is one.
func TwinOf(g *game.Game, me uuid.UUID, equipment uuid.UUID, of int) (game.EquipTwin, bool) {
	var out game.EquipTwin
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != equipment {
				continue
			}
			rows, origins := game.ActivatedAbilitiesWithOrigins(c)
			for _, tw := range g.EquipTwinsLocked(me, c, rows, origins) {
				if tw.Of == of {
					out, found = tw, true
				}
			}
		}
	})
	return out, found
}

func EquipRow(t *testing.T, g *game.Game, equipment uuid.UUID) int {
	t.Helper()
	rows := game.ActivatedAbilitiesForCard(game.Card{OracleID: OracleOfBattlefieldCard(t, g, equipment)})
	for i, r := range rows {
		if r.Equip {
			return i
		}
	}
	t.Fatal("no equip row")
	return -1
}
