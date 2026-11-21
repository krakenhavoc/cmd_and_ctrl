package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// raffine_a_dauthi_test.go — Dauthi Voidwalker (#3212, slice raffine-a).

const DauthiOracle = "f1c2dbe2-fbe0-4058-bdf1-91d1b1832786"

func VoidCounters(g *game.Game, id uuid.UUID) int {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c.Counters["void"]
		}
	}
	return -1
}
