package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bury in Books — {4}{U} Instant:
//
//	"This spell costs {2} less to cast if it targets an attacking creature.
//	Put target creature into its owner's library second from the top."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5e67016-1740-42f8-8be2-9eb9d4adb142",
		Name:         "Bury in Books",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets an attacking creature.",
				effects.AttackingCreature()),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			return effects.PutIntoLibrary{Card: t.ID, Depth: 2}.Apply(ctx)
		},
	})
}
