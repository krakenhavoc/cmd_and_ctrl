package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leyline Invocation — Sorcery {5}{G} (Strixhaven: School of Mages, #5502):
//
//	"Create a 0/0 green and blue Fractal creature token. Put X +1/+1
//	 counters on it, where X is the number of lands you control."
//
// Fractal Anomaly's create-then-put shape (SocB8FractalWithCountersSteps).
// X is the number of lands the caster controls as the spell resolves (CR
// 608.2h), read once before the token is made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f2e713f-c7d5-4c91-9934-9e36888244fd",
		Name:         "Leyline Invocation",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(effects.SocB8FractalWithCountersSteps("leyline-invocation/fractal", func(ctx *effects.Context) int {
			return effects.CountYouControl(ctx, effects.Land())
		})...),
	})
}
