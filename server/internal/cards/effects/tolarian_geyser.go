package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d2240ce8-0bd6-4c4f-a73e-fb0a2af5bfac",
		Name:         "Tolarian Geyser",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Draws: 1, Answers: game.AnswerRemove,
			Kicked: &game.Purpose{LifeGain: 3}},
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{Kicker("{W}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return BounceToHand{Target: t.ID}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return GainLife{Amount: 3}.Apply(ctx)
			}),
		),
	})
}
