package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azure Mage — Creature — Human Wizard {1}{U}, 2/1 (Magic 2012, #5573):
//
//	"{3}{U}: Draw a card."
//
// Archivist's draw row at a mana cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bbf9fd47-10e1-4749-b06f-4fe72200afc7",
		Name:         "Azure Mage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    ManaCost("{3}{U}"),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
