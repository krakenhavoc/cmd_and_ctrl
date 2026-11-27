package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glimmerdust Nap — Enchantment — Aura {2}{U} (Lorwyn, S82 OLD, #5593):
//
//	"Enchant tapped creature
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Entangling Vines' shape: the enchant clause is the Aura's legality check
// (CR 704.5m) and the untap-step restriction follows the Aura.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "8b476e35-77df-4332-b29d-8bf6f1bbbc3e",
		Name:                  "Glimmerdust Nap",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.TargetCreature("enchant tapped creature", effects.TappedPermanent()),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
	})
}
