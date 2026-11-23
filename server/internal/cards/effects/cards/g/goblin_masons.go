package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cfec9d17-2d11-4b89-8b30-84bd25a38466",
		Name:         "Goblin Masons",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Goblin Masons — destroy target Wall", effects.DestroyFirstLegalTarget),
				effects.TargetPermanent("target Wall", effects.OfCreatureType("Wall")),
			),
		},
	})
}
