package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kami of Jealous Thirst — Creature — Spirit {2}{B}, 1/3 (Modern Horizons 3):
//
//	"Deathtouch
//	 {4}{B}: Each opponent loses 2 life and you gain 2 life. This ability costs
//	 {4}{B} less to activate if you've drawn three or more cards this turn.
//	 Activate only once each turn."
func init() {
	const label = "{4}{B}: Each opponent loses 2 life and you gain 2 life. This ability costs {4}{B} less to activate if you've drawn three or more cards this turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "5973d8f6-242c-4014-b66e-aa5c78440be7",
		Name:            "Kami of Jealous Thirst",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerValue, LifeGain: 2},
			Cost:      effects.ManaCost("{4}{B}"),
			Condition: effects.OncePerTurnActivation(label),
			CostModifiers: []game.CostModifier{
				effects.ActivationCostsLessMana("{4}{B}",
					"This ability costs {4}{B} less to activate if you've drawn three or more cards this turn.",
					effects.GapB49DrawnThreeOrMoreThisTurn),
			},
			Effect: effects.Do(effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2}),
		}},
	})
}
