package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hammerheim Deadeye — Creature — Giant Warrior, {3}{R}, 3/3:
//
//	"Echo {5}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, destroy target creature with flying."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6962ccc-9883-48ca-9a5f-aeb68e66fd97",
		Name:         "Hammerheim Deadeye",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Hammerheim Deadeye", "{5}{R}"),
			Targeting(WhenThisEnters("Hammerheim Deadeye — destroy target creature with flying", destroyFirstLegalCardTarget),
				TargetCreature("target creature with flying", HasKeyword("flying"))),
		},
	})
}
