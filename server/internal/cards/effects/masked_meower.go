package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Masked Meower — Creature — Spider Cat Hero {R}, 1/1 (Marvel Super Heroes
// Commander, #4455):
//
//	"Haste
//	 Discard a card, Sacrifice this creature: Draw a card."
//
// Haste rides PrintedKeywords. Both costs are paid together at announce
// (CR 602.2b): the discard is picked from hand, and the sacrifice is the
// Meower itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a0037ff8-64db-45dd-bba1-db5c49095083",
		Name:            "Masked Meower",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "Discard a card, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(DiscardACard(), SacrificeThis()),
			Effect:  ebDrawACard,
		}},
	})
}
