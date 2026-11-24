package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snow-Covered Island — Basic Snow Land — Island (Secret Lair Drop, #5462):
//
//	"({T}: Add {U}.)"
//
// The mana ability is the intrinsic one of the Island land type (CR 305.6),
// derived from the type line, and the mana it makes comes from a snow
// source (CR 205.4g, CR 107.4h), so it pays {S}. Like every basic land,
// it is Basic, so it is not nonbasic for Blood Moon and the like.
//
// No simplification.
func init() {
	effects.Register(effects.SnowBasicSpec("5b2460a5-6ae5-4cad-ba94-1a9e98e6e4c0", "Snow-Covered Island"))
}
