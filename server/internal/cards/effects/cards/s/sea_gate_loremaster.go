package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea Gate Loremaster — Creature — Merfolk Wizard Ally {4}{U}, 1/3
// (Zendikar, #5583):
//
//	"{T}: Draw a card for each Ally you control."
//
// The Allies are counted as the ability resolves (CR 608.2h), the Loremaster
// among them while it is still an Ally on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6eed122b-9760-47fd-8ba2-adeda8054e0d",
		Name:         "Sea Gate Loremaster",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Draw a card for each Ally you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.DrawCards{N: effects.CountYouControl(ctx, effects.HasSubtype("Ally"))}.Apply(ctx)
			},
		}},
	})
}
