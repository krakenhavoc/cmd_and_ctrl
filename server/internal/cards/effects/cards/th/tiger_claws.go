package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tiger Claws — Enchantment — Aura {2}{G} (Mercadian Masques, S82 OLD batch 685, #5619):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +1/+1 and has trample."
//
// Flash, and an Aura's +1/+1 and trample on the creature it is attached to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2ef8ebc9-4f95-42f8-86e6-85eff0b8f021",
		Name:            "Tiger Claws",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("trample"),
		},
	})
}
