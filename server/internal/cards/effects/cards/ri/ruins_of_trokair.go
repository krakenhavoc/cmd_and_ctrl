package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruins of Trokair — Land (Fallen Empires, #5637):
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 {T}, Sacrifice this land: Add {W}{W}."
//
// The white one of Fallen Empires' five sacrifice lands, built by
// tappedSacrificeLand (sacrifice_lands.go): a tapped entry, a tap for
// {W}, and a one-shot tap-and-sacrifice for {W}{W}.
//
// No simplification.
func init() {
	effects.Register(effects.TappedSacrificeLand("ad7b610c-e276-4c71-9b53-1ccb6d2dafd2", "Ruins of Trokair", "W"))
}
