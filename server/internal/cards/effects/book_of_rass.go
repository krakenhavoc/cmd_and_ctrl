package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Book of Rass — Artifact — Book {6} (The Dark, #5638):
//
//	"{2}, Pay 2 life: Draw a card."
//
// Greed's row with a generic {2}: the mana and the life are both
// checked before either is paid, so a player below 2 life cannot
// activate it, and paying the last 2 is legal. No tap, so it can be
// activated any number of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5e3ca74-ec33-4679-9366-8b1a45048783",
		Name:         "Book of Rass",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Pay 2 life: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), PayLife(2)),
			Effect:  drawOneCardEffect,
		}},
	})
}
