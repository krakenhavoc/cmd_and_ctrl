package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Maggot Therapy — Enchantment — Aura {2}{B} (Mercadian Masques, S82 OLD batch 686, #5619):
//
//	"Flash
//	Enchant creature
//	Enchanted creature gets +2/-2."
//
// Flash, and an Aura's layer 7c +2/-2 on the creature it is attached to (a creature whose toughness reaches 0 dies as a state-based action).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ab4887e6-f71b-462d-9243-9ebe54da98f4",
		Name:            "Maggot Therapy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static:          []game.StaticAbility{effects.PumpAttached(2, -2)},
	})
}
