package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fractal Summoning — Sorcery — Lesson {X}{G/U}{G/U} (Strixhaven: School of
// Mages, #5502):
//
//	"Create a 0/0 green and blue Fractal creature token. Put X +1/+1
//	 counters on it."
//
// Fractal Anomaly's create-then-put shape with the X announced as the spell
// was cast (CR 107.3f). At X = 0 the Fractal is a 0/0 and dies to CR 704.5f,
// so the spell leaves nothing behind and declares XMatters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce690421-db1d-4177-a2db-338f29aab623",
		Name:         "Fractal Summoning",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: effects.ResolveSteps(effects.SocB8FractalWithCountersSteps("fractal-summoning/fractal", func(ctx *effects.Context) int {
			return ctx.X()
		})...),
	})
}
