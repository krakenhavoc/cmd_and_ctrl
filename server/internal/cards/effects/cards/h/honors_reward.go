package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Honor's Reward — Instant {2}{W} (Fate Reforged,
// #5552):
//
//	"You gain 4 life. Bolster 2. (Choose a creature with the least
//	 toughness among creatures you control and put two +1/+1 counters on
//	 it.)"
//
// The life, then bolster (CR 701.39a), in printed order (CR 608.2c): the
// creature is chosen as the spell resolves, with no target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87e6aabb-41f5-41d2-b662-cd09ad61a1c8",
		Name:         "Honor's Reward",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 4},
		OnResolve:    effects.ResolveSteps(effects.GainLife{Amount: 4}, effects.Bolster{N: 2}),
	})
}
