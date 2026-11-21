package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_b_crystal_test.go — The Darkness Crystal (#3213, slice kuja-b).

const KbDarknessCrystalOracle = "660f370a-b25b-4ea5-b765-7abf96899c1c"

func KbCrystal(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Darkness Crystal", TypeLine: "Legendary Artifact",
		OracleID: KbDarknessCrystalOracle, Owner: owner, Controller: owner,
	})
}
