package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waterknot — Enchantment — Aura {1}{U}{U} (Rivals of Ixalan, #5531):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Claustrophobia's text: the entry trigger taps the creature the Aura
// enchants as it resolves, and the untap-step restriction holds it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "002100cd-1ab1-44da-93a0-1269d47a712a",
		Name:                  "Waterknot",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered:             []game.TriggeredAbility{effects.WhenThisEnters("Waterknot — tap enchanted creature", effects.TapEnchantedCreatureOnEntry)},
	})
}
