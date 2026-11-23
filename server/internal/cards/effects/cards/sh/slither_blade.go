package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slither Blade — Creature {U}, 1/2: "This creature can't be blocked."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a55641b-0818-4418-a2e5-44157056353c",
		Name:         "Slither Blade",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
	})
}
