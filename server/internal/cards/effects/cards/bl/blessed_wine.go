package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blessed Wine — Instant {1}{W} (Ice Age, S82 OLD, #5636):
//
//	"You gain 1 life.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8079007c-a07a-4d71-bb08-e1dd215297de",
		Name:         "Blessed Wine",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GainLife{Amount: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Blessed Wine — draw a card").Apply(ctx)
		},
	})
}
