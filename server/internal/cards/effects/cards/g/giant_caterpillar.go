package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Caterpillar — {3}{G} Creature — Insect 3/3 (Jumpstart 2022, #5948):
//
//	"{G}, Sacrifice this creature: Create a 1/1 green Insect creature token
//	 with flying named Butterfly at the beginning of the next end step."
//
// The token is made by a delayed trigger (CR 603.7) scheduled as the
// ability resolves, controlled by whoever controlled the ability; the
// token is named Butterfly.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43a515f2-b53a-4bb4-bb6e-0d79c9162359",
		Name:         "Giant Caterpillar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, Sacrifice this creature: Create a 1/1 green Insect creature token with flying named Butterfly at the beginning of the next end step.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ScheduleDelayedTrigger{
					At:    game.StepEnd,
					Label: "Giant Caterpillar — create a 1/1 green Insect creature token with flying named Butterfly",
					Body:  giantCaterpillarButterflyBody,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

var giantCaterpillarButterflyBody = game.SimpleDelayedBody("giant-caterpillar/butterfly", func(g *game.Game, item *game.StackItem) error {
	return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 green Butterfly Insect with flying"), N: 1}.Apply(effects.NewContext(g, item))
})
