package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// An-Havva Inn — Sorcery {1}{G}{G} (Homelands, S82 OLD b827, #5635):
//
//	"You gain X plus 1 life, where X is the number of green creatures on the
//	 battlefield."
//
// War Report's count over every player's creatures, read through the layers
// (a creature a layer-5 effect made green counts), plus one. The spell is a
// flat gain of one even when no green creature is out.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3cd72848-7b99-4d46-96f5-53f297c9d2a0",
		Name:         "An-Havva Inn",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := len(effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.OfColor("G"))))
			return effects.GainLife{Amount: x + 1}.Apply(ctx)
		},
	})
}
