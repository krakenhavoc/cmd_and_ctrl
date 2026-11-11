package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slither Blade — Creature {U}, 1/2: "This creature can't be blocked."
func init() {
	Register(Spec{
		OracleID:     "5a55641b-0818-4418-a2e5-44157056353c",
		Name:         "Slither Blade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
	})
}
