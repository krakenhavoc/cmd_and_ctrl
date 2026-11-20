package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5604f04e-fd72-41bf-906f-eb89d4a3475d",
		Name:         "Ogre Taskmaster",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
