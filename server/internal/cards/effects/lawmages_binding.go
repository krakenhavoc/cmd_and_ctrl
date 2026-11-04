package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "10f0e797-3887-4c2b-9cff-80f110b952e1",
		Name:            "Lawmage's Binding",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(
				game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana,
			),
		},
	})
}
