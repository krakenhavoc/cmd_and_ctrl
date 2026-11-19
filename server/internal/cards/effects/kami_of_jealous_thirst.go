package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kami of Jealous Thirst — Creature — Spirit {2}{B}, 1/3 (Modern Horizons 3):
//
//	"Deathtouch
//	 {4}{B}: Each opponent loses 2 life and you gain 2 life. This ability costs
//	 {4}{B} less to activate if you've drawn three or more cards this turn.
//	 Activate only once each turn."
func init() {
	const label = "{4}{B}: Each opponent loses 2 life and you gain 2 life. This ability costs {4}{B} less to activate if you've drawn three or more cards this turn. Activate only once each turn."
	Register(Spec{
		OracleID:        "5973d8f6-242c-4014-b66e-aa5c78440be7",
		Name:            "Kami of Jealous Thirst",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerValue, LifeGain: 2},
			Cost:      ManaCost("{4}{B}"),
			Condition: OncePerTurnActivation(label),
			CostModifiers: []game.CostModifier{
				ActivationCostsLessMana("{4}{B}",
					"This ability costs {4}{B} less to activate if you've drawn three or more cards this turn.",
					gapB49DrawnThreeOrMoreThisTurn),
			},
			Effect: Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2}),
		}},
	})
}
