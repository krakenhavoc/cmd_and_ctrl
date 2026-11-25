package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snow-Covered Forest — Basic Snow Land — Forest (Secret Lair Drop, #5462):
//
//	"({T}: Add {G}.)"
//
// The mana ability is the intrinsic one of the Forest land type (CR 305.6),
// derived from the type line, and the mana it makes comes from a snow
// source (CR 205.4g, CR 107.4h), so it pays {S}. Like every basic land,
// it is Basic, so it is not nonbasic for Blood Moon and the like.
//
// No simplification.
func init() {
	effects.Register(effects.SnowBasicSpec("5f0d3be8-e63e-4ade-ae58-6b0c14f2ce6d", "Snow-Covered Forest"))
}
