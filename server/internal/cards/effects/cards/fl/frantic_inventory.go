package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frantic Inventory — Instant {1}{U} (Core Set 2021, #5511):
//
//	"Draw a card, then draw cards equal to the number of cards named Frantic
//	 Inventory in your graveyard."
//
// The count is taken after the first draw, as the sentence orders it. The
// spell itself is still on the stack as it resolves, so it is not counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7865331-4c53-46db-b2d7-7e5cb21b6b9d",
		Name:         "Frantic Inventory",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 1},
			effects.Instruction(func(ctx *effects.Context) error {
				p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
				if p == nil {
					return nil
				}
				n := 0
				for _, c := range p.Graveyard.Cards {
					if c.Name == "Frantic Inventory" {
						n++
					}
				}
				return effects.DrawCards{N: n}.Apply(ctx)
			}),
		),
	})
}
