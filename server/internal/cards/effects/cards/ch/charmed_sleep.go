package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charmed Sleep — Enchantment — Aura {1}{U}{U} (Bloomburrow, #4467):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap
//	 step."
//
// Claustrophobia's text at a different cost, built from the same parts:
// the entry trigger reads the Aura's attachment as it resolves
// (tapEnchantedCreatureOnEntry, CR 608.2h), and the untap-step hold
// lasts only while the Aura enchants the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "cb33edaf-b604-48b8-8b9c-d6e462a8bebb",
		Name:                  "Charmed Sleep",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered:             []game.TriggeredAbility{effects.WhenThisEnters("Charmed Sleep — tap enchanted creature", effects.TapEnchantedCreatureOnEntry)},
	})
}
