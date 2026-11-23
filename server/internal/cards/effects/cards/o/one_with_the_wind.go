package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "9203bee0-718a-4ede-9556-5521417c4080",
		Name:         "One With the Wind",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying"),
		},
	})
}
