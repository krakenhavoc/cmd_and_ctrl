package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desynchronize — Instant for {4}{U} (The Brothers' War):
//
//	"Target nonland permanent's owner puts it on their choice of the top or
//	bottom of their library. Scry 2."
//
// The owner chooses the lane; the spell's controller scries afterwards, once
// the owner's prompt has been answered (CR 608.2c).
func init() {
	Register(Spec{
		OracleID:     "dae81b1b-28dd-4f7f-9bcb-65e72eaa37e2",
		Name:         "Desynchronize",
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 2},
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ownerTucksFirstTargetTopOrBottom(ctx, "Desynchronize — put it on your choice of the top or bottom of your library")
			}),
			Instruction(func(ctx *Context) error {
				return Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
			}),
		),
	})
}
