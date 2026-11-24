package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thran Temporal Gateway — Legendary Artifact {4} (Secret Lair Drop):
//
//	"{4}, {T}: You may put a historic permanent card from your hand onto the
//	 battlefield. (Artifacts, legendaries, and Sagas are historic.)"
//
// PutFromHandOntoBattlefield over the historic test (Weatherlight's), optional:
// the hand is offered only its permanent cards that are artifacts, legendary
// or Sagas, read through the layers (CR 700.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36282c29-a44b-4156-9441-03aa4bf386a8",
		Name:         "Thran Temporal Gateway",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: You may put a historic permanent card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutFromHandOntoBattlefield{
					Match:    effects.B14Historic(),
					Optional: true,
					Label:    "Thran Temporal Gateway — you may put a historic permanent card from your hand onto the battlefield",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
