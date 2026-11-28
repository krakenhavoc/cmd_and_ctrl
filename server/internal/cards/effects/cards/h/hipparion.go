package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hipparion — Creature — Horse {1}{W}, 1/3 (Ice Age, S82 OLD, #5636):
//
//	"This creature can't block creatures with power 3 or greater unless you pay {1}."
//
// A block cost priced per attacker blocked (CR 509.1d, BlockTax.PerAttacker): blocking a 2/2 is free,
// blocking a 3/3 costs {1}, and the attacker's power is read as blockers are declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "306fd9d2-5ded-4857-90f4-e62d73c6aed2",
		Name:         "Hipparion",
		Completeness: effects.CompletenessFull,
		BlockTaxes: []game.BlockTax{
			effects.CantBlockAttackersUnlessYouPay("{1}", effects.PowerAtLeast(3), "This creature can't block creatures with power 3 or greater unless you pay {1}."),
		},
	})
}
