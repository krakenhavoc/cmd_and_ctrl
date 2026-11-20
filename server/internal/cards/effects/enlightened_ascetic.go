package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enlightened Ascetic — Creature — Cat Monk {1}{W}, 1/1 (Magic Origins, #5549):
//
//	"When this creature enters, you may destroy target enchantment."
//
// Conclave Naturalists' entry, limited to an enchantment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8fe257e5-b54a-4e56-8040-d2eaa0bb0a6a",
		Name:         "Enlightened Ascetic",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WhenThisEnters("Enlightened Ascetic — destroy target enchantment", destroyChosenPermanent),
				TargetPermanent("target enchantment", Enchantment())),
				"Enlightened Ascetic — destroy an enchantment?"),
		},
	})
}
