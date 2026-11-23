package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Craven Giant — Creature — Giant {2}{R}, 4/1 (Portal, #5630):
//
//	"This creature can't block."
//
// A static restriction on itself (RestrictSelf), read by the block check
// (CR 509.1b). It can still attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ebd9c8d-c840-4c2e-ace6-352a68d1c20f",
		Name:         "Craven Giant",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
