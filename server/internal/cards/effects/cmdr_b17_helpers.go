package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b17_helpers.go — shared pieces of Commander deck cards, batch 17
// (#5368).

// cmdrB17DefendersControlled is the number of creatures with defender
// `controller` controls right now, read through the layers (Blight Pile's
// X, Wingmantle Chaplain's count).
func cmdrB17DefendersControlled(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && game.HasKeyword(&c, "defender") {
			n++
		}
	}
	return n
}
