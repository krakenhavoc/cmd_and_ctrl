package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Body of Research — Sorcery {G}{G}{G}{U}{U}{U} (Strixhaven: School of Mages,
// #5502):
//
//	"Create a 0/0 green and blue Fractal creature token. Put X +1/+1
//	 counters on it, where X is the number of cards in your library."
//
// X is the size of the caster's library as the spell resolves (CR 608.2h),
// read once before the token is made. An empty library makes a 0/0 that dies
// to CR 704.5f.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0c21b32-f672-4496-91ed-7b7f351eb539",
		Name:         "Body of Research",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(effects.SocB8FractalWithCountersSteps("body-of-research/fractal", func(ctx *effects.Context) int {
			if p := ctx.PlayerByID(ctx.Controller()); p != nil && p.Library != nil {
				return p.Library.Size()
			}
			return 0
		})...),
	})
}
