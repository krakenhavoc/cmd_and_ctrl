package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dehydration — Enchantment — Aura {3}{U} (Eighth Edition, S82 older sets
// batch 617, #5607):
//
//	"Enchant creature
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Bitter Chill's untap lock, without the tap or the leaves trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "db27c686-e202-4b60-9a10-0a0fef25576c",
		Name:                  "Dehydration",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
	})
}
