package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Case the Joint — Instant {3}{U} (Streets of New Capenna, #5491):
//
//	"Draw two cards, then look at the top card of each player's library."
//
// Looking marks the caster a knower of each top card and nobody else
// (LookAtTopOfPlayersLibraryForEffect); nothing is moved. Active player
// first, as every table-wide walk is (CR 101.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "575d067a-83a0-49e7-8a74-e3e46a946c8c",
		Name:         "Case the Joint",
		Purpose:      game.Purpose{Draws: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				for _, p := range effects.ApnapPlayers(ctx.Game) {
					ctx.Game.LookAtTopOfPlayersLibraryForEffect(ctx.Controller(), p, 1)
				}
				return nil
			}),
		),
	})
}
