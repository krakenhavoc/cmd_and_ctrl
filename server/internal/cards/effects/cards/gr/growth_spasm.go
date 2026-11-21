package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Growth Spasm — Sorcery {2}{G}:
//
//	"Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle. Create a 0/1 colorless Eldrazi
//	 Spawn creature token. It has "Sacrifice this token: Add {C}.""
//
// Rampant Growth's search, then the Spawn Awakening Zone makes; the
// token is created even when the search finds nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c022656-2b97-46c5-8f7b-72a71996457a",
		Name:         "Growth Spasm",
		Purpose:      game.Purpose{Lands: 1, Tokens: 1, CreatureTokens: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.RampantGrowthSearch()(ctx.Item, ctx)
			}),
			effects.CreateToken{Template: effects.EldraziSpawnToken(), N: 1},
		),
	})
}
