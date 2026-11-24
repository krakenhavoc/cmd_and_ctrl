package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snow-Covered Plains — Basic Snow Land — Plains (Secret Lair Drop, #5462):
//
//	"({T}: Add {W}.)"
//
// The mana ability is the intrinsic one of the Plains land type (CR 305.6),
// derived from the type line, and the mana it makes comes from a snow
// source (CR 205.4g, CR 107.4h), so it pays {S}. Like every basic land,
// it is Basic, so it is not nonbasic for Blood Moon and the like.
//
// No simplification.
func init() {
	effects.Register(effects.SnowBasicSpec("ac8cc74d-e43b-4118-bba0-dfa8b9c04d45", "Snow-Covered Plains"))
}
