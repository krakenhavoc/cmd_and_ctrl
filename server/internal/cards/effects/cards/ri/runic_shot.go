package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runic Shot — {W} Sorcery (Dominaria United):
//
//	"Kicker {U} (You may pay an additional {U} as you cast this spell.)
//	 Destroy target tapped creature. If this spell was kicked, scry 2."
//
// "Tapped" is rechecked on resolution (CR 608.2b): a creature that
// untapped in response is an illegal target, and the spell does nothing.
// The scry follows the destruction (ResolveSteps).
func init() {
	effects.Register(effects.Spec{
		OracleID:      "aa597788-4149-48b8-9f43-8c0de063eb68",
		Name:          "Runic Shot",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove, Kicked: &game.Purpose{Scry: 2}},
		Targets:       effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{U}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.KineticOozeDestroy),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
			}),
		),
	})
}
