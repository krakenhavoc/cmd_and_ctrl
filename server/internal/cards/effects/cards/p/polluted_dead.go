package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Polluted Dead — {4}{B} Creature — Zombie 3/3 (Avacyn Restored, #5570):
//
//	"When this creature dies, destroy target land."
//
// Goblin Gardener's dies trigger: destroy target land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ffca600-5286-4086-9efe-c398babeb6ae",
		Name:         "Polluted Dead",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Polluted Dead — destroy target land", effects.DestroyFirstLegalTarget),
				effects.TargetPermanent("target land", effects.Land()),
			),
		},
	})
}
