package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB123AnAbilityYouDontControl is "…ability you don't control" (Nimble
// Obstructionist): the opposite of AnAbilityYouControl, read against the
// player choosing the target.
func S82OldB123AnAbilityYouDontControl() AbilityPredicate {
	return func(_ *game.Game, chooser uuid.UUID, item *game.StackItem) bool {
		return item.Controller != chooser
	}
}
