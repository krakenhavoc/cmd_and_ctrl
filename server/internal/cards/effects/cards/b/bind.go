package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bind — Instant {1}{G} (Invasion, S82 #5616):
//
//	"Counter target activated ability. (Mana abilities can't be targeted.)
//	 Draw a card."
//
// Stifle's clause narrowed to activated abilities: a triggered ability is not a
// legal target. Mana abilities never use the stack, so the parenthetical needs
// no code (CR 605.3b). If the ability has already left the stack the spell does
// nothing, and so does not draw (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b9557df-7158-4ada-be77-5c346851a568",
		Name:         "Bind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets: effects.AbilityOnStack("target activated ability", func(_ *game.Game, _ uuid.UUID, item *game.StackItem) bool {
			return item.Kind == game.StackItemActivated
		}),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if _, ok := ctx.ClauseTarget(0); !ok {
				return nil
			}
			if err := effects.CounterTheTargetSpell(item, ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
