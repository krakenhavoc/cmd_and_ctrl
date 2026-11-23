package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Snake — Creature — Snake {4}{R}, 3/1 (Portal, #5630):
//
//	"When this creature dies, destroy target land."
//
// Goblin Gardener's dies trigger, word for word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e96542ed-1931-4da1-9d9e-d10878c4ae6b",
		Name:         "Fire Snake",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Fire Snake — destroy target land", effects.DestroyFirstLegalTarget),
				effects.TargetPermanent("target land", effects.Land()),
			),
		},
	})
}
