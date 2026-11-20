package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feebleness — Enchantment — Aura {1}{B} (Time Spiral, #5596):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets -2/-1."
//
// Despondency's static at -2/-1 with flash and no return: a creature at
// toughness 1 dies to the state-based check (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a1a7757f-7e9b-4b5b-b9df-e98d8918caac",
		Name:            "Feebleness",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			PumpAttached(-2, -1),
		},
	})
}
