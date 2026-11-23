package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const AzoriusChanceryOracle = "189fc8f4-17ac-4f1d-82c8-8401445bdaf4"

func PlayerByIDForTest(g *game.Game, id uuid.UUID) *game.Player {
	for _, p := range g.Seats {
		if p != nil && p.ID == id {
			return p
		}
	}
	return nil
}
