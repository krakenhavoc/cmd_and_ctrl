package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maniacal Rage — Enchantment — Aura {1}{R} (Core Set 2020, #5518):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and can't block."
//
// A layer 7c pump and Sluggishness's can't-block restriction, which is checked
// as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5dbb212-0399-475b-a831-921f7c76cdc4",
		Name:         "Maniacal Rage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.RestrictAttached(game.CantBlock),
		},
	})
}
