package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spitebellows — Creature — Elemental {5}{R}, 6/1 (Morningtide, #5592):
//
//	"When this creature leaves the battlefield, it deals 6 damage to
//	 target creature.
//	 Evoke {1}{R}{R} (You may cast this spell for its evoke cost. If you
//	 do, it's sacrificed when it enters.)"
//
// Reveillark's leave trigger (any exit, not only dying) with Bogardan
// Firefiend's damage body: the departed creature is the source, read from
// its last-known information (CR 608.2h). Evoke is Mulldrifter's
// alternative cost (CR 702.74a): the evoked creature enters, is
// sacrificed, and the leave trigger fires.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "fee18f57-4ae5-4df8-af3b-f82ed2157ac3",
		Name:             "Spitebellows",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Evoke("{1}{R}{R}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisLeaves("Spitebellows — 6 damage to target creature", sourceDealsDamageToEachLegalTarget(6)),
				TargetCreature("target creature")), ForTargets(DamageToTarget(0, 6))),
		},
	})
}
