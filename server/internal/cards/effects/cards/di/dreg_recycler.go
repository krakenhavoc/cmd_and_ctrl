package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreg Recycler — Creature — Phyrexian Beast 2/2 {1}{B} (March of the Machine, S82):
//
//	"{T}, Sacrifice an artifact or creature: Each opponent loses 1 life and you gain 1 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "556c4dad-002e-41a8-8ba3-8b4eda70232f",
		Name:         "Dreg Recycler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact or creature: Each opponent loses 1 life and you gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, LifeGain: 1},
			Cost: game.AbilityCost{
				Tap:            true,
				SacrificeOther: effects.SacrificeSpec("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			},
			Effect: effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1}),
		}},
	})
}
