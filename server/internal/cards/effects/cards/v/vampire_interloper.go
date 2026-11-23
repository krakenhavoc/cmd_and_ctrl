package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire Interloper — Creature — Vampire Scout {1}{B}, 2/1
// (Foundations, #4465):
//
//	"Flying
//	 This creature can't block."
//
// Flying from PrintedKeywords, and Carrion Feeder's "can't block" as a
// RestrictSelf static.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b753d7dd-4c6b-4472-a009-671572f927eb",
		Name:            "Vampire Interloper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
