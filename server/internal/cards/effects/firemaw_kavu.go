package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Firemaw Kavu — Creature — Kavu, {5}{R}, 4/2:
//
//	"Echo {5}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, it deals 2 damage to target creature.
//	 When this creature leaves the battlefield, it deals 4 damage to target creature."
//
// "Leaves the battlefield" is any zone change from the battlefield, not
// only dying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ac09735-f05c-4e28-83d6-262cddc9eeb3",
		Name:         "Firemaw Kavu",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Firemaw Kavu", "{5}{R}"),
			Targeting(WhenThisEnters("Firemaw Kavu — 2 damage to target creature", sourceDealsDamageToEachLegalTarget(2)),
				TargetCreature("target creature")),
			Targeting(WhenThisLeaves("Firemaw Kavu — 4 damage to target creature", sourceDealsDamageToEachLegalTarget(4)),
				TargetCreature("target creature")),
		},
	})
}
