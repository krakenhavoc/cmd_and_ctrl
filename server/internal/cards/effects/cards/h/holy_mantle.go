package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Holy Mantle — Enchantment — Aura {2}{W}{W} (Gatecrash, S82 OLD b407, #5565):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has protection from creatures."
//
// Spirit Mantle at +2/+2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ed8f5bc-022b-4b1b-8592-57694bf59a0b",
		Name:         "Holy Mantle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("protection from creatures"),
		},
	})
}
