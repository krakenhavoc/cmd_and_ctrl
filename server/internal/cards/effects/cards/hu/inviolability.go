package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inviolability — Enchantment — Aura {1}{W} (Mercadian Masques, S82 OLD batch 685, #5619):
//
//	"Enchant creature
//	 Prevent all damage that would be dealt to enchanted creature."
//
// A CR 615.1a prevention static whose recipient is the creature the Aura is attached to (Panther Habit's ToEquippedCreature), any damage from any source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f336198-2425-447a-88b4-9be21ef7e90a",
		Name:         "Inviolability",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToEquippedCreature,
				Label: "Inviolability — prevent all damage that would be dealt to enchanted creature",
			}),
		},
	})
}
