package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Felhide Brawler — {1}{B} Creature — Minotaur 2/2 (Born of the Gods):
//
//	"This creature can't block unless you control another Minotaur."
//
// A CR 509.1b block rule on the creature itself, read as the block is declared: another
// Minotaur is a different permanent the controller controls that has the type, through the
// layers and the card's text.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ddd33e1-b48e-42e2-a60a-f09f151d3c1d",
		Name:         "Felhide Brawler",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockUnlessYouControlAnother("Minotaur", "it can't block unless its controller controls another Minotaur"),
		},
	})
}
