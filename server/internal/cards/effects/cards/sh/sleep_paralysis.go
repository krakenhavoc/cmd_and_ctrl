package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sleep Paralysis — Enchantment — Aura {3}{U}
// (Shadows over Innistrad, #5543):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Claustrophobia's text exactly: the enters trigger taps the enchanted
// creature (tapEnchantedCreatureOnEntry) and an untap-step restriction
// holds it (enchantedDoesntUntap).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "5a8bd4a4-8b10-4eef-8341-28be2afa2c00",
		Name:                  "Sleep Paralysis",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered:             []game.TriggeredAbility{effects.WhenThisEnters("Sleep Paralysis — tap enchanted creature", effects.TapEnchantedCreatureOnEntry)},
	})
}
