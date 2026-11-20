package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alchemist's Apprentice — {1}{U} Creature — Human Wizard 1/1 (Avacyn Restored, #5570):
//
//	"Sacrifice this creature: Draw a card."
//
// Brass Secretary's row without the mana: sacrificing it is the
// whole cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de085c51-2555-450e-a479-ddbb3bea9e8c",
		Name:         "Alchemist's Apprentice",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    SacrificeThis(),
			Effect:  ebDrawACard,
		}},
	})
}
