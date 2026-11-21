package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// power_pack_test.go — Power Pack (#1815): a graveyard target chosen at
// random, then a free cast at your next upkeep.

const PowerPackOracle = "3d5c98ff-fcf1-42bd-9533-2648791a4f45"

func PpGraveyardCard(g *game.Game, owner *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		owner.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Owner: owner.ID, Controller: owner.ID})
	})
	return id
}
