package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hidden Stockpile — Enchantment {W}{B}:
//
//	"Revolt — At the beginning of your end step, if a permanent left
//	 the battlefield under your control this turn, create a 1/1
//	 colorless Servo artifact creature token.
//	 {1}, Sacrifice a creature: Scry 1."
//
// Revolt is an intervening if (revolt.go). The sacrifice is a cost, so
// the sacrificed creature itself counts toward revolt at the end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5ded323-75c8-471e-a516-238b9d6b06d3",
		Name:         "Hidden Stockpile",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfRevolt("Hidden Stockpile — create a 1/1 colorless Servo artifact creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Servo artifact"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
