package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ray of Erasure — Instant {U} (Ice Age, S82 OLD, #5636):
//
//	"Target player mills a card.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29d828fa-a813-48f6-8dcb-43f6047b88a3",
		Name:         "Ray of Erasure",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if p := effects.TargetedPlayerOf(ctx); p != uuid.Nil {
				if err := (effects.MillCards{Player: p, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawACardAtTheNextUpkeep("Ray of Erasure — draw a card").Apply(ctx)
		},
	})
}
