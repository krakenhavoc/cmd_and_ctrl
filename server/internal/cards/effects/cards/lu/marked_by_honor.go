package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marked by Honor — Enchantment — Aura {3}{W} (Magic 2015, #5556):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has vigilance."
//
// Knightly Valor's two statics on the host without the token: the enchant
// clause is the Aura spell's target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa5a7940-976d-4d96-9ddf-1c3891885c7e",
		Name:         "Marked by Honor",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("vigilance"),
		},
	})
}
