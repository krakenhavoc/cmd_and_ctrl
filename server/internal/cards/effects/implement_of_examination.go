package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Implement of Examination — Artifact {3} (Aether Revolt,
// #5536):
//
//	"{U}, Sacrifice this artifact: Draw a card.
//	 When this artifact is put into a graveyard from the battlefield,
//	 draw a card."
//
// Two draws for one card: the ability's, and the trigger's as the
// sacrifice puts the Implement into the graveyard.
// The draw is Chromatic Star's leaves-for-the-graveyard trigger: it
// fires however the Implement got there, sacrificed to its own ability
// or destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0baf75b3-cc02-47cc-873a-8d5f17da0ec9",
		Name:         "Implement of Examination",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{U}"), SacrificeThis()),
			Effect:  Do(DrawCards{N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Implement of Examination — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
