package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bury in Books — {4}{U} Instant:
//
//	"This spell costs {2} less to cast if it targets an attacking creature.
//	Put target creature into its owner's library second from the top."
func init() {
	Register(Spec{
		OracleID:     "d5e67016-1740-42f8-8be2-9eb9d4adb142",
		Name:         "Bury in Books",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets an attacking creature.",
				AttackingCreature()),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			return PutIntoLibrary{Card: t.ID, Depth: 2}.Apply(ctx)
		},
	})
}
