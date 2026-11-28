package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cave Sense — Enchantment — Aura {1}{R} (Mercadian Masques, S82 OLD batch 685, #5619):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has mountainwalk.
//	 (It can't be blocked as long as defending player controls a Mountain.)"
//
// Volcanic Strength's shape at +1/+1: the keyword is the engine's own mountainwalk.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d2fcdac-c8d5-43ca-81bc-130744ed8d3d",
		Name:         "Cave Sense",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("mountainwalk"),
		},
	})
}
