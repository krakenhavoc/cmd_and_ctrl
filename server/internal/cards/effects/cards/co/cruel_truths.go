package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cruel Truths — Instant {3}{B} (Tarkir: Dragonstorm, #4463):
//
//	"Surveil 2, then draw two cards. You lose 2 life."
//
// Risky Research's text, so its resolution: the surveil prompt carries
// the keyed tail (riskyResearchThen), the cards drawn come from the
// library the player arranged and the life loss follows the draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e68e3f7-b168-4541-b8a0-791a8a6ebde8",
		Name:         "Cruel Truths",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Surveil{Player: item.Controller, N: 2, ThenTail: ctx.Tail(effects.RiskyResearchThen)}.Apply(ctx)
		},
	})
}
