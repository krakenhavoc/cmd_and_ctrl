package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Scrivener — Creature — Zombie Wizard {1}{B}, 2/1 (Dragon's Maze, S82
// OLD b406, #5564):
//
//	"If you would draw a card while you have no cards in hand, instead you draw
//	 two cards and you lose 1 life."
//
// Phial of Galadriel's empty-hand draw replacement plus the life loss
// (S82OldB406BloodScrivenerDraw). It reads the hand as each draw would happen,
// so "draw three" on an empty hand doubles only the first draw, and the draw
// step's card is doubled too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af66da4b-2f36-4dd2-a5c6-18e6023c5aba",
		Name:         "Blood Scrivener",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.S82OldB406BloodScrivenerDraw("Blood Scrivener: draw two cards and lose 1 life instead"),
		},
	})
}
