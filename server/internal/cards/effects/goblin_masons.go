package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Masons — Creature — Goblin {1}{R}, 2/1 (Urza's Destiny, #5621):
//
//	"When this creature dies, destroy target Wall."
//
// Goblin Gardener's dies trigger aimed at a Wall: a creature with the
// Wall subtype, any player's. With no Wall on the battlefield the
// trigger has no legal target and is removed (CR 603.3d). A Wall gone
// by resolution is not destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfec9d17-2d11-4b89-8b30-84bd25a38466",
		Name:         "Goblin Masons",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisDies("Goblin Masons — destroy target Wall", destroyFirstLegalTarget),
				TargetPermanent("target Wall", OfCreatureType("Wall")),
			),
		},
	})
}
