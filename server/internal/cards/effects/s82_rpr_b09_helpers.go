package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82RprB09NoOtherThopter is Thopter Assembly's intervening "if" (CR
// 603.4): the controller controls no Thopter other than the source.
func S82RprB09NoOtherThopter(g *game.Game, controller, source uuid.UUID) bool {
	thopter := Subtype("Thopter")
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != source && c.Controller == controller && thopter(g, controller, c) {
			return false
		}
	}
	return true
}
