package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning Bolt — "Lightning Bolt deals 3 damage to any target."
// Canonical S14 exit-criteria card: cast at a player, pass priority,
// life drops by 3 automatically. The target's kind is discriminated
// in DealDamage.Apply (player vs. battlefield creature).
func init() {
	Register(Spec{
		OracleID:     "4457ed35-7c10-48c8-9776-456485fdf070",
		Name:         "Lightning Bolt",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove}, // #4734: damages a creature, planeswalker or player
		OnResolve:    damageToFirstTarget(3),
	})
}
