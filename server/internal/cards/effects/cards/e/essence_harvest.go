package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Essence Harvest — {2}{B} Sorcery: "Target player loses X life and you
// gain X life, where X is the greatest power among creatures you control."
//
// X is read once, as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae9a398e-f3a7-4878-a14d-0f6c2dac3d83",
		Name:         "Essence Harvest",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerLosesLifeAndYouGainThatMuch(ctx.Game, item)
		},
	})
}
