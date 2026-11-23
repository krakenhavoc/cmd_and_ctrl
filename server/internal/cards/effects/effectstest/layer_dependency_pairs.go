package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Maskwood Nexus + a permanent that turns itself off -------------

func PushNexusFor(g *game.Game, owner uuid.UUID) {
	PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Maskwood Nexus", TypeLine: "Artifact",
		OracleID: MaskwoodNexusOracle, Owner: owner, Controller: owner,
	})
}
