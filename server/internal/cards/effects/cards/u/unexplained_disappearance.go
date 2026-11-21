package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unexplained Disappearance — Instant {1}{U} (Guilds of Ravnica, #5525):
//
//	"Return target creature to its owner's hand.
//	 Surveil 1."
//
// Two steps: the shared bounceTheTarget, then surveil 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18d16129-da25-4120-a343-64ce962abba2",
		Name:         "Unexplained Disappearance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.BounceTheTarget(nil, ctx) }),
			effects.Surveil{N: 1},
		),
	})
}
