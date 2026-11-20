package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Perilous Myr — Artifact Creature — Phyrexian Myr {2}, 1/1 (Scars of
// Mirrodin, #5577):
//
//	"When this creature dies, it deals 2 damage to any target."
//
// Shady Informant's dies trigger. The Myr is the source of the damage as
// it last existed on the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1e8f746-1837-4ea8-9a07-e57946883353",
		Name:         "Perilous Myr",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisDies("Perilous Myr — 2 damage to any target",
				sourceDealsDamageToEachLegalTarget(2)), TargetAny()), ForTargets(DamageToTarget(0, 2))),
		},
	})
}
