package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Settler — Creature — Goblin {3}{R}, 1/1 (Starter 1999, #5620):
//
//	"When this creature enters, destroy target land."
//
// Avalanche Riders' enters trigger without its haste and echo. With no
// land on the battlefield the trigger is removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04dd449f-ff3a-4ed7-8241-244d58b371dc",
		Name:         "Goblin Settler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Goblin Settler — destroy target land", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target land", effects.Land())),
		},
	})
}
