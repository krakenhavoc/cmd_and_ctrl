package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Take Inventory — Sorcery {1}{U} (Eldritch Moon, S82 OLD, #5541):
//
//	"Draw a card, then draw cards equal to the number of cards named Take
//	 Inventory in your graveyard."
//
// The graveyard is counted after the first draw (the "then"), as the spell
// resolves, so the spell itself, still on the stack, is not counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "925b3d1b-98cd-44f5-8aa0-e2b17df2808a",
		Name:         "Take Inventory",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 1},
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: effects.S82OldB276NamedInGraveyard(ctx.Game, ctx.Controller(), "Take Inventory")}.Apply(ctx)
			}),
		),
	})
}
