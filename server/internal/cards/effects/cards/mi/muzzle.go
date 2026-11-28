package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Muzzle — Enchantment — Aura {1}{W} (Mercadian Masques, S82 OLD batch 686, #5619):
//
//	"Enchant creature
//	Prevent all damage that would be dealt by enchanted creature."
//
// Defang's CR 615.1a prevention: the source is the permanent the Aura is attached to, whatever it would hit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bdf497b5-166f-46a7-8e18-ae0b8f97768c",
		Name:         "Muzzle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtByAttached("Muzzle: prevent all damage that would be dealt by enchanted creature."),
		},
	})
}
