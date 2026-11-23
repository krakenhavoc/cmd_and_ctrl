package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lawmage's Binding — Enchantment — Aura {1}{W}{U} (Jumpstart Azorius
// deck request, #4894):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature can't attack or block, and its activated
//	 abilities can't be activated."
//
// Arrest with flash. "Its activated abilities" has no carve-out, so a
// mana ability is stopped too (CR 602.5): both bits, as Arrest sets
// them, checked before any cost is paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "10f0e797-3887-4c2b-9cff-80f110b952e1",
		Name:            "Lawmage's Binding",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(
				game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana,
			),
		},
	})
}
