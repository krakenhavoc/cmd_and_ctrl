package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Prismatic Piper — Legendary Creature — Shapeshifter {5}, 3/3
// (Commander Masters, #5948):
//
//	"If The Prismatic Piper is your commander, choose a color before the
//	 game begins. The Prismatic Piper is the chosen color.
//	 Partner (You can have two commanders if both have partner.)"
//
// The first paragraph is CR 903.4b (ChoosesCommanderColor,
// game/commander_color.go): the deck checker reads the chosen colour off
// the deck's cards (internal/deck/commander_color.go), and as the game
// begins the engine records it, or asks its owner when the deck shows
// none. The Piper is that colour in every zone, and the colour joins its
// colour identity. Partner is a deck-construction rule (partner.go in
// internal/deck).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "b2aeee90-7b8d-47a6-9167-cc21f7fbfd43",
		Name:                  "The Prismatic Piper",
		Completeness:          effects.CompletenessFull,
		ChoosesCommanderColor: true,
	})
}
