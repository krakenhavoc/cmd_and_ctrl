package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// One With the Wind — Enchantment — Aura {1}{U} (Ixalan, #5532):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has flying."
//
// Wings of Hope's shape: a +2/+2 and flying for whatever creature the Aura
// is attached to (layers 7c and 6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9203bee0-718a-4ede-9556-5521417c4080",
		Name:         "One With the Wind",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("flying"),
		},
	})
}
