package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burrenton Shield-Bearers — Creature — Kithkin Soldier {4}{W}, 3/3
// (Morningtide, #5592):
//
//	"Whenever this creature attacks, target creature gets +0/+3 until end
//	 of turn."
//
// Mayhem Patrol's attack trigger with a +0/+3 body (boostTargetUntilEOT):
// any creature, the attacker itself included, and a target gone by
// resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d038c81-b43f-4721-abda-5dacfc194ed3",
		Name:         "Burrenton Shield-Bearers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WheneverThisAttacks("Burrenton Shield-Bearers — target creature gets +0/+3",
				boostTargetUntilEOT("Burrenton Shield-Bearers — +0/+3", 0, 3)), TargetCreature("target creature")),
				ForTargets(BenefitUntilEOT(0, 0, 3))),
		},
	})
}
