package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Settler — Creature — Goblin {3}{R}, 1/1 (Starter 1999, #5620):
//
//	"When this creature enters, destroy target land."
//
// Avalanche Riders' enters trigger without its haste and echo. With no
// land on the battlefield the trigger is removed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04dd449f-ff3a-4ed7-8241-244d58b371dc",
		Name:         "Goblin Settler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Goblin Settler — destroy target land", destroyFirstLegalCardTarget),
				TargetPermanent("target land", Land())),
		},
	})
}
