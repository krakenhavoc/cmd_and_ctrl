package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d53ae01b-87e1-48e1-9fc3-23a734ce6622",
		Name:         "No Admittance",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 {
				if err := (DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			return EmpowerJace{N: 1}.Apply(ctx)
		},
	})
}
