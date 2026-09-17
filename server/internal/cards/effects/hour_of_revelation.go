package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hour of Revelation — Sorcery {3}{W}{W}{W}:
//
//	"This spell costs {3} less to cast if there are ten or more nonland
//	 permanents on the battlefield.
//	 Destroy all nonland permanents."
//
// #746: a conditional self cost modifier. The reduction spends
// generic mana only, so the best it can do is {W}{W}{W}.
func init() {
	Register(Spec{
		OracleID:     "920c1dd4-c3f6-4020-a6f0-2e8acad2c212",
		Name:         "Hour of Revelation",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(3, "This spell costs {3} less to cast if there are ten or more nonland permanents on the battlefield.",
				PermanentsOnBattlefieldAtLeast(10, Nonland())),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Nonland()}.Apply(ctx)
		},
	})
}
