package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enlightened Ascetic — Creature — Cat Monk {1}{W}, 1/1 (Magic Origins, #5549):
//
//	"When this creature enters, you may destroy target enchantment."
//
// Conclave Naturalists' entry, limited to an enchantment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fe257e5-b54a-4e56-8040-d2eaa0bb0a6a",
		Name:         "Enlightened Ascetic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Enlightened Ascetic — destroy target enchantment", effects.DestroyChosenPermanent),
				effects.TargetPermanent("target enchantment", effects.Enchantment())),
				"Enlightened Ascetic — destroy an enchantment?"),
		},
	})
}
