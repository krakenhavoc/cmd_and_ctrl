package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tidal Kraken — Creature — Kraken {5}{U}{U}{U}, 6/6 (Eighth Edition, #5607):
//
//	"This creature can't be blocked."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "104a8f76-d178-4a33-bb59-e8767da3d982",
		Name:         "Tidal Kraken",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
	})
}
