package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gift of Granite — Enchantment — Aura {W} (Future Sight, #5594):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Enchant creature
//	 Enchanted creature gets +0/+2."
//
// A flash Aura with the shared attached pump, so it can be cast in
// response to damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1abac67-ae07-4bb2-a048-7e41e0fc0371",
		Name:            "Gift of Granite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{effects.PumpAttached(0, 2)},
	})
}
