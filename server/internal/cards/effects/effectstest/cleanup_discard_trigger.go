package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counterOnBattlefieldCard reads one counter kind off a battlefield
// card, or 0 when the card has left.
func CounterOnBattlefieldCard(g *game.Game, id uuid.UUID, kind string) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Counters[kind]
		}
	}
	return 0
}
