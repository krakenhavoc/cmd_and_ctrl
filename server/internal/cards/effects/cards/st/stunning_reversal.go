package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stunning Reversal — Instant {3}{B} (Battlebond, #5529):
//
//	The next time you would lose the game this turn, instead draw seven cards and your life total becomes 1.
//	Exile Stunning Reversal.
//
// The one-shot replacement is Game.ReplaceNextLossThisTurnForEffect (#6244); the spell exiles itself as it resolves instead of going to the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3db00947-d568-45d7-82cd-b54b4fcb1312",
		Name:         "Stunning Reversal",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx.Game.ReplaceNextLossThisTurnForEffect(ctx.Controller(), item.SourceCardID, 7, 1, "Stunning Reversal — draw seven cards and your life total becomes 1")
			if effects.SourceIsNewObject(ctx.Game, item) {
				return nil
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
