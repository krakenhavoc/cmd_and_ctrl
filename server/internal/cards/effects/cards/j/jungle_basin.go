package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jungle Basin — Land (Visions, #5631):
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you return an untapped
//	 Forest you control to its owner's hand.
//	 {T}: Add {C}{G}."
//
// karooReturningUntapped (karoo_return_untapped_land.go), with the rest
// of its cycle.
//
// No simplification.
func init() {
	effects.Register(effects.KarooReturningUntapped("f922f90a-b1a2-4630-9266-40726ca89f74", "Jungle Basin", "Forest", "G", jungleBasinReturn))
}

var jungleBasinReturn = effects.NewReturnLandsOrSacrifice(
	"Jungle Basin — return an untapped Forest you control to its owner's hand (or sacrifice Jungle Basin)", 1,
	"choose-cards/jungle-basin-return-an-untapped-forest", "vis/jungle-basin-after-return",
	effects.UntappedLandOfType("Forest"))
