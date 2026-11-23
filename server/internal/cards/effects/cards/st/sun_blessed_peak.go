package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sun-Blessed Peak — Land (Avatar: The Last Airbender, #4459):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {W}.
//	 {4}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's ten common duals, registered through
// avatarB20CyclingTapland (avatar_b20_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.AvatarB20CyclingTapland("d188a397-c7d9-4f16-bc21-d9e1b5ffc951", "Sun-Blessed Peak", "R", "W"))
}
