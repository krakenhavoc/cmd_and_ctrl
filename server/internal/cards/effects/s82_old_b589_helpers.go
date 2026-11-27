package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b589_helpers.go — Saviors of Kamigawa, batch 589 (#5601): the
// activation gate of "Activate only if you have more cards in hand than each
// opponent".

// OldB589MoreCardsInHandThanEachOpponent is "Activate only if you have more
// cards in hand than each opponent" (Kitsune Bonesetter). Public information
// only: it counts the hands as sizes. An eliminated opponent is not one
// (CR 800.4a); with no opponent left the comparison is vacuously true.
func OldB589MoreCardsInHandThanEachOpponent() ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		mine := CardsInHand(g, controller)
		return !EachOpponent(g, controller, func(opp uuid.UUID) bool {
			return CardsInHand(g, opp) >= mine
		})
	}
}
