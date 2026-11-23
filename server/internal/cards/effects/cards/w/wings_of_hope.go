package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wings of Hope — Enchantment — Aura {W}{U} (Invasion, #5616):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+3 and has flying."
//
// Drake Umbra's two statics (PumpAttached, GrantToAttached) at +1/+3,
// without the umbra armor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1df6359-edf4-48cf-b8d1-6240ac291cf7",
		Name:         "Wings of Hope",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 3),
			effects.GrantToAttached("flying"),
		},
	})
}
