package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tolarian Geyser — {2}{U} Sorcery (Dominaria United):
//
//	"Kicker {W} (You may pay an additional {W} as you cast this spell.)
//	 Return target creature to its owner's hand. Draw a card. If this
//	 spell was kicked, you gain 3 life."
//
// Three steps (ResolveSteps). The draw and the life do not depend on the
// bounce having worked; a spell whose only target is gone does not
// resolve at all (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2240ce8-0bd6-4c4f-a73e-fb0a2af5bfac",
		Name:         "Tolarian Geyser",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Draws: 1, Answers: game.AnswerRemove,
			Kicked: &game.Purpose{LifeGain: 3}},
		Targets:       effects.TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{W}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BounceToHand{Target: t.ID}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.GainLife{Amount: 3}.Apply(ctx)
			}),
		),
	})
}
