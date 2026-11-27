package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kyren Glider — Creature — Goblin {1}{R}, 1/1 (Mercadian Masques, S82 OLD
// b684, #5619):
//
//	"Flying
//	 This creature can't block."
//
// Flying is a printed keyword; the restriction is Aesthir Glider's
// (RestrictSelf, CantBlock).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6ae3a901-9f9a-4a99-94db-22c38fda7aef",
		Name:            "Kyren Glider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
