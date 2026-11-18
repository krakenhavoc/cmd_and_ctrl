package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyre Spawn — Creature — Elemental {4}{R}{R}, 6/4 (Innistrad: Crimson
// Vow):
//
//	"When this creature dies, it deals 3 damage to any target."
//
// Dynamite Diver's dies trigger for 3: dealt by the Spawn as it last
// existed (CR 608.2h), target chosen as the trigger goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13b0a3f3-04a5-465f-898b-13a28f795a16",
		Name:         "Pyre Spawn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisDies("Pyre Spawn — 3 damage to any target", sourceDealsDamageToEachLegalTarget(3)),
				TargetAny()), ForTargets(DamageToTarget(0, 3))),
		},
	})
}
