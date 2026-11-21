package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Containment Protocol — Enchantment — Aura {2}{U} (TMNT Commander,
// #5011):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap
//	 step."
//
// Claustrophobia's text: its entry trigger (tapEnchantedCreatureOnEntry,
// which reads the creature the Aura last enchanted if it has left in
// response, CR 608.2h) and its untap restriction (CR 502.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "d33c7ea2-4d17-4cdb-aa58-c25c57836297",
		Name:                  "Containment Protocol",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered:             []game.TriggeredAbility{effects.WhenThisEnters("Containment Protocol — tap enchanted creature", effects.TapEnchantedCreatureOnEntry)},
	})
}
