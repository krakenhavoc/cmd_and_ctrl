package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Capture Sphere — {3}{U} Enchantment — Aura, flash: "Enchant creature.
// When this Aura enters, tap enchanted creature. Enchanted creature
// doesn't untap during its controller's untap step."
func init() {
	effects.Register(effects.Spec{
		OracleID:              "2dc59566-10cb-468a-b6a2-17d6c0186e6d",
		Name:                  "Capture Sphere",
		Completeness:          effects.CompletenessFull,
		PrintedKeywords:       []string{"flash"},
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Capture Sphere — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
		},
	})
}
