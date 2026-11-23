package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Courier's Capsule — Artifact {1}{U} (Shards of Alara, #5588):
//
//	"{1}{U}, {T}, Sacrifice this artifact: Draw two cards."
//
// Hedron Archive's draw on a cheaper Capsule: the mana, the tap and the
// sacrifice are all costs (CR 602.2b), and the draw resolves for the
// activator.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e694613-e2e1-4ae3-baa8-f372259e91cf",
		Name:         "Courier's Capsule",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
