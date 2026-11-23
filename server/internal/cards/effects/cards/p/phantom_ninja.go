package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phantom Ninja — Creature — Illusion Ninja {1}{U}{U}, 2/2 (Foundations
// Jumpstart):
//
//	"This creature can't be blocked."
//
// A static restriction on itself, as Phantom Warrior's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7a13f43-1171-4808-9964-c5ad424027cb",
		Name:         "Phantom Ninja",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
	})
}
