package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:              "d33c7ea2-4d17-4cdb-aa58-c25c57836297",
		Name:                  "Containment Protocol",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered:             []game.TriggeredAbility{WhenThisEnters("Containment Protocol — tap enchanted creature", tapEnchantedCreatureOnEntry)},
	})
}
