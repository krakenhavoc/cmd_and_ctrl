package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moment of Triumph — Instant {W} (Foundations, #4465):
//
//	"Target creature gets +2/+2 until end of turn. You gain 2 life."
//
// Moment of Craving's shape with a pump: the still-legal target gets
// +2/+2 (boostTargetUntilEOT), then the caster gains 2 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57579799-2582-417f-9864-4560cac39ac3",
		Name:         "Moment of Triumph",
		Purpose:      game.Purpose{LifeGain: 2, Targets: game.ForTargets(BenefitUntilEOT(0, 2, 2))},
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := boostTargetUntilEOT("Moment of Triumph — +2/+2", 2, 2)(ctx.Game, item); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
		},
	})
}
