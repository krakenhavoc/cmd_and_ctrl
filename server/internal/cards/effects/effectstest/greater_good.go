package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const GreaterGoodOracle = "dc0593c2-ccb4-4648-a592-c5bcd121dc72"

// pushGloriousAnthemFor puts a Glorious Anthem under `owner`'s control:
// creatures they control get +1/+1.
func PushGloriousAnthemFor(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: GloriousAnthemOracle, Owner: owner, Controller: owner,
	})
}
