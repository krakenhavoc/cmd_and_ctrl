package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dromar's Cavern — Land — Lair (Dominaria Remastered, #5948):
//
//	"When this land enters, sacrifice it unless you return a non-Lair land
//	 you control to its owner's hand.
//	 {T}: Add {W}, {U}, or {B}."
//
// The trigger is the bounce lands' keyed return with "non-Lair land" as its
// match. Declining, or having no such land, sacrifices it.
//
// No simplification.
func init() {
	effects.Register(effects.GapLairSpec("d8b57707-796d-4488-8f91-65bb75bc6281", "Dromar's Cavern", "W|U|B", dromarsCavernReturn))
}

var dromarsCavernReturn = effects.NewReturnLandsOrSacrifice(
	"Dromar's Cavern — return a non-Lair land you control to its owner's hand (or sacrifice Dromar's Cavern)", 1,
	"choose-cards/s82-gap-b142-dromars-cavern-return", "vis/s82-gap-b142-dromars-cavern-after-return",
	effects.GapNonLairLand)
