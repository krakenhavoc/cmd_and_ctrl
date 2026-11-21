package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swashbuckling — Enchantment — Aura {1}{R} (Ixalan, #5532):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has haste."
//
// Wings of Hope's shape: a +2/+2 and haste for whatever creature the Aura
// is attached to (layers 7c and 6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4992144f-7a36-4f30-9e8a-4355288b50c5",
		Name:         "Swashbuckling",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("haste"),
		},
	})
}
