package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desynchronize — Instant for {4}{U} (The Brothers' War):
//
//	"Target nonland permanent's owner puts it on their choice of the top or
//	bottom of their library. Scry 2."
//
// The owner chooses the lane; the spell's controller scries afterwards, once
// the owner's prompt has been answered (CR 608.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dae81b1b-28dd-4f7f-9bcb-65e72eaa37e2",
		Name:         "Desynchronize",
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 2},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.OwnerTucksFirstTargetTopOrBottom(ctx, "Desynchronize — put it on your choice of the top or bottom of your library")
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
			}),
		),
	})
}
