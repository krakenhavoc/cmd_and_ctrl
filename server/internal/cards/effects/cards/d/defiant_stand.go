package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defiant Stand — Instant {1}{W} (Portal, S82 OLD batch 784, #5630):
//
//	"Cast this spell only during the declare attackers step and only if you've
//	 been attacked this step.
//	 Target creature gets +1/+3 until end of turn. Untap that creature."
//
// Rally the Troops' cast restriction over one creature: the pump, then the
// untap as the next instruction (CR 608.2c). Nothing happens when the target is
// illegal on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "05e61628-cdae-4664-b866-be438fb3a6ba",
		Name:               "Defiant Stand",
		Completeness:       effects.CompletenessFull,
		CastCondition:      effects.YouWereAttackedThisStep,
		CastConditionLabel: effects.AttackedThisStepLabel,
		Targets:            effects.TargetCreature("target creature"),
		Purpose:            effects.ForTargets(effects.BenefitUntilEOT(0, 1, 3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.BoostTheTargetUntilEOT(1, 3, "Defiant Stand — +1/+3 until end of turn")(item, ctx); err != nil {
				return err
			}
			return effects.UntapTheTarget(ctx.Game, item)
		},
	})
}
