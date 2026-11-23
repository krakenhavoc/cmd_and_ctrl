package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Winged Words — Sorcery {2}{U}:
//
//	"This spell costs {1} less to cast if you control a creature with
//	 flying.
//	 Draw two cards."
//
// #746: a conditional self cost modifier.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c623aeb1-e6d4-48fe-bd2a-a7a6729aa4df",
		Name:         "Winged Words",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control a creature with flying.",
				effects.YouControlA(effects.And(effects.Creature(), effects.HasKeyword("flying")))),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
