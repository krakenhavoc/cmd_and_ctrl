package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Capture Sphere — {3}{U} Enchantment — Aura, flash: "Enchant creature.
// When this Aura enters, tap enchanted creature. Enchanted creature
// doesn't untap during its controller's untap step."
func init() {
	Register(Spec{
		OracleID:              "2dc59566-10cb-468a-b6a2-17d6c0186e6d",
		Name:                  "Capture Sphere",
		Completeness:          CompletenessFull,
		PrintedKeywords:       []string{"flash"},
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Capture Sphere — tap enchanted creature", tapEnchantedCreatureOnEntry),
		},
	})
}
