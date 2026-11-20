package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ice Over — Enchantment — Aura {1}{U} (Aether Revolt,
// #5536):
//
//	"Enchant artifact or creature
//	 Enchanted permanent doesn't untap during its controller's untap step."
//
// Malfunction without the enters tap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "80f8a64e-b22e-413e-a143-f1943be3a574",
		Name:                  "Ice Over",
		Completeness:          CompletenessFull,
		Targets:               TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
	})
}
