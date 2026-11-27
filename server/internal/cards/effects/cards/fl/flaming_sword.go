package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flaming Sword — Enchantment — Aura {1}{R} (Mercadian Masques, S82 OLD batch 685, #5619):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +1/+0 and has first strike."
//
// Flash, and Lightning Reflexes' pump and keyword: layer 7c +1/+0 and layer 6 first strike on the creature it is attached to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "05c62f91-2a5b-4cba-8e66-78fb370ea409",
		Name:            "Flaming Sword",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("first strike"),
		},
	})
}
