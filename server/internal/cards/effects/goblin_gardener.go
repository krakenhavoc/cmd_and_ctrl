package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Gardener — Creature — Goblin {3}{R}, 2/1 (Urza's Destiny,
// #5621):
//
//	"When this creature dies, destroy target land."
//
// Reliquary Monk's targeted dies trigger aimed at a land, any player's.
// The trigger is mandatory, so with only your own lands on the
// battlefield one of them is the target. A land that has left by
// resolution is not destroyed and nothing else happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "29781b19-0917-4d8f-abe3-df0b7e96d8a6",
		Name:         "Goblin Gardener",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisDies("Goblin Gardener — destroy target land", destroyFirstLegalTarget),
				TargetPermanent("target land", Land()),
			),
		},
	})
}
