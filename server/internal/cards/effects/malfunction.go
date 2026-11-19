package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malfunction — Enchantment — Aura {3}{U} (Foundations Jumpstart):
//
//	"Enchant artifact or creature
//	 When this Aura enters, tap enchanted permanent.
//	 Enchanted permanent doesn't untap during its controller's untap step."
//
// Claustrophobia's pair of abilities on a wider enchant clause. The tap
// reads the Aura's attachment as the trigger resolves, so it needs no
// creature check.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "d2f98983-f395-4166-aa5a-309b4769baab",
		Name:                  "Malfunction",
		Completeness:          CompletenessFull,
		Targets:               TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered:             []game.TriggeredAbility{WhenThisEnters("Malfunction — tap enchanted permanent", tapEnchantedCreatureOnEntry)},
	})
}
