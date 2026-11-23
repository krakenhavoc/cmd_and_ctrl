package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darigaaz's Caldera — Land — Lair (Dominaria Remastered, #5948):
//
//	"When this land enters, sacrifice it unless you return a non-Lair land
//	 you control to its owner's hand.
//	 {T}: Add {B}, {R}, or {G}."
//
// The trigger is the bounce lands' keyed return with "non-Lair land" as its
// match. Declining, or having no such land, sacrifices it.
//
// No simplification.
func init() {
	effects.Register(effects.GapLairSpec("19b58ec9-bb88-4193-8ea8-c8f09ceec1ed", "Darigaaz's Caldera", "B|R|G", darigaazsCalderaReturn))
}

var darigaazsCalderaReturn = effects.NewReturnLandsOrSacrifice(
	"Darigaaz's Caldera — return a non-Lair land you control to its owner's hand (or sacrifice Darigaaz's Caldera)", 1,
	"choose-cards/s82-gap-b142-darigaazs-caldera-return", "vis/s82-gap-b142-darigaazs-caldera-after-return",
	effects.GapNonLairLand)
