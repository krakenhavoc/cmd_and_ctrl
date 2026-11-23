package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rith's Grove — Land — Lair (Dominaria Remastered, #5948):
//
//	"When this land enters, sacrifice it unless you return a non-Lair land
//	 you control to its owner's hand.
//	 {T}: Add {R}, {G}, or {W}."
//
// The trigger is the bounce lands' keyed return with "non-Lair land" as its
// match. Declining, or having no such land, sacrifices it.
//
// No simplification.
func init() {
	effects.Register(effects.GapLairSpec("e13289e5-370b-435b-a38e-cf57c3078cec", "Rith's Grove", "R|G|W", rithsGroveReturn))
}

var rithsGroveReturn = effects.NewReturnLandsOrSacrifice(
	"Rith's Grove — return a non-Lair land you control to its owner's hand (or sacrifice Rith's Grove)", 1,
	"choose-cards/s82-gap-b142-riths-grove-return", "vis/s82-gap-b142-riths-grove-after-return",
	effects.GapNonLairLand)
