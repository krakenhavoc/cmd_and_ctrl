package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hulking Ogre — Creature — Ogre {2}{R}, 3/3 (Urza's Destiny, #5621):
//
//	"This creature can't block."
//
// Hulking Cyclops' restriction on a 3/3: the defending player can't
// declare it as a blocker (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a95aa9ba-fa61-4a3d-a493-1ea790bd20ad",
		Name:         "Hulking Ogre",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
