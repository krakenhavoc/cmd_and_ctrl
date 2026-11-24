package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serpentine Curve — Sorcery {3}{U} (Strixhaven: School of Mages, #5502):
//
//	"Create a 0/0 green and blue Fractal creature token. Put X +1/+1
//	 counters on it, where X is one plus the total number of instant and
//	 sorcery cards you own in exile and in your graveyard."
//
// X is read as the spell resolves (CR 608.2h), so Serpentine Curve itself,
// still on the stack, is not counted. Face-down exiled cards have no types
// (CR 708.2) and do not count; a card you own that an opponent exiled does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e07ed462-17b8-46ec-8a3f-3f5c89f4bf78",
		Name:         "Serpentine Curve",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(effects.SocB8FractalWithCountersSteps("serpentine-curve/fractal", func(ctx *effects.Context) int {
			return 1 + effects.OldB41InstantsAndSorceriesOwnedInExileAndGraveyard(ctx.Game, ctx.Controller())
		})...),
	})
}
