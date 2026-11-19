package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// PowerAtMostWarriorsYouControl is "creature with power less than or
// equal to the number of Warriors you control" (Goma Fada Vanguard),
// counted as the target is chosen and again as the ability resolves
// (CR 608.2b). "You" is the player choosing.
func PowerAtMostWarriorsYouControl() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		warrior := c.LentSubtype("Warrior")
		n := 0
		for _, p := range g.BattlefieldCardsForEffect() {
			if p.Controller == caster && p.HasSubtype(warrior) {
				n++
			}
		}
		return c.CurrentPower() <= n
	}
}
