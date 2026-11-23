package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boiling Rock Prison — Land (Avatar: The Last Airbender, #4459):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}.
//	 {4}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's ten common duals, registered through
// avatarB20CyclingTapland (avatar_b20_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.AvatarB20CyclingTapland("6caabc45-4644-4941-a2f3-f04ee7b50a89", "Boiling Rock Prison", "B", "R"))
}
