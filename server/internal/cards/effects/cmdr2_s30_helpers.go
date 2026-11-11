package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2S30CreatureCardsIn is the number of creature cards in `player`'s
// graveyard and the greatest power among them (0 when there are none).
func cmdr2S30CreatureCardsIn(g *game.Game, player uuid.UUID) (count, greatestPower int) {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0, 0
	}
	for _, c := range p.Graveyard.Cards {
		if !c.IsCreature() {
			continue
		}
		count++
		if c.Power > greatestPower {
			greatestPower = c.Power
		}
	}
	return count, greatestPower
}
