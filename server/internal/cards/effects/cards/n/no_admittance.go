package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// No Admittance — Sorcery {1}{R} (Reality Fracture):
//
//	"No Admittance deals 3 damage to any target.
//	 Empower Jace 1."
//
// ADR 0139 proof card: the keyword action after an ordinary targeted
// effect. The Empower Jace half does not depend on the target, so a
// target that left in response still empowers Jace (CR 608.2b stops
// only the part that needs the illegal target; the spell still has a
// legal target if it had one, and with none at all it does not
// resolve, CR 608.2b's first sentence, which the engine applies).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d53ae01b-87e1-48e1-9fc3-23a734ce6622",
		Name:         "No Admittance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		// Empower Jace waits for a prompt the damage asks (a CR 616 order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if len(ctx.Item.Targets) == 0 {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: ctx.Item.Targets[0].ID, Amount: 3}.Apply(ctx)
			}),
			effects.EmpowerJace{N: 1},
		),
	})
}
