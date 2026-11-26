package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB275DeliriumActivation is "Activate only if there are four or more
// card types among cards in your graveyard" (delirium, CR 207.2c), an
// ActivationCondition read for the activating player.
func S82OldB275DeliriumActivation(g *game.Game, controller, _ uuid.UUID) bool {
	return DrcDeliriumMet(g, controller)
}
