package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcane Encyclopedia — Artifact — Book {3} (Core Set 2019, #5527):
//
//	"{3}, {T}: Draw a card."
//
// A mana-and-tap ability that draws one card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d91ea728-2d69-42cb-bb5d-e2b058f0d7b1",
		Name:         "Arcane Encyclopedia",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
