package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selesnya Sanctuary — Land (EDHREC rank 572):
//
//	"This land enters tapped.
//	 When this land enters, return a land you control to its owner's
//	 hand.
//	 {T}: Add {G}{W}."
//
// The bounce-land shape — see b04BounceLand in batch04_helpers.go.
func init() {
	effects.Register(effects.B04BounceLand("00ef1c55-dea1-4564-bd57-66de86cba4df", "Selesnya Sanctuary", "{G}{W}"))
}
