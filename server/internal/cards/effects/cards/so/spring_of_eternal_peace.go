package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spring of Eternal Peace — Sorcery {3}{G}{G} (Portal Three Kingdoms, #5622):
//
//	"You gain 8 life."
//
// One life-gain step for its controller, as Sacred Nectar's at 8.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6fcdb870-a1b6-49fe-8b5c-cdde074a7f7d",
		Name:         "Spring of Eternal Peace",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.GainLife{Amount: 8}),
	})
}
