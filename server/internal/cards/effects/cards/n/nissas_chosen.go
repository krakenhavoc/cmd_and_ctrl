package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nissa's Chosen — Creature — Elf Warrior {G}{G}, 2/3 (Zendikar, #5583):
//
//	"If this creature would die, put it on the bottom of its owner's library
//	 instead."
//
// A replacement of the move from the battlefield to a graveyard (CR 614.1a,
// 700.4): the card goes to the bottom of its owner's library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d14c9d2b-41c7-4c9a-a66a-cda34d29c7ba",
		Name:         "Nissa's Chosen",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB497ThisWouldDieToLibraryBottom("Nissa's Chosen")},
	})
}
