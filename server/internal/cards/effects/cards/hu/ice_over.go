package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:              "80f8a64e-b22e-413e-a143-f1943be3a574",
		Name:                  "Ice Over",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.TargetPermanent("enchant artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
	})
}
