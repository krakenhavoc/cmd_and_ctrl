package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Needletooth Raptor — Creature — Dinosaur {3}{R}, 2/2 (Rivals of Ixalan,
// #5531):
//
//	"Enrage — Whenever this creature is dealt damage, it deals 5 damage to
//	 target creature an opponent controls."
//
// The shared enrage trigger with a target creature an opponent controls,
// picked as the trigger goes on the stack (CR 603.3d) and re-checked as it
// resolves (CR 608.2b). The Raptor is the source, even when the damage
// killed it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c431df9-095a-44f1-993e-74eb215d75fe",
		Name:         "Needletooth Raptor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(cmdr2B08EnrageTrigger("Needletooth Raptor — Enrage: 5 damage to target creature an opponent controls",
				DealDamageToTheTarget(5)), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
