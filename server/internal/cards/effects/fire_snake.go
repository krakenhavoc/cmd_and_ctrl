package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire Snake — Creature — Snake {4}{R}, 3/1 (Portal, #5630):
//
//	"When this creature dies, destroy target land."
//
// Goblin Gardener's dies trigger, word for word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e96542ed-1931-4da1-9d9e-d10878c4ae6b",
		Name:         "Fire Snake",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisDies("Fire Snake — destroy target land", destroyFirstLegalTarget),
				TargetPermanent("target land", Land()),
			),
		},
	})
}
