package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orcish Oriflamme — Enchantment {3}{R} (Eternal Masters, S82 OLD, #5542):
//
//	"Attacking creatures you control get +1/+0."
//
// War Horn's static: a layer 7c +1/+0 for each creature of yours declared as an
// attacker, applied only while it attacks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b16a650-68b0-44dc-a9e1-15b7966e0b18",
		Name:         "Orcish Oriflamme",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttackingCreaturesYouControlGetPower(1)},
	})
}
