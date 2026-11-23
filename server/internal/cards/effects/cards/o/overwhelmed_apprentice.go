package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Overwhelmed Apprentice — Creature — Human Wizard {U}, 1/2 (Jumpstart 2022, #5948):
//
//	"When this creature enters, each opponent mills two cards. Then you
//	 scry 2."
//
// Each opponent mills in turn order, one prompt at a time; the scry is the
// mill's registered tail, so it waits for the last library (Aether Syphon's
// mill, Fading Hope's "then scry").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e68d6459-2f32-4320-a062-6d33f85348f9",
		Name:         "Overwhelmed Apprentice",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Overwhelmed Apprentice — each opponent mills two cards, then you scry 2",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillEachPlayerInTurn(ctx, ctx.Opponents(), effects.ConstantMill(2), ctx.TailN(effects.S82MidThenScry, 2))
				}),
		},
	})
}
