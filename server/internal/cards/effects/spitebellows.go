package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spitebellows — Creature — Elemental {5}{R}, 6/1 (Commander Masters):
//
//	"When this creature leaves the battlefield, it deals 6 damage to target
//	 creature.
//	 Evoke {1}{R}{R}"
//
// Firemaw Kavu's leave-the-battlefield damage, at six, with Shriekmaw's
// evoke: an evoked Spitebellows is sacrificed as it enters, which is a
// leave, so the trigger fires (CR 702.74a). The damage source is the
// Elemental as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fee18f57-4ae5-4df8-af3b-f82ed2157ac3",
		Name:         "Spitebellows",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Evoke("{1}{R}{R}"),
		},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisLeaves("Spitebellows — 6 damage to target creature", sourceDealsDamageToEachLegalTarget(6)),
				TargetCreature("target creature")), ForTargets(DamageToTarget(0, 6))),
		},
	})
}
