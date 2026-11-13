package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bogardan Firefiend — Creature — Elemental Spirit {2}{R}, 2/1 (Weatherlight, #5629):
//
//	"When this creature dies, it deals 2 damage to target creature."
//
// Munitions Enthusiast's dies trigger with a creature target: the
// Firefiend deals the damage as it last existed (CR 608.2h), and the
// target is chosen as the trigger goes on the stack (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d61f3a7-5861-479b-8312-f2d308c35b68",
		Name:         "Bogardan Firefiend",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisDies("Bogardan Firefiend — 2 damage to target creature", sourceDealsDamageToEachLegalTarget(2)),
				TargetCreature("target creature")), ForTargets(DamageToTarget(0, 2))),
		},
	})
}
