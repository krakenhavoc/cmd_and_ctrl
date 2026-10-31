package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mist-Cloaked Herald — Creature — Merfolk Warrior {U}, 1/1 (Marvel
// Super Heroes Commander, #4455): "This creature can't be blocked."
// Blighted Agent's self-restriction, with no keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "74255add-2894-4c31-81c1-d38873fa325f",
		Name:         "Mist-Cloaked Herald",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
	})
}
