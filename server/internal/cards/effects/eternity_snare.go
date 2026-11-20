package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eternity Snare — Enchantment — Aura {5}{U} (Time Spiral, #5596):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Abundant Growth's enters draw and Bitter Chill's untap restriction on
// the enchanted creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "b9bd1833-d4f9-46c0-bb94-331c456193fc",
		Name:                  "Eternity Snare",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm, Draws: 1},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Eternity Snare — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
