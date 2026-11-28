package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Strength — Enchantment — Aura {R}{R} (Tempest, S82 OLD, #5628):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2."
//
// Hero's Resolve's shape: one layer 7c modification read off the attachment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db85ba13-f00d-4cdd-99e1-22a4d39c8837",
		Name:         "Giant Strength",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
	})
}
