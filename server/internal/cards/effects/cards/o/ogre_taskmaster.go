package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Taskmaster — Creature — Ogre {3}{R}, 4/3 (Portal Second Age,
// #5625):
//
//	"This creature can't block."
//
// Jungle Lion's block ban: a RestrictSelf static the declare-blockers
// check reads (CR 509.1b). It can still attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5604f04e-fd72-41bf-906f-eb89d4a3475d",
		Name:         "Ogre Taskmaster",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
