package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func BattlefieldCountersOf(g *game.Game, id uuid.UUID, kind string) int {
	c, ok := BattlefieldCard(g, id)
	if !ok {
		return 0
	}
	return c.Counters[kind]
}
