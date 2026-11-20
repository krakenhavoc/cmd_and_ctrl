package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Polluted Dead — {4}{B} Creature — Zombie 3/3 (Avacyn Restored, #5570):
//
//	"When this creature dies, destroy target land."
//
// Goblin Gardener's dies trigger: destroy target land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ffca600-5286-4086-9efe-c398babeb6ae",
		Name:         "Polluted Dead",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisDies("Polluted Dead — destroy target land", destroyFirstLegalTarget),
				TargetPermanent("target land", Land()),
			),
		},
	})
}
