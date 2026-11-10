package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Red Herring — Artifact Creature — Clue Fish {1}{R}, 2/2 (Murders at
// Karlov Manor, #4469):
//
//	"Haste
//	 This creature attacks each combat if able.
//	 {2}, Sacrifice this creature: Draw a card."
//
// Amarant Coral's forced attack (a creature that can't attack, or whose
// attack costs something, isn't forced) and Candlestick's printed draw.
// It is a Clue as well as a creature, so sacrificing it counts as
// sacrificing a Clue.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f0ee45b5-697b-48f8-b1a6-cb3ae75e4a29",
		Name:            "Red Herring",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
