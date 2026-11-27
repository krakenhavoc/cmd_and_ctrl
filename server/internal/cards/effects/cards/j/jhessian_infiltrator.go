package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jhessian Infiltrator — {G}{U} Creature — Human Rogue 2/2 (Shards of Alara,
// S82 OLD b513):
//
//	"This creature can't be blocked."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ec8dc98-adc7-4fc3-94e5-31037235ceab",
		Name:         "Jhessian Infiltrator",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
	})
}
