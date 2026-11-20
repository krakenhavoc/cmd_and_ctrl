package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blossoming Wreath — Instant {G} (Weatherlight, #5629):
//
//	"You gain life equal to the number of creature cards in your graveyard."
//
// The creature cards are counted once, as the spell resolves
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbd99a88-8926-455e-8e23-58fed6d14719",
		Name:         "Blossoming Wreath",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Amount: effects.B11CreatureCardsInGraveyard(ctx.Game, ctx.Controller())}.Apply(ctx)
		},
	})
}
