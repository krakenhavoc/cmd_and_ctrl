package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meditation Pools — Land (Avatar: The Last Airbender, #4459):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {U}.
//	 {4}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's ten common duals, registered through
// avatarB20CyclingTapland (avatar_b20_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.AvatarB20CyclingTapland("66208129-20fb-4cb4-8f00-8d4574c9c653", "Meditation Pools", "G", "U"))
}
