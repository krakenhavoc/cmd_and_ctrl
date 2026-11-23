package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hour of Revelation — Sorcery {3}{W}{W}{W}:
//
//	"This spell costs {3} less to cast if there are ten or more nonland
//	 permanents on the battlefield.
//	 Destroy all nonland permanents."
//
// #746: a conditional self cost modifier. The reduction spends
// generic mana only, so the best it can do is {W}{W}{W}.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "920c1dd4-c3f6-4020-a6f0-2e8acad2c212",
		Name:         "Hour of Revelation",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy}},
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(3, "This spell costs {3} less to cast if there are ten or more nonland permanents on the battlefield.",
				effects.PermanentsOnBattlefieldAtLeast(10, effects.Nonland())),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Nonland()}.Apply(ctx)
		},
	})
}
