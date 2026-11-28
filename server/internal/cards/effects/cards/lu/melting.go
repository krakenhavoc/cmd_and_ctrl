package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Melting — Enchantment {3}{R} (Ice Age, S82 OLD batch 848, #5636):
//
//	"All lands are no longer snow."
//
// The #6199 layer 4 static (AllLandsAreNoLongerSupertype): every land on the
// battlefield, whoever controls it, loses the snow supertype while Melting is
// there, in timestamp order with other type-changing effects (CR 613.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "599247f2-7f88-40ab-be08-e284e5f231d9",
		Name:         "Melting",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AllLandsAreNoLongerSupertype("Snow")},
	})
}
