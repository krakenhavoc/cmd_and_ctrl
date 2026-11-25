package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mystic Subdual — Enchantment — Aura {1}{U} (Ikoria: Lair of Behemoths, S82
// OLD batch 120, #5512):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets -2/-0 and loses all abilities. (Mutating onto the
//	 creature won't give it new abilities. It can gain abilities in other
//	 ways.)"
//
// The -2/-0 is PumpAttached (layer 7c). The ability loss is Kasmina's
// Transmutation's layer 6 removal, timestamp ordered with the Aura's own, so
// an ability granted afterwards still works.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fdd423d5-f19b-4bb2-98ca-75cc3c058d32",
		Name:            "Mystic Subdual",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.PumpAttached(-2, 0),
			effects.LoseAllAbilities(),
		},
	})
}
