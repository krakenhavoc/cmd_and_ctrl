package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghitu Slinger — Creature — Human Nomad, {2}{R}, 2/2:
//
//	"Echo {2}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, it deals 2 damage to any target."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e2b5f90-e405-4b24-a71a-0988c37cf67a",
		Name:         "Ghitu Slinger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Ghitu Slinger", "{2}{R}"),
			Targeting(WhenThisEnters("Ghitu Slinger — 2 damage to any target", sourceDealsDamageToEachLegalTarget(2)), TargetAny()),
		},
	})
}
