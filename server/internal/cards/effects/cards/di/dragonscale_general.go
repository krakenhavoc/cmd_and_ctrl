package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragonscale General — {3}{W} Creature — Human Warrior 2/3 (Fate Reforged):
//
//	"At the beginning of your end step, bolster X, where X is the number of
//	 tapped creatures you control."
//
// X is counted as the trigger resolves (CR 608.2c), on the layered board; the
// General counts itself if it is tapped. Bolster (CR 701.39a) picks the
// creature you control with the least toughness then, with no target; X of
// zero puts nothing anywhere.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fc6e1db-cf09-4b8c-be4f-bb02ffe7a188",
		Name:         "Dragonscale General",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.ByYou, "Dragonscale General — bolster X, where X is the number of tapped creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					g.RecomputeLayersIfStaleLocked()
					x := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && c.Tapped {
							x++
						}
					}
					if x <= 0 {
						return nil
					}
					return effects.Bolster{N: x}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
