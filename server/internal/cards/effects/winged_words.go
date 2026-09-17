package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winged Words — Sorcery {2}{U}:
//
//	"This spell costs {1} less to cast if you control a creature with
//	 flying.
//	 Draw two cards."
//
// #746: a conditional self cost modifier.
func init() {
	Register(Spec{
		OracleID:     "c623aeb1-e6d4-48fe-bd2a-a7a6729aa4df",
		Name:         "Winged Words",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you control a creature with flying.",
				YouControlA(And(Creature(), HasKeyword("flying")))),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
