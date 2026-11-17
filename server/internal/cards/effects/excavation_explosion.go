package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Excavation Explosion — Sorcery {2}{R} (The Brothers' War, #5482):
//
//	"Excavation Explosion deals 3 damage to any target. Create a tapped
//	 Powerstone token."
//
// Lightning Strike's damage, then the Powerstone as a second step (it
// waits for any prompt the damage asks, CR 608.2c). A target that is
// illegal on resolution means nothing resolves, so no token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4bb75182-6dd4-42f7-9541-d484d0b57897",
		Name:         "Excavation Explosion",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				targets := ctx.Targets()
				if len(targets) == 0 {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: targets[0].ID, Amount: 3}.Apply(ctx)
			}),
			CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1},
		),
	})
}
