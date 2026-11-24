package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snow-Covered Swamp — Basic Snow Land — Swamp (Secret Lair Drop, #5462):
//
//	"({T}: Add {B}.)"
//
// The mana ability is the intrinsic one of the Swamp land type (CR 305.6),
// derived from the type line, and the mana it makes comes from a snow
// source (CR 205.4g, CR 107.4h), so it pays {S}. Like every basic land,
// it is Basic, so it is not nonbasic for Blood Moon and the like.
//
// No simplification.
func init() {
	effects.Register(effects.SnowBasicSpec("d8239a86-7184-4005-ba1e-2dddcd756c47", "Snow-Covered Swamp"))
}
