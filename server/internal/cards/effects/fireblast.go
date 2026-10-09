package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fireblast — Instant {4}{R}{R}:
//
//	"You may sacrifice two Mountains rather than pay this spell's mana
//	 cost.
//	 Fireblast deals 4 damage to any target."
//
// A #1727 proof card: the sacrifice alternative cost paid from HAND,
// with no zone binding, which is the other half of the component
// Dread Return's flashback uses. The two Mountains are sacrificed with
// Fireblast already on the stack, so a countered Fireblast still costs
// them.
func init() {
	Register(Spec{
		OracleID:     "9dd7f27a-e862-47d7-9158-034cf4d353b8",
		Name:         "Fireblast",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		AlternativeCosts: []game.AlternativeCost{
			SacrificeInstead(2, "two Mountains", 0, HasSubtype("Mountain")),
		},
		OnResolve: damageToFirstTarget(4),
	})
}
