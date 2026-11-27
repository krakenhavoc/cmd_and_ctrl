package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Entangling Vines — Enchantment — Aura {3}{G} (Magic 2010, #5585):
//
//	"Enchant tapped creature
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// The enchant clause is a target spec, so it is re-run as the Aura's legality
// check (CR 704.5m): the Aura goes away if the creature is untapped by an
// effect, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "51e7405e-a714-4c76-8913-10d626d2361e",
		Name:                  "Entangling Vines",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.TargetCreature("enchant tapped creature", effects.TappedPermanent()),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
	})
}
