package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saddled Rimestag — Snow Creature — Elk {1}{G}, 2/2 (Modern Horizons):
//
//	"This creature gets +2/+2 as long as you had another creature enter the
//	 battlefield under your control this turn."
//
// A layer 7c self pump (Armory Mice's shape) that is on while the turn's
// creature entry tally says another creature entered under its controller's
// control: the Rimestag's own entry does not count, so when it entered this
// turn a second creature must (BristlebaneOutriderAnotherCreatureEntered). A
// creature that entered under an opponent's control does not count. It ends
// with the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "731898d3-465b-4ecb-a346-b91881714ebc",
		Name:         "Saddled Rimestag",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "This creature gets +2/+2 as long as you had another creature enter the battlefield under your control this turn.",
			Layer: game.Layer7PT, SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfCreature(target, g, source) && effects.BristlebaneOutriderAnotherCreatureEntered(g, source)
			},
			Apply: effects.PumpBy(2, 2),
		}},
	})
}
