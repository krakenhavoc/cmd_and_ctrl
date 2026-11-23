package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crosis's Catacombs — Land — Lair (Dominaria Remastered, #5948):
//
//	"When this land enters, sacrifice it unless you return a non-Lair land
//	 you control to its owner's hand.
//	 {T}: Add {U}, {B}, or {R}."
//
// The trigger is the bounce lands' keyed return with "non-Lair land" as its
// match. Declining, or having no such land, sacrifices it.
//
// No simplification.
func init() {
	effects.Register(effects.GapLairSpec("e9a7dede-3968-4b0e-a707-419d46a6fec9", "Crosis's Catacombs", "U|B|R", crosissCatacombsReturn))
}

var crosissCatacombsReturn = effects.NewReturnLandsOrSacrifice(
	"Crosis's Catacombs — return a non-Lair land you control to its owner's hand (or sacrifice Crosis's Catacombs)", 1,
	"choose-cards/s82-gap-b142-crosiss-catacombs-return", "vis/s82-gap-b142-crosiss-catacombs-after-return",
	effects.GapNonLairLand)
