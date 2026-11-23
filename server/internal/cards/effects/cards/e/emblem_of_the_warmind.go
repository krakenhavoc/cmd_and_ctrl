package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emblem of the Warmind — Enchantment — Aura {1}{R} (Future Sight, #5594):
//
//	"Enchant creature you control
//	 Creatures you control have haste."
//
// Fervor's static on an Aura that may only be attached to a creature its
// controller controls (Mantle of the Ancients' enchant clause). The
// grant does not depend on the enchanted creature beyond the Aura staying
// on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86e88219-0f8e-405d-8439-c3b4acf11330",
		Name:         "Emblem of the Warmind",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{YoursOnly: true}, "haste"),
		},
	})
}
