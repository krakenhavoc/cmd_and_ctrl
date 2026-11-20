package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ashenmoor Gouger — Creature — Elemental Warrior {B/R}{B/R}{B/R}, 4/4
// (Shadowmoor, #5591):
//
//	"This creature can't block."
//
// RestrictSelf, Changeling Outcast's shape: the creature prints a
// restriction on itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eaedc09f-43b8-42db-a645-0ac3baf1de12",
		Name:         "Ashenmoor Gouger",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
