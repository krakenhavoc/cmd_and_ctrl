package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sage of Lat-Nam — Creature — Human Artificer {1}{U}, 1/2
// (Antiquities, #5640):
//
//	"{T}, Sacrifice an artifact: Draw a card."
//
// Trading Post's last row without its {1}: a tap and the sacrifice of
// one artifact the controller picks at announce, both paid before the
// ability goes on the stack, then a draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f34be3cc-ff47-4415-a6a8-ed142891dc0c",
		Name:         "Sage of Lat-Nam",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
