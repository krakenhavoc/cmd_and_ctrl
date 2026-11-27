package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paralyzing Grasp — Enchantment — Aura {2}{U} (Return to Ravnica, S82 OLD):
//
//	"Enchant creature
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Roots' untap-step restriction without the tap on entry: a creature that is
// already tapped stays tapped, an untapped one is not stopped from tapping
// and untapping by other means.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "c1459d75-2bc2-4dae-8c7e-4e759a4dffc9",
		Name:                  "Paralyzing Grasp",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
	})
}
