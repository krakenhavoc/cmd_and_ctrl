package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scorching Winds — Instant {R} (Portal, #5630):
//
//	"Cast this spell only during the declare attackers step and only if you've been attacked this step.
//	Scorching Winds deals 1 damage to each attacking creature."
//
// Heavy Fog's cast restriction (youWereAttackedThisStep): the declare
// attackers step, in which a creature was declared attacking you.
// Each creature attacking as it resolves is dealt 1 damage, whoever
// it attacks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:           "db025c94-1182-412e-9665-b5ae89d26616",
		Name:               "Scorching Winds",
		Completeness:       CompletenessFull,
		CastCondition:      youWereAttackedThisStep,
		CastConditionLabel: attackedThisStepLabel,
		// ADR 0126 §6: a creature that isn't attacking is spared.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, AttackingCreature(), 1)
		},
	})
}
