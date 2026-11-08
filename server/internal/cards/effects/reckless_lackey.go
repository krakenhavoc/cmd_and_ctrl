package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reckless Lackey — Creature — Goblin Pirate {R}, 1/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"First strike, haste
//	 {2}{R}, Sacrifice this creature: Draw a card and create a Treasure
//	 token. (It's an artifact with "{T}, Sacrifice this token: Add one
//	 mana of any color.")"
//
// Illvoi Galeblade's sacrifice-to-draw row with the catalog's Treasure
// after the card. The sacrifice is part of the cost (CR 602.2b), so the
// card and the Treasure come even though the Lackey is gone by the time
// the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "73733fa5-d292-40ec-8f17-05ba5feb40ab",
		Name:            "Reckless Lackey",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice this creature: Draw a card and create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{R}"), SacrificeThis()),
			Effect:  otjRDrawThenTreasure(),
		}},
	})
}
