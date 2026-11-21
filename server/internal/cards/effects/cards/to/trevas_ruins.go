package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treva's Ruins — Land — Lair (Dominaria Remastered, #5948):
//
//	"When this land enters, sacrifice it unless you return a non-Lair land
//	 you control to its owner's hand.
//	 {T}: Add {G}, {W}, or {U}."
//
// The trigger is the bounce lands' keyed return with "non-Lair land" as its
// match. Declining, or having no such land, sacrifices it.
//
// No simplification.
func init() {
	effects.Register(effects.GapB143LairSpec("7b2c7758-2b89-49ff-8838-8dc9880c7209", "Treva's Ruins", "G|W|U", trevasRuinsReturn))
}

var trevasRuinsReturn = effects.NewReturnLandsOrSacrifice(
	"Treva's Ruins — return a non-Lair land you control to its owner's hand (or sacrifice Treva's Ruins)", 1,
	"choose-cards/s82-gap-b143-trevas-ruins-return", "vis/s82-gap-b143-trevas-ruins-after-return",
	effects.GapB143NonLairLand)
