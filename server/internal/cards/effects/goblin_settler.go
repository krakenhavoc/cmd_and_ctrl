package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Settler — Creature — Goblin {3}{R}, 1/1 (Mystery Booster 2):
//
//	"When this creature enters, destroy target land."
//
// Goblin Gardener's targeted trigger on entry: mandatory, any player's land
// (so with only your own lands out, one of them is the target, and the Settler
// may be aimed at a land that is itself). A land that has left by resolution
// is not destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04dd449f-ff3a-4ed7-8241-244d58b371dc",
		Name:         "Goblin Settler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Goblin Settler — destroy target land", destroyFirstLegalTarget),
				TargetPermanent("target land", Land()),
			),
		},
	})
}
