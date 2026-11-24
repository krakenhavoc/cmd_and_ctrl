package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faceless One — Legendary Enchantment Creature — Background {5}, 3/3
// (Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"If Faceless One is your commander, choose a color before the game
//	 begins. Faceless One is the chosen color.
//	 Choose a Background (You can have a Background as a second
//	 commander.)"
//
// The first paragraph is CR 903.4b (ChoosesCommanderColor,
// game/commander_color.go): the deck checker reads the chosen colour off
// the deck's cards, and as the game begins the engine records it, or asks
// its owner when the deck shows none. Faceless One is that colour in every
// zone, and the colour joins its colour identity. Choose a Background is a
// deck-construction rule (CR 702.124k, partner.go in internal/deck).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "ef220170-c01a-4ca5-b887-f248c50227ee",
		Name:                  "Faceless One",
		Completeness:          effects.CompletenessFull,
		ChoosesCommanderColor: true,
	})
}
