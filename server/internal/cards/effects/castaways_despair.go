package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Castaway's Despair — Enchantment — Aura {3}{U} (Ixalan, #5532):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Claustrophobia's text: the Aura taps its creature as it enters (a
// trigger, CR 603.6a), and the untap-step restriction holds it down for as
// long as the Aura is attached (CR 502.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "9dc4b172-da33-40c0-ab29-d68807dbe93f",
		Name:                  "Castaway's Despair",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered:             []game.TriggeredAbility{WhenThisEnters("Castaway's Despair — tap enchanted creature", tapEnchantedCreatureOnEntry)},
	})
}
