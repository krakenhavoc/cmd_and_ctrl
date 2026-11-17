package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Runic Shot — {W} Sorcery (Dominaria United):
//
//	"Kicker {U} (You may pay an additional {U} as you cast this spell.)
//	 Destroy target tapped creature. If this spell was kicked, scry 2."
//
// "Tapped" is rechecked on resolution (CR 608.2b): a creature that
// untapped in response is an illegal target, and the spell does nothing.
// The scry follows the destruction (ResolveSteps).
func init() {
	Register(Spec{
		OracleID:      "aa597788-4149-48b8-9f43-8c0de063eb68",
		Name:          "Runic Shot",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove, Kicked: &game.Purpose{Scry: 2}},
		Targets:       TargetCreature("target tapped creature", tappedPermanent()),
		OptionalCosts: []game.AdditionalCost{Kicker("{U}")},
		OnResolve: ResolveSteps(
			Instruction(kineticOozeDestroy),
			Instruction(func(ctx *Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
			}),
		),
	})
}
