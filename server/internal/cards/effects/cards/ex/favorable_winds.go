package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Favorable Winds — Enchantment {1}{U} (EDHREC rank 3216):
//
//	"Creatures you control with flying get +1/+1."
//
// The fliers anthem. A layer 7c modify over the controller's
// creatures that have flying when layer 7c runs — printed, granted
// by a lord, or granted until end of turn, since layer 6 has already
// been applied to every permanent by then
// (b30CreaturesYouControlWithFlying). A creature that loses flying
// loses the bonus with it, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2361ca87-6352-4ba3-8d91-b3d71242914d",
		Name:         "Favorable Winds",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.CreaturesYouControlWithFlyingGetPlusOne()},
	})
}
