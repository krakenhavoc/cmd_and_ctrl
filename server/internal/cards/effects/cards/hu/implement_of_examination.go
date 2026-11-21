package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0baf75b3-cc02-47cc-873a-8d5f17da0ec9",
		Name:         "Implement of Examination",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Implement of Examination — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
