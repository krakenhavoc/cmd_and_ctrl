package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Due Respect — Instant {1}{W}:
//
//	"Permanents enter tapped this turn.
//	 Draw a card."
//
// Every permanent, whoever controls it, enters tapped until cleanup.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c367fa2-844a-4083-9484-c6def1dee8a8",
		Name:         "Due Respect",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ctx.Game.PermanentsEnterTappedThisTurnForEffect(ctx.Source(), uuid.Nil, nil, "Due Respect — permanents enter tapped this turn")
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
