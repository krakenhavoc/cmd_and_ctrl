package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angel's Mercy — Instant {2}{W}{W} (Magic 2010, #5585):
//
//	"You gain 7 life."
//
// Sacred Nectar's one instruction with 7 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b232bb7-d372-4174-a049-5f8d620810e6",
		Name:         "Angel's Mercy",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.GainLife{Amount: 7}),
	})
}
